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
// "fmt"
)

func setup_square_masks() {
	for i := 0; i < 64; i++ {
		sq_mask_on[i] = BB(1 << uint(i))
		sq_mask_off[i] = (^sq_mask_on[i])
		mask_of_length[i] = uint64(sq_mask_on[i] - 1)
	}
}

func setup_pawn_masks() {
	var sq int
	for i := 0; i < 64; i++ {
		pawn_side_masks[i] = (king_masks[i] & row_masks[row(i)])
		if i < 56 {
			pawn_stop_masks[WHITE][i] = sq_mask_on[i] << 8
			pawn_stop_sq[WHITE][i] = i + 8
			for j := 0; j < 2; j++ {
				sq = i + pawn_attack_offsets[j]
				if manhattan_distance(sq, i) == 2 {
					pawn_attack_masks[WHITE][i].Add(sq)
				}
			}
		}
		if i > 7 {
			pawn_stop_masks[BLACK][i] = sq_mask_on[i] >> 8
			pawn_stop_sq[BLACK][i] = i - 8
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
	var center BB
	for i := 0; i < 64; i++ {
		for j := 0; j < 8; j++ {
			sq = i + king_offsets[j]
			if on_board(sq) && manhattan_distance(sq, i) <= 2 {
				king_masks[i].Add(sq)
			}
		}
		center = king_masks[i] | sq_mask_on[i]
		// The king zone is the 3 x 4 square area consisting of the squares around the king and
		// the squares facing the enemy side.
		king_zone_masks[WHITE][i] = center | (center << 8)
		king_zone_masks[BLACK][i] = center | (center >> 8)
		// The king shield is the three squares adjacent to the king and closest to the enemy side.
		king_shield_masks[WHITE][i] = (king_zone_masks[WHITE][i] ^ center) >> 8
		king_shield_masks[BLACK][i] = (king_zone_masks[BLACK][i] ^ center) << 8
	}

}

func setup_row_masks() {
	row_masks[0] = 0xff // set the first row to binary 11111111, or 255.
	for i := 1; i < 8; i++ {
		row_masks[i] = (row_masks[i-1] << 8) // create the remaining rows by shifting the previous
	} // row up by 8 squares.
	// middle_rows = row_masks[2] | row_masks[3] | row_masks[4] | row_masks[5]
}

func setup_column_masks() {
	column_masks[0] = 1
	for i := 0; i < 8; i++ { // create the first column
		column_masks[0] |= column_masks[0] << 8
	}
	for i := 1; i < 8; i++ { // create the remaining columns by transposing the first column rightward.
		column_masks[i] = (column_masks[i-1] << 1)
	}
}

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
				if sq_mask_on[j]&ray > 0 {
					directions[i][j] = dir
					intervening[i][j] = ray ^ (ray_masks[dir][j] | sq_mask_on[j])
					line_masks[i][j] = ray | ray_masks[opposite_dir[dir]][j]
				}
			}
		}
	}
}

func setup_pawn_structure_masks() {
	var col int
	for i := 0; i < 64; i++ {
		col = column(i)
		pawn_isolated_masks[i] = (king_masks[i] & (^column_masks[col]))

		pawn_passed_masks[WHITE][i] = ray_masks[NORTH][i]
		pawn_passed_masks[BLACK][i] = ray_masks[SOUTH][i]
		if col < 7 {
			pawn_passed_masks[WHITE][i] |= pawn_passed_masks[WHITE][i] << BB(1)
			pawn_passed_masks[BLACK][i] |= pawn_passed_masks[BLACK][i] << BB(1)
		}
		if col > 0 {
			pawn_passed_masks[WHITE][i] |= pawn_passed_masks[WHITE][i] >> BB(1)
			pawn_passed_masks[BLACK][i] |= pawn_passed_masks[BLACK][i] >> BB(1)
		}

		pawn_attack_spans[WHITE][i] = pawn_passed_masks[WHITE][i] & (^column_masks[col])
		pawn_attack_spans[BLACK][i] = pawn_passed_masks[BLACK][i] & (^column_masks[col])

		pawn_front_spans[WHITE][i] = pawn_passed_masks[WHITE][i] & (column_masks[col])
		pawn_front_spans[BLACK][i] = pawn_passed_masks[BLACK][i] & (column_masks[col])

		pawn_doubled_masks[i] = pawn_front_spans[WHITE][i] | pawn_front_spans[BLACK][i]

		pawn_promote_sq[WHITE][i] = msb(pawn_front_spans[WHITE][i])
		pawn_promote_sq[BLACK][i] = lsb(pawn_front_spans[BLACK][i])
	}
}

func setup_castle_masks() {
	castle_queenside_intervening[WHITE] |= (sq_mask_on[B1] | sq_mask_on[C1] | sq_mask_on[D1])
	castle_kingside_intervening[WHITE] |= (sq_mask_on[F1] | sq_mask_on[G1])
	castle_queenside_intervening[BLACK] = (castle_queenside_intervening[WHITE] << 56)
	castle_kingside_intervening[BLACK] = (castle_kingside_intervening[WHITE] << 56)
}

func setup_masks() {
	setup_row_masks() // Create bitboard masks for each row and column.
	setup_column_masks()
	setup_square_masks() // First set up masks used to add/remove bits by their index.
	setup_knight_masks() // For each square, calculate bitboard attack maps showing
	setup_bishop_masks() // the squares to which the given piece type may move. These are
	setup_rook_masks()   // used as bitmasks during move generation to find pseudolegal moves.
	setup_queen_masks()
	setup_king_masks()
	setup_directions()
	setup_pawn_masks()
	setup_pawn_structure_masks()
	setup_castle_masks()
}
