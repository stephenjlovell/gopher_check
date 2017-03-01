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

func (s *MoveSelector) CurrentStage() int {
	return s.stage - 1
}

func (s *QMoveSelector) CurrentStage() int {
	return s.stage - 1
}

type MoveSelector struct {
	winning        MoveList
	losing         MoveList
	remainingMoves MoveList
	mu             sync.Mutex
	brd            *Board
	thisStk        *StackItem
	htable         *HistoryTable
	stage          int
	index          int
	finished       int
	firstMove      Move
	inCheck        bool
}

type QMoveSelector struct {
	winning        MoveList
	losing         MoveList
	remainingMoves MoveList
	checks         MoveList
	brd            *Board
	thisStk        *StackItem
	htable         *HistoryTable
	stage          int
	index          int
	finished       int
	inCheck        bool
	canCheck       bool
}

func NewMoveSelector(brd *Board, thisStk *StackItem, htable *HistoryTable, inCheck bool, firstMove Move) *MoveSelector {
	return &MoveSelector{
		brd:            brd,
		thisStk:        thisStk,
		htable:         htable,
		inCheck:        inCheck,
		firstMove:      firstMove,
		winning:        NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
		losing:         NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
		remainingMoves: NewMoveList(QUIET_MOVE_LIST_LENGTH),
	}
}

func NewQMoveSelector(brd *Board, thisStk *StackItem, htable *HistoryTable, inCheck, canCheck bool) *QMoveSelector {
	return &QMoveSelector{
		brd:            brd,
		thisStk:        thisStk,
		htable:         htable,
		inCheck:        inCheck,
		canCheck:       canCheck,
		winning:        NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
		losing:         NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
		checks:         NewMoveList(DEFAULT_MOVE_LIST_LENGTH),
		remainingMoves: NewMoveList(QUIET_MOVE_LIST_LENGTH),
	}
}

func (s *MoveSelector) Next(spType int) (Move, int) {
	if spType == SP_NONE {
		return s.NextMove()
	} else {
		return s.NextSPMove()
	}
}

func (s *MoveSelector) NextSPMove() (Move, int) {
	s.mu.Lock()
	m, stage := s.NextMove()
	s.mu.Unlock()
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
			if s.brd.ValidMove(s.firstMove, s.inCheck) && s.brd.LegalMove(s.firstMove, s.inCheck) {
				return s.firstMove, STAGE_FIRST
			}
		case STAGE_WINNING:
			m := s.winning[s.index].move
			s.index++
			if m != s.firstMove && s.brd.AvoidsCheck(m, s.inCheck) {
				return m, STAGE_WINNING
			}
		case STAGE_KILLER:
			m := s.thisStk.killers[s.index]
			s.index++
			if m != s.firstMove && s.brd.ValidMove(m, s.inCheck) && s.brd.LegalMove(m, s.inCheck) {
				return m, STAGE_KILLER
			}
		case STAGE_LOSING:
			m := s.losing[s.index].move
			s.index++
			if m != s.firstMove && s.brd.AvoidsCheck(m, s.inCheck) {
				return m, STAGE_LOSING
			}
		case STAGE_REMAINING:
			m := s.remainingMoves[s.index].move
			s.index++
			if m != s.firstMove && !s.thisStk.IsKiller(m) && s.brd.AvoidsCheck(m, s.inCheck) {
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
		if s.inCheck {
			getEvasions(s.brd, s.htable, &s.winning, &s.losing, &s.remainingMoves)
		} else {
			getCaptures(s.brd, s.htable, &s.winning, &s.losing)
		}
		s.winning.Sort()
		s.finished = len(s.winning)
	case STAGE_KILLER:
		s.finished = KILLER_COUNT
	case STAGE_LOSING:
		s.losing.Sort()
		s.finished = len(s.losing)
	case STAGE_REMAINING:
		if !s.inCheck {
			getNonCaptures(s.brd, s.htable, &s.remainingMoves)
		}
		s.remainingMoves.Sort()
		s.finished = len(s.remainingMoves)
	default:
		s.finished = 0
		done = true
	}
	s.stage++
	return done
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
			if s.brd.AvoidsCheck(m, s.inCheck) {
				return m
			}
		case Q_STAGE_LOSING:
			m := s.losing[s.index].move
			s.index++
			if s.brd.AvoidsCheck(m, s.inCheck) {
				return m
			}
		case Q_STAGE_REMAINING:
			m := s.remainingMoves[s.index].move
			s.index++
			if s.brd.AvoidsCheck(m, s.inCheck) {
				return m
			}
		case Q_STAGE_CHECKS:
			m := s.checks[s.index].move
			s.index++
			if s.brd.AvoidsCheck(m, s.inCheck) {
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
		if s.inCheck {
			getEvasions(s.brd, s.htable, &s.winning, &s.losing, &s.remainingMoves)
		} else {
			getWinningCaptures(s.brd, s.htable, &s.winning)
		}
		s.winning.Sort()
		s.finished = len(s.winning)
	case Q_STAGE_LOSING:
		s.losing.Sort()
		s.finished = len(s.losing)
	case Q_STAGE_REMAINING:
		s.remainingMoves.Sort()
		s.finished = len(s.remainingMoves)
	case Q_STAGE_CHECKS:
		if !s.inCheck && s.canCheck {
			getChecks(s.brd, s.htable, &s.checks)
			s.checks.Sort()
		}
		s.finished = len(s.checks)
	default:
		done = true
	}
	s.stage++
	return done
}
