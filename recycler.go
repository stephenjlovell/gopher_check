//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "sync"

type Recycler struct {
	moveLists      []MoveList
	qMoveSelectors []*QMoveSelector
	sync.Mutex
}

const (
	DEFAULT_MOVE_LIST_LENGTH = 12
	QUIET_MOVE_LIST_LENGTH   = 20
	DEFAULT_RECYCLER_SIZE    = 256
)

func NewRecycler() *Recycler {
	r := &Recycler{
		moveLists:      make([]MoveList, DEFAULT_RECYCLER_SIZE>>1, DEFAULT_RECYCLER_SIZE),
		qMoveSelectors: make([]*QMoveSelector, DEFAULT_RECYCLER_SIZE>>3, DEFAULT_RECYCLER_SIZE>>2),
	}
	r.init()
	return r
}

func (r *Recycler) init() {
	for i := range r.moveLists {
		r.moveLists[i] = NewMoveList(QUIET_MOVE_LIST_LENGTH)
	}
	for i := range r.qMoveSelectors {
		r.qMoveSelectors[i] = &QMoveSelector{
			winning:        NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
			losing:         NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
			checks:         NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
			remainingMoves: NewMoveList(QUIET_MOVE_LIST_LENGTH),
		}
	}
}

func (r *Recycler) RecycleMoveList(moves MoveList) {
	if len(r.moveLists) < cap(r.moveLists) {
		r.moveLists = append(r.moveLists, moves)
	}
}

func (r *Recycler) ReuseMoveList(length int) MoveList {
	if len(r.moveLists) > 0 {
		var moves MoveList
		moves, r.moveLists = r.moveLists[len(r.moveLists)-1], r.moveLists[:len(r.moveLists)-1]
		return moves
	} else {
		return NewMoveList(length)
	}
}

func (r *Recycler) RecycleQMoveSelector(selector *QMoveSelector) {
	if len(r.qMoveSelectors) < cap(r.qMoveSelectors) {
		r.qMoveSelectors = append(r.qMoveSelectors, selector)
	}
}

func (r *Recycler) ReuseQMoveSelector(brd *Board, thisStk *StackItem, htable *HistoryTable,
	inCheck, canCheck bool) *QMoveSelector {

	if len(r.qMoveSelectors) > 0 {
		var selector *QMoveSelector
		selector, r.qMoveSelectors = r.qMoveSelectors[len(r.qMoveSelectors)-1], r.qMoveSelectors[:len(r.qMoveSelectors)-1]
		selector.brd, selector.thisStk, selector.htable = brd, thisStk, htable
		selector.inCheck, selector.canCheck = inCheck, canCheck
		selector.winning = selector.winning[0:0]
		selector.losing = selector.losing[0:0]
		selector.checks = selector.checks[0:0]
		selector.remainingMoves = selector.remainingMoves[0:0]
		selector.stage, selector.finished, selector.index = 0, 0, 0
		return selector
	} else {
		return NewQMoveSelector(brd, thisStk, htable, inCheck, canCheck)
	}

}
