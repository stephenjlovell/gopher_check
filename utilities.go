//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"time"
)

func RunTestSuite(testSuite string, depth, timeout int) {
	test, err := LoadEpdFile(testSuite)
	if err != nil {
		fmt.Println(err)
		return
	}
	var moveStr string
	sum, score := 0, 0
	var gt *GameTimer
	var search *Search

	start := time.Now()
	for i, epd := range test {
		gt = NewGameTimer(0, epd.brd.c)
		gt.SetMoveTime(time.Duration(timeout) * time.Millisecond)
		search = NewSearch(SearchParams{depth, false, false, false}, gt, nil, nil)
		search.Start(epd.brd)

		moveStr = ToSAN(epd.brd, search.bestMove)
		if correctMove(epd, moveStr) {
			score += 1
			fmt.Printf("-")
		} else {
			fmt.Printf("%d.", i+1)
			// fmt.Printf("\n%s\n", epd.fen)
			// fmt.Println(moveStr)
			// fmt.Printf("best moves: %s\n", strings.Join(epd.bestMoves, ", "))
			// fmt.Printf("avoid moves: %s\n", strings.Join(epd.avoidMoves, ", "))
		}
		sum += search.nodes
		// search.htable.PrintMax()
	}
	secondsElapsed := time.Since(start).Seconds()
	mNodes := float64(sum) / 1000000.0
	fmt.Printf("\n%.3fm nodes searched in %.4fs (%.3fm NPS)\n",
		mNodes, secondsElapsed, mNodes/secondsElapsed)
	fmt.Printf("Total score: %d/%d\n", score, len(test))
	fmt.Printf("Overhead: %.3fm\n", float64(loadBalancer.Overhead())/1000000.0)
	fmt.Printf("Timeout: %.1fs\n", float64(timeout)/1000.0)
}

func correctMove(epd *EPD, moveStr string) bool {
	for _, a := range epd.avoidMoves {
		if moveStr == a {
			return false
		}
	}
	for _, b := range epd.bestMoves {
		if moveStr == b {
			return true
		}
	}
	return false
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

func isBoardConsistent(brd *Board) bool {
	var squares [64]Piece
	var occupied [2]BB
	var material [2]int16

	var sq int
	for sq = 0; sq < 64; sq++ {
		squares[sq] = NO_PIECE
	}
	consistent := true

	for c := uint8(BLACK); c <= WHITE; c++ {
		for pc := Piece(PAWN); pc <= KING; pc++ {
			if occupied[c]&brd.pieces[c][pc] > 0 {
				fmt.Printf("brd.pieces[%d][%d] overlaps with another pieces bitboard.\n", c, pc)
				consistent = false
			}
			occupied[c] |= brd.pieces[c][pc]

			for bb := brd.pieces[c][pc]; bb > 0; bb.Clear(sq) {
				sq = furthestForward(c, bb)
				material[c] += int16(pc.Value() + mainPst[c][pc][sq])
				if squares[sq] != NO_PIECE {
					fmt.Printf("brd.pieces[%d][%d] overlaps with another pieces bitboard at %s.\n", c, pc, SquareString(sq))
					consistent = false
				}
				squares[sq] = pc
			}
		}
	}

	if squares != brd.squares {
		fmt.Println("brd.squares inconsistent")
		consistent = false
	}
	if occupied != brd.occupied {
		fmt.Println("brd.occupied inconsistent")
		consistent = false
	}
	if material != brd.material {
		fmt.Println("brd.material inconsistent")
		consistent = false
	}

	return consistent
}
