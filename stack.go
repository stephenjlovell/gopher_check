//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

// "fmt"

const (
	MAX_STACK = 128
)

// Stack provides a way for worker goroutines to reach data from above or below the current ply
// that would otherwise be trapped in the call stack due to recursion.
type Stack []StackItem

type StackItem struct {
	killers      KEntry
	hashKey      uint64 // use hash key to search for repetitions
	pv           *PV
	singularMove Move
	eval         int16
	inCheck      bool
	canNull      bool
}

func (thisStk *StackItem) Copy() *StackItem {
	copy := *thisStk
	return &copy
}

func NewStack() Stack {
	stk := make(Stack, MAX_STACK, MAX_STACK)
	for i := 0; i < MAX_STACK; i++ {
		stk[i].canNull = true
		stk[i].singularMove = NO_MOVE
	}
	return stk
}

func (stk Stack) CopyUpTo(otherStk Stack, ply int) {
	for i := 0; i < ply; i++ {
		// other_stk[i].sp = stk[i].sp
		// other_stk[i].value = stk[i].value
		// other_stk[i].eval = stk[i].eval
		// other_stk[i].pv_move = stk[i].pv_move
		// other_stk[i].killers = stk[i].killers
		otherStk[i].hashKey = stk[i].hashKey
		// other_stk[i].depth = stk[i].depth
		// other_stk[i].in_check = stk[i].in_check
	}
}
