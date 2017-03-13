//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

// UCI Protocol specification:  http://wbec-ridderkerk.nl/html/UCIProtocol.html

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var columnChars = map[string]int{"a": 0, "b": 1, "c": 2, "d": 3, "e": 4, "f": 5, "g": 6, "h": 7}

var fenPieceChars = map[string]int{"p": 0, "n": 1, "b": 2, "r": 3, "q": 4, "k": 5, "P": 8, "N": 9,
	"B": 10, "R": 11, "Q": 12, "K": 13}

var columnNames = [8]string{"a", "b", "c", "d", "e", "f", "g", "h"}

var pieceNames = [2][8]string{{"p", "n", "b", "r", "q", "k"}, {"P", "N", "B", "R", "Q", "K"}}
var sideNames = [2]string{"b", "w"}

type EPD struct {
	brd        *Board
	bestMoves  []string
	avoidMoves []string
	nodeCount  map[int]int
	id         string
	fen        string
}

func (epd *EPD) Print() {
	fmt.Println(epd.id)
}

func (epd *EPD) PrintDetails() {
	epd.Print()
	epd.brd.PrintDetails()
	fmt.Println("Best moves:")
	fmt.Println(epd.bestMoves)
	fmt.Println("Avoid moves:")
	fmt.Println(epd.avoidMoves)
}

func LoadEpdFile(dir string) ([]*EPD, error) {
	epdFile, err := os.Open(dir)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("The specified EPD file could not be loaded.:\n%s\n", dir))
	}
	var testPositions []*EPD
	scanner := bufio.NewScanner(epdFile)
	for scanner.Scan() {
		epd := ParseEPDString(scanner.Text())
		testPositions = append(testPositions, epd)
	}
	return testPositions, err
}

// 2k4B/bpp1qp2/p1b5/7p/1PN1n1p1/2Pr4/P5PP/R3QR1K b - - bm Ng3+ g3; id "WAC.273";
func ParseEPDString(str string) *EPD {
	epd := &EPD{
		nodeCount: make(map[int]int),
	}
	epdFields := strings.Split(str, ";")
	fenFields := strings.Split(epdFields[0], " ")

	epd.brd = ParseFENSlice(fenFields[:4])
	epd.fen = strings.Join(fenFields[:4], " ")

	bm := regexp.MustCompile("bm")
	am := regexp.MustCompile("am")
	id := regexp.MustCompile("id")
	depth := regexp.MustCompile("D[1-9][0-9]?")
	var loc []int
	var subFields []string
	var d, nodeCount int64
	for _, field := range epdFields {
		loc = bm.FindStringIndex(field)
		if loc != nil {
			field = field[loc[1]:]
			subFields = strings.Split(field, " ")
			epd.bestMoves = append(epd.bestMoves, subFields...)
			continue
		}
		loc = am.FindStringIndex(field)
		if loc != nil {
			field = field[:loc[1]+1]
			subFields = strings.Split(field, " ")
			epd.avoidMoves = append(epd.avoidMoves, subFields...)
			continue
		}
		loc = id.FindStringIndex(field)
		if loc != nil {
			subFields = strings.Split(field, " ")
			epd.id = strings.Join(subFields[2:], "")
		}

		loc = depth.FindStringIndex(field)
		if loc != nil { // map each depth to expected node count
			subFields = strings.Split(field, " ")
			d, _ = strconv.ParseInt(subFields[0][1:], 10, 64)
			nodeCount, _ = strconv.ParseInt(subFields[1], 10, 64)
			epd.nodeCount[int(d)] = int(nodeCount)
		}
	}
	return epd
}

var sanChars = [8]string{"P", "N", "B", "R", "Q", "K"}

// <SAN move descriptor piece moves>
// 		::= <Piece symbol>[<from file>|<from rank>|<from square>]['x']<to square>
// <SAN move descriptor pawn captures>
// 		::= <from file>[<from rank>] 'x' <to square>[<promoted to>]
// <SAN move descriptor pawn push>
// 		::= <to square>[<promoted to>]

