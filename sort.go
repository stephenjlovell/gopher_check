//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

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
	occ := brd.AllOccupied()
	if IsAttackedBy(brd, occ&sqMaskOff[from],
		to, brd.Enemy(), brd.c) { // defended
		see := GetSee(brd, occ, from, to, NO_PIECE)
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
	occ := brd.AllOccupied()
	if IsAttackedBy(brd, occ&sqMaskOff[from], to, brd.Enemy(), brd.c) { // defended
		return uint32(SORT_WINNING_PROMOTION + GetSee(brd, occ, from, to, capturedPiece))
	} else { // undefended
		return SORT_WINNING_PROMOTION | uint32(promotedTo.PromoteValue()+capturedPiece.Value())
	}
}

// TODO: test with SEE as primary ordering factor.
func SortCapture(victim, attacker Piece, see int) uint32 {
	return (mVVLVA(victim, attacker) << 22) + uint32(see-SEE_MIN)
}

func mVVLVA(victim, attacker Piece) uint32 {
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

func (l MoveList) Sort() {
	l.qSort()
	// assert(l.isSorted(), "list not sorted")
}

func (l MoveList) InsertionSort() {
	l.insertionSort(0, len(l))
}

func (l MoveList) Len() int { return len(l) }

func (l MoveList) Less(i, j int) bool { return l[i].order > l[j].order }

func (l MoveList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l *MoveList) Push(item SortItem) {
	*l = append(*l, item)
}

func (l MoveList) insertionSort(a, b int) {
	for i := a + 1; i < b; i++ {
		for j := i; j > a && l.Less(j, j-1); j-- {
			l.Swap(j, j-1)
		}
	}
}

func (l MoveList) qSort() {
	if len(l) < 9 {
		if len(l) < 2 {
			return
		}
		l.insertionSort(0, len(l)) // use insertion sort for shorter slices.
	}
	left, right := 0, len(l)-1
	pivotIndex := right >> 1  // initial pivot location
	l.Swap(pivotIndex, right) // Move the pivot to the right

	for i := range l {
		if l.Less(i, right) { // Pile elements larger than the pivot on the left
			l.Swap(i, left)
			left++
		}
	}

	l.Swap(left, right) // relocate pivot
	l[:left].qSort()
	l[left+1:].qSort()
}
