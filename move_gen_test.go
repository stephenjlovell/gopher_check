//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"strconv"
	// "testing"
	"time"
)

// var legalMaxTree = [10]int{1, 20, 400, 8902, 197281, 4865609, 119060324, 3195901860, 84998978956, 2439530234167}
var legalMaxTree = [10]int{1, 24, 496, 9483, 182838, 3605103, 71179139}

// func TestLegalMoveGen(t *testing.T) {
// 	setup()
// 	depth := 5
// 	legalMovegen(Perft, StartPos(), depth, legalMaxTree[depth], true)
// }

// func TestMoveValidation(t *testing.T) {
// 	setup()
// 	depth := 5
// 	legalMovegen(PerftValidation, StartPos(), depth, legalMaxTree[depth], true)
// }
//

// func TestPerftSuite(t *testing.T) {
// 	setup()
// 	depth := 4
// 	testPositions := loadEpdFile("test_suites/perftsuite.epd")  // http://www.rocechess.ch/perft.html
// 	for i, epd := range testPositions {
// 		if expected, ok := epd.nodeCount[depth]; ok {
// 			fmt.Printf("%d.", i+1)
// 			epd.brd.Print()
// 			legalMovegen(Perft, epd.brd, depth, expected, false)
// 		}
// 	}
// }

func legalMovegen(fn func(*Board, *HistoryTable, Stack, int, int) int, brd *Board, depth, expected int, verbose bool) {
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
	thisStk := stk[ply]
	memento := brd.NewMemento()
	recycler := loadBalancer.RootWorker().recycler
	generator := NewMoveSelector(brd, &thisStk, htable, inCheck, NO_MOVE)
	for m, _ := generator.Next(recycler, SP_NONE); m != NO_MOVE; m, _ = generator.Next(recycler, SP_NONE) {
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
	thisStk := stk[ply]
	memento := brd.NewMemento()
	// intentionally disregard whether king is in check while generating moves.
	recycler := loadBalancer.RootWorker().recycler
	generator := NewMoveSelector(brd, &thisStk, htable, false, NO_MOVE)
	for m, _ := generator.Next(recycler, SP_NONE); m != NO_MOVE; m, _ = generator.Next(recycler, SP_NONE) {
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

func CompareBoards(brd, other *Board) bool {
	equal := true
	if brd.pieces != other.pieces {
		fmt.Println("Board.pieces unequal")
		equal = false
	}
	if brd.squares != other.squares {
		fmt.Println("Board.squares unequal")
		fmt.Println("original:")
		brd.Print()
		fmt.Println("new board:")
		other.Print()
		equal = false
	}
	if brd.occupied != other.occupied {
		fmt.Println("Board.occupied unequal")
		for i := 0; i < 2; i++ {
			fmt.Printf("side: %d\n", i)
			fmt.Println("original:")
			brd.occupied[i].Print()
			fmt.Println("new board:")
			other.occupied[i].Print()
		}
		equal = false
	}
	if brd.material != other.material {
		fmt.Println("Board.material unequal")
		equal = false
	}
	if brd.hashKey != other.hashKey {
		fmt.Println("Board.hashKey unequal")
		equal = false
	}
	if brd.pawnHashKey != other.pawnHashKey {
		fmt.Println("Board.pawnHashKey unequal")
		equal = false
	}
	if brd.c != other.c {
		fmt.Println("Board.c unequal")
		equal = false
	}
	if brd.castle != other.castle {
		fmt.Println("Board.castle unequal")
		equal = false
	}
	if brd.enpTarget != other.enpTarget {
		fmt.Println("Board.enpTarget unequal")
		equal = false
	}
	if brd.halfmoveClock != other.halfmoveClock {
		fmt.Println("Board.halfmoveClock unequal")
		equal = false
	}
	if brd.endgameCounter != other.endgameCounter {
		fmt.Println("Board.endgameCounter unequal")
		equal = false
	}
	return equal
}
