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
	"sync/atomic"
)

type HTable [2][8][64]uint64

var main_htable HTable

// Store atomically adds count to the history table h.
func (h *HTable) Store(m Move, c uint8, count int) {
	atomic.AddUint64(&h[c][m.Piece()][m.To()], uint64((count >> 2) | 1))
}

// Probe atomically reads the history table h.
func (h *HTable) Probe(pc Piece, c uint8, to int) uint64 {
	v := atomic.LoadUint64(&h[c][pc][to])
	if v > 0 {
		return ((((v >> 3) & mask_of_length[21]) | 1) << 1)
	}
	return 0
}

func (h *HTable) Clear() {
	for i := 0; i < 2; i++ {
		for j := 0; j < 6; j++ {
			for k := 0; k < 64; k++ {
				atomic.StoreUint64(&h[i][j][k], 0)
			}
		}
	}
}
