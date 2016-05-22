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

var uci_mode, uci_ponder, uci_debug bool

var current_board *Board = EmptyBoard()

func Milliseconds(d time.Duration) int64 {
	return int64(d.Seconds() * float64(time.Second/time.Millisecond))
}

// * info
// 	the engine wants to send infos to the GUI. This should be done whenever one of the info has changed.
// 	The engine can send only selected infos and multiple infos can be send with one info command,
// 	e.g. "info currmove e2e4 currmovenumber 1" or
// 	     "info depth 12 nodes 123456 nps 100000".
// 	Also all infos belonging to the pv should be sent together
// 	e.g. "info depth 2 score cp 214 time 1242 nodes 2124 nps 34928 pv e2e4 e7e5 g1f3"
// 	I suggest to start sending "currmove", "currmovenumber", "currline" and "refutation" only after one second
// 	to avoid too much traffic.
// 	Additional info:
// 	* depth
// 		search depth in plies
// 	* seldepth
// 		selective search depth in plies,
// 		if the engine sends seldepth there must also a "depth" be present in the same string.
// 	* time
// 		the time searched in ms, this should be sent together with the pv.
// 	* nodes
// 		x nodes searched, the engine should send this info regularly
// 	* pv  ...
// 		the best line found
// 	* multipv
// 		this for the multi pv mode.
// 		for the best move/pv add "multipv 1" in the string when you send the pv.
// 		in k-best mode always send all k variants in k strings together.
// 	* score
// 		* cp
// 			the score from the engine's point of view in centipawns.
// 		* mate
// 			mate in y moves, not plies.
// 			If the engine is getting mated use negativ values for y.
// 		* lowerbound
// 	      the score is just a lower bound.
// 		* upperbound
// 		   the score is just an upper bound.
// 	* currmove
// 		currently searching this move
// 	* currmovenumber
// 		currently searching move number x, for the first move x should be 1 not 0.
// 	* hashfull
// 		the hash is x permill full, the engine should send this info regularly
// 	* nps
// 		x nodes per second searched, the engine should send this info regularly
// 	* tbhits
// 		x positions where found in the endgame table bases
// 	* cpuload
// 		the cpu usage of the engine is x permill.
// 	* string
// 		any string str which will be displayed be the engine,
// 		if there is a string command the rest of the line will be interpreted as .
// 	* refutation   ...
// 	   move  is refuted by the line  ... , i can be any number >= 1.
// 	   Example: after move d1h5 is searched, the engine can send
// 	   "info refutation d1h5 g6h5"
// 	   if g6h5 is the best answer after d1h5 or if g6h5 refutes the move d1h5.
// 	   if there is norefutation for d1h5 found, the engine should just send
// 	   "info refutation d1h5"
// 		The engine should only send this if the option "UCI_ShowRefutations" is set to true.
// 	* currline   ...
// 	   this is the current line the engine is calculating.  is the number of the cpu if
// 	   the engine is running on more than one cpu.  = 1,2,3....
// 	   if the engine is just using one cpu,  can be omitted.
// 	   If  is greater than 1, always send all k lines in k strings together.
// 		The engine should only send this if the option "UCI_ShowCurrLine" is set to true.

type Info struct {
	score, depth, node_count int
	t time.Duration // time elapsed
	stk Stack
}

func UCISend(s string) { // log the UCI command s and print to standard I/O.
	log.Print(s)
	fmt.Printf(s)
}

// Printed to standard output at end of each non-trivial iterative deepening pass.
// Score given in centipawns. Time given in milliseconds. PV given as list of moves.
// Example: info score cp 13  depth 1 nodes 13 time 15 pv f1b5 h1h2
func UCIInfo(info Info) {
	nps := int64(float64(info.node_count) / info.t.Seconds())
	UCISend(fmt.Sprintf("info score cp %d depth %d nodes %d nps %d time %d pv %s\n", info.score, info.depth,
		info.node_count, nps, int(info.t / time.Millisecond), info.stk[0].pv.ToUCI()))
}

