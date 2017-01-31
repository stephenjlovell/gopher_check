# GopherCheck

An open-source, UCI chess engine written in Go!

GopherCheck supports a subset of the Universal Chess Interface (UCI) protocol. To use GopherCheck, you'll need a UCI-compatible chess GUI such as [Arena Chess](http://www.playwitharena.com/ "Arena Chess") or [Scid vs. PC](http://scidvspc.sourceforge.net/ "Scid vs. PC").

## Installation

Binaries are available for Windows and Mac. You can get the latest stable release from the [releases page](https://github.com/stephenjlovell/gopher_check/releases).

To compile from source, you'll need the [latest version of Go](https://golang.org/doc/install). Once you've set up your Go workspace, run [go get](https://golang.org/cmd/go/#hdr-Download_and_install_packages_and_dependencies) to download and install GopherCheck:

    $ go get -u github.com/stephenjlovell/gopher_check

## Usage

```
$ gopher_check --help
  Usage of gopher_check:
    -cpuprofile
      	Runs cpu profiler on test suite.
    -memprofile
      	Runs memory profiler on test suite.
    -version
      	Prints version number and exits.
```
Starting GopherCheck without any arguments will start the engine in UCI (command-line) mode:
```
$ gopher_check
  Magics read from disk.

$ uci
  id name GopherCheck 0.2.0
  id author Steve Lovell
  option name Ponder type check default false
  option name CPU type spin default 0 min 1 max 4
  uciok

$ position startpos
  readyok

$ print
  Side to move: WHITE
      A   B   C   D   E   F   G   H
    ---------------------------------
  8 | ♜ | ♞ | ♝ | ♛ | ♚ | ♝ | ♞ | ♜ |
    ---------------------------------
  7 | ♟ | ♟ | ♟ | ♟ | ♟ | ♟ | ♟ | ♟ |
    ---------------------------------
  6 |   |   |   |   |   |   |   |   |
    ---------------------------------
  5 |   |   |   |   |   |   |   |   |
    ---------------------------------
  4 |   |   |   |   |   |   |   |   |
    ---------------------------------
  3 |   |   |   |   |   |   |   |   |
    ---------------------------------
  2 | ♙ | ♙ | ♙ | ♙ | ♙ | ♙ | ♙ | ♙ |
    ---------------------------------
  1 | ♖ | ♘ | ♗ | ♕ | ♔ | ♗ | ♘ | ♖ |
    ---------------------------------
      A   B   C   D   E   F   G   H

$ go movetime 2000
  info score cp 16 depth 7 nodes 9531 nps 1188663 time 8 pv e2e4 d7d5 e4e5 c8f5 d2d4 e7e6 b1c3
  info score cp 13 depth 8 nodes 55569 nps 1173934 time 47 pv d2d4 d7d5 c1f4 e7e6 e2e3 b8c6 b1c3 b7b6
  info score cp 19 depth 9 nodes 129949 nps 1282122 time 101 pv e2e4 e7e6 d2d4 d7d5 e4e5 d5e4 b1c3 d8g5 c1g5
  info score cp 12 depth 10 nodes 213058 nps 1194635 time 178 pv e2e4 e7e6 d2d4 d7d5 e4e5 c7c5 c2c3 d8a5 g2g3 c5d4
  info score cp 18 depth 11 nodes 917781 nps 1528904 time 600 pv e2e4 e7e6 f2f4 d7d5 e4e5 c7c5 d2d4 d8b6 b1c3 f8b4 f1a6
  info score cp 11 depth 12 nodes 2394738 nps 1661600 time 1441 pv e2e4 e7e6 f2f4 d7d5 e4e5 c7c5 g2g3 b8c6 c2c3 d8b6 b2b3 g7g6
  bestmove e2e4 ponder e7e6

$ quit
```
## Search Features

GopherCheck supports [parallel search](https://chessprogramming.wikispaces.com/Parallel+Search "Parallel Search"), defaulting to one search process (goroutine) per logical core. You can set the number of search goroutines via the options panel in your GUI, or by using ```setoption name CPU value <number of goroutines>``` when in command-line mode.

GopherCheck uses a version of iterative deepening, nega-max search known as [Principal Variation Search (PVS)](https://chessprogramming.wikispaces.com/Principal+Variation+Search "Principal Variation Search"). Notable search features include:

- Shared hash table
- Young-brothers wait concept (YBWC)
- Null-move pruning with verification search
- Mate-distance pruning
- Internal iterative deepening (IID)
- Search extensions:
  - Singular extensions
  - Check extensions
  - Promotion extensions
- Search reductions:
  - Late-move reductions  
- Pruning:
  - Futility pruning

## Evaluation Features

Evaluation in GopherCheck is symmetric: values for each heuristic are calculated for both sides, and a net score is returned for the current side to move.  GopherCheck uses the following evaluation heuristics:

- Material balance - material is a simple sum of the value of each non-king piece in play. This is the largest evaluation factor.
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
      - they cannot be defended by friendly pawns (no friendly pawn can move up to defend them),
      - their stop square is defended by an enemy sentry pawn,
      - their stop square is not defended by a friendly pawn
- Pawn hash table - Evaluation features that depend only on the location of each side's pawns are cached in a special pawn hash table.

## Contributing

Pull requests are welcome! To contribute to GopherCheck, you'll need to do the following:

1. Make sure you have [Go (>= 1.7.0)](https://golang.org/doc/install) installed.
- Install a UCI-compatible chess GUI such as [Arena Chess](http://www.playwitharena.com/ "Arena Chess") or [Scid vs. PC](http://scidvspc.sourceforge.net/ "Scid vs. PC").
- Fork this repo.
- Run ```go install``` and ```gopher_check --version``` to ensure GopherCheck installed correctly.
- Hack on your changes.
- Run tests frequently to make sure everything is still working:
  - Run ```go test -run=TestPlayingStrength``` to benchmark GopherCheck's performance on your hardware. This takes about 10 minutes.
  - Use your chess GUI to pit GopherCheck against other engines, or against older versions of GopherCheck.
- Document the reasoning behind your changes along with any test results in your pull request.

## License

GopherCheck is available under the MIT License.
