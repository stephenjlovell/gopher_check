//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------
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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"regexp"
)

type EPD struct {
	brd         *Board
	best_moves  []string
	avoid_moves []string
	id          string
}

func (epd *EPD) Print() {
	fmt.Println(epd.id)
	// epd.brd.PrintDetails()
	// fmt.Println("Best moves:")
	// fmt.Println(epd.best_moves)
	// fmt.Println("Avoid moves:")
	// fmt.Println(epd.avoid_moves)
}

func load_epd_file(dir string) []*EPD {
	epd_file, err := os.Open(dir)
	if err != nil {
		panic(err)
	}
	var test_positions []*EPD
	scanner := bufio.NewScanner(epd_file)
	for scanner.Scan() {
		epd := ParseEPDString(scanner.Text())
		test_positions = append(test_positions, epd)
	}
	return test_positions
}

// 2k4B/bpp1qp2/p1b5/7p/1PN1n1p1/2Pr4/P5PP/R3QR1K b - - bm Ng3+ g3; id "WAC.273";
func ParseEPDString(str string) *EPD {
	epd := &EPD{}
	epd_fields := strings.Split(str, ";")
	fen_fields := strings.Split(epd_fields[0], " ")

	epd.brd = ParseFENSlice(fen_fields[:4])

	bm := regexp.MustCompile("bm")
	am := regexp.MustCompile("am")
	id := regexp.MustCompile("id")
	var loc []int
	var move_fields, id_fields []string
	for _, field := range epd_fields {
		loc = bm.FindStringIndex(field)
		if loc != nil {
			field = field[loc[1]:]
			move_fields = strings.Split(field, " ")
			for _, move_field := range move_fields {
				epd.best_moves = append(epd.best_moves, move_field)
			}
			continue
		}
		loc = am.FindStringIndex(field)
		if loc != nil {
			field = field[:loc[1]+1]
			move_fields = strings.Split(field, " ")
			for _, move_field := range move_fields {
				epd.avoid_moves = append(epd.avoid_moves, move_field)
			}
			continue
		}
		loc = id.FindStringIndex(field)
		if loc != nil {
			id_fields = strings.Split(field, " ")
			epd.id = strings.Join(id_fields[2:], "")
		}
	}
	return epd
}

var san_chars = [6]string{"P", "N", "B", "R", "Q", "K"}

func ToSAN(brd *Board, m Move) string { // convert move to Standard Algebraic Notation (SAN)
	piece, from, to := m.Piece(), m.From(), m.To()
	san := SquareString(to)

	if piece == PAWN {
		return PawnSAN(brd, m, san)
	}

	if piece == KING {
		if to-from == 2 { // kingside castling
			return "O-O"
		} else if to-from == -2 { // queenside castling
			return "O-O-O"
		}
	}

	if m.IsCapture() {
		san = "x" + san
	}

	// disambiguate moving piece
	if pop_count(brd.pieces[brd.c][piece]) > 1 {
		occ := brd.AllOccupied()
		c := brd.c
		var t BB
		switch piece {
		case KNIGHT:
			t = knight_masks[to] & brd.pieces[c][piece]
		case BISHOP:
			t = bishop_attacks(occ, to) & brd.pieces[c][piece]
		case ROOK:
			t = rook_attacks(occ, to) & brd.pieces[c][piece]
		case QUEEN:
			t = queen_attacks(occ, to) & brd.pieces[c][piece]
		}
		if pop_count(t) > 1 {
			if pop_count(column_masks[column(from)]&t) == 1 {
				san = column_names[column(from)] + san
			} else if pop_count(row_masks[row(from)]&t) == 1 {
				san = strconv.Itoa(row(from)+1) + san
			} else {
				san = SquareString(from) + san
			}
		}
	}

	if GivesCheck(brd, m) {
		san += "+"
	}
	san = san_chars[piece] + san
	return san
}

func PawnSAN(brd *Board, m Move, san string) string {
	if m.IsPromotion() {
		san += san_chars[m.PromotedTo()]
	}
	if m.IsCapture() {
		from, to := m.From(), m.To()
		san = "x" + san
		// disambiguate capturing pawn
		if brd.TypeAt(to) == EMPTY { // en passant
			san = column_names[column(from)] + san
		} else {
			t := pawn_attack_masks[brd.Enemy()][to] & brd.pieces[brd.c][PAWN]
			if pop_count(t) > 1 {
				san = column_names[column(from)] + san
			}
		}
	}
	if GivesCheck(brd, m) {
		san += "+"
	}

	return san
}

func GivesCheck(brd *Board, m Move) bool {
	memento := brd.NewMemento()
	make_move(brd, m)
	in_check := brd.InCheck()
	unmake_move(brd, m, memento)
	return in_check
}

func ParseFENSlice(fen_fields []string) *Board {
	brd := EmptyBoard()

	ParsePlacement(brd, fen_fields[0])
	brd.c = ParseSide(fen_fields[1])
	brd.castle = ParseCastleRights(brd, fen_fields[2])
	brd.hash_key ^= castle_zobrist(brd.castle)
	if len(fen_fields) > 3 {
		brd.enp_target = ParseEnpTarget(fen_fields[3])
		if len(fen_fields) > 4 {
			brd.halfmove_clock = ParseHalfmoveClock(fen_fields[4])
		}
	}
	brd.hash_key ^= enp_zobrist(brd.enp_target)
	return brd
}

func ParseFENString(str string) *Board {
	brd := EmptyBoard()
	fen_fields := strings.Split(str, " ")

	ParsePlacement(brd, fen_fields[0])
	brd.c = ParseSide(fen_fields[1])
	brd.castle = ParseCastleRights(brd, fen_fields[2])
	brd.hash_key ^= castle_zobrist(brd.castle)

	brd.enp_target = ParseEnpTarget(fen_fields[3])
	brd.hash_key ^= enp_zobrist(brd.enp_target)

	if len(fen_fields) > 4 {
		brd.halfmove_clock = ParseHalfmoveClock(fen_fields[4])
	}
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
				if piece_type == PAWN {
					brd.pawn_hash_key ^= pawn_zobrist(sq, c)
				}
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
		halmove_clock, _ := strconv.ParseInt(str, 10, 8)
		return uint8(halmove_clock)
	}
}

func ParseMove(brd *Board, str string) Move {
	// make sure the move is valid.
	if !IsMove(str) {
		return NO_MOVE
	}

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

var column_names = [8]string{"a", "b", "c", "d", "e", "f", "g", "h"}

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