func ReadUCICommand() {
	var input string
	var wg sync.WaitGroup
	var move_counter int

	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
	}
	defer f.Close()
	log.SetOutput(f)

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ = reader.ReadString('\n')
		log.Print(input)
		uci_fields := strings.Fields(input)
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
				uci_mode = true
				UCIIdentify()
	// * debug [ on | off ]
	// 	switch the debug mode of the engine on and off.
	// 	In debug mode the engine should sent additional infos to the GUI, e.g. with the "info string" command,
	// 	to help debugging, e.g. the commands that the engine has received etc.
	// 	This mode should be switched off by default and this command can be sent
	// 	any time, also when the engine is thinking.
			case "debug":
				if len(uci_fields) > 1 {
					UCIDebug(uci_fields[1:])
				}
				UCISend(fmt.Sprintf("readyok\n"))
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
				wg.Wait()
				UCISend(fmt.Sprintf("readyok\n"))
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
					UCISetOption(uci_fields[2:])
				}
				UCISend(fmt.Sprintf("readyok\n"))
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
					UCIRegister(uci_fields[1:])
				}
				UCISend(fmt.Sprintf("readyok\n"))
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
				current_board = ParseFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
				UCISend(fmt.Sprintf("readyok\n"))
	// * position [fen  | startpos ]  moves  ....
	// 	set up the position described in fenstring on the internal board and
	// 	play the moves on the internal chess board.
	// 	if the game was played  from the start position the string "startpos" will be sent
	// 	Note: no "new" command is needed. However, if this position is from a different game than
	// 	the last position sent to the engine, the GUI should have sent a "ucinewgame" inbetween.
			case "position":
				wg.Wait()
				current_board = UCIPosition(uci_fields[1:])
				fmt.Sprintf("readyok\n")
	// * go
	// 	start calculating on the current position set up with the "position" command.
	// 	There are a number of commands that can follow this command, all will be sent in the same string.
	// 	If one command is not send its value should be interpreted as it would not influence the search.
			case "go":
				wg.Add(1)
				go UCIGo(uci_fields[1:], &wg, move_counter) // parse any parameters given by GUI and begin searching.
				move_counter += 1
	// * stop
	// 	stop calculating as soon as possible,
	// 	don't forget the "bestmove" and possibly the "ponder" token when finishing the search
			case "stop": // stop calculating and return a result as soon as possible.
				AbortSearch()
	// * ponderhit
	// 	the user has played the expected move. This will be sent if the engine was told to ponder on the same move
	// 	the user has played. The engine should continue searching but switch from pondering to normal search.
			case "ponderhit":
				UCIInvalid(uci_fields) // placeholder until pondering is implemented.
	// * quit
	// 	quit the program as soon as possible
			case "quit":
				AbortSearch()
				return
			case "print": // Not a UCI command. Used to print the board for debugging while in UCI mode.
				current_board.Print()
			default:
				UCIInvalid(uci_fields)
			}
		}
	}
}

func UCIDebug(uci_fields []string) {
	switch uci_fields[0] {
	case "on":
		uci_debug = true
	case "off":
		uci_debug = false
	default:
		UCIInvalid(uci_fields)
	}
}

func UCIInvalid(uci_fields []string) {
	UCISend(fmt.Sprintf("invalid command.\n"))
}

func UCIIdentify() {
	UCISend(fmt.Sprintf("id name GopherCheck\n"))
	UCISend(fmt.Sprintf("id author Steve Lovell\n"))
	UCISend(fmt.Sprintf("uciok\n"))
	UCIOption()
}

func UCIOption() { // option name option_name [ parameters ]
	// tells the GUI which parameters can be changed in the engine.
	UCISend(fmt.Sprintf("option\n"))
}

func UCISetOption(uci_fields []string) {
	// sent to the engine when GUI user wants to change the internal parameters
	// of the engine
}

func UCIRegister(uci_fields []string) {
	// The following tokens are allowed:
	// * later
	//    the user doesn't want to register the engine now.
	// * name
	//    the engine should be registered with the name
	// * code
	//    the engine should be registered with the code
	// Example:
	//    "register later"
	//    "register name Stefan MK code 4359874324"

}

