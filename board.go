//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"sync"
)

const ( // color
	BLACK = iota
	WHITE
)

var printMutex sync.Mutex

// When spawning new goroutines for subtree search, a deep copy of the Board struct will have to be made
// and passed to the new goroutine.  Keep this struct as small as possible.
type Board struct {
	pieces          [2][6]BB  // 768 bits
	squares         [64]Piece // 512 bits
	occupied        [2]BB     // 128 bits
	material        [2]int32  // 64  bits
	hashKey        uint64    // 64  bits
	pawnHashKey   uint32    // 32  bits
	c               uint8     // 8   bits
	castle          uint8     // 8   bits
	enpTarget      uint8     // 8 	bits
	halfmoveClock  uint8     // 8 	bits
	endgameCounter uint8     // 8 	bits
	worker          *Worker
}

type BoardMemento struct { // memento object used to store board state to unmake later.
	hashKey       uint64
	pawnHashKey  uint32
	castle         uint8
	enpTarget     uint8
	halfmoveClock uint8
}

func (brd *Board) NewMemento() *BoardMemento {
	return &BoardMemento{
		hashKey:       brd.hashKey,
		pawnHashKey:  brd.pawnHashKey,
		castle:         brd.castle,
		enpTarget:     brd.enpTarget,
		halfmoveClock: brd.halfmoveClock,
	}
}

func (brd *Board) InCheck() bool { // determines if side to move is in check
	return isAttackedBy(brd, brd.AllOccupied(), brd.KingSq(brd.c), brd.Enemy(), brd.c)
}

func (brd *Board) KingSq(c uint8) int {
	// assert(brd.pieces[c][KING] > 0, "King missing from board")
	return furthestForward(c, brd.pieces[c][KING])
}

// Used to verify that a killer or hash move is legal.
func (brd *Board) LegalMove(m Move, inCheck bool) bool {
	if inCheck {
		return brd.EvadesCheck(m)
	} else {
		return brd.PseudolegalAvoidsCheck(m)
	}
}

// Moves generated while in check should already be legal, since we determine this
// as a side-effect of generating evasions.
func (brd *Board) AvoidsCheck(m Move, inCheck bool) bool {
	return inCheck || brd.PseudolegalAvoidsCheck(m)
}

func (brd *Board) PseudolegalAvoidsCheck(m Move) bool {
	switch m.Piece() {
	case PAWN:
		if m.CapturedPiece() == PAWN && brd.TypeAt(m.To()) == EMPTY { // En-passant
			return pinnedCanMove(brd, m.From(), m.To(), brd.c, brd.Enemy()) &&
				isPinned(brd, int(brd.enpTarget), brd.c, brd.Enemy())&sqMaskOn[m.To()] > 0
		} else {
			return pinnedCanMove(brd, m.From(), m.To(), brd.c, brd.Enemy())
		}
	case KING:
		return !isAttackedBy(brd, brd.AllOccupied(), m.To(), brd.Enemy(), brd.c)
	default:
		return pinnedCanMove(brd, m.From(), m.To(), brd.c, brd.Enemy())
	}
}

// Only called when in check
func (brd *Board) EvadesCheck(m Move) bool {
	piece, from, to := m.Piece(), m.From(), m.To()
	c, e := brd.c, brd.Enemy()

	if piece == KING {
		return !isAttackedBy(brd, occAfterMove(brd.AllOccupied(), from, to), to, e, c)
	} else {
		occ := brd.AllOccupied()
		kingSq := brd.KingSq(c)
		threats := colorAttackMap(brd, occ, kingSq, e, c)

		// TODO: EvadesCheck() called from non-check position in rare cases. Examples:
		// 5r1k/1b3p1p/pp3p1q/3n4/1P2R3/P2B1PP1/7P/6K1 w - - 0 1
		// 8/PPKR4/1Bn4P/3P3R/8/2p4r/pp4p1/r6k w - - 5 2  (r h3h5 x r)...?

		if threats == 0 {
			fmt.Println("EvadesCheck() called from non-check position!")
			m.Print()
			brd.PrintDetails()
			return true // no threats to evade.
		}

		if popCount(threats) > 1 {
			return false // only king moves can escape from double check.
		}
		if (threats|intervening[furthestForward(e, threats)][kingSq])&sqMaskOn[to] == 0 {
			return false // the moving piece must kill or block the attacking piece.
		}
		if !pinnedCanMove(brd, from, to, c, e) {
			return false // the moving piece can't be pinned to the king.
		}
		if brd.enpTarget != SQ_INVALID && piece == PAWN && m.CapturedPiece() == PAWN && // En-passant
			brd.TypeAt(to) == EMPTY {
			occ = occAfterMove(occ, from, to) & sqMaskOff[brd.enpTarget]

			return colorAttackMap(brd, occ, kingSq, e, c) == 0
			// return attacks_after_move(brd, occ, occ&brd.occupied[e], king_sq, e, c) == 0
		}
	}
	return true
}

