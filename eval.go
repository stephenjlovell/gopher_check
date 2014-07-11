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


var non_king_value, endgame_value, mate_value int

var passed_pawn_bonus = [2][8]int{ { 0, 49, 28, 16, 9,  5,  3,  0 },   
                                   { 0,  3,  5, 9, 16, 16, 28, 49 } }

var promote_row = [2][2]int{ {1, 2}, {6, 5} }


const isolated_pawn_penalty int = -5;
const double_pawn_penalty int  = -10;
const pawn_duo_bonus int       = 3;

var main_pst = [2][5][64]int{
  { // Black
    // Pawn
   {  0,  0,  0,  0,  0,  0,  0,  0, 
     -1,  1,  1,  1,  1,  1,  1, -1, 
     -2,  0,  1,  2,  2,  1,  0, -2, 
     -3, -1,  2, 10, 10,  2, -1, -3, 
     -4, -2,  4, 14, 14,  4, -2, -4, 
     -5, -3,  0,  9,  9,  0, -3, -5, 
     -6, -4,  0,-20,-20,  0, -4, -6, 
      0,  0,  0,  0,  0,  0,  0,  0 },
  // Knight
   { -8, -8, -6, -6, -6, -6, -8, -8, 
     -8,  0,  0,  0,  0,  0,  0, -8, 
     -6,  0,  4,  4,  4,  4,  0, -6, 
     -6,  0,  4,  8,  8,  4,  0, -6, 
     -6,  0,  4,  8,  8,  4,  0, -6, 
     -6,  0,  4,  4,  4,  4,  0, -6, 
     -8,  0,  1,  2,  2,  1,  0, -8, 
    -10,-12, -6, -6, -6, -6,-12,-10 },
  // Bishop
   { -3, -3, -3, -3, -3, -3, -3, -3, 
     -3,  0,  0,  0,  0,  0,  0, -3, 
     -3,  0,  2,  4,  4,  2,  0, -3, 
     -3,  0,  4,  5,  5,  4,  0, -3, 
     -3,  0,  4,  5,  5,  4,  0, -3, 
     -3,  1,  2,  4,  4,  2,  1, -3, 
     -3,  2,  1,  1,  1,  1,  2, -3, 
     -3, -3,-10, -3, -3,-10, -3, -3 },
  // Rook
  {   4,  4,  4,  4,  4,  4,  4,  4,
     16, 16, 16, 16, 16, 16, 16, 16,
     -4,  0,  0,  0,  0,  0,  0, -4,
     -4,  0,  0,  0,  0,  0,  0, -4,
     -4,  0,  0,  0,  0,  0,  0, -4,
     -4,  0,  0,  0,  0,  0,  0, -4,
     -4,  0,  0,  0,  0,  0,  0, -4,
      0,  0,  0,  2,  2,  0,  0,  0 },
  // Queen
   {  0,  0,  0,  1,  1,  0,  0,  0, 
      0,  0,  1,  2,  2,  1,  0,  0, 
      0,  1,  2,  2,  2,  2,  1,  0, 
      0,  1,  2,  3,  3,  2,  1,  0, 
      0,  1,  2,  3,  3,  2,  1,  0, 
      0,  1,  1,  2,  2,  1,  1,  0, 
      0,  0,  1,  1,  1,  1,  0,  0, 
     -6, -6, -6, -6, -6, -6, -6, -6 },
}, // White
{  // Pawn
    { 0,  0,  0,  0,  0,  0,  0,  0, 
     -6, -4,  0,-20,-20,  0, -4, -6, 
     -5, -3,  0,  9,  9,  0, -3, -5,
     -4, -2,  4, 14, 14,  4, -2, -4, 
     -3, -1,  2, 10, 10,  2, -1, -3, 
     -2,  0,  1,  2,  2,  1,  0, -2, 
     -1,  1,  1,  1,  1,  1,  1, -1, 
      0,  0,  0,  0,  0,  0,  0,  0 },
  // Knight
   {-10,-12, -6, -6, -6, -6,-12,-10, 
     -8,  0,  1,  2,  2,  1,  0, -8, 
     -6,  0,  4,  4,  4,  4,  0, -6, 
     -6,  0,  4,  8,  8,  4,  0, -6, 
     -6,  0,  4,  8,  8,  4,  0, -6,
     -6,  0,  4,  4,  4,  4,  0, -6, 
     -8,  0,  0,  0,  0,  0,  0, -8, 
     -8, -8, -6, -6, -6, -6, -8, -8 },
    // Bishop
   { -3, -3,-10, -3, -3,-10, -3, -3, 
     -3,  2,  1,  1,  1,  1,  2, -3, 
     -3,  1,  2,  4,  4,  2,  1, -3, 
     -3,  0,  4,  5,  5,  4,  0, -3, 
     -3,  0,  4,  5,  5,  4,  0, -3, 
     -3,  0,  2,  4,  4,  2,  0, -3, 
     -3,  0,  0,  0,  0,  0,  0, -3, 
     -3, -3, -3, -3, -3, -3, -3, -3 },
    // Rook
   {  0,  0,  0,  2,  2,  0,  0,  0, 
     -4,  0,  0,  0,  0,  0,  0, -4, 
     -4,  0,  0,  0,  0,  0,  0, -4, 
     -4,  0,  0,  0,  0,  0,  0, -4, 
     -4,  0,  0,  0,  0,  0,  0, -4, 
     -4,  0,  0,  0,  0,  0,  0, -4, 
     16, 16, 16, 16, 16, 16, 16, 16, 
      4,  4,  4,  4,  4,  4,  4,  4 },
      // Queen
   { -6, -6, -6, -6, -6, -6, -6, -6,
      0,  0,  1,  1,  1,  1,  0,  0, 
      0,  1,  1,  2,  2,  1,  1,  0, 
      0,  1,  2,  3,  3,  2,  1,  0, 
      0,  1,  2,  3,  3,  2,  1,  0, 
      0,  1,  2,  2,  2,  2,  1,  0, 
      0,  0,  1,  2,  2,  1,  0,  0, 
      0,  0,  0,  1,  1,  0,  0,  0 },
  },
}

