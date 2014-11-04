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

// import (
//   "fmt"
// )

func is_on_pv(brd *Board) bool {
	return false
}

type PV struct {
	m     Move
	value int
	next  *PV
}

func (pv *PV) ToUCI() string {
	str := ""
	remaining := pv
	if remaining != nil {
		str = remaining.m.ToUCI()
		for remaining.next != nil {
			remaining = remaining.next
			str += " " + remaining.m.ToUCI()
		}
	}
	return str
}

// Save the PV entries to the main transposition table.
func (pv *PV) Save(brd *Board, depth int) {
	copy := brd.Copy() // create a local copy of the board to avoid having to unmake moves.
	remaining := pv
	extensions_left := MAX_EXT
	var m Move

	for remaining != nil {
		m = remaining.m
		if !m.IsValid(copy) {
			break
		}

		if is_in_check(brd) && extensions_left > 0 {
			if MAX_EXT > extensions_left { // only extend after the first check.
				depth += 1
			}
			extensions_left -= 1
		}

		main_tt.store(copy, m, depth, EXACT, remaining.value)

		make_move(copy, m)

		if m.IsPromotion() {
			extensions_left -= 1
		} else {
			depth -= 1
		}

		remaining = remaining.next
	}
}

// Implementation options - linked list.

// Creation:

// At each node:
// Create a new PV struct.
// Each search returns a pointer to the finalized local pv on completion.
// When a move is > alpha and < beta, copy the move to the local struct and append its PV to the local.
// When no move is a PV move, any local PVs beneath are discarded. If there is a best move, it should still be returned as the last item in the pv list

// Root returns a pointer to final PV for current iteration to ID.

// Node criteria as defined by Onno Garms:
// http://www.talkchess.com/forum/viewtopic.php?t=38408&postdays=0&postorder=asc&topic_view=flat&start=10

// The root node is a PV node.
// The first child of a PV node is a PV node.
// The further children are searched by a scout search as CUT nodes.

// Research is done as PV nodes.

// The node after a null move is a CUT node.
// Internal iterative deeping does not change the node type.

// The first child of a CUT node is an ALL node.
// Further children of a CUT node are CUT nodes.
// Children of ALL nodes are CUT nodes.


// The first node of bad pruning is a CUT node.
// The first node of null move verification is a CUT node





