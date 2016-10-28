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

import "github.com/stephenjlovell/go-datastructures/queue"

func init() {
	recycler = NewRecycler()
}

var recycler *Recycler

type Recycler struct {
	ring *queue.RingBuffer
}

func NewRecycler() *Recycler {
	r := &Recycler{
		ring: queue.NewRingBuffer(512),
	}
	r.init()
	return r
}

func (r *Recycler) init() {
	for i := uint64(0); i < 512/uint64(2); i++ {
		r.ring.Offer(NewMoveList())
	}
}

func (r *Recycler) Recycle(moves MoveList) {
	r.ring.Offer(moves)
}

func (r *Recycler) AttemptReuse() MoveList {
	moves, err := r.ring.TryGet()
	if err != nil {
		panic(err)
	}
	if moves != nil {
		// fmt.Printf("-")
		return moves.(MoveList)
	}
	// fmt.Printf("+")
	return NewMoveList()
}
