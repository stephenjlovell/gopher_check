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

// Implements time management features for different time rules:
// https://chessprogramming.wikispaces.com/Time+Management

// Time control can be per-move, or per-game.
// Per-game time control consists of a base amount of time, plus an increment of additional
// time granted at the beginning of each move.

package main

import (
	"time"
	// "fmt"
)

const (
	AVG_MOVES_PER_GAME = 50
	MAX_TIME       = time.Duration(8) * time.Hour // default search time limit
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
		moves_remaining: max(20, AVG_MOVES_PER_GAME-moves_played),
		remaining:       [2]time.Duration{MAX_TIME, MAX_TIME},
		side_to_move:    side_to_move,
		start_time:      time.Now(),
	}
}

func (gt *GameTimer) SetMoveTime(time_limit time.Duration) {
	gt.remaining = [2]time.Duration{time_limit, time_limit}
	gt.moves_remaining = 1
	// UCIInfoString(fmt.Sprintln(gt.TimeLimit()))
}

func (gt *GameTimer) Start() {
	// UCIInfoString(fmt.Sprint(gt.TimeLimit()) +
	// 	fmt.Sprintf(" with %d moves remaining\n", gt.moves_remaining))
	gt.timer = time.AfterFunc(gt.TimeLimit(), gt.s.Abort)
}

func (gt *GameTimer) TimeLimit() time.Duration {
	return (gt.remaining[gt.side_to_move] / time.Duration(gt.moves_remaining)) -
		(time.Duration(20) * time.Millisecond) // safety margin
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
