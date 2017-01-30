//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
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

var mainTt TT

func resetMainTt() {
	for i := 0; i < SLOT_COUNT; i++ {
		for j := 0; j < 4; j++ {
			// main_tt[i][j].key = uint64(0)
			// main_tt[i][j].data = uint64(NewData(NO_MOVE, 0, EXACT, NO_SCORE, 511))
			mainTt[i][j].Store(NewData(NO_MOVE, 0, EXACT, NO_SCORE, 511), uint64(0))
		}
	}
}

type TT [SLOT_COUNT]Slot
type Slot [4]Bucket // sized to fit in a single cache line

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

func NewData(move Move, depth, entryType, value, id int) BucketData {
	return BucketData(depth) | (BucketData(move) << 5) | (BucketData(entryType) << 26) |
		(BucketData(value+INF) << 28) | (BucketData(id) << 45)
}

func (b *Bucket) Store(newData BucketData, hashKey uint64) {
	atomic.StoreUint64(&b.data, uint64(newData))
	atomic.StoreUint64(&b.key, uint64(newData)^hashKey)
	// b.data = uint64(new_data)
	// b.key = (uint64(new_data) ^ hash_key)
}

func (b *Bucket) Load() (BucketData, BucketData) {
	return BucketData(atomic.LoadUint64(&b.data)), BucketData(atomic.LoadUint64(&b.key))
	// return BucketData(b.data), BucketData(b.key)
}

type BucketData uint64

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

func (data BucketData) NewID(id int) BucketData {
	return (data & BucketData(35184372088831)) | (BucketData(id) << 45)
}

func (tt *TT) getSlot(hashKey uint64) *Slot {
	return &tt[hashKey&TT_MASK]
}

// Use Hyatt's lockless hashing approach to avoid having to lock/unlock shared TT memory
// during parallel search:  https://cis.uab.edu/hyatt/hashing.html
func (tt *TT) probe(brd *Board, depth, nullDepth, alpha, beta int, score *int) (Move, int) {

	// return NO_MOVE, NO_MATCH  // uncomment to disable transposition table

	var data, key BucketData
	hashKey := brd.hashKey
	slot := tt.getSlot(hashKey)

	for i := 0; i < 4; i++ {
		data, key = slot[i].Load()

		// XOR out data to return the original hash key.  If data has been modified by another goroutine
		// due to a data race, the key returned will no longer match and probe() will reject the entry.
		if hashKey == uint64(data^key) { // look for an entry uncorrupted by lockless access.

			slot[i].Store(data.NewID(searchId), hashKey) // update age (search id) of entry.

			entryValue := data.Value()
			*score = entryValue // set the current search score

			entryDepth := data.Depth()
			if entryDepth >= depth {
				switch data.Type() {
				case LOWER_BOUND: // failed high last time (at CUT node)
					if entryValue >= beta {
						return data.Move(), (CUTOFF_FOUND | BETA_FOUND)
					}
					return data.Move(), BETA_FOUND
				case UPPER_BOUND: // failed low last time. (at ALL node)
					if entryValue <= alpha {
						return data.Move(), (CUTOFF_FOUND | ALPHA_FOUND)
					}
					return data.Move(), ALPHA_FOUND
				case EXACT: // score was inside bounds.  (at PV node)
					if entryValue > alpha && entryValue < beta {
						return data.Move(), (CUTOFF_FOUND | EXACT_FOUND)
					}
					return data.Move(), EXACT_FOUND
				}
			} else if entryDepth >= nullDepth {
				// if the entry is too shallow for an immediate cutoff but at least as deep as a potential
				// null-move search, check if a null move search would have any chance of causing a beta cutoff.
				if data.Type() == UPPER_BOUND && data.Value() < beta {
					return data.Move(), AVOID_NULL
				}
			}
			return data.Move(), ORDERING_ONLY
		}
	}
	return NO_MOVE, NO_MATCH
}

// use lockless storing to avoid concurrent write issues without incurring locking overhead.
func (tt *TT) store(brd *Board, move Move, depth, entryType, value int) {
	hashKey := brd.hashKey
	slot := tt.getSlot(hashKey)
	var key BucketData
	var data [4]BucketData

	newData := NewData(move, depth, entryType, value, searchId)

	for i := 0; i < 4; i++ {
		data[i], key = slot[i].Load()
		if hashKey == uint64(data[i]^key) {
			slot[i].Store(newData, hashKey) // exact match found.  Always replace.
			return
		}
	}
	// If entries from a previous search exist, find/replace shallowest old entry.
	replaceIndex, replaceDepth := 4, 32
	for i := 0; i < 4; i++ {
		if searchId != data[i].Id() { // entry is not from the current search.
			if data[i].Depth() < replaceDepth {
				replaceIndex, replaceDepth = i, data[i].Depth()
			}
		}
	}
	if replaceIndex != 4 {
		slot[replaceIndex].Store(newData, hashKey)
		return
	}
	// No exact match or entry from previous search found. Replace the shallowest entry.
	replaceIndex, replaceDepth = 4, 32
	for i := 0; i < 4; i++ {
		if data[i].Depth() < replaceDepth {
			replaceIndex, replaceDepth = i, data[i].Depth()
		}
	}
	slot[replaceIndex].Store(newData, hashKey)
}
