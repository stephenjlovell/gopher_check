//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

// Current search stages:
// 1. Hash move if available
// 2. IID move if no hash move available.
// 3. Evasions or Winning captures/promotions via getBestMoves(). No pruning - extensions only.
// 4. All other moves via getRemainingMoves().  Futility pruning and Late-move reductions applied.
// Q-search stages
// 1. Evasions or winning captures/promotions getBestMoves(). Specialized futility pruning.
// 2. Non-captures that give check via getChecks().

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
	inCheck        bool
	winning         MoveList
	losing          MoveList
	remainingMoves MoveList
	brd             *Board
	thisStk        *StackItem
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
	firstMove Move
}

type QMoveSelector struct {
	AbstractSelector
	checks    MoveList
	recycler  *Recycler
	canCheck bool
}

func NewMoveSelector(brd *Board, thisStk *StackItem, htable *HistoryTable, inCheck bool, firstMove Move) *MoveSelector {
	return &MoveSelector{
		AbstractSelector: AbstractSelector{
			brd:      brd,
			thisStk: thisStk,
			htable:   htable,
			inCheck: inCheck,
		},
		firstMove: firstMove,
	}
}

func NewQMoveSelector(brd *Board, thisStk *StackItem, htable *HistoryTable, recycler *Recycler, inCheck, canCheck bool) *QMoveSelector {
	return &QMoveSelector{
		AbstractSelector: AbstractSelector{
			brd:      brd,
			thisStk: thisStk,
			htable:   htable,
			inCheck: inCheck,
		},
		canCheck: canCheck,
		recycler:  recycler,
	}
}

func (s *MoveSelector) Next(recycler *Recycler, spType int) (Move, int) {
	if spType == SP_NONE {
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

func (s *MoveSelector) NextBatch(recycler *Recycler) bool {
	done := false
	s.index = 0
	switch s.stage {
	case STAGE_FIRST:
		s.finished = 1
	case STAGE_WINNING:
		if s.inCheck {
			s.winning = recycler.AttemptReuse()
			s.losing = recycler.AttemptReuse()
			s.remainingMoves = recycler.AttemptReuse()
			getEvasions(s.brd, s.htable, &s.winning, &s.losing, &s.remainingMoves)
			// fmt.Printf("%t,%t,%t,", len(s.winning) > 8, len(s.losing) > 8, len(s.remainingMoves) > 8)
		} else {
			s.winning = recycler.AttemptReuse()
			s.losing = recycler.AttemptReuse()
			getCaptures(s.brd, s.htable, &s.winning, &s.losing)
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
		if !s.inCheck {
			s.remainingMoves = recycler.AttemptReuse()
			getNonCaptures(s.brd, s.htable, &s.remainingMoves)
			// fmt.Printf("%t,", len(s.remainingMoves) > 8)
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

func (s *MoveSelector) Recycle(recycler *Recycler) {
	s.recycleList(recycler, s.winning)
	s.recycleList(recycler, s.losing)
	s.recycleList(recycler, s.remainingMoves)
	s.winning, s.losing, s.remainingMoves = nil, nil, nil
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
			s.winning = s.recycler.AttemptReuse()
			s.losing = s.recycler.AttemptReuse()
			s.remainingMoves = s.recycler.AttemptReuse()
			getEvasions(s.brd, s.htable, &s.winning, &s.losing, &s.remainingMoves)
		} else {
			s.winning = s.recycler.AttemptReuse()
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
			s.checks = s.recycler.AttemptReuse()
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

func (s *QMoveSelector) Recycle() {
	s.recycleList(s.recycler, s.winning)
	s.recycleList(s.recycler, s.losing)
	s.recycleList(s.recycler, s.remainingMoves)
	s.recycleList(s.recycler, s.checks)
	s.winning, s.losing, s.remainingMoves, s.checks = nil, nil, nil, nil
}
