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
