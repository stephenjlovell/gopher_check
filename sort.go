//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "sort"

// Root Sorting
// At root, moves should be sorted based on subtree value rather than standard sorting.

// bit pos. (LSB order)
// 28  Winning promotions  (1 bits)
// 1	 <<padding>> (1 bit)
// 22  MVV/LVA  (6 bits)  - Used to choose between captures of equal material gain/loss
// 1   History heuristic : (21 bits)
// 0 	 Castles  (1 bit)

const (
	SORT_WINNING_PROMOTION = (1 << 31)
	SORT_LOSING_PROMOTION  = (1 << 29)
)

// TODO: promotions not ordered correctly.

// Promotion Captures:
// if undefended, gain is promoteValues[promotedPiece] + pieceValues[capturedPiece]
// is defended, gain is SEE score.
// Non-capture promotions:
// if square undefended, gain is promoteValues[promotedPiece].
// If defended, gain is SEE score where capturedPiece == EMPTY

// func sortPromotion(brd *Board, m Move) uint64 {
// 	if isAttackedBy(brd, brd.AllOccupied()&sqMaskOff[m.From()],
// 		m.To(), brd.Enemy(), brd.c) { // defended
// 		if getSee(brd, m.From(), m.To(), m.CapturedPiece()) >= 0 {
// 			return SORT_WINNING_PROMOTION | mvvLva(m.CapturedPiece(), PAWN)
// 		} else {
// 			return mvvLva(m.CapturedPiece(), PAWN)
// 		}
// 	} else {
// 		// val = m.PromotedTo().PromoteValue() + m.CapturedPiece().Value() // undefended
// 		return SORT_WINNING_PROMOTION | mvvLva(m.CapturedPiece(), PAWN)
// 	}
// }

func sortPromotionAdvances(brd *Board, from, to int, promotedTo Piece) uint64 {
	if isAttackedBy(brd, brd.AllOccupied()&sqMaskOff[from],
		to, brd.Enemy(), brd.c) { // defended
		see := getSee(brd, from, to, EMPTY)
		if see >= 0 {
			return SORT_WINNING_PROMOTION | uint64(see)
		} else {
			return uint64(SORT_LOSING_PROMOTION + see)
		}
	} else { // undefended
		return SORT_WINNING_PROMOTION | uint64(promotedTo.PromoteValue())
	}
}

func sortPromotionCaptures(brd *Board, from, to int, capturedPiece, promotedTo Piece) uint64 {
	if isAttackedBy(brd, brd.AllOccupied()&sqMaskOff[from], to, brd.Enemy(), brd.c) { // defended
		return uint64(SORT_WINNING_PROMOTION + getSee(brd, from, to, capturedPiece))
	} else { // undefended
		return SORT_WINNING_PROMOTION | uint64(promotedTo.PromoteValue()+capturedPiece.Value())
	}
}

func mvvLva(victim, attacker Piece) uint64 { // returns value between 0 and 64
	return uint64(((victim+1)<<3)-attacker) << 22
}

type SortItem struct {
	order uint64
	move  Move
}

type MoveList []SortItem

func NewMoveList() MoveList {
	return make(MoveList, 0, 8)
}

func (l *MoveList) Sort() {
	sort.Sort(l)
	// printMutex.Lock()
	// fmt.Println("-------------")
	// for i, item := range *l {
	// 	fmt.Printf("%d  %s  %b\n", i+1, item.move.ToString(), item.order)
	// }
	// printMutex.Unlock()
}

func (l MoveList) Len() int { return len(l) }

func (l MoveList) Less(i, j int) bool { return l[i].order > l[j].order }

func (l MoveList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l *MoveList) Push(item SortItem) {
	*l = append(*l, item)
}
