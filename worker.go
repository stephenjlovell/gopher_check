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
	MAX_WORKER_GOROUTINES = 8
	MAX_SP_PER_WORKER     = 8
	MAX_SP                = MAX_WORKER_GOROUTINES * MAX_SP_PER_WORKER
)

var node_count [MAX_WORKER_GOROUTINES]SafeCounter

var load_balancer *Balancer

func setup_load_balancer() {
	load_balancer = NewLoadBalancer()
	load_balancer.Start()
}

func NewLoadBalancer() *Balancer {
	b := &Balancer{
		workers: make([]*Worker, MAX_WORKER_GOROUTINES),
		done:    make(chan *Worker, MAX_WORKER_GOROUTINES),
	}
	for i := uint8(0); i < MAX_WORKER_GOROUTINES; i++ {
		b.workers[i] = &Worker{
			mask:      1 << i,
			index:     i,
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

func (b *Balancer) RootWorker() *Worker {
	return b.workers[0]
}

type Worker struct {
	sync.Mutex
	mask  uint8
	index uint8

	current_sp *SplitPoint
	stk        Stack
	ptt        *PawnTT

	// polling_order []int

	assign_sp chan *SplitPoint
	// cancel chan bool
}

func (w *Worker) RemoveSP() {
	load_balancer.Lock()

	last_sp := w.current_sp.parent
	close(w.current_sp.cancel)
	w.current_sp = last_sp

	load_balancer.Unlock()
}

func (w *Worker) Help(b *Balancer) {

	go func() {
		var sp, best_sp *SplitPoint

		for {

			// to do: Randomize the order of workers when looking for best SP

			b.Lock()
			sp, best_sp = nil, nil
			for _, master := range b.workers { // try to find a good SP
				if master.index == w.index {
					continue
				}
				sp = master.current_sp
				for sp != nil {
					if best_sp == nil || sp.Order() > best_sp.Order() {
						best_sp = sp
					}
					sp = sp.parent // walk the SP list in search of the best place to split
				}
			}
			b.Unlock()

			if best_sp == nil { // no SP was available.
				b.done <- w
				sp = <-w.assign_sp // wait for the next SP to be discovered.
			} else {
				sp = best_sp
			}

			brd := sp.brd.Copy()
			brd.worker = w
			sp.master.stk.CopyUpTo(w.stk, sp.ply)
			w.stk[sp.ply].sp = sp

			sp.wg.Add(1)
			// sp.Lock()
			// sp.servant_mask |= w.mask
			// sp.Unlock()

			// Once the SP is fully evaluated, The SP master will handle returning its value to parent node.
			_, _ = ybw(brd, w.stk, sp.alpha, sp.beta, sp.depth, sp.ply,
				sp.extensions_left, sp.can_null, sp.node_type, SP_SERVANT)

			// At this point, any additional SPs found by the worker during the search rooted at a.sp
			// should be fully resolved.  The SP list for this worker should be empty again.

			sp.wg.Done()
			// sp.Lock()
			// sp.servant_mask &= (^w.mask)
			// sp.Unlock()

		}
	}()
}
