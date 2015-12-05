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
	MAX_WORKERS = 8
)

var node_count []SafeCounter

var load_balancer *Balancer

func setup_load_balancer(num_cpu int) {
	num_workers := uint8(min(num_cpu, MAX_WORKERS))
	node_count = make([]SafeCounter, num_workers, num_workers)
	load_balancer = NewLoadBalancer(num_workers)
	load_balancer.Start()
}

func NewLoadBalancer(num_workers uint8) *Balancer {
	b := &Balancer{
		workers: make([]*Worker, num_workers),
		done:    make(chan *Worker, num_workers),
	}
	for i := uint8(0); i < uint8(num_workers); i++ {
		b.workers[i] = &Worker{
			mask:      1 << i,
			index:     i,
			sp_list:   make(SPList, 0, MAX_PLY),
			stk:       NewStack(),
			ptt:       NewPawnTT(),
			assign_sp: make(chan *SplitPoint, 1),
		}
	}
	return b
}

type Balancer struct {
	workers []*Worker
	sync.Mutex

	done chan *Worker
}

func (b *Balancer) Start() {
	for _, w := range b.workers[1:] {
		w.Help(b) // Start each worker except for the root worker.
	}
}

func (b *Balancer) Overhead() int {
	overhead := 0
	for _, w := range b.workers {
		overhead += w.search_overhead
	}
	return overhead
}

func (b *Balancer) RootWorker() *Worker {
	return b.workers[0]
}

type Worker struct {
	sync.Mutex
	mask  uint8
	index uint8

	search_overhead int

	sp_list    SPList
	current_sp *SplitPoint

	stk Stack
	ptt *PawnTT

	assign_sp chan *SplitPoint
}

func (w *Worker) CancelSP() { // Should only be called by the SP master
	load_balancer.Lock()

	w.sp_list.Pop()
	last_sp := w.current_sp.parent
	w.current_sp.cancel = true
	w.current_sp = last_sp

	load_balancer.Unlock()
}

func (w *Worker) RemoveSP() { // Prevent new workers from being assigned to w.current_sp without
	load_balancer.Lock() // cancelling any ongoing work at this SP.

	w.sp_list.Pop()
	last_sp := w.current_sp.parent
	w.current_sp = last_sp

	load_balancer.Unlock()
}

func (w *Worker) IsCancelled() bool {
	for sp := w.current_sp; sp != nil; sp = sp.parent {
		if sp.cancel {
			return true
		}
	}
	return false
}

func (w *Worker) HelpServants(current_sp *SplitPoint) {
	var best_sp *SplitPoint
	var worker *Worker
	// Check for SPs underneath current_sp

	for mask := current_sp.servant_mask; mask > 0; mask = current_sp.servant_mask {
		best_sp = nil

		load_balancer.Lock()
		for temp_mask := mask; temp_mask > 0; temp_mask &= (^worker.mask) {
			worker = load_balancer.workers[lsb(BB(temp_mask))]
			for _, this_sp := range worker.sp_list {
				// If a worker has already finished searching, then either a beta cutoff has already
				// occurred at sp, or no moves are left to search.
				if !this_sp.servant_finished && (best_sp == nil || this_sp.Order() > best_sp.Order()) {
					best_sp = this_sp
					temp_mask |= this_sp.servant_mask // If this SP has servants of its own, check them as well.
				}
			}
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
					if !this_sp.servant_finished && (best_sp == nil || this_sp.Order() > best_sp.Order()) {
						best_sp = this_sp
					}
				}
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

	sp.AddServant(w.mask)

	// Once the SP is fully evaluated, The SP master will handle returning its value to parent node.
	_, total := ybw(brd, w.stk, sp.alpha, sp.beta, sp.depth, sp.ply, sp.node_type, SP_SERVANT, sp.checked)
	w.search_overhead += total

	sp.RemoveServant(w.mask)
	// At this point, any additional SPs found by the worker during the search rooted at sp
	// should be fully resolved.  The SP list for this worker should be empty again.
}
