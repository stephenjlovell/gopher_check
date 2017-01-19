//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

const ( // TODO: expose these options via UCI interface.
	LAZY_EVAL_MARGIN = BISHOP_VALUE
	TEMPO_BONUS      = 5
)

const (
	MAX_ENDGAME_COUNT = 24
)

const (
	MIDGAME = iota
	ENDGAME
)

var chebyshev_distance_table [64][64]int

func chebyshev_distance(from, to int) int {
	return chebyshev_distance_table[from][to]
}

func setup_chebyshev_distance() {
	for from := 0; from < 64; from++ {
		for to := 0; to < 64; to++ {
			chebyshev_distance_table[from][to] = max(abs(row(from)-row(to)), abs(column(from)-column(to)))
		}
	}
}

// 0 indicates endgame.  Initial position phase is 24.  Maximum possible is 48.
var endgame_phase [64]int

// piece values used to determine endgame status. 0-12 per side,
var endgame_count_values = [8]uint8{0, 1, 1, 2, 4, 0}

var main_pst = [2][8][64]int{ // Black. White PST will be set in setup_eval.
	{ // Pawn
		{0, 0, 0, 0, 0, 0, 0, 0,
			-11, 1, 1, 1, 1, 1, 1, -11,
			-12, 0, 1, 2, 2, 1, 0, -12,
			-13, -1, 2, 10, 10, 2, -1, -13,
			-14, -2, 4, 14, 14, 4, -2, -14,
			-15, -3, 0, 9, 9, 0, -3, -15,
			-16, -4, 0, -20, -20, 0, -4, -16,
			0, 0, 0, 0, 0, 0, 0, 0},

		// Knight
		{-8, -8, -6, -6, -6, -6, -8, -8,
			-8, 0, 0, 0, 0, 0, 0, -8,
			-6, 0, 4, 4, 4, 4, 0, -6,
			-6, 0, 4, 8, 8, 4, 0, -6,
			-6, 0, 4, 8, 8, 4, 0, -6,
			-6, 0, 4, 4, 4, 4, 0, -6,
			-8, 0, 1, 2, 2, 1, 0, -8,
			-10, -12, -6, -6, -6, -6, -12, -10},
		// Bishop
		{-3, -3, -3, -3, -3, -3, -3, -3,
			-3, 0, 0, 0, 0, 0, 0, -3,
			-3, 0, 2, 4, 4, 2, 0, -3,
			-3, 0, 4, 5, 5, 4, 0, -3,
			-3, 0, 4, 5, 5, 4, 0, -3,
			-3, 1, 2, 4, 4, 2, 1, -3,
			-3, 2, 1, 1, 1, 1, 2, -3,
			-3, -3, -10, -3, -3, -10, -3, -3},
		// Rook
		{4, 4, 4, 4, 4, 4, 4, 4,
			16, 16, 16, 16, 16, 16, 16, 16,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			0, 0, 0, 2, 2, 0, 0, 0},
		// Queen
		{0, 0, 0, 1, 1, 0, 0, 0,
			0, 0, 1, 2, 2, 1, 0, 0,
			0, 1, 2, 2, 2, 2, 1, 0,
			0, 1, 2, 3, 3, 2, 1, 0,
			0, 1, 2, 3, 3, 2, 1, 0,
			0, 1, 1, 2, 2, 1, 1, 0,
			0, 0, 1, 1, 1, 1, 0, 0,
			-6, -6, -6, -6, -6, -6, -6, -6},
	},
}

