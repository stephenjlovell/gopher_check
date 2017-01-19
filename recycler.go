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

import "sync"

type Recycler struct {
	stack []MoveList
	sync.Mutex
}

func NewRecycler(capacity uint64) *Recycler {
	r := &Recycler{
		stack: make([]MoveList, capacity/2, capacity),
	}
	r.init()
	return r
}

func (r *Recycler) init() {
	for i := 0; i < len(r.stack); i++ {
		r.stack[0] = NewMoveList()
	}
}

func (r *Recycler) Recycle(moves MoveList) {
	if len(r.stack) < cap(r.stack) {
		r.stack = append(r.stack, moves)
	}
}

func (r *Recycler) AttemptReuse() MoveList {
	var moves MoveList
	if len(r.stack) > 0 {
		moves, r.stack = r.stack[len(r.stack)-1], r.stack[:len(r.stack)-1]
	}
	if moves == nil {
		moves = NewMoveList()
	}
	return moves
}
