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
	"sync/atomic"
)

const (
	SLOT_COUNT = 1048576        // number of main TT slots. 4 buckets per slot.
	TT_MASK    = SLOT_COUNT - 1 // a set bitmask used to index into TT.
)

const (
	NO_MATCH = 1 << iota
	ORDERING_ONLY
	AVOID_NULL
	ALPHA_FOUND
	BETA_FOUND
	EXACT_FOUND
	CUTOFF_FOUND
)

const (
	LOWER_BOUND = iota
	EXACT
	UPPER_BOUND
)

var main_tt TT

func setup_main_tt() {
	for i, _ := range main_tt {
		main_tt[i] = &Slot{
			NewBucket(0, NO_MOVE, 0, 0, NO_SCORE),
			NewBucket(0, NO_MOVE, 0, 0, NO_SCORE),
			NewBucket(0, NO_MOVE, 0, 0, NO_SCORE),
			NewBucket(0, NO_MOVE, 0, 0, NO_SCORE),
		}
	}
}

type TT [SLOT_COUNT]*Slot

type Slot [4]Bucket // 512 bits

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

type BucketData uint64

func NewBucket(hash_key uint64, move Move, depth, entry_type, value int) Bucket {
	entry_data := uint64(depth) // maximum allowable depth of 31
	entry_data |= (uint64(move) << 5) | (uint64(entry_type) << 26) |
		(uint64(value+INF) << 28) | (uint64(search_id) << 45)
	return Bucket{
		key:  (hash_key ^ entry_data), // XOR in entry_data to provide a way to check for race conditions.
		data: entry_data,
	}
}

// XOR out b.data and return the original hash key.  If b.data has been modified by another goroutine
// due to a race condition, the key returned will no longer match and probe() will reject the entry.

// func (b Bucket) HashKey() uint64 {
// 	return (b.key ^ b.data)
// }
func (data BucketData) Depth() int {
	return int(uint64(data) & uint64(31))
}
func (data BucketData) Move() Move {
	return Move((uint64(data) >> 5) & uint64(2097151))
}
func (data BucketData) Type() int {
	return int((uint64(data) >> 26) & uint64(3))
}
func (data BucketData) Value() int {
	return int(((uint64(data) >> 28) & uint64(131071)) - INF)
}
func (data BucketData) Id() int {
	return int((uint64(data) >> 45) & uint64(511))
}

func (data BucketData) NewID(id int) uint64 {
	return (uint64(data) & uint64(35184372088831)) | (uint64(id) << 45)
}

// func (b *Bucket) Replace(hash_key uint64, move Move, depth, entry_type, value int) {
// 	// maximum allowable depth of 31
// 	entry_data := uint64(depth) | (uint64(move) << 5) | (uint64(entry_type) << 26) |
// 		(uint64(value+INF) << 28) | (uint64(search_id) << 45)
// 	b.key = (hash_key ^ entry_data) // XOR in entry_data to provide a way to check for race conditions.
// 	b.data = entry_data
// }

func NewData(move Move, depth, entry_type, value int) uint64 {
	return uint64(depth) | (uint64(move) << 5) | (uint64(entry_type) << 26) |
		(uint64(value+INF) << 28) | (uint64(search_id) << 45)
}


func (tt *TT) get_slot(hash_key uint64) *Slot {
	return tt[hash_key&TT_MASK]
}

// Use Hyatt's lockless hashing approach to avoid having to lock/unlock shared TT memory
// during parallel search:  https://cis.uab.edu/hyatt/hashing.html
func (tt *TT) probe(brd *Board, depth, null_depth, alpha, beta int, score *int) (Move, int) {

	// return NO_MOVE, NO_MATCH  // uncomment to disable transposition table

	var data, key BucketData
	hash_key := brd.hash_key
	slot := tt.get_slot(hash_key)

	for i := 0; i < 4; i++ {
		data = BucketData(atomic.LoadUint64(&slot[i].data))
		key = BucketData(atomic.LoadUint64(&slot[i].key))

		if hash_key == uint64(data ^ key) { // look for an entry uncorrupted by lockless access.

			fmt.Printf("Full Key match: %d", hash_key)

			atomic.StoreUint64(&slot[i].key, key.NewID(search_id)) // update age (search id) of entry.

			entry_value := data.Value()

			*score = entry_value // set the current search score

			entry_depth := data.Depth()
			if entry_depth >= depth {
				entry_type := data.Type()

				switch entry_type {
				case LOWER_BOUND: // failed high last time (at CUT node)
					if entry_value >= beta {
						return data.Move(), (CUTOFF_FOUND | BETA_FOUND)
					}
					return data.Move(), BETA_FOUND
				case UPPER_BOUND: // failed low last time. (at ALL node)
					if entry_value <= alpha {
						return data.Move(), (CUTOFF_FOUND | ALPHA_FOUND)
					}
					return data.Move(), ALPHA_FOUND
				case EXACT: // score was inside bounds.  (at PV node)
					if entry_value > alpha && entry_value < beta {
						// to do: if exact entry is valid for current bounds, save the full PV.
						return data.Move(), (CUTOFF_FOUND | EXACT_FOUND)
					}
					return data.Move(), EXACT_FOUND
				}

			} else if entry_depth >= null_depth {
				entry_type := data.Type()
				entry_value := data.Value()
				// if the entry is too shallow for an immediate cutoff but at least as deep as a potential
				// null-move search, check if a null move search would have any chance of causing a beta cutoff.
				if entry_type == UPPER_BOUND && entry_value < beta {
					return data.Move(), AVOID_NULL
				}
			}
			return data.Move(), ORDERING_ONLY
		}

	}
	return NO_MOVE, NO_MATCH
}

