# GopherCheck

An open-source, UCI chess engine written in Go!

GopherCheck currently supports a subset of the Universal Chess Interface (UCI) protocol. To use GopherCheck, you'll need a UCI-compatible chess GUI.

## Search Features



## Evaluation Features

Evaluation in GopherCheck is symmetric: values for each heuristic are calculated for both sides, and a net score is returned for the current side to move.  GopherCheck uses the following evaluation heuristics:

1. Material balance - material is a simple sum of the value of each non-king piece in play.

- Lazy evaluation - if the material balance is well outside the search window, evaluation is cut short and returns the material balance. This prevents the engine from wasting a lot of time evaluating unrealistic positions.

- Piece-square tables - Small static bonuses/penalties are applied based on the type of piece and its location on the board.

- Mobility - major pieces are awarded bonuses based on the type of piece and the available moves from its current location (excluding squares guarded by enemy pawns).  GopherCheck will generally prefer to position its major pieces where they can control the largest amount of space on the board.

- King safety - Each side receives a scaled bonus for the number of attacks it can make into squares adjacent to the enemy king.

- Tapered evaluation - Some heuristics are adjusted based on how close we are to the endgame. This prevents 'evaluation discontinuity' where the score changes significantly when moving from mid-game to end-game, causing the search to chase after changes in endgame status instead of real positional gain.

- Pawn structure - Pawn values are adjusted by looking for several structures considered in chess to be particularly strong/weak.
    - Passed pawns - If no enemy pawns can block a pawn's advance, it is considered 'passed' and is more likely to eventually get promoted.  A bonus is awarded for each passed pawn based on how close it is to promotion.
    - Defended/chained pawns - Pawns that are defended by at least one other pawn are awarded a bonus.
    - Isolated pawns - Pawns that are separated from other friendly pawns are vulnerable and may tie down valuable resources in defending them. A small penalty is given for each isolated pawn.
    - Pawn duos - Pawns that are side by side to one another create an interlocking wall of defended squares. A small bonus is given to each pawn that has at least one other pawn directly to its left or right.
    - Doubled/tripled pawns - A penalty is given for each pawn on the same file (column) as another friendly pawn. Having multiple pawns on the same file (column) limits their ability to advance, as they can easily be blocked by a single enemy piece and cannot defend one another.
    - Backward pawns - A small penalty is given to backward pawns, i.e.:
      - they cannot be defended by friendly pawns,
      - their stop square is defended by an enemy sentry pawn,
      - their stop square is not defended by a friendly pawn

- Pawn hash table - Evaluation features that depend only on the location of each side's pawns are cached in a special pawn hash table.
