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

func attack_map(brd *BRD, sq int) BB {
  var attacks, b_attackers, r_attackers BB
  occ := Occupied(brd)
  // Pawns
  attacks |= (pawn_attack_masks[BLACK][sq] & brd.pieces[WHITE][PAWN]) |
             (pawn_attack_masks[WHITE][sq] & brd.pieces[BLACK][PAWN])
  // Knights
  attacks |= (knight_masks[sq] & (brd.pieces[WHITE][KNIGHT]|brd.pieces[BLACK][KNIGHT]))
  // Bishops and Queens
  b_attackers = brd.pieces[WHITE][BISHOP] | brd.pieces[BLACK][BISHOP] | 
                brd.pieces[WHITE][QUEEN]  | brd.pieces[BLACK][QUEEN]
  attacks |= bishop_attacks(occ, sq) & b_attackers
  // Rooks and Queens
  r_attackers = brd.pieces[WHITE][ROOK]  | brd.pieces[BLACK][ROOK] | 
                brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]
  attacks |= rook_attacks(occ, sq) & r_attackers
  // Kings
  attacks |= king_masks[sq] & (brd.pieces[WHITE][KING]|brd.pieces[BLACK][KING])
  return attacks
}

func color_attack_map(brd *BRD, sq, c, e int) BB {
  var attacks, b_attackers, r_attackers BB
  occ := Occupied(brd)
  // Pawns
  attacks |= pawn_attack_masks[e][sq] & brd.pieces[c][PAWN]
  // Knights
  attacks |= knight_masks[sq] & brd.pieces[c][KNIGHT]
  // Bishops and Queens
  b_attackers = brd.pieces[c][BISHOP] | brd.pieces[c][QUEEN]
  attacks |= bishop_attacks(occ, sq) & b_attackers
  // Rooks and Queens
  r_attackers = brd.pieces[c][ROOK] | brd.pieces[c][QUEEN]
  attacks |= rook_attacks(occ, sq) & r_attackers
  // Kings
  attacks |= king_masks[sq] & brd.pieces[c][KING]
  return attacks
}

func is_attacked_by(brd *BRD, sq, attacker, defender int) bool {
  occ := Occupied(brd)
  // Pawns
  if pawn_attack_masks[defender][sq] & brd.pieces[attacker][PAWN] >0 { return true }
  // Knights
  if knight_masks[sq] & (brd.pieces[attacker][KNIGHT]) > 0 { return true }
  // Bishops and Queens
  if bishop_attacks(occ, sq) & (brd.pieces[attacker][BISHOP]|brd.pieces[attacker][QUEEN]) > 0 { return true }
  // Rooks and Queens
  if rook_attacks(occ, sq) & (brd.pieces[attacker][ROOK]|brd.pieces[attacker][QUEEN]) > 0 { return true }
  // Kings
  if king_masks[sq] & (brd.pieces[attacker][KING]) > 0 { return true }
  return false
}

// Determines if a piece is blocking a ray attack to its king, and cannot move off this ray
// without placing its king in check.
// 1. Find the displacement vector between the piece at sq and its own king and determine if it
//    lies along a valid ray attack.  If the vector is a valid ray attack:
// 2. Scan toward the king to see if there are any other pieces blocking this route to the king.
// 3. Scan in the opposite direction to see detect any potential threats along this ray.
func is_pinned(brd *BRD, sq, c, e int) bool {
  occ := Occupied(brd)
  var threat, guarded_king BB
   //get direction toward king
  dir := directions[sq][furthest_forward(c, brd.pieces[c][KING])]
  switch dir {
    case NW, NE:
      threat = scan_down(occ, dir+2, sq) & (brd.pieces[e][BISHOP]|brd.pieces[e][QUEEN])
      guarded_king = scan_up(occ, dir, sq) & (brd.pieces[c][KING])
    case SE, SW:
      threat = scan_up(occ, dir-2, sq) & (brd.pieces[e][BISHOP]|brd.pieces[e][QUEEN])
      guarded_king = scan_down(occ, dir, sq) & (brd.pieces[c][KING])
    case NORTH, EAST:
      threat = scan_down(occ, dir+2, sq) & (brd.pieces[e][ROOK]|brd.pieces[e][QUEEN])
      guarded_king = scan_up(occ, dir, sq) & (brd.pieces[c][KING])
    case SOUTH, WEST: 
      threat = scan_up(occ, dir-2, sq) & (brd.pieces[e][ROOK]|brd.pieces[e][QUEEN])
      guarded_king = scan_down(occ, dir, sq) & (brd.pieces[c][KING])
    case INVALID: return false
  }
  return (threat & guarded_king > 0)
}

