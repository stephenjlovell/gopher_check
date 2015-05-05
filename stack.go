//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//-----------------------------------------------------------------------------------

package main

const (
	MAX_STACK = 128
)

type Stack []StackItem

type StackItem struct {
	sp *SplitPoint

	// pv_move      Move
	// value int
	// depth int
	pv PV

	killers  KEntry
	eval     int
	hash_key uint64 // use hash key to search for repetitions
	in_check bool
}

func (this_stk *StackItem) Copy() *StackItem {
	return &StackItem{
		// sp: this_stk.sp,
		pv: this_stk.pv,

		killers:  this_stk.killers,
		eval:     this_stk.eval,
		hash_key: this_stk.hash_key,
		in_check: this_stk.in_check,
	}
}

func NewStack() Stack {
	return make(Stack, MAX_STACK, MAX_STACK)
}

func (stk Stack) CopyUpTo(other_stk Stack, ply int) {
	for i := 0; i <= ply; i++ {
		this_stk := &stk[i]
		this_cpy := &other_stk[i]

		// this_cpy.sp = this_stk.sp
		// this_cpy.value = this_stk.value
		// this_cpy.eval = this_stk.eval
		// this_cpy.pv_move = this_stk.pv_move
		// this_cpy.killers = this_stk.killers
		this_cpy.hash_key = this_stk.hash_key
		// this_cpy.depth = this_stk.depth
		// this_cpy.in_check = this_stk.in_check
	}
}
