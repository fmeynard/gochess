package board

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

func TestSetPieceAt(t *testing.T) {
	data := map[string]struct {
		fenPos string
		idx    int8
		piece  Piece
	}{
		"simple pawn move": {
			fenPos: EmptyBoard,
			idx:    A1,
			piece:  Piece(Rook | White),
		},
		"capture by white": {
			fenPos: "r7/8/8/8/8/8/8/R7 w - - 0 1",
			idx:    A8,
			piece:  Piece(Rook | White),
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			pos.setPieceAt(d.idx, d.piece)

			assert.Equal(t, d.piece, pos.PieceAt(d.idx))

			isOccupiedExpected := d.piece != NoPiece
			assert.Equal(t, isOccupiedExpected, pos.IsOccupied(d.idx))

			if !isOccupiedExpected {
				assert.Equal(t, false, pos.IsColorOccupied(White, d.idx))
				assert.Equal(t, false, pos.IsColorOccupied(Black, d.idx))
			} else {
				assert.Equal(t, true, pos.IsColorOccupied(pos.activeColor, d.idx))
				assert.Equal(t, false, pos.IsColorOccupied(pos.OpponentColor(), d.idx))
			}
		})
	}
}
