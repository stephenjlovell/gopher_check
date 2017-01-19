//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"sync"
	// "time"
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

var load_balancer *Balancer

func setup_load_balancer(num_cpu int) {
	num_workers := uint8(min(num_cpu, MAX_WORKERS))
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
			sp_list:   make(SPList, 0, MAX_DEPTH),
			stk:       NewStack(),
			ptt:       NewPawnTT(),
			assign_sp: make(chan *SplitPoint, 1),
			recycler:  NewRecycler(512),
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
	w.Unlock()
	w.current_sp = sp

FlushIdle: // If there are any idle workers, assign them now.
	for {
		select {
		case idle := <-b.done:
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
	w.Unlock()
	w.current_sp = w.current_sp.parent
}

func (b *Balancer) Print() {
	for i, w := range b.workers {
		if len(w.sp_list) > 0 {
			w.Lock()
			fmt.Printf("w%d: ", i)
			for _, sp := range w.sp_list {
				fmt.Printf("%d, ", (sp.brd.hash_key >> 48))
			}
			fmt.Printf("\n")
			w.Unlock()
		}
	}
}
