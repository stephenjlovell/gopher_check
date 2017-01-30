//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	// "fmt"
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

type Worker struct {
	sync.RWMutex
	searchOverhead int

	spList SPList
	stk     Stack

	assignSp chan *SplitPoint

	ptt        *PawnTT
	recycler   *Recycler
	currentSp *SplitPoint

	mask  uint8
	index uint8
}

func (w *Worker) IsCancelled() bool {
	for sp := w.currentSp; sp != nil; sp = sp.parent {
		if sp.Cancel() {
			return true
		}
	}
	return false
}

func (w *Worker) HelpServants(currentSp *SplitPoint) {
	var bestSp *SplitPoint
	var worker *Worker
	// Check for SPs underneath current_sp

	// assert(w.current_sp == current_sp.parent, "not current sp")

	for mask := currentSp.ServantMask(); mask > 0; mask = currentSp.ServantMask() {
		bestSp = nil

		for tempMask := mask; tempMask > 0; tempMask &= (^worker.mask) {
			worker = loadBalancer.workers[lsb(BB(tempMask))]
			worker.RLock()
			for _, thisSp := range worker.spList {
				// If a worker has already finished searching, then either a beta cutoff has already
				// occurred at sp, or no moves are left to search.
				if !thisSp.WorkerFinished() && (bestSp == nil || thisSp.Order() > bestSp.Order()) {
					bestSp = thisSp
					tempMask |= thisSp.ServantMask() // If this SP has servants of its own, check them as well.
				}
			}
			worker.RUnlock()
		}

		if bestSp == nil || bestSp.WorkerFinished() {
			break
		} else {
			bestSp.AddServant(w.mask)
			w.currentSp = bestSp
			w.SearchSP(bestSp)
		}
	}

	w.currentSp = currentSp.parent

	// If at any point we can't find another viable servant SP, wait for remaining servants to complete.
	// This prevents us from continually acquiring the worker locks.
	currentSp.Wait()
}

func (w *Worker) Help(b *Balancer) {
	go func() {
		var bestSp *SplitPoint
		for {

			bestSp = nil
			for _, master := range b.workers { // try to find a good SP
				if master.index == w.index {
					continue
				}
				master.RLock()
				for _, thisSp := range master.spList {
					if !thisSp.WorkerFinished() && (bestSp == nil || thisSp.Order() > bestSp.Order()) {
						bestSp = thisSp
					}
				}
				master.RUnlock()
			}

			if bestSp == nil || bestSp.WorkerFinished() { // No best SP was available.
				b.done <- w             // Worker is completely idle and available to help any processor.
				bestSp = <-w.assignSp // Wait for the next SP to be discovered.
			} else {
				bestSp.AddServant(w.mask)
			}

			w.currentSp = bestSp
			w.SearchSP(bestSp)
			w.currentSp = nil

		}
	}()
}

func (w *Worker) SearchSP(sp *SplitPoint) {
	brd := sp.brd.Copy()
	brd.worker = w

	sp.stk.CopyUpTo(w.stk, sp.ply)
	w.stk[sp.ply].sp = sp

	sp.RLock()
	alpha, beta := sp.alpha, sp.beta
	sp.RUnlock()

	// Once the SP is fully evaluated, The SP master will handle returning its value to parent node.
	_, total := sp.s.ybw(brd, w.stk, alpha, beta, sp.depth, sp.ply, sp.nodeType, SP_SERVANT, sp.checked)
	w.searchOverhead += total

	sp.RemoveServant(w.mask)
	// At this point, any additional SPs found by the worker during the search rooted at sp
	// should be fully resolved.  The SP list for this worker should be empty again.
}
