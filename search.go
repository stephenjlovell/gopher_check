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
	MAX_DEPTH    = 16
	MAX_EXT      = 16
	MAX_PLY			 = MAX_DEPTH + MAX_EXT
	SPLIT_MIN    = 2 // set > MAX_DEPTH to disable parallel search.

	F_PRUNE_MAX  = 3  // should always be less than SPLIT_MIN
	LMR_MIN      = 2
	IID_MIN      = 4
	
	MAX_Q_CHECKS = 2 
	COMMS_MIN    = 6 // minimum depth at which to send info to GUI.
)

const (
	Y_CUT = iota
	Y_ALL
	Y_PV
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

	load_balancer = NewLoadBalancer()
	load_balancer.Start() 
	brd.worker = load_balancer.RootWorker() // Send SPs generated by root goroutine to root worker.

	move, sum := iterative_deepening(brd, depth, start)
	timer.Stop() // cancel the timer to prevent it from interfering with the next search if it's not
	// garbage collected before then.
	return move, sum
}

var id_move [2]Move
var id_score [2]int
var id_alpha, id_beta int

func iterative_deepening(brd *Board, depth int, start time.Time) (Move, int) {
	var guess, total, sum int
	c := brd.c
	var stk Stack

	id_alpha, id_beta = -INF, INF // first iteration is always full-width.
	in_check := is_in_check(brd)

	for d := 1; d <= depth; d++ {
		if cancel_search {
			return id_move[c], sum
		} 

		stk = NewStack()
		stk[0].in_check = in_check
		guess, total = ybw(brd, stk, id_alpha, id_beta, d, 0, MAX_EXT, true, Y_PV, SP_NONE)
		sum += total

		if stk[0].pv_move.IsMove() {
			id_move[c], id_score[c] = stk[0].pv_move, guess
			stk.SavePV(brd, d) // install PV to transposition table prior to next iteration.
		} else {
			fmt.Printf("Nil PV returned to ID\n")
		}

		nodes_per_iteration[d] += total
		// if d > COMMS_MIN && print_info && uci_mode { // don't print info for first few plys to reduce communication traffic.
			fmt.Printf("\n")
			PrintInfo(guess, d, sum, time.Since(start), stk)
		// }

	}

	if print_info {
		PrintInfo(guess, depth, sum, time.Since(start), stk)
	}

	return id_move[c], sum
}

