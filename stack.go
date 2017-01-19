//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

// "fmt"

const (
	MAX_STACK = 128
)

type Stack []StackItem

type StackItem struct {
	hash_key      uint64 // use hash key to search for repetitions
	eval          int
	killers       KEntry
	singular_move Move

	sp *SplitPoint
	pv *PV

	in_check bool
	can_null bool
}

func (this_stk *StackItem) Copy() *StackItem {
	return &StackItem{
		// split point is not copied over.
		pv:            this_stk.pv,
		killers:       this_stk.killers,
		singular_move: this_stk.singular_move,
		eval:          this_stk.eval,
		hash_key:      this_stk.hash_key,
		in_check:      this_stk.in_check,
		can_null:      this_stk.can_null,
	}
}

func NewStack() Stack {
	stk := make(Stack, MAX_STACK, MAX_STACK)
	for i := 0; i < MAX_STACK; i++ {
		stk[i].can_null = true
		stk[i].singular_move = NO_MOVE
	}
	return stk
}

func (stk Stack) CopyUpTo(other_stk Stack, ply int) {
	for i := 0; i < ply; i++ {
		// other_stk[i].sp = stk[i].sp
		// other_stk[i].value = stk[i].value
		// other_stk[i].eval = stk[i].eval
		// other_stk[i].pv_move = stk[i].pv_move
		// other_stk[i].killers = stk[i].killers
		other_stk[i].hash_key = stk[i].hash_key
		// other_stk[i].depth = stk[i].depth
		// other_stk[i].in_check = stk[i].in_check
	}
}
