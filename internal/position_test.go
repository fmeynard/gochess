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

func TestPosition_IsCheck(t *testing.T) {
	data := map[string]struct {
		fenPos  string
		isCheck bool
	}{
		"startPos white no check": {
			fenPos:  FenStartPos,
			isCheck: false,
		},
		"Check By Queen": {
			fenPos:  "rnb1kbnr/ppppqppp/8/5p2/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			isCheck: true,
		},
		"Black NOT inCheck (white is)": {
			fenPos:  "rnb1kbnr/ppppqppp/8/5p2/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1",
			isCheck: false,
		},
		"Check by Knight": {
			fenPos:  "rnb1kb1r/ppppqppp/8/3n1p2/5K2/8/PPPP1PPP/RNBQ1BNR w KQkq - 0 1",
			isCheck: true,
		},
		"Check by pawn": {
			fenPos:  "rnb1kb1r/ppppqppp/8/3n1p2/4K3/8/PPPP1PPP/RNBQ1BNR w KQkq - 0 1",
			isCheck: true,
		},
		"NOT Check by pawns": {
			fenPos:  "8/8/8/4p3/3pK3/8/8/8 w - - 0 1",
			isCheck: false,
		},
		"NOT Check by bishop same file": {
			fenPos:  "4b3/8/8/8/4K3/8/8/8 w - - 0 1",
			isCheck: false,
		},
		"NOT Check by bishop same rank": {
			fenPos:  "8/8/8/8/b3K3/8/8/8 w - - 0 1",
			isCheck: false,
		},
		"NOT Check by black pawn, pawnRank < king rank": {
			fenPos:  "8/8/8/8/4K3/3p4/8/8 w - - 0 1",
			isCheck: false,
		},
		"NOT Check by white pawn, pawnRank > king rank": {
			fenPos:  "8/8/8/3P4/4k3/8/8/8 b - - 0 1",
			isCheck: false,
		},
		"NOT Check by pawn same rank": {
			fenPos:  "rnb1k2r/ppppbppp/8/3np2q/4K3/8/PPPP1PPP/RNBQ1BNR w KQkq - 0 1",
			isCheck: false,
		},
		"Check By Rook same file": {
			fenPos:  "4r3/8/8/8/4K3/8/8/8 w - - 0 1",
			isCheck: true,
		},
		"Check By Rook same rank": {
			fenPos:  "8/8/8/8/R3k3/8/8/8 b - - 0 1",
			isCheck: true,
		},
		"NOT Check By Rook in diagonal": {
			fenPos:  "R7/8/8/8/4k3/8/8/8 b - - 0 1",
			isCheck: false,
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			assert.Equal(t, d.isCheck, pos.IsCheck())
		})
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
