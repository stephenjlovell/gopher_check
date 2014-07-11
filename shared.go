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

import(
  "fmt"
)

type BB uint64

type BRD struct {
  pieces [2][6]BB
  occupied [2]BB
  material [2]int
  squares [64]int
}

const (
  NW = iota; NE; SE; SW; NORTH; EAST; SOUTH; WEST; INVALID;
)

const (
  A1=iota; B1; C1; D1; E1; F1; G1; H1; 
  A2; B2; C2; D2; E2; F2; G2; H2; 
  A3; B3; C3; D3; E3; F3; G3; H3; 
  A4; B4; C4; D4; E4; F4; G4; H4; 
  A5; B5; C5; D5; E5; F5; G5; H5; 
  A6; B6; C6; D6; E6; F6; G6; H6; 
  A7; B7; C7; D7; E7; F7; G7; H7; 
  A8; B8; C8; D8; E8; F8; G8; H8; 
)

const (
  BLACK = iota; WHITE
)

const (
  PAWN = iota; KNIGHT; BISHOP; ROOK; QUEEN; KING;
)

var uni_mask BB = 0xffffffffffffffff;
var empty_mask BB = 0x0;

var row_masks [8]BB
var column_masks [8]BB
var ray_masks [8][64]BB

var pawn_attack_masks, pawn_passed_masks [2][64]BB

var pawn_isolated_masks, pawn_side_masks [64]BB

var intervening [64][64]BB

var knight_masks, bishop_masks, rook_masks, queen_masks, king_masks, sq_mask_on, sq_mask_off [64]BB

var pawn_from_offsets = [2][4]int{ {8, 16, 9, 7 }, {-8, -16, -7, -9 } }
var knight_offsets = [8]int{-17, -15, -10, -6, 6, 10, 15, 17}
var bishop_offsets = [4]int{7, 9, -7, -9}
var rook_offsets = [4]int{8, 1, -8, -1}
var king_offsets = [8]int{-9, -7, 7, 9, -8, -1, 1, 8}
var pawn_attack_offsets = [4]int{9, 7, -9, -7}
var pawn_advance_offsets = [4]int{8, 16, -8, -16}
var pawn_enpassant_offsets = [2]int{1, -1}

var piece_values = [6]int{100, 320, 333, 510, 880, 100000} // default piece values

var directions [64][64]int

func max(a,b int) int { if a > b { return a } else { return b } }
func min(a,b int) int { if a > b { return b } else { return a } }
func abs(x int) int { if x < 0 { return -x } else { return x } }
func round(x float64) int { if x >= 0 { return int(x+0.5) } else { return int(x-0.5) } }

func on_board(sq int) bool { return 0 <= sq && sq <= 63 }
func row(sq int) int { return sq >> 3 }
func column(sq int) int { return sq & 7 }

func manhattan_distance(from, to int) int { return abs(row(from)-row(to)) + abs(column(from)-column(to)) }
func chebyshev_distance(from, to int) int { return max(abs(row(from)-row(to)),abs(column(from)-column(to))) }

func clear_sq(sq int, b BB) BB { return (b & sq_mask_off[sq]) }  // no longer modifies b by reference
func add_sq(sq int, b BB) BB { return  (b | sq_mask_on[sq]) }  // no longer modifies b by reference


func lsb(b BB) int { return 0 }
func msb(b BB) int { return 0 }
func furthest_forward(c int, b BB) int { return 0 }
func pop_count(b BB) int { return 0 }

// #define lsb(bitboard) (__builtin_ctzl(bitboard))
// #define msb(bitboard) (63-__builtin_clzl(bitboard))
// #define furthest_forward(color, bitboard) (color ? lsb(bitboard) : msb(bitboard))  
// #define pop_count(bitboard) (__builtin_popcountl(bitboard))

func Occupied(brd *BRD) BB { return brd.occupied[0]|brd.occupied[1] }
func Placement(c int, brd *BRD) BB { return brd.occupied[c] }
func piece_type(piece_id int) int { return (piece_id & 0xe) >> 1 }
func piece_color(piece_id int) int { return piece_id & 1 }

// #define Occupied() ((brd->occupied[0])|(brd->occupied[1]))
// #define Placement(color) (brd->occupied[color])
// #define piece_type(piece_id)  ((piece_id & 0xe) >> 1 )
// #define piece_color(piece_id)  (piece_id & 0x1)

func main() {
  fmt.Println("Hello Chess World")
}


