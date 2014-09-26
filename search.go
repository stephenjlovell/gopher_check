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
	"fmt"
	"math"
	"time"
)

const (
	MAX_TIME  = 120 // default search time limit in seconds
	MAX_DEPTH = 12
	SPLIT_MIN = 13 // set > MAX_DEPTH to disable parallel search.
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
var cancel_search bool

func AbortSearch() {
	cancel_search = true
}

func search_timer(timer *time.Timer) {
	select {
	case <-timer.C:
		fmt.Println("Time up. Aborting search.")
		AbortSearch()
	}
}

func Search(brd *Board, restrict_search []Move, depth, time_limit int) (Move, int) {
	cancel_search = false
	iid_move[brd.c] = 0
	start := time.Now()
	timer := time.NewTimer(time.Duration(time_limit) * time.Second)

	if search_id >= 512 { // only 9 bits are available to store the id in each TT entry.
		search_id = 1
	} else {
		search_id += 1
	}

	go search_timer(timer) // abort the current search after time_limit seconds.

	move, sum := iterative_deepening(brd, depth, start)
	timer.Stop() // cancel the timer to prevent it from interfering with the next search if it's not
	// garbage collected before then.

	return move, sum
}

func iterative_deepening(brd *Board, depth int, start time.Time) (Move, int) {
	var move Move
	var guess, count, first_count, sum int
	// var previous_count int
	c := brd.c

	for d := 1; d <= depth; d++ {
		move, guess, count = ybw_root(brd, -INF, INF, iid_score[c], d)
		sum += count

		if cancel_search {
			if depth > 1 {
				avg_branch := math.Pow(float64(sum)/float64(first_count), float64(1)/float64(depth-1))
				// fmt.Println("------------------------------------------------------------------")
				fmt.Printf("Average Branching: %.4f\n", avg_branch)
			}
			return iid_move[c], sum
		} else {
			// if d > 5 { // don't print info for first few plys to reduce communication traffic.
			// PrintInfo(guess, d, sum, time.Since(start))
			// 	fmt.Printf("  -Branching factor: %v\n", float64(count)/float64(previous_count))
			// }
			if d == 1 {
				first_count = count
			}
			iid_move[c], iid_score[c] = move, guess
			// previous_count = count
		}
	}

	PrintInfo(guess, depth, sum, time.Since(start))

	if depth > 1 {
		avg_branch := math.Pow(float64(sum)/float64(first_count), float64(1)/float64(depth-1))
		// fmt.Println("------------------------------------------------------------------")
		fmt.Printf("Average Branching: %.4f\n", avg_branch)
	}
	return iid_move[c], sum
}

func ybw_root(brd *Board, alpha, beta, guess, depth int) (Move, int, int) {
	if cancel_search {
		return 0, 0, 1
	}

	sum, count, legal_searched := 1, 0, 0

	in_check := is_in_check(brd)
	// if in_check {
	// 	fmt.Println("In check at root.")
	// }
	score, best, old_alpha := -INF, -INF, alpha

	var best_move Move
	// search hash move
	first_move, hash_result := main_tt.probe(brd, depth, depth-2, &alpha, &beta, &score)
	if hash_result != NO_MATCH && is_valid_move(brd, first_move, depth) &&
		avoids_check(brd, first_move, in_check) {

		legal_searched += 1

		score, count = ybw_make(brd, first_move, alpha, beta, depth-1, 1, true)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, first_move, depth, count)
					main_tt.store(brd, first_move, depth, LOWER_BOUND, score)
					return first_move, score, sum
				}
				alpha = score
			}
			best_move = first_move
			best = score
		}
	}

	// Generate tactical (non-quiet) moves.  Good moves will be searched sequentially to establish good bounds
	// before remaining nodes are searched in parallel.
	best_moves, remaining_moves := get_best_moves(brd, in_check)
	var m Move
	for _, item := range *best_moves { // search the best moves sequentially.
		m = item.move
		if m == first_move || !avoids_check(brd, m, in_check) {
			continue
		}
		legal_searched += 1
		score, count = ybw_make(brd, m, alpha, beta, depth-1, 1, true)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, m, depth, count)
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return m, score, sum
				}
				alpha = score
			}
			best_move = m
			best = score
		}
	}

	// Delay the generation of remaining moves until all promotions and winning captures have been searched.
	// if a cutoff occurs, this will reduce move generation effort substantially.
	get_remaining_moves(brd, in_check, remaining_moves)

	if depth <= SPLIT_MIN { // Depth is too shallow for parallel search to be worthwhile.
		for _, item := range *remaining_moves { // search remaining moves sequentially.
			m = item.move
			if m == first_move || !avoids_check(brd, m, in_check) {
				continue
			}
			legal_searched += 1
			score, count = ybw_make(brd, m, alpha, beta, depth-1, 1, true)
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						store_cutoff(brd, m, depth, count)
						main_tt.store(brd, m, depth, LOWER_BOUND, score)
						return m, score, sum
					}
					alpha = score
				}
				best_move = m
				best = score
			}
		}
	} else { // now that decent bounds have been established, parallel search is possible.
		// Make sure at least 3 nodes have been searched serially before spawning.
		// for ; legal_searched < 3; legal_searched++ {
		// 	item := remaining_moves.Pop()
		// 	for item != nil && !avoids_check(brd, item.move, in_check) {
		// 		item = remaining_moves.Pop()  // get the highest-sorted legal move from the remaining_moves list
		// 	}
		// 	if item == nil {
		// 		break
		// 	}
		// 	score, count = ybw_make(brd, item.move, alpha, beta, depth-1, 1, true)
		// 	sum += count
		// 	if score > best {
		// 		if score > alpha {
		// 			if score >= beta {
		// 				store_cutoff(brd, item.move, depth, count)
		// 				main_tt.store(brd, item.move, depth, LOWER_BOUND, score)
		// 				return item.move, score, sum
		// 			}
		// 			alpha = score
		// 		}
		// 		best_move = item.move
		// 		best = score
		// 	}
		// }

		result_child := make(chan SearchResult, 40)
		var child_counter int
		for _, item := range *remaining_moves {
			m := item.move
			if m == first_move || !avoids_check(brd, m, in_check) {
				continue
			}
			new_brd := brd.Copy() // create a locally scoped deep copy of the board.
			legal_searched += 1
			child_counter++
			go func() {
				score, count := ybw_make(new_brd, m, alpha, beta, depth-1, 1, true)
				result_child <- SearchResult{m, score, count}
			}()
		}

		if child_counter > 0 {
		remaining_pieces:
			for {
				select { // wait for a message to come in on one of the channels.
				case result := <-result_child: // one of the child subtrees has been completely searched.
					sum += result.count
					if result.score > best {
						if result.score > alpha {
							if result.score >= beta {
								store_cutoff(brd, result.move, depth, result.count)
								main_tt.store(brd, result.move, depth, LOWER_BOUND, result.score)
								return result.move, result.score, sum
							}
							alpha = result.score
							// for _, update_child := range listeners {
							// 	update_child <- BoundUpdate{-alpha, true} // send the updated bound to child processes.
							// }
						}
						best_move = result.move
						best = result.score
					}

					child_counter--
					if child_counter == 0 {
						break remaining_pieces // exit the for loop
					}
				}
			}
		}
	}

	if legal_searched > 0 {
		if alpha > old_alpha {
			// if best_move == 0 {
			// 	fmt.Println("No best move for root EXACT node.")
			// }
			main_tt.store(brd, best_move, depth, EXACT, best)
		} else {
			// if best_move == 0 {
			// 	fmt.Println("No best move for root UPPER_BOUND node.")
			// }
			main_tt.store(brd, best_move, depth, UPPER_BOUND, best)
		}
		return best_move, best, sum
	} else { // draw or checkmate detected.
		if in_check {
			// fmt.Printf("Checkmate detected at root %#x\n", brd.hash_key)
			main_tt.store(brd, 0, MAX_PLY, EXACT, -INF)
			return 0, -INF, sum
		} else {
			// fmt.Printf("Draw detected at root %#x\n", brd.hash_key)
			main_tt.store(brd, 0, MAX_PLY, EXACT, 0)
			return 0, 0, sum
		}
	}

}