var mirror = [64]int{
     56, 57, 58, 59, 60, 61, 62, 63, // Used to create a mirror image of the base PST
     48, 49, 50, 51, 52, 53, 54, 55, // during initialization.
     40, 41, 42, 43, 44, 45, 46, 47,
     32, 33, 34, 35, 36, 37, 38, 39,
     24, 25, 26, 27, 28, 29, 30, 31,
     16, 17, 18, 19, 20, 21, 22, 23,
      8,  9, 10, 11, 12, 13, 14, 15,
      0,  1,  2,  3,  4,  5,  6,  7 }


var king_pst = [2][2][64]int{ { // Black // False
  { -52,-50,-50,-50,-50,-50,-50,-52,   // In early game, encourage the king to stay on back 
    -50,-48,-48,-48,-48,-48,-48,-50,   // row defended by friendly pieces.
    -48,-46,-46,-46,-46,-46,-46,-48,
    -46,-44,-44,-44,-44,-44,-44,-46,
    -44,-42,-42,-42,-42,-42,-42,-44,
    -42,-40,-40,-40,-40,-40,-40,-42,
    -16,-15,-20,-20,-20,-20,-15,-16,
      0, 20, 30,-30,  0,-20, 30, 20 },
    { // True
    -30,-20,-10,  0,  0,-10,-20,-30,     // In end game (when few friendly pieces are available
    -20,-10,  0, 10, 10,  0,-10,-20,     // to protect king), the king should move toward the center
    -10,  0, 10, 20, 20, 10,  0,-10,     // and avoid getting trapped in corners.
      0, 10, 20, 30, 30, 20, 10,  0,
      0, 10, 20, 30, 30, 20, 10,  0,
    -10,  0, 10, 20, 20, 10,  0,-10,
    -20,-10,  0, 10, 10,  0,-10,-20,
    -30,-20,-10,  0,  0,-10,-20,-30 },
  }, { // White // False
   {  0, 20, 30,-30,  0,-20, 30, 20, 
    -16,-15,-20,-20,-20,-20,-15,-16, 
    -42,-40,-40,-40,-40,-40,-40,-42, 
    -44,-42,-42,-42,-42,-42,-42,-44, 
    -46,-44,-44,-44,-44,-44,-44,-46, 
    -48,-46,-46,-46,-46,-46,-46,-48, 
    -50,-48,-48,-48,-48,-48,-48,-50, 
    -52,-50,-50,-50,-50,-50,-50,-52 },
    { // True
    -30,-20,-10,  0,  0,-10,-20,-30, 
    -20,-10,  0, 10, 10,  0,-10,-20, 
    -10,  0, 10, 20, 20, 10,  0,-10, 
      0, 10, 20, 30, 30, 20, 10,  0, 
      0, 10, 20, 30, 30, 20, 10,  0, 
    -10,  0, 10, 20, 20, 10,  0,-10,
    -20,-10,  0, 10, 10,  0,-10,-20, 
    -30,-20,-10,  0,  0,-10,-20,-30 }, }, }

