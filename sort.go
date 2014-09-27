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
	// "container/heap"
	"sort"
)

// Ordering: PV/hash (handled by search), promotions, winning captures, killers,
// 					 losing captures, quiet moves (history heuristic order)

// what is the range of return values for SEE function?  {-4900, 5000} (min 13 bits)
// if king saftey were handled properly by SEE, range would be {-780, 880} (min 11 bits)

// In MSB order:
// promotion captures : (10 bits) always set 1st bit
// promotions : (1 bit)
// winning captures : (10 bits)
// killer 1 / killer 2 :  (2 bits)
// castles : (1 bit)
// losing captures : (8 bits)
// history heuristic : (18 bits)
// hopeless captures : (10 bits)

// split up losing captures

const (
	SORT_CASTLE    = (1 << 36)
	SORT_K2        = (1 << 37)
	SORT_K1        = (1 << 38)
	SORT_PROMOTION = (1 << 49)
)

func mvv_lva(victim, attacker Piece) int { // returns value between 0 and 64
	return int((victim << 3) | attacker)
}

//refactor to take up less space

func SortPromotionCapture(see int) uint64 {
	return (uint64(see) | 1) << 50
}

func SortWinningCapture(see int) uint64 {
	return (uint64(see) | 1) << 39
}

// var piece_values = [8]int{100, 320, 333, 510, 880, 5000, 0, 0} // default piece values
// 													// 220  13, 177,  370
// 																 // 190

func SortLosingCapture(see int) uint64 {
	if see > -200 { // {-200, 0}
		return uint64(see+200) << 28 // 8 bits, losing capture
	} else { // {-780, -200}
		return uint64(see + 780) // 10 bits, hopeless capture
	}
}

func SortHistory(h int) uint64 {
	if h > 0 {
		return (((uint64(h) >> 5) & mask_of_length[18]) | 3) << 10
	}
	return uint64(1 << 10)
}

// 134217727

// "Promising" moves (winning captures, promotions, and killers) are searched sequentially.

type SortItem struct {
	move  Move
	order uint64
}

func (s SortItem) See() int {
	return int(((s.order >> 42) & 2047) - 780)
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

func (l *MoveList) Push(item *SortItem) {
	*l = append(*l, item)
}

func (l *MoveList) Dequeue() *SortItem {
	old := *l
	if len(old) == 0 {
		return nil
	}
	item := old[0]
	*l = old[1:]
	return item
}
