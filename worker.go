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
// "fmt"
// "sync"
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
	// sync.Mutex
	search_overhead int

	sp_list SPList
	stk     Stack

	assign_sp chan *SplitPoint

	ptt        *PawnTT
	current_sp *SplitPoint

	mask  uint8
	index uint8
}

func (w *Worker) IsCancelled() bool {
	for sp := w.current_sp; sp != nil; sp = sp.parent {
		if sp.Cancel() {
			return true
		}
	}
	return false
}

func (w *Worker) HelpServants(current_sp *SplitPoint) {
	var best_sp *SplitPoint
	var worker *Worker
	// Check for SPs underneath current_sp

	for mask := current_sp.ServantMask(); mask > 0; mask = current_sp.ServantMask() {
		best_sp = nil

		load_balancer.Lock()
		for temp_mask := mask; temp_mask > 0; temp_mask &= (^worker.mask) {
			worker = load_balancer.workers[lsb(BB(temp_mask))]
			for _, this_sp := range worker.sp_list {
				// If a worker has already finished searching, then either a beta cutoff has already
				// occurred at sp, or no moves are left to search.
				if !this_sp.ServantFinished() && (best_sp == nil || this_sp.Order() > best_sp.Order()) {
					best_sp = this_sp
					temp_mask |= this_sp.ServantMask() // If this SP has servants of its own, check them as well.
				}
			}
		}
		if best_sp != nil {
			best_sp.AddServant(w.mask)
		}
		load_balancer.Unlock()

		if best_sp == nil {
			break
		} else {
			w.SearchSP(best_sp)
		}
	}
	// If at any point we can't find another viable servant SP, wait for remaining servants to complete.
	// This prevents us from continually acquiring the load balancer lock.
	current_sp.wg.Wait()
}

func (w *Worker) Help(b *Balancer) {
	go func() {
		var best_sp *SplitPoint
		for {

			b.Lock()
			best_sp = nil
			for _, master := range b.workers { // try to find a good SP
				if master.index == w.index {
					continue
				}
				for _, this_sp := range master.sp_list {
					if !this_sp.ServantFinished() && (best_sp == nil || this_sp.Order() > best_sp.Order()) {
						best_sp = this_sp
					}
				}
			}
			if best_sp != nil {
				best_sp.AddServant(w.mask)
			}
			b.Unlock()

			if best_sp == nil { // No best SP was available.
				b.done <- w             // worker is completely idle and available to help any processor.
				best_sp = <-w.assign_sp // wait for the next SP to be discovered.
			}

			w.current_sp = best_sp
			w.SearchSP(best_sp)
			w.current_sp = nil

		}
	}()
}

func (w *Worker) SearchSP(sp *SplitPoint) {
	brd := sp.brd.Copy()
	brd.worker = w

	sp.master.stk.CopyUpTo(w.stk, sp.ply)
	w.stk[sp.ply].sp = sp

	sp.Lock()
	alpha, beta := sp.alpha, sp.beta
	sp.Unlock()

	// Once the SP is fully evaluated, The SP master will handle returning its value to parent node.
	_, total := ybw(brd, w.stk, alpha, beta, sp.depth, sp.ply, sp.node_type, SP_SERVANT, sp.checked)
	w.search_overhead += total

	sp.RemoveServant(w.mask)
	// At this point, any additional SPs found by the worker during the search rooted at sp
	// should be fully resolved.  The SP list for this worker should be empty again.
}
