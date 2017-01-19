//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

const (
	MAGIC_INDEX_SIZE = 12
	MAGIC_DB_SIZE    = 1 << MAGIC_INDEX_SIZE
	MAGICS_JSON      = "magics.json"
)

// In testing, homogenous array move DB actually outperformed a 'Fancy'
// magic bitboard approach implemented using slices.
var bishop_magic_moves, rook_magic_moves [64][MAGIC_DB_SIZE]BB

var bishop_magics, rook_magics [64]BB
var bishop_magic_masks, rook_magic_masks [64]BB

func bishop_attacks(occ BB, sq int) BB {
	return bishop_magic_moves[sq][magic_index(occ, bishop_magic_masks[sq], bishop_magics[sq])]
}

func rook_attacks(occ BB, sq int) BB {
	return rook_magic_moves[sq][magic_index(occ, rook_magic_masks[sq], rook_magics[sq])]
}

func queen_attacks(occ BB, sq int) BB {
	return (bishop_attacks(occ, sq) | rook_attacks(occ, sq))
}

func magic_index(occ, sq_mask, magic BB) int {
	return int(((occ & sq_mask) * magic) >> (64 - MAGIC_INDEX_SIZE))
}

// if magics have already been generated, just fetch them from 'magics.json'.
// otherwise, generate the magics and write them to disk.
func setup_magic_move_gen() {
	var wg sync.WaitGroup

	magics_needed := false
	if _, err := os.Stat(MAGICS_JSON); err == nil {
		if !load_magics() {
			// if magics failed to load for any reason, we'll have to generate them.
			magics_needed = true
		}
	} else {
		magics_needed = true
		fmt.Printf("Calculating magics")
		wg.Add(64 * 2)
	}
	setup_magics_for_piece(magics_needed, &wg, &bishop_magic_masks, &bishop_masks, &bishop_magics,
		&bishop_magic_moves, generate_bishop_attacks)
	setup_magics_for_piece(magics_needed, &wg, &rook_magic_masks, &rook_masks, &rook_magics,
		&rook_magic_moves, generate_rook_attacks)

	if magics_needed {
		wg.Wait()
		write_magics_to_disk()
		fmt.Printf("done!\n\n")
	}
}

type MagicData struct {
	Bishop_magics [64]BB
	Rook_magics   [64]BB
}

func check_error(e error) {
	if e != nil {
		panic(e)
	}
}

func write_magics_to_disk() {
	magics := MagicData{
		Bishop_magics: bishop_magics,
		Rook_magics:   rook_magics,
	}

	f, err := os.Create(MAGICS_JSON)
	check_error(err)
	defer f.Close()

	data, err := json.Marshal(magics)

	check_error(err)
	f.Write(data) // write the magics to disk as JSON.
}

func load_magics() (success bool) {
	defer func() {
		if r := recover(); r != nil {
			success = false // recover any panic
		}
	}()

	data, err := ioutil.ReadFile(MAGICS_JSON)
	check_error(err)

	magics := &MagicData{}

	json.Unmarshal(data, magics) // will panic if malformed json present.

	bishop_magics = magics.Bishop_magics
	rook_magics = magics.Rook_magics
	fmt.Printf("Magics read from disk.\n")
	return true
}

func setup_magics_for_piece(magics_needed bool, wg *sync.WaitGroup, magic_masks, masks, magics *[64]BB,
	moves *[64][MAGIC_DB_SIZE]BB, gen_fn func(BB, int) BB) {

	for sq := 0; sq < 64; sq++ {
		edge_mask := (column_masks[0]|column_masks[7])&(^column_masks[column(sq)]) |
			(row_masks[0]|row_masks[7])&(^row_masks[row(sq)])
		magic_masks[sq] = masks[sq] & (^edge_mask)

		// Enumerate all subsets of the sq_mask using the Carry-Rippler technique:
		// https://chessprogramming.wikispaces.com/Traversing+Subsets+of+a+Set#Enumerating%20All%20Subsets-All%20Subsets%20of%20any%20Set
		ref_attacks, occupied := [MAGIC_DB_SIZE]BB{}, [MAGIC_DB_SIZE]BB{}
		n := 0
		for occ := BB(0); occ != 0 || n == 0; occ = (occ - magic_masks[sq]) & magic_masks[sq] {
			ref_attacks[n] = gen_fn(occ, sq) // save the attack bitboard for each subset for later use.
			occupied[n] = occ
			n++ // count the number of subsets
		}

		if magics_needed {
			go func(sq int) { // Calculate a magic for square sq in parallel
				rand_generator := NewRngKiss(73) // random number generator optimized for finding magics
				i := 0
				for i < n {
					// try random numbers until a suitable candidate is found.
					for magics[sq] = rand_generator.RandomMagic(sq); pop_count((magic_masks[sq]*magics[sq])>>(64-MAGIC_INDEX_SIZE)) < MAGIC_INDEX_SIZE; {
						magics[sq] = rand_generator.RandomMagic(sq)
					}
					// if the last candidate magic failed, clear out any attack maps already placed in the moves DB
					moves[sq] = [MAGIC_DB_SIZE]BB{}
					for i = 0; i < n; i++ {
						// verify the candidate magic will index each possible occupancy subset to either a new slot,
						// or a slot with the same attack map (only benign collisions are allowed).
						attack := &moves[sq][magic_index(occupied[i], magic_masks[sq], magics[sq])]

						if *attack != BB(0) && *attack != ref_attacks[i] {
							break // keep going unless we hit a harmful collision
						}
						*attack = ref_attacks[i] // populate the moves DB so we can detect collisions.
					}
				} // if every possible occupancy has been mapped to the correct attack set, we are done.
				fmt.Printf(".")
				wg.Done()
			}(sq)
		} else {
			for i := 0; i < n; i++ {
				moves[sq][magic_index(occupied[i], magic_masks[sq], magics[sq])] = ref_attacks[i]
			}
		}
	}
}
