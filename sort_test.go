//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "fmt"

func (l MoveList) isSorted() bool {
	n := len(l)
	for i := 1; i < n; i++ {
		if l[i-1].order < l[i].order {
			printMutex.Lock()
			fmt.Println(l)
			printMutex.Unlock()
			return false
		}
	}
	return true
}
