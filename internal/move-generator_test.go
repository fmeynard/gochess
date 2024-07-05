package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type SlideCheckData struct {
	fenString     string
	piecePos      string
	moves         []int
	capturesMoves []int
}

func TestPosition_SliderPseudoLegalMovesRookNoCastle(t *testing.T) {
	data := []SlideCheckData{
		{
			fenString: "8/8/8/8/8/8/8/R7 w - - 0 1",
			piecePos:  "a1",
			moves:     []int{8, 16, 24, 32, 40, 48, 56, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			fenString: "8/8/4P3/8/3PR1P1/4P3/8/8 w - - 0 1",
			piecePos:  "e4",
			moves:     []int{29, 36},
		},
		{
			fenString:     "8/8/4P3/8/3PR1p1/4P3/8/8 w - - 0 1",
			piecePos:      "e4",
			moves:         []int{29, 30, 36},
			capturesMoves: []int{30},
		},
		{
			fenString:     "rn2kbnr/1pp1pppp/8/p7/R5bq/8/P1PPPPPP/1NBQKBNR w - - 0 1",
			piecePos:      "a4",
			moves:         []int{16, 32, 25, 26, 27, 28, 29, 30},
			capturesMoves: []int{32, 30},
		},
	}

	for _, d := range data {
		pos, err := NewPositionFromFEN(d.fenString)
		if err != nil {
			t.Fatal(err.Error())
		}

		moves, capturesMoves := pos.SliderPseudoLegalMoves(SquareToIdx(d.piecePos))
		assert.ElementsMatch(t, d.moves, moves)
		assert.ElementsMatch(t, d.capturesMoves, capturesMoves)
	}
}
