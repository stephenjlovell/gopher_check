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
	hashKey      uint64 // use hash key to search for repetitions
	eval          int
	killers       KEntry
	singularMove Move

	sp *SplitPoint
	pv *PV

	inCheck bool
	canNull bool
}

func (thisStk *StackItem) Copy() *StackItem {
	return &StackItem{
		// split point is not copied over.
		pv:            thisStk.pv,
		killers:       thisStk.killers,
		singularMove: thisStk.singularMove,
		eval:          thisStk.eval,
		hashKey:      thisStk.hashKey,
		inCheck:      thisStk.inCheck,
		canNull:      thisStk.canNull,
	}
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
		// otherStk[i].sp = stk[i].sp
		// otherStk[i].value = stk[i].value
		// otherStk[i].eval = stk[i].eval
		// otherStk[i].pvMove = stk[i].pvMove
		// otherStk[i].killers = stk[i].killers
		otherStk[i].hashKey = stk[i].hashKey
		// otherStk[i].depth = stk[i].depth
		// otherStk[i].inCheck = stk[i].inCheck
	}
}
