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
// "fmt"
)

type PV struct {
	m     Move
	value int
	depth int
	next  *PV
}

func (pv *PV) ToUCI() string {
	if pv == nil || !pv.m.IsMove() {
		return ""
	}
	str := pv.m.ToUCI()
	for curr_pv := pv.next; curr_pv != nil; curr_pv = curr_pv.next {
		if !curr_pv.m.IsMove() {
			break
		}
		str += " " + curr_pv.m.ToUCI()
	}
	return str
}

var pv_accuracy [2]int

func (pv *PV) SavePV(brd *Board, value, depth int) {
	var last_value int
	var m Move
	var in_check bool

	copy := brd.Copy() // create a local copy of the board to avoid having to unmake moves.
	for pv != nil {
		m = pv.m
		in_check = copy.InCheck()

		if !copy.ValidMove(m, in_check) || !copy.LegalMove(m, in_check) {
			break
		}

		last_value = pv.value
		main_tt.store(copy, m, pv.depth, EXACT, pv.value)

		make_move(copy, m)
		pv = pv.next
	}

	if abs(value) == abs(last_value) {
		pv_accuracy[1] += 1
	} else {
		pv_accuracy[0] += 1
	}

}

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
