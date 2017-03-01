//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"sync"
)

const ( // TODO: expose these as options via UCI interface.
	MIN_SPLIT       = 2  // Do not begin parallel search below this depth.
	F_PRUNE_MAX     = 2  // Do not use futility pruning when above this depth.
	LMR_MIN         = 2  // Do not use late move reductions below this depth.
	IID_MIN         = 4  // Do not use internal iterative deepening below this depth.
	NULL_MOVE_MIN   = 3  // Do not use null-move pruning below this depth.
	MIN_CHECK_DEPTH = -2 // During Q-Search, consider all evasions when in check at or above this depth.

	DRAW_VALUE = KNIGHT_VALUE // The value to assign to a draw
)

const (
	INF      = 10000            // an arbitrarily large score used for initial bounds
	NO_SCORE = INF - 1          // sentinal value indicating a meaningless score.
	MATE     = NO_SCORE - 1     // maximum checkmate score (i.e. mate in 0)
	MIN_MATE = MATE - MAX_STACK // minimum possible checkmate score (mate in MAX_STACK)
)

const (
	MAX_DEPTH = 32 // default maximum search depth
	COMMS_MIN = 1  // minimum depth at which to send info to GUI.
)

const (
	Y_CUT = iota // YBWC node types
	Y_ALL
	Y_PV
)

var searchId int

type Search struct {
	htable HistoryTable // must be listed first to ensure cache alignment for atomic w/r
	SearchParams
	sideToMove           uint8 // SearchParams would otherwise create padding
	once                 sync.Once
	allowedMoves         []Move
	bestScore            [2]int
	cancel               chan bool
	bestMove, ponderMove Move
	gt                   *GameTimer
	uci                  *UCIAdapter
	alpha, beta, nodes   int
}

type SearchParams struct {
	maxDepth                        int
	verbose, ponder, restrictSearch bool
}

type SearchResult struct {
	bestMove, ponderMove Move
}

func NewSearch(params SearchParams, gt *GameTimer, uci *UCIAdapter, allowedMoves []Move) *Search {
	s := &Search{
		bestScore:    [2]int{-INF, -INF},
		cancel:       make(chan bool),
		uci:          uci,
		bestMove:     NO_MOVE,
		ponderMove:   NO_MOVE,
		alpha:        -INF,
		beta:         -INF,
		gt:           gt,
		SearchParams: params,
		allowedMoves: allowedMoves,
	}
	gt.s = s
	if !s.ponder {
		gt.Start()
	}
	return s
}

func (s *Search) sendResult() {
	s.once.Do(func() {
		if s.uci != nil {
			if s.ponder {
				s.uci.result <- s.Result() // queue result to be sent when requested by GUI.
			} else {
				s.uci.BestMove(s.Result()) // send result immediately
			}
		}
	})
}

func (s *Search) Result() SearchResult {
	return SearchResult{s.bestMove, s.ponderMove}
}

func (s *Search) Abort() {
	select {
	case <-s.cancel:
	default:
		close(s.cancel)
	}
}

func (s *Search) moveAllowed(m Move) bool {
	for _, permittedMove := range s.allowedMoves {
		if m == permittedMove {
			return true
		}
	}
	return false
}

func (s *Search) sendInfo(str string) {
	if s.uci != nil {
		s.uci.InfoString(str)
	} else if s.verbose {
		fmt.Print(str)
	}
}

func (s *Search) Start(brd *Board) {
	s.sideToMove = brd.c
	brd.worker = loadBalancer.RootWorker() // Send SPs generated by root goroutine to root worker.

	s.nodes = s.iterativeDeepening(brd)

	if searchId >= 512 { // only 9 bits are available to store the id in each TT entry.
		searchId = 0
	} else {
		searchId += 1
	}
	s.gt.Stop() // s.cancel the timer to prevent it from interfering with the next search if it's not
	// garbage collected before then.
	s.sendResult()
	if s.uci != nil {
		s.uci.wg.Done()
	}
}