// * go
// 	start calculating on the current position set up with the "position" command.
// 	There are a number of commands that can follow this command, all will be sent in the same string.
// 	If one command is not send its value should be interpreted as it would not influence the search.
func UCIGo(uci_fields []string, wg *sync.WaitGroup, move_counter int) {
	var time_limit int
	per_move := false

	var restrict_search []Move
	gt := NewGameTimer(move_counter)
	for len(uci_fields) > 0 {
		switch uci_fields[0] {
		// 	* searchmoves  ....
		// 		restrict search to this moves only
		// 		Example: After "position startpos" and "go infinite searchmoves e2e4 d2d4"
		// 		the engine should only search the two moves e2e4 and d2d4 in the initial position.
		case "searchmoves":
			uci_fields = uci_fields[:1]
			for len(uci_fields) > 0 && IsMove(uci_fields[0]) {
				restrict_search = append(restrict_search, ParseMove(current_board, uci_fields[0]))
				uci_fields = uci_fields[:1]
			}
			per_move = true
		// 	* ponder
		// 		start searching in pondering mode.
		// 		Do not exit the search in ponder mode, even if it's mate!
		// 		This means that the last move sent in in the position string is the ponder move.
		// 		The engine can do what it wants to do, but after a "ponderhit" command
		// 		it should execute the suggested move to ponder on. This means that the ponder move sent by
		// 		the GUI can be interpreted as a recommendation about which move to ponder. However, if the
		// 		engine decides to ponder on a different move, it should not display any mainlines as they are
		// 		likely to be misinterpreted by the GUI because the GUI expects the engine to ponder
		// 	   on the suggested move.
		case "ponder":
			uci_ponder = true  // TODO: actually implement pondering.
			per_move = true

		// 		white has x msec left on the clock
		case "wtime":
			time_limit, _ = strconv.Atoi(uci_fields[1])
			gt.remaining[WHITE] = time.Duration(time_limit) * time.Millisecond
			uci_fields = uci_fields[2:]

		// black has x msec left on the clock
		case "btime":
			time_limit, _ = strconv.Atoi(uci_fields[1])
			gt.remaining[BLACK] = time.Duration(time_limit) * time.Millisecond
			uci_fields = uci_fields[2:]

		// 		white increment per move in mseconds if x > 0
		case "winc":
			time_limit, _ = strconv.Atoi(uci_fields[1])
			gt.inc[WHITE] = time.Duration(time_limit) * time.Millisecond
			uci_fields = uci_fields[2:]

		// 		black increment per move in mseconds if x > 0
		case "binc":
			time_limit, _ = strconv.Atoi(uci_fields[1])
			gt.inc[BLACK] = time.Duration(time_limit) * time.Millisecond
			uci_fields = uci_fields[2:]

		// 	* movestogo: there are x moves to the next time control, this will only be sent if x > 0,
		// 		if you don't get this and get the wtime and btime it's sudden death
		case "movestogo":
			gt.moves_remaining, _ = strconv.Atoi(uci_fields[1])
			uci_fields = uci_fields[2:]

		// search x plies only.
		case "depth":
			gt.max_depth, _ = strconv.Atoi(uci_fields[1])
			uci_fields = uci_fields[2:]
			per_move = true
		// search x nodes only,
		case "nodes":
			uci_fields = uci_fields[2:]
			per_move = true
		// search for a mate in x moves
		case "mate":
			uci_fields = uci_fields[2:]

		// 	* movetime
		// 	search exactly x mseconds
		case "movetime":
			time_limit, _ = strconv.Atoi(uci_fields[1])
			uci_fields = uci_fields[2:]
			per_move = true

		// 	* infinite
		// 	search until the "stop" command. Do not exit the search without being told so in this mode!
		case "infinite":
			per_move = true
		default:
			uci_fields = uci_fields[:1]
		}
	}

	if per_move {
		gt.PerMoveStart(time.Duration(time_limit) * time.Millisecond)
	} else {
		gt.PerGameStart(current_board.c)
	}

	Search(current_board, gt)

	wg.Done()
}

// position [fen  | startpos ]  moves  ....
func UCIPosition(uci_fields []string) *Board {
	var brd *Board
	if len(uci_fields) == 0 || uci_fields[0] == "startpos" {
		brd = ParseFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		if len(uci_fields) > 1 {
			PlayMoveSequence(brd, uci_fields[1:])
		}
	} else if uci_fields[0] == "fen" {
		brd = ParseFENSlice(uci_fields[1:])
		if len(uci_fields) > 7 {
			PlayMoveSequence(brd, uci_fields[7:])
		}
	} else {
		UCIInvalid(uci_fields)
	}
	return brd
}

func PlayMoveSequence(brd *Board, uci_fields []string) {
	var move Move
	if uci_fields[0] == "moves" {
		uci_fields = uci_fields[1:]
	} else if uci_fields[1] == "moves" {
		uci_fields = uci_fields[2:]
	}
	for _, move_str := range uci_fields {
		move = ParseMove(brd, move_str)
		make_move(brd, move)
	}
}

func StartPos() *Board {
	return ParseFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}