func setup_eval_constants(){
  non_king_value = piece_values[PAWN]*8 + piece_values[KNIGHT]*2 + piece_values[BISHOP]*2 +
                   piece_values[ROOK]*2 + piece_values[QUEEN];
  endgame_value =  piece_values[KING]   - (non_king_value/4);
  mate_value = non_king_value + piece_values[KING];
}

func net_placement(c, e int, brd *BRD) int {
  return adjusted_placement(c, e, brd) - adjusted_placement(c, e, brd)
}



func adjusted_placement(c, e int, brd *BRD) int {
  var ratio float64
  var sq, placement int
  var b BB
  enemy_king_sq := furthest_forward(e, brd.pieces[e][KING])
  
  for t := PAWN; t < KING; t++ {
    for b = brd.pieces[c][t]; b>0; clear_sq(sq, b) {
      sq = furthest_forward(c, b)
      placement += (main_pst[c][t][sq] + tropism_bonus[sq][enemy_king_sq][t])
    }
  }
  for b = brd.pieces[c][KING]; b>0; clear_sq(sq, b){
    sq = furthest_forward(c, b)
    placement += king_pst[c][in_endgame(c, brd)][sq]
  }
  // Base material is incrementally updated as moves are made/unmade.
  return brd.material[c] + placement + mobility(c, e, brd) + pawn_structure(c, e, brd);
}

// // Counts the total possible moves for the given side, not including any target squares defended by enemy pawns.
// static int mobility(int c, int e, BRD *brd){
//   BB friendly = Placement(c);
//   BB available = ~friendly;
//   BB enemy = Placement(e);
//   BB occ = friendly|enemy;
//   BB empty = ~occ;
//   BB unguarded;
//   int sq;
//   int mobility=0;

//   // pawn mobility
//   BB single_advances, double_advances, left_temp, right_temp;
//   if(c){ // white to move
//     single_advances = (brd->pieces[WHITE][PAWN]<<8) & empty; // promotions generated in get_captures
//     double_advances = ((single_advances & row_masks[2])<<8) & empty;
//     left_temp = ((brd->pieces[c][PAWN] & (~column_masks[0]))<<7) & enemy;
//     right_temp = ((brd->pieces[c][PAWN] & (~column_masks[7]))<<9) & enemy;
//     unguarded = ~(((brd->pieces[e][PAWN]&(~column_masks[0]))>>9) | ((brd->pieces[e][PAWN]&(~column_masks[7]))>>7));
//   } else { // black to move
//     single_advances = (brd->pieces[BLACK][PAWN]>>8) & empty;  
//     double_advances = ((single_advances & row_masks[5])>>8) & empty;
//     left_temp = ((brd->pieces[c][PAWN] & (~column_masks[0]))>>9) & enemy;
//     right_temp = ((brd->pieces[c][PAWN] & (~column_masks[7]))>>7) & enemy;
//     unguarded = ~(((brd->pieces[e][PAWN]& (~column_masks[0]))<<7) | ((brd->pieces[e][PAWN]& (~column_masks[7]))<<9));
//   }

//   mobility += (pop_count((single_advances|double_advances) & unguarded) 
//                + pop_count(left_temp & unguarded) + pop_count(right_temp & unguarded));
//   // knight mobility
//   for(BB b = brd->pieces[c][KNIGHT]; b; clear_sq(sq, b)){
//     sq = furthest_forward(c, b);
//     mobility += pop_count(knight_masks[sq] & available & unguarded);
//   }
//   // bishop mobility
//   for(BB b = brd->pieces[c][BISHOP]; b; clear_sq(sq, b)){
//     sq = furthest_forward(c, b);
//     mobility += pop_count(bishop_attacks(occ, sq) & available & unguarded);
//   }
//   // rook mobility
//   for(BB b = brd->pieces[c][ROOK]; b; clear_sq(sq, b)){
//     sq = furthest_forward(c, b);
//     mobility += pop_count(rook_attacks(occ, sq) & available & unguarded);
//   }
//   // queen mobility
//   for(BB b = brd->pieces[c][QUEEN]; b; clear_sq(sq, b)){
//     sq = furthest_forward(c, b);
//     mobility += pop_count(queen_attacks(occ, sq) & available & unguarded);
//   }
//   // king mobility
//   for(BB b = brd->pieces[c][KING]; b; clear_sq(sq, b)){
//     sq = furthest_forward(c, b);
//     mobility += pop_count(king_masks[sq] & available & unguarded);
//   }
//   return mobility;
// }

