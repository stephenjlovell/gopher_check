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
	"fmt"
	"time"
)

const (
	MAX_TIME  = 120 // default search time limit in seconds
	MAX_DEPTH = 12
	SPLIT_MIN = 3
	EXT_MAX   = 4
	MAX_PLY   = MAX_DEPTH + EXT_MAX
	IID_MIN   = 4
)

type PV []Move

type SearchResult struct {
	move  Move
	score int
	count int
}

type BoundUpdate struct {
	bound        int
	alpha_update bool
}

var search_id int
var iid_move [2]Move
var iid_score [2]int
var iid_cancel chan bool

func AbortSearch() {
	if iid_cancel != nil {
		iid_cancel <- true
		close(iid_cancel)
	}
}

func search_timer(timer *time.Timer) {
	select {
	case <-timer.C:
		fmt.Println("Time up. Aborting search.")
		AbortSearch()
	}
}

func Search(brd *Board, restrict_search []Move, depth, time_limit int) Move {
	iid_cancel = make(chan bool, 1) // set up the search.
	iid_move[brd.c] = 0
	start := time.Now()
	timer := time.NewTimer(time.Duration(time_limit) * time.Second)

	go search_timer(timer) // abort the current search after time_limit seconds.

	move := iterative_deepening(brd, depth, start, iid_cancel)
	timer.Stop() // cancel the timer to prevent it from interfering with the next search if it's not
	// garbage collected before then.
	return move
}

func iterative_deepening(brd *Board, depth int, start time.Time, cancel chan bool) Move {
	var move Move
	var guess, count, sum int
	c := brd.c

	for d := 1; d <= depth; d++ {
		move, guess, count = ybw_root(brd, -INF, INF, iid_score[c], d, cancel)
		sum += count

		select {
		case <-cancel:
			return iid_move[c]
		default:
			if time.Since(start) > 15*time.Millisecond { // don't print info for first few plys to reduce communication traffic.
				PrintInfo(guess, d, sum, time.Since(start))
			}
			iid_move[c], iid_score[c] = move, guess
		}
	}
	PrintInfo(guess, depth, sum, time.Since(start))
	return iid_move[c]
}

