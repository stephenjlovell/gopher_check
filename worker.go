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
  "time"
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
  MAX_SP = MAX_WORKER_GOROUTINES * MAX_SP_PER_WORKER

  ALL_SERVANTS_SEARCHING = (1<<MAX_WORKER_GOROUTINES)-1
)


var load_balancer *Balancer

func NewLoadBalancer() *Balancer {
  b := &Balancer{
    work: make(chan SPListItem, MAX_SP),
    waiting: make(chan Done, MAX_WORKER_GOROUTINES),
    cancel_sp: make(chan SPCancellation, MAX_SP),

    done: make(chan uint8, MAX_WORKER_GOROUTINES),
    add_sp: make(chan uint8, MAX_SP),
    remove_sp: make(chan uint8, MAX_SP),
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
  sp_mutex sync.Mutex

  // All communication with the load balancer should be done via channel.
  work      chan SPListItem
  waiting   chan Done
  cancel_sp chan SPCancellation

  done      chan uint8
  add_sp    chan uint8
  remove_sp chan uint8
}

func (b *Balancer) Start() {
  for _, w := range b.workers[1:] {
    w.Help(b.done) // Start each worker except for the root worker.
  }
  b.Dispatch()
  b.AddSP()
  b.CleanupSP()
  // b.StartPrinting()
}

func (b *Balancer) StartPrinting() {
  go func() {
    for _ = range time.Tick(500 * time.Millisecond) {  
      b.Print()  
    }
  }()
}

func (b *Balancer) Print() {
  b.sp_mutex.Lock()
  fmt.Printf("\n| ")
  for i := 0; i < MAX_WORKER_GOROUTINES; i++ {
    list := b.sp_lists[i]
    fmt.Printf("%d ", len(*list))
  }
  fmt.Printf("|")
  b.sp_mutex.Unlock()
}


func (b *Balancer) Dispatch() {
  go func() {
    var idle_mask, sp_mask, assigned uint8

    // wait for a new worker or SP to become available, then try to dispatch work.
    for {
      fmt.Printf("\nidle_mask:%b  sp_mask:%b", idle_mask, sp_mask)
      select {
      case w_mask := <- b.done:
        fmt.Printf("\nAssigning...")
        idle_mask |= w_mask
        assigned = b.AttemptAssignment(idle_mask, sp_mask)
        idle_mask &= (^assigned)
        fmt.Printf("Assigned %b", assigned)
      case add_mask := <-b.add_sp:
        fmt.Printf("\nadding SP...")
        sp_mask |= add_mask
        assigned = b.AttemptAssignment(idle_mask, sp_mask)
        idle_mask &= (^assigned)
        fmt.Printf("assigned %b", assigned)
      case remove_mask := <-b.remove_sp:
        sp_mask &= (^remove_mask)
      }
    }

  }()
}

func (b *Balancer) AttemptAssignment(idle_mask, sp_mask uint8) uint8 {
  var w_index, sp_index, w_temp, sp_temp, w_current, sp_current, assigned uint8
  var item *SPListItem
  var w *Worker

  w_temp = idle_mask
  b.sp_mutex.Lock()

  for ; w_temp > 0; w_temp &= (^w_current) {
    w_index = uint8(lsb(BB(w_temp)))
    w_current = 1 << w_index

    sp_temp = sp_mask & (^w_current)
    for ; sp_temp > 0; sp_temp &= (^sp_current) {
      sp_index = uint8(lsb(BB(sp_temp)))
      sp_current = 1 << sp_index

      item = heap.Pop(b.sp_lists[sp_index]).(*SPListItem)
      w = b.workers[w_index]
      item.worker_mask |= w.mask
      
      assert(item != nil, "Nil item in SP List!")

      // fmt.Printf(" Assigning...")
      w.assign_sp <- *item // send the item to the worker
      // fmt.Printf("Assigned")

      assigned |= w.mask
      heap.Push(b.sp_lists[sp_index], item)
      break
    }
  }
  b.sp_mutex.Unlock()
  return assigned
}


func (b *Balancer) AddSP() {
  go func() {
    for {
      select {
      case item := <-b.work:
        // A worker has discovered a good split point.  Add the SP to the worker's SP List. 
        // This allows other workers to begin searching this SP node when they're ready, without having
        // to wait for this worker to return to this point in the search.
        // fmt.Printf(" LB: New SP @ worker%d", item.sp.master.index)

        master := item.sp.master
        item.worker_mask |= master.mask
        b.sp_mutex.Lock()
        heap.Push(b.sp_lists[item.sp.master.index], &item)
        b.add_sp <- master.mask
        b.sp_mutex.Unlock()
        // fmt.Printf(" Added SP%x", item.sp.brd.hash_key)
      }
    }
  }()
}

type SPCancellation struct {
  w *Worker
  hash_key uint64
  cancel_servant bool
}

// Only masters should send cancellation signal.
func (b *Balancer) CleanupSP() {
  go func() {
    var helpers uint8
    var l, i, index int
    var item *SPListItem

    for {
      c := <-b.cancel_sp  // iteratively remove the SP and all its descendants.
      // fmt.Printf(" removing SP%x", c.hash_key)
      b.sp_mutex.Lock()
      list := b.sp_lists[c.w.index]
      l = len(*list)
      if c.cancel_servant {
        // remove all SPs for this worker, since it's master node has been cancelled.
        for ; l > 0; l-- {
          item = heap.Pop(list).(*SPListItem)
          close(item.sp.cancel)

          helpers = item.worker_mask
          helpers &= (^c.w.mask)
          index = 0
          for ; helpers > 0; helpers &= ^(1<<uint8(index)) {
            index = lsb(BB(helpers))
            b.cancel_sp <- SPCancellation{b.workers[index], 0, true}
          }  
        }
        b.remove_sp <- c.w.mask
      } else {
        for i, item = range *list {
          if item.sp.brd.hash_key == c.hash_key {
            helpers = item.worker_mask
            heap.Remove(list, i)
            close(item.sp.cancel)
            c.w.available_slots <- true
            if l == 1 {
              b.remove_sp <- c.w.mask // If this was the only SP for this worker, update the SP mask.
            }
            break
          }
        }
        helpers &= (^c.w.mask)
        index = 0
        for ; helpers > 0; helpers &= ^(1<<uint8(index)) {
          index = lsb(BB(helpers))
          b.cancel_sp <- SPCancellation{b.workers[index], 0, true}
        }      
      }
      b.sp_mutex.Unlock()
    }
  }()
}



// func (item *SPListItem) AllSearching() bool {
//   return (item.worker_mask>>1) == ALL_SERVANTS_SEARCHING
// }

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

func (w *Worker) Help(done chan uint8) {
  go func() {
    for {
      done <- w.mask

      item := <-w.assign_sp  // Wait for LB to assign this worker as a servant to another worker.

      sp := item.sp
      brd := sp.brd.Copy()
      stk := item.stk.CopyUpTo(sp.ply)
      stk[sp.ply].sp = sp
      brd.worker = w

      // fmt.Printf(" worker%d searching SP%x\n", w.index, sp.brd.hash_key)
      // fmt.Printf("%d %d %d %d %d\n", sp.alpha, sp.beta, sp.depth, sp.ply, sp.extensions_left)
      // Once the SP is fully evaluated, The SP master will handle returning its value to parent node.
      _, _ = ybw(brd, stk, sp.alpha, sp.beta, sp.depth, sp.ply, 
                 sp.extensions_left, sp.can_null, sp.node_type, SP_SERVANT)

      // At this point, any additional SPs found by the worker during the search rooted at a.sp
      // should be fully resolved.  The SP list for this worker should be empty again.
      // fmt.Printf(" worker%d finished SP%x", w.index, sp.brd.hash_key)

      assert(len(*load_balancer.sp_lists[w.index]) == 0, " worker" + string(w.index) + " returned with work remaining")

      for i := 0; i < MAX_SP_PER_WORKER; i++ {
        select {
        case w.available_slots <- true:  // Refill the availability channel.
        default:
        }
      }
    }
  }()
}


type Done struct {
  w *Worker
}







