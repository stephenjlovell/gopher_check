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
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Info
type Info struct {
	score, depth, nodeCount int
	t                       time.Duration // time elapsed
	stk                     Stack
}

// TODO: add proper error handling in UCI adapter.
type UCIAdapter struct {
	brd    *Board
	search *Search
	wg     *sync.WaitGroup
	result chan SearchResult

	move_counter int

	option_ponder bool
	option_debug  bool
}

func NewUCIAdapter() *UCIAdapter {
	return &UCIAdapter{
		wg:     new(sync.WaitGroup),
		result: make(chan SearchResult),
	}
}

func (uci *UCIAdapter) Send(s string) { // log the UCI command s and print to standard I/O.
	log.Printf("engine: " + s)
	fmt.Printf(s)
}

func (uci *UCIAdapter) BestMove(result SearchResult) {
	uci.Send(fmt.Sprintf("bestmove %s ponder %s\n", result.best_move.ToUCI(),
		result.ponder_move.ToUCI()))
}

// Printed to standard output at end of each non-trivial iterative deepening pass.
// Score given in centipawns. Time given in milliseconds. PV given as list of moves.
// Example: info score cp 13  depth 1 nodes 13 time 15 pv f1b5 h1h2
func (uci *UCIAdapter) Info(info Info) {
	nps := int64(float64(info.nodeCount) / info.t.Seconds())
	uci.Send(fmt.Sprintf("info score cp %d depth %d nodes %d nps %d time %d pv %s\n", info.score,
		info.depth, info.nodeCount, nps, int(info.t/time.Millisecond), info.stk[0].pv.ToUCI()))
}

func (uci *UCIAdapter) InfoString(s string) {
	uci.Send("info string " + s)
}

