//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
)

const (
	ANY_SQUARE_MASK = (1 << 64) - 1
)

type BB uint64

func (b *BB) Clear(sq int) {
	*b &= sqMaskOff[sq]
}

func (b *BB) Add(sq int) {
	*b |= sqMaskOn[sq]
}

func (b BB) Print() {
	row, sq := "", ""
	// fmt.Printf("%d\n", b)
	fmt.Println("  A B C D E F G H")
	for i := 63; i >= 0; i-- {
		if sqMaskOn[i]&b > 0 {
			sq = " 1"
		} else {
			sq = " 0"
		}
		row = sq + row
		if i%8 == 0 {
			fmt.Printf("%d%s\n", (i/8)+1, row)
			row = ""
		}
	}
	fmt.Printf("  A B C D E F G H\n\n")
}

func slidingAttacks(piece Piece, occ BB, sq int) BB {
	switch piece {
	case BISHOP:
		return bishopAttacks(occ, sq)
	case ROOK:
		return rookAttacks(occ, sq)
	case QUEEN:
		return queenAttacks(occ, sq)
	default:
		return BB(0)
	}
}

// TODO: incorporate pawn_attacks() into movegen

func pawnAttacks(brd *Board, c uint8) (BB, BB) { // returns (left_attacks, right_attacks) separately
	if c == WHITE {
		return ((brd.pieces[WHITE][PAWN] & (^columnMasks[0])) << 7), ((brd.pieces[WHITE][PAWN] & (^columnMasks[7])) << 9)
	} else {
		return ((brd.pieces[BLACK][PAWN] & (^columnMasks[7])) >> 7), ((brd.pieces[BLACK][PAWN] & (^columnMasks[0])) >> 9)
	}
}

func generateBishopAttacks(occ BB, sq int) BB {
	return scanUp(occ, NW, sq) | scanUp(occ, NE, sq) | scanDown(occ, SE, sq) | scanDown(occ, SW, sq)
}

func generateRookAttacks(occ BB, sq int) BB {
	return scanUp(occ, NORTH, sq) | scanUp(occ, EAST, sq) | scanDown(occ, SOUTH, sq) | scanDown(occ, WEST, sq)
}

func scanDown(occ BB, dir, sq int) BB {
	ray := rayMasks[dir][sq]
	blockers := (ray & occ)
	if blockers > 0 {
		ray ^= (rayMasks[dir][msb(blockers)]) // chop off end of ray after first blocking piece.
	}
	return ray
}

func scanUp(occ BB, dir, sq int) BB {
	ray := rayMasks[dir][sq]
	blockers := (ray & occ)
	if blockers > 0 {
		ray ^= (rayMasks[dir][lsb(blockers)]) // chop off end of ray after first blocking piece.
	}
	return ray
}
