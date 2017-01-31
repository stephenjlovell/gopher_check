//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
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

	depth int
	ply   int
	// extensionsLeft int  // TODO: verify if extension counter needs lock protection.
	nodeType      int
	alpha         int // shared
	beta          int // shared
	best          int // shared
	nodeCount     int // shared
	legalSearched int

	bestMove Move // shared

	s        *Search
	selector *MoveSelector
	parent   *SplitPoint
	master   *Worker
	brd      *Board
	thisStk  *StackItem
	cond     *sync.Cond

	servantMask uint8

	cancel         bool
	workerFinished bool
	canNull        bool
	checked        bool
}

func (sp *SplitPoint) Wait() {
	sp.cond.L.Lock()
	for sp.servantMask > 0 {
		sp.cond.Wait() // unlocks, sleeps thread, then locks sp.cond.L
	}
	sp.cond.L.Unlock()
}

func (sp *SplitPoint) Order() int {
	sp.RLock()
	searched := sp.legalSearched
	nodeType := sp.nodeType
	sp.RUnlock()
	return (max(searched, 16) << 3) | nodeType
}

func (sp *SplitPoint) WorkerFinished() bool {
	sp.RLock()
	finished := sp.workerFinished
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
	servantMask := sp.servantMask
	sp.cond.L.Unlock()
	return servantMask
}

func (sp *SplitPoint) AddServant(wMask uint8) {
	sp.cond.L.Lock()
	sp.servantMask |= wMask
	sp.cond.L.Unlock()
}

func (sp *SplitPoint) RemoveServant(wMask uint8) {
	sp.cond.L.Lock()
	sp.servantMask &= (^wMask)
	sp.cond.L.Unlock()

	sp.Lock()
	sp.workerFinished = true
	sp.Unlock()

	sp.cond.Signal()
}

func CreateSP(s *Search, brd *Board, stk Stack, ms *MoveSelector, bestMove Move, alpha, beta, best,
	depth, ply, legalSearched, nodeType, sum int, checked bool) *SplitPoint {

	sp := &SplitPoint{
		cond:     sync.NewCond(new(sync.Mutex)),
		selector: ms,
		master:   brd.worker,
		parent:   brd.worker.currentSp,

		brd:     brd.Copy(),
		thisStk: stk[ply].Copy(),
		s:       s,

		depth: depth,
		ply:   ply,

		nodeType: nodeType,

		alpha:    alpha,
		beta:     beta,
		best:     best,
		bestMove: bestMove,

		checked: checked,

		nodeCount:     sum,
		legalSearched: legalSearched,
		cancel:        false,
	}

	sp.stk = make(Stack, ply, ply)
	stk.CopyUpTo(sp.stk, ply)

	ms.brd = sp.brd // make sure the move selector points to the static SP board.
	ms.thisStk = sp.thisStk
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
