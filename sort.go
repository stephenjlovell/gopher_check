//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
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

import (
// "container/heap"
)

// using a Pair to associate moves with their ordering will incur substantial GC overhead,
// since each new struct will be allocated on the heap...

type SortItem struct {
	move     Move
	priority int
}

type MoveList []SortItem

func (l MoveList) Len() int { return len(l) }

func (l MoveList) Less(i, j int) bool { return l[i].priority < l[j].priority }

func (l MoveList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
	// l[i].index, l[j].index = j, i
}

func (l *MoveList) Push(sort_item interface{}) {
	// n := len(*l)
	item := sort_item.(SortItem)
	*l = append(*l, item)
}

func (l *MoveList) Pop() interface{} {
	old := *l
	n := len(old)
	// if n > 0 {
	item := old[n-1] // not safe for zero-length slice.
	*l = old[0 : n-1]
	return item
	// } else {
	//   return nil
	// }

}

// SEE score can be stored in 16 bits (max value +/- 20000).  Could add INF to
// score to make it always positive.

func mvv_lva(victim, attacker Piece) int { // returns value between 0 and 64
	return int((victim << 3) | attacker)
}
