package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewPositionFromFEN(t *testing.T) {

	type Check struct {
		square int8
		piece  Piece
	}

	data := []struct {
		fenPos string
		checks []Check
	}{
		{
			fenPos: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			checks: []Check{
				{A1, Piece(Rook | White)},
				{A8, Piece(Rook | Black)},
				{A4, NoPiece},
			},
		},
		{
			fenPos: "1k6/8/8/8/3Q4/8/8/K7 w KQkq - 0 1",
			checks: []Check{
				{B8, Piece(King | Black)},
				{A1, Piece(King | White)},
			},
		},
	}

	for _, d := range data {
		pos, err := NewPositionFromFEN(d.fenPos)
		if err != nil {
			t.Error(err)
		}

		for _, check := range d.checks {
			assert.Equal(t, check.piece, pos.PieceAt(check.square))
		}
	}
}

func TestPosition_CanCastle(t *testing.T) {
	data := []struct {
		fenPos             string
		color              int8
		canCastleQueenSide bool
		canCastleKingSide  bool
	}{
		{
			fenPos:             "8/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			color:              White,
			canCastleQueenSide: true,
			canCastleKingSide:  true,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K3 w KQkq - 0 1",
			color:              White,
			canCastleQueenSide: true,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "r3k2r/8/8/8/8/8/8/8 w KQkq - 0 1",
			color:              Black,
			canCastleQueenSide: true,
			canCastleKingSide:  true,
		},
		{
			fenPos:             "r3k3/8/8/8/8/8/8/8 w KQkq - 0 1",
			color:              Black,
			canCastleQueenSide: true,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K3 w - - 0 1",
			color:              White,
			canCastleQueenSide: false,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K3 w - - 0 1",
			color:              Black,
			canCastleQueenSide: false,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K2R w kq - 0 1",
			color:              White,
			canCastleQueenSide: false,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K2R w KQ - 0 1",
			color:              White,
			canCastleQueenSide: true,
			canCastleKingSide:  true,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/r3K2r w KQkq - 0 1",
			color:              White,
			canCastleQueenSide: false,
			canCastleKingSide:  false,
		},
	}

	for _, d := range data {
		pos, err := NewPositionFromFEN(d.fenPos)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, d.canCastleQueenSide, pos.CanCastle(d.color, QueenSideCastle))
		assert.Equal(t, d.canCastleKingSide, pos.CanCastle(d.color, KingSideCastle))
	}
}
