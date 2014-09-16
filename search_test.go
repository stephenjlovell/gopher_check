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

package main

// import (
//   "fmt"
//   "testing"
// )

// func TestSearch(t *testing.T) {
//   setup()
//   ResetAll() // reset all shared data structures and prepare to start a new game.
//   current_board := StartPos()
//   // brd.PrintDetails()
//   move := Search(current_board, make([]Move, 0), MAX_DEPTH, MAX_TIME)
//   fmt.Printf("bestmove %s\n", move.ToString())

// }


// func TestBoardCopy(t *testing.T) {
//   setup()
//   brd := StartPos()

//   result := make(chan *Board)
//   counter := 0

//   for i := 8; i < 16; i++ {
//     copy := brd.Copy()
//     move := NewMove(i, i+8, PAWN, EMPTY, EMPTY)
//     counter++
//     go func(i int) {
//       make_move(copy, move)
//       fmt.Printf("%d ", i)
//       result <- copy
//     }(i)
//   }

//   fmt.Printf("\n")

//   for counter > 0 {
//     select {
//     case r := <-result:
//       r.Print()
//       counter--
//     }
//   }
//   brd.Print()
//   fmt.Println("TestBoardCopy complete.")
// }



















