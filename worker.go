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
  "fmt"
  // "sync"
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
  MAX_SP = MAX_WORKER_GOROUTINES * MAX_SP_PER_WORKER

  ALL_SERVANTS_SEARCHING = (1<<MAX_WORKER_GOROUTINES)-1
)


var load_balancer *Balancer

func NewLoadBalancer() *Balancer {
  b := &Balancer{
    work: make(chan SPListItem, MAX_SP+1),
    done: make(chan Done, MAX_WORKER_GOROUTINES+1),
    waiting: make(chan Done, MAX_WORKER_GOROUTINES+1),
    remove_sp: make(chan SPCancellation, MAX_SP+1),
  }

  for i := uint8(0); i < MAX_WORKER_GOROUTINES; i++ {
    b.workers[i] = &Worker{
      mask: 1 << i,
      index: i,
      available_slots: make(chan bool, MAX_SP_PER_WORKER),
      assign_sp: make(chan SPListItem, 1),
    }
    sp_list := make(SPList, 0, MAX_SP_PER_WORKER)
    b.sp_lists[i] = &sp_list
    heap.Init(b.sp_lists[i])

    for j := 0; j < MAX_WORKER_GOROUTINES; j++ {
      b.workers[i].available_slots <- true  // fill up the availability buffer since the SP list is empty.
    }
  }
  return b
}

type Balancer struct{
  workers [MAX_WORKER_GOROUTINES]*Worker
  sp_lists [MAX_WORKER_GOROUTINES]*SPList

  sp_counts [MAX_WORKER_GOROUTINES]int  // cache the number of available SP per worker.
  sp_mask   uint8

  // All communication with the load balancer should be done via channel.
  work chan SPListItem
  done chan Done
  waiting chan Done
  remove_sp chan SPCancellation
}

func (b *Balancer) Start() {
  for _, w := range b.workers[1:] {
    w.Help(b.done) // Start each worker except for the root worker.
  }
  b.Balance() // start assigning tasks to workers
}

// when work is complete at an SP node, there's no guarantee that the SP cancellation will happen
// before the workers finish.

func (b *Balancer) Balance() {
  go func() {
    for {
      fmt.Printf(">")
      select {
      case c := <-b.remove_sp:
        b.RemoveSP(c)

      case item := <-b.work:
        // A worker has discovered a good split point.  Add the SP to the worker's SP List. 
        // This allows other workers to begin searching this SP node when they're ready, without having
        // to wait for this worker to return to this point in the search.
        // fmt.Printf(" LB: New SP @ worker%d", item.sp.master.index)
        b.AddSP(item)

        FlushStaging: // Any workers waiting for a relevant SP should be sent back to the done queue.
          for {
            select {
            case d := <-b.waiting:
              b.done <- d
            default:
              break FlushStaging    
            }
          }

      case d := <-b.done:
        b.AssignBestSP(d.w) // Assign the worker as servant to best available SP, or move the worker 
                            // to waiting area if the worker can't participate in any available SP.  
      }
    }
  }()
}


func (b *Balancer) AddSP(item SPListItem) {
  master := item.sp.master

  b.sp_mask |= master.mask
  b.sp_counts[master.index] += 1
  item.worker_mask |= master.mask
  
  heap.Push(b.sp_lists[item.sp.master.index], &item)
  fmt.Printf(" Added SP%d", item.sp.brd.hash_key)
}

func (b *Balancer) RemoveSP(c SPCancellation) {
  sp_list := b.sp_lists[c.sp.master.index]
  for i, item := range *sp_list {
    if item.sp.brd.hash_key == c.hash_key {
      heap.Remove(sp_list, i)
      b.sp_counts[i] -= 1
      if b.sp_counts[i] == 0 {
        b.sp_mask &= (^item.sp.master.mask)
      }
      close(c.sp.cancel)
      c.sp.master.available_slots <- true // let the worker know it now has another SP slot available.
      fmt.Printf(" Removed SP%d", c.hash_key)
      return
    }
  }
  fmt.Printf(" Missing SP%d", c.hash_key)
}

func (b *Balancer) AssignBestSP(w *Worker) bool { // Select the best available split point.
  var sp_list SPList
  var item, best_item *SPListItem
  var best_order uint8
  var l int
  index := w.index
  for i := uint8(0); i < MAX_WORKER_GOROUTINES; i++ {
    if i == index {
      continue
    }
    sp_list = *b.sp_lists[i]
    l = b.sp_counts[i]
    if l > 0 {
      item = sp_list[l-1]
      if item.worker_mask & w.mask > 0 {
        // fmt.Printf(" worker%d already assigned to this node: %b!", w.index, item.worker_mask)
      }
      if item.worker_mask & w.mask == 0 && item.order > best_order {
        best_item = item
      }
    }
  }

  if best_item == nil {
    b.waiting <- Done{w}  // Put the worker back into the queue.
    return false
  }

  best_item.worker_mask |= w.mask 
  w.assign_sp <- *best_item
  return true
}

func (item *SPListItem) AllSearching() bool {
  return (item.worker_mask>>1) == ALL_SERVANTS_SEARCHING
}

func (b *Balancer) RootWorker() *Worker {
  return b.workers[0] 
}


type Worker struct {
  // sync.Mutex
  mask uint8
  index uint8

  available_slots chan bool
	assign_sp chan SPListItem
}

func (w *Worker) Help(done chan Done) {
  go func() {
    for {
      done <- Done{w}
      // fmt.Printf(" worker%d requested more work.", w.index)

      item := <-w.assign_sp  // Wait for LB to assign this worker as a servant to another worker.

      sp := item.sp
      brd := sp.brd.Copy()
      stk := item.stk.CopyUpTo(sp.ply)
      stk[sp.ply].sp = sp
      brd.worker = w

      fmt.Printf(" worker%d searching:\n", w.index)
      // fmt.Printf("%d %d %d %d %d\n", sp.alpha, sp.beta, sp.depth, sp.ply, sp.extensions_left)
      // Once the SP is fully evaluated, The SP master will handle returning its value to parent node.
      _, _ = ybw(brd, stk, sp.alpha, sp.beta, sp.depth, sp.ply, 
                 sp.extensions_left, sp.can_null, sp.node_type, SP_SERVANT)

      // At this point, any additional SPs found by the worker during the search rooted at a.sp
      // should be fully resolved.  The SP list for this worker should be empty again.
      fmt.Printf(" worker%d finished.", w.index)
    }
  }()
}


type Done struct {
  w *Worker
}







