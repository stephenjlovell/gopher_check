//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
)

const (
	NO_MOVE = (Move(EMPTY) << 15) | (Move(EMPTY) << 18)
)

type Move uint32

// To to fit into transposition table entries, moves are encoded using 21 bits as follows (in LSB order):
// From square - first 6 bits
// To square - next 6 bits
// Piece - next 3 bits
// Captured piece - next 3 bits
// promoted to - next 3 bits

func (m Move) From() int {
	return int(uint32(m) & uint32(63))
}

func (m Move) To() int {
	return int((uint32(m) >> 6) & uint32(63))
}

func (m Move) Piece() Piece {
	return Piece((uint32(m) >> 12) & uint32(7))
}

func (m Move) CapturedPiece() Piece {
	return Piece((uint32(m) >> 15) & uint32(7))
}

func (m Move) PromotedTo() Piece {
	return Piece((uint32(m) >> 18) & uint32(7))
}

func (m Move) IsCapture() bool {
	return m.CapturedPiece() != EMPTY
}

func (m Move) IsPromotion() bool {
	return m.PromotedTo() != EMPTY
}

func (m Move) IsQuiet() bool {
	return !(m.IsCapture() || m.IsPromotion())
}

func (m Move) IsMove() bool {
	return m != 0 && m != NO_MOVE
}

var piece_chars = [6]string{"p", "n", "b", "r", "q", "k"}

func (m Move) Print() {
	fmt.Println(m.ToString())
}

func (m Move) ToString() string { // string representation used for debugging only.
	var str string
	if !m.IsMove() {
		return "NO_MOVE"
	}
	str += piece_chars[m.Piece()] + " "
	str += ParseCoordinates(row(m.From()), column(m.From()))
	str += ParseCoordinates(row(m.To()), column(m.To()))
	if m.IsCapture() {
		str += " x " + piece_chars[m.CapturedPiece()]
	}
	if m.IsPromotion() {
		str += " promoted to " + piece_chars[m.PromotedTo()]
	}
	return str
}

func (m Move) ToUCI() string {
	if !m.IsMove() {
		return "0000"
	}
	str := ParseCoordinates(row(m.From()), column(m.From())) +
		ParseCoordinates(row(m.To()), column(m.To()))
	if m.PromotedTo() != EMPTY {
		str += piece_chars[m.PromotedTo()]
	}
	return str
}

func NewMove(from, to int, piece, captured_piece, promoted_to Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) |
		(Move(captured_piece) << 15) | (Move(promoted_to) << 18)
}

func NewRegularMove(from, to int, piece Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) |
		(Move(EMPTY) << 15) | (Move(EMPTY) << 18)
}

func NewCapture(from, to int, piece, captured_piece Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) |
		(Move(captured_piece) << 15) | (Move(EMPTY) << 18)
}

// since moving piece is always PAWN (0) for promotions, no need to merge in the moving piece.
func NewPromotion(from, to int, piece, promoted_to Piece) Move {
	return Move(from) | (Move(to) << 6) | (Move(piece) << 12) |
		(Move(EMPTY) << 15) | (Move(promoted_to) << 18)
}
