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

import (
	"fmt"
	"time"
)

const (
	MAX_TIME     = 120000 // default search time limit in milliseconds (2m)
	MAX_DEPTH    = 12
	MAX_EXT      = 12
	SPLIT_MIN    = 13 // set > MAX_DEPTH to disable parallel search.
	F_PRUNE_MAX  = 3  // should always be less than SPLIT_MIN
	LMR_MIN      = 2
	MAX_PLY      = MAX_DEPTH + MAX_EXT
	IID_MIN      = 4
	MAX_Q_CHECKS = 2
	COMMS_MIN    = 6 // minimum depth at which to send info to GUI.
)

const (
	Y_PV = iota
	Y_CUT
	Y_ALL
)

var side_to_move uint8
var search_id int
var cancel_search bool
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

func Search(brd *Board, depth, time_limit int) (Move, int) {
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

	move, sum := iterative_deepening(brd, depth, start)
	timer.Stop() // cancel the timer to prevent it from interfering with the next search if it's not
	// garbage collected before then.
	return move, sum
}

var id_move [2]Move
var id_score [2]int
var id_alpha, id_beta int

const (
	STEP_SIZE = 8
)

func iterative_deepening(brd *Board, depth int, start time.Time) (Move, int) {
	var guess, count, sum int
	c := brd.c

	id_alpha, id_beta = -INF, INF // first iteration is always full-width.
	stk := make(Stack, MAX_STACK, MAX_STACK)

	guess, count = ybw(brd, stk, id_alpha, id_beta, 1, 0, MAX_EXT, true, true, Y_PV)
	nodes_per_iteration[1] += count
	sum += count

	if stk[0].pv_move != 0 {
		id_move[c], id_score[c] = stk[0].pv_move, guess
		save_pv(brd, stk, 1) // install PV to transposition table prior to next iteration.
	}

	var d int
	for d = 2; d <= depth; {

		guess, count = ybw(brd, stk, id_alpha, id_beta, d, 0, MAX_EXT, true, true, Y_PV)
		sum += count

		if cancel_search {
			return id_move[c], sum
		} else if stk[0].pv_move != 0 {
			id_move[c], id_score[c] = stk[0].pv_move, guess
			save_pv(brd, stk, 1) // install PV to transposition table prior to next iteration.
		} else {
			fmt.Printf("Nil PV returned to ID\n")
		}

		nodes_per_iteration[d] += count
		if d > COMMS_MIN && print_info && uci_mode { // don't print info for first few plys to reduce communication traffic.
			PrintInfo(guess, d, sum, time.Since(start), stk)
		}

		d++
	}

	if print_info {
		PrintInfo(guess, depth, sum, time.Since(start), stk)
	}

	return id_move[c], sum
}