func ybw(brd *Board, stk Stack, alpha, beta, depth, ply, extensions_left int, can_null bool, node_type, sp_type int) (int, int) {
	var this_stk *StackItem
	var in_check bool
	var sp *SplitPoint
	var selector *MoveSelector

	score, best, old_alpha := -INF, -INF, alpha
	sum := 1

	var best_move, first_move Move
	var null_depth, hash_result, eval, subtotal, total, legal_searched, child_type, r_depth, r_extensions int

	f_prune, can_reduce := false, false

	// if the is_sp flag is set, a worker has just been assigned to this split point.
	// the SP master has already handled most of the pruning, so just read the latest values
	// from the SP and jump to the moves loop.
	if sp_type == SP_SERVANT {
		stk[ply].sp.Lock()
		sp 				= stk[ply].sp
		this_stk 	= sp.this_stk
		eval 		 	= sp.this_stk.eval
		in_check 	= this_stk.in_check
		best 			= sp.best
		selector 	= sp.selector
		stk[ply].sp.Unlock()
		goto search_moves
	}

	if depth <= 0 {
		if node_type == Y_PV {
			stk[ply].pv_move = NO_MOVE
		}
		score, sum := quiescence(brd, stk, alpha, beta, depth, ply, MAX_Q_CHECKS) // q-search is always sequential.
		return score, sum
	}

	if cancel_search {
		return 0, 0
	}

	this_stk = &stk[ply]

	this_stk.hash_key = brd.hash_key
	if stk.IsRepetition(ply) { // check for draw by threefold repetition
		return 0, 1
		// return -ply, 1
	}

	in_check = this_stk.in_check

	if in_check != is_in_check(brd) {
		fmt.Println(ply)
		brd.Print()
	}

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

	if depth > 6 {
		null_depth = depth - 3
	} else {
		null_depth = depth - 2
	}

	first_move, hash_result = main_tt.probe(brd, depth, null_depth, &alpha, &beta, &score)

	eval = evaluate(brd, alpha, beta)	
	this_stk.eval = eval

	if node_type != Y_PV {
		if (hash_result & CUTOFF_FOUND) > 0 { // Hash hit valid for current bounds.
			return score, sum
		} else if !in_check && can_null && hash_result != AVOID_NULL && depth > 2 && in_endgame(brd) == 0 &&
		!brd.PawnsOnly() && eval >= beta {
			score, subtotal = null_make(brd, stk, beta, null_depth, ply, extensions_left)
			sum += subtotal
			if score >= beta {
				main_tt.store(brd, 0, depth, LOWER_BOUND, score)
				return score, sum
			}
		}
	}

	// // skip IID when in check?
	// if !in_check && node_type == Y_PV && hash_result == NO_MATCH && can_null && depth >= IID_MIN { 
	// 	// No hash move available. Use IID to get a decent first move to try.
	// 	score, subtotal = ybw(brd, stk, alpha, beta, depth-2, ply, extensions_left, can_null, node_type, SP_NONE)
	// 	sum += subtotal
	// 	first_move = this_stk.pv_move
	// }

	if !in_check && ply > 0 && node_type != Y_PV && alpha > 100-MATE {
		if depth <= F_PRUNE_MAX && eval+piece_values[BISHOP] < alpha {
			f_prune = true
		}
		if depth >= LMR_MIN {
			can_reduce = true
		}
	}

	selector = NewMoveSelector(brd, this_stk, in_check, first_move)

search_moves:
	memento := brd.NewMemento()

	for m, stage := selector.Next(sp_type); m != NO_MOVE; m, stage = selector.Next(sp_type) {

		if sp_type == SP_SERVANT {
			select {
			case <-sp.cancel:
				return NO_SCORE, 0
			default:
			}
		}

		make_move(brd, m)

		gives_check := is_in_check(brd)

		if f_prune && stage == STAGE_REMAINING && legal_searched > 0 && m.IsQuiet() &&
			!is_passed_pawn(brd, m) && !brd.PawnsOnly() && !gives_check {
			unmake_move(brd, m, memento)
			continue
		}

		child_type = determine_child_type(node_type, legal_searched)

		r_depth, r_extensions = depth, extensions_left
		if m.IsPromotion() && stage == STAGE_WINNING && extensions_left > 0 {
			r_depth = depth + 1  // extend winning promotions only
			r_extensions = extensions_left - 1
		} else if can_reduce && stage == STAGE_REMAINING && 
			((node_type == Y_ALL && legal_searched > 2) || legal_searched > 6) && !is_passed_pawn(brd, m) && 
			!gives_check {
			r_depth = depth - 1 // Late move reductions
		}

		stk[ply+1].in_check = gives_check // avoid having to recalculate in_check at beginning of search.
		total = 0
		if node_type == Y_PV && alpha > old_alpha {
			score, subtotal = ybw(brd, stk, -alpha-1, -alpha, r_depth-1, ply+1, r_extensions, can_null, child_type, SP_NONE)
			score = -score
			total += subtotal
			if score > alpha {
				score, subtotal = ybw(brd, stk, -beta, -alpha, r_depth-1, ply+1, r_extensions, can_null, Y_PV, SP_NONE)
				score = -score
				total += subtotal
			}
		} else {
			score, subtotal = ybw(brd, stk, -beta, -alpha, r_depth-1, ply+1, r_extensions, can_null, child_type, SP_NONE)
			score = -score
			total += subtotal
		}

		unmake_move(brd, m, memento) 

		if sp_type != SP_NONE {

			sp.Lock()

			alpha = sp.alpha // get the latest info
			beta = sp.beta
			best = sp.best
			best_move = sp.best_move
			legal_searched = sp.legal_searched
			sum = sp.node_count

			sp.node_count += total
			sp.legal_searched += 1

			if score > best {
				best_move = m
				sp.best_move = m
				best = score
				sp.best = score
				if score > alpha {
					alpha = score
					sp.alpha = score

					if node_type == Y_PV { 	// will need to update this for parallel search
						this_stk.pv_move, this_stk.value, this_stk.depth = m, score, depth
					}

					if score >= beta {
						sp.Unlock()
						if sp_type == SP_MASTER {
							// The SP master has finished evaluating the node. Remove the SP from the worker's SP List.
							load_balancer.remove_sp <- SPCancellation{sp, brd.hash_key}
							store_cutoff(this_stk, m, brd.c, total) // what happens on refutation of main pv?
							main_tt.store(brd, m, depth, LOWER_BOUND, score)
							return score, sum
						} else {
							return NO_SCORE, 0
						}
					}
				}
			}
			sp.Unlock()
		} else {
			sum += total
			if score > best {
				if score > alpha {
					if node_type == Y_PV {
						this_stk.pv_move, this_stk.value, this_stk.depth = m, score, depth
					}
					if score >= beta {
						store_cutoff(this_stk, m, brd.c, total) // what happens on refutation of main pv?
						main_tt.store(brd, m, depth, LOWER_BOUND, score)
						return score, sum
					}
					alpha = score
				}
				best_move = m
				best = score
			}
			legal_searched += 1
		}

		// Determine if this would be a good location to begin searching in parallel.
		if sp_type == SP_NONE && can_split(brd, ply, depth, node_type, legal_searched, stage) {
			if setup_sp(brd, stk, selector, best_move, alpha, beta, best, depth, ply, 
							 		extensions_left, legal_searched, can_null, node_type, total) {
				sp = this_stk.sp
				sp_type = SP_MASTER
			}
		}

	} // end of moves loop


	switch sp_type {
	case SP_MASTER: // The SP master has finished evaluating the node. Ask the load balancer to remove the SP 
									// from the worker's SP List.
		load_balancer.remove_sp <- SPCancellation{sp, brd.hash_key}
	case SP_SERVANT:
		return NO_SCORE, 0
	default:
	}


	if legal_searched > 0 {
		if node_type == Y_PV {
			this_stk.pv_move, this_stk.value, this_stk.depth = best_move, best, depth
		}
		if alpha > old_alpha {
			main_tt.store(brd, best_move, depth, EXACT, best) 
			return best, sum
		} else {
			main_tt.store(brd, best_move, depth, UPPER_BOUND, best)
			return best, sum
		}
	} else {
		if in_check { // Checkmate.
			main_tt.store(brd, 0, depth, EXACT, ply-MATE)
			return ply - MATE, sum
		} else { // Draw.
			main_tt.store(brd, 0, depth, EXACT, 0)
			return 0, sum
		}
	}
}