func young_brothers_wait(brd *Board, alpha, beta, depth, ply int, can_null bool) (int, int) {
	if cancel_search {
		return 0, 0
	}

	if depth <= 0 {
		return quiescence(brd, alpha, beta, depth, ply) // q-search is always sequential.
	}

	in_check := is_in_check(brd)
	if brd.halfmove_clock >= 100 {
		if is_checkmate(brd, in_check) {
			return ply - INF, 1
		} else {
			fmt.Printf("Draw by repetition detected at ply %d\n", ply)
			return 0, 1
		}
	}

	score, best := -INF, -INF
	legal_searched := 0
	old_alpha := alpha
	sum, count := 1, 0
	var null_depth int
	if depth > 6 {
		null_depth = depth - 3
	} else {
		null_depth = depth - 2
	}

	var best_move Move
	first_move, hash_result := main_tt.probe(brd, depth, null_depth, &alpha, &beta, &score)

	if hash_result == CUTOFF_FOUND {
		return score, sum
	} else if hash_result != AVOID_NULL { // Null-Move Pruning
		if !in_check && can_null && depth > 2 && in_endgame(brd, brd.c) == 0 &&
			!pawns_only(brd, brd.c) && evaluate(brd, alpha, beta) >= beta {
			score, count = null_make(brd, beta, null_depth-1, ply+1)
			sum += count
			if score >= beta {
				main_tt.store(brd, 0, depth, LOWER_BOUND, score)
				return score, sum
			}
		}
	}

	if hash_result == NO_MATCH { // No hash move available. Use IID to get a decent first move to try.
		// implementation will depend on PVS implementation.
	}

	if is_valid_move(brd, first_move, depth) && avoids_check(brd, first_move, in_check) {
		legal_searched += 1
		score, count = ybw_make(brd, first_move, alpha, beta, depth-1, ply+1, can_null)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, first_move, depth, count)
					main_tt.store(brd, first_move, depth, LOWER_BOUND, score)
					return score, sum
				}
				alpha = score
			}
			best_move = first_move
			best = score
		}
	}

	// Generate tactical (non-quiet) moves.  Good moves will be searched sequentially to establish good bounds
	// before remaining nodes are searched in parallel.
	best_moves, remaining_moves := get_best_moves(brd, in_check)
	var m Move
	for _, item := range *best_moves { // search the best moves sequentially.
		m = item.move
		if m == first_move || !avoids_check(brd, m, in_check) {
			continue
		}
		legal_searched += 1
		score, count = ybw_make(brd, m, alpha, beta, depth-1, ply+1, can_null)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, m, depth, count)
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return score, sum
				}
				alpha = score
			}
			best_move = m
			best = score
		}
	}

	// Delay the generation of remaining moves until all promotions and winning captures have been searched.
	// if a cutoff occurs, this will reduce move generation effort substantially.
	get_remaining_moves(brd, in_check, remaining_moves)

	if depth <= SPLIT_MIN { // Depth is too shallow for parallel search to be worthwhile.
		for _, item := range *remaining_moves { // search remaining moves sequentially.
			m = item.move
			if m == first_move || !avoids_check(brd, m, in_check) {
				continue
			}
			legal_searched += 1
			score, count = ybw_make(brd, m, alpha, beta, depth-1, ply+1, can_null)
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						store_cutoff(brd, m, depth, count)
						main_tt.store(brd, m, depth, LOWER_BOUND, score)
						return score, sum
					}
					alpha = score
				}
				best_move = m
				best = score
			}
		}
	} else { // now that decent bounds have been established, parallel search is possible.
		// // Make sure at least 3 nodes have been searched serially before spawning.
		// for ; legal_searched < 3; legal_searched++ {
		// 	item := remaining_moves.Pop()
		// 	for item != nil && !avoids_check(brd, item.move, in_check) {
		// 		item = remaining_moves.Pop()  // get the highest-sorted legal move from the remaining_moves list
		// 	}
		// 	if item == nil {
		// 		break
		// 	}
		// 	score, count = ybw_make(brd, item.move, alpha, beta, depth-1, 1, true)
		// 	sum += count
		// 	if score > best {
		// 		if score > alpha {
		// 			if score >= beta {
		// 				store_cutoff(brd, item.move, depth, count)
		// 				main_tt.store(brd, item.move, depth, LOWER_BOUND, score)
		// 				return score, sum
		// 			}
		// 			alpha = score
		// 		}
		// 		best_move = item.move
		// 		best = score
		// 	}
		// }

		result_child := make(chan SearchResult, 10)
		var child_counter int
		for _, item := range *remaining_moves {
			m := item.move
			if m == first_move || !avoids_check(brd, m, in_check) {
				continue
			}
			new_brd := brd.Copy() // create a locally scoped deep copy of the board.
			legal_searched += 1
			child_counter++
			go func() {
				score, count := ybw_make(new_brd, m, alpha, beta, depth-1, ply+1, can_null)
				result_child <- SearchResult{m, score, count}
			}()
		}

		if child_counter > 0 {
		remaining_pieces:
			for {
				select { // wait for a message to come in on one of the channels.
				case result := <-result_child: // one of the child subtrees has been completely searched.
					sum += result.count
					if result.score > best {
						if result.score > alpha {
							if result.score >= beta {
								store_cutoff(brd, result.move, depth, result.count)
								main_tt.store(brd, result.move, depth, LOWER_BOUND, score)
								return result.score, sum
							}
							alpha = result.score
							// for _, update_child := range listeners {
							// 	update_child <- BoundUpdate{-alpha, true} // send the updated bound to child processes.
							// }
						}
						best_move = result.move
						best = score
					}

					child_counter--
					if child_counter == 0 {
						break remaining_pieces // exit the for loop
					}
				}
			}
		}
	}

	if legal_searched > 0 {
		if alpha > old_alpha {
			main_tt.store(brd, best_move, depth, EXACT, best)
		} else {
			main_tt.store(brd, best_move, depth, UPPER_BOUND, best)
		}
		return best, sum
	} else { // draw or checkmate detected.
		if in_check {
			main_tt.store(brd, 0, depth, EXACT, ply-INF)
			return ply - INF, sum
		} else {
			main_tt.store(brd, 0, depth, EXACT, 0)
			return 0, sum
		}
	}
}

