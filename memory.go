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
	"math/rand"
)

const (
	SLOT_COUNT   = 131072         // number of main TT slots. 4 buckets per slot.
	TT_MASK      = SLOT_COUNT - 1 // a set bitmask of length 17
)

const (
	NO_MATCH = iota
	ORDERING_ONLY
	AVOID_NULL
	MATCH_FOUND
)

const (
	LOWER_BOUND = iota
	EXACT
	UPPER_BOUND
)

var main_tt TT

func setup_main_tt() {
	for i, _ := range main_tt {
		main_tt[i] = &Slot{ Bucket{}, Bucket{}, Bucket{}, Bucket{} }
	}
}

type TT [SLOT_COUNT]*Slot

type Slot [4]Bucket

// data stores the following: (54 bits total)
// depth remaining - 5 bits
// move - 21 bits
// bound/node type (exact, upper, lower) - 2 bits
// value - 17 bits
// search id (age of entry) - 9 bits
type Bucket struct {
	key  uint64
	data uint64
}

// XOR out b.data and return the original hash key.  If b.data has been modified by another goroutine
// due to a race condition, the key returned will no longer match and probe() will reject the entry.
func (b *Bucket) HashKey() uint64 {
	return (b.key ^ b.data)
}

func (b *Bucket) Depth() int {
	return int(b.data & uint64(31))
}
func (b *Bucket) Move() Move {
	return Move((b.data >> 5) & uint64(2097151))
}
func (b *Bucket) Type() int {
	return int((b.data >> 26) & uint64(3))
}
func (b *Bucket) Value() int {
	return int((b.data >> 28) & uint64(131071))
}
func (b *Bucket) Id() int {
	return int((b.data >> 45) & uint64(511))
}

func NewBucket(hash_key uint64, move Move, depth, entry_type, value int) Bucket {
	entry_data := uint64(depth) // maximum allowable depth of 31
	entry_data |= (uint64(move) << 5) | (uint64(entry_type) << 26) |
		(uint64(value) << 28) | (uint64(search_id) << 45)
	return Bucket{
		key:  (hash_key ^ entry_data), // XOR in entry_data to provide a way to check for race conditions.
		data: entry_data,
	}
}

func (tt *TT) get_slot(hash_key uint64) *Slot {
	return tt[hash_key&TT_MASK]
}

func (tt *TT) probe(brd *Board, depth, null_depth, alpha, beta int, value *int) (Move, int) {
	hash_key := brd.hash_key
	slot := tt.get_slot(hash_key)

	for i := 0; i < 4; i++ {
		if hash_key == slot[i].HashKey() { // look for an entry uncorrupted by lockless access.

			// to do: update age (search id) of entry.
			entry_depth := slot[i].Depth()

			if entry_depth >= depth {
				entry_type := slot[i].Type()
				entry_value := slot[i].Value()
				*value = entry_value // set the initial value for the subtree search at this node.
				switch entry_type {
				case LOWER_BOUND:
					if entry_value <= alpha {
						return slot[i].Move(), MATCH_FOUND
					}
				case UPPER_BOUND:
					if entry_value >= beta {
						return slot[i].Move(), MATCH_FOUND
					}
				case EXACT:
					if alpha < entry_value && entry_value < beta {
						// to do: if exact entry is valid for current bounds, save the PV.
					}
					return slot[i].Move(), MATCH_FOUND
				}
			} else if entry_depth >= null_depth {
				entry_type := slot[i].Type()
				entry_value := slot[i].Value()
				// if the entry is too shallow for an immediate cutoff but at least as deep as a potential
				// null-move search, check if a null move search would have any chance of causing a beta cutoff.
				if entry_type == UPPER_BOUND && entry_value < beta {
					return slot[i].Move(), AVOID_NULL
				}
			}
			return slot[i].Move(), ORDERING_ONLY
			// break
		}
	}
	return Move(0), NO_MATCH
}

// use lockless storing to avoid concurrent write issues without incurring locking overhead.
func (tt *TT) store(brd *Board, move Move, depth, entry_type, value int) {
	hash_key := brd.hash_key
	slot := tt.get_slot(hash_key)

	// to do:  store PV for exact entries.

	for i := 0; i < 4; i++ {
		if hash_key == slot[i].HashKey() { // exact match found.  Always replace.
			slot[i] = NewBucket(hash_key, move, depth, entry_type, value)
			return
		}
	}
	// If entries from a previous search exist, find/replace shallowest old entry.
	replace_index, replace_depth := 4, 32
	for i := 0; i < 4; i++ {
		if search_id != slot[i].Id() { // entry is not from the current search.
			if slot[i].Depth() < replace_depth {
				replace_index, replace_depth = i, slot[i].Depth()
			}
		}
	}
	if replace_index != 4 {
		slot[replace_index] = NewBucket(hash_key, move, depth, entry_type, value)
		return
	}
	// No exact match or entry from previous search found. Replace the shallowest entry.
	replace_index, replace_depth = 4, 32
	for i := 0; i < 4; i++ {
		if slot[i].Depth() < replace_depth {
			replace_index, replace_depth = i, slot[i].Depth()
		}
	}
	slot[replace_index] = NewBucket(hash_key, move, depth, entry_type, value)

}

// Zobrist Hashing
//
// Each possible square and piece combination is assigned a unique 64-bit integer key at startup.
// A unique hash key for a chess position can be generated by merging (via XOR) the keys for each
// piece/square combination, and merging in keys representing the side to move, castling rights,
// and any en-passant target square.

var zobrist_table [2][8][64]uint64 // keep array dimensions powers of 2 for faster array access.

var enp_table [65]uint64 // integer keys representing the en-passant target square, if any.
var castle_table [16]uint64
var side_key = random_key() // Integer key representing a change in side-to-move.

func random_key() uint64 { // create a pseudorandom 64-bit unsigned int key
	return uint64(rand.Uint32()) | (uint64(rand.Uint32()) << 32)
}

func setup_zobrist() {
	for c := 0; c < 2; c++ {
		for pc := 0; pc < 6; pc++ {
			for sq := 0; sq < 64; sq++ {
				zobrist_table[c][pc][sq] = random_key()
			}
		}
	}
	for i := 0; i < 16; i++ {
		castle_table[i] = random_key()
	}
	for sq := 0; sq < 64; sq++ {
		enp_table[sq] = random_key()
	}
	enp_table[64] = 0
}

func zobrist(pc Piece, sq int, c uint8) uint64 {
	return zobrist_table[c][pc][sq]
}

func enp_zobrist(sq int) uint64 {
	return enp_table[sq]
}

func castle_zobrist(castle uint8) uint64 {
	return castle_table[castle]
}