func ybw_root(brd *Board, alpha, beta, guess, depth int, cancel chan bool) (Move, int, int) {

	sum, count := 1, 0

	in_check := is_in_check(brd)
	if in_check {
		fmt.Println("In check at root.")
	}
	old_alpha, score := alpha, -INF
	legal_moves := false

	cancel_child := make(chan bool, 1)
	var listeners []chan BoundUpdate

	if brd.halfmove_clock >= 100 {
		return 0, 0, 1
		// to do: if this is not checkmate, return a draw
	}

	// // search hash move
	// first_move, hash_result := main_tt.probe(brd, depth, depth-2, alpha, beta, &score)
	// if hash_result != NO_MATCH && first_move != 0 {

	// 	if is_cancelled(cancel, cancel_child, listeners) {
	// 		return first_move, score, sum
	// 	}
	// 	fmt.Printf("first_move: %s hash_result: %d\n", first_move.ToString(), hash_result)
	// 	update_child := make(chan BoundUpdate, 3)
	// 	listeners = append(listeners, update_child)
	// 	legal_moves = true

	// 	score, count = ybw_make(brd, first_move, alpha, beta, depth-1, 1, cancel_child, update_child)
	// 	sum += count
	// 	if score > alpha {
	// 		if score >= beta {
	// 			main_tt.store(brd, first_move, depth, LOWER_BOUND, score)
	// 			return first_move, beta, sum
	// 		}
	// 		alpha = score
	// 	}
	// }

	// Generate tactical (non-quiet) moves.  Good moves will be searched sequentially to establish good bounds
	// before remaining nodes are searched in parallel.

	best_moves, remaining_moves := get_best_moves(brd, in_check)
	var m, best_move Move
	for _, item := range *best_moves { // search the best moves sequentially.

		if is_cancelled(cancel, cancel_child, listeners) {
			return best_move, 0, sum
		} // make sure the job hasn't been cancelled.
		m = item.move
		if !avoids_check(brd, m, in_check) {
			// brd.Print()
			// fmt.Printf("illegal move: %s\n", m.ToString())
			continue
		}

		update_child := make(chan BoundUpdate, 3)
		listeners = append(listeners, update_child)
		legal_moves = true

		score, count = ybw_make(brd, m, alpha, beta, depth-1, 1, cancel_child, update_child)
		sum += count
		if score > alpha {
			if score >= beta {
				main_tt.store(brd, m, depth, LOWER_BOUND, score)
				return m, beta, sum
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

			if is_cancelled(cancel, cancel_child, listeners) {
				return best_move, 0, sum
			} // make sure the job hasn't been cancelled.

			m = item.move
			if !avoids_check(brd, m, in_check) {
				// brd.Print()
				// fmt.Printf("in check:%v\n", in_check)
				// fmt.Printf("illegal move: %s\n", m.ToString())
				continue
			}

			update_child := make(chan BoundUpdate, 3)
			listeners = append(listeners, update_child)
			legal_moves = true

			score, count = ybw_make(brd, m, alpha, beta, depth-1, 1, cancel_child, update_child)
			sum += count
			if score > alpha {
				if score >= beta {
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return m, beta, sum
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
			if !avoids_check(brd, m, in_check) {
				continue
			}
			new_brd := brd.Copy() // create a locally scoped deep copy of the board.
			update_child := make(chan BoundUpdate, 3)
			listeners = append(listeners, update_child)
			legal_moves = true
			child_counter++
			go func() {
				score, count := ybw_make(new_brd, m, alpha, beta, depth-1, 1, cancel_child, update_child)
				result_child <- SearchResult{m, score, count}
			}()
		}

		if child_counter > 0 {
		remaining_pieces:
			for {
				select { // wait for a message to come in on one of the channels.
				case <-cancel: // task was cancelled.
					cancel_work(cancel_child, listeners)
					return best_move, 0, sum

				case result := <-result_child: // one of the child subtrees has been completely searched.
					sum += result.count
					if result.score > alpha {
						if score >= beta {
							main_tt.store(brd, result.move, depth, LOWER_BOUND, result.score)
							cancel_work(cancel_child, listeners)
							return result.move, beta, sum
						}
						alpha = result.score
						best_move = result.move
						// for _, update_child := range listeners {
						// 	update_child <- BoundUpdate{-alpha, true} // send the updated bound to child processes.
						// }
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
		return best_move, alpha, sum
	} else { // draw or checkmate detected.
		// fmt.Println("No moves available at root.")
		if in_check {
			score = -INF
		} else {
			score = 0
		}
		main_tt.store(brd, 0, MAX_PLY, EXACT, score)
		return 0, score, sum
	}

}

func young_brothers_wait(brd *Board, alpha, beta, depth, ply int, cancel chan bool, update chan BoundUpdate) (int, int) {

	if depth <= 0 {
		return quiescence(brd, alpha, beta, depth, ply, cancel) // q-search is sequential.
	}

	sum, count := 1, 0
	in_check := is_in_check(brd)
	old_alpha, score := alpha, -INF
	legal_moves := false

	cancel_child := make(chan bool, 1)
	var listeners []chan BoundUpdate

	if brd.halfmove_clock >= 100 {
		return 0, 1
		// to do: if this is not checkmate, return a draw
	}

	// to do: add adaptive depth for null move search.
	// null_depth := depth - 2

	// // search hash move
	// first_move, hash_result := main_tt.probe(brd, depth, null_depth, alpha, beta, &score)
	// switch hash_result {
	// case MATCH_FOUND:
	// 	return score, sum
	// case NO_MATCH:
	// 	if depth > IID_MIN {
	// 		if is_cancelled(cancel, cancel_child, listeners) {
	// 			return 0, sum
	// 		}
	// 		// To do: use IID to get a decent first move to try.
	// 	}
	// default:
	// 	if is_cancelled(cancel, cancel_child, listeners) {
	// 		return 0, sum
	// 	}

	// 	update_child := make(chan BoundUpdate, 3)
	// 	listeners = append(listeners, update_child)
	// 	legal_moves = true

	// 	score, count = ybw_make(brd, first_move, alpha, beta, depth-1, ply+1, cancel_child, update_child)
	// 	sum += count
	// 	if score > alpha {
	// 		if score >= beta {
	// 			main_tt.store(brd, first_move, depth, LOWER_BOUND, score)
	// 			return beta, sum
	// 		}
	// 		alpha = score
	// 	}
	// }

	// Generate tactical (non-quiet) moves.  Good moves will be searched sequentially to establish good bounds
	// before remaining nodes are searched in parallel.
	best_moves, remaining_moves := get_best_moves(brd, in_check)
	var m, best_move Move
	for _, item := range *best_moves { // search the best moves sequentially.
		if is_cancelled(cancel, cancel_child, listeners) {
			return 0, sum
		} // make sure the job hasn't been cancelled.
		m = item.move
		if !avoids_check(brd, m, in_check) {
			continue
		}
		update_child := make(chan BoundUpdate, 3)
		listeners = append(listeners, update_child)
		legal_moves = true

		score, count = ybw_make(brd, m, alpha, beta, depth-1, ply+1, cancel_child, update_child)
		sum += count
		if score > alpha {
			if score >= beta {
				main_tt.store(brd, m, depth, LOWER_BOUND, score)
				return beta, sum
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

			if is_cancelled(cancel, cancel_child, listeners) {
				return 0, sum
			} // make sure the job hasn't been cancelled.
			m = item.move
			if !avoids_check(brd, m, in_check) {
				continue
			}
			update_child := make(chan BoundUpdate, 3)
			listeners = append(listeners, update_child)
			legal_moves = true

			score, count = ybw_make(brd, m, alpha, beta, depth-1, ply+1, cancel_child, update_child)
			sum += count
			if score > alpha {
				if score >= beta {
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return beta, sum
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
			if !avoids_check(brd, m, in_check) {
				continue
			}
			new_brd := brd.Copy() // create a locally scoped deep copy of the board.
			update_child := make(chan BoundUpdate, 3)
			listeners = append(listeners, update_child)
			legal_moves = true
			child_counter++
			go func() {
				score, count := ybw_make(new_brd, m, alpha, beta, depth-1, ply+1, cancel_child, update_child)
				result_child <- SearchResult{m, score, count}
			}()

		}

		if child_counter > 0 {
		remaining_pieces:
			for {
				select { // wait for a message to come in on one of the channels.
				case <-cancel: // task was cancelled.
					cancel_work(cancel_child, listeners)
					return 0, sum
					// case bound_update := <-update: // an updated bound was received from the parent node.
					// // may also want to check this before sequential searches.
					// if bound_update.alpha_update {
					// 	if bound_update.bound < beta {
					// 		beta = bound_update.bound // update relevant local bound
					// 		for _, update_child := range listeners {
					// 			update_child <- BoundUpdate{-beta, false} // broadcast update to child nodes.
					// 		}
					// 	}
					// } else {
					// 	if bound_update.bound > alpha {
					// 		alpha = bound_update.bound
					// 		for _, update_child := range listeners {
					// 			update_child <- BoundUpdate{-alpha, true}
					// 		}
					// 	}
					// }

				case result := <-result_child: // one of the child subtrees has been completely searched.
					sum += result.count
					if result.score > alpha {
						if score >= beta {
							main_tt.store(brd, result.move, depth, LOWER_BOUND, result.score)
							cancel_work(cancel_child, listeners)
							return beta, sum
						}
						alpha = result.score
						best_move = result.move
						// for _, update_child := range listeners {
						// 	update_child <- BoundUpdate{-alpha, true} // send the updated bound to child processes.
						// }
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
		return alpha, sum
	} else { // draw or checkmate detected.
		if in_check {
			score = ply - INF
		} else {
			score = 0
		}
		main_tt.store(brd, 0, MAX_PLY, EXACT, score)
		return score, sum
	}

}

// Q-Search will always be done sequentially.
// Q-search subtrees are taller and narrower than in the main search making benefit of parallelism
// smaller and raising communication and synchronization overhead.
func quiescence(brd *Board, alpha, beta, depth, ply int, cancel chan bool) (int, int) {

	// return brd.Evaluate(), 1
	sum, count := 1, 0

	select {
	case <-cancel:
		return 0, sum
	default:
	}

	in_check := is_in_check(brd)
	score := -INF
	legal_moves := false

	if brd.halfmove_clock >= 100 {
		return 0, 1
		// to do: if this is not checkmate, return a draw
	}

	if !in_check {
		score = brd.Evaluate() // stand pat
		if score > alpha {
			alpha = score
			if score >= beta {
				return beta, sum
			}
		}
	}

	var m Move
	if in_check {
		best_moves, remaining_moves := &MoveList{}, &MoveList{}
		get_evasions(brd, best_moves, remaining_moves)
		for _, item := range *best_moves { // search the best moves sequentially.
			m = item.move
			legal_moves = true
			score, count = q_make(brd, m, alpha, beta, depth-1, ply+1, cancel)
			sum += count
			if score > alpha {
				if score >= beta {
					return beta, sum
				}
				alpha = score
			}
		}
		for _, item := range *remaining_moves { // search the best moves sequentially.
			m = item.move
			legal_moves = true
			score, count = q_make(brd, m, alpha, beta, depth-1, ply+1, cancel)
			sum += count
			if score > alpha {
				if score >= beta {
					return beta, sum
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
			score, count = q_make(brd, m, alpha, beta, depth-1, ply+1, cancel)
			sum += count
			if score > alpha {
				if score >= beta {
					return beta, sum
				}
				alpha = score
			}
		}
	}

	if legal_moves {
		return alpha, sum
	} else { // draw or checkmate detected.
		if in_check {
			return ply - INF, sum
		} else {
			return 0, sum
		}
	}
}

func ybw_make(brd *Board, m Move, alpha, beta, depth, ply int, cancel chan bool, update chan BoundUpdate) (int, int) {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score, sum := young_brothers_wait(brd, -beta, -alpha, depth-1, ply+1, cancel, update)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return -score, sum
}

func ybw_parallell_make(brd *Board, m Move, alpha, beta, depth, ply int, cancel chan bool, update chan BoundUpdate) (int, int) {
	make_move(brd, m) // to do: make move
	score, sum := young_brothers_wait(brd, -beta, -alpha, depth-1, ply+1, cancel, update)
	return -score, sum
}

func q_make(brd *Board, m Move, alpha, beta, depth, ply int, cancel chan bool) (int, int) {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score, sum := quiescence(brd, -beta, -alpha, depth-1, ply+1, cancel)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return -score, sum
}

func is_cancelled(cancel, cancel_child chan bool, listeners []chan BoundUpdate) bool {
	select {
	case <-cancel:
		// fmt.Println("task cancelled")
		cancel_work(cancel_child, listeners)
		return true
	default:
		return false
	}
}

func cancel_work(cancel_child chan bool, listeners []chan BoundUpdate) {
	cancel_child <- true
	close(cancel_child)
	for _, update_child := range listeners {
		close(update_child)
	}
}
