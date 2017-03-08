//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"time"
)

const (
	AVG_MOVES_PER_GAME  = 55
	MIN_MOVES_REMAINING = 15
	MAX_TIME            = time.Duration(8) * time.Hour        // default search time limit
	SAFETY_MARGIN       = time.Duration(5) * time.Millisecond // minimal amount of time to keep on clock
)

// GameTimer implements time management features for different time controls.
// https://chessprogramming.wikispaces.com/Time+Management
// Time control can be per-move, or per-game.
// Per-game time control consists of a base amount of time, plus an increment of additional
// time granted at the beginning of each move.
type GameTimer struct {
	inc            [2]time.Duration
	remaining      [2]time.Duration
	movesRemaining int
	startTime      time.Time
	timer          *time.Timer
	s              *Search
	sideToMove     uint8
}

func NewGameTimer(movesPlayed int, sideToMove uint8) *GameTimer {
	return &GameTimer{
		movesRemaining: Max(MIN_MOVES_REMAINING, AVG_MOVES_PER_GAME-movesPlayed),
		remaining:      [2]time.Duration{MAX_TIME, MAX_TIME},
		sideToMove:     sideToMove,
		startTime:      time.Now(),
	}
}

func (gt *GameTimer) SetMoveTime(timeLimit time.Duration) {
	gt.remaining = [2]time.Duration{timeLimit, timeLimit}
	gt.inc = [2]time.Duration{0, 0}
	gt.movesRemaining = 1
}

func (gt *GameTimer) Start() {
	gt.timer = time.AfterFunc(gt.TimeLimit(), gt.s.Abort)
}

func (gt *GameTimer) TimeLimit() time.Duration {
	return (gt.remaining[gt.sideToMove] - SAFETY_MARGIN) / time.Duration(gt.movesRemaining)
}

func (gt *GameTimer) Elapsed() time.Duration {
	return time.Since(gt.startTime)
}

func (gt *GameTimer) Stop() {
	if gt.timer != nil {
		gt.timer.Stop()
	}
}
