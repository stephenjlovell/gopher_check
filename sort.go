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

// Promotions and captures SEE >= 0  11 bits
// Killers  1 bit
// castles  1 bit
// Promotions and captures SEE < 0  10 bits
// history heuristic : (21 bits)

func SortPromotion(brd *Board, m Move) uint64 {
	// Promotion Captures:
	// if undefended, gain is promote_values[promoted_piece] + piece_values[captured_piece]
	// is defended, gain is SEE score.
	// Non-capture promotions:
	// if square undefended, gain is promote_values[promoted_piece].
	// If defended, gain is SEE score where captured_piece == EMPTY
	var val int
	if is_attacked_by(brd, m.To(), brd.Enemy(), brd.c) {
		val = get_see(brd, m.From(), m.To(), m.CapturedPiece())
	} else {
		val = promote_values[m.PromotedTo()] + m.CapturedPiece().Value()
	}
	if val >= 0 {
		return SortWinningCapture(val)
	} else {
		return SortLosingCapture(val)
	}
}

const (
	SORT_CASTLE = (1 << 31)
	SORT_KILLER = (1 << 32)
	WINNING     = (1 << 33)
)

// {-780, 1660}, 12 bits  promotions and captures

func SortWinningCapture(see int) uint64 { // 11 bits
	return uint64(see|1) << 33
}

func SortLosingCapture(see int) uint64 { // 10 bits
	return uint64((see+780)|1) << 21
}

// "Promising" moves (winning captures, promotions, and killers) are searched sequentially.

type SortItem struct {
	move  Move
	order uint64
}

// func (s SortItem) See() int {
// 	return int(((s.order >> 42) & mask_of_length[11]) - 780)
// }

func mvv_lva(victim, attacker Piece) int { // returns value between 0 and 64
	return int((victim << 3) | attacker)
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
