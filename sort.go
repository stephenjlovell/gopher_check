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
	"sort"
)

// Root Sorting
// At root, moves should be sorted based on subtree value rather than standard sorting.

// bit pos. (LSB order)
// 28  Winning promotions  (1 bits)
// 22  MVV/LVA  (6 bits)  - Used to choose between captures of equal material gain/loss
// 1   History heuristic : (21 bits)
// 0 	 Castles  (1 bit)

const (
	SORT_WINNING = (1 << 28)
)

func mvv_lva(victim, attacker Piece) uint64 { // returns value between 0 and 64
	return uint64(((victim+1)<<3)-attacker) << 22
}

// TODO: promotions not ordered correctly.

// Promotion Captures:
// if undefended, gain is promote_values[promoted_piece] + piece_values[captured_piece]
// is defended, gain is SEE score.
// Non-capture promotions:
// if square undefended, gain is promote_values[promoted_piece].
// If defended, gain is SEE score where captured_piece == EMPTY
func SortPromotion(brd *Board, m Move) uint64 {
	var val int
	if is_attacked_by(brd, brd.AllOccupied()&sq_mask_off[m.From()], m.To(), brd.Enemy(), brd.c) {
		val = get_see(brd, m.From(), m.To(), m.CapturedPiece()) // defended
	} else {
		val = m.PromotedTo().PromoteValue() + m.CapturedPiece().Value() // undefended
	}
	if val >= 0 {
		return SORT_WINNING | mvv_lva(m.CapturedPiece(), PAWN)
	} else {
		return mvv_lva(m.CapturedPiece(), PAWN)
	}
}

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

func (l *MoveList) Push(item *SortItem) {
	*l = append(*l, item)
}