func ybw(brd *Board, stk Stack, alpha, beta, depth, ply, extensions_left int, can_null, can_split bool, node_type int) (int, int) {

	// always extend 1-ply when at end of PV?

	if depth <= 0 {
		score, sum := quiescence(brd, stk, alpha, beta, depth, ply, MAX_Q_CHECKS) // q-search is always sequential.
		return score, sum
	}

	if cancel_search {
		return 0, 0
	}

	// if old_reps.Scan(brd.hash_key) {
	// 	return 0, 1, nil
	// }

	in_check := is_in_check(brd)
	if in_check && extensions_left > 0 {
		if MAX_EXT > extensions_left { // only extend after the first check.
			depth += 1
		}
		extensions_left -= 1
	}

	if brd.halfmove_clock >= 100 {
		if is_checkmate(brd, in_check) {
			return ply - MATE, 1
		} else {
			return 0, 1
		}
	}

	score, best := -INF, -INF
	old_alpha := alpha
	sum, count, legal_searched := 1, 0, 0
	this_stk := stk[ply]


	var best_move, first_move Move
	var null_depth int
	if depth > 6 {
		null_depth = depth - 3
	} else {
		null_depth = depth - 2
	}
	var hash_result int
	first_move, hash_result = main_tt.probe(brd, depth, null_depth, &alpha, &beta, &score)

	if node_type != Y_PV {

		if hash_result == CUTOFF_FOUND { // Hash hit
			// pv.m, pv.value = first_move, score
			this_stk.pv_move, this_stk.value = first_move, score
			return score, sum

		} else if hash_result != AVOID_NULL { // Null-Move Pruning

			if !in_check && can_null && depth > 2 && in_endgame(brd) == 0 &&
				!brd.pawns_only() && evaluate(brd, alpha, beta) >= beta {
				score, count = null_make(brd, stk, beta, null_depth-1, ply+1, extensions_left, can_split)
				sum += count
				if score >= beta {
					main_tt.store(brd, 0, depth, LOWER_BOUND, score)
					return score, sum
				}
			}
		}
	}

	// IID
	if hash_result == NO_MATCH && can_null && depth >= IID_MIN { // skip IID when in check?
		// No hash move available. Use IID to get a decent first move to try.
		score, count = ybw(brd, stk, alpha, beta, depth-2, ply, extensions_left, can_null, false, node_type)
		sum += count
		first_move = this_stk.pv_move
	}



	var child_type, r_depth, r_extensions int
	var m Move
	this_stk.eval = evaluate(brd, alpha, beta)

	// Restrict pruning to STAGE_REMAINING
	f_prune, can_reduce := false, false
	if !in_check && ply > 0 && node_type != Y_PV && alpha > 100-MATE {
		if depth <= F_PRUNE_MAX && this_stk.eval+piece_values[BISHOP] < alpha {
			f_prune = true
		}
		if depth >= LMR_MIN {
			can_reduce = true
		}
	}

	memento := brd.NewMemento()
	generator := NewMoveSelector(brd, &this_stk, in_check, first_move)

	for m = generator.next(); m != NO_MOVE; m = generator.next() {

		make_move(brd, m)

		if f_prune && generator.stage >= STAGE_REMAINING && legal_searched > 0 && m.IsQuiet() &&
			!is_passed_pawn(brd, m) && !is_in_check(brd) {
			unmake_move(brd, m, &memento)
			continue
		}

		child_type = determine_child_type(node_type, legal_searched)

		r_depth, r_extensions = depth, extensions_left
		if m.IsPromotion() && extensions_left > 0 {
			r_depth = depth + 1
			r_extensions = extensions_left - 1
		} else if can_reduce && generator.stage >= STAGE_REMAINING && !is_passed_pawn(brd, m) && !is_in_check(brd) {
			r_depth = depth - 1 // Late move reductions
		}

		if node_type == Y_PV && alpha > old_alpha {
			score, count = ybw(brd, stk, -alpha-1, -alpha, r_depth-1, ply+1, r_extensions, can_null, can_split, child_type)
			score = -score
			sum += count
			if score > alpha {
				score, count = ybw(brd, stk, -beta, -alpha, r_depth-1, ply+1, r_extensions, can_null, can_split, Y_ALL)
				score = -score
				sum += count
			}
		} else {
			score, count = ybw(brd, stk, -beta, -alpha, r_depth-1, ply+1, r_extensions, can_null, can_split, child_type)
			sum += count
			score = -score
		}
		legal_searched += 1

		unmake_move(brd, m, &memento) // to do: unmake move

		if score > best {
			if score > alpha {
				if score >= beta {
					this_stk.pv_move, this_stk.value = m, score
					store_cutoff(brd, &this_stk, m, depth, ply, count) // what happens on refutation of main pv?
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return score, sum
				}
				alpha = score
			}
			best_move = m
			best = score
		}
	}

	// may want to look at how depth is being stored: is it accounting for non-check extensions/reductions?

	// at split nodes the legal_searched counter will need to be shared via the SP struct.

	if legal_searched > 0 {
		this_stk.pv_move, this_stk.value = best_move, best
		if alpha > old_alpha {
			main_tt.store(brd, best_move, depth, EXACT, best) // local PV node found.
			return best, sum
		} else {
			stk[ply+1].pv_move = 0
			main_tt.store(brd, best_move, depth, UPPER_BOUND, best)
			return best, sum
		}
	} else { // draw or checkmate detected.
		// stk[ply+1].pv_move = 0  // <- is this needed?
		if in_check {
			main_tt.store(brd, 0, depth, EXACT, ply-MATE)
			return ply-MATE, sum
		} else {
			main_tt.store(brd, 0, depth, EXACT, 0)
			return 0, sum
		}
	}
}

