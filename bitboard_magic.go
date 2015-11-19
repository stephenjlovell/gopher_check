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

const (
	MAGIC_INDEX_SIZE = 11
	MAGIC_DB_SIZE = 1 << MAGIC_INDEX_SIZE
)

var bishop_magic_moves, bishop_magic_moves BB[64][MAGIC_DB_SIZE] // replace with variable size

var bishop_magics, rook_magics BB[64]

var bishop_magic_masks, rook_magic_masks BB[64]


func bishop_attacks(occ BB, sq int) BB {
	return bishop_magic_moves[sq][((occ&bishop_magic_masks[sq])*bishop_magics[sq])>>(64-MAGIC_INDEX_SIZE)]
}

func rook_attacks(occ BB, sq int) BB {
	return rook_magic_moves[sq][((occ&rook_magic_masks[sq])*rook_magics[sq])>>(64-MAGIC_INDEX_SIZE)]
}

func setup_magic_move_gen() {
	setup_magic_masks()
	setup_magics()
	setup_magic_moves()
}

func setup_magic_masks() {
	edge_mask := column_masks[0]|column_masks[7]|row_masks[0]|row_masks[7]
	for i := 0; i < 64; i++ {
		bishop_magic_masks[i] = bishop_masks[i]&(^edge_mask)
		bishop_magic_shifts[i] = 64 - pop_count(bishop_magic_masks[i])

		rook_magic_masks[i]	=	rook_magic_masks[i]&(^edge_mask)
		rook_magic_shifts[i] = 64 - pop_count(rook_magic_shifts[i])
	}
}

// calculate 'magic numbers' used to index into the magic_moves tables.
func setup_magics() {

}

// For each possible key (occupancy & square mask), generate the move bitboard conventionally.
// Use the pre-computed magic number to cache the move bitboard to the appropriate slot in the magic_moves table.
func setup_magic_moves() {


}