func (s *Search) iterativeDeepening(brd *Board) int {
	var guess, total, sum int
	c := brd.c
	stk := brd.worker.stk
	s.alpha, s.beta = -INF, INF // first iteration is always full-width.
	inCheck := brd.InCheck()

	for d := 1; d <= s.maxDepth; d++ {

		stk[0].inCheck = inCheck
		guess, total = s.ybw(brd, stk, s.alpha, s.beta, d, 0, Y_PV, SP_NONE, false)
		sum += total

		select { // if the cancel signal was received mid-search, the current guess is not useful.
		case <-s.cancel:
			return sum
		default:
		}

		if stk[0].pv.m.IsMove() {
			s.bestMove, s.bestScore[c] = stk[0].pv.m, guess
			if stk[0].pv.next != nil {
				s.ponderMove = stk[0].pv.next.m
			}

			stk[0].pv.SavePV(brd, d, guess) // install PV to transposition table prior to next iteration.

		} else {
			s.sendInfo("Nil PV returned to ID\n")
		}
		if d >= COMMS_MIN && (s.verbose || s.uci != nil) { // don't print info for first few plies to reduce communication traffic.
			s.uci.Info(Info{guess, d, sum, s.gt.Elapsed(), stk})
		}
	}

	return sum
}

func (s *Search) ybw(brd *Board, stk Stack, alpha, beta, depth, ply, nodeType,
	spType int, checked bool) (int, int) {
	select {
	case <-s.cancel:
		return NO_SCORE, 1
	default:
	}

	if depth <= 0 {
		if nodeType == Y_PV {
			stk[ply].pv = nil
		}
		return s.quiescence(brd, stk, alpha, beta, 0, ply) // q-search is always sequential.
	}

	var thisStk *StackItem
	var inCheck bool
	var sp *SplitPoint
	var pv *PV
	var selector *MoveSelector

	score, best, oldAlpha := -INF, -INF, alpha
	sum := 1

	var nullDepth, hashResult, eval, subtotal, total, legalSearched, childType, rDepth int
	// var hashScore int
	canPrune, fPrune, canReduce := false, false, false
	bestMove, firstMove := NO_MOVE, NO_MOVE

	recycler := brd.worker.recycler
	// if the is_sp flag is set, a worker has just been assigned to this split point.
	// the SP master has already handled most of the pruning, so just read the latest values
	// from the SP and jump to the moves loop.
	if spType == SP_SERVANT {
		sp = stk[ply].sp
		sp.RLock()
		selector = sp.selector
		thisStk = sp.thisStk
		eval = int(thisStk.eval)
		inCheck = thisStk.inCheck
		sp.RUnlock()
		goto searchMoves
	}

	thisStk = &stk[ply]

	if nodeType != Y_PV { // Mate Distance Pruning
		mateValue := max(ply-MATE, alpha)
		if mateValue >= min(MATE-ply, beta) {
			return mateValue, sum
		}
	}

	thisStk.hashKey = brd.hashKey
	if stk.IsRepetition(ply, brd.halfmoveClock) { // check for draw by threefold repetition
		return ply - DRAW_VALUE, 1
	}

	inCheck = thisStk.inCheck

	if brd.halfmoveClock >= 100 { // check for draw by halfmove rule
		// TODO: handle non-king moves that escape from check
		if isCheckmate(brd, inCheck) {
			return ply - MATE, 1
		} else {
			return ply - DRAW_VALUE, 1
		}
	}

	nullDepth = depth - 4
	firstMove, hashResult = mainTt.probe(brd, depth, nullDepth, alpha, beta, &score)
	// hashScore = score

	eval = evaluate(brd, alpha, beta)
	thisStk.eval = int16(eval)

	if nodeType != Y_PV {
		if (hashResult & CUTOFF_FOUND) > 0 { // Hash hit valid for current bounds.
			return score, sum
		} else if !inCheck && thisStk.canNull && hashResult != AVOID_NULL && depth >= NULL_MOVE_MIN &&
			!brd.PawnsOnly() && eval >= beta { // Null-move pruning

			score, subtotal = s.nullMake(brd, stk, beta, nullDepth, ply, checked)
			sum += subtotal
			if score >= beta && score < MIN_MATE {
				if depth >= 8 { //  Null-move Verification search
					thisStk.canNull = false
					score, subtotal = s.ybw(brd, stk, beta-1, beta, nullDepth-1, ply, nodeType, SP_NONE, checked)
					thisStk.canNull = true
					sum += subtotal
					if score >= beta && score < MIN_MATE {
						return score, sum
					}
				} else {
					return score, sum
				}
			}
		}
	}

	// skip IID when in check?
	if !inCheck && nodeType == Y_PV && hashResult == NO_MATCH && depth >= IID_MIN {
		// No hash move available. Use IID to get a decent first move to try.
		score, subtotal = s.ybw(brd, stk, alpha, beta, depth-2, ply, Y_PV, SP_NONE, checked)
		sum += subtotal
		if thisStk.pv != nil {
			firstMove = thisStk.pv.m
		}
	}

	// recycler = brd.worker.recycler
	selector = recycler.ReuseMoveSelector(brd, thisStk, &s.htable, inCheck, firstMove)

searchMoves:

	if nodeType == Y_PV { // remove any stored pv move from a previous iteration.
		pv = &PV{}
	}

	if inCheck {
		checked = true // Don't extend on the first check in the current variation.
	} else if ply > 0 && alpha > -MIN_MATE {
		if depth <= F_PRUNE_MAX && !brd.PawnsOnly() {
			canPrune = true
			if eval+BISHOP_VALUE < alpha {
				fPrune = true
			}
		}
		if depth >= LMR_MIN {
			canReduce = true
		}
	}

	// if nodeType == Y_PV {
	// 	fmt.Printf("%d:%t, %b\n", ply, hashResult&EXACT_FOUND > 0, hashResult)
	// }

	// singularNode := ply > 0 && nodeType == Y_CUT && (hashResult&BETA_FOUND) > 0 &&
	// 	firstMove.IsMove() && depth > 6 && thisStk.canNull

	memento := brd.NewMemento()

	var mayPromote, tryPrune, givesCheck bool

	for m, stage := selector.Next(spType); m != NO_MOVE; m, stage = selector.Next(spType) {

		if ply == 0 && s.restrictSearch {
			if !s.moveAllowed(m) { // restrict search to only those moves requested by the GUI.
				continue
			}
		}

		if m == thisStk.singularMove {
			continue
		}

		mayPromote = brd.MayPromote(m)
		tryPrune = canPrune && stage == STAGE_REMAINING && legalSearched > 0 && !mayPromote

		if tryPrune && getSee(brd, m.From(), m.To(), EMPTY) < 0 {
			continue // prune quiet moves that result in loss of moving piece
		}

		total = 0
		rDepth = depth

		// // TODO: verify safety for parallel search
		// // Singular extension
		// if singularNode && spType == SP_NONE && m == firstMove {
		// 	sBeta := hashScore - (depth << 1)
		// 	thisStk.singularMove, thisStk.canNull = m, false
		// 	score, total = s.ybw(brd, stk, sBeta-1, sBeta, depth/2, ply, Y_CUT, SP_NONE, checked)
		// 	thisStk.singularMove, thisStk.canNull = NO_MOVE, true
		// 	if score < sBeta {
		// 		rDepth = depth + 1 // extend moves that are expected to be the only move searched.
		// 	}
		// }

		makeMove(brd, m)

		givesCheck = brd.InCheck()

		if fPrune && tryPrune && !givesCheck {
			unmakeMove(brd, m, memento)
			continue
		}

		childType = s.determineChildType(nodeType, legalSearched)

		if rDepth == depth {
			if stage == STAGE_WINNING && mayPromote && m.IsPromotion() {
				rDepth = depth + 1 // extend winning promotions only
			} else if givesCheck && checked && ply > 0 && (stage < STAGE_LOSING ||
				// don't extend suicidal checks
				(stage == STAGE_REMAINING && getSee(brd, m.From(), m.To(), EMPTY) >= 0)) {
				rDepth = depth + 1 // only extend "useful" checks after the first check in a variation.
			} else if canReduce && !mayPromote && !givesCheck &&
				stage >= STAGE_REMAINING && ((nodeType == Y_ALL && legalSearched > 2) ||
				legalSearched > 6) {
				rDepth = depth - 1 // Late move reductions
			}
		}

		stk[ply+1].inCheck = givesCheck // avoid having to recalculate in_check at beginning of search.

		// time to search deeper:
		if nodeType == Y_PV && alpha > oldAlpha {
			score, subtotal = s.ybw(brd, stk, (-alpha)-1, -alpha, rDepth-1, ply+1, childType, SP_NONE, checked)
			score = -score
			total += subtotal
			if score > alpha { // re-search with full-window on fail high
				score, subtotal = s.ybw(brd, stk, -beta, -alpha, rDepth-1, ply+1, Y_PV, SP_NONE, checked)
				score = -score
				total += subtotal
			}
		} else {
			score, subtotal = s.ybw(brd, stk, -beta, -alpha, rDepth-1, ply+1, childType, SP_NONE, checked)
			score = -score
			total += subtotal
			// re-search reduced moves that fail high at full depth.
			if rDepth < depth && score > alpha {
				score, subtotal = s.ybw(brd, stk, -beta, -alpha, depth-1, ply+1, childType, SP_NONE, checked)
				score = -score
				total += subtotal
			}
		}

		unmakeMove(brd, m, memento)

		if brd.worker.IsCancelled() {
			switch spType {
			case SP_MASTER:
				sp.Lock()
				if sp.cancel { // A servant has found a cutoff
					best, bestMove, sum = sp.best, sp.bestMove, sp.nodeCount
					sp.Unlock()
					loadBalancer.RemoveSP(brd.worker)
					// the servant that found the cutoff has already stored the cutoff info.
					mainTt.store(brd, bestMove, depth, LOWER_BOUND, best)
					return best, sum
				} else { // A cutoff has been found somewhere above this SP.
					sp.cancel = true
					sp.Unlock()
					loadBalancer.RemoveSP(brd.worker)
					return NO_SCORE, sum
				}
			case SP_SERVANT:
				return NO_SCORE, sum // servant aborts its search and reports the nodes searched as overhead.
			case SP_NONE:
				// selector.Recycle(recycler)
				recycler.RecycleMoveSelector(selector)
				return NO_SCORE, sum
			default:
				s.sendInfo("unknown SP type\n")
			}
		}

		if spType != SP_NONE {
			sp.Lock() // get the latest info under lock protection
			alpha, beta, best, bestMove = sp.alpha, sp.beta, sp.best, sp.bestMove
			if nodeType == Y_PV {
				pv = thisStk.pv
				stk[ply].pv = pv
			}

			sp.legalSearched += 1
			sp.nodeCount += total
			legalSearched, sum = sp.legalSearched, sp.nodeCount

			if score > best {
				bestMove, sp.bestMove, best, sp.best = m, m, score, score
				if nodeType == Y_PV {
					pv.m, pv.value, pv.depth, pv.next = m, score, depth, stk[ply+1].pv
					thisStk.pv = pv
					stk[ply].pv = pv
				}
				if score > alpha {
					alpha, sp.alpha = score, score
					if score >= beta {
						storeCutoff(&stk[ply], &s.htable, m, brd.c, total)
						sp.cancel = true
						sp.Unlock()
						if spType == SP_MASTER {
							loadBalancer.RemoveSP(brd.worker)
							mainTt.store(brd, m, depth, LOWER_BOUND, score)
							// selector.Recycle(recycler)
							return score, sum
						} else { // sp_type == SP_SERVANT
							return NO_SCORE, 0
						}
					}
				}
			}
			sp.Unlock()
		} else { // sp_type == SP_NONE
			sum += total
			if score > best {
				if nodeType == Y_PV {
					pv.m, pv.value, pv.depth, pv.next = m, score, depth, stk[ply+1].pv
					thisStk.pv = pv
				}
				if score > alpha {
					if score >= beta {
						storeCutoff(thisStk, &s.htable, m, brd.c, total) // what happens on refutation of main pv?
						mainTt.store(brd, m, depth, LOWER_BOUND, score)
						// selector.Recycle(recycler)
						recycler.RecycleMoveSelector(selector)
						return score, sum
					}
					alpha = score
				}
				bestMove, best = m, score
			}
			legalSearched += 1
			// Determine if this would be a good location to begin searching in parallel.
			if canSplit(brd, ply, depth, nodeType, legalSearched, stage) {
				sp = CreateSP(s, brd, stk, selector, bestMove, alpha, beta, best, depth, ply,
					legalSearched, nodeType, sum, checked)
				// register the split point in the appropriate SP list, and notify any idle workers.
				loadBalancer.AddSP(brd.worker, sp)
				thisStk = sp.thisStk
				spType = SP_MASTER
			}
		}

	} // end of moves loop

	switch spType {
	case SP_MASTER:
		sp.Lock()
		sp.workerFinished = true
		sp.Unlock()
		loadBalancer.RemoveSP(brd.worker)

		// Helpful Master Concept:
		// All moves at this SP may have been consumed, but servant workers may still be busy evaluating
		// subtrees rooted at this SP.  If that's the case, offer to help only those workers assigned to
		// this split point.
		brd.worker.HelpServants(sp) // Blocks until all servants have finished processing.

		sp.Lock() // make sure to capture any improvements contributed by servant workers:
		alpha, best, bestMove = sp.alpha, sp.best, sp.bestMove
		sum, legalSearched = sp.nodeCount, sp.legalSearched
		if nodeType == Y_PV {
			stk[ply].pv = thisStk.pv
		}
		sp.cancel = true
		sp.Unlock()
		// since all servants have finished processing, we can safely recycle the move buffers.
		// selector.Recycle(recycler)
	case SP_SERVANT:
		return NO_SCORE, 0
	default:
		// selector.Recycle(recycler)
		recycler.RecycleMoveSelector(selector)
	}

	if legalSearched > 0 {
		if alpha > oldAlpha {
			mainTt.store(brd, bestMove, depth, EXACT, best)
			return best, sum
		} else {
			mainTt.store(brd, bestMove, depth, UPPER_BOUND, best)
			return best, sum
		}
	} else {
		if inCheck { // Checkmate.
			mainTt.store(brd, NO_MOVE, depth, EXACT, ply-MATE)
			return ply - MATE, sum
		} else { // Draw.
			mainTt.store(brd, NO_MOVE, depth, EXACT, 0)
			return ply - DRAW_VALUE, sum
		}
	}
}

