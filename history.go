//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"sync/atomic"
)

type HistoryTable [2][8][64]uint64

// Store atomically adds count to the history table h.
func (h *HistoryTable) Store(m Move, c uint8, count int) {
	atomic.AddUint64(&h[c][m.Piece()][m.To()], uint64((count>>2)|1))
}

// Probe atomically reads the history table h.
func (h *HistoryTable) Probe(pc Piece, c uint8, to int) uint64 {
	v := atomic.LoadUint64(&h[c][pc][to])
	if v > 0 {
		return ((((v >> 3) & maskOfLength[21]) | 1) << 1)
	}
	return 0
}
