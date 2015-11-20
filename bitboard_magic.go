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

var bishop_magic_moves, rook_magic_moves [64][MAGIC_DB_SIZE]BB // replace with variable size?
var bishop_magics, rook_magics [64]BB
var bishop_magic_masks, rook_magic_masks [64]BB


func bishop_attacks(occ BB, sq int) BB {
	return bishop_magic_moves[sq][magic_index(occ, bishop_magic_masks[sq], bishop_magics[sq])]
}

func rook_attacks(occ BB, sq int) BB {
	return rook_magic_moves[sq][magic_index(occ, rook_magic_masks[sq], rook_magics[sq])]
}

func magic_index(occ, mask, magic BB) int {
	return int(((occ&mask)*magic) >> (64-MAGIC_INDEX_SIZE))
}


func setup_magic_move_gen() {
	setup_magics_for_piece(&bishop_magic_masks, &bishop_masks, &bishop_magics, &bishop_magic_moves, generate_bishop_attacks)
	setup_magics_for_piece(&rook_magic_masks, &rook_masks, &rook_magics, &rook_magic_moves, generate_rook_attacks)
}

func setup_magics_for_piece(magic_masks, masks, magics *[64]BB, moves *[64][MAGIC_DB_SIZE]BB,  gen_fn func(BB, int) BB) {
	edge_mask := column_masks[0]|column_masks[7]|row_masks[0]|row_masks[7]


	var mask, magic BB
	for sq := 0; sq < 64; sq++ {
		ref, occupied := [MAGIC_DB_SIZE]BB{}, [MAGIC_DB_SIZE]BB{}
		mask = masks[sq]&(^edge_mask)
		magic_masks[sq] = mask

		// Enumerate all subsets of the mask using the Carry-Rippler technique:
		// https://chessprogramming.wikispaces.com/Traversing+Subsets+of+a+Set#Enumerating%20All%20Subsets-All%20Subsets%20of%20any%20Set
		n := 0
		for occ := BB(0); occ != 0; occ = (occ-mask) & mask {
			ref[n] = gen_fn(occ, sq) // save the attack bitboard for each subset for later use.
			occupied[n] = occ
			n++
		}

		// Calculate a magic for the current square
		for i := 0; i < n; {
			// try random numbers until a suitable candidate is found.
			for magics[sq] = random_magic(0); pop_count((mask*magic)>>56) < 6; magics[sq] = random_magic(0) {
			}
			for i = 0; i < n; i++ {
				// verify the candidate magic will correctly index for each possible occupancy subset
				attack := &moves[sq][magic_index(occupied[i], mask, magics[sq])]

				if *attack != BB(0) && *attack != ref[i] {
					break
				}
				*attack = ref[i];
			}
		}

	}
}

func random_magic(starter BB) BB {
	// will probably need a better starting-point than a pseudo-random number...
	return BB(random_key64())
}