// use lockless storing to avoid concurrent write issues without incurring locking overhead.
func (tt *TT) store(brd *Board, move Move, depth, entry_type, value int) {
	hash_key := brd.hash_key
	slot := tt.get_slot(hash_key)
	var key BucketData
	var data [4]BucketData

	for i := 0; i < 4; i++ {
		data[i] = BucketData(atomic.LoadUint64(&slot[i].data))
		key = BucketData(atomic.LoadUint64(&slot[i].key))

		if hash_key == uint64(data[i] ^ key) { // exact match found.  Always replace.
			data[i] = BucketData(NewData(move, depth, entry_type, value))
			atomic.StoreUint64(&slot[i].data, uint64(data[i]))
			atomic.StoreUint64(&slot[i].key, uint64(data[i] ^ key))
			return
		}
	}
	// If entries from a previous search exist, find/replace shallowest old entry.
	replace_index, replace_depth := 4, 32
	for i := 0; i < 4; i++ {
		if search_id != data[i].Id() { // entry is not from the current search.
			if data[i].Depth() < replace_depth {
				replace_index, replace_depth = i, data[i].Depth()
			}
		}
	}
	if replace_index != 4 {
		atomic.StoreUint64(&slot[replace_index].data, uint64(data[replace_index]))
		atomic.StoreUint64(&slot[replace_index].key, uint64(data[replace_index] ^ key))
		return
	}
	// No exact match or entry from previous search found. Replace the shallowest entry.
	replace_index, replace_depth = 4, 32
	for i := 0; i < 4; i++ {
		if data[i].Depth() < replace_depth {
			replace_index, replace_depth = i, data[i].Depth()
		}
	}
	atomic.StoreUint64(&slot[replace_index].data, uint64(data[replace_index]))
	atomic.StoreUint64(&slot[replace_index].key, uint64(data[replace_index] ^ key))
}

// Zobrist Hashing
//
// Each possible square and piece combination is assigned a unique 64-bit integer key at startup.
// A unique hash key for a chess position can be generated by merging (via XOR) the keys for each
// piece/square combination, and merging in keys representing the side to move, castling rights,
// and any en-passant target square.
var pawn_zobrist_table [2][64]uint32
var zobrist_table [2][8][64]uint64 // keep array dimensions powers of 2 for faster array access.
var enp_table [65]uint64           // integer keys representing the en-passant target square, if any.
var castle_table [16]uint64

var side_key64 uint64 // keys representing a change in side-to-move.
var side_key32 uint32

const (
	MAX_RAND = (1 << 32) - 1
)

func setup_zobrist() {
	for c := 0; c < 2; c++ {
		for sq := 0; sq < 64; sq++ {
			pawn_zobrist_table[c][sq] = random_key32()
			for pc := 0; pc < 6; pc++ {
				zobrist_table[c][pc][sq] = random_key64()
			}
		}
	}
	for i := 0; i < 16; i++ {
		castle_table[i] = random_key64()
	}
	for sq := 0; sq < 64; sq++ {
		enp_table[sq] = random_key64()
	}
	enp_table[64] = 0
	side_key64 = random_key64()
	side_key32 = random_key32()
}

func zobrist(pc Piece, sq int, c uint8) uint64 {
	return zobrist_table[c][pc][sq]
}

func pawn_zobrist(sq int, c uint8) uint32 {
	return pawn_zobrist_table[c][sq]
}

func enp_zobrist(sq uint8) uint64 {
	return enp_table[sq]
}

func castle_zobrist(castle uint8) uint64 {
	return castle_table[castle]
}
