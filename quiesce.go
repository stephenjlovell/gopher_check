//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

// Q-Search will always be done sequentially: Q-search subtrees are taller and narrower than in the main search,
// making benefit of parallelism smaller and raising communication and synchronization overhead.
func (s *Search) quiescence(brd *Board, stk Stack, alpha, beta, depth, ply int) (int, int) {

	thisStk := &stk[ply]

	thisStk.hashKey = brd.hashKey
	if stk.IsRepetition(ply, brd.halfmoveClock) { // check for draw by threefold repetition
		return ply - DRAW_VALUE, 1
	}

	inCheck := thisStk.inCheck
	if brd.halfmoveClock >= 100 {
		if isCheckmate(brd, inCheck) {
			return ply - MATE, 1
		} else {
			return ply - DRAW_VALUE, 1
		}
	}

	best, sum := -INF, 1
	var score, total int

	if !inCheck {
		score = evaluate(brd, alpha, beta) // stand pat
		thisStk.eval = int16(score)
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

	legalMoves := false
	memento := brd.NewMemento()
	recycler := brd.worker.recycler
	selector := recycler.ReuseQMoveSelector(brd, thisStk, &s.htable, inCheck, depth >= MIN_CHECK_DEPTH)

	var mayPromote, givesCheck bool
	for m := selector.Next(); m != NO_MOVE; m = selector.Next() {

		mayPromote = brd.MayPromote(m)

		makeMove(brd, m)

		givesCheck = brd.InCheck()

		if !inCheck && !givesCheck && !mayPromote && alpha > -MIN_MATE &&
			best+m.CapturedPiece().Value()+ROOK_VALUE < alpha {
			unmakeMove(brd, m, memento)
			continue
		}

		stk[ply+1].inCheck = givesCheck // avoid having to recalculate in_check at beginning of search.

		score, total = s.quiescence(brd, stk, -beta, -alpha, depth-1, ply+1)
		score = -score
		sum += total
		unmakeMove(brd, m, memento)

		if score > best {
			if score > alpha {
				if score >= beta {
					recycler.RecycleQMoveSelector(selector)
					return score, sum
				}
				alpha = score
			}
			best = score
		}
		legalMoves = true
	}

	recycler.RecycleQMoveSelector(selector)
	if inCheck && !legalMoves {
		return ply - MATE, 1 // detect checkmate.
	}
	return best, sum
}
