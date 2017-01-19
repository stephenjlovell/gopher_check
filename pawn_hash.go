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
	left_attacks  [2]BB
	right_attacks [2]BB
	all_attacks   [2]BB
	passed_pawns  [2]BB
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
