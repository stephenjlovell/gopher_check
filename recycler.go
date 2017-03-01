//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "sync"

type Recycler struct {
	moveLists []MoveList
	sync.Mutex
}

const (
	DEFAULT_MOVE_LIST_LENGTH = 12
	QUIET_MOVE_LIST_LENGTH   = 20
)

func NewRecycler(capacity uint64) *Recycler {
	r := &Recycler{
		moveLists: make([]MoveList, capacity, capacity),
	}
	r.init()
	return r
}

func (r *Recycler) init() {
	for i := 0; i < len(r.moveLists); i++ {
		r.moveLists[i] = NewMoveList(QUIET_MOVE_LIST_LENGTH)
	}
}

func (r *Recycler) Recycle(moves MoveList) {
	if len(r.moveLists) < cap(r.moveLists) {
		r.moveLists = append(r.moveLists, moves)
	}
}

func (r *Recycler) AttemptReuse(length int) MoveList {
	var moves MoveList
	if len(r.moveLists) > 0 {
		moves, r.moveLists = r.moveLists[len(r.moveLists)-1], r.moveLists[:len(r.moveLists)-1]
	}
	if moves == nil {
		moves = NewMoveList(length)
	}
	return moves
}
