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
// TODO: change material to uint16
type Board struct {
	pieces         [2][8]BB  // 1024 bits
	squares        [64]Piece //  512 bits
	occupied       [2]BB     //  128 bits
	hashKey        uint64    //   64 bits
	worker         *Worker   //   64 bits
	material       [2]int16  //   32 bits
	pawnHashKey    uint32    //   32 bits
	c              uint8     //    8 bits
	castle         uint8     //    8 bits
	enpTarget      uint8     //    8 bits
	halfmoveClock  uint8     //    8 bits
	endgameCounter uint8     //    8 bits
	// ...24 bits padding
}

type BoardMemento struct { // memento object used to store board state to unmake later.
	hashKey       uint64
	pawnHashKey   uint32
	castle        uint8
	enpTarget     uint8
	halfmoveClock uint8
}

func (brd *Board) NewMemento() *BoardMemento {
	return &BoardMemento{
		hashKey:       brd.hashKey,
		pawnHashKey:   brd.pawnHashKey,
		castle:        brd.castle,
		enpTarget:     brd.enpTarget,
		halfmoveClock: brd.halfmoveClock,
	}
}

func (brd *Board) InCheck() bool { // determines if side to move is in check
	return isAttackedBy(brd, brd.AllOccupied(), brd.KingSq(brd.c), brd.Enemy(), brd.c)
}

func (brd *Board) KingSq(c uint8) int {
	return furthestForward(c, brd.pieces[c][KING])
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
		pieces:         brd.pieces,
		squares:        brd.squares,
		occupied:       brd.occupied,
		material:       brd.material,
		hashKey:        brd.hashKey,
		pawnHashKey:    brd.pawnHashKey,
		c:              brd.c,
		castle:         brd.castle,
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
func row(sq int) int      { return sq >> 3 }
func column(sq int) int   { return sq & 7 }

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
