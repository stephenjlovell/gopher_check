//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

// Implements time management features for different time rules:
// https://chessprogramming.wikispaces.com/Time+Management

// Time control can be per-move, or per-game.
// Per-game time control consists of a base amount of time, plus an increment of additional
// time granted at the beginning of each move.

package main

import (
	"time"
)

const (
	AVG_MOVES_PER_GAME  = 55
	MIN_MOVES_REMAINING = 15
	MAX_TIME            = time.Duration(8) * time.Hour          // default search time limit
	SAFETY_MARGIN       = time.Duration(500) * time.Millisecond // minimal amount of time to keep on clock
)

type GameTimer struct {
	inc             [2]time.Duration
	remaining       [2]time.Duration
	moves_remaining int
	start_time      time.Time
	timer           *time.Timer
	s               *Search
	side_to_move    uint8
}

func NewGameTimer(moves_played int, side_to_move uint8) *GameTimer {
	return &GameTimer{
		moves_remaining: max(MIN_MOVES_REMAINING, AVG_MOVES_PER_GAME-moves_played),
		remaining:       [2]time.Duration{MAX_TIME, MAX_TIME},
		side_to_move:    side_to_move,
		start_time:      time.Now(),
	}
}

func (gt *GameTimer) SetMoveTime(time_limit time.Duration) {
	gt.remaining = [2]time.Duration{time_limit, time_limit}
	gt.inc = [2]time.Duration{0, 0}
	gt.moves_remaining = 1
}

func (gt *GameTimer) Start() {
	gt.timer = time.AfterFunc(gt.TimeLimit(), gt.s.Abort)
}

func (gt *GameTimer) TimeLimit() time.Duration {
	return (gt.remaining[gt.side_to_move] - SAFETY_MARGIN) / time.Duration(gt.moves_remaining)
}

func (gt *GameTimer) Elapsed() time.Duration {
	return time.Since(gt.start_time)
}

func (gt *GameTimer) Stop() {
	if gt.timer != nil {
		gt.timer.Stop()
	}
}

//
