//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
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

// UCI Protocol specification:  http://wbec-ridderkerk.nl/html/UCIProtocol.html

package main

import (
	// "fmt"
	"regexp"
	"strconv"
	"strings"
)

func ParseEPD(str string) {

}

// 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'

func ParseFENSlice(fen_fields []string) *Board {
	brd := EmptyBoard()

	ParsePlacement(brd, fen_fields[0])
	brd.c = ParseSide(fen_fields[1])
	brd.castle = ParseCastleRights(brd, fen_fields[2])
	brd.enp_target = ParseEnpTarget(fen_fields[3])
	if len(fen_fields) > 4 {
		brd.halfmove_clock = ParseHalfmoveClock(fen_fields[4])
	}
	return brd
}

func ParseFENString(str string) *Board {
	brd := EmptyBoard()
	fen_fields := strings.Split(str, " ")

	ParsePlacement(brd, fen_fields[0])
	brd.c = ParseSide(fen_fields[1])
	brd.castle = ParseCastleRights(brd, fen_fields[2])
	brd.enp_target = ParseEnpTarget(fen_fields[3])
	brd.halfmove_clock = ParseHalfmoveClock(fen_fields[4])

	return brd
}

var fen_piece_chars = map[string]int{
	"p": 0,
	"n": 1,
	"b": 2,
	"r": 3,
	"q": 4,
	"k": 5,
	"P": 8,
	"N": 9,
	"B": 10,
	"R": 11,
	"Q": 12,
	"K": 13,
}

func ParsePlacement(brd *Board, str string) {
	var row_str string
	row_fields := strings.Split(str, "/")
	sq := 0
	match_digit, _ := regexp.Compile("\\d")
	for row := len(row_fields) - 1; row >= 0; row-- {

		row_str = row_fields[row]

		for _, r := range row_str {
			chr := string(r)
			if match_digit.MatchString(chr) {
				digit, _ := strconv.ParseInt(chr, 10, 5)
				sq += int(digit)
			} else {
				c := uint8(fen_piece_chars[chr] >> 3)
				piece_type := Piece(fen_piece_chars[chr] & 7)
				add_piece(brd, piece_type, sq, c) // place the piece on the board.
				sq += 1
			}
		}

	}
}

func ParseSide(str string) uint8 {
	if str == "w" {
		return 1
	} else if str == "b" {
		return 0
	} else {
		// something's wrong.
		return 1
	}
}

func ParseCastleRights(brd *Board, str string) uint8 {
	var castle uint8
	if str != "-" {
		match, _ := regexp.MatchString("K", str)
		if match {
			castle |= C_WK
		}
		match, _ = regexp.MatchString("Q", str)
		if match {
			castle |= C_WQ
		}
		match, _ = regexp.MatchString("k", str)
		if match {
			castle |= C_BK
		}
		match, _ = regexp.MatchString("q", str)
		if match {
			castle |= C_BQ
		}
	}
	return castle
}

func ParseEnpTarget(str string) uint8 {
	if str == "-" {
		return SQ_INVALID
	} else {
		return uint8(ParseSquare(str))
	}
}

func ParseHalfmoveClock(str string) uint8 {
	non_numeric, _ := regexp.MatchString("\\D", str)
	if non_numeric {
		return 0
	} else {
		halmove_clock, _ := strconv.ParseInt(str, 10, 7)
		return uint8(halmove_clock)
	}
}

func ParseMove(brd *Board, str string) Move {
	from := ParseSquare(string(str[:2]))
	to := ParseSquare(string(str[2:4]))
	piece := brd.TypeAt(from)
	captured_piece := brd.TypeAt(to)
	if piece == PAWN && captured_piece == EMPTY { // check for en-passant capture
		if abs(to-from) == 9 || abs(to-from) == 7 {
			captured_piece = PAWN // en-passant capture detected.
		}
	}
	var promoted_to Piece
	if len(str) == 5 { // check for promotion.
		promoted_to = Piece(fen_piece_chars[string(str[4])]) // will always be lowercase.
	} else {
		promoted_to = Piece(EMPTY)
	}
	return NewMove(from, to, piece, captured_piece, promoted_to)
}

// A1 through H8.  test with Regexp.

var column_chars = map[string]int{
	"a": 0,
	"b": 1,
	"c": 2,
	"d": 3,
	"e": 4,
	"f": 5,
	"g": 6,
	"h": 7,
}

var column_names = [8]string{ "a", "b", "c", "d", "e", "f", "g", "h" }


// create regular expression to match valid move string.
func IsMove(str string) bool {
	match, _ := regexp.MatchString("[a-h][1-8][a-h][1-8][nbrq]", str)
	return match
}



func ParseSquare(str string) int {
	column := column_chars[string(str[0])]
	row, _ := strconv.ParseInt(string(str[1]), 10, 5)
	return Square(int(row-1), column)
}

func SquareString(sq int) string {
	return ParseCoordinates(row(sq), column(sq))
}

func ParseCoordinates(row, col int) string {
	return column_names[col] + strconv.FormatInt(int64(row+1), 10)
}








