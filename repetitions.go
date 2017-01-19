//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

// "fmt"

func (stk Stack) IsRepetition(ply int, halfmoveClock uint8) bool {
	hashKey := stk[ply].hashKey
	if halfmoveClock < 4 {
		return false
	}
	for repetitionCount := 0; ply >= 2; ply -= 2 {
		if stk[ply-2].hashKey == hashKey {
			repetitionCount += 1
			if repetitionCount == 2 {
				return true
			}
		}
	}
	return false
}
