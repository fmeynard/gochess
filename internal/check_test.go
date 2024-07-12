package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsSquareAttackedByPawn(t *testing.T) {
	data := map[string]struct {
		fenPos    string
		kingIdx   int8
		kingColor int8
		result    bool
	}{
		"startPos no check": {
			FenStartPos,
			E1,
			White, false,
		},
		"Double doublemove no check": {
			"rnbqkbnr/pppppppp/8/4K3/8/8/PPPPPPPP/RNBQ1BNR b KQkq - 0 1",
			E5,
			White,
			false,
		},
		"Pawn vertical no check": {
			"8/8/3p4/3K4/3p4/8/8/8 b - - 0 1",
			D5,
			White,
			false,
		},
		"Passed Pawns no check": {
			"8/8/8/3K4/2p1p3/8/8/8 w - - 0 1",
			D5,
			White,
			false,
		},
		"Pawn check from left": {
			"8/8/8/8/2p5/3K4/8/8 b - - 0 1",
			D3,
			White,
			true,
		},
		"Pawn check from right": {
			"8/8/8/8/4p3/3K4/8/8 w - - 0 1",
			D3,
			White,
			true,
		},
		"No check by same color": {
			"8/8/8/8/2PPP3/2PKP3/2PPP3/8 w - - 0 1",
			D3,
			White,
			false,
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			assert.Equal(t, d.result, isSquareAttackedByPawn(pos, d.kingIdx, d.kingColor))
		})
	}
}

func TestIsSquareAttackedByKnight(t *testing.T) {
	data := map[string]struct {
		fenPos    string
		kingIdx   int8
		kingColor int8
		result    bool
	}{
		"startPos no check": {
			FenStartPos,
			E1,
			White, false,
		},
		"All invalid captures no check": {
			"nnnnnnnn/nnnnnnnn/nn1n1nnn/n1nnn1nn/nnnKnnnn/n1nnn1nn/nn1n1nnn/nnnnnnnn b - - 0 1",
			D4,
			White,
			false,
		},
		"No cross-board no check": {
			"nnnnnnnn/nnnnnnnn/nnnnnnnn/nnnnnnnn/nnnnnnnn/n1nnnnnn/nn1nnnnn/Knnnnnnn b - - 0 1",
			A1,
			White,
			false,
		},
		"Queens no check": {
			"8/8/8/2Q1Q3/1Q3Q2/3k4/1Q3Q2/2Q1Q3 w - - 0 1",
			D5,
			Black,
			false,
		},
		"Top left check": {
			"8/8/2N5/8/3k4/8/8/8 b - - 0 1",
			D4,
			Black,
			true,
		},
		"Top right check": {
			"8/8/4N3/8/3k4/8/8/8 b - - 0 1",
			D4,
			Black,
			true,
		},
		"Bottom left check": {
			"8/8/8/8/3k4/8/2N5/8 b - - 0 1",
			D4,
			Black,
			true,
		},
		"Bottom right check": {
			"8/8/8/8/3k4/8/4N3/8 b - - 0 1",
			D4,
			Black,
			true,
		},
		"Left top check": {
			"8/8/8/1N6/3k4/8/8/8 b - - 0 1",
			D4,
			Black,
			true,
		},
		"Left bottom check": {
			"8/8/8/8/3k4/1N6/8/8 b - - 0 1",
			D4,
			Black,
			true,
		},
		"Right top check": {
			"8/8/8/5n2/3K4/8/8/8 w - - 0 1",
			D4,
			White,
			true,
		},
		"Right bottom check": {
			"8/8/8/8/3K4/5n2/8/8 w - - 0 1",
			D4,
			White,
			true,
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			assert.Equal(t, d.result, isSquareAttackedByKnight(pos, d.kingIdx, d.kingColor))
		})
	}
}

func TestIsSquareAttackedBySlidingPiece(t *testing.T) {
	data := map[string]struct {
		fenPos    string
		kingIdx   int8
		kingColor int8
		result    bool
	}{
		"startPos rook directions no check": {
			FenStartPos,
			E1,
			White,
			false,
		},
		"startPos bishop directions no check": {
			FenStartPos,
			E1,
			White,
			false,
		},
		"Queen/Bishop blocked no check": {
			"8/3q4/3n4/8/qP1K4/8/5n2/6b1 w - - 0 1",
			D4,
			White,
			false,
		},
		"Queen/Rook blocked no check": {
			"8/3r4/3n4/8/rP1K2Nq/8/8/8 w - - 0 1",
			D4,
			White,
			false,
		},
		"Diagonals misses no check": {
			"bbbb1bbb/b6b/7b/bK5b/7b/b6b/b6b/bbbbb1bb w - - 0 1",
			B5,
			White,
			false,
		},
		"Vertical misses no check": {
			"R7/R7/R7/R7/R7/R7/4k3/1RRR1RRR w - - 0 1",
			E2,
			Black,
			false,
		},
		"Bishop diagonal check": {
			"8/8/8/8/3K4/8/8/b7 w - - 0 1",
			D4,
			White,
			true,
		},
		"Queen diagonal check": {
			"7q/8/8/8/3K4/8/8/8 w - - 0 1",
			D4,
			White,
			true,
		},
		"Rook vertical check": {
			"8/8/8/8/3K4/8/8/3r4 w - - 0 1",
			D4,
			White,
			true,
		},
		"Queen horizontal check": {
			"8/8/8/8/3K4/8/8/3q4 w - - 0 1",
			D4,
			White,
			true,
		},
		"Rook horizontal check": {
			"8/8/8/8/r2K4/8/8/8 w - - 0 1",
			D4,
			White,
			true,
		},
		"Queen lateral check": {
			"8/8/8/8/q2K4/8/8/8 w - - 0 1",
			D4,
			White,
			true,
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			assert.Equal(
				t,
				d.result,
				isSquareAttackedBySlidingPiece(pos, d.kingIdx, d.kingColor),
			)
		})
	}
}