// Determine if the current node is a good place to start searching in parallel.
func can_split(brd *Board, ply, depth, node_type, legal_searched, stage int) bool {
	if depth >= SPLIT_MIN {
		switch node_type {
		case Y_PV:
			return legal_searched > 0 && ply > 0
		case Y_CUT:
			return legal_searched > 3 && stage == STAGE_REMAINING
		case Y_ALL:
			return legal_searched > 0
		}		
	}
	return false
}


func setup_sp(brd *Board, stk Stack, ms *MoveSelector, best_move Move, alpha, beta, best, depth, ply, extensions_left, legal_searched int,
							 can_null bool, node_type, total int) bool {
	worker := brd.worker
	select {
	case <-worker.available_slots:
		brd_copy := brd.Copy()
		stk_copy := stk[ply].Copy()
		ms.brd = brd_copy // make sure the move selector points to the static SP board. 
		ms.this_stk = stk_copy

		sp := &SplitPoint{
			selector: ms,
			master: worker,
			brd: brd_copy,
			this_stk: stk_copy,

			depth: depth,
			ply: ply,
			extensions_left: extensions_left,
			can_null: can_null,
			node_type: node_type,

			alpha: alpha,
			beta: beta,
			best: best,
			best_move: best_move,

			node_count: total,
			legal_searched: legal_searched,
			cancel: make(chan bool),
		}
		stk[ply].sp = sp

		load_balancer.work <- SPListItem{sp: sp, stk: stk.CopyUpTo(ply), order: uint8((sp.depth << 2)|sp.node_type) }
		return true
	default:
		return false
	}
}