func (uci *UCIAdapter) Read(reader *bufio.Reader) {
	var input string
	var uci_fields []string

	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
	}
	defer f.Close()
	log.SetOutput(f)

	ponder := false

	for {
		input, _ = reader.ReadString('\n')
		log.Println("gui: " + input)
		uci_fields = strings.Fields(input)

		if len(uci_fields) > 0 {
			switch uci_fields[0] {
			case "":
				continue
				// uci
				// 	tell engine to use the uci (universal chess interface),
				// 	this will be send once as a first command after program boot
				// 	to tell the engine to switch to uci mode.
				// 	After receiving the uci command the engine must identify itself with the "id" command
				// 	and sent the "option" commands to tell the GUI which engine settings the engine supports if any.
				// 	After that the engine should sent "uciok" to acknowledge the uci mode.
				// 	If no uciok is sent within a certain time period, the engine task will be killed by the GUI.
			case "uci":
				uci.identify()
				// * debug [ on | off ]
				// 	switch the debug mode of the engine on and off.
				// 	In debug mode the engine should sent additional infos to the GUI, e.g. with the "info string" command,
				// 	to help debugging, e.g. the commands that the engine has received etc.
				// 	This mode should be switched off by default and this command can be sent
				// 	any time, also when the engine is thinking.
			case "debug":
				if len(uci_fields) > 1 {
					uci.option_debug = uci.debug(uci_fields[1:])
				}
				uci.Send("readyok\n")
				// * isready
				// 	this is used to synchronize the engine with the GUI. When the GUI has sent a command or
				// 	multiple commands that can take some time to complete,
				// 	this command can be used to wait for the engine to be ready again or
				// 	to ping the engine to find out if it is still alive.
				// 	E.g. this should be sent after setting the path to the tablebases as this can take some time.
				// 	This command is also required once before the engine is asked to do any search
				// 	to wait for the engine to finish initializing.
				// 	This command must always be answered with "readyok" and can be sent also when the engine is calculating
				// 	in which case the engine should also immediately answer with "readyok" without stopping the search.
			case "isready":
				uci.wg.Wait()
				uci.Send("readyok\n")
				// * setoption name  [value ]
				// 	this is sent to the engine when the user wants to change the internal parameters
				// 	of the engine. For the "button" type no value is needed.
				// 	One string will be sent for each parameter and this will only be sent when the engine is waiting.
				// 	The name of the option in  should not be case sensitive and can inludes spaces like also the value.
				// 	The substrings "value" and "name" should be avoided in  and  to allow unambiguous parsing,
				// 	for example do not use  = "draw value".
				// 	Here are some strings for the example below:
				// 	   "setoption name Nullmove value true\n"
				//       "setoption name Selectivity value 3\n"
				// 	   "setoption name Style value Risky\n"
				// 	   "setoption name Clear Hash\n"
				// 	   "setoption name NalimovPath value c:\chess\tb\4;c:\chess\tb\5\n"
			case "setoption": // setoption name option_name
				if len(uci_fields) > 2 && uci_fields[1] == "name" {
					uci.setOption(uci_fields[2:])
				}
				uci.Send("readyok\n")
				// * register
				// 	this is the command to try to register an engine or to tell the engine that registration
				// 	will be done later. This command should always be sent if the engine	has send "registration error"
				// 	at program startup.
				// 	The following tokens are allowed:
				// 	* later
				// 	   the user doesn't want to register the engine now.
				// 	* name
				// 	   the engine should be registered with the name
				// 	* code
				// 	   the engine should be registered with the code
				// 	Example:
				// 	   "register later"
				// 	   "register name Stefan MK code 4359874324"
				//
			case "register":
				if len(uci_fields) > 1 {
					uci.register(uci_fields[1:])
				}
				uci.Send("readyok\n")
				// * ucinewgame
				//    this is sent to the engine when the next search (started with "position" and "go") will be from
				//    a different game. This can be a new game the engine should play or a new game it should analyse but
				//    also the next position from a testsuite with positions only.
				//    If the GUI hasn't sent a "ucinewgame" before the first "position" command, the engine shouldn't
				//    expect any further ucinewgame commands as the GUI is probably not supporting the ucinewgame command.
				//    So the engine should not rely on this command even though all new GUIs should support it.
				//    As the engine's reaction to "ucinewgame" can take some time the GUI should always send "isready"
				//    after "ucinewgame" to wait for the engine to finish its operation.
			case "ucinewgame":
				reset_main_tt()
				uci.brd = StartPos()
				uci.Send("readyok\n")
				// * position [fen  | startpos ]  moves  ....
				// 	set up the position described in fenstring on the internal board and
				// 	play the moves on the internal chess board.
				// 	if the game was played  from the start position the string "startpos" will be sent
				// 	Note: no "new" command is needed. However, if this position is from a different game than
				// 	the last position sent to the engine, the GUI should have sent a "ucinewgame" inbetween.
			case "position":
				uci.wg.Wait()
				uci.position(uci_fields[1:])
				uci.Send("readyok\n")
				// * go
				// 	start calculating on the current position set up with the "position" command.
				// 	There are a number of commands that can follow this command, all will be sent in the same string.
				// 	If one command is not send its value should be interpreted as it would not influence the search.
			case "go":
				ponder = uci.start(uci_fields[1:]) // parse any parameters given by GUI and begin searching.
				if !uci.option_ponder || !ponder {
					uci.move_counter++
				}
				// * stop
				// 	stop calculating as soon as possible,
				// 	don't forget the "bestmove" and possibly the "ponder" token when finishing the search
			case "stop": // stop calculating and return a result as soon as possible.
				if uci.search != nil {
					uci.search.Abort()
					if ponder {
						uci.BestMove(<-uci.result)
					}
				}
				// * ponderhit
				// 	the user has played the expected move. This will be sent if the engine was told to ponder on the same move
				// 	the user has played. The engine should continue searching but switch from pondering to normal search.
			case "ponderhit":
				if uci.search != nil && ponder {
					uci.search.gt.Start()
					uci.BestMove(<-uci.result)
				}
				uci.move_counter++
			case "quit": // quit the program as soon as possible
				return

			case "print": // Not a UCI command. Used to print the board for debugging from console
				uci.brd.Print() // while in UCI mode.
			default:
				uci.invalid(uci_fields)
			}
		}
	}
}

func (uci *UCIAdapter) debug(uci_fields []string) bool {
	switch uci_fields[0] {
	case "on":
		return true
	case "off":
	default:
		uci.invalid(uci_fields)
	}
	return false
}

