//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//-----------------------------------------------------------------------------------

package main

// Modified Young Brothers Wait (YBW) approach

// At each node, search the most promising (leftmost) child sequentially first,
// then send the rest of the successors to the load balancer.
// The requester then blocks until it receives a cancellation flag, a result, or an updated bound.

// Completion

// When each child completes, its result is sent back via a channel to the requester node.
// On completion, each node sends a cancellation flag indicating no more work is needed.

// Alpha Updates

// Bounds are stored locally in the call stack.
// When alpha is updated, the update is piped down to all running child requests via a channel.  For requests still
// in queue, use a closure to scope the arguments so that when a worker executes the job it does so with the latest
// bounds from the requestor.

// When an alpha update is received from the node above, use the received value to update the locally scoped alpha
// value.

// Cancellation

// If a beta cutoff occurs, a cancellation flag is sent on the cancellation channel for the current master node.

// Each request made by the master node contains a pointer to a cancellation channel.
// Prior to starting work on a new request, the worker checks the cancellation channel to see if the work is still
// needed, and discards the request if not.

// Each node periodically reads the cancellation channel from the frame above it.  When a cancellation message
// is received, the frame sends a cancellation message on its own channel to be read by the frames below it and any
// workers who have the current frame's requests in queue.

// Spawning behavior

// If some heuristic for the strength of move ordering were available at the current node
// (variance of ordering scores?), this info could be used to influence spawning behavior.
// When some moves are scored far better than others, those nodes would be searched sequentially in hopes of
// achieving a cutoff without incurring communication overhead.

// Without a meaningful heuristic, spawning tactic could alternatively be based on node type.

import (
	// "sync"
	"github.com/stephenjlovell/gopher_check/load_balancer"
)

// channel of channels for managed closing?

var work chan load_balancer.Request

func young_brothers_wait(brd *BRD, old_alpha, old_beta, depth, ply int, cancel chan bool, update chan int) int {

	alpha, beta, score := -old_beta, -old_alpha, -INF

	update_child := make(chan int)
	cancel_child := make(chan bool)

	if depth <= 0 {
		return quiescence(brd, alpha, beta, depth, ply, cancel_child)
	} // call standard sequential q-search

	in_check := is_in_check(brd /* c, e */) // move c, e into BRD struct to avoid constantly passing these around.

	best_moves, other_moves := split_moves(brd, in_check) // slice off the best few nodes to search sequentially

	// search the best moves sequentially.
	for _, m := range best_moves {
		if is_cancelled(cancel, cancel_child, update_child) {
			return 0
		} // make sure the job hasn't been cancelled.
		make_move(brd, m) // to do: make move
		score = young_brothers_wait(brd, alpha, beta, depth-1, ply+1, cancel_child, update_child) * -1
		unmake_move(brd, m) // to do: unmake move
		if score >= beta {
			// to do: save result to transposition table before returning.
			return beta // no communication necessary; nothing has been spawned in parallel from this node yet.
		}
		if score > alpha {
			alpha = score // no communication necessary; nothing has been spawned in parallel from this node yet.
		}
	}

	// now that decent bounds have been established, search the remaining nodes in parallel.
	result_child := make(chan int)
	var child_counter int
	for _, m := range other_moves { // search the remaining moves in parallel.
		new_brd := brd.Copy() // create a locally scoped deep copy of the board.

		req := load_balancer.Request{ // package the subtree search into a Request object
			Cancel: cancel_child,
			Result: result_child,
			Size:   (3 << uint(depth-1)), // estimate of the number of main search leaf nodes remaining
			Fn: func() int {
				make_move(new_brd, m) // to do: make move
				val := young_brothers_wait(new_brd, alpha, beta, depth-1, ply+1, cancel_child, update_child) * -1
				unmake_move(new_brd, m) // to do: unmake move
				return val
			},
		}
		work <- req // pipe the new request object to the load balancer to execute in parallel.
		child_counter++
	}

	if child_counter > 0 {
	remaining_pieces:
		for { // wait for a message to come in on one of the channels
			select {
			case <-cancel: // task was cancelled.

				cancel_work(cancel_child, update_child)
				return 0

			case updated := <-update: // an updated bound was received from the parent node.

				if updated > alpha {
					alpha = updated
				}
				update_child <- updated // propegate updated bound to child nodes

			case score = <-result_child: // one of the child subtrees has been completely searched.
				if score >= beta {
					// to do: save result to transposition table before returning.
					cancel_work(cancel_child, update_child)
					return beta
				}
				if score > alpha {
					alpha = score
					update_child <- alpha // send the updated bound to child processes.
				}
				child_counter--
				if child_counter == 0 {
					break remaining_pieces
				}
			}
		}
	}

	// to do: check for draw or checkmate
	// to do: save result to transposition table before returning.
	return 0
}

func is_cancelled(cancel, cancel_child chan bool, update_child chan int) bool {
	select {
	case <-cancel:
		cancel_work(cancel_child, update_child)
		return true
	default:
		return false
	}
}

func cancel_work(cancel_child chan bool, update_child chan int) {
	cancel_child <- true
	close(cancel_child)
	close(update_child)
}

// Q-Search will always be done sequentially.
// Q-search subtrees are taller and narrower than in the main search making benefit of parallelism
// smaller and raising communication and synchronization overhead.
func quiescence(brd *BRD, alpha, beta, depth, ply int, cancel chan bool) int {

	return 0
}
