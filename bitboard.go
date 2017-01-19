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
	*b &= sq_mask_off[sq]
}

func (b *BB) Add(sq int) {
	*b |= sq_mask_on[sq]
}

func (b BB) Print() {
	row, sq := "", ""
	// fmt.Printf("%d\n", b)
	fmt.Println("  A B C D E F G H")
	for i := 63; i >= 0; i-- {
		if sq_mask_on[i]&b > 0 {
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

func sliding_attacks(piece Piece, occ BB, sq int) BB {
	switch piece {
	case BISHOP:
		return bishop_attacks(occ, sq)
	case ROOK:
		return rook_attacks(occ, sq)
	case QUEEN:
		return queen_attacks(occ, sq)
	default:
		return BB(0)
	}
}

// TODO: incorporate pawn_attacks() into movegen

func pawn_attacks(brd *Board, c uint8) (BB, BB) { // returns (left_attacks, right_attacks) separately
	if c == WHITE {
		return ((brd.pieces[WHITE][PAWN] & (^column_masks[0])) << 7), ((brd.pieces[WHITE][PAWN] & (^column_masks[7])) << 9)
	} else {
		return ((brd.pieces[BLACK][PAWN] & (^column_masks[7])) >> 7), ((brd.pieces[BLACK][PAWN] & (^column_masks[0])) >> 9)
	}
}

func generate_bishop_attacks(occ BB, sq int) BB {
	return scan_up(occ, NW, sq) | scan_up(occ, NE, sq) | scan_down(occ, SE, sq) | scan_down(occ, SW, sq)
}

func generate_rook_attacks(occ BB, sq int) BB {
	return scan_up(occ, NORTH, sq) | scan_up(occ, EAST, sq) | scan_down(occ, SOUTH, sq) | scan_down(occ, WEST, sq)
}

func scan_down(occ BB, dir, sq int) BB {
	ray := ray_masks[dir][sq]
	blockers := (ray & occ)
	if blockers > 0 {
		ray ^= (ray_masks[dir][msb(blockers)]) // chop off end of ray after first blocking piece.
	}
	return ray
}

func scan_up(occ BB, dir, sq int) BB {
	ray := ray_masks[dir][sq]
	blockers := (ray & occ)
	if blockers > 0 {
		ray ^= (ray_masks[dir][lsb(blockers)]) // chop off end of ray after first blocking piece.
	}
	return ray
}
