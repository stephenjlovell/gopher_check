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

var loadBalancer *Balancer

func setupLoadBalancer(numCpu int) {
	numWorkers := uint8(min(numCpu, MAX_WORKERS))
	loadBalancer = NewLoadBalancer(numWorkers)
	loadBalancer.Start()
}

func NewLoadBalancer(numWorkers uint8) *Balancer {
	b := &Balancer{
		workers: make([]*Worker, numWorkers),
		done:    make(chan *Worker, numWorkers),
	}
	for i := uint8(0); i < uint8(numWorkers); i++ {
		b.workers[i] = &Worker{
			mask:      1 << i,
			index:     i,
			spList:   make(SPList, 0, MAX_DEPTH),
			stk:       NewStack(),
			ptt:       NewPawnTT(),
			assignSp: make(chan *SplitPoint, 1),
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
		overhead += w.searchOverhead
	}
	return overhead
}

func (b *Balancer) RootWorker() *Worker {
	return b.workers[0]
}

func (b *Balancer) AddSP(w *Worker, sp *SplitPoint) {
	w.Lock()
	w.spList.Push(sp)
	w.Unlock()
	w.currentSp = sp

FlushIdle: // If there are any idle workers, assign them now.
	for {
		select {
		case idle := <-b.done:
			sp.AddServant(idle.mask)
			idle.assignSp <- sp
		default:
			break FlushIdle
		}
	}
}

// RemoveSP prevents new workers from being assigned to w.current_sp without cancelling
// any ongoing work at this SP.
func (b *Balancer) RemoveSP(w *Worker) {
	w.Lock()
	w.spList.Pop()
	w.Unlock()
	w.currentSp = w.currentSp.parent
}

func (b *Balancer) Print() {
	for i, w := range b.workers {
		if len(w.spList) > 0 {
			w.Lock()
			fmt.Printf("w%d: ", i)
			for _, sp := range w.spList {
				fmt.Printf("%d, ", (sp.brd.hashKey >> 48))
			}
			fmt.Printf("\n")
			w.Unlock()
		}
	}
}
