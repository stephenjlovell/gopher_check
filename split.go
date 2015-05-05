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
	wg sync.WaitGroup

	selector     *MoveSelector
	parent       *SplitPoint
	master       *Worker
	servant_mask uint8

	brd      *Board
	this_stk *StackItem

	depth           int
	ply             int
	extensions_left int
	can_null        bool
	node_type       int

	alpha     int  // shared
	beta      int  // shared
	best      int  // shared
	best_move Move // shared

	node_count     int // shared
	legal_searched int
	cancel         chan bool
}

func (sp *SplitPoint) Order() int {
	return (sp.depth << 11) | (sp.legal_searched << 3) | (sp.node_type)
}

func (sp *SplitPoint) ServantMask() uint8 {
	sp.Lock()
	servant_mask := sp.servant_mask
	sp.Unlock()
	return servant_mask
}
