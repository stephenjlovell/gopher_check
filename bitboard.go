//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
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

func setup_square_masks() {
	for i := 0; i < 64; i++ {
		sq_mask_on[i] = BB(1 << uint(i))
		sq_mask_off[i] = (^sq_mask_on[i])
		mask_of_length[i] = sq_mask_on[i] - 1
	}
}

func setup_pawn_masks() {
	var sq int
	for i := 0; i < 64; i++ {

		pawn_side_masks[i] = (king_masks[i] & row_masks[row(i)])

		if i < 56 {
			for j := 0; j < 2; j++ {
				sq = i + pawn_attack_offsets[j]
				if manhattan_distance(sq, i) == 2 {
					pawn_attack_masks[WHITE][i].Add(sq)
				}
			}
		}
		if i > 7 {
			for j := 2; j < 4; j++ {
				sq = i + pawn_attack_offsets[j]
				if manhattan_distance(sq, i) == 2 {
					pawn_attack_masks[BLACK][i].Add(sq)
				}
			}
		}
	}
}

func setup_knight_masks() {
	var sq int
	for i := 0; i < 64; i++ {
		for j := 0; j < 8; j++ {
			sq = i + knight_offsets[j]
			if on_board(sq) && manhattan_distance(sq, i) == 3 {
				knight_masks[i] |= sq_mask_on[sq]
			}
		}
	}
}

func setup_bishop_masks() {
	var previous, current, offset int
	for i := 0; i < 64; i++ {
		for j := 0; j < 4; j++ {
			previous = i
			offset = bishop_offsets[j]
			current = i + offset
			for on_board(current) && manhattan_distance(current, previous) == 2 {
				ray_masks[j][i].Add(current)
				previous = current
				current += offset
			}
		}
		bishop_masks[i] = ray_masks[NW][i] | ray_masks[NE][i] | ray_masks[SE][i] | ray_masks[SW][i]
	}
}

func setup_rook_masks() {
	var previous, current, offset int
	for i := 0; i < 64; i++ {
		for j := 0; j < 4; j++ {
			previous = i
			offset = rook_offsets[j]
			current = i + offset
			for on_board(current) && manhattan_distance(current, previous) == 1 {
				ray_masks[j+4][i].Add(current)
				previous = current
				current += offset
			}
		}
		rook_masks[i] = ray_masks[NORTH][i] | ray_masks[SOUTH][i] | ray_masks[EAST][i] | ray_masks[WEST][i]
	}
}

func setup_queen_masks() {
	for i := 0; i < 64; i++ {
		queen_masks[i] = bishop_masks[i] | rook_masks[i]
	}
}

func setup_king_masks() {
	var sq int
	for i := 0; i < 64; i++ {
		for j := 0; j < 8; j++ {
			sq = i + king_offsets[j]
			if on_board(sq) && manhattan_distance(sq, i) <= 2 {
				king_masks[i].Add(sq)
			}
		}
	}
}

func setup_row_masks() {
	row_masks[0] = 0xff // set the first row to binary 11111111, or 255.
	for i := 1; i < 8; i++ {
		row_masks[i] = (row_masks[i-1] << 8) // create the remaining rows by shifting the previous
	} // row up by 8 squares.
}

func setup_column_masks() {
	column_masks[0] = 1
	for i := 0; i < 8; i++ {
		column_masks[0] |= column_masks[0] << 8
	} // set the first column
	for i := 1; i < 8; i++ {
		column_masks[i] = (column_masks[i-1] << 1)
	} // create the remaining columns by transposing the
} // previous column rightward.

func setup_directions() {
	var ray BB
	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			directions[i][j] = DIR_INVALID // initialize array.
		}
	}
	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			for dir := 0; dir < 8; dir++ {
				ray = ray_masks[dir][i]
				if sq_mask_on[j]&ray != 0 {
					directions[i][j] = dir
					intervening[i][j] = ray ^ (ray_masks[dir][j] | sq_mask_on[j])
				}
			}
		}
	}
}

func setup_pawn_structure_masks() {
	for i := 0; i < 64; i++ { // initialize arrays
		pawn_passed_masks[WHITE][i] = 0
		pawn_passed_masks[BLACK][i] = 0
		pawn_isolated_masks[i] = 0
	}
	var sq, col int
	var center BB

	for i := 0; i < 64; i++ {
		col = column(i)
		pawn_isolated_masks[i] = (king_masks[i] & (^column_masks[col]))

		sq = i + 8
		for sq < 64 {
			pawn_passed_masks[WHITE][i] |= sq_mask_on[sq] // center row
			sq += 8
		}
		sq = i - 8
		for sq > 0 {
			pawn_passed_masks[BLACK][i] |= sq_mask_on[sq] // center row
			sq -= 8
		}
		center = pawn_passed_masks[WHITE][i]
		if col != 0 {
			pawn_passed_masks[WHITE][i] |= (center >> 1)
		} // queenside row
		if col != 7 {
			pawn_passed_masks[WHITE][i] |= (center << 1)
		} // kingside row
		center = pawn_passed_masks[BLACK][i]
		if col != 0 {
			pawn_passed_masks[BLACK][i] |= (center >> 1)
		} // queenside row
		if col != 7 {
			pawn_passed_masks[BLACK][i] |= (center << 1)
		} // kingside row
	}
}

func setup_castle_masks() {
	castle_queenside_intervening[WHITE] |= (sq_mask_on[B1] | sq_mask_on[C1] | sq_mask_on[D1])
	castle_kingside_intervening[WHITE] |= (sq_mask_on[F1] | sq_mask_on[G1])
	castle_queenside_intervening[BLACK] = (castle_queenside_intervening[WHITE] << 56)
	castle_kingside_intervening[BLACK] = (castle_kingside_intervening[WHITE] << 56)
}

func setup_masks() {
	setup_square_masks() // First set up masks used to add/remove bits by their index.
	setup_king_masks()   // For each square, calculate bitboard attack maps showing
	setup_knight_masks() // the squares to which the given piece type may move. These are
	setup_bishop_masks() // used as bitmasks during move generation to find pseudolegal moves.
	setup_rook_masks()
	setup_queen_masks()
	setup_row_masks() // Create bitboard masks for each row and column.
	setup_pawn_masks()
	setup_column_masks()
	setup_directions()
	setup_pawn_structure_masks()
	setup_castle_masks()

}
