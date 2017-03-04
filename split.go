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
	mu                               sync.RWMutex // 24 (bytes)
	stk                              Stack        // 12
	depth, ply, nodeType, nodeCount  int          // 8 x 8
	alpha, beta, best, legalSearched int
	s                                *Search // 8 x 7
	selector                         *MoveSelector
	parent                           *SplitPoint
	master                           *Worker
	brd                              *Board
	thisStk                          *StackItem
	cond                             *sync.Cond
	bestMove                         Move // 4
	servantMask                      uint32
	cancel, workerFinished, checked  bool
	// extensionsLeft int  // TODO: verify if extension counter needs lock protection.
}

func (sp *SplitPoint) Wait() {
	sp.mu.Lock()
	for sp.servantMask > 0 {
		sp.cond.Wait() // unlocks, sleeps thread, then locks sp.cond.L
	}
	sp.mu.Unlock()
}

func (sp *SplitPoint) Order() int {
	sp.mu.RLock()
	searched := sp.legalSearched
	nodeType := sp.nodeType
	sp.mu.RUnlock()
	return (max(searched, 16) << 2) | nodeType
}

func (sp *SplitPoint) WorkerFinished() bool {
	sp.mu.RLock()
	finished := sp.workerFinished
	sp.mu.RUnlock()
	return finished
}

func (sp *SplitPoint) Cancel() bool {
	sp.mu.RLock()
	cancel := sp.cancel
	sp.mu.RUnlock()
	return cancel
}

func (sp *SplitPoint) HelpWanted() bool {
	// return !sp.Cancel() && sp.ServantMask() > 0
	sp.mu.RLock()
	cancel := sp.cancel
	servantMask := sp.servantMask
	sp.mu.RUnlock()
	return !cancel && servantMask > 0
}

func (sp *SplitPoint) ServantMask() uint32 {
	// sp.cond.L.Lock()
	sp.mu.Lock()
	servantMask := sp.servantMask
	// sp.cond.L.Unlock()
	sp.mu.Unlock()
	return servantMask
}

func (sp *SplitPoint) AddServant(wMask uint32) {
	// sp.cond.L.Lock()
	sp.mu.Lock()
	sp.servantMask |= wMask
	// sp.cond.L.Unlock()
	sp.mu.Unlock()
}

func (sp *SplitPoint) RemoveServant(wMask uint32) {
	// sp.cond.L.Lock()
	// sp.servantMask &= (^wMask)
	// sp.cond.L.Unlock()
	sp.mu.Lock()
	sp.servantMask &= (^wMask)
	sp.workerFinished = true
	sp.mu.Unlock()

	sp.cond.Signal() // there should only ever be one sp master sleeping & awaiting this signal.
}

func CreateSP(s *Search, brd *Board, stk Stack, ms *MoveSelector, bestMove Move, alpha, beta, best,
	depth, ply, legalSearched, nodeType, sum int, checked bool) *SplitPoint {
	sp := &SplitPoint{
		mu:       sync.RWMutex{},
		selector: ms,
		master:   brd.worker,
		parent:   brd.worker.currentSp,

		brd:     brd.Copy(),
		thisStk: stk[ply].Copy(),

		s:             s,
		depth:         depth,
		ply:           ply,
		nodeType:      nodeType,
		alpha:         alpha,
		beta:          beta,
		best:          best,
		bestMove:      bestMove,
		checked:       checked,
		nodeCount:     sum,
		legalSearched: legalSearched,
		cancel:        false,
	}

	sp.cond = sync.NewCond(&sp.mu)

	// TODO: If possible, recycle this slice when SplitPoint is discarded.
	sp.stk = make(Stack, ply, ply)
	stk.CopyUpTo(sp.stk, ply)

	ms.brd = sp.brd // make sure the move selector points to the static SP board.
	ms.thisStk = sp.thisStk

	return sp
}

type SPList []*SplitPoint

func (l *SPList) Push(sp *SplitPoint) {
	*l = append(*l, sp)
}

func (l *SPList) Pop() *SplitPoint {
	old := *l
	n := len(old)
	var sp *SplitPoint
	sp, *l = old[n-1], old[0:n-1]
	return sp
}
