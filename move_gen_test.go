//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

var legalMaxTree = [10]int{1, 20, 400, 8902, 197281, 4865609, 119060324, 3195901860, 84998978956, 2439530234167}

func TestLegalMoveGen(t *testing.T) {
	depth := 5
	LegalMovegen(Perft, StartPos(), depth, legalMaxTree[depth], true)
}

func TestMoveValidation(t *testing.T) {
	depth := 5
	LegalMovegen(PerftValidation, StartPos(), depth, legalMaxTree[depth], true)
}

func TestPerftSuite(t *testing.T) {
	depth := 6
	testPositions, err := LoadEpdFile("test_suites/perftsuite.epd") // http://www.rocechess.ch/perft.html
	if err != nil {
		panic("could not load epd file")
	}

	for i, epd := range testPositions {
		if expected, ok := epd.nodeCount[depth]; ok {
			fmt.Printf("%d.", i+1)
			epd.brd.Print()
			fmt.Println(epd.fen)
			LegalMovegen(Perft, epd.brd, depth, expected, false)
		}
	}
}

// TODO: add parallelism

func LegalMovegen(fn func(*Board, *HistoryTable, Stack, int, int) int, brd *Board, depth, expected int, verbose bool) {
	htable := new(HistoryTable)
	copy := brd.Copy()
	start := time.Now()
	stk := make(Stack, MAX_STACK, MAX_STACK)

	sum := fn(brd, htable, stk, depth, 0)

	if verbose {
		elapsed := time.Since(start)
		nps := int64(float64(sum) / elapsed.Seconds())
		fmt.Printf("%d nodes at depth %d. %d NPS\n", sum, depth, nps)
		CompareBoards(copy, brd)
	}
	assert(*brd == *copy, "move generation did not return to initial board state.")
	assert(sum == expected, "Expected "+strconv.Itoa(expected)+" nodes, got "+strconv.Itoa(sum))
}

func Perft(brd *Board, htable *HistoryTable, stk Stack, depth, ply int) int {
	sum := 0
	inCheck := brd.InCheck()
	memento := brd.NewMemento()

	recycler := loadBalancer.RootWorker().recycler
	selector := recycler.ReuseMoveSelector(brd, htable, inCheck, NO_MOVE, stk[ply].killers)
	for m, _ := selector.Next(SP_NONE); m != NO_MOVE; m, _ = selector.Next(SP_NONE) {
		if depth > 1 {
			makeMove(brd, m)
			sum += Perft(brd, htable, stk, depth-1, ply+1)
			unmakeMove(brd, m, memento)
		} else {
			sum += 1
		}
	}

	return sum
}

func PerftValidation(brd *Board, htable *HistoryTable, stk Stack, depth, ply int) int {
	if depth == 0 {
		return 1
	}
	sum := 0
	memento := brd.NewMemento()
	// intentionally disregard whether king is in check while generating moves.
	recycler := loadBalancer.RootWorker().recycler
	selector := recycler.ReuseMoveSelector(brd, htable, false, NO_MOVE, stk[ply].killers)
	for m, _ := selector.Next(SP_NONE); m != NO_MOVE; m, _ = selector.Next(SP_NONE) {
		inCheck := brd.InCheck()
		if !brd.ValidMove(m, inCheck) || !brd.LegalMove(m, inCheck) {
			continue // rely on validation to prevent illegal moves...
		}
		makeMove(brd, m)
		sum += PerftValidation(brd, htable, stk, depth-1, ply+1)
		unmakeMove(brd, m, memento)
	}
	return sum
}
