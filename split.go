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
  SP_NONE = iota
  SP_SERVANT
  SP_MASTER
)

type SplitPoint struct {
	sync.Mutex

	selector          *MoveSelector
	parent            *SplitPoint
  master            *Worker

	brd               *Board
  this_stk          *StackItem

  depth             int
  ply               int
  extensions_left   int
  can_null          bool
  node_type         int

	alpha             int // shared
	beta              int // shared
	best              int // shared
  best_move         Move // shared

	node_count        int    // shared
  legal_searched    int
	cancel            chan bool
}

type SPCancellation struct {
  sp *SplitPoint
  hash_key uint64
}


type SPListItem struct {
  sp *SplitPoint
  stk Stack
  index uint8
  order uint8
  worker_mask uint8  // this will limit us to no more than 8 Worker goroutines...
}

type SPList []*SPListItem

func (l SPList) Len() int { return len(l) }

func (l SPList) Less(i, j int) bool { return l[i].order < l[j].order }

func (l SPList) Swap(i, j int) {
  l[i], l[j] = l[j], l[i]
  l[i].index, l[j].index = uint8(j), uint8(i)
}

func (l *SPList) Push(li interface{}) {
  n := len(*l)
  item := li.(*SPListItem)
  item.index = uint8(n)
  *l = append(*l, item)
}

func (l *SPList) Pop() interface{} {
  old := *l
  n := len(old)
  item := old[n-1]
  *l = old[0 : n-1]
  return item
}







