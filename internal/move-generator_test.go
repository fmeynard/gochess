package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type CheckData map[string]struct {
	fenString     string
	piecePos      int8
	moves         []int8
	capturesMoves []int8
}

type generatorFn func(p Position, pieceIdx int8) (moves []int8, capturesMoves []int8)

func TestPosition_SliderPseudoLegalMovesRookNoCastle(t *testing.T) {
	data := CheckData{
		"A1 - Empty board": {
			fenString: "8/8/8/8/8/8/8/R7 w - - 0 1",
			piecePos:  A1,
			moves:     []int8{A2, A3, A4, A5, A6, A7, A8, B1, C1, D1, E1, F1, G1, H1},
		},
		"E4 - Blocked left&down - 1 up&right- No captures": {
			fenString: "8/8/4P3/8/3PR1P1/4P3/8/8 w - - 0 1",
			piecePos:  E4,
			moves:     []int8{F4, E5},
		},
		"E4 - Blocked left&down - 1 up- Capture G4": {
			fenString:     "8/8/4P3/8/3PR1p1/4P3/8/8 w - - 0 1",
			piecePos:      E4,
			moves:         []int8{F4, G4, E5},
			capturesMoves: []int8{G4},
		},
		"E4 - 1 down - Capture A5,G4": {
			fenString:     "rn2kbnr/1pp1pppp/8/p7/R5bq/8/P1PPPPPP/1NBQKBNR w - - 0 1",
			piecePos:      A4,
			moves:         []int8{A3, A5, B4, C4, D4, E4, F4, G4},
			capturesMoves: []int8{A5, G4},
		},
	}

	execMovesCheck(t, data, SliderPseudoLegalMoves)
}

func TestPosition_KnightPseudoLegalMoves(t *testing.T) {
	data := CheckData{
		"All Directions - No Captures": {
			fenString: "8/8/8/8/4N3/8/8/8 w - - 0 1",
			piecePos:  E4,
			moves:     []int8{11, 13, 18, 22, 34, 38, 43, 45},
		},
		"Captures All Directions": {
			fenString:     "8/8/3P1P2/2P3P1/4n3/2P3P1/3P1P2/8 w - - 0 1",
			piecePos:      E4,
			capturesMoves: []int8{C3, C5, D2, D6, F2, F6, G3, G5},
			moves:         []int8{C3, C5, D2, D6, F2, F6, G3, G5},
		},
		"B1 : Edges check": {
			fenString: "8/8/8/8/8/8/8/1N6 w - - 0 1",
			piecePos:  B1,
			moves:     []int8{A3, C3, D2},
		},
		"G8 : Edges check": {
			fenString: "8/8/8/8/4N3/8/8/8 w - - 0 1",
			piecePos:  G8,
			moves:     []int8{E7, H6, F6},
		},
	}

	execMovesCheck(t, data, KnightPseudoLegalMoves)
}

func Test_KingPseudoLegalMoves(t *testing.T) {
	data := CheckData{
		"Start pos no moves white": {
			fenString: FenStartPos,
			piecePos:  E1,
		},
		"All directions white - no capture": {
			fenString: "8/8/8/8/8/8/2k5/8 w - - 0 1",
			piecePos:  C2,
			moves:     []int8{B1, C1, D1, B2, D2, B3, C3, D3},
		},
		"All directions capture": {
			fenString:     "8/1QRB4/1PkP4/1BRQ4/8/8/8/8 b - - 0 1",
			piecePos:      C6,
			moves:         []int8{B7, C7, D7, B6, D6, B5, C5, D5},
			capturesMoves: []int8{B7, C7, D7, B6, D6, B5, C5, D5},
		},
		"All directions no capture same color": {
			fenString: "8/1QRB4/1PKP4/1BRQ4/8/8/8/8 b - - 0 1",
			piecePos:  C6,
		},
		"QueenSide castle white": {
			fenString: "rnbqkbnr/pppppppp/8/8/8/B1NP4/PPPQPPPP/R3KBNR b KQkq - 0 1",
			piecePos:  E1,
			moves:     []int8{C1, D1},
		},
		"QueenSide castle white - Only kingSide allowed": {
			fenString: "rnbqkbnr/pppppppp/8/8/8/B1NP4/PPPQPPPP/R3KBNR b Kkq - 0 1",
			piecePos:  E1,
			moves:     []int8{D1},
		},
		"QueenSide castle white - no rights": {
			fenString: "rnbqkbnr/pppppppp/8/8/8/B1NP4/PPPQPPPP/R3KBNR b kq - 0 1",
			piecePos:  E1,
			moves:     []int8{D1},
		},
		"KingSide castle black": {
			fenString: "rnbqk2r/ppppnppp/3bp3/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			piecePos:  E8,
			moves:     []int8{F8, G8},
		},
		"KingSide castle black - Only queen allowed": {
			fenString: "rnbqk2r/ppppnppp/3bp3/8/8/8/PPPPPPPP/RNBQKBNR w KQq - 0 1",
			piecePos:  E8,
			moves:     []int8{F8},
		},
		"KingSide castle black - ro rights": {
			fenString: "rnbqk2r/ppppnppp/3bp3/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
			piecePos:  E8,
			moves:     []int8{F8},
		},
	}

	execMovesCheck(t, data, KingPseudoLegalMoves)
}

