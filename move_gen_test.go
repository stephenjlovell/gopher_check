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

// var legal_max_tree = [10]int{1, 20, 400, 8902, 197281, 4865609, 119060324, 3195901860, 84998978956, 2439530234167}
var legal_max_tree = [10]int{1, 24, 496, 9483, 182838, 3605103, 71179139}

// func TestLegalMoveGen(t *testing.T) {
// 	setup()
// 	depth := 5
// 	legal_movegen(Perft, StartPos(), depth, legal_max_tree[depth], true)
// }

// func TestMoveValidation(t *testing.T) {
// 	setup()
// 	depth := 5
// 	legal_movegen(PerftValidation, StartPos(), depth, legal_max_tree[depth], true)
// }
//

// func TestPerftSuite(t *testing.T) {
// 	setup()
// 	depth := 4
// 	test_positions := load_epd_file("test_suites/perftsuite.epd")  // http://www.rocechess.ch/perft.html
// 	for i, epd := range test_positions {
// 		if expected, ok := epd.node_count[depth]; ok {
// 			fmt.Printf("%d.", i+1)
// 			epd.brd.Print()
// 			legal_movegen(Perft, epd.brd, depth, expected, false)
// 		}
// 	}
// }

func legal_movegen(fn func(*Board, *HistoryTable, Stack, int, int) int, brd *Board, depth, expected int, verbose bool) {
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
	in_check := brd.InCheck()
	this_stk := stk[ply]
	memento := brd.NewMemento()
	recycler := load_balancer.RootWorker().recycler
	generator := NewMoveSelector(brd, &this_stk, htable, in_check, NO_MOVE)
	for m, _ := generator.Next(recycler, SP_NONE); m != NO_MOVE; m, _ = generator.Next(recycler, SP_NONE) {
		if depth > 1 {
			make_move(brd, m)
			sum += Perft(brd, htable, stk, depth-1, ply+1)
			unmake_move(brd, m, memento)
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
	this_stk := stk[ply]
	memento := brd.NewMemento()
	// intentionally disregard whether king is in check while generating moves.
	recycler := load_balancer.RootWorker().recycler
	generator := NewMoveSelector(brd, &this_stk, htable, false, NO_MOVE)
	for m, _ := generator.Next(recycler, SP_NONE); m != NO_MOVE; m, _ = generator.Next(recycler, SP_NONE) {
		in_check := brd.InCheck()
		if !brd.ValidMove(m, in_check) || !brd.LegalMove(m, in_check) {
			continue // rely on validation to prevent illegal moves...
		}
		make_move(brd, m)
		sum += PerftValidation(brd, htable, stk, depth-1, ply+1)
		unmake_move(brd, m, memento)
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
	if brd.hash_key != other.hash_key {
		fmt.Println("Board.hash_key unequal")
		equal = false
	}
	if brd.pawn_hash_key != other.pawn_hash_key {
		fmt.Println("Board.pawn_hash_key unequal")
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
	if brd.enp_target != other.enp_target {
		fmt.Println("Board.enp_target unequal")
		equal = false
	}
	if brd.halfmove_clock != other.halfmove_clock {
		fmt.Println("Board.halfmove_clock unequal")
		equal = false
	}
	if brd.endgame_counter != other.endgame_counter {
		fmt.Println("Board.endgame_counter unequal")
		equal = false
	}
	return equal
}
