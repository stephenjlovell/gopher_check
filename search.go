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

// Young Brothers Wait (YBW) approach
// At each node, search the leftmost child sequentially before searching the rest of the successors concurrently.
// Goal is to avoid wasted processing effort where a subtree is expanded that otherwise would have been pruned.

// Values are passed up call stack as normal.

// When this causes bounds in the node to update, ideally would want updated bounds to be piped down to each 
// frame of local search stack for each goroutine subtree search rooted at that node






// example of concurrent move generation.  
// Communication and deep copy cost probably would outweigh benefit of concurrent move gen...
func example_movegen_call(brd *BRD, depth, alpha, beta int) {

  moves := make(chan MV, 10) // create a channel that will receive moves created by MoveGen.

  go GenerateMoves(brd, moves) // generate moves concurrently.

  for {
    m, more_moves := <-moves // blocks until a move is ready to try.
    if more_moves {
      make_move(brd, m)  // proceed normally.
      value, count := alpha_beta(brd, depth, alpha, beta)
      unmake_move(brd, m)

      // test bounds etc.

    } else {
      break // no more moves to try.  Exit the loop.
    }
  }
}









