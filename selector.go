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
	STAGE_KILLER
	STAGE_LOSING
	STAGE_REMAINING
)
const (
	Q_STAGE_WINNING = iota
	Q_STAGE_LOSING
	Q_STAGE_REMAINING
	Q_STAGE_CHECKS
)

type SelectorInterface interface { // concrete Selector types must implement this interface
	Next(bool) Move
	NextBatch() bool
	CurrentStage() int
}

type AbstractSelector struct {
	sync.Mutex
	brd             *Board
	this_stk        *StackItem
	stage           int
	index           int
	finished        int
	in_check        bool
	winning         MoveList
	losing          MoveList
	remaining_moves MoveList
}

func (s *AbstractSelector) CurrentStage() int {
	return s.stage-1
}

type MoveSelector struct {
	AbstractSelector
	first_move Move
}

type QMoveSelector struct {
	AbstractSelector
	checks 						MoveList
	can_check 				bool
}

func NewMoveSelector(brd *Board, this_stk *StackItem, in_check bool, first_move Move) SelectorInterface {
	return &MoveSelector{
		AbstractSelector: AbstractSelector{
			brd:             brd,
			this_stk:        this_stk,
			in_check:        in_check,
			winning:         MoveList{},
			losing:          MoveList{},
			remaining_moves: MoveList{},
		},
		first_move: first_move,
	}
}

func NewQMoveSelector(brd *Board, this_stk *StackItem, in_check, can_check bool) SelectorInterface {
	return &QMoveSelector{
		AbstractSelector: AbstractSelector{
			brd:             brd,
			this_stk:        this_stk,
			in_check:        in_check,
			winning:         MoveList{},
			losing:          MoveList{},
			remaining_moves: MoveList{},
		},
		checks:						 MoveList{},
		can_check: can_check,
	}
}

func (s *MoveSelector) Next(is_sp bool) Move {
	if is_sp {
		return s.NextSPMove()
	} else {
		return s.NextMove()
	}
}

func (s *MoveSelector) NextSPMove() Move {
	sp_selector := s.this_stk.sp.selector
	sp_selector.Lock()
	m := sp_selector.NextMove()
	sp_selector.Unlock()
	return m
}

func (s *MoveSelector) NextMove() Move {
	for {
		for s.index == s.finished {
			if s.NextBatch() {
				return NO_MOVE
			}
		}
		switch s.stage - 1 {
		case STAGE_FIRST: 
			s.index++
			if s.brd.ValidMove(s.first_move, s.in_check) && s.brd.LegalMove(s.first_move, s.in_check) {
				return s.first_move
			} else {
				// if s.first_move.IsMove() { 
				// 	fmt.Println("first move invalid:")
				// 	s.brd.Print()
				// 	s.first_move.Print()
				// 	fmt.Println(s.brd.castle) 
				// }
			}
		case STAGE_WINNING:
			m := s.winning[s.index].move
			s.index++
			if m != s.first_move && s.brd.AvoidsCheck(m, s.in_check) {
				return m
			}
		case STAGE_KILLER:
			m := s.this_stk.killers[s.index]
			s.index++
			if m != s.first_move && s.brd.ValidMove(m, s.in_check) && s.brd.LegalMove(m, s.in_check) {
				return m
			}
		case STAGE_LOSING:
			m := s.losing[s.index].move
			s.index++
			if m != s.first_move && s.brd.AvoidsCheck(m, s.in_check) {
				return m
			}
		case STAGE_REMAINING:
			m := s.remaining_moves[s.index].move
			s.index++
			if m != s.first_move && !s.this_stk.IsKiller(m) && s.brd.AvoidsCheck(m, s.in_check) {
				return m
			}
		default:
		}
	}
}

func (s *MoveSelector) NextBatch() bool {
	done := false
	s.index = 0
	switch s.stage {
	case STAGE_FIRST:
		s.finished = 1
	case STAGE_WINNING:
		if s.in_check {
			get_evasions(s.brd, &s.winning, &s.losing, &s.remaining_moves)
		} else {
			get_captures(s.brd, &s.winning, &s.losing)
		}
		s.winning.Sort()
		s.finished = len(s.winning)
	case STAGE_KILLER:
		s.finished = KILLER_COUNT			
	case STAGE_LOSING:
		s.losing.Sort()
		s.finished = len(s.losing)
	case STAGE_REMAINING:
		if !s.in_check {
			get_non_captures(s.brd, &s.remaining_moves)
		}
		s.remaining_moves.Sort()
		s.finished = len(s.remaining_moves)
	default:
		done = true
	}
	s.stage++
	return done
}

func (s *QMoveSelector) Next(is_sp bool) Move {
	for {
		for s.index == s.finished {
			if s.NextBatch() {
				return NO_MOVE
			}
		}
		switch s.stage - 1 {
		case Q_STAGE_WINNING:
			m := s.winning[s.index].move
			s.index++
			if s.brd.AvoidsCheck(m, s.in_check) {
				return m
			}
		case Q_STAGE_LOSING:
			m := s.losing[s.index].move
			s.index++
			if s.brd.AvoidsCheck(m, s.in_check) {
				return m
			}
		case Q_STAGE_REMAINING:
			m := s.remaining_moves[s.index].move
			s.index++
			if s.brd.AvoidsCheck(m, s.in_check) {
				return m
			}
		case Q_STAGE_CHECKS:
			m := s.checks[s.index].move
			s.index++
			if s.brd.AvoidsCheck(m, s.in_check) {
				return m
			}
		default:
		}
	}
}

func (s *QMoveSelector) NextBatch() bool {
	done := false
	s.index = 0
	switch s.stage {
	case Q_STAGE_WINNING:
		if s.in_check {
			get_evasions(s.brd, &s.winning, &s.losing, &s.remaining_moves)
		} else {
			get_winning_captures(s.brd, &s.winning)
		}
		s.winning.Sort()
		s.finished = len(s.winning)
	case Q_STAGE_LOSING:
		s.losing.Sort()
		s.finished = len(s.losing)
	case Q_STAGE_REMAINING:
		s.remaining_moves.Sort()
		s.finished = len(s.remaining_moves)
	case Q_STAGE_CHECKS:
		if !s.in_check && s.can_check {
			get_checks(s.brd, &s.checks)
			s.checks.Sort()
		}
		s.finished = len(s.checks)
	default:
		done = true
	}
	s.stage++
	return done
}











