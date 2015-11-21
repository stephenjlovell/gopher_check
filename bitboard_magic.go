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
	// "math/rand"
)

const (
	MAGIC_INDEX_SIZE = 12
	MAGIC_DB_SIZE    = 1 << MAGIC_INDEX_SIZE
)

var bishop_magic_moves, rook_magic_moves [64][MAGIC_DB_SIZE]BB
var bishop_magics, rook_magics [64]BB
var bishop_magic_masks, rook_magic_masks [64]BB

func bishop_attacks(occ BB, sq int) BB {
	return bishop_magic_moves[sq][magic_index(occ, bishop_magic_masks[sq], bishop_magics[sq])]
}

func rook_attacks(occ BB, sq int) BB {
	return rook_magic_moves[sq][magic_index(occ, rook_magic_masks[sq], rook_magics[sq])]
}

func magic_index(occ, sq_mask, magic BB) int {
	return int(((occ & sq_mask) * magic) >> (64 - MAGIC_INDEX_SIZE))
}

func setup_magic_move_gen() {
	fmt.Printf("Calculating magics")
	setup_magics_for_piece(&bishop_magic_masks, &bishop_masks, &bishop_magics, &bishop_magic_moves, generate_bishop_attacks)
	setup_magics_for_piece(&rook_magic_masks, &rook_masks, &rook_magics, &rook_magic_moves, generate_rook_attacks)
	fmt.Printf("\n")
}

func setup_magics_for_piece(magic_masks, masks, magics *[64]BB, moves *[64][MAGIC_DB_SIZE]BB, gen_fn func(BB, int) BB) {

	rand_generator := NewRngKiss(73)

	for sq := 0; sq < 64; sq++ {
		fmt.Printf(".")
		edge_mask := (column_masks[0]|column_masks[7])&(^column_masks[column(sq)]) |
			(row_masks[0]|row_masks[7])&(^row_masks[row(sq)])
		sq_mask := masks[sq] & (^edge_mask)

		magic_masks[sq] = sq_mask

		// Enumerate all subsets of the sq_mask using the Carry-Rippler technique:
		// https://chessprogramming.wikispaces.com/Traversing+Subsets+of+a+Set#Enumerating%20All%20Subsets-All%20Subsets%20of%20any%20Set
		ref_attacks, occupied := [MAGIC_DB_SIZE]BB{}, [MAGIC_DB_SIZE]BB{}
		n := 0
		for occ := BB(0); occ != 0 || n == 0; occ = (occ - sq_mask) & sq_mask {

			ref_attacks[n] = gen_fn(occ, sq) // save the attack bitboard for each subset for later use.
			occupied[n] = occ
			n++
		}

		// fmt.Printf("domain n = %d. calculating magic for square %d", n, sq)
		// Calculate a magic for the current square
		i := 0
		for i < n {
			// try random numbers until a suitable candidate is found.
			for magics[sq] = rand_generator.RandomMagic(sq); pop_count((sq_mask*magics[sq])>>(64-MAGIC_INDEX_SIZE)) < MAGIC_INDEX_SIZE; {
				magics[sq] = rand_generator.RandomMagic(sq)
			}

			// fmt.Printf(".")

			// if the last candidate magic failed, clear out any attack maps already placed in the moves DB
			moves[sq] = [MAGIC_DB_SIZE]BB{}

			for i = 0; i < n; i++ {
				// verify the candidate magic will index each possible occupancy subset to either a new slot,
				// or a slot with the same attack map (only useful collisions are allowed).
				attack := &moves[sq][magic_index(occupied[i], sq_mask, magics[sq])]

				if *attack != BB(0) && *attack != ref_attacks[i] {
					break
				}
				*attack = ref_attacks[i] // populate the moves DB so we can detect collisions.
			}
		}

		// fmt.Printf("\n   magic found for sq: %d, magic: %x\n", sq, magics[sq])
		// magics[sq].Print()
	}
}
