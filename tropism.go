//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
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

var tropism_bonus [64][64][6]int

func setup_bonus_table(){
  base_bonus_ratio := 0.15
  bonus := 0.0
  for f := 0; f < 64; f++ {
    for to := 0; to < 64; to++ {
      for t := PAWN; t < KING; t++ {
        // bonus = piece_values[t] * base_bonus_ratio * manhattan_distance_ratio(f, to);
        bonus = piece_values[t] * base_bonus_ratio * chebyshev_distance_ratio(f, to);
        tropism_bonus[f][to][t] = round(bonus);
      }
    }
  }
}

// Returns 1 (maximum bonus) at minimum distance, and 0 (no bonus) at max distance.
func chebyshev_distance_ratio(from, to int) float {
  return (-float(chebyshev_distance(from, to))/6.0) + (7.0/6.0)
}
// Returns 1 (maximum bonus) at minimum distance, and 0 (no bonus) at max distance.
func manhattan_distance_ratio(from, to int) float {
  return (-float(manhattan_distance(from, to))/13.0) + (14.0/13.0)
}




