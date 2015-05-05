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

// Bit manipulation resources:
// https://chessprogramming.wikispaces.com/Bit-Twiddling

package main

import (
// "fmt"
)

// const (
// 	DEBRUIJN = 285870213051386505
// )

func furthest_forward(c uint8, b BB) int {
	if c == WHITE {
		return lsb(b)
	} else {
		return msb(b)
	}
}

func msb(b BB) int {
	if b>>48 > 0 {
		return msb_table[b>>48] + 48
	}
	if (b>>32)&65535 > 0 {
		return msb_table[(b>>32)&65535] + 32
	}
	if (b>>16)&65535 > 0 {
		return msb_table[(b>>16)&65535] + 16
	}
	return msb_table[b&65535]
}

func lsb(b BB) int {
	if b&65535 > 0 {
		return lsb_table[b&65535]
	}
	if (b>>16)&65535 > 0 {
		return lsb_table[(b>>16)&65535] + 16
	}
	if (b>>32)&65535 > 0 {
		return lsb_table[(b>>32)&65535] + 32
	}
	return lsb_table[b>>48] + 48
}

// func debruijn_lsb(b BB) int {
// 	return lsb_table[lsb_index(b)] LSB by Debruijn multiplication is actually slower...
// }

// func lsb_index(b BB) BB {
// 	return ((b^(b-1)) * DEBRUIJN) >> 58
// }

func pop_count(b BB) int {
	b = (b & 6148914691236517205) + ((b >> 1) & 6148914691236517205)
	b = (b & 3689348814741910323) + ((b >> 2) & 3689348814741910323)
	b = (b & 1085102592571150095) + ((b >> 4) & 1085102592571150095)
	b = (b & 71777214294589695) + ((b >> 8) & 71777214294589695)
	b = (b & 70367670468607) + ((b >> 16) & 70367670468607)
	return int((b & 4294967295) + ((b >> 32) & 4294967295))
}

var msb_table [65536]int
var debruijn_lsb_table [64]int

var lsb_table [65536]int

func setup_bitwise_ops() {
	// for i := 0; i < 64; i++ {
	// 	debruijn_lsb_table[lsb_index(1<<uint64(i))] = i
	// }

	msb_table[0] = 64
	lsb_table[0] = 16
	for i := 1; i < 65536; i++ {
		lsb_table[i] = 16
		for j := 0; j < 16; j++ {
			if (1<<uint(j))&i > 0 {
				msb_table[i] = j
				if lsb_table[i] == 16 {
					lsb_table[i] = j
				}
			}
		}
	}
}
