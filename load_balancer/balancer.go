//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
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

package load_balancer

// The load balancer uses a priority queue to divide up subtree searches evenly among available 'worker'
// goroutines, based on the expected effort required to search the subtree and its relative importance
// based on how promising that part of the tree appears to be.

// Expected effort is the average branching factor for the game tree raised by the depth remaining to search.

// Subtrees rooted along the Principal Variation (PV-Nodes) should be searched first, 
// followed by fail-high nodes (Cut-Nodes). Fail-low nodes (All-Nodes) should be searched last.
// Nodes of the same type should be processed left to right (making use of move ordering heuristics)


import (
  "container/heap"
  "fmt"
  // "runtime"
)

type Balancer struct {
  pool Pool
  done chan Done
}

func (b *Balancer) start() {
  for _, worker := range b.pool {
    worker.work(b.done)
  }
}

func (b *Balancer) balance(work chan Request) {
  go func() {
    for {
      select {
      case req := <-work: // request received
        b.dispatch(req) // forward request to a worker
      case d := <-b.done: // worker finished with a request
        b.completed(d) 
      }
    }
  }()
}

func (b *Balancer) dispatch(req Request) { // route the request to the most lightly loaded 
  w := heap.Pop(&b.pool).(*Worker)         // worker in the priority queue, and adjust queue
  w.requests <- req                        // ordering if needed.
  w.pending += req.Size
  heap.Push(&b.pool, w)
}

func (b *Balancer) completed(d Done) {  // adjust the ordering of the priority queue.
  d.w.pending -= d.size
  heap.Remove(&b.pool, d.w.index)
  heap.Push(&b.pool, d.w)
}

func (b *Balancer) Print() {
  fmt.Printf("\n")
  total_pending := 0
  for _, worker := range b.pool {
    pending := worker.pending
    fmt.Printf("%d  ", pending)
    total_pending += pending
  }
  fmt.Printf("| %d  ", total_pending)
}

func (b *Balancer) Setup(work chan Request) {
  b.start()
  b.balance(work)
}

func NewBalancer(work chan Request) *Balancer {  // Balancer constructor
  nworker := 4
  b := &Balancer{
    done: make(chan Done, 100),
    pool: make(Pool, nworker),
  }
  for i := 0; i < nworker; i++ {
    b.pool[i] = &Worker{
      requests: make(chan Request, 100), // each worker needs its own channel on which to receive work
      index:    i,                       // from the load balancer.
    }
  }
  heap.Init(&b.pool)
  return b
}

func SetupNewBalancer(work chan Request) *Balancer {
  b := NewBalancer(work)
  b.start()
  b.balance(work)
  return b
}





