// Q-Search will always be done sequentially.
// Q-search subtrees are taller and narrower than in the main search making benefit of parallelism
// smaller and raising communication and synchronization overhead.
func quiescence(brd *Board, alpha, beta, depth, ply int) (int, int) {

	sum, count := 1, 0

	in_check := is_in_check(brd)
	score, best := -INF, -INF

	if brd.halfmove_clock >= 100 {
		if is_checkmate(brd, in_check) {
			return ply - INF, 1
		} else {
			fmt.Printf("Draw by repetition detected at ply %d\n", ply)
			return 0, 1
		}
	}

	legal_moves := false
	var m Move
	if in_check {
		best_moves, remaining_moves := &MoveList{}, &MoveList{}
		get_evasions(brd, best_moves, remaining_moves) // only legal moves generated here.
		for _, item := range *best_moves {
			m = item.move
			legal_moves = true
			score, count = q_make(brd, m, alpha, beta, depth-1, ply+1)
			sum += count
			if score > best {
				best = score
				if score > alpha {
					if score >= beta {
						return score, sum
					}
					alpha = score
				}
			}
		}
		for _, item := range *remaining_moves {
			m = item.move
			legal_moves = true
			score, count = q_make(brd, m, alpha, beta, depth-1, ply+1)
			sum += count
			if score > best {
				best = score
				if score > alpha {
					if score >= beta {
						return score, sum
					}
					alpha = score
				}
			}
		}
		if !legal_moves {
			return ply - INF, 1 // detect checkmate.
		}
	} else {

		score = evaluate(brd, alpha, beta) // stand pat
		if score > best {
			if score > alpha {
				if score >= beta {
					return score, sum
				}
				alpha = score
			}
			best = score
		}

		best_moves := get_winning_captures(brd)
		for _, item := range *best_moves { // search the best moves sequentially.
			m = item.move
			if !avoids_check(brd, m, in_check) {
				continue // prune illegal moves
			}
			if best+m.CapturedPiece().Value()+m.PromotedTo().PromoteValue()+piece_values[ROOK] < alpha {
				continue // prune futile moves with no chance of raising alpha.
			}

			score, count = q_make(brd, m, alpha, beta, depth-1, ply+1)
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						return score, sum
					}
					alpha = score
				}
				best = score
			}
		}
	}

	return best, sum

}

