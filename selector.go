//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
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

import "sync"

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

type AbstractSelector struct {
	sync.Mutex
	stage           int
	index           int
	finished        int
	in_check        bool
	winning         MoveList
	losing          MoveList
	remaining_moves MoveList
	brd             *Board
	this_stk        *StackItem
	htable          *HistoryTable
}

func (s *AbstractSelector) CurrentStage() int {
	return s.stage - 1
}

func (s *AbstractSelector) recycleList(recycler *Recycler, moves MoveList) {
	if moves != nil {
		recycler.Recycle(moves[0:0])
	}
}

type MoveSelector struct {
	AbstractSelector
	first_move Move
}

type QMoveSelector struct {
	AbstractSelector
	checks    MoveList
	recycler  *Recycler
	can_check bool
}

func NewMoveSelector(brd *Board, this_stk *StackItem, htable *HistoryTable, in_check bool, first_move Move) *MoveSelector {
	return &MoveSelector{
		AbstractSelector: AbstractSelector{
			brd:      brd,
			this_stk: this_stk,
			htable:   htable,
			in_check: in_check,
		},
		first_move: first_move,
	}
}

func NewQMoveSelector(brd *Board, this_stk *StackItem, htable *HistoryTable, recycler *Recycler, in_check, can_check bool) *QMoveSelector {
	return &QMoveSelector{
		AbstractSelector: AbstractSelector{
			brd:      brd,
			this_stk: this_stk,
			htable:   htable,
			in_check: in_check,
		},
		can_check: can_check,
		recycler:  recycler,
	}
}

func (s *MoveSelector) Next(recycler *Recycler, sp_type int) (Move, int) {
	if sp_type == SP_NONE {
		return s.NextMove(recycler)
	} else {
		return s.NextSPMove(recycler)
	}
}

func (s *MoveSelector) NextSPMove(recycler *Recycler) (Move, int) {
	s.Lock()
	m, stage := s.NextMove(recycler)
	s.Unlock()
	return m, stage
}

func (s *MoveSelector) NextMove(recycler *Recycler) (Move, int) {
	for {
		for s.index == s.finished {
			if s.NextBatch(recycler) {
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

func (s *MoveSelector) NextBatch(recycler *Recycler) bool {
	done := false
	s.index = 0
	switch s.stage {
	case STAGE_FIRST:
		s.finished = 1
	case STAGE_WINNING:
		if s.in_check {
			s.winning = recycler.AttemptReuse()
			s.losing = recycler.AttemptReuse()
			s.remaining_moves = recycler.AttemptReuse()
			get_evasions(s.brd, s.htable, &s.winning, &s.losing, &s.remaining_moves)
			// fmt.Printf("%t,%t,%t,", len(s.winning) > 8, len(s.losing) > 8, len(s.remaining_moves) > 8)
		} else {
			s.winning = recycler.AttemptReuse()
			s.losing = recycler.AttemptReuse()
			get_captures(s.brd, s.htable, &s.winning, &s.losing)
			// fmt.Printf("%t,%t,", len(s.winning) > 8, len(s.losing) > 8)
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
			s.remaining_moves = recycler.AttemptReuse()
			get_non_captures(s.brd, s.htable, &s.remaining_moves)
			// fmt.Printf("%t,", len(s.remaining_moves) > 8)
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

func (s *MoveSelector) Recycle(recycler *Recycler) {
	s.recycleList(recycler, s.winning)
	s.recycleList(recycler, s.losing)
	s.recycleList(recycler, s.remaining_moves)
	s.winning, s.losing, s.remaining_moves = nil, nil, nil
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
			s.winning = s.recycler.AttemptReuse()
			s.losing = s.recycler.AttemptReuse()
			s.remaining_moves = s.recycler.AttemptReuse()
			get_evasions(s.brd, s.htable, &s.winning, &s.losing, &s.remaining_moves)
		} else {
			s.winning = s.recycler.AttemptReuse()
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
			s.checks = s.recycler.AttemptReuse()
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
	s.recycleList(s.recycler, s.winning)
	s.recycleList(s.recycler, s.losing)
	s.recycleList(s.recycler, s.remaining_moves)
	s.recycleList(s.recycler, s.checks)
	s.winning, s.losing, s.remaining_moves, s.checks = nil, nil, nil, nil
}
