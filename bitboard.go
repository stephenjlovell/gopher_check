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
	fmt.Printf("%d\n", b)
	for i := 63; i >= 0; i-- {
		if sq_mask_on[i]&b > 0 {
			sq = " 1"
		} else {
			sq = " 0"
		}
		row = sq + row
		if i%8 == 0 {
			fmt.Printf("%s\n", row)
			row = ""
		}
	}
	fmt.Printf("\n")
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
