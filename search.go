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

type SearchResult struct {
	move  Move
	score int
}

type BoundUpdate struct {
	bound        int
	alpha_update bool
}

var search_id int = 0

func iterative_deepening(brd *Board) int {
	return 0
}

func young_brothers_wait(brd *Board, alpha, beta, depth, ply int, cancel chan bool, update chan BoundUpdate) int {

	if depth <= 0 {
		return quiescence(brd, alpha, beta, depth, ply, cancel) // q-search is sequential.
	}

	in_check := is_in_check(brd)
	old_alpha, score := alpha, -INF
	legal_moves := false

	cancel_child := make(chan bool)
	var listeners []chan BoundUpdate

	if brd.halfmove_clock >= 100 {
		// to do: if this is not checkmate, return a draw
	}

	// to do: add adaptive depth for null move search.
	null_depth := depth - 2

	// search hash move
	first_move, hash_result := main_tt.probe(brd, depth, null_depth, alpha, beta, &score)
	switch hash_result {
	case MATCH_FOUND:
		return score
	case NO_MATCH:
		if depth > IID_MIN {
			if is_cancelled(cancel, cancel_child, listeners) {
				return 0
			}
			// To do: use IID to get a decent first move to try.
		}
	default:
		if is_cancelled(cancel, cancel_child, listeners) {
			return 0
		}

		update_child := make(chan BoundUpdate, 3)
		listeners = append(listeners, update_child)
		legal_moves = true

		score = ybw_make_unmake(brd, first_move, alpha, beta, depth-1, ply+1, cancel_child, update_child)
		if score > alpha {
			if score >= beta {
				main_tt.store(brd, first_move, depth, LOWER_BOUND, score)
				return beta
			}
			alpha = score
		}
	}
	// Generate tactical (non-quiet) moves.  Good moves will be searched sequentially to establish good bounds
	// before remaining nodes are searched in parallel.

	best_moves, remaining_moves := get_best_moves(brd, in_check)
	var m, best_move Move
	for _, item := range *best_moves { // search the best moves sequentially.
		m = item.move
		if is_cancelled(cancel, cancel_child, listeners) {
			return 0
		} // make sure the job hasn't been cancelled.

		update_child := make(chan BoundUpdate, 3)
		listeners = append(listeners, update_child)
		legal_moves = true

		score = ybw_make_unmake(brd, m, alpha, beta, depth-1, ply+1, cancel_child, update_child)
		if score > alpha {
			if score >= beta {
				main_tt.store(brd, m, depth, LOWER_BOUND, score)
				return beta
			}
			alpha = score
			best_move = m
		}
	}
	// Delay the generation of remaining moves until all promotions and winning captures have been searched.
	// if a cutoff occurs, this will reduce move generation effort substantially.
	get_remaining_moves(brd, in_check, remaining_moves)

	if depth <= SPLIT_MIN { // Depth is too shallow for parallel search to be worthwhile.
		for _, item := range *remaining_moves { // search remaining moves sequentially.
			m = item.move
			if is_cancelled(cancel, cancel_child, listeners) {
				return 0
			} // make sure the job hasn't been cancelled.

			update_child := make(chan BoundUpdate, 3)
			listeners = append(listeners, update_child)
			legal_moves = true

			score = ybw_make_unmake(brd, m, alpha, beta, depth-1, ply+1, cancel_child, update_child)
			if score > alpha {
				if score >= beta {
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return beta
				}
				alpha = score
				best_move = m
			}
		}
	} else { // now that decent bounds have been established, search the remaining nodes in parallel.
		result_child := make(chan SearchResult, 10)
		var child_counter int
		for _, item := range *remaining_moves {
			m := item.move
			new_brd := brd.Copy() // create a locally scoped deep copy of the board.
			update_child := make(chan BoundUpdate, 3)
			listeners = append(listeners, update_child)
			legal_moves = true

			go func() {
				score := ybw_make_unmake(new_brd, m, alpha, beta, depth-1, ply+1, cancel_child, update_child)
				result_child <- SearchResult{m, score}
			}()
			child_counter++
		}

		if child_counter > 0 {
		remaining_pieces:
			for {
				select { // wait for a message to come in on one of the channels.
				case <-cancel: // task was cancelled.

					cancel_work(cancel_child, listeners)
					return 0

				case bound_update := <-update: // an updated bound was received from the parent node.
					// may also want to check this before sequential searches.
					if bound_update.alpha_update {
						if bound_update.bound < beta {
							beta = bound_update.bound // update relevant local bound
							for _, update_child := range listeners {
								update_child <- BoundUpdate{-beta, false} // broadcast update to child nodes.
							}
						}
					} else {
						if bound_update.bound > alpha {
							alpha = bound_update.bound
							for _, update_child := range listeners {
								update_child <- BoundUpdate{-alpha, true}
							}
						}
					}

				case result := <-result_child: // one of the child subtrees has been completely searched.

					if result.score > alpha {
						if score >= beta {
							main_tt.store(brd, result.move, depth, LOWER_BOUND, result.score)
							cancel_work(cancel_child, listeners)
							return beta
						}
						alpha = result.score
						best_move = result.move
						for _, update_child := range listeners {
							update_child <- BoundUpdate{-alpha, true} // send the updated bound to child processes.
						}
					}
					child_counter--
					if child_counter == 0 {
						break remaining_pieces // exit the for loop
					}
				}
			}
		}
	}

	if legal_moves { 
		var result_type int
		if alpha > old_alpha {
			result_type = EXACT
		} else {
			result_type = UPPER_BOUND
		}
		main_tt.store(brd, best_move, depth, result_type, alpha)
		return alpha
	} else { // draw or checkmate detected.
		if in_check {
			score = ply - INF
		} else {
			score = 0
		}
		main_tt.store(brd, 0, MAX_PLY, EXACT, score)
		return score
	}



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
	legal_moves := false

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

	var m Move
	if in_check {
		var best_moves, remaining_moves *MoveList
		get_evasions(brd, best_moves, remaining_moves)
		for _, item := range *best_moves { // search the best moves sequentially.
			m = item.move
			legal_moves = true
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
			legal_moves = true
			score = q_make_unmake(brd, m, alpha, beta, depth-1, ply+1, cancel)
			if score > alpha {
				if score >= beta {
					return beta
				}
				alpha = score
			}
		}
	} else {
		// to do:  add futility pruning
		best_moves := get_winning_captures(brd)
		for _, item := range *best_moves { // search the best moves sequentially.
			m = item.move
			legal_moves = true
			score = q_make_unmake(brd, m, alpha, beta, depth-1, ply+1, cancel)
			if score > alpha {
				if score >= beta {
					return beta
				}
				alpha = score
			}
		}
	}

	if legal_moves { 
		return alpha
	} else { // draw or checkmate detected.
		if in_check {
			return ply - INF
		} else {
			return 0
		}
	}
}

func ybw_make_unmake(brd *Board, m Move, alpha, beta, depth, ply int, cancel chan bool, update chan BoundUpdate) int {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score := -young_brothers_wait(brd, -beta, -alpha, depth-1, ply+1, cancel, update)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return score
}

func q_make_unmake(brd *Board, m Move, alpha, beta, depth, ply int, cancel chan bool) int {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score := -quiescence(brd, -beta, -alpha, depth-1, ply+1, cancel)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return score
}

func is_cancelled(cancel, cancel_child chan bool, listeners []chan BoundUpdate) bool {
	select {
	case <-cancel:
		cancel_work(cancel_child, listeners)
		return true
	default:
		return false
	}
}

func cancel_work(cancel_child chan bool, listeners []chan BoundUpdate) {
	close(cancel_child)
	for _, update_child := range listeners {
		close(update_child)
	}
}
