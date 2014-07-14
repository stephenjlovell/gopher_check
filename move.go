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




// define an interface shared by all moves.
type AnyMove interface {
  Make() 
  Unmake()
  // Hash()
  // PawnHash()

  // 24 bits, in LSB order:
  Piece() uint32  // 4 bits
  From() uint32  // 6 bits
  To() uint32 // 6 bits
  CapturedPiece() uint32  // 4 bits
  PromotedTo() uint32  // 4 bits
} // Space remaining => 8 bits

type MV uint32

func (m MV) Make(brd *BRD) {

}
func (m MV) Unmake(brd *BRD) {

}
func (m MV) Piece() uint32 {
  return uint32(m) & uint32(15)
}
func (m MV) From() uint32 {
  return (uint32(m) >> 4) & uint32(63)  // discard the first 4 bits and return the next 6 bits
}
func (m MV) To() uint32 {
  return (uint32(m) >> 10) & uint32(63)  
}
func (m MV) CapturedPiece() uint32 {
  return (uint32(m) >> 16) & uint32(15)
}
func (m MV) PromotedTo() uint32 {
  return (uint32(m) >> 20) & uint32(15)
}


// use automatic delegation to call make/unmake of strategy


// regular_move:           Proc.new { |*args| RegularMove.new                },  
// regular_capture:        Proc.new { |*args| RegularCapture.new(*args)      },  # captured_piece
// castle:                 Proc.new { |*args| Castle.new(*args)              },  # rook, rook_from, rook_to
// enp_capture:            Proc.new { |*args| EnPassantCapture.new(*args)    },  # captured_piece, enp_target
// pawn_move:              Proc.new { |*args| PawnMove.new                   },  
// enp_advance:            Proc.new { |*args| EnPassantAdvance.new           },
// pawn_promotion:         Proc.new { |*args| PawnPromotion.new(*args)       },  # promoted_piece
// pawn_promotion_capture: Proc.new { |*args| PawnPromotionCapture.new(*args)} } 




















