//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	// "fmt"
	"runtime"
	"sync"
)

// Each worker maintains a list of active split points for which it is responsible.

// When a worker's search reaches a new SP node, it creates a new SP struct, (including the current
// []Stack info for nodes above the SP) and adds the SP to its active SP list.

// When workers are idle (they've finished searching and have no split points of their own),
// they request more work from the load balancer. The load balancer selects the best
// available SP and assigns the worker to the SP.

// The assigned worker begins a search rooted at the chosen SP node. Each worker searching the SP
// node requests moves from the SP node's move generator.

// Cancellation:

// When a beta cutoff occurs at an SP node, the worker sends a cancellation signal on a channel
// read by the other workers collaborating on the current split point.
// If there are more SPs below the current one, the cancellation signal will be fanned out to
// each child SP.
const (
	MAX_WORKERS = 32
)

type Worker struct {
	sync.RWMutex
	searchOverhead int

	spList SPList
	stk    Stack

	assignSp chan *SplitPoint

	ptt       *PawnTT
	recycler  *Recycler
	currentSP *SplitPoint

	mask  uint32
	index uint8
}

func MaxWorkers() int {
	return Min(MAX_WORKERS, runtime.NumCPU())
}

func (w *Worker) IsCancelled() bool {
	for sp := w.currentSP; sp != nil; sp = sp.parent {
		if sp.Cancel() {
			return true
		}
	}
	return false
}

// HelpServants is called by the master currentSP when there are no moves remaining for the master
// to search, but there are still servant workers searching moves at currentSP.
// Rather than simply sleeping until all servant workers are finished, the master worker looks for
// any viable split points underneath currentSP at which to help one of it's servant workers. If
// a good place to help is found, the master of currentSP temporarily assists its servant worker(s)
// until all servants are finished processing at currentSP.
func (w *Worker) HelpServants(currentSP *SplitPoint) {
	var bestSp *SplitPoint
	var worker *Worker

	currentSP.mu.Lock()
	currentSP.workerFinished = true // no need to update sp.servantMask
	currentSP.mu.Unlock()

	loadBalancer.RemoveSP(w)

	var bestOrder, thisOrder int
	for mask := currentSP.ServantMask(); mask > 0; mask = currentSP.ServantMask() {
		bestSp = nil

		for tempMask := mask; tempMask > 0; tempMask &= (^worker.mask) {
			worker = loadBalancer.workers[lsb(BB(tempMask))]
			worker.RLock()
			for _, thisSp := range worker.spList {
				thisOrder = thisSp.Order()
				// If a worker has already finished searching, then either a beta cutoff has already
				// occurred at sp, or no moves are left to search.
				if thisSp.HelpWanted() && (bestSp == nil || thisOrder > bestOrder) {
					bestSp = thisSp
					bestOrder = thisOrder
					tempMask |= thisSp.ServantMask() // If this SP has servants of its own, check them as well.
				}
			}
			worker.RUnlock()
		}

		if bestSp == nil {
			break
		} else {
			w.SearchSP(bestSp)
		}
	}

	w.currentSP = currentSP.parent

	// If at any point we can't find another viable servant SP, wait for remaining servants to complete.
	// This prevents us from continually acquiring the worker locks.
	currentSP.Wait()
}

func (w *Worker) Help(b *Balancer) {
	go func() {
		var bestSp *SplitPoint
		var bestOrder, thisOrder int
		for {
			w.currentSP = nil
			bestSp = nil
			for _, master := range b.workers { // try to find a good SP
				if master.index == w.index {
					continue
				}
				master.RLock()
				for _, thisSp := range master.spList {
					thisOrder = thisSp.Order()
					if thisSp.HelpWanted() && (bestSp == nil || thisOrder > bestOrder) {
						bestSp = thisSp
						bestOrder = thisOrder
					}
				}
				master.RUnlock()
			}

			if bestSp == nil { // No best SP was available.
				b.done <- w           // Worker is completely idle and available to help any processor.
				bestSp = <-w.assignSp // Wait for the next SP to be discovered.
			}
			w.SearchSP(bestSp)
		}
	}()
}

func (w *Worker) SearchSP(sp *SplitPoint) {
	sp.AddServant(w.mask)
	w.currentSP = sp

	brd := sp.brd.Copy()
	brd.worker = w

	CopyToStack(sp.stk, w.stk, sp.ply)

	sp.mu.RLock()
	alpha, beta := sp.alpha, sp.beta
	sp.mu.RUnlock()

	// Once the SP is fully evaluated, The SP master will handle returning its value to parent node.
	_, total := sp.s.ybw(brd, alpha, beta, sp.depth, sp.ply, sp.nodeType, SP_SERVANT, sp.checked)
	w.searchOverhead += total

	sp.RemoveServant(w.mask)
	// At this point, any additional SPs found by the worker during the search rooted at sp
	// should be fully resolved.  The SP list for this worker should be empty again.
}