func (brd *Board) ValidMove(m Move, inCheck bool) bool {
	if !m.IsMove() {
		return false
	}
	c, e := brd.c, brd.Enemy()
	piece, from, to, capturedPiece := m.Piece(), m.From(), m.To(), m.CapturedPiece()

	if brd.TypeAt(from) != piece || brd.pieces[c][piece]&sqMaskOn[from] == 0 {
		// fmt.Printf("No piece of this type available at from square!{%s}", m.ToString())
		return false
	}
	if sqMaskOn[to]&brd.occupied[c] > 0 {
		// fmt.Printf("To square occupied by own piece!{%s}", m.ToString())
		return false
	}
	if capturedPiece == KING || brd.pieces[c][KING] == 0 {
		// fmt.Printf("King capture detected!{%s}", m.ToString())
		return false
	}
	switch piece {
	case PAWN:
		var diff int
		if c == WHITE {
			diff = to - from
		} else {
			diff = from - to
		}
		if diff < 0 {
			// fmt.Printf("Invalid pawn movement direction!{%s}", m.ToString())
			return false
		} else if diff == 8 {
			return brd.TypeAt(to) == EMPTY
		} else if diff == 16 {
			return brd.TypeAt(to) == EMPTY && brd.TypeAt(pawnStopSq[c][from]) == EMPTY
		} else if capturedPiece == EMPTY {
			// fmt.Printf("Invalid pawn move!{%s}", m.ToString())
			return false
		} else {
			if capturedPiece == PAWN && brd.TypeAt(to) == EMPTY {
				if brd.enpTarget != SQ_INVALID && pawnStopSq[e][to] == int(brd.enpTarget) {
					return true
				} else {
					// fmt.Printf("Invalid En-passant move!{%s}", m.ToString())
					return false
				}
			} else {
				return brd.TypeAt(to) == capturedPiece
			}
		}
	case KING:
		if abs(to-from) == 2 { // validate castle moves
			occ := brd.AllOccupied()
			if c == WHITE && (brd.castle&12) > 0 {
				switch to {
				case C1:
					if !inCheck && (brd.castle&C_WQ > uint8(0)) && castleQueensideIntervening[WHITE]&brd.AllOccupied() == 0 &&
						!isAttackedBy(brd, occ, B1, e, c) && !isAttackedBy(brd, occ, C1, e, c) && !isAttackedBy(brd, occ, D1, e, c) {
						return true
					}
				case G1:
					if !inCheck && (brd.castle&C_WK > uint8(0)) && castleKingsideIntervening[WHITE]&brd.AllOccupied() == 0 &&
						!isAttackedBy(brd, occ, F1, e, c) && !isAttackedBy(brd, occ, G1, e, c) {
						return true
					}
				}
			} else if c == BLACK && (brd.castle&3) > 0 {
				switch to {
				case C8:
					if !inCheck && (brd.castle&C_BQ > uint8(0)) && castleQueensideIntervening[BLACK]&brd.AllOccupied() == 0 &&
						!isAttackedBy(brd, occ, B8, e, c) && !isAttackedBy(brd, occ, C8, e, c) && !isAttackedBy(brd, occ, D8, e, c) {
						return true
					}
				case G8:
					if !inCheck && (brd.castle&C_BK > uint8(0)) && castleKingsideIntervening[BLACK]&brd.AllOccupied() == 0 &&
						!isAttackedBy(brd, occ, F8, e, c) && !isAttackedBy(brd, occ, G8, e, c) {
						return true
					}
				}
			}
			// fmt.Printf("Invalid castle move!{%s}. Castle rights: %d", m.ToString(), brd.castle)
			return false
		}
	case KNIGHT: // no special treatment needed for knights.

	default:
		if slidingAttacks(piece, brd.AllOccupied(), from)&sqMaskOn[to] == 0 {
			// fmt.Printf("Invalid sliding attack!{%s}", m.ToString())
			return false
		}
	}

	if brd.TypeAt(to) != capturedPiece {
		// fmt.Printf("Captured piece not found on to square!{%s}", m.ToString())
		return false
	}

	return true
}