// The Static Exchange Evaluation (SEE) heuristic provides a way to determine if a capture 
// is a 'winning' or 'losing' capture.
// 1. When a capture results in an exchange of pieces by both sides, SEE is used to determine the 
//    net gain/loss in material for the side initiating the exchange.
// 2. SEE scoring of moves is used for move ordering of captures at critical nodes.
// 3. During quiescence search, SEE is used to prune losing captures. This provides a very low-risk
//    way of reducing the size of the q-search without impacting playing strength.

func get_see(brd *BRD, from, to, c int) int {
  var next_victim, t, last_t int
  temp_color := c^1

  // get initial map of all squares directly attacking this square (does not include 'discovered'/hidden attacks)
  b_attackers := brd.pieces[WHITE][BISHOP] | brd.pieces[BLACK][BISHOP] | 
                 brd.pieces[WHITE][QUEEN]  | brd.pieces[BLACK][QUEEN]
  r_attackers := brd.pieces[WHITE][ROOK]   | brd.pieces[BLACK][ROOK]   | 
                 brd.pieces[WHITE][QUEEN]  | brd.pieces[BLACK][QUEEN]

  temp_map := attack_map(brd, to)
  temp_occ := Occupied(brd)
  var temp_pieces BB

  var piece_list [20]int
  count := 1
  // before entering the main loop, perform each step once for the initial attacking piece.  This ensures that the
  // moved piece is the first to capture.

  piece_list[0] = piece_value(to)

  next_victim = piece_value(from)
  t = piece_type(from)
  clear_sq(from, temp_occ)
  if t != KNIGHT && t != KING { // if the attacker was a pawn, bishop, rook, or queen, re-scan for hidden attacks:
    if t == PAWN || t == BISHOP || t == QUEEN { temp_map |= bishop_attacks(temp_occ, to) & b_attackers }
    if t == PAWN || t == ROOK   || t == QUEEN { temp_map |= rook_attacks(temp_occ, to) & r_attackers }
  }
  last_t = t

  for temp_map &= temp_occ; temp_map>0; temp_map &= temp_occ {
    for t = PAWN; t <= KING; t++ { // loop over piece ts in order of value.
      temp_pieces = brd.pieces[temp_color][t] & temp_map;
      if temp_pieces>0 { break } // stop as soon as a match is found.
    }
    if t > KING { break }

    piece_list[count] = -piece_list[count-1] + next_victim
    next_victim = piece_values[t]

    count++
    if (piece_list[count-1] - next_victim) > 0 { break }

    if last_t == KING { break }

    temp_occ ^= (temp_pieces & -temp_pieces)  // merge the first set bit of temp_pieces into temp_occ
    if t != KNIGHT && t != KING {
      if t == PAWN || t == BISHOP || t == QUEEN { temp_map |= (bishop_attacks(temp_occ, to) & b_attackers) }
      if t == ROOK || t == QUEEN { temp_map |= (rook_attacks(temp_occ, to) & r_attackers) }
    }
    temp_color ^= 1
    last_t = t
  }

  for count-1 > 0 { 
    count--
    piece_list[count-1] = -max(-piece_list[count-1], piece_list[count]) 
  }

  return piece_list[0]
}


func is_in_check(brd *BRD, c, e int) bool {
  if brd.pieces[c][KING] == 0 { return true }
  if is_attacked_by(brd, furthest_forward(c, brd.pieces[c][KING]), e, c) { return true } else { return false }
}

// static VALUE move_evades_check(VALUE self, VALUE p_board, VALUE sq_board, VALUE from, VALUE to, VALUE color){
//   BRD *brd = get_brd(p_board);
//   int c = SYM2COLOR(color);
//   int e = c^1;
//   int f = NUM2INT(from), t = NUM2INT(to);
//   int check;

//   int piece = NUM2INT(rb_ary_entry(sq_board, f));  // ?
//   int captured_piece = NUM2INT(rb_ary_entry(sq_board, t));

//   if(!brd.pieces[c][KING]) return Qfalse;

//   BB delta = (sq_mask_on(t)|sq_mask_on(f));
//   brd.pieces[c][piece_type(piece)] ^= delta;
//   brd.occupied[c] ^= delta;

//   if(captured_piece){
//     clear_sq(t, brd.pieces[e][piece_type(captured_piece)]);
//     clear_sq(t, brd.occupied[e]);
//     // determine if in check
//     check = is_attacked_by(brd, furthest_forward(c, brd.pieces[c][KING]), e, c);
//     add_sq(t, brd.pieces[e][piece_type(captured_piece)]);
//     add_sq(t, brd.occupied[e]);
//   } else {
//     // determine if in check
//     check = is_attacked_by(brd, furthest_forward(c, brd.pieces[c][KING]), e, c);
//   }
//   brd.pieces[c][piece_type(piece)] ^= delta;
//   brd.occupied[c] ^= delta;