func (s *Search) nullMake(brd *Board, stk Stack, beta, nullDepth, ply int, checked bool) (int, int) {
	hashKey, enpTarget := brd.hashKey, brd.enpTarget
	brd.c ^= 1
	brd.hashKey ^= sideKey64
	brd.hashKey ^= enpZobrist(enpTarget)
	brd.enpTarget = SQ_INVALID
	stk[ply+1].inCheck = false // Impossible to give check from a legal position by standing pat.
	stk[ply+1].canNull = false
	score, sum := s.ybw(brd, stk, -beta, (-beta)+1, nullDepth-1, ply+1, Y_CUT, SP_NONE, checked)
	stk[ply+1].canNull = true
	brd.c ^= 1
	brd.hashKey = hashKey
	brd.enpTarget = enpTarget
	return -score, sum
}

func (s *Search) determineChildType(nodeType, legalSearched int) int {
	switch nodeType {
	case Y_PV:
		if legalSearched == 0 {
			return Y_PV
		} else {
			return Y_CUT
		}
	case Y_CUT:
		if legalSearched == 0 {
			return Y_ALL
		} else {
			return Y_CUT
		}
	case Y_ALL:
		return Y_CUT
	default:
		s.sendInfo("Invalid node type detected.\n")
		return nodeType
	}
}

// Determine if the current node is a good place to start searching in parallel.
func canSplit(brd *Board, ply, depth, nodeType, legalSearched, stage int) bool {
	if depth >= MIN_SPLIT {
		switch nodeType {
		case Y_PV:
			return ply > 0 && legalSearched > 0
		case Y_CUT:
			return legalSearched > 6 && stage >= STAGE_REMAINING
		case Y_ALL:
			return legalSearched > 1
		}
	}
	return false
}

func storeCutoff(thisStk *StackItem, htable *HistoryTable, m Move, c uint8, total int) {
	if m.IsQuiet() {
		htable.Store(m, c, total)
		thisStk.StoreKiller(m) // store killer moves in stack for this Goroutine.
	}
}
