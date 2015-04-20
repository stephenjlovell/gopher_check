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
  "container/heap"
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
  MAX_SP_PER_WORKER = 8
)

var load_balancer *Balancer

func NewLoadBalancer() *Balancer {
  b := &Balancer{
    workers: make([]*Worker, MAX_WORKER_GOROUTINES, MAX_WORKER_GOROUTINES),
    done: make(chan Done, MAX_WORKER_GOROUTINES),
  }
  for i := uint8(0); i < MAX_WORKER_GOROUTINES; i++ {
    b.workers[i] = &Worker{
      id: 1 << i,
      sp_list: make(SPList, 0, MAX_SP_PER_WORKER),
      sp_available: b.sp_available,
    }
  }
  return b
}

type Balancer struct{
  workers []*Worker
  // sp_list SPList
  done chan Done
  // work chan SPListItem
  sp_available chan bool
}

func (b *Balancer) Start() {
  for _, w := range b.workers[1:] {
    w.Help(b.done) // Start each worker except for the root worker.
  }
  b.Balance() // start assigning tasks to workers
}

func (b *Balancer) Balance() {
  go func() {
    for {
      select {
      // case item := <-b.work:

      case d := <-b.done: 
        // A slave worker has finished searching. Any SPs generated during search have been dealt with,
        // and the worker's SP list should now be empty.
        <-b.sp_available // wait for at least one split point to be available
        d.w.assign_sp <- b.GetAvailable() // Assign the worker as slave to best available SP
      }
    }
  }()
}

func (b *Balancer) RootWorker() *Worker {
  return b.workers[0] 
}

// Select the best available split point.
func (b *Balancer) GetAvailable() SPListItem {
  var item, best_item SPListItem
  var best_order int
  for _, w := range b.workers {
    w.Lock()
    if w.sp_count > 0 {
      item = *w.sp_list[w.sp_count-1]
      if item.order > best_order {
        best_item = item
      }      
    }
    w.Unlock()
  }
  return best_item
}



type Worker struct {
  sync.Mutex
  id int
  sp_available chan bool
  sp_count int  // cache the size of the SP list
	assign_sp chan SPListItem
	sp_list  SPList  // stores the SPs for which this worker is responsible.
}

type Done struct {
  w *Worker
  // item *SPListItem
}

func (w *Worker) RemoveSP(hash_key uint64) { // find and remove a finished SP
  w.Lock()
  for i, item := range w.sp_list {
    if item.sp.brd.hash_key == hash_key {
      heap.Remove(&w.sp_list, i)
      break
    }
  }
  w.sp_count--
  w.Unlock()
}

func (w *Worker) CanAddSP() bool {
  return w.sp_count < MAX_SP_PER_WORKER
}

func (w *Worker) AddSP(sp *SplitPoint, stk Stack) {
  // A worker has discovered a good split point.  Add the SP to the worker's SP List. 
  // This allows other workers to begin searching this SP node when they're ready, without having
  // to wait for this worker to return to this point in the search.
  item := SPListItem{sp: sp, stk: stk, order: (sp.depth << 3) | sp.node_type }
  heap.Push(&w.sp_list, item)
  w.sp_available <- true // signal load balancer that another SP is available.
}


func (w *Worker) Help(done chan Done) {
  go func() {
    for {
      a := <-w.assign_sp  // Wait for LB to assign this worker as a slave to another worker.
      sp := a.sp
      brd := sp.brd.Copy()
      brd.worker = w
      // Once the SP is fully evaluated, The SP master will handle returning its value to parent node.
      _, _ = ybw(brd, a.stk.CopyUpTo(sp.ply), sp.alpha, sp.beta, sp.depth, sp.ply, 
                sp.this_stk.extensions_left, sp.this_stk.can_null, true, sp.node_type)
      // At this point, any additional SPs found by the worker during the search rooted at a.sp
      // should be fully resolved.  The worker's SP list should be empty again.
      done <- Done{w}
    }
  }()
}









