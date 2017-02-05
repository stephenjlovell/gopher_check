//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"sync/atomic"
)

type HistoryTable [2][8][64]uint32

// Store atomically adds count to the history table h.
func (h *HistoryTable) Store(m Move, c uint8, count int) {
	atomic.AddUint32(&h[c][m.Piece()][m.To()], uint32((count>>4)|1))
}

// Probe atomically reads the history table h.
func (h *HistoryTable) Probe(pc Piece, c uint8, to int) uint32 {
	v := atomic.LoadUint32(&h[c][pc][to])
	if v > 0 {
		return ((v >> 3) & uint32(1<<21)) | 1
	}
	return 0
}
