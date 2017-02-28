//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "math/rand"

// Root Sorting
// At root, moves should be sorted based on subtree value rather than standard sorting.

// bit pos. (LSB order)
// 31  Winning promotions (1 bits)
// 30	 <<padding>> (1 bit)
// 29  Losing promotions  (1 bits)
// 28	 <<padding>> (1 bit)
// 22  MVV/LVA  (6 bits)  - Used to choose between captures of equal material gain/loss
// 0   History heuristic : (22 bits). Castles will always have the first bit set.

const (
	SORT_WINNING_PROMOTION = (1 << 31)
	SORT_LOSING_PROMOTION  = (1 << 29)
)

// Promotion Captures:
// if undefended, gain is promote_values[promoted_piece] + piece_values[captured_piece]
// is defended, gain is SEE score.
// Non-capture promotions:
// if square undefended, gain is promote_values[promoted_piece].
// If defended, gain is SEE score where captured_piece == EMPTY

func SortPromotionAdvances(brd *Board, from, to int, promotedTo Piece) uint32 {
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

func SortPromotionCaptures(brd *Board, from, to int, capturedPiece, promotedTo Piece) uint32 {
	if isAttackedBy(brd, brd.AllOccupied()&sqMaskOff[from], to, brd.Enemy(), brd.c) { // defended
		return uint32(SORT_WINNING_PROMOTION + getSee(brd, from, to, capturedPiece))
	} else { // undefended
		return SORT_WINNING_PROMOTION | uint32(promotedTo.PromoteValue()+capturedPiece.Value())
	}
}

func SortCapture(victim, attacker Piece, see int) uint32 {
	return (MVVLVA(victim, attacker) << 22) + uint32(see-SEE_MIN)
}

func MVVLVA(victim, attacker Piece) uint32 {
	return uint32(((victim + 1) << 3) - attacker)
}

type SortItem struct {
	order uint32
	move  Move
}

type MoveList []SortItem

// Generating non-captures usually results in longer move lists.
func NewMoveList(length int) MoveList {
	return make(MoveList, 0, length)
}

// sort.Sort() takes an interface. This prevents proper escape analysis by the compiler,
// resulting in additional heap allocations.
// TODO: write native sort implementation to replace package sort.
func (l MoveList) Sort() {
	// sort.Sort(l)
	l.QSort()

	// if len(l) > 1 && l[0].order > 0 {
	// 	printMutex.Lock()
	// 	fmt.Println(l)
	// 	printMutex.Unlock()
	// }
}

func (l MoveList) Len() int { return len(l) }

func (l MoveList) Less(i, j int) bool { return l[i].order > l[j].order }

func (l MoveList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l *MoveList) Push(item SortItem) {
	*l = append(*l, item)
}

// func (l *MoveList) BubbleSort() {
// }

func (l MoveList) QSort() MoveList {
	if len(l) < 2 {
		return l
	}
	left, right := 0, len(l)-1
	pivotIndex := rand.Int() % len(l) // initial pivot location
	// Move the pivot to the right
	l[pivotIndex], l[right] = l[right], l[pivotIndex]
	// Pile elements larger than the pivot on the left
	for i := range l {
		if l[i].order > l[right].order {
			l[i], l[left] = l[left], l[i]
			left++
		}
	}
	// relocate pivot
	l[left], l[right] = l[right], l[left]
	l[:left].QSort()
	l[left+1:].QSort()
	return l
}
