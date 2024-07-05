package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type SlideCheckData struct {
	fenString     string
	piecePos      int8
	moves         []int8
	capturesMoves []int8
}

func TestPosition_SliderPseudoLegalMovesRookNoCastle(t *testing.T) {
	data := []SlideCheckData{
		{
			fenString: "8/8/8/8/8/8/8/R7 w - - 0 1",
			piecePos:  A1,
			moves:     []int8{A2, A3, A4, A5, A6, A7, A8, B1, C1, D1, E1, F1, G1, H1},
		},
		{
			fenString: "8/8/4P3/8/3PR1P1/4P3/8/8 w - - 0 1",
			piecePos:  E4,
			moves:     []int8{F4, E5},
		},
		{
			fenString:     "8/8/4P3/8/3PR1p1/4P3/8/8 w - - 0 1",
			piecePos:      E4,
			moves:         []int8{F4, G4, E5},
			capturesMoves: []int8{G4},
		},
		{
			fenString:     "rn2kbnr/1pp1pppp/8/p7/R5bq/8/P1PPPPPP/1NBQKBNR w - - 0 1",
			piecePos:      A4,
			moves:         []int8{A3, A5, B4, C4, D4, E4, F4, G4},
			capturesMoves: []int8{A5, G4},
		},
	}

	for _, d := range data {
		pos, err := NewPositionFromFEN(d.fenString)
		if err != nil {
			t.Fatal(err.Error())
		}

		moves, capturesMoves := pos.SliderPseudoLegalMoves(d.piecePos)
		assert.ElementsMatch(t, d.moves, moves)
		assert.ElementsMatch(t, d.capturesMoves, capturesMoves)
	}
}
