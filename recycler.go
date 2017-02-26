//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "sync"

type Recycler struct {
	stack []MoveList
	sync.Mutex
}

const (
	DEFAULT_MOVE_LIST_LENGTH = 12
	QUIET_MOVE_LIST_LENGTH   = 20
)

func NewRecycler(capacity uint64) *Recycler {
	r := &Recycler{
		stack: make([]MoveList, capacity/2, capacity),
	}
	r.init()
	return r
}

func (r *Recycler) init() {
	for i := 0; i < len(r.stack); i++ {
		r.stack[0] = NewMoveList(QUIET_MOVE_LIST_LENGTH)
	}
}

func (r *Recycler) Recycle(moves MoveList) {
	if len(r.stack) < cap(r.stack) {
		r.stack = append(r.stack, moves)
	}
}

func (r *Recycler) AttemptReuse(length int) MoveList {
	var moves MoveList
	if len(r.stack) > 0 {
		moves, r.stack = r.stack[len(r.stack)-1], r.stack[:len(r.stack)-1]
	}
	if moves == nil {
		moves = NewMoveList(length)
	}
	return moves
}
