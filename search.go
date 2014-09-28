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
	"math"
	"time"
)

const (
	MAX_TIME  = 120000 // default search time limit in seconds (2m)
	MAX_DEPTH = 12
	SPLIT_MIN = 13 // set > MAX_DEPTH to disable parallel search.
	F_PRUNE_MIN = 3 // should always be less than SPLIT_MIN
	EXT_MAX   = 4
	MAX_PLY   = MAX_DEPTH + EXT_MAX
	IID_MIN   = 6
	COMMS_MIN = 5 // minimum depth at which to send info to GUI.
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

var search_id int
var cancel_search bool
var uci_mode bool = false
var uci_ponder bool = false

func AbortSearch() {
	cancel_search = true
	fmt.Println("Search aborted by GUI")
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

func iterative_deepening(brd *Board, depth int, start time.Time) (Move, int) {
	var move Move
	var guess, count, first_count, sum int
	var previous_count int
	var old_pv, current_pv *PV
	c := brd.c

	// first iteration is always full-depth.
	id_alpha, id_beta = -INF, INF
	move, guess, count, old_pv = ybw_root(brd, id_alpha, id_beta, id_score[c], 1, nil)

	sum, first_count, previous_count = count, count, count
	id_move[c], id_score[c] = move, guess

	for d := 2; d <= depth; d++ {
		// to do: add aspiration windows

		move, guess, count, current_pv = ybw_root(brd, id_alpha, id_beta, id_score[c], d, old_pv)

		if cancel_search {
			avg_branch := math.Pow(float64(sum)/float64(first_count), float64(1)/float64(depth-1))
			if !uci_mode {
				fmt.Printf("Average Branching: %.4f\n", avg_branch)
			}
			return id_move[c], sum
		} else {
			sum += count
			if d > COMMS_MIN { // don't print info for first few plys to reduce communication traffic.
				if uci_mode {
					PrintInfo(guess, d, sum, time.Since(start), current_pv)
				} else {
					fmt.Printf("Depth: %d Branching: %.4f\n", d, float64(count)/float64(previous_count))
				}
			}
			id_move[c], id_score[c] = move, guess
			previous_count, old_pv = count, current_pv
		}
	}
	PrintInfo(guess, depth, sum, time.Since(start), current_pv)
	avg_branch := math.Pow(float64(sum)/float64(first_count), float64(1)/float64(depth-1))
	fmt.Printf("Average Branching: %.4f\n", avg_branch)

	return id_move[c], sum
}

func ybw_root(brd *Board, alpha, beta, guess, depth int, old_pv *PV) (Move, int, int, *PV) {
	if cancel_search {
		return 0, 0, 1, nil
	}
	sum, count, legal_searched := 1, 0, 0
	in_check, extension := is_in_check(brd), 0
	if in_check {
		extension = 1
	}

	score, best, old_alpha := -INF, -INF, alpha
	pv := &PV{}
	var next_pv *PV
	var best_move, first_move Move

	// root is always on the main PV line.  No need to search hash table.
	if old_pv != nil {
		first_move = old_pv.m
		legal_searched += 1
		if uci_mode && uci_ponder && depth > COMMS_MIN {
			fmt.Printf("info currmove %s currmovenumber %d\n", first_move.ToString(), legal_searched)
		}
		score, count, next_pv = ybw_make(brd, first_move, alpha, beta, depth-1+extension, 1, true, nil)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, first_move, depth, count)
					main_tt.store(brd, first_move, depth, LOWER_BOUND, score)
					return first_move, score, sum, nil
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
	best_moves, remaining_moves := get_best_moves(brd, in_check)
	var m Move
	for _, item := range *best_moves { // search the best moves sequentially.
		m = item.move
		if m == first_move || !avoids_check(brd, m, in_check) {
			continue
		}
		legal_searched += 1
		if uci_mode && uci_ponder && depth > COMMS_MIN {
			fmt.Printf("info currmove %s currmovenumber %d\n", m.ToString(), legal_searched)
		}
		score, count, next_pv = ybw_make(brd, m, alpha, beta, depth-1+extension, 1, true, nil)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, m, depth, count)
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return m, score, sum, nil
				}
				alpha = score
				pv.m = m
				pv.next = next_pv
			}
			best_move = m
			best = score
		}
	}

	// Delay the generation of remaining moves until all promotions and winning captures have been searched.
	// if a cutoff occurs, this will reduce move generation effort substantially.
	get_remaining_moves(brd, in_check, remaining_moves)

	for _, item := range *remaining_moves { // search remaining moves sequentially.
		m = item.move
		if m == first_move || !avoids_check(brd, m, in_check) {
			continue
		}
		legal_searched += 1
		if uci_mode && depth > 5 {
			fmt.Printf("info currmove %s currmovenumber %d\n", m.ToString(), legal_searched)
		}
		score, count, next_pv = ybw_make(brd, m, alpha, beta, depth-1+extension, 1, true, nil)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, m, depth, count)
					main_tt.store(brd, m, depth, LOWER_BOUND, score)
					return m, score, sum, nil
				}
				alpha = score
				pv.m = m
				pv.next = next_pv
			}
			best_move = m
			best = score
		}
	}

	if legal_searched > 0 {
		if alpha > old_alpha {
			main_tt.store(brd, best_move, depth, EXACT, best)
			return best_move, best, sum, pv
		} else {
			main_tt.store(brd, best_move, depth, UPPER_BOUND, best)
			return best_move, best, sum, nil
		}
	} else {
		if in_check {
			main_tt.store(brd, 0, MAX_PLY, EXACT, -INF) // should these be EXACT nodes?
			return 0, -INF, sum, nil                    // checkmate.
		} else {
			main_tt.store(brd, 0, MAX_PLY, EXACT, 0)
			return 0, 0, sum, nil // draw.
		}
	}
}


