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
  MAX_GOPROCS = 8
  MAX_SP_PER_WORKER = 8
)

var load_balancer *Balancer

func setup_load_balancer() {
  load_balancer = &Balancer{
    workers: make([]*Worker, 0, MAX_GOPROCS),
    done: make(chan *Worker, MAX_GOPROCS),
  }
  workers := load_balancer.workers
  for i := uint8(0); i < MAX_GOPROCS; i++ {
    workers[i] = &Worker{
      id: 1 << i,
      sp_list: make(SPList, MAX_SP_PER_WORKER),
    }
  }
}



type Balancer struct{
  workers []*Worker
  done chan *Worker
}


func (b *Balancer) Start() {
  for _, w := range b.workers {
    w.Work(b.done) // Start each worker
  }
}

func (b *Balancer) GetAvailable() {}

type Worker struct {
  sync.Mutex
  id int
  sp_count int  // cache the size of the SP list
	assignments chan Assignment
	sp_list  SPList  // stores the SPs for which this worker is responsible.
}

func (w *Worker) CanAddSP() bool {
  w.Lock()
  can_split := w.sp_count < MAX_SP_PER_WORKER
  w.Unlock()
  return can_split
}

func (w *Worker) AddSP(sp *SplitPoint) {
  w.Lock()
  w.sp_list = append(w.sp_list, sp) // may want to organize as a heap in order to 
                                    // keep sorted. Otherwise implement insertion sort
  w.Unlock()
}


func (w *Worker) Work(done chan *Worker) {

  go func() {
    for {
      a := <-w.assignments  // Wait for LB to assign this worker as a slave to another worker.
      sp := a.sp
      brd := sp.brd.Copy()
      brd.worker = w
      _, _ = ybw(brd, a.stk, sp.alpha, sp.beta, sp.depth, sp.ply, 
                sp.this_stk.extensions_left, sp.this_stk.can_null, true, sp.node_type)
    }
  }()

}



type Assignment struct{
  sp *SplitPoint
  stk Stack
}







