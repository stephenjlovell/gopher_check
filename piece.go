//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

const ( // type
	PAWN = iota
	KNIGHT
	BISHOP
	ROOK
	QUEEN
	KING
	EMPTY // no piece located at this square
)

const (
	PAWN_VALUE   = 100 // piece values are given in centipawns
	KNIGHT_VALUE = 320
	BISHOP_VALUE = 333
	ROOK_VALUE   = 510
	QUEEN_VALUE  = 880
	KING_VALUE   = 5000
)

type Piece uint8

var pieceValues = [8]int{PAWN_VALUE, KNIGHT_VALUE, BISHOP_VALUE, ROOK_VALUE, QUEEN_VALUE, KING_VALUE} // default piece values

var promoteValues = [8]int{0, KNIGHT_VALUE - PAWN_VALUE, BISHOP_VALUE - PAWN_VALUE, ROOK_VALUE - PAWN_VALUE,
	QUEEN_VALUE - PAWN_VALUE}

func (pc Piece) Value() int { return pieceValues[pc] }

func (pc Piece) PromoteValue() int { return promoteValues[pc] }