func Test_PawnPseudoLegalMoves(t *testing.T) {
	data := CheckData{
		"start pos white": {
			fenString: "8/8/8/8/8/8/P7/8 w - - 0 1",
			piecePos:  A2,
			moves:     []int8{A3, A4},
		},
		"start pos black": {
			fenString: "8/p7/8/8/8/8/8/8 w - - 0 1",
			piecePos:  A7,
			moves:     []int8{A6, A5},
		},
		"no capture left white": {
			fenString:     "8/8/8/pp5p/P7/8/8/8 w - - 0 1",
			piecePos:      A4,
			moves:         []int8{B5},
			capturesMoves: []int8{B5},
		},
		"no capture left black": {
			fenString:     "8/8/8/p7/1P5P/8/8/8 b - - 0 1",
			piecePos:      A5,
			moves:         []int8{A4, B4},
			capturesMoves: []int8{B4},
		},
		"both capture white": {
			fenString:     "8/8/8/8/8/ppp4p/1P6/8 b - - 0 1",
			piecePos:      B2,
			moves:         []int8{A3, C3},
			capturesMoves: []int8{A3, C3},
		},
		"both capture black": {
			fenString:     "8/8/8/8/3PPP2/3PpP2/PPPPPPPP/8 w - - 0 1",
			piecePos:      E3,
			moves:         []int8{D2, F2},
			capturesMoves: []int8{D2, F2},
		},
		"enPassant left white": {
			fenString:     "8/8/1p6/pPp5/8/8/8/8 w - a6 0 1",
			piecePos:      B5,
			moves:         []int8{A6},
			capturesMoves: []int8{A6},
		},
		"enPassant right white": {
			fenString:     "8/8/1p6/pPp5/8/8/8/8 w - c6 0 1",
			piecePos:      B5,
			moves:         []int8{C6},
			capturesMoves: []int8{C6},
		},
		"no enPassant white": {
			fenString: "8/8/1p6/pPp5/8/8/8/8 w - - 0 1",
			piecePos:  B5,
		},
		"enPassant left black": {
			fenString:     "8/8/8/8/PpP5/1P6/8/8 b - a3 0 1",
			piecePos:      B4,
			moves:         []int8{A3},
			capturesMoves: []int8{A3},
		},
		"No-EnPassant cross-board black": {
			fenString: "8/8/8/8/PpP4p/1P6/8/8 b - a3 0 1",
			piecePos:  H4,
			moves:     []int8{H3},
		},
	}

	execMovesCheck(t, data, PawnPseudoLegalMoves)
}

func execMovesCheck(t *testing.T, data CheckData, generator generatorFn) {
	for tName, d := range data {
		t.Run(tName, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fenString)
			if err != nil {
				t.Fatal(err.Error())
			}

			moves, capturesMoves := generator(pos, d.piecePos)
			assert.ElementsMatch(t, d.moves, moves, "Moves do not match")
			assert.ElementsMatch(t, d.capturesMoves, capturesMoves, "Captures moves do not match")
		})
	}
}
