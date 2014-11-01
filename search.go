//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------
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
	// "math"
	"time"
)

const (
	MAX_TIME    = 120000 // default search time limit in seconds (2m)
	MAX_DEPTH   = 12
	MAX_EXT     = 12
	SPLIT_MIN   = 13 // set > MAX_DEPTH to disable parallel search.
	F_PRUNE_MAX = 3  // should always be less than SPLIT_MIN
	LMR_MIN     = 2
	MAX_PLY     = MAX_DEPTH + MAX_EXT
	IID_MIN     = 4
	MAX_Q_CHECKS = 2
	COMMS_MIN   = 6 // minimum depth at which to send info to GUI.
)

const (
	D_PV = iota
	D_CUT
	D_ALL
)

type SearchResult struct {
	move  Move
	score int
	count int
	pv    *PV
}

type BoundUpdate struct {
	bound        int
	alpha_update bool
}

var side_to_move uint8
var search_id int
var cancel_search bool
var uci_mode bool = false
var uci_ponder bool = false
var print_info bool = true

var nodes_per_iteration [MAX_DEPTH + 1]int

func AbortSearch() {
	cancel_search = true
	if print_info {
		fmt.Println("Search aborted by GUI")
	}
}

func search_timer(timer *time.Timer) {
	select {
	case <-timer.C:
		AbortSearch()
	}
}

func Search(brd *Board, reps *RepList, depth, time_limit int) (Move, int) {
	cancel_search = false
	side_to_move = brd.c
	id_move[brd.c] = 0
	start := time.Now()
	timer := time.NewTimer(time.Duration(time_limit) * time.Millisecond)

	if search_id >= 512 { // only 9 bits are available to store the id in each TT entry.
		search_id = 1
	} else {
		search_id += 1
	}

	go search_timer(timer) // abort the current search after time_limit seconds.

	move, sum := iterative_deepening(brd, reps, depth, start)
	timer.Stop() // cancel the timer to prevent it from interfering with the next search if it's not
	// garbage collected before then.
	return move, sum
}

var id_move [2]Move
var id_score [2]int
var id_alpha, id_beta int

func iterative_deepening(brd *Board, reps *RepList, depth int, start time.Time) (Move, int) {
	var guess, count, sum int
	var current_pv *PV
	c := brd.c

	id_alpha, id_beta = -INF, INF // first iteration is always full-width.

	guess, count, current_pv = young_brothers_wait(brd, id_alpha, id_beta, 1, 0, MAX_EXT, true, reps)

	nodes_per_iteration[1] += count
	sum += count
	id_move[c], id_score[c] = current_pv.m, guess

	var d int
	for d = 2; d <= depth; d++ {

		// to do: add aspiration windows

		guess, count, current_pv = young_brothers_wait(brd, id_alpha, id_beta, d, 0, MAX_EXT, true, reps)

		if cancel_search {
			return id_move[c], sum
		}

		nodes_per_iteration[d] += count
		sum += count
		if d > COMMS_MIN && print_info && uci_mode { // don't print info for first few plys to reduce communication traffic.
			PrintInfo(guess, d, sum, time.Since(start), current_pv)
		}

		if current_pv == nil || current_pv.m == 0 {
			fmt.Printf("No root PV returned after depth %d.\n", d)
		} else {
			id_move[c], id_score[c] = current_pv.m, guess
		}
	}
	if print_info {
		PrintInfo(guess, depth, sum, time.Since(start), current_pv)
	}

	return id_move[c], sum
}

