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
  "bufio"
  "os"
	"fmt"
	"strings"
  "strconv"
	"time"
)

// var ponder_mode bool = false

func Milliseconds(d time.Duration) int64 {
	return int64(d.Seconds() * float64(time.Second/time.Millisecond))
}

// Printed to standard output at end of each iterative deepening pass. Score given in centipawns.
// Time given in milliseconds. PV given as list of moves.
// Example: info score cp 13  depth 1 nodes 13 time 15 pv f1b5
func PrintInfo(score, depth, node_count int, time_elapsed time.Duration) {
	fmt.Printf("info score cp %d depth %d nodes %d time %d\n", score, depth, node_count, Milliseconds(time_elapsed))
}

func ReadUCICommand() {
  var input string
  reader := bufio.NewReader(os.Stdin)
	UCIStart()
	for {
    input, _ = reader.ReadString('\n')
		uci_fields := strings.Fields(input)

		switch uci_fields[0] {
		case "":
			continue
		case "uci":
			UCIStart()
		case "isready":
			// to do: check if any tasks are still running.
			fmt.Printf("readyok\n")

		case "position":
			current_board = ParseUCIPosition(uci_fields[1:])
      fmt.Printf("readyok\n")

		case "ucinewgame":
			ResetAll() // reset all shared data structures and prepare to start a new game.
			current_board = ParseFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
      fmt.Printf("readyok\n")

		case "ponderhit":

		case "go":
			ParseUCIGo(uci_fields[1:])

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

func UCIStart() {
	fmt.Println("id name GopherCheck")
	fmt.Println("id author Steve Lovell")
	fmt.Println("uciok")
}

func ParseUCIGo(uci_fields []string) {
  depth, time := int64(MAX_DEPTH), MAX_TIME
  var restrict_search []Move

  for len(uci_fields) > 0 {
    switch uci_fields[0] {
    case "depth":
      depth, _ = strconv.ParseInt(uci_fields[1], 10, 8)
      uci_fields = uci_fields[2:]
    case "ponder":
      depth = 32
    case "searchmoves":
      uci_fields = uci_fields[:1]
      for len(uci_fields) > 0 && IsMove(uci_fields[0]) {
        restrict_search = append(restrict_search, ParseMove(current_board, uci_fields[0]))
        uci_fields = uci_fields[:1]
      }
    default:
      uci_fields = uci_fields[:1]
    }
  }

  go func() {
    move := Search(current_board, restrict_search, int(depth), time)
    fmt.Printf("bestmove %s\n", move.ToString())
  }()
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

}
