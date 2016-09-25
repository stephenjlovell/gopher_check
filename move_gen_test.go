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
// 	legal_movegen(Perft)
// }

// func TestMoveValidation(t *testing.T) {
// 	legal_movegen(PerftValidation)
// }

func legal_movegen(fn func(brd *Board, htable *HistoryTable, stk Stack, depth, ply int) int) {
	htable := new(HistoryTable)
	// brd := StartPos()
	brd := ParseFENString("n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1")
	// brd := ParseFENString("5r1k/1b3p1p/pp3p1q/3n4/1P2R3/P2B1PP1/7P/6K1 w - - 0 1")

	brd.Print()
	copy := brd.Copy()
	depth := 5
	start := time.Now()
	stk := make(Stack, MAX_STACK, MAX_STACK)
	sum := fn(brd, htable, stk, depth, 0)
	elapsed := time.Since(start)
	nps := int64(float64(sum) / elapsed.Seconds())

	fmt.Printf("%d nodes at depth %d. %d NPS\n", sum, depth, nps)

	CompareBoards(copy, brd)
	assert(*brd == *copy, "move generation did not return to initial board state.")
	assert(sum == legal_max_tree[depth], "Expected "+strconv.Itoa(legal_max_tree[depth])+
		" nodes, got "+strconv.Itoa(sum))
}

func Perft(brd *Board, htable *HistoryTable, stk Stack, depth, ply int) int {
	sum := 0
	in_check := brd.InCheck()
	this_stk := stk[ply]
	memento := brd.NewMemento()
	generator := NewMoveSelector(brd, &this_stk, htable, in_check, NO_MOVE)

	for m, _ := generator.Next(SP_NONE); m != NO_MOVE; m, _ = generator.Next(SP_NONE) {
		if depth > 1 {
			make_move(brd, m)
			sum += Perft(brd, htable, stk, depth-1, ply+1)
			// if depth-1 > 0 {
			// 	fmt.Printf("%s\n", m.ToUCI())
			// }
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
	generator := NewMoveSelector(brd, &this_stk, htable, false, NO_MOVE)
	for m, _ := generator.Next(SP_NONE); m != NO_MOVE; m, _ = generator.Next(SP_NONE) {
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
