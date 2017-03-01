//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestRecyclerThreadSafety(t *testing.T) {
	r := NewRecycler()
	count := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(r *Recycler) {
			defer wg.Done()
			var moves MoveList
			for j := 0; j < 100; j++ {
				moves = r.ReuseMoveList(DEFAULT_MOVE_LIST_LENGTH)
				time.Sleep(time.Microsecond * time.Duration(rand.Intn(1000)))
				fmt.Println(len(moves))
				// r.g.Dump()
				r.RecycleMoveList(moves)
			}
		}(r)
	}
	wg.Wait()
}
