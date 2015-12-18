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
	"time"
	"sync"
	"fmt"
)

// 2-Level Locking Scheme
//
// Load Balancer (Global-level) Lock
//
//   - Protects integrity of SP lists maintained by workers (Adding / Removing SPs)
//   USAGE:
//   - When a worker self-assigns (searches for an SP from available SP lists), it should lock
//     the load balancer. Load balancer should only be unlocked after the worker is registered
//     with the SP.
//   - Finding the best SP should ideally be encapsulated by the load balancer.
//
// Split Point (local) Lock
//
//   - Protects search state for workers collaborating on same SP.
//   - Protects info on which workers are collaborating at this SP.
//   USAGE:
//   - When the master directly assigns a worker, it should register the worker immediately
//     with the SP before sending the SP to the worker to avoid a data race with the SP's
// 		 WaitGroup.

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

func (b *Balancer) AddSP(w *Worker, sp *SplitPoint) {
	w.Lock()
	w.sp_list.Push(sp)
	w.current_sp = sp
	w.Unlock()

FlushIdle: // If there are any idle workers, assign them now.
	for {
		select {
		case idle := <-load_balancer.done:
			sp.AddServant(idle.mask)
			idle.assign_sp <- sp
		default:
			break FlushIdle
		}
	}
}

// RemoveSP prevents new workers from being assigned to w.current_sp without cancelling
// any ongoing work at this SP.
func (b *Balancer) RemoveSP(w *Worker) {
	w.Lock()
	w.sp_list.Pop()
	last_sp := w.current_sp.parent
	w.current_sp = last_sp
	w.Unlock()
}

func (b *Balancer) Print() {
	for i, w := range b.workers {
		if len(w.sp_list) > 0 {
			w.Lock()
			fmt.Printf("w%d: ", i)
			for _, sp := range w.sp_list {
				fmt.Printf("%d, ", (sp.brd.hash_key>>48))
			}
			fmt.Printf("\n")
			w.Unlock()
		}
	}
}

func (b *Balancer) KeepPrinting(cancel chan bool) {
	go func() {
		for {
			select {
			case <-cancel:
				return
			default:
				b.Print()
				time.Sleep(time.Second)
			}
		}
		// time.Sleep(time.Second * 5)
		// panic("Gimme all your stack traces...")
	}()
}
