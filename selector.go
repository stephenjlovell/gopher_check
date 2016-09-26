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
// 1. Evasions or winning captures/promotions get_best_moves(). Specialized futility pruning.
// 2. Non-captures that give check via get_checks().

import (
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

var move_list_pool chan MoveList = make(chan MoveList, 200)

// type SelectorInterface interface { // concrete Selector types must implement this interface
// 	Next(bool) Move
// 	NextBatch() bool
// 	CurrentStage() int
// }

type AbstractSelector struct {
	sync.Mutex
	brd             *Board
	this_stk        *StackItem
	htable          *HistoryTable
	stage           int
	index           int
	finished        int
	in_check        bool
	winning         MoveList
	losing          MoveList
	remaining_moves MoveList
}


func (s *AbstractSelector) allocate() MoveList {
	var moves MoveList
	select {
  case moves = <-move_list_pool:
	default:
		moves = make(MoveList, 0, 32)
	}
	return moves
}

func (s *AbstractSelector) recycleList(moves MoveList) {
	if moves != nil {
		select {
		case move_list_pool <- moves[0:0]:
		default:
		}
	}
}


func (s *AbstractSelector) CurrentStage() int {
	return s.stage - 1
}

type MoveSelector struct {
	AbstractSelector
	first_move Move
}

type QMoveSelector struct {
	AbstractSelector
	checks    MoveList
	can_check bool
}

func NewMoveSelector(brd *Board, this_stk *StackItem, htable *HistoryTable, in_check bool, first_move Move) *MoveSelector {
	return &MoveSelector{
		AbstractSelector: AbstractSelector{
			brd:             brd,
			this_stk:        this_stk,
			htable:          htable,
			in_check:        in_check,
		},
		first_move: first_move,
	}
}

func NewQMoveSelector(brd *Board, this_stk *StackItem, htable *HistoryTable, in_check, can_check bool) *QMoveSelector {
	return &QMoveSelector{
		AbstractSelector: AbstractSelector{
			brd:             brd,
			this_stk:        this_stk,
			htable:          htable,
			in_check:        in_check,
		},
		can_check: can_check,
	}
}

func (s *MoveSelector) Next(sp_type int) (Move, int) {
	if sp_type == SP_NONE {
		return s.NextMove()
	} else {
		return s.NextSPMove()
	}
}

func (s *MoveSelector) NextSPMove() (Move, int) {
	s.Lock()
	m, stage := s.NextMove()
	s.Unlock()
	return m, stage
}

func (s *MoveSelector) NextMove() (Move, int) {
	for {
		for s.index == s.finished {
			if s.NextBatch() {
				return NO_MOVE, s.CurrentStage()
			}
		}
		switch s.CurrentStage() {
		case STAGE_FIRST:
			s.index++
			if s.brd.ValidMove(s.first_move, s.in_check) && s.brd.LegalMove(s.first_move, s.in_check) {
				return s.first_move, STAGE_FIRST
			}
		case STAGE_WINNING:
			m := s.winning[s.index].move
			s.index++
			if m != s.first_move && s.brd.AvoidsCheck(m, s.in_check) {
				return m, STAGE_WINNING
			}
		case STAGE_KILLER:
			m := s.this_stk.killers[s.index]
			s.index++
			if m != s.first_move && s.brd.ValidMove(m, s.in_check) && s.brd.LegalMove(m, s.in_check) {
				return m, STAGE_KILLER
			}
		case STAGE_LOSING:
			m := s.losing[s.index].move
			s.index++
			if m != s.first_move && s.brd.AvoidsCheck(m, s.in_check) {
				return m, STAGE_LOSING
			}
		case STAGE_REMAINING:
			m := s.remaining_moves[s.index].move
			s.index++
			if m != s.first_move && !s.this_stk.IsKiller(m) && s.brd.AvoidsCheck(m, s.in_check) {
				return m, STAGE_REMAINING
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
			s.winning = s.allocate()
			s.losing = s.allocate()
			s.remaining_moves = s.allocate()
			get_evasions(s.brd, s.htable, &s.winning, &s.losing, &s.remaining_moves)
		} else {
			s.winning = s.allocate()
			s.losing = s.allocate()
			get_captures(s.brd, s.htable, &s.winning, &s.losing)
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
			s.remaining_moves = s.allocate()
			get_non_captures(s.brd, s.htable, &s.remaining_moves)
		}
		s.remaining_moves.Sort()
		s.finished = len(s.remaining_moves)
	default:
		s.finished = 0
		done = true
	}
	s.stage++
	return done
}

func (s *MoveSelector) Recycle() {
	s.recycleList(s.winning)
	s.recycleList(s.losing)
	s.recycleList(s.remaining_moves)
}

func (s *QMoveSelector) Next() Move {
	for {
		for s.index == s.finished {
			if s.NextBatch() {
				return NO_MOVE
			}
		}
		switch s.CurrentStage() {
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
			s.winning = s.allocate()
			s.losing = s.allocate()
			s.remaining_moves = s.allocate()
			get_evasions(s.brd, s.htable, &s.winning, &s.losing, &s.remaining_moves)
		} else {
			s.winning = s.allocate()
			get_winning_captures(s.brd, s.htable, &s.winning)
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
			s.checks = s.allocate()
			get_checks(s.brd, s.htable, &s.checks)
			s.checks.Sort()
		}
		s.finished = len(s.checks)
	default:
		done = true
	}
	s.stage++
	return done
}

func (s *QMoveSelector) Recycle() {
	s.recycleList(s.winning)
	s.recycleList(s.losing)
	s.recycleList(s.remaining_moves)
	s.recycleList(s.checks)
}
