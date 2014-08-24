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

// Each request made by the master node contains a pointer to a cancellation channel.
// If a beta cutoff occurs, a cancellation flag is sent on the cancellation channel for the current master node.

// Prior to starting work on a new request, the worker checks the cancellation channel to see if the work is still
// needed, and discards the request if not.

// Each node periodically reads the cancellation channel from the frame above it.  When a cancellation message
// is received, the frame sends a cancellation message on its own channel to be read by the frames below it and any
// workers who have the current frame's requests in queue.

// Spawning behavior

// When some moves are scored far better than others, those nodes would be searched sequentially in hopes of
// achieving a cutoff without incurring communication overhead.

// Search phases

// To reduce synchronization overhead, all search below some depth threshold will be handled sequentially.

// Hash Move (Always)
// YBW/Parallel search allowed (Ply <= 5)
// IID (Depth >= 4)
// Null Move (Depth >= 3)
// Futility pruning (Depth <= 2)

import (
// "github.com/stephenjlovell/gopher_check/load_balancer"
)

const (
	MAX_DEPTH = 10
	SPLIT_MIN = 2
	EXT_MAX   = 4
	MAX_PLY   = MAX_DEPTH + EXT_MAX
	IID_MIN   = 4
)

type PV []Move

func iterative_deepening(brd *Board) int {
	return 0
}

func young_brothers_wait(brd *Board, old_alpha, old_beta, depth, ply int, cancel chan bool, update chan int) int {

	if depth <= 0 {
		return quiescence(brd, old_alpha, old_beta, depth, ply, cancel) // q-search is sequential.
	}

	in_check := is_in_check(brd)
	alpha, beta, score := -old_beta, -old_alpha, -INF
	update_child := make(chan int)
	cancel_child := make(chan bool)

	if brd.halfmove_clock >= 100 {
		// to do: if this is not checkmate, return a draw
	}

	// search hash move
	best_move := main_tt.probe(brd, depth)
	if best_move > 0 {
		if is_cancelled(cancel, cancel_child, update_child) {
			return 0
		} // make sure the job hasn't been cancelled.
		score = ybw_make_unmake(brd, best_move, alpha, beta, depth-1, ply+1, cancel_child, update_child)
		if score > alpha {
			if score >= beta {
				// to do: save result to transposition table before returning.
				return beta // no communication necessary; nothing has been spawned in parallel from this node yet.
			}
			alpha = score // no communication necessary; nothing has been spawned in parallel from this node yet.
		}
	} else if depth > IID_MIN {
		// To do: use IID to get a decent best move to try.
	}

	// Generate tactical (non-quiet) moves.  Good moves will be searched sequentially to establish good bounds
	// before remaining nodes are searched in parallel.
	best_moves, remaining_moves := get_best_moves(brd, in_check)
	var m Move
	for _, item := range *best_moves { // search the best moves sequentially.
		m = item.move
		if is_cancelled(cancel, cancel_child, update_child) {
			return 0
		} // make sure the job hasn't been cancelled.
		score = ybw_make_unmake(brd, m, alpha, beta, depth-1, ply+1, cancel_child, update_child)

		if score > alpha {
			if score >= beta {
				// to do: save result to transposition table before returning.
				return beta // no communication necessary; nothing has been spawned in parallel from this node yet.
			}
			alpha = score // no communication necessary; nothing has been spawned in parallel from this node yet.
		}
	}
	// Delay the generation of remaining moves until all promotions and winning captures have been searched.
	// if a cutoff occurs, this will reduce move generation effort substantially.
	get_remaining_moves(brd, in_check, remaining_moves)

	if depth <= SPLIT_MIN { // Depth is too shallow for parallel search to be worthwhile.
		for _, item := range *remaining_moves { // search the best moves sequentially.
			m = item.move
			if is_cancelled(cancel, cancel_child, update_child) {
				return 0
			} // make sure the job hasn't been cancelled.
			score = ybw_make_unmake(brd, m, alpha, beta, depth-1, ply+1, cancel_child, update_child)
			if score > alpha {
				if score >= beta {
					// to do: save result to transposition table before returning.
					return beta
				}
				alpha = score
			}
		}
	} else { // now that decent bounds have been established, search the remaining nodes in parallel.
		result_child := make(chan int, 10)
		var child_counter int
		for _, item := range *remaining_moves {
			m := item.move
			new_brd := brd.Copy() // create a locally scoped deep copy of the board.
			go func() {
				result_child <- ybw_make_unmake(new_brd, m, alpha, beta, depth-1, ply+1, cancel_child, update_child)
			}()
			child_counter++
		}
		if child_counter > 0 { // wait for a message to come in on one of the channels
		remaining_pieces:
			for {
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
					if score > alpha {
						if score >= beta {
							// to do: save result to transposition table before returning.
							cancel_work(cancel_child, update_child)
							return beta
						}
						alpha = score
						update_child <- alpha // send the updated bound to child processes.
					}
					child_counter--
					if child_counter == 0 {
						break remaining_pieces // exit the for loop
					}
				}
			}
		}
	}

	// to do: check for draw or checkmate
	// to do: save result to transposition table before returning.
	return 0
}

// Q-Search will always be done sequentially.
// Q-search subtrees are taller and narrower than in the main search making benefit of parallelism
// smaller and raising communication and synchronization overhead.
func quiescence(brd *Board, old_alpha, old_beta, depth, ply int, cancel chan bool) int {

	select {
	case <-cancel:
		return 0
	default:
	}

	in_check := is_in_check(brd)
	alpha, beta, score := -old_beta, -old_alpha, -INF

	if brd.halfmove_clock >= 100 {
		// to do: if this is not checkmate, return a draw
	}

	if !in_check {
		score = brd.Evaluate() // stand pat
		if score > alpha {
			alpha = score
			if score >= beta {
				return beta
			}
		}
	}

	// to do:  add futility pruning
	var m Move
	if in_check {
		var best_moves, remaining_moves MoveList
		get_evasions(brd, &best_moves, &remaining_moves)
		for _, item := range *best_moves { // search the best moves sequentially.
			m = item.move
			score = q_make_unmake(brd, m, alpha, beta, depth-1, ply+1, cancel)
			if score > alpha {
				if score >= beta {
					return beta
				}
				alpha = score
			}
		}
		for _, item := range *remaining_moves { // search the best moves sequentially.
			m = item.move
			score = q_make_unmake(brd, m, alpha, beta, depth-1, ply+1, cancel)
			if score > alpha {
				if score >= beta {
					return beta
				}
				alpha = score
			}
		}
	} else {
		best_moves := get_winning_captures(brd, in_check)
		for _, item := range *best_moves { // search the best moves sequentially.
			m = item.move
			score = q_make_unmake(brd, m, alpha, beta, depth-1, ply+1, cancel)
			if score > alpha {
				if score >= beta {
					return beta
				}
				alpha = score
			}
		}
	}

	// to do: check for draw or checkmate

}

func ybw_make_unmake(brd *Board, m Move, alpha, beta, depth, ply int, cancel chan bool, update chan int) int {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score := -1 * young_brothers_wait(brd, alpha, beta, depth-1, ply+1, cancel, update)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return score
}

func q_make_unmake(brd *Board, m Move, alpha, beta, depth, ply int, cancel chan bool) int {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score := -1 * quiescence(brd, alpha, beta, depth-1, ply+1, cancel)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return score
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
	close(cancel_child)
	close(update_child)
}