func (uci *UCIAdapter) invalid(uci_fields []string) {
	uci.InfoString("invalid command.\n")
}

func (uci *UCIAdapter) identify() {
	uci.Send(fmt.Sprintf("id name GopherCheck %s\n", version))
	uci.Send("id author Steve Lovell\n")
	uci.option()
	uci.Send("uciok\n")
}

func (uci *UCIAdapter) option() { // option name option_name [ parameters ]
	// tells the GUI which parameters can be changed in the engine.
	uci.Send("option name Ponder type check default false\n")

}

// some example options from Toga 1.3.1:

// Engine: option name Hash type spin default 16 min 4 max 1024
// Engine: option name Search Time type spin default 0 min 0 max 3600
// Engine: option name Search Depth type spin default 0 min 0 max 20
// Engine: option name Ponder type check default false
// Engine: option name OwnBook type check default true
// Engine: option name BookFile type string default performance.bin
// Engine: option name MultiPV type spin default 1 min 1 max 10
// Engine: option name NullMove Pruning type combo default Always var Always var Fail High var Never
// Engine: option name NullMove Reduction type spin default 3 min 1 max 4
// Engine: option name Verification Search type combo default Always var Always var Endgame var Never
// Engine: option name Verification Reduction type spin default 5 min 1 max 6
// Engine: option name History Pruning type check default true
// Engine: option name History Threshold type spin default 70 min 0 max 100
// Engine: option name Futility Pruning type check default true
// Engine: option name Futility Margin type spin default 100 min 0 max 500
// Engine: option name Extended Futility Margin type spin default 300 min 0 max 900
// Engine: option name Delta Pruning type check default true
// Engine: option name Delta Margin type spin default 50 min 0 max 500
// Engine: option name Quiescence Check Plies type spin default 1 min 0 max 2
// Engine: option name Material type spin default 100 min 0 max 400
// Engine: option name Piece Activity type spin default 100 min 0 max 400
// Engine: option name King Safety type spin default 100 min 0 max 400
// Engine: option name Pawn Structure type spin default 100 min 0 max 400
// Engine: option name Passed Pawns type spin default 100 min 0 max 400
// Engine: option name Toga Lazy Eval type check default true
// Engine: option name Toga Lazy Eval Margin type spin default 200 min 0 max 900
// Engine: option name Toga King Safety type check default false
// Engine: option name Toga King Safety Margin type spin default 1700 min 500 max 3000
// Engine: option name Toga Extended History Pruning type check default false

func (uci *UCIAdapter) setOption(uci_fields []string) {
	switch uci_fields[0] {
	case "Ponder": // example: setoption name Ponder value true
		if len(uci_fields) == 3 {
			switch uci_fields[2] {
			case "true":
				uci.option_ponder = true
			case "false":
				uci.option_ponder = false
			default:
				uci.invalid(uci_fields)
			}
		}
	default:
	}
}

func (uci *UCIAdapter) register(uci_fields []string) {
	// The following tokens are allowed:
	// * later - the user doesn't want to register the engine now.
	// * name - the engine should be registered with the name
	// * code - the engine should be registered with the code
	// Examples: "register later"  "register name Stefan MK code 4359874324"
}