// // PAWN EVALUATION
// // 
// // Good structures:
// //   -Passed pawns - Bonus for pawns unblocked by an enemy pawn on the same or adjacent file. 
// //                   May eventually get promoted.
// //   -Pawn duos - Pawns side by side to another friendly pawn receive a small bonus
// // 
// // Bad structures:
// //   -Isolated pawns - Penalty for any pawn without friendly pawns on adjacent files.  
// //   -Double/tripled pawns - Penalty for having multiple pawns on the same file.
// static int pawn_structure(int c, int e, BRD *brd){
//   int structure = 0;
//   int sq;
//   BB own_pawns = brd->pieces[c][PAWN];
//   BB enemy_pawns = brd->pieces[e][PAWN];

//   for(BB b = own_pawns; b; clear_sq(sq, b)){
//     sq = furthest_forward(c, b);
//     // passed pawns
//     if(!(pawn_passed_masks[c][sq] & enemy_pawns)) {
//       structure += passed_pawn_bonus[c][row(sq)];        
//       if(row(sq) == promote_row[c][0]){
//         if(!is_attacked_by(brd, (c ? sq+8 : sq-8), c^1, c)){
//           structure += passed_pawn_bonus[c][row(sq)];  // double the value of the bonus if path to promotion is undefended.          
//         }
//       } else if(row(sq) == promote_row[c][1]) {
//         if(!is_attacked_by(brd, (c ? sq+8 : sq-8), c^1, c) && 
//            !is_attacked_by(brd, (c ? sq+16 : sq-16), c^1, c)){
//           structure += passed_pawn_bonus[c][row(sq)];  // double the value of the bonus if path to promotion is undefended.
//         }
//       }
//     }
//     // isolated pawns
//     if(!(pawn_isolated_masks[sq] & own_pawns)) structure += isolated_pawn_penalty;
//     // pawn duos 
//     if(pawn_side_masks[sq] & own_pawns) structure += pawn_duo_bonus;
//   }
//   int column_count;
//   for(int i=0; i<8; i++){
//     // doubled/tripled pawns
//     column_count = pop_count(column_masks[i] & own_pawns);
//     if (column_count > 1){
//       structure += (double_pawn_penalty<<(column_count-2));
//     }
//   }
//   return structure;
// }



// static VALUE net_material(VALUE self, VALUE pc_board, VALUE color){
//   BRD *brd = get_brd(pc_board);  
//   int c = SYM2COLOR(color);
//   int e = c^1;
//   int sq, placement = 0;
//   BB b;
//   return INT2NUM(adjusted_material(c, brd)-adjusted_material(e, brd));
// }

// static int adjusted_material(int c, BRD *brd){
//   int sq, placement = 0;
//   BB b;

//   for(int t = PAWN; t < QUEEN; t++){
//     for(b = brd->pieces[c][t]; b; clear_sq(sq, b)){
//       sq = furthest_forward(c, b);
//       placement += main_pst[c][t][sq];
//     }
//   }
//   for(b = brd->pieces[c][KING]; b; clear_sq(sq, b)){
//     sq = furthest_forward(c, b);
//     placement += king_pst[c][in_endgame(c)][sq];
//   }
//   return brd->material[c] + placement;
// }


// static VALUE evaluate_material(VALUE self, VALUE pc_board, VALUE color){
//   BRD *brd = get_brd(pc_board);  
//   int c = SYM2COLOR(color);
//   int sq, placement = 0;
//   BB b;

//   for(int t = PAWN; t < QUEEN; t++){
//     b = brd->pieces[c][t];
//     for(b = brd->pieces[c][t]; b; clear_sq(sq, b)){
//       sq = furthest_forward(c, b);
//       placement += main_pst[c][t][sq];
//     }
//   }
//   for(b = brd->pieces[c][KING]; b; clear_sq(sq, b)){
//     sq = furthest_forward(c, b);
//     placement += king_pst[c][in_endgame(c)][sq];
//   }
//   return INT2NUM(brd->material[c] + placement);
// }

// extern void Init_eval(){
//   printf("  -Loading eval extension...");
//   setup_eval_constants();

//   VALUE mod_chess = rb_define_module("Chess");
//   VALUE mod_eval = rb_define_module_under(mod_chess, "Evaluation");

//   rb_define_module_function(mod_eval, "evaluate_material", evaluate_material, 2);
//   rb_define_module_function(mod_eval, "net_material", net_material, 2);
//   rb_define_module_function(mod_eval, "net_placement", net_placement, 2);

//   printf("done.\n");
// }

func in_endgame(c int, brd *BRD) bool { if brd.material[c] <= endgame_value { return 1 } else { return 0 } }

// #define in_endgame(color) (cBoard->material[color] <= endgame_value ? 1 : 0)


