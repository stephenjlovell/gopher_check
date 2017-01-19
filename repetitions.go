//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

// "fmt"

func (stk Stack) IsRepetition(ply int, halfmove_clock uint8) bool {
	hash_key := stk[ply].hash_key
	if halfmove_clock < 4 {
		return false
	}
	for repetition_count := 0; ply >= 2; ply -= 2 {
		if stk[ply-2].hash_key == hash_key {
			repetition_count += 1
			if repetition_count == 2 {
				return true
			}
		}
	}
	return false
}
