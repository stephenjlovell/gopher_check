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

type MagicData struct {
	BishopMagics [64]BB `json:"bishopMagics"`
	RookMagics   [64]BB `json:"rookMagics"`
}

// In testing, homogenous array move DB actually outperformed a 'Fancy'
// magic bitboard approach implemented using slices.
var bishopMagicMoves, rookMagicMoves [64][MAGIC_DB_SIZE]BB

var bishopMagics, rookMagics [64]BB
var bishopMagicMasks, rookMagicMasks [64]BB

func bishopAttacks(occ BB, sq int) BB {
	return bishopMagicMoves[sq][magicIndex(occ, bishopMagicMasks[sq], bishopMagics[sq])]
}

func rookAttacks(occ BB, sq int) BB {
	return rookMagicMoves[sq][magicIndex(occ, rookMagicMasks[sq], rookMagics[sq])]
}

func queenAttacks(occ BB, sq int) BB {
	return (bishopAttacks(occ, sq) | rookAttacks(occ, sq))
}

func magicIndex(occ, sqMask, magic BB) int {
	return int(((occ & sqMask) * magic) >> (64 - MAGIC_INDEX_SIZE))
}

// if magics have already been generated, just fetch them from 'magics.json'.
// otherwise, generate the magics and write them to disk.
func setupMagicMoveGen() {
	var wg sync.WaitGroup

	magicsNeeded := false
	if _, err := os.Stat(MAGICS_JSON); err == nil {
		if !loadMagics() { // if magics failed to load for any reason, we'll have to generate them.
			magicsNeeded = true
		}
	} else {
		magicsNeeded = true
	}
	if magicsNeeded {
		fmt.Printf("Calculating magics")
		wg.Add(64 * 2)
	}

	setupMagicsForPiece(magicsNeeded, &wg, &bishopMagicMasks, &bishopMasks, &bishopMagics,
		&bishopMagicMoves, generateBishopAttacks)
	setupMagicsForPiece(magicsNeeded, &wg, &rookMagicMasks, &rookMasks, &rookMagics,
		&rookMagicMoves, generateRookAttacks)

	if magicsNeeded {
		wg.Wait()
		writeMagicsToDisk()
		fmt.Printf("done!\n\n")
	}
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func writeMagicsToDisk() {
	magics := MagicData{
		BishopMagics: bishopMagics,
		RookMagics:   rookMagics,
	}

	f, err := os.Create(MAGICS_JSON)
	checkError(err)
	defer f.Close()

	data, err := json.Marshal(magics)
	checkError(err)

	_, err = f.Write(data) // write the magics to disk as JSON.
	checkError(err)
}

func loadMagics() (success bool) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Failure reading magics from disk.")
			success = false // recover any panic
		}
	}()

	data, err := ioutil.ReadFile(MAGICS_JSON)
	checkError(err)

	magics := &MagicData{}

	err = json.Unmarshal(data, magics)
	checkError(err)

	bishopMagics = magics.BishopMagics
	rookMagics = magics.RookMagics
	fmt.Printf("Magics read from disk.\n")
	return true
}

func setupMagicsForPiece(magicsNeeded bool, wg *sync.WaitGroup, magicMasks, masks, magics *[64]BB,
	moves *[64][MAGIC_DB_SIZE]BB, genFn func(BB, int) BB) {

	for sq := 0; sq < 64; sq++ {
		edgeMask := (columnMasks[0]|columnMasks[7])&(^columnMasks[column(sq)]) |
			(rowMasks[0]|rowMasks[7])&(^rowMasks[row(sq)])
		magicMasks[sq] = masks[sq] & (^edgeMask)

		// Enumerate all subsets of the sq_mask using the Carry-Rippler technique:
		// https://chessprogramming.wikispaces.com/Traversing+Subsets+of+a+Set#Enumerating%20All%20Subsets-All%20Subsets%20of%20any%20Set
		refAttacks, occupied := [MAGIC_DB_SIZE]BB{}, [MAGIC_DB_SIZE]BB{}
		n := 0
		for occ := BB(0); occ != 0 || n == 0; occ = (occ - magicMasks[sq]) & magicMasks[sq] {
			refAttacks[n] = genFn(occ, sq) // save the attack bitboard for each subset for later use.
			occupied[n] = occ
			n++ // count the number of subsets
		}

		if magicsNeeded {
			go func(sq int) { // Calculate a magic for square sq in parallel
				randGenerator := NewRngKiss(73) // random number generator optimized for finding magics
				i := 0
				for i < n {
					// try random numbers until a suitable candidate is found.
					for magics[sq] = randGenerator.RandomMagic(sq); popCount((magicMasks[sq]*magics[sq])>>(64-MAGIC_INDEX_SIZE)) < MAGIC_INDEX_SIZE; {
						magics[sq] = randGenerator.RandomMagic(sq)
					}
					// if the last candidate magic failed, clear out any attack maps already placed in the moves DB
					moves[sq] = [MAGIC_DB_SIZE]BB{}
					for i = 0; i < n; i++ {
						// verify the candidate magic will index each possible occupancy subset to either a new slot,
						// or a slot with the same attack map (only benign collisions are allowed).
						attack := &moves[sq][magicIndex(occupied[i], magicMasks[sq], magics[sq])]

						if *attack != BB(0) && *attack != refAttacks[i] {
							break // keep going unless we hit a harmful collision
						}
						*attack = refAttacks[i] // populate the moves DB so we can detect collisions.
					}
				} // if every possible occupancy has been mapped to the correct attack set, we are done.
				fmt.Printf(".")
				wg.Done()
			}(sq)
		} else {
			for i := 0; i < n; i++ {
				moves[sq][magicIndex(occupied[i], magicMasks[sq], magics[sq])] = refAttacks[i]
			}
		}
	}
}