//   return (check ? Qfalse : Qtrue);
// }

// // Determines if a move will put the enemy's king in check.
// static VALUE move_gives_check(VALUE self, VALUE p_board, VALUE sq_board, VALUE from, VALUE to, 
//                               VALUE color, VALUE promoted_piece){
//   BRD *brd = get_brd(p_board);
//   int c = SYM2COLOR(color);
//   int e = c^1;
//   int f = NUM2INT(from), t = NUM2INT(to);
//   int check;

//   int piece = NUM2INT(rb_ary_entry(sq_board, f));  // ?
//   int captured_piece = NUM2INT(rb_ary_entry(sq_board, t));

//   if(!brd.pieces[e][KING]) return Qtrue;

//   BB delta = (sq_mask_on(t)|sq_mask_on(f));
//   brd.occupied[c] ^= delta;
//   if(promoted_piece != Qnil){
//     clear_sq(f, brd.pieces[c][piece_type(piece)]);
//     add_sq(t, brd.pieces[c][piece_type(NUM2INT(promoted_piece))]);
//     if(captured_piece){
//       clear_sq(t, brd.pieces[e][piece_type(captured_piece)]);
//       clear_sq(t, brd.occupied[e]);
//       // determine if in check
//       check = is_attacked_by(brd, furthest_forward(e, brd.pieces[e][KING]), c, e);
//       add_sq(t, brd.pieces[e][piece_type(captured_piece)]);
//       add_sq(t, brd.occupied[e]);
//     } else { // determine if in check
//       check = is_attacked_by(brd, furthest_forward(e, brd.pieces[e][KING]), c, e);
//     }
//     add_sq(f, brd.pieces[c][piece_type(piece)]);
//     clear_sq(t, brd.pieces[c][piece_type(NUM2INT(promoted_piece))]);

//   } else {
//     brd.pieces[c][piece_type(piece)] ^= delta;
//     if(captured_piece){
//       clear_sq(t, brd.pieces[e][piece_type(captured_piece)]);
//       clear_sq(t, brd.occupied[e]);
//       // determine if in check
//       check = is_attacked_by(brd, furthest_forward(e, brd.pieces[e][KING]), c, e);
//       add_sq(t, brd.pieces[e][piece_type(captured_piece)]);
//       add_sq(t, brd.occupied[e]);
//     } else { // determine if in check
//       check = is_attacked_by(brd, furthest_forward(e, brd.pieces[e][KING]), c, e);
//     }
//     brd.pieces[c][piece_type(piece)] ^= delta;
//   }
//   brd.occupied[c] ^= delta;

//   return check ? Qtrue : Qfalse;
// }


// static VALUE is_pseudolegal_move_legal(VALUE self, VALUE p_board, VALUE piece, VALUE f, VALUE t, VALUE color){
//   int c = SYM2COLOR(color);
//   int e = c^1;
//   BRD *brd = get_brd(p_board);
//   if(piece_type(NUM2INT(piece)) == KING){ // determine if the to square is attacked by an enemy piece.
//     return is_attacked_by(brd, NUM2INT(t), e, c) ? Qfalse : Qtrue;  // castle moves are pre-checked for legality
//   } else { // determine if the piece being moved is pinned on the king and can't move without putting king at risk.
//     BB pinned = is_pinned(brd, NUM2INT(f), c, e);
//     return pinned && (~pinned & sq_mask_on(NUM2INT(t))) ? Qfalse : Qtrue;
//   }
// }


// static VALUE static_exchange_evaluation(VALUE self, VALUE p_board, VALUE from, VALUE to, VALUE side_to_move, VALUE sq_board){
//   return INT2NUM(get_see(get_brd(p_board), NUM2INT(from), NUM2INT(to), SYM2COLOR(side_to_move), sq_board));
// }


// extern void Init_attack(){
//   VALUE mod_chess = rb_define_module("Chess");
//   VALUE cls_position = rb_define_class_under(mod_chess, "Position", rb_cObject);
//   rb_define_method(cls_position, "side_in_check?", RUBY_METHOD_FUNC(is_in_check), 2);
//   rb_define_method(cls_position, "move_is_legal?", RUBY_METHOD_FUNC(move_evades_check), 5);
//   rb_define_method(cls_position, "move_gives_check?", RUBY_METHOD_FUNC(move_gives_check), 6);
//   rb_define_method(cls_position, "move_avoids_check?", RUBY_METHOD_FUNC(is_pseudolegal_move_legal), 5);

//   VALUE mod_search = rb_define_module_under(mod_chess, "Search");
//   rb_define_module_function(mod_search, "static_exchange_evaluation", static_exchange_evaluation, 5);
// }