func young_brothers_wait(brd *Board, alpha, beta, depth, ply, extensions_left int, can_null bool, old_reps *RepList) (int, int, *PV) {

	if depth <= 0 {
		score, sum := quiescence(brd, alpha, beta, depth, ply, MAX_Q_CHECKS, old_reps) // q-search is always sequential.
		return score, sum, nil
	}

	if cancel_search {
		return 0, 0, nil
	}

	if old_reps.Scan(brd.hash_key) {
		return 0, 1, nil
	}

	in_check := is_in_check(brd)
	if in_check && extensions_left > 0 {
		if MAX_EXT > extensions_left { // only extend after the first check.
			depth += 1
		}
		extensions_left -= 1
	}

	if brd.halfmove_clock >= 100 {
		if is_checkmate(brd, in_check) {
			return ply - INF, 1, nil
		} else {
			return 0, 1, nil
		}
	}

	score, best := -INF, -INF
	old_alpha := alpha
	sum, count, legal_searched := 1, 0, 0
	pv := &PV{}
	reps := &RepList{uint32(brd.hash_key), old_reps}
	var next_pv *PV
	var best_move, first_move Move
	var null_depth int
	if depth > 6 {
		null_depth = depth - 3
	} else {
		null_depth = depth - 2
	}
	var hash_result int
	first_move, hash_result = main_tt.probe(brd, depth, null_depth, &alpha, &beta, &score)

	if ply > 0 {
		if hash_result == CUTOFF_FOUND {
			return score, sum, nil
		} else if hash_result != AVOID_NULL {
			// Null-Move Pruning
			if !in_check && can_null && depth > 2 && in_endgame(brd) == 0 &&
				!pawns_only(brd, brd.c) && evaluate(brd, alpha, beta) >= beta {
				score, count = null_make(brd, beta, null_depth-1, ply+1, extensions_left, reps)
				sum += count
				if score >= beta {
					main_tt.store(brd, 0, depth, LOWER_BOUND, score)
					return score, sum, nil
				}
			}
		}
	}

	// Determine expected node type.
	node_type := D_ALL
	var local_id_alpha, local_id_beta int
	if (ply & 1) == 1 { // odd-ply
		local_id_alpha, local_id_beta = -id_beta, -id_alpha
		if beta == local_id_beta {
			node_type = D_CUT
		}
	} else { // even-ply
		local_id_alpha, local_id_beta = id_alpha, id_beta
		if alpha == local_id_alpha {
			node_type = D_CUT
		}
	}
	if alpha == local_id_alpha && beta == local_id_beta {
		node_type = D_PV
	}
	// No hash move available. If this is a PV node, use IID to get a decent first move to try.
	if hash_result == NO_MATCH && can_null && depth >= IID_MIN { //&& node_type != D_ALL {
		var local_pv *PV
		score, count, local_pv = young_brothers_wait(brd, alpha, beta, depth-2, ply, extensions_left, can_null, old_reps)
		sum += count
		if local_pv != nil {
			first_move = local_pv.m
		}
	}

	// If a hash move or IID move is available, try it first.
	if first_move.IsValid(brd) && avoids_check(brd, first_move, in_check) {
		legal_searched += 1
		score, count, next_pv = ybw_make(brd, first_move, alpha, beta, depth-1, ply+1, extensions_left, can_null, reps)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, first_move, depth, ply, count)
					main_tt.store(brd, first_move, depth, LOWER_BOUND, score)
					return score, sum, nil
				}
				alpha = score
				pv.m = first_move
				pv.next = next_pv
			}
			best_move = first_move
			best = score
		}
	}

	// Generate tactical (non-quiet) moves.  Good moves will be searched sequentially to establish good bounds
	// before remaining nodes are searched in parallel.
	best_moves, remaining_moves := get_best_moves(brd, in_check, &main_ktable[ply])
	var m Move
	for _, item := range *best_moves { // search the best moves sequentially.
		m = item.move
		if m == first_move || !avoids_check(brd, m, in_check) {
			continue
		}

		// if alpha > old_alpha && node_type != D_PV {
		// 	score, count, next_pv = ybw_make(brd, m, alpha, alpha+1, depth-1, ply+1, extensions_left, can_null, reps)
		// 	sum += count
		// 	if alpha < score && score < beta {
		// 		score, count, next_pv = ybw_make(brd, m, alpha, beta, depth-1, ply+1, extensions_left, can_null, reps)
		// 		sum += count
		// 	}
		// } else {
		if m.IsPromotion() && extensions_left > 0 {
			score, count, next_pv = ybw_make(brd, m, alpha, beta, depth, ply+1, extensions_left-1, can_null, reps)
		} else {
			score, count, next_pv = ybw_make(brd, m, alpha, beta, depth-1, ply+1, extensions_left, can_null, reps)
		}

		sum += count
		// }

		legal_searched += 1
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, m, depth, ply, count)
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return score, sum, nil
				}
				alpha = score
				pv.m = m
				pv.next = next_pv
			}
			best_move = m
			best = score
		}
	}

	// Delay the generation of remaining moves until all promotions, winning captures, and killer moves have been searched.
	// if a cutoff occurs, this will reduce move generation effort substantially.
	get_remaining_moves(brd, in_check, remaining_moves, &main_ktable[ply])

	// if depth <= SPLIT_MIN { // Depth is too shallow for parallel search to be worthwhile.

	f_prune, can_reduce := false, false
	if !in_check && ply > 0 && node_type != D_PV && alpha > 100-INF {
		if depth <= F_PRUNE_MAX && evaluate(brd, alpha, beta)+piece_values[BISHOP] < alpha {
			f_prune = true
		}
		if depth >= LMR_MIN {
			can_reduce = true
		}
	}

	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock
	var r_depth int
	for _, item := range *remaining_moves { // search remaining moves sequentially.
		m = item.move
		if m == first_move || !avoids_check(brd, m, in_check) {
			continue
		}

		make_move(brd, m)

		if f_prune && legal_searched > 0 && m.IsQuiet() && !is_passed_pawn(brd, m) && !is_in_check(brd) {
			unmake_move(brd, m, enp_target)
			brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
			brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
			continue
		}

		// Late move reductions:
		r_depth = depth
		if can_reduce && item.order == 0 && !is_passed_pawn(brd, m) && !is_in_check(brd) {
			r_depth = depth - 1
		}

		// if alpha > old_alpha && node_type != D_PV {
		// 	score, count, next_pv = young_brothers_wait(brd, -alpha-1, -alpha, r_depth-1, ply+1, extensions_left, can_null, reps)
		// 	sum += count
		// 	if alpha < -score && -score < beta {
		// 		score, count, next_pv = young_brothers_wait(brd, -beta, -alpha, r_depth-1, ply+1, extensions_left, can_null, reps)
		// 		sum += count
		// 	}
		// } else {

		if m.IsPromotion() && extensions_left > 0 {
			score, count, next_pv = young_brothers_wait(brd, -beta, -alpha, r_depth, ply+1, extensions_left-1, can_null, reps)
		} else {
			score, count, next_pv = young_brothers_wait(brd, -beta, -alpha, r_depth-1, ply+1, extensions_left, can_null, reps)
		}

		sum += count
		// }

		score = -score
		unmake_move(brd, m, enp_target) // to do: unmake move
		brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
		brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock

		legal_searched += 1
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, m, depth, ply, count) // what happens on refutation of main pv?
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return score, sum, nil
				}
				alpha = score
				pv.m = m
				pv.next = next_pv
			}
			best_move = m
			best = score
		}
	}
	// } else { // now that decent bounds have been established, parallel search is possible.
	// 	// Make sure at least 3 nodes have been searched serially before spawning.
	// 	for ; legal_searched < 3; legal_searched++ {
	// 		item := remaining_moves.Dequeue()
	// 		for item != nil && (item.move == first_move || !avoids_check(brd, item.move, in_check)) {
	// 			item = remaining_moves.Dequeue() // get the highest-sorted legal move from remaining_moves
	// 		}
	// 		if item == nil {
	// 			break
	// 		}

	// 		score, count, next_pv = ybw_make(brd, item.move, alpha, beta, depth-1, 1, extensions_left, true, reps)
	// 		sum += count
	// 		if score > best {
	// 			if score > alpha {
	// 				if score >= beta {
	// 					store_cutoff(brd, item.move, depth, ply, count)
	// 					main_tt.store(brd, item.move, depth, LOWER_BOUND, score)
	// 					return score, sum, nil
	// 				}
	// 				alpha = score
	// 				pv.m = m
	// 				pv.next = next_pv
	// 			}
	// 			best_move = item.move
	// 			best = score
	// 		}
	// 	}

	// 	result_child := make(chan SearchResult, 10)
	// 	var child_counter int
	// 	for _, item := range *remaining_moves {
	// 		m := item.move
	// 		if m == first_move || !avoids_check(brd, m, in_check) {
	// 			continue
	// 		}
	// 		new_brd := brd.Copy() // create a locally scoped deep copy of the board.
	// 		legal_searched += 1
	// 		child_counter++
	// 		go func() {
	// 			score, count, next_pv := ybw_make(new_brd, m, alpha, beta, depth-1, ply+1, extensions_left, can_null, reps)
	// 			result_child <- SearchResult{m, score, count, next_pv}
	// 		}()
	// 	}

	// 	if child_counter > 0 {
	// 	remaining_pieces:
	// 		for {
	// 			select { // wait for a message to come in on one of the channels.
	// 			case result := <-result_child: // one of the child subtrees has been completely searched.
	// 				sum += result.count
	// 				if result.score > best {
	// 					if result.score > alpha {
	// 						if result.score >= beta {
	// 							store_cutoff(brd, result.move, depth, ply, result.count)
	// 							main_tt.store(brd, result.move, depth, LOWER_BOUND, score)
	// 							return result.score, sum, nil
	// 						}
	// 						alpha = result.score
	// 						pv.m = result.move
	// 						pv.next = result.pv
	// 					}
	// 					best_move = result.move
	// 					best = score
	// 				}
	// 				child_counter--
	// 				if child_counter == 0 {
	// 					break remaining_pieces // exit the for loop
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	if legal_searched > 0 {
		if alpha > old_alpha {
			main_tt.store(brd, best_move, depth, EXACT, best)
			return best, sum, pv
		} else {
			main_tt.store(brd, best_move, depth, UPPER_BOUND, best)
			return best, sum, nil
		}
	} else { // draw or checkmate detected.
		if in_check {
			main_tt.store(brd, 0, depth, EXACT, ply-INF)
			return ply - INF, sum, nil
		} else {
			main_tt.store(brd, 0, depth, EXACT, 0)
			return 0, sum, nil
		}
	}
}

