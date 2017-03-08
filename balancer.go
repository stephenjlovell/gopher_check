//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"sync"
)

var loadBalancer *Balancer

func setupLoadBalancer(numCPU int) {
	loadBalancer = NewLoadBalancer(uint8(numCPU))
	loadBalancer.Start(numCPU)
}

func NewLoadBalancer(numWorkers uint8) *Balancer {
	b := &Balancer{
		workers: make([]*Worker, numWorkers),
		done:    make(chan *Worker, numWorkers),
	}
	for i := uint8(0); i < numWorkers; i++ {
		b.workers[i] = &Worker{
			mask:     uint32(1 << i),
			index:    i,
			spList:   make(SPList, 0, MAX_DEPTH),
			stk:      NewStack(),
			ptt:      NewPawnTT(),
			assignSp: make(chan *SplitPoint, 1),
			recycler: NewRecycler(),
		}
	}
	return b
}

type Balancer struct {
	workers []*Worker
	once    sync.Once
	done    chan *Worker
}

func (b *Balancer) Start(numCPU int) {
	b.once.Do(func() {
		for _, w := range b.workers[1:] {
			w.Help(b) // Start each worker except for the root worker.
		}
	})
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

// AddSP registers Worker w as the 'master' of sp, meaning w is responsible for collecting results
// from all goroutines collaborating at this Split Point. AddSP is only called by the worker that
// found sp.
func (b *Balancer) AddSP(w *Worker, sp *SplitPoint) {
	sp.master = w
	w.Lock()
	sp.parent = w.currentSP
	w.spList.Push(sp)
	w.currentSP = sp
	w.Unlock()

FlushIdle: // If there are any idle workers, assign them now.
	for {
		select {
		case idle := <-b.done:
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
	w.currentSP = w.currentSP.parent
	w.Unlock()
}

func (b *Balancer) Print() {
	printMutex.Lock()
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
	printMutex.Unlock()
}
