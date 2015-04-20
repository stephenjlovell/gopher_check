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

import (
	"sync"
)

const (
	MAX_STACK = 128
)



type SplitPoint struct {
	sync.Mutex

	selector  *MoveSelector
	parent    *SplitPoint
	master    *Worker
	brd       *Board
	this_stk  *StackItem
	depth     int
	node_type int

	sort_key	int 			

	alpha int // shared
	beta  int
	best  int // shared
	ply int
	node_count           int    // shared
	
	// slave_mask           uint32 // shared
	// all_slaves_searching bool   // shared


	best_move    Move // shared
	move_count   int  // shared. number of moves fully searched so far.
	// cutoff_found bool // shared
	cancel chan bool
}


type SPListItem struct{
  sp *SplitPoint
  stk Stack
  index int
  order int
}


type SPList []*SPListItem

func (l SPList) Len() int { return len(l) }

func (l SPList) Less(i, j int) bool { return l[i].order > l[j].order }

func (l SPList) Swap(i, j int) {
  l[i], l[j] = l[j], l[i]
  l[i].index, l[j].index = j, i
}

func (l *SPList) Push(li interface{}) {
  n := len(*l)
  item := li.(*SPListItem)
  item.index = n
  *l = append(*l, item)
}

func (l *SPList) Pop() interface{} {
  old := *l
  n := len(old)
  item := old[n-1]
  *l = old[0 : n-1]
  return item
}







