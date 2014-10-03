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

// This module implements communication over standard I/O using the Universal Chess
// Interface (UCI) protocol.  This allows the engine to communicate with any other
// chess software that also implements UCI.

// UCI Protocol specification:  http://wbec-ridderkerk.nl/html/UCIProtocol.html

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// var ponder_mode bool = false

func Milliseconds(d time.Duration) int64 {
	return int64(d.Seconds() * float64(time.Second/time.Millisecond))
}

// Printed to standard output at end of each iterative deepening pass. Score given in centipawns.
// Time given in milliseconds. PV given as list of moves.
// Example: info score cp 13  depth 1 nodes 13 time 15 pv f1b5
func PrintInfo(score, depth, node_count int, time_elapsed time.Duration, pv *PV) {
	ms := Milliseconds(time_elapsed)
	nps := int64(float64(node_count) / (float64(ms) / float64(1000.0)))
	fmt.Printf("info score cp %d depth %d nodes %d nps %d time %d", score, depth, node_count, nps, ms)
	if pv != nil {
		fmt.Printf(" pv %s\n", pv.ToUCI())
	}
	fmt.Printf("NPS: %.4f m\n", float64(nps)/1000000)
}

func ReadUCICommand() {
	var input string
	var wg sync.WaitGroup

	reader := bufio.NewReader(os.Stdin)
	UCIIdentify()
	for {
		input, _ = reader.ReadString('\n')
		uci_fields := strings.Fields(input)

		switch uci_fields[0] {
		case "":
			continue
		case "uci":
			uci_mode = true
			UCIIdentify()
		case "isready":
			wg.Wait()
			fmt.Printf("readyok\n")
		case "position":
			wg.Wait()
			current_board = ParseUCIPosition(uci_fields[1:])
			fmt.Printf("readyok\n")
		case "ucinewgame":
			ResetAll() // reset all shared data structures and prepare to start a new game.
			current_board = ParseFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
			fmt.Printf("readyok\n")

		case "ponderhit":

		case "go":
			wg.Add(1)
			go ParseUCIGo(uci_fields[1:], &wg) // parse any parameters given by GUI and begin searching.

		case "print":
			current_board.Print()

		case "stop":
			AbortSearch()
			// stop calculating and return a result as soon as possible.

		case "quit":
			AbortSearch()
			return

		default:
			fmt.Println("invalid command.")
		}
	}
}

func UCIIdentify() {
	fmt.Println("id name GopherCheck")
	fmt.Println("id author Steve Lovell")
	fmt.Println("uciok")
}

func ParseUCIGo(uci_fields []string, wg *sync.WaitGroup) {
	depth, time := int64(MAX_DEPTH), int64(MAX_TIME)
	var restrict_search []Move

	for len(uci_fields) > 0 {
		switch uci_fields[0] {
		case "depth":
			depth, _ = strconv.ParseInt(uci_fields[1], 10, 8)
			uci_fields = uci_fields[2:]
		case "ponder":
			depth = 32
			uci_ponder = true
		case "searchmoves":
			uci_fields = uci_fields[:1]
			for len(uci_fields) > 0 && IsMove(uci_fields[0]) {
				restrict_search = append(restrict_search, ParseMove(current_board, uci_fields[0]))
				uci_fields = uci_fields[:1]
			}
		case "movetime":
			time, _ = strconv.ParseInt(uci_fields[1], 10, 24)
			uci_fields = uci_fields[2:]
		case "infinite":
			depth = 32
		default:
			uci_fields = uci_fields[:1]
		}
	}
	move, _ := Search(current_board, restrict_search, int(depth), int(time))
	fmt.Printf("bestmove %s\n", move.ToUCI())
	wg.Done()
}

// position [fen  | startpos ]  moves  ....
func ParseUCIPosition(uci_fields []string) *Board {
	var brd *Board
	if len(uci_fields) == 0 || uci_fields[0] == "startpos" {
		brd = ParseFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		if len(uci_fields) > 2 {
			PlayMoveSequence(brd, uci_fields[1:])
		}
	} else if uci_fields[0] == "fen" {
		brd = ParseFENSlice(uci_fields[1:])
		if len(uci_fields) > 6 {
			PlayMoveSequence(brd, uci_fields[5:])
		}
	} else {
		fmt.Println("Empty board created.")
		brd = EmptyBoard()
	}
	return brd
}

func PlayMoveSequence(brd *Board, uci_fields []string) {
	var move Move
	if uci_fields[0] == "moves" {
		for _, move_str := range uci_fields[1:] {
			move = ParseMove(brd, move_str)
			make_move(brd, move)
		}
	} else if uci_fields[1] == "moves" {
		for _, move_str := range uci_fields[2:] {
			move = ParseMove(brd, move_str)
			make_move(brd, move)
		}
	}
}

func StartPos() *Board {
	return ParseFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}

func ResetAll() {
	main_htable.Clear()
	main_ktable.Clear()
}
