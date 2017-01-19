//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

// "fmt"

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
	for currPv := pv.next; currPv != nil; currPv = currPv.next {
		if !currPv.m.IsMove() {
			break
		}
		str += " " + currPv.m.ToUCI()
	}
	return str
}

var pvAccuracy [2]int

func (pv *PV) SavePV(brd *Board, value, depth int) {
	var lastValue int
	var m Move
	var inCheck bool

	copy := brd.Copy() // create a local copy of the board to avoid having to unmake moves.
	for pv != nil {
		m = pv.m
		inCheck = copy.InCheck()

		if !copy.ValidMove(m, inCheck) || !copy.LegalMove(m, inCheck) {
			break
		}

		lastValue = pv.value
		mainTt.store(copy, m, pv.depth, EXACT, pv.value)

		makeMove(copy, m)
		pv = pv.next
	}

	if abs(value) == abs(lastValue) {
		pvAccuracy[1] += 1
	} else {
		pvAccuracy[0] += 1
	}

}

// Node criteria as defined by Onno Garms:
// http://www.talkchess.com/forum/viewtopic.php?t=38408&postdays=0&postorder=asc&topicView=flat&start=10

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
