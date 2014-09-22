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
	// "container/heap"
	"sort"
)

// Ordering: PV/hash (handled by search), promotions, winning captures, killers,
// 					 losing captures, quiet moves (history heuristic order)

// what is the range of return values for SEE function?  {-4900, 5000} (min 13 bits)
// if king saftey were handled properly by SEE, range would be {-780, 880} (min 11 bits)

// In MSB order:
// promotion captures : (11 bits) always set 1st bit
// winning captures : (11 bits)
// killer 1 / killer 2 :  (2 bits)
// castles : (1 bit)
// losing captures : (11 bits)
// history heuristic : (28 bits)  shift history score right by 2 bits
// 268435455

const (
	SORT_CASTLE    = (1 << 39)
	SORT_K2        = (1 << 40)
	SORT_K1        = (1 << 41)
	SORT_PROMOTION = (1 << 53)
)

func mvv_lva(victim, attacker Piece) int { // returns value between 0 and 64
	return int((victim << 3) | attacker)
}

func SortPromotionCapture(see int) uint64 {
	return ((uint64(see) + 780) | 1) << 53
}

func SortWinningCapture(see int) uint64 {
	return ((uint64(see) + 780) | 1) << 42
}

func SortLosingCapture(see int) uint64 {
	return ((uint64(see) + 780) | 1) << 28
}

func SortHistory(h int) uint64 {
	return (uint64(h) >> 2) & 268435455 // 28 bit-wide set bitmask
}

// "Promising" moves (winning captures, promotions, and killers) are searched sequentially.

type SortItem struct {
	move  Move
	order uint64
}

type MoveList []*SortItem

func (l *MoveList) Sort() {
	sort.Sort(l)
}

func (l MoveList) Len() int { return len(l) }

func (l MoveList) Less(i, j int) bool { return l[i].order > l[j].order }

func (l MoveList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l *MoveList) Push(sort_item interface{}) {
	*l = append(*l, sort_item.(*SortItem))
}
