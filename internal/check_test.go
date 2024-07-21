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

	NewEngine()
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
		"Rook left horizontal check": {
			"8/8/8/8/r2K4/8/8/8 w - - 0 1",
			D4,
			White,
			true,
		},
		"Rook left horizontal blocked": {
			"8/8/8/8/rR1K4/8/8/8 w - - 0 1",
			D4,
			White,
			false,
		},
		"Rook right horizontal blocked": {
			"8/8/8/8/3K2nr/8/8/8 w - - 0 1",
			D4,
			White,
			false,
		},
		"Rook right horizontal check": {
			"8/8/8/8/3K3r/8/8/8 w - - 0 1",
			D4,
			White,
			true,
		},

		"Rook right blocked by opponent rook": {
			"8/8/8/8/3K2Rr/8/8/8 w - - 0 1",
			D4,
			White,
			false,
		},
		"Queen lateral check": {
			"8/8/8/8/q2K4/8/8/8 w - - 0 1",
			D4,
			White,
			true,
		},
		"Rook top check": {
			"3r4/8/8/8/3K4/8/8/8 w - - 0 1",
			D4,
			White,
			true,
		},
		"Rook top blocked": {
			"3r4/3R4/8/8/3K4/8/8/8 w - - 0 1",
			D4,
			White,
			false,
		},
		"Rook everywhere but blocked": {
			"2rrr2r/r2R4/8/r6r/rR1K2Rr/r6r/3R4/r1rrr1r1 w - - 0 1",
			D4,
			White,
			false,
		},
		"random position 1 : no check": {
			"r1bqkbnr/1ppppppp/8/p7/8/P2KP3/1PPP1PPP/RNB1QBnR b KQkq - 0 1",
			E8,
			Black,
			false,
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

func TestIsFileAttackedByEnemy(t *testing.T) {
	data := map[string]struct {
		fenPos   string
		pieceIdx int8
		result   bool
	}{
		"startPos rook directions no check": {
			"r1bqkbnr/1ppppppp/8/p7/8/P2KP3/1PPP1PPP/RNB1QBnR b KQkq - 0 1",
			E8,
			false,
		},
		"random position 1 : no check by white": {
			"r1bqkbnr/1ppppppp/8/p7/8/P2KP3/1PPP1PPP/RNB1QBnR b KQkq - 0 1",
			D3,
			false,
		},
		"random position 2 : no check by white": {
			"r1bqkbnr/1pp1pppp/1p6/p7/8/P2KP3/1PPP1PPP/RNB1QBnR w KQkq - 0 1",
			D3,
			true,
		},
		"Rook vertical check": {
			"8/8/8/8/3K4/8/8/3r4 w - - 0 1",
			D4,
			true,
		},
		"Rook everywhere but blocked": {
			"2rrr2r/r2R4/8/r6r/rR1K2Rr/r6r/3R4/r1rrr1r1 w - - 0 1",
			D4,
			false,
		},
		"Rook top blocked": {
			"3r4/3R4/8/8/3K4/8/8/8 w - - 0 1",
			D4,
			false,
		},
		"Random position 2": {
			"2bqkbnr/1ppppppp/8/rK6/p7/P3P3/1PPP1PPP/RNB1QBnR w KQkq - 0 1",
			B5,
			false,
		},
		"Random position 3": {
			"2bqkbnr/1ppppppp/8/2r5/p1K5/P3P3/RPPP1PPP/1NB1QBnR w KQkq - 0 1",
			C4,
			true,
		},
		"Random position 4": {
			"2bqkbnr/1ppppppp/8/1r6/pK6/P3P3/RPPP1PPP/1NB1QBnR w KQkq - 0 1",
			B4,
			true,
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			assert.Equal(
				t,
				d.result,
				isFileAttackedByEnemy(pos, d.pieceIdx, pos.OpponentColor()),
			)
		})
	}
}

func TestIsRankAttackedByEnemy(t *testing.T) {
	data := map[string]struct {
		fenPos   string
		pieceIdx int8
		result   bool
	}{
		"Random position 2": {
			"2bqkbnr/1ppppppp/8/rK6/p7/P3P3/1PPP1PPP/RNB1QBnR w KQkq - 0 1",
			B5,
			true,
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			assert.Equal(
				t,
				d.result,
				isRankAttackedByEnemy(pos, d.pieceIdx, pos.OpponentColor()),
			)
		})
	}
}
