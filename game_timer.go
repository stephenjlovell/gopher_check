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
)

const (
  MOVES_PER_GAME = 55
  MAX_TIME = time.Duration(120000) * time.Millisecond // default search time limit
)

type GameTimer struct {
  inc [2]time.Duration
  remaining [2]time.Duration
  moves_remaining, max_depth int
  start_time time.Time
  timer *time.Timer
}

func NewGameTimer(moves_played int) *GameTimer {
  return &GameTimer{
    moves_remaining: max(1, MOVES_PER_GAME - moves_played),
    max_depth: MAX_DEPTH,
    remaining: [2]time.Duration{ MAX_TIME, MAX_TIME },
  }
}

func (g *GameTimer) DepthBasedStart() {
  g.start_time = time.Now()
}

func (g *GameTimer) PerMoveStart(time_limit time.Duration) {
  g.start_time = time.Now()
  g.timer = time.AfterFunc(time_limit, AbortSearch)
}

func (g *GameTimer) PerGameStart(c uint8) {
  // time_limit := (g.remaining[c] + (g.inc[c] * time.Duration(g.moves_remaining))) /
  //   time.Duration(g.moves_remaining)
  time_limit := g.remaining[c] / time.Duration(g.moves_remaining)
  g.PerMoveStart(time_limit)
}

func (g *GameTimer) Elapsed() time.Duration {
  return time.Since(g.start_time)
}

func (g *GameTimer) Stop() {
  g.timer.Stop()
}






//
