//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "sync"

type Recycler struct {
	qMoveSelectors []*QMoveSelector
	moveSelectors  []*MoveSelector
	sync.Mutex
}

const (
	DEFAULT_MOVE_LIST_LENGTH = 12
	QUIET_MOVE_LIST_LENGTH   = 20
	DEFAULT_RECYCLER_SIZE    = 256
)

func NewRecycler() *Recycler {
	r := &Recycler{
		moveSelectors:  make([]*MoveSelector, DEFAULT_RECYCLER_SIZE>>2, DEFAULT_RECYCLER_SIZE>>1),
		qMoveSelectors: make([]*QMoveSelector, DEFAULT_RECYCLER_SIZE>>3, DEFAULT_RECYCLER_SIZE>>2),
	}
	r.init()
	return r
}

func (r *Recycler) init() {
	for i := range r.qMoveSelectors {
		r.qMoveSelectors[i] = &QMoveSelector{
			winning:        NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
			losing:         NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
			checks:         NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
			remainingMoves: NewMoveList(QUIET_MOVE_LIST_LENGTH),
		}
	}
	for i := range r.moveSelectors {
		r.moveSelectors[i] = &MoveSelector{
			winning:        NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
			losing:         NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
			remainingMoves: NewMoveList(QUIET_MOVE_LIST_LENGTH),
		}
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
		selector.brd, selector.htable = brd, htable
		selector.inCheck, selector.canCheck = inCheck, canCheck
		selector.winning = selector.winning[0:0]
		selector.losing = selector.losing[0:0]
		selector.checks = selector.checks[0:0]
		selector.remainingMoves = selector.remainingMoves[0:0]
		selector.stage, selector.finished, selector.index = 0, 0, 0
		return selector
	} else {
		return NewQMoveSelector(brd, htable, inCheck, canCheck)
	}
}

func (r *Recycler) RecycleMoveSelector(selector *MoveSelector) {
	if len(r.moveSelectors) < cap(r.moveSelectors) {
		r.moveSelectors = append(r.moveSelectors, selector)
	}
}

func (r *Recycler) ReuseMoveSelector(brd *Board, htable *HistoryTable,
	inCheck bool, firstMove Move, killers KEntry) *MoveSelector {
	if len(r.moveSelectors) > 0 {
		var selector *MoveSelector
		selector, r.moveSelectors = r.moveSelectors[len(r.moveSelectors)-1], r.moveSelectors[:len(r.moveSelectors)-1]
		selector.brd, selector.htable = brd, htable
		selector.killers = killers
		selector.inCheck = inCheck
		selector.firstMove = firstMove
		selector.winning = selector.winning[0:0]
		selector.losing = selector.losing[0:0]
		selector.remainingMoves = selector.remainingMoves[0:0]
		selector.stage, selector.finished, selector.index = 0, 0, 0
		return selector
	} else {
		return NewMoveSelector(brd, htable, inCheck, firstMove, killers)
	}
}