// * go
// 	start calculating on the current position set up with the "position" command.
// 	There are a number of commands that can follow this command, all will be sent in the same string.
// 	If one command is not send its value should be interpreted as it would not influence the search.
func (uci *UCIAdapter) start(uci_fields []string) bool {
	var time_limit int
	max_depth := MAX_DEPTH
	gt := NewGameTimer(uci.move_counter, uci.brd.c) // TODO: this will be inaccurate in pondering mode.
	ponder := false
	var allowed_moves []Move
	for len(uci_fields) > 0 {
		// fmt.Println(uci_fields[0])
		switch uci_fields[0] {

		// 	* searchmoves  ....
		// 		restrict search to this moves only
		// 		Example: After "position startpos" and "go infinite searchmoves e2e4 d2d4"
		// 		the engine should only search the two moves e2e4 and d2d4 in the initial position.
		case "searchmoves":
			uci_fields = uci_fields[1:]
			for len(uci_fields) > 0 && IsMove(uci_fields[0]) {
				allowed_moves = append(allowed_moves, ParseMove(uci.brd, uci_fields[0]))
				uci_fields = uci_fields[1:]
			}

		// 	* ponder - start searching in pondering mode.
		case "ponder":
			if uci.option_ponder {
				ponder = true
			}
			uci_fields = uci_fields[1:]

		case "wtime": // white has x msec left on the clock
			time_limit, _ = strconv.Atoi(uci_fields[1])
			gt.remaining[WHITE] = time.Duration(time_limit) * time.Millisecond
			uci_fields = uci_fields[2:]

		case "btime": // black has x msec left on the clock
			time_limit, _ = strconv.Atoi(uci_fields[1])
			gt.remaining[BLACK] = time.Duration(time_limit) * time.Millisecond
			uci_fields = uci_fields[2:]

		case "winc": //	white increment per move in mseconds if x > 0
			time_limit, _ = strconv.Atoi(uci_fields[1])
			gt.inc[WHITE] = time.Duration(time_limit) * time.Millisecond
			uci_fields = uci_fields[2:]

		case "binc": //	black increment per move in mseconds if x > 0
			time_limit, _ = strconv.Atoi(uci_fields[1])
			gt.inc[BLACK] = time.Duration(time_limit) * time.Millisecond
			uci_fields = uci_fields[2:]

		// 	* movestogo: there are x moves to the next time control, this will only be sent if x > 0,
		// 		if you don't get this and get the wtime and btime it's sudden death
		case "movestogo":
			remaining, _ := strconv.Atoi(uci_fields[1])
			gt.moves_remaining = remaining
			uci_fields = uci_fields[2:]

		case "depth": // search x plies only
			max_depth, _ = strconv.Atoi(uci_fields[1])
			uci_fields = uci_fields[2:]

		case "nodes": // search x nodes only
			uci.invalid(uci_fields)
			uci_fields = uci_fields[2:]

		case "mate": // search for a mate in x moves
			uci.invalid(uci_fields)
			uci_fields = uci_fields[2:]

		case "movetime": // search exactly x mseconds
			time_limit, _ = strconv.Atoi(uci_fields[1])
			gt.SetMoveTime(time.Duration(time_limit) * time.Millisecond)
			uci_fields = uci_fields[2:]
		// * infinite: search until the "stop" command. Do not exit the search without being
		// told so in this mode!
		case "infinite":
			gt.SetMoveTime(MAX_TIME)
			uci_fields = uci_fields[1:]
		default:
			uci_fields = uci_fields[1:]
		}
	}
	uci.wg.Add(1)

	// type SearchParams struct {
	// 	max_depth         int
	// 	verbose, ponder, restrict_search bool
	// }
	uci.search = NewSearch(SearchParams{max_depth, uci.option_debug, ponder, len(allowed_moves) > 0},
		gt, uci, allowed_moves)
	go uci.search.Start(uci.brd.Copy()) // starting the search also starts the clock
	return ponder
}

// position [fen  | startpos ]  moves  ....
func (uci *UCIAdapter) position(uci_fields []string) {
	if len(uci_fields) == 0 {
		uci.brd = StartPos()
	} else if uci_fields[0] == "startpos" {
		uci.brd = StartPos()
		uci_fields = uci_fields[1:]
		if len(uci_fields) > 1 && uci_fields[0] == "moves" {
			uci.playMoveSequence(uci_fields[1:])
		}
	} else if uci_fields[0] == "fen" {
		uci.brd = ParseFENSlice(uci_fields[1:])
		if len(uci_fields) > 7 {
			uci.playMoveSequence(uci_fields[7:])
		}
	} else {
		uci.invalid(uci_fields)
	}
}

func (uci *UCIAdapter) playMoveSequence(uci_fields []string) {
	var move Move
	if uci_fields[0] == "moves" {
		uci_fields = uci_fields[1:]
	}
	for _, move_str := range uci_fields {
		move = ParseMove(uci.brd, move_str)
		make_move(uci.brd, move)
	}
}

func StartPos() *Board {
	return ParseFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}
