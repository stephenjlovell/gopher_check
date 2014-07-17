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



type MV uint32
// To to fit into transposition table entries, moves are encoded using 21 bits as follows (in LSB order):
// From square - first 6 bits
// To square - next 6 bits
// Piece - next 3 bits
// Captured piece - next 3 bits
// promoted to - next 3 bits


// define an interface shared by all moves.
type AnyMove interface {
  From() int 
  To() int 
  Piece() PC  
  CapturedPiece() PC  
  PromotedTo() PC
} 


func (m MV) From() int {
  return int(uint32(m) & uint32(63))  
}

func (m MV) To() int {
  return int((uint32(m) >> 6) & uint32(63))
}

func (m MV) Piece(c int) PC {
  return (PC((uint32(m) >> 12) & uint32(7))<<1) | PC(c)
}

func (m MV) CapturedPiece(e int) PC {
  return (PC((uint32(m) >> 15) & uint32(7))<<1) | PC(e)
}

func (m MV) PromotedTo(c int) PC {
  return PC(((uint32(m) >> 18) & uint32(7))<<1) | PC(c)
}


// regular_move:           Proc.new { |*args| RegularMove.new                },  
// regular_capture:        Proc.new { |*args| RegularCapture.new(*args)      },  # captured_piece
// castle:                 Proc.new { |*args| Castle.new(*args)              },  # rook, rook_from, rook_to
// enp_capture:            Proc.new { |*args| EnPassantCapture.new(*args)    },  # captured_piece, enp_target
// pawn_move:              Proc.new { |*args| PawnMove.new                   },  
// enp_advance:            Proc.new { |*args| EnPassantAdvance.new           },
// pawn_promotion:         Proc.new { |*args| PawnPromotion.new(*args)       },  # promoted_piece
// pawn_promotion_capture: Proc.new { |*args| PawnPromotionCapture.new(*args)} } 




















