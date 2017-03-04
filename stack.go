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

// ThinStackItem provides a lighter version of Stack for use with SPs.

type ThinStack []ThinStackItem

type ThinStackItem struct {
	hashKey uint64
	eval    int16
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

func CopyToThinStack(from Stack, to ThinStack, upTo int) {
	for i := 0; i < upTo; i++ {
		to[i].eval = from[i].eval
		to[i].hashKey = from[i].hashKey
	}
}

func CopyToStack(from ThinStack, to Stack, upTo int) {
	for i := 0; i < upTo; i++ {
		to[i].eval = from[i].eval
		to[i].hashKey = from[i].hashKey
	}
}
