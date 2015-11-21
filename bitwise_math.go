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

const (
	DEBRUIJN = 285870213051386505
)

func furthest_forward(c uint8, b BB) int {
	if c == WHITE {
		return lsb(b)
	} else {
		return msb(b)
	}
}

func msb(b BB) int {
	b |= b >> 1
	b |= b >> 2
	b |= b >> 4
	b |= b >> 8
	b |= b >> 16
	b |= b >> 32
	return debruijn_msb_table[(b*DEBRUIJN)>>58]
}

func lsb(b BB) int {
	return debruijn_lsb_table[((b&-b)*DEBRUIJN)>>58]
}

func pop_count(b BB) int {
	b = b - ((b >> 1) & 0x5555555555555555)
	b = (b & 0x3333333333333333) + ((b >> 2) & 0x3333333333333333)
	b = (b + (b >> 4)) & 0x0f0f0f0f0f0f0f0f
	b = b + (b >> 8)
	b = b + (b >> 16)
	b = b + (b >> 32)
	return int(b & 0x7f)
}

var debruijn_lsb_table = [64]int{
	0, 1, 48, 2, 57, 49, 28, 3,
	61, 58, 50, 42, 38, 29, 17, 4,
	62, 55, 59, 36, 53, 51, 43, 22,
	45, 39, 33, 30, 24, 18, 12, 5,
	63, 47, 56, 27, 60, 41, 37, 16,
	54, 35, 52, 21, 44, 32, 23, 11,
	46, 26, 40, 15, 34, 20, 31, 10,
	25, 14, 19, 9, 13, 8, 7, 6,
}

var debruijn_msb_table = [64]int{
	0, 47, 1, 56, 48, 27, 2, 60,
	57, 49, 41, 37, 28, 16, 3, 61,
	54, 58, 35, 52, 50, 42, 21, 44,
	38, 32, 29, 23, 17, 11, 4, 62,
	46, 55, 26, 59, 40, 36, 15, 53,
	34, 51, 20, 43, 31, 22, 10, 45,
	25, 39, 14, 33, 19, 30, 9, 24,
	13, 18, 8, 12, 7, 6, 5, 63,
}

func setup_bitwise_ops() {

}