// Q-Search will always be done sequentially: Q-search subtrees are taller and narrower than in the main search,
// making benefit of parallelism smaller and raising communication and synchronization overhead.
func quiescence(brd *Board, alpha, beta, depth, ply, checks_remaining int, old_reps *RepList) (int, int) {
	if cancel_search {
		return 0, 0
	}

	if old_reps.Scan(brd.hash_key) {
		return 0, 1
	}

	in_check := is_in_check(brd)
	if brd.halfmove_clock >= 100 {
		if is_checkmate(brd, in_check) {
			return ply - INF, 1
		} else {
			return 0, 1
		}
	}

	score, best := -INF, -INF
	sum, count := 1, 0
	reps := &RepList{uint32(brd.hash_key), old_reps}
	legal_moves := false

	var m Move
	if in_check {
		checks_remaining -= 1
		best_moves, remaining_moves := &MoveList{}, &MoveList{}
		get_evasions(brd, best_moves, remaining_moves, &main_ktable[ply]) // only legal moves generated here.
		best_moves.Sort()
		for _, item := range *best_moves {
			m = item.move
			legal_moves = true
			score, count = q_make(brd, m, alpha, beta, depth-1, ply+1, checks_remaining, reps)
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
		remaining_moves.Sort()
		for _, item := range *remaining_moves {
			m = item.move
			legal_moves = true
			score, count = q_make(brd, m, alpha, beta, depth-1, ply+1, checks_remaining, reps)
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

		hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
		castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock
		best_moves := get_winning_captures(brd)
		for _, item := range *best_moves { // search the best moves sequentially.
			m = item.move
			if !avoids_check(brd, m, in_check) {
				continue // prune illegal moves
			}

			make_move(brd, m) // to do: make move
			if alpha > 100-INF &&
				best+m.CapturedPiece().Value()+m.PromotedTo().PromoteValue()+piece_values[ROOK] < alpha &&
				!is_in_check(brd) {
				unmake_move(brd, m, enp_target) // to do: unmake move
				brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
				brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
				continue
			}
			score, count := quiescence(brd, -beta, -alpha, depth, ply, checks_remaining, reps)
			unmake_move(brd, m, enp_target) // to do: unmake move
			brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
			brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock

			score = -score
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

		if checks_remaining > 0 {
			checking_moves := get_checks(brd, &main_ktable[ply])
			for _, item := range *checking_moves {
				m = item.move
				m.IsValid(brd)
				score, count = q_make(brd, m, alpha, beta, depth-1, ply+1, checks_remaining, reps)
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

	}

	return best, sum
}

func ybw_make(brd *Board, m Move, alpha, beta, depth, ply, extensions_left int, can_null bool, reps *RepList) (int, int, *PV) {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score, sum, pv := young_brothers_wait(brd, -beta, -alpha, depth, ply, extensions_left, can_null, reps)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return -score, sum, pv
}

func q_make(brd *Board, m Move, alpha, beta, depth, ply, checks_remaining int, reps *RepList) (int, int) {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score, sum := quiescence(brd, -beta, -alpha, depth, ply, checks_remaining, reps)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return -score, sum
}

func null_make(brd *Board, beta, depth, ply, extensions_left int, reps *RepList) (int, int) {
	hash_key, enp_target := brd.hash_key, brd.enp_target
	brd.c ^= 1
	brd.hash_key ^= side_key
	brd.hash_key ^= enp_zobrist(enp_target)
	brd.enp_target = SQ_INVALID

	score, sum, _ := young_brothers_wait(brd, -beta, -beta+1, depth, ply, extensions_left, false, reps)

	brd.c ^= 1
	brd.hash_key = hash_key
	brd.enp_target = enp_target
	return -score, sum
}

func store_cutoff(brd *Board, m Move, depth, ply, count int) {
	if !m.IsCapture() {
		main_htable.Store(m, brd.c, count)
		if !m.IsPromotion() { // By the time killer moves are tried, any promotions will already have been searched.
			main_ktable.Store(m, ply) // store killer move in killer list for this Goroutine.
		}
	}
}