// ToSan converts Move m to Standard Algebraic Notation (SAN)
func ToSAN(brd *Board, m Move) string {
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
	if popCount(brd.pieces[brd.c][piece]) > 1 {
		occ := brd.AllOccupied()
		c := brd.c
		var t BB
		switch piece {
		case KNIGHT:
			t = knightMasks[to] & brd.pieces[c][piece]
		case BISHOP:
			t = BishopAttacks(occ, to) & brd.pieces[c][piece]
		case ROOK:
			t = RookAttacks(occ, to) & brd.pieces[c][piece]
		case QUEEN:
			t = QueenAttacks(occ, to) & brd.pieces[c][piece]
		}
		if popCount(t) > 1 {
			if popCount(columnMasks[column(from)]&t) == 1 {
				san = columnNames[column(from)] + san
			} else if popCount(rowMasks[row(from)]&t) == 1 {
				san = strconv.Itoa(row(from)+1) + san
			} else {
				san = SquareString(from) + san
			}
		}
	}

	if GivesCheck(brd, m) {
		san += "+"
	}
	san = sanChars[piece] + san
	return san
}

func PawnSAN(brd *Board, m Move, san string) string {
	if m.IsPromotion() {
		san += sanChars[m.PromotedTo()]
	}
	if m.IsCapture() {
		from := m.From()
		san = columnNames[column(from)] + "x" + san
		// disambiguate capturing pawn
		// if brd.TypeAt(to) == EMPTY { // en passant
		// 	san = columnNames[column(from)] + san
		// } else {
		// 	if popCount(pawnAttackMasks[brd.Enemy()][to]&brd.pieces[brd.c][PAWN]) > 1 {
		// 		san = columnNames[column(from)] + san
		// 	}
		// }
	}
	if GivesCheck(brd, m) {
		san += "+"
	}
	return san
}

func GivesCheck(brd *Board, m Move) bool {
	memento := brd.NewMemento()
	makeMove(brd, m)
	inCheck := brd.InCheck()
	unmakeMove(brd, m, memento)
	return inCheck
}

// Fields: placement, side to move, castle rights, en-passant target
// 2k4B/bpp1qp2/p1b5/7p/1PN1n1p1/2Pr4/P5PP/R3QR1K b - -
// Start Pos: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
// white pieces are capitalized.
func BoardToFEN(brd *Board) string {
	fenFields := []string{
		GetFENPlacement(brd),
		sideNames[brd.c],
		GetFENCastleRights(brd),
		SquareString(int(brd.enpTarget)),
		// string(brd.halfmoveClock),
	}
	return strings.Join(fenFields, " ")
}

func GetFENPlacement(brd *Board) string {
	placement := ""
	var row []Piece
	var emptySquares int
	for sq := 0; sq < 64; sq = sq + 8 {
		row = brd.squares[sq:(sq + 8)]
		rowStr := ""
		emptySquares = 0
		for i, pc := range row {
			if pc == EMPTY {
				emptySquares += 1
			} else {
				if emptySquares > 0 {
					rowStr += fmt.Sprint(emptySquares)
					emptySquares = 0
				}
				if sqMaskOn[sq+i]&brd.occupied[BLACK] > 0 {
					rowStr += pieceNames[BLACK][pc]
				} else {
					rowStr += pieceNames[WHITE][pc]
				}
			}
		}
		if emptySquares > 0 {
			rowStr += fmt.Sprint(emptySquares)
		}
		placement = rowStr + "/" + placement
	}
	return strings.TrimRight(placement, "/")
}

func ParseFENSlice(fenFields []string) *Board {
	brd := EmptyBoard()

	ParsePlacement(brd, fenFields[0])
	brd.c = ParseSide(fenFields[1])
	brd.castle = ParseCastleRights(brd, fenFields[2])
	brd.hashKey ^= CastleZobrist(brd.castle)
	if len(fenFields) > 3 {
		brd.enpTarget = ParseEnpTarget(fenFields[3])
		if len(fenFields) > 4 {
			brd.halfmoveClock = ParseHalfmoveClock(fenFields[4])
		}
	}
	brd.hashKey ^= EnpZobrist(brd.enpTarget)
	return brd
}