var king_pst = [2][2][64]int{ // Black
	{ // Early game
		{
			-52, -50, -50, -50, -50, -50, -50, -52, // In early game, encourage the king to stay on back
			-50, -48, -48, -48, -48, -48, -48, -50, // row defended by friendly pieces.
			-48, -46, -46, -46, -46, -46, -46, -48,
			-46, -44, -44, -44, -44, -44, -44, -46,
			-44, -42, -42, -42, -42, -42, -42, -44,
			-42, -40, -40, -40, -40, -40, -40, -42,
			-16, -15, -20, -20, -20, -20, -15, -16,
			0, 20, 30, -30, 0, -20, 30, 20,
		},
		{ // Endgame
			-30, -20, -10, 0, 0, -10, -20, -30, // In end game (when few friendly pieces are available
			-20, -10, 0, 10, 10, 0, -10, -20, // to protect king), the king should move toward the center
			-10, 0, 10, 20, 20, 10, 0, -10, // and avoid getting trapped in corners.
			0, 10, 20, 30, 30, 20, 10, 0,
			0, 10, 20, 30, 30, 20, 10, 0,
			-10, 0, 10, 20, 20, 10, 0, -10,
			-20, -10, 0, 10, 10, 0, -10, -20,
			-30, -20, -10, 0, 0, -10, -20, -30,
		},
	},
}

var square_mirror = [64]int{
	H1, H2, H3, H4, H5, H6, H7, H8,
	G1, G2, G3, G4, G5, G6, G7, G8,
	F1, F2, F3, F4, F5, F6, F7, F8,
	E1, E2, E3, E4, E5, E6, E7, E8,
	D1, D2, D3, D4, D5, D6, D7, D8,
	C1, C2, C3, C4, C5, C6, C7, C8,
	B1, B2, B3, B4, B5, B6, B7, B8,
	A1, A2, A3, A4, A5, A6, A7, A8,
}

var king_threat_bonus = [64]int{
	0, 2, 3, 5, 9, 15, 24, 37,
	55, 79, 111, 150, 195, 244, 293, 337,
	370, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
}

var king_saftey_base = [2][64]int{
	{ // Black
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 3, 3, 3, 3, 3, 3, 4,
		3, 1, 1, 1, 1, 1, 1, 3,
		2, 0, 0, 0, 0, 0, 0, 2,
	},
}

// adjusts value of knights and rooks based on number of own pawns in play.
var knight_pawns = [16]int{-20, -16, -12, -8, -4, 0, 4, 8, 12}
var rook_pawns = [16]int{16, 12, 8, 4, 2, 0, -2, -4, -8}

// adjusts the value of bishop pairs based on number of enemy pawns in play.
var bishop_pair_pawns = [16]int{10, 10, 9, 8, 6, 4, 2, 0, -2}

var knight_mobility = [16]int{-16, -12, -6, -3, 0, 1, 3, 5, 6, 0, 0, 0, 0, 0, 0}

var bishop_mobility = [16]int{-24, -16, -8, -4, -2, 0, 2, 4, 6, 7, 8, 9, 10, 11, 12, 13}

