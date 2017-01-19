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
// if undefended, gain is promote_values[promoted_piece] + piece_values[captured_piece]
// is defended, gain is SEE score.
// Non-capture promotions:
// if square undefended, gain is promote_values[promoted_piece].
// If defended, gain is SEE score where captured_piece == EMPTY

// func sort_promotion(brd *Board, m Move) uint64 {
// 	if is_attacked_by(brd, brd.AllOccupied()&sq_mask_off[m.From()],
// 		m.To(), brd.Enemy(), brd.c) { // defended
// 		if get_see(brd, m.From(), m.To(), m.CapturedPiece()) >= 0 {
// 			return SORT_WINNING_PROMOTION | mvv_lva(m.CapturedPiece(), PAWN)
// 		} else {
// 			return mvv_lva(m.CapturedPiece(), PAWN)
// 		}
// 	} else {
// 		// val = m.PromotedTo().PromoteValue() + m.CapturedPiece().Value() // undefended
// 		return SORT_WINNING_PROMOTION | mvv_lva(m.CapturedPiece(), PAWN)
// 	}
// }

func sort_promotion_advances(brd *Board, from, to int, promoted_to Piece) uint64 {
	if is_attacked_by(brd, brd.AllOccupied()&sq_mask_off[from],
		to, brd.Enemy(), brd.c) { // defended
		see := get_see(brd, from, to, EMPTY)
		if see >= 0 {
			return SORT_WINNING_PROMOTION | uint64(see)
		} else {
			return uint64(SORT_LOSING_PROMOTION + see)
		}
	} else { // undefended
		return SORT_WINNING_PROMOTION | uint64(promoted_to.PromoteValue())
	}
}

func sort_promotion_captures(brd *Board, from, to int, captured_piece, promoted_to Piece) uint64 {
	if is_attacked_by(brd, brd.AllOccupied()&sq_mask_off[from], to, brd.Enemy(), brd.c) { // defended
		return uint64(SORT_WINNING_PROMOTION + get_see(brd, from, to, captured_piece))
	} else { // undefended
		return SORT_WINNING_PROMOTION | uint64(promoted_to.PromoteValue()+captured_piece.Value())
	}
}

func mvv_lva(victim, attacker Piece) uint64 { // returns value between 0 and 64
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
