//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

const (
	PAWN_ENTRY_COUNT = 16384
	PAWN_TT_MASK     = PAWN_ENTRY_COUNT - 1
)

type PawnTT [PAWN_ENTRY_COUNT]PawnEntry

type PawnEntry struct {
	leftAttacks  [2]BB
	rightAttacks [2]BB
	allAttacks   [2]BB
	passedPawns  [2]BB
	value         [2]int
	key           uint32
	count         [2]uint8
}

func NewPawnTT() *PawnTT {
	return new(PawnTT)
}

// Typical hit rate is around 97 %
func (ptt *PawnTT) Probe(key uint32) *PawnEntry {
	return &ptt[key&PAWN_TT_MASK]
}