var rook_mobility = [16]int{-12, -8, -4, -2, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

var queen_mobility = [32]int{-24, -18, -12, -6, -3, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 24, 24, 24}

var queen_tropism_bonus = [8]int{0, 12, 9, 6, 3, 0, -3, -6}

func evaluate(brd *Board, alpha, beta int) int {
	c, e := brd.c, brd.Enemy()
	// lazy evaluation: if material balance is already outside the search window by an amount that outweighs
	// the largest likely placement evaluation, return the material as an approximate evaluation.
	// This prevents the engine from wasting a lot of time evaluating unrealistic positions.
	score := int(brd.material[c]-brd.material[e]) + TEMPO_BONUS
	if score+LAZY_EVAL_MARGIN < alpha || score-LAZY_EVAL_MARGIN > beta {
		return score
	}

	pentry := brd.worker.ptt.Probe(brd.pawn_hash_key)
	if pentry.key != brd.pawn_hash_key { // pawn hash table miss.
		// collisions can occur, but are too infrequent to matter much (1 / 20+ million)
		set_pawn_structure(brd, pentry) // evaluate pawn structure and save to pentry.
	}

	score += net_pawn_placement(brd, pentry, c, e)
	score += net_major_placement(brd, pentry, c, e) // 3x as expensive as pawn eval...

	return score
}

func net_major_placement(brd *Board, pentry *PawnEntry, c, e uint8) int {
	king_sq, enemy_king_sq := brd.KingSq(c), brd.KingSq(e)
	return major_placement(brd, pentry, c, e, king_sq, enemy_king_sq) -
		major_placement(brd, pentry, e, c, enemy_king_sq, king_sq)
}

var pawn_shield_bonus = [4]int{-9, -3, 3, 9}

func major_placement(brd *Board, pentry *PawnEntry, c, e uint8, king_sq,
	enemy_king_sq int) (total_placement int) {

	friendly := brd.Placement(c)
	occ := brd.AllOccupied()

	available := (^friendly) & (^(pentry.all_attacks[e]))

	var sq, mobility, placement, king_threats int
	var b, attacks BB

	enemy_king_zone := king_zone_masks[e][enemy_king_sq]

	pawn_count := pentry.count[c]

	for b = brd.pieces[c][KNIGHT]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += knight_pawns[pawn_count]
		attacks = knight_masks[sq] & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += knight_mobility[pop_count(attacks)]
	}

	for b = brd.pieces[c][BISHOP]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		attacks = bishop_attacks(occ, sq) & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += bishop_mobility[pop_count(attacks)]
	}
	if pop_count(brd.pieces[c][BISHOP]) > 1 { // bishop pairs
		placement += 40 + bishop_pair_pawns[pentry.count[e]]
	}

	phase := endgame_phase[brd.endgame_counter]

	for b = brd.pieces[c][ROOK]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += rook_pawns[pawn_count]
		attacks = rook_attacks(occ, sq) & available
		king_threats += pop_count(attacks & enemy_king_zone)
		// only reward rook mobility in the late-game.
		mobility += weight_score(phase, 0, rook_mobility[pop_count(attacks)])
	}

	for b = brd.pieces[c][QUEEN]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		attacks = queen_attacks(occ, sq) & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += queen_mobility[pop_count(attacks)]
		placement += weight_score(phase, 0, // encourage queen to move toward enemy king in the late-game.
			queen_tropism_bonus[chebyshev_distance(sq, enemy_king_sq)])
	}

	placement += weight_score(phase,
		pawn_shield_bonus[pop_count(brd.pieces[c][PAWN]&king_shield_masks[c][king_sq])], 0)

	placement += weight_score(phase,
		king_pst[c][MIDGAME][king_sq], king_pst[c][ENDGAME][king_sq])

	placement += weight_score(phase,
		king_threat_bonus[king_threats+king_saftey_base[e][enemy_king_sq]], 0)

	return placement + mobility
}

// Tapered Evaluation: adjust the score based on how close we are to the endgame.
// This prevents 'evaluation discontinuity' where the score changes significantly when moving from
// mid-game to end-game, causing the search to chase after changes in endgame status instead of real
// positional gain.
// Uses the scaling function first implemented in Fruit and described here:
// https://chessprogramming.wikispaces.com/Tapered+Eval
func weight_score(phase, mg_score, eg_score int) int {
	return ((mg_score * (256 - phase)) + (eg_score * phase)) / 256
}

func setup_eval() {
	for piece := PAWN; piece < KING; piece++ { // Main PST
		for sq := 0; sq < 64; sq++ {
			main_pst[WHITE][piece][sq] = main_pst[BLACK][piece][square_mirror[sq]]
		}
	}
	for endgame := 0; endgame < 2; endgame++ { // King PST
		for sq := 0; sq < 64; sq++ {
			king_pst[WHITE][endgame][sq] = king_pst[BLACK][endgame][square_mirror[sq]]
		}
	}
	for sq := 0; sq < 64; sq++ { // King saftey counters
		king_saftey_base[WHITE][sq] = king_saftey_base[BLACK][square_mirror[sq]]
	}
	for i := 0; i <= 24; i++ { // Endgame phase scaling factor
		endgame_phase[i] = (((MAX_ENDGAME_COUNT - i) * 256) + (MAX_ENDGAME_COUNT / 2)) / MAX_ENDGAME_COUNT
	}
}