func ParseFENString(str string) *Board {
	brd := EmptyBoard()
	fenFields := strings.Split(str, " ")

	ParsePlacement(brd, fenFields[0])
	brd.c = ParseSide(fenFields[1])
	brd.castle = ParseCastleRights(brd, fenFields[2])
	brd.hashKey ^= CastleZobrist(brd.castle)

	brd.enpTarget = ParseEnpTarget(fenFields[3])
	brd.hashKey ^= EnpZobrist(brd.enpTarget)

	if len(fenFields) > 4 {
		brd.halfmoveClock = ParseHalfmoveClock(fenFields[4])
	}
	return brd
}

func ParsePlacement(brd *Board, str string) {
	var rowStr string
	rowFields := strings.Split(str, "/")
	sq := 0
	matchDigit, _ := regexp.Compile(`\d`)
	for row := len(rowFields) - 1; row >= 0; row-- {
		rowStr = rowFields[row]
		for _, r := range rowStr {
			chr := string(r)
			if matchDigit.MatchString(chr) {
				digit, _ := strconv.ParseInt(chr, 10, 5)
				sq += int(digit)
			} else {
				c := uint8(fenPieceChars[chr] >> 3)
				pieceType := Piece(fenPieceChars[chr] & 7)
				addPiece(brd, pieceType, sq, c) // place the piece on the board.
				switch pieceType {
				case PAWN:
					brd.pawnHashKey ^= PawnZobrist(sq, c)
				case KING:
					brd.kingSq[c] = uint8(sq)
				default:
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

func GetFENCastleRights(brd *Board) string {
	rights := ""
	if brd.castle&C_WK > 0 {
		rights += "K"
	}
	if brd.castle&C_WQ > 0 {
		rights += "Q"
	}
	if brd.castle&C_BK > 0 {
		rights += "k"
	}
	if brd.castle&C_BQ > 0 {
		rights += "q"
	}
	if rights == "" {
		rights = "-"
	}
	return rights
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
	nonNumeric, _ := regexp.MatchString("\\D", str)
	if nonNumeric {
		return 0
	} else {
		halmoveClock, _ := strconv.ParseInt(str, 10, 8)
		return uint8(halmoveClock)
	}
}

func ParseMove(brd *Board, str string) Move {
	// make sure the move is valid.
	if !IsMove(str) {
		return NO_MOVE
	}

	from := ParseSquare(str[:2])
	to := ParseSquare(str[2:4])
	piece := brd.TypeAt(from)
	capturedPiece := brd.TypeAt(to)
	if piece == PAWN && capturedPiece == EMPTY { // check for en-passant capture
		if Abs(to-from) == 9 || Abs(to-from) == 7 {
			capturedPiece = PAWN // en-passant capture detected.
		}
	}
	var promotedTo Piece
	if len(str) == 5 { // check for promotion.
		promotedTo = Piece(fenPieceChars[string(str[4])]) // will always be lowercase.
	} else {
		promotedTo = Piece(EMPTY)
	}
	return NewMove(from, to, piece, capturedPiece, promotedTo)
}

// create regular expression to match valid move string.
func IsMove(str string) bool {
	match, _ := regexp.MatchString("[a-h][1-8][a-h][1-8][nbrq]?", str)
	return match
}

func Square(row, column int) int { return (row << 3) + column }

func ParseSquare(str string) int {
	column := columnChars[string(str[0])]
	row, _ := strconv.ParseInt(string(str[1]), 10, 5)
	return Square(int(row-1), column)
}

func SquareString(sq int) string {
	if sq >= SQ_INVALID {
		return "-"
	} else {
		return ParseCoordinates(row(sq), column(sq))
	}
}

func ParseCoordinates(row, col int) string {
	return columnNames[col] + strconv.Itoa(row+1)
}