// Q-Search will always be done sequentially: Q-search subtrees are taller and narrower than in the main search,
// making benefit of parallelism smaller and raising communication and synchronization overhead.
func quiescence(brd *Board, stk Stack, alpha, beta, depth, ply, checks_remaining int) (int, int) {
	if cancel_search {
		return 0, 0
	}

	this_stk := &stk[ply]

	this_stk.hash_key = brd.hash_key
	if stk.IsRepetition(ply) {
		return 0, 1
		// return -ply, 1
	}

	in_check := this_stk.in_check
	if brd.halfmove_clock >= 100 {
		if is_checkmate(brd, in_check) {
			return ply - MATE, 1
		} else {
			return 0, 1
		}
	}


	score, best, sum, total := -INF, -INF, 1, 0
	r_checks_remaining := checks_remaining

	if in_check {
		r_checks_remaining = checks_remaining - 1
	} else {
		score = evaluate(brd, alpha, beta) // stand pat
		this_stk.eval = score
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

	legal_moves := false
	memento := brd.NewMemento()
	selector := NewQMoveSelector(brd, this_stk, in_check, checks_remaining > 0)

	for m := selector.Next(false); m != NO_MOVE; m = selector.Next(false) {

		make_move(brd, m)

		gives_check := is_in_check(brd)

		if !in_check && !gives_check && alpha > 100-MATE &&
			best+m.CapturedPiece().Value()+m.PromotedTo().PromoteValue()+piece_values[ROOK] < alpha {
			unmake_move(brd, m, memento)
			continue
		}

		stk[ply+1].in_check = gives_check // avoid having to recalculate in_check at beginning of search.

		score, total = quiescence(brd, stk, -beta, -alpha, depth-1, ply+1, r_checks_remaining)
		score = -score
		sum += total
		unmake_move(brd, m, memento)

		if score > best {
			if score > alpha {
				if score >= beta {
					return score, sum
				}
				alpha = score
			}
			best = score
		}
		legal_moves = true
	}

	if in_check && !legal_moves {
		return ply - MATE, 1 // detect checkmate.
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

func null_make(brd *Board, stk Stack, beta, null_depth, ply, extensions_left int) (int, int) {
	hash_key, enp_target := brd.hash_key, brd.enp_target
	brd.c ^= 1
	brd.hash_key ^= side_key
	brd.hash_key ^= enp_zobrist(enp_target)
	brd.enp_target = SQ_INVALID

	stk[ply+1].in_check = is_in_check(brd)
	score, sum := ybw(brd, stk, -beta, -beta+1, null_depth-1, ply+1, extensions_left, false, Y_CUT, SP_NONE)

	brd.c ^= 1
	brd.hash_key = hash_key
	brd.enp_target = enp_target
	return -score, sum
}

func store_cutoff(this_stk *StackItem, m Move, c uint8, total int) {
	if m.IsQuiet() {
		main_htable.Store(m, c, total)
		this_stk.StoreKiller(m) // store killer moves in stack for this Goroutine.
	}
}
