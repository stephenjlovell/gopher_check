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
// "fmt"
	"sync"
)

type HTable [2][6][64]uint64

var main_htable HTable
var htable_mutex sync.Mutex

func (h *HTable) Store(m Move, c uint8, count int) {
	htable_mutex.Lock()
	h[c][m.Piece()][m.To()] += uint64((count >> 2) | 1)
	htable_mutex.Unlock()
}

func (h *HTable) Probe(pc Piece, c uint8, to int) uint64 {
	value := uint64(0)
	htable_mutex.Lock()
	if h[c][pc][to] > 0 {
		value = (((h[c][pc][to] >> 3) & mask_of_length[21]) | 1) << 1
	}
	htable_mutex.Unlock()
	return value
}

func (h *HTable) Clear() {
	for i := 0; i < 2; i++ {
		for j := 0; j < 6; j++ {
			for k := 0; k < 64; k++ {
				h[i][j][k] = 0
			}
		}
	}
}