func (brd *Board) MayPromote(m Move) bool {
	if m.Piece() != PAWN {
		return false
	}
	if m.IsPromotion() {
		return true
	}
	if brd.c == WHITE {
		return m.To() >= A5 || brd.isPassedPawn(m)
	} else {
		return m.To() < A5 || brd.isPassedPawn(m)
	}
}

func (brd *Board) isPassedPawn(m Move) bool {
	return pawnPassedMasks[brd.c][m.To()]&brd.pieces[brd.Enemy()][PAWN] == 0
}

func (brd *Board) ValueAt(sq int) int {
	return brd.squares[sq].Value()
}

func (brd *Board) TypeAt(sq int) Piece {
	return brd.squares[sq]
}

func (brd *Board) Enemy() uint8 {
	return brd.c ^ 1
}

func (brd *Board) AllOccupied() BB { return brd.occupied[0] | brd.occupied[1] }

func (brd *Board) Placement(c uint8) BB { return brd.occupied[c] }

func (brd *Board) PawnsOnly() bool {
	return brd.occupied[brd.c] == brd.pieces[brd.c][PAWN]|brd.pieces[brd.c][KING]
}

func (brd *Board) ColorPawnsOnly(c uint8) bool {
	return brd.occupied[c] == brd.pieces[c][PAWN]|brd.pieces[c][KING]
}

func (brd *Board) Copy() *Board {
	return &Board{
		pieces:          brd.pieces,
		squares:         brd.squares,
		occupied:        brd.occupied,
		material:        brd.material,
		hashKey:        brd.hashKey,
		pawnHashKey:   brd.pawnHashKey,
		c:               brd.c,
		castle:          brd.castle,
		enpTarget:      brd.enpTarget,
		halfmoveClock:  brd.halfmoveClock,
		endgameCounter: brd.endgameCounter,
	}
}

func (brd *Board) PrintDetails() {
	pieceNames := [6]string{"Pawn", "Knight", "Bishop", "Rook", "Queen", "KING"}
	sideNames := [2]string{"White", "Black"}
	printMutex.Lock()

	fmt.Printf("hashKey: %x, pawnHashKey: %x\n", brd.hashKey, brd.pawnHashKey)
	fmt.Printf("castle: %d, enpTarget: %d, halfmoveClock: %d\noccupied:\n", brd.castle, brd.enpTarget, brd.halfmoveClock)
	for i := 0; i < 2; i++ {
		fmt.Printf("side: %s, material: %d\n", sideNames[i], brd.material[i])
		brd.occupied[i].Print()
		for pc := 0; pc < 6; pc++ {
			fmt.Printf("%s\n", pieceNames[pc])
			brd.pieces[i][pc].Print()
		}
	}
	printMutex.Unlock()
	brd.Print()
}

func (brd *Board) Print() {
	printMutex.Lock()
	if brd.c == WHITE {
		fmt.Println("\nSide to move: WHITE")
	} else {
		fmt.Println("\nSide to move: BLACK")
	}
	fmt.Printf("    A   B   C   D   E   F   G   H\n")
	fmt.Printf("  ---------------------------------\n")
	row := brd.squares[56:]
	fmt.Printf("8 ")
	brd.PrintRow(56, row)

	for i := 48; i >= 0; i -= 8 {
		row = brd.squares[i : i+8]
		fmt.Printf("%v ", 1+(i/8))
		brd.PrintRow(i, row)
	}
	fmt.Printf("    A   B   C   D   E   F   G   H\n")
	printMutex.Unlock()
}

func (brd *Board) PrintRow(start int, row []Piece) {
	fmt.Printf("| ")
	for i, piece := range row {
		if piece == EMPTY {
			fmt.Printf("  | ")
		} else {
			if brd.occupied[WHITE]&sqMaskOn[start+i] > 0 {
				fmt.Printf("%v | ", pieceGraphics[WHITE][piece])
			} else {
				fmt.Printf("%v | ", pieceGraphics[BLACK][piece])
			}
		}
	}
	fmt.Printf("\n  ---------------------------------\n")
}

func EmptyBoard() *Board {
	brd := &Board{
		enpTarget: SQ_INVALID,
	}
	for sq := 0; sq < 64; sq++ {
		brd.squares[sq] = EMPTY
	}
	return brd
}

func onBoard(sq int) bool { return 0 <= sq && sq <= 63 }
func row(sq int) int       { return sq >> 3 }
func column(sq int) int    { return sq & 7 }

var pieceGraphics = [2][6]string{
	{"\u265F", "\u265E", "\u265D", "\u265C", "\u265B", "\u265A"},
	{"\u2659", "\u2658", "\u2657", "\u2656", "\u2655", "\u2654"},
}

const (
	A1 = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
	SQ_INVALID
)