func ybw_make(brd *Board, m Move, alpha, beta, depth, ply int, can_null bool) (int, int) {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	// if !is_valid_move(brd, m, depth) {
	// 	fmt.Printf("Warning: invalid move made.\n")
	// }

	make_move(brd, m) // to do: make move
	score, sum := young_brothers_wait(brd, -beta, -alpha, depth, ply, can_null)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return -score, sum
}

// func ybw_parallell_make(brd *Board, m Move, alpha, beta, depth, ply int) (int, int) {
// 	make_move(brd, m) // to do: make move
// 	score, sum := young_brothers_wait(brd, -beta, -alpha, depth, ply)
// 	return -score, sum
// }

func q_make(brd *Board, m Move, alpha, beta, depth, ply int) (int, int) {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score, sum := quiescence(brd, -beta, -alpha, depth, ply)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return -score, sum
}

func null_make(brd *Board, beta, depth, ply int) (int, int) {
	hash_key, enp_target := brd.hash_key, brd.enp_target
	brd.c ^= 1
	brd.hash_key ^= side_key
	brd.hash_key ^= enp_zobrist(enp_target)
	brd.enp_target = SQ_INVALID

	score, sum := young_brothers_wait(brd, -beta, -beta+1, depth, ply, false)

	brd.c ^= 1
	brd.hash_key = hash_key
	brd.enp_target = enp_target
	return -score, sum
}

func store_cutoff(brd *Board, m Move, depth, count int) {
	if !m.IsCapture() {
		main_htable.Store(m, brd.c, count)
	}

	// store killer move in killer list for this Goroutine.

}

// what makes this so slow?  Is it just the context switching between threads?
