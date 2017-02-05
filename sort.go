//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "sort"

// Root Sorting
// At root, moves should be sorted based on subtree value rather than standard sorting.

// bit pos. (LSB order)
// 30  Winning promotions (1 bits)
// 29  Losing promotions  (1 bits)
// 28	 <<padding>> (1 bit)
// 22  MVV/LVA  (6 bits)  - Used to choose between captures of equal material gain/loss
// 1   History heuristic : (21 bits)
// 0 	 Castles  (1 bit)

const (
	SORT_WINNING_PROMOTION = (1 << 30)
	SORT_LOSING_PROMOTION  = (1 << 29)
)

// TODO: promotions not ordered correctly.

// Promotion Captures:
// if undefended, gain is promote_values[promoted_piece] + piece_values[captured_piece]
// is defended, gain is SEE score.
// Non-capture promotions:
// if square undefended, gain is promote_values[promoted_piece].
// If defended, gain is SEE score where captured_piece == EMPTY

func sortPromotionAdvances(brd *Board, from, to int, promotedTo Piece) uint32 {
	if isAttackedBy(brd, brd.AllOccupied()&sqMaskOff[from],
		to, brd.Enemy(), brd.c) { // defended
		see := getSee(brd, from, to, EMPTY)
		if see >= 0 {
			return SORT_WINNING_PROMOTION | uint32(see)
		} else {
			return uint32(SORT_LOSING_PROMOTION + see)
		}
	} else { // undefended
		return SORT_WINNING_PROMOTION | uint32(promotedTo.PromoteValue())
	}
}

func sortPromotionCaptures(brd *Board, from, to int, capturedPiece, promotedTo Piece) uint32 {
	if isAttackedBy(brd, brd.AllOccupied()&sqMaskOff[from], to, brd.Enemy(), brd.c) { // defended
		return uint32(SORT_WINNING_PROMOTION + getSee(brd, from, to, capturedPiece))
	} else { // undefended
		return SORT_WINNING_PROMOTION | uint32(promotedTo.PromoteValue()+capturedPiece.Value())
	}
}

func mvvLva(victim, attacker Piece) uint32 { // returns value between 0 and 64
	return uint32(((victim+1)<<3)-attacker) << 22
}

type SortItem struct {
	order uint32
	move  Move
}

type MoveList []SortItem

func NewMoveList() MoveList {
	return make(MoveList, 0, 8)
}

func (l *MoveList) Sort() {
	sort.Sort(l)
	// print_mutex.Lock()
	// fmt.Println("-------------")
	// for i, item := range *l {
	// 	fmt.Printf("%d  %s  %b\n", i+1, item.move.ToString(), item.order)
	// }
	// print_mutex.Unlock()
}

func (l MoveList) Len() int { return len(l) }

func (l MoveList) Less(i, j int) bool { return l[i].order > l[j].order }

func (l MoveList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l *MoveList) Push(item SortItem) {
	*l = append(*l, item)
}