func young_brothers_wait(brd *Board, alpha, beta, depth, ply int, can_null bool, old_pv *PV) (int, int, *PV) {
	if cancel_search {
		return 0, 0, nil
	}

	if depth <= 0 {
		return quiescence(brd, alpha, beta, depth, ply, nil) // q-search is always sequential.
	}

	in_check, extension := is_in_check(brd), 0
	if in_check {
		extension = 1
	}

	if brd.halfmove_clock >= 100 {
		if is_checkmate(brd, in_check) {
			return ply - INF, 1, nil
		} else {
			// fmt.Printf("Draw by Halfmove Rule at ply %d\n", ply)
			return 0, 1, nil
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
	pv := &PV{}
	var next_pv *PV
	var best_move, first_move Move

	if old_pv == nil { // if on main PV line from previous iteration, no need to probe hash table.
		// Never try null move pruning while on main PV line.
		var hash_result int
		first_move, hash_result = main_tt.probe(brd, depth, null_depth, &alpha, &beta, &score)
		if hash_result == CUTOFF_FOUND {
			return score, sum, nil
		} else if hash_result != AVOID_NULL { // Null-Move Pruning
			if !in_check && can_null && depth > 2 && in_endgame(brd, brd.c) == 0 &&
				!pawns_only(brd, brd.c) && evaluate(brd, alpha, beta) >= beta {
				score, count = null_make(brd, beta, null_depth-1, ply+1)
				sum += count
				if score >= beta {
					main_tt.store(brd, 0, depth, LOWER_BOUND, score)
					return score, sum, nil
				}
			}
		}
		if hash_result == NO_MATCH && can_null && depth >= IID_MIN { // No hash move available. Use IID to get a decent first move to try.
			var local_id_alpha, local_id_beta int
			if (ply&1) == 0 { // test if odd-ply
				local_id_alpha, local_id_beta = -id_beta, -id_alpha
			} else {
				local_id_alpha, local_id_beta = id_alpha, id_beta
			}
			if alpha == local_id_alpha && beta == local_id_beta {  // Only use IID at expected PV nodes.
				// This assumes use of PVSearch where non-PV nodes are searched with null windows. (as in PVS)
				var local_pv *PV
				score, count, local_pv = young_brothers_wait(brd, alpha, beta, depth-2+extension, ply, true, nil)
				sum += count
				if local_pv != nil {
					first_move = local_pv.m						
				}
			}
		}
	}

	// If a PV move, hash move or IID move is available, try it first.
	if old_pv != nil || (is_valid_move(brd, first_move, depth)&&avoids_check(brd, first_move, in_check)) {
		var temp_pv *PV
		if old_pv != nil {
			first_move = old_pv.m
			temp_pv = old_pv.next
		}
		legal_searched += 1
		score, count, next_pv = ybw_make(brd, first_move, alpha, beta, depth-1+extension, ply+1, can_null, temp_pv)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					// what happens if a move is refuted here while on the main PV?
					store_cutoff(brd, first_move, depth, count)
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
	best_moves, remaining_moves := get_best_moves(brd, in_check)
	var m Move
	for _, item := range *best_moves { // search the best moves sequentially.
		m = item.move
		if m == first_move || !avoids_check(brd, m, in_check) {
			continue
		}
		legal_searched += 1
		score, count, next_pv = ybw_make(brd, m, alpha, beta, depth-1+extension, ply+1, can_null, nil)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta {
					store_cutoff(brd, m, depth, count)
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

	// Delay the generation of remaining moves until all promotions and winning captures have been searched.
	// if a cutoff occurs, this will reduce move generation effort substantially.
	get_remaining_moves(brd, in_check, remaining_moves)

	if depth <= SPLIT_MIN { // Depth is too shallow for parallel search to be worthwhile.
		// Extended futility pruning:
		f_prune := false
		if depth <= F_PRUNE_MIN {
			f_prune = !in_check && old_pv == nil && alpha > 100-INF && evaluate(brd, alpha, beta)+piece_values[BISHOP] < alpha
		}
		hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
		castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

		for _, item := range *remaining_moves { // search remaining moves sequentially.
			m = item.move
			if m == first_move || !avoids_check(brd, m, in_check) {
				continue
			}

			make_move(brd, m) // to do: make move
			if f_prune && legal_searched > 1 && m.IsQuiet() && m.Piece() != PAWN && !is_in_check(brd) {
				unmake_move(brd, m, enp_target) // to do: unmake move
				brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
				brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
				continue
			}
			score, count, next_pv := young_brothers_wait(brd, -beta, -alpha, depth-1+extension, ply+1, can_null, nil)
			legal_searched += 1
			score = -score
			unmake_move(brd, m, enp_target) // to do: unmake move
			brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
			brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock

			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						store_cutoff(brd, m, depth, count) // what happens on refutation of main pv?
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
	} else { // now that decent bounds have been established, parallel search is possible.
		// Make sure at least 3 nodes have been searched serially before spawning.
		for ; legal_searched < 3; legal_searched++ {
			item := remaining_moves.Dequeue()
			for item != nil && !avoids_check(brd, item.move, in_check) {
				item = remaining_moves.Dequeue() // get the highest-sorted legal move from remaining_moves
			}
			if item == nil {
				break
			}
			score, count, next_pv = ybw_make(brd, item.move, alpha, beta, depth-1+extension, 1, true, nil)
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						store_cutoff(brd, item.move, depth, count)
						main_tt.store(brd, item.move, depth, LOWER_BOUND, score)
						return score, sum, nil
					}
					alpha = score
					pv.m = m
					pv.next = next_pv
				}
				best_move = item.move
				best = score
			}
		}

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
				score, count, next_pv := ybw_make(new_brd, m, alpha, beta, depth-1+extension, ply+1, can_null, nil)
				result_child <- SearchResult{m, score, count, next_pv}
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
								return result.score, sum, nil
							}
							alpha = result.score
							pv.m = result.move
							pv.next = result.pv
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

// Q-Search will always be done sequentially.
// Q-search subtrees are taller and narrower than in the main search making benefit of parallelism
// smaller and raising communication and synchronization overhead.
func quiescence(brd *Board, alpha, beta, depth, ply int, old_pv *PV) (int, int, *PV) {
	in_check := is_in_check(brd)
	if brd.halfmove_clock >= 100 {
		if is_checkmate(brd, in_check) {
			return ply - INF, 1, nil
		} else {
			// fmt.Printf("Draw by Halfmove Rule at ply %d\n", ply)
			return 0, 1, nil
		}
	}

	score, best := -INF, -INF
	sum, count := 1, 0
	pv := &PV{}
	var next_pv *PV
	var first_move Move
	legal_moves := false

	if old_pv != nil {	// Search the main pv line first, if any.
		legal_moves = true
		score, count, next_pv = q_make(brd, first_move, alpha, beta, depth-1, ply+1, old_pv.next)
		sum += count
		if score > best {
			if score > alpha {
				if score >= beta { // what happens if a move is refuted here while on the main PV?
					return score, sum, nil
				}
				alpha = score
				pv.m = first_move
				pv.next = next_pv
			}
			best = score
		}
	}

	var m Move
	if in_check {
		best_moves, remaining_moves := &MoveList{}, &MoveList{}
		get_evasions(brd, best_moves, remaining_moves) // only legal moves generated here.
		for _, item := range *best_moves {
			m = item.move
			legal_moves = true
			score, count, next_pv = q_make(brd, m, alpha, beta, depth-1, ply+1, nil)
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						return score, sum, nil
					}
					alpha = score
					pv.m = m
					pv.next = next_pv
				}
				best = score
			}
		}
		for _, item := range *remaining_moves {
			m = item.move
			legal_moves = true
			score, count, next_pv = q_make(brd, m, alpha, beta, depth-1, ply+1, nil)
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						return score, sum, nil
					}
					alpha = score
					pv.m = m
					pv.next = next_pv
				}
				best = score
			}
		}
		if !legal_moves {
			return ply-INF, 1, nil // detect checkmate.
		}
	} else {

		score = evaluate(brd, alpha, beta) // stand pat
		if score > best {
			if score > alpha {
				if score >= beta {
					return score, sum, nil
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
			if old_pv == nil && alpha > 100-INF && best+m.CapturedPiece().Value()+m.PromotedTo().PromoteValue()+piece_values[ROOK] < alpha {
				continue // prune futile moves with no chance of raising alpha.
			}
			score, count, next_pv = q_make(brd, m, alpha, beta, depth-1, ply+1, nil)
			sum += count
			if score > best {
				if score > alpha {
					if score >= beta {
						return score, sum, nil
					}
					alpha = score
					pv.m = first_move
					pv.next = next_pv
				}
				best = score
			}
		}
	}

	if pv.m > 0 {
		return best, sum, pv
	} else {
		return best, sum, nil
	}
}

func ybw_make(brd *Board, m Move, alpha, beta, depth, ply int, can_null bool, old_pv *PV) (int, int, *PV) {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score, sum, pv := young_brothers_wait(brd, -beta, -alpha, depth, ply, can_null, old_pv)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return -score, sum, pv
}

func q_make(brd *Board, m Move, alpha, beta, depth, ply int, old_pv *PV) (int, int, *PV) {
	hash_key, pawn_hash_key := brd.hash_key, brd.pawn_hash_key
	castle, enp_target, halfmove_clock := brd.castle, brd.enp_target, brd.halfmove_clock

	make_move(brd, m) // to do: make move
	score, sum, pv := quiescence(brd, -beta, -alpha, depth, ply, old_pv)
	unmake_move(brd, m, enp_target) // to do: unmake move

	brd.hash_key, brd.pawn_hash_key = hash_key, pawn_hash_key
	brd.castle, brd.enp_target, brd.halfmove_clock = castle, enp_target, halfmove_clock
	return -score, sum, pv
}

func null_make(brd *Board, beta, depth, ply int) (int, int) {
	hash_key, enp_target := brd.hash_key, brd.enp_target
	brd.c ^= 1
	brd.hash_key ^= side_key
	brd.hash_key ^= enp_zobrist(enp_target)
	brd.enp_target = SQ_INVALID

	score, sum, _ := young_brothers_wait(brd, -beta, -beta+1, depth, ply, false, nil)

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