// Q-Search will always be done sequentially: Q-search subtrees are taller and narrower than in the main search,
// making benefit of parallelism smaller and raising communication and synchronization overhead.
func quiescence(brd *Board, stk Stack, alpha, beta, depth, ply, checks_remaining int) (int, int) {
	if cancel_search {
		return 0, 0
	}

	// if old_reps.Scan(brd.hash_key) {
	// 	return 0, 1
	// }
	this_stk := stk[ply]

	in_check := is_in_check(brd)
	if brd.halfmove_clock >= 100 {
		if is_checkmate(brd, in_check) {
			return ply - MATE, 1
		} else {
			return 0, 1
		}
	}

	score, best, sum, count := -INF, -INF, 1, 0
	legal_moves := false
	memento := brd.NewMemento()

	var m Move
	if in_check {
		checks_remaining -= 1
		best_moves, remaining_moves := &MoveList{}, &MoveList{}
		get_evasions(brd, best_moves, remaining_moves, &this_stk.killers) // only legal moves generated here.
		best_moves.Sort()
		for _, item := range *best_moves {
			m = item.move
			legal_moves = true
			score, count = q_make(brd, stk, m, alpha, beta, depth-1, ply+1, checks_remaining, &memento)
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						this_stk.pv_move, this_stk.value = m, score
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
			score, count = q_make(brd, stk, m, alpha, beta, depth-1, ply+1, checks_remaining, &memento)
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						this_stk.pv_move, this_stk.value = m, score
						return score, sum
					}
					alpha = score
				}
				best = score
			}
		}
		if !legal_moves {
			return ply-MATE, 1 // detect checkmate.
		}
	} else {
		score = evaluate(brd, alpha, beta) // stand pat
		this_stk.eval = score

		if score > best {
			if score > alpha {
				if score >= beta {
					// stk[ply+1].pv_move = 0
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

			make_move(brd, m) // to do: make move
			if alpha > 100-MATE &&
				best+m.CapturedPiece().Value()+m.PromotedTo().PromoteValue()+piece_values[ROOK] < alpha &&
				!is_in_check(brd) {
				unmake_move(brd, m, &memento) // to do: unmake move
				continue
			}
			score, count := quiescence(brd, stk, -beta, -alpha, depth, ply, checks_remaining)
			unmake_move(brd, m, &memento) // to do: unmake move

			score = -score
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						this_stk.pv_move, this_stk.value = m, score
						return score, sum
					}
					alpha = score
				}
				best = score
			}
		}

		if checks_remaining > 0 {
			checking_moves := get_checks(brd, &this_stk.killers)
			for _, item := range *checking_moves {
				m = item.move
				if !avoids_check(brd, m, false) {
					continue
				}
				score, count = q_make(brd, stk, m, alpha, beta, depth-1, ply+1, checks_remaining, &memento)
				sum += count
				if score > best {
					if score > alpha {
						if score >= beta {
							this_stk.pv_move, this_stk.value = m, score
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

func determine_child_type(node_type, legal_searched int) int {
	switch node_type {
	case Y_PV:
		if legal_searched == 0 {
			return Y_PV
		} else {
			return Y_CUT
		}
	case Y_CUT:
		if legal_searched == 0 {
			return Y_ALL
		} else {
			return Y_CUT
		}
	case Y_ALL:
		return Y_CUT
	default:
		fmt.Println("Invalid node type detected.")
		return node_type
	}
}



func q_make(brd *Board, stk Stack, m Move, alpha, beta, depth, ply, checks_remaining int, memento *BoardMemento) (int, int) {
	make_move(brd, m) // to do: make move
	score, sum := quiescence(brd, stk, -beta, -alpha, depth, ply, checks_remaining)
	unmake_move(brd, m, memento) // to do: unmake move
	return -score, sum
}

func null_make(brd *Board, stk Stack, beta, depth, ply, extensions_left int, can_split bool) (int, int) {
	hash_key, enp_target := brd.hash_key, brd.enp_target
	brd.c ^= 1
	brd.hash_key ^= side_key
	brd.hash_key ^= enp_zobrist(enp_target)
	brd.enp_target = SQ_INVALID

	score, sum := ybw(brd, stk, -beta, -beta+1, depth, ply, extensions_left, false, can_split, Y_CUT)

	brd.c ^= 1
	brd.hash_key = hash_key
	brd.enp_target = enp_target
	return -score, sum
}

func store_cutoff(brd *Board, this_stk *StackItem, m Move, depth, ply, count int) {
	if !m.IsCapture() {
		main_htable.Store(m, brd.c, count)
		if !m.IsPromotion() { // By the time killer moves are tried, any promotions will already have been searched.
			
			store_killers(this_stk, m) // store killer moves in stack for this Goroutine.
		}
	}
}





