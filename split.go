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
	sync.RWMutex
	// wg 								sync.WaitGroup
	stk Stack

	depth           int
	ply             int
	extensions_left int
	node_type       int
	alpha           int // shared
	beta            int // shared
	best            int // shared
	node_count      int // shared
	legal_searched  int

	best_move Move // shared

	selector *MoveSelector
	parent   *SplitPoint
	master   *Worker
	brd      *Board
	this_stk *StackItem
	cond     *sync.Cond

	servant_mask uint8

	cancel          bool
	worker_finished bool
	can_null        bool
	checked         bool
}

func (sp *SplitPoint) Wait() {
	sp.cond.L.Lock()
	for sp.servant_mask > 0 {
		sp.cond.Wait() // unlocks, sleeps thread, then locks sp.cond.L
	}
	sp.cond.L.Unlock()
}

func (sp *SplitPoint) Order() int {
	sp.RLock()
	searched := sp.legal_searched
	node_type := sp.node_type
	sp.RUnlock()

	return (max(searched, 16) << 3) | node_type
}

func (sp *SplitPoint) WorkerFinished() bool {
	sp.RLock()
	finished := sp.worker_finished
	sp.RUnlock()
	return finished
}

func (sp *SplitPoint) Cancel() bool {
	sp.RLock()
	cancel := sp.cancel
	sp.RUnlock()
	return cancel
}

func (sp *SplitPoint) HelpWanted() bool {
	return !sp.Cancel() && sp.ServantMask() > 0
}

func (sp *SplitPoint) ServantMask() uint8 {
	sp.cond.L.Lock()
	servant_mask := sp.servant_mask
	sp.cond.L.Unlock()
	return servant_mask
}

func (sp *SplitPoint) AddServant(w_mask uint8) {
	sp.cond.L.Lock()
	sp.servant_mask |= w_mask
	sp.cond.L.Unlock()
}

func (sp *SplitPoint) RemoveServant(w_mask uint8) {
	sp.cond.L.Lock()
	sp.servant_mask &= (^w_mask)
	sp.cond.L.Unlock()

	sp.Lock()
	sp.worker_finished = true
	sp.Unlock()

	sp.cond.Signal()
}

func CreateSP(brd *Board, stk Stack, ms *MoveSelector, best_move Move, alpha, beta, best, depth, ply,
	legal_searched, node_type, sum int, checked bool) *SplitPoint {

	sp := &SplitPoint{
		cond:     sync.NewCond(new(sync.Mutex)),
		selector: ms,
		master:   brd.worker,
		parent:   brd.worker.current_sp,

		brd:      brd.Copy(),
		this_stk: stk[ply].Copy(),

		depth: depth,
		ply:   ply,

		node_type: node_type,

		alpha:     alpha,
		beta:      beta,
		best:      best,
		best_move: best_move,

		checked: checked,

		node_count:     sum,
		legal_searched: legal_searched,
		cancel:         false,
	}

	sp.stk = make(Stack, ply, ply)
	stk.CopyUpTo(sp.stk, ply)

	ms.brd = sp.brd // make sure the move selector points to the static SP board.
	ms.this_stk = sp.this_stk
	stk[ply].sp = sp

	return sp
}

type SPList []*SplitPoint

func (l *SPList) Push(sp *SplitPoint) {
	*l = append(*l, sp)
}

func (l *SPList) Pop() *SplitPoint {
	old := *l
	n := len(old)
	sp := old[n-1]
	*l = old[0 : n-1]
	return sp
}
