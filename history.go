//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"sync/atomic"
)

type HistoryTable [2][8][64]uint32

// Store atomically adds count to the history table h.
func (h *HistoryTable) Store(m Move, c uint8, count int) {
	atomic.AddUint32(&h[c][m.Piece()][m.To()], uint32((count>>3)|1))
}

// Probe atomically reads the history table h.
func (h *HistoryTable) Probe(pc Piece, c uint8, to int) uint32 {
	if v := atomic.LoadUint32(&h[c][pc][to]); v > 0 {
		return ((v >> 3) & (uint32(1<<22) - 1)) | 1
	}
	return 0
}

func (h *HistoryTable) PrintMax() {
	var val uint32
	for i := 0; i < 2; i++ {
		for j := 0; j < 8; j++ {
			for k := 0; k < 64; k++ {
				if h[i][j][k] > val {
					val = h[i][j][k]
				}
			}
		}
	}
	fmt.Printf("%d\n", val)
}
