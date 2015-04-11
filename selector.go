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

// Current search stages:
// 1. Hash move if available
// 2. IID move if no hash move available.
// 3. Evasions or Winning captures/promotions via get_best_moves(). No pruning - extensions only.
// 4. All other moves via get_remaining_moves().  Futility pruning and Late-move reductions applied.
// Q-search stages
// 1. Evasions or winning captures/promotions get_best_moves(). Specialized futility prunins.
// 2. Non-captures that give check via get_checks().

import (
	// "fmt"
	"sync"
)
const (
	STAGE_FIRST = iota
	STAGE_WINNING
	// STAGE_KILLER
	STAGE_REMAINING
)
const (
	Q_STAGE_WINNING = iota
	Q_STAGE_REMAINING
	Q_STAGE_CHECKS
)

type SelectorInterface interface { // concrete Selector types must implement this interface
	next() Move
	next_batch() bool
}

type AbstractSelector struct {
	sync.Mutex
	brd             *Board
	stk             *StackItem
	stage           int
	index           int
	in_check        bool
	moves           MoveList 
	remaining_moves MoveList
}

type MoveSelector struct {
	AbstractSelector
	first_move      Move
}

type QMoveSelector struct {
	AbstractSelector
	can_check				bool
}

func NewMoveSelector(brd *Board, stk *StackItem, in_check bool, first_move Move) SelectorInterface {
	return &MoveSelector{
		AbstractSelector: AbstractSelector{
			brd:             brd,
			stk:             stk,
			in_check:        in_check,
			moves:           MoveList{},
			remaining_moves: MoveList{},
		},
		first_move:      first_move,
	}
}

func NewQMoveSelector(brd *Board, stk *StackItem, in_check, can_check bool) SelectorInterface {
	return &QMoveSelector{
		AbstractSelector: AbstractSelector{
			brd:             brd,
			stk:             stk,
			in_check:        in_check,
			moves:           MoveList{},
			remaining_moves: MoveList{},			
		},
		can_check:			 can_check,
	}
}

func (s *MoveSelector) next_shared() Move {
	s.Lock()
	m := s.next()
	s.Unlock()
	return m
}

func  (s *MoveSelector) next() Move {
	for {
		for s.index == len(s.moves) {
			if s.next_batch() {
				return NO_MOVE
			}
		}
		m := s.moves[s.index].move
		s.index++
		if (s.stage != STAGE_FIRST && m == s.first_move) || !avoids_check(s.brd, m, s.in_check) {
			continue
		} else {
			return m
		}
	}
}

func (s *MoveSelector) next_batch() bool {
	finished := false
	switch s.stage {
	case STAGE_FIRST:
		if s.first_move.IsValid(s.brd) && avoids_check(s.brd, s.first_move, s.in_check) {
			s.moves.Push(&SortItem{s.first_move, SORT_FIRST})
		}
	case STAGE_WINNING:
		if s.in_check {
			get_evasions(s.brd, &s.moves, &s.remaining_moves, &s.stk.killers)
		} else {
			get_captures(s.brd, &s.moves, &s.remaining_moves)
		}
		s.moves.Sort()
	case STAGE_REMAINING:
		if !s.in_check {
			get_non_captures(s.brd, &s.remaining_moves, &s.stk.killers)
		}
		s.remaining_moves.Sort()
		s.moves = append(s.moves, s.remaining_moves...)
	default:
		finished = true
	}
	s.stage++
	return finished
}

func (s *QMoveSelector) next() Move {
	for {
		for s.index == len(s.moves) {
			if s.next_batch() {
				return NO_MOVE
			}
		}
		m := s.moves[s.index].move
		s.index++
		if avoids_check(s.brd, m, s.in_check) {
			return m
		} else {
			continue
		}
	}
}

func (s *QMoveSelector) next_batch() bool {
	finished := false
	switch s.stage {
	case Q_STAGE_WINNING:
		if s.in_check {
			get_evasions(s.brd, &s.moves, &s.remaining_moves, &s.stk.killers)
		} else {
			get_captures(s.brd, &s.moves, &s.remaining_moves)
		}
		s.moves.Sort()
	case Q_STAGE_REMAINING:
		if s.in_check {
			if len(s.remaining_moves) > 0 {
				s.remaining_moves.Sort()
				s.moves = append(s.moves, s.remaining_moves...)
			}			
		} else if s.can_check {
			get_checks(s.brd, &s.remaining_moves, &s.stk.killers)
			s.remaining_moves.Sort()
			s.moves = append(s.moves, s.remaining_moves...)
		}
	default:
		finished = true
	}
	s.stage++
	return finished
}



