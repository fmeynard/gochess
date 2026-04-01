package eval

import (
	board "chessV2/internal/board"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticEvaluatorEvaluate(t *testing.T) {
	evaluator := NewStaticEvaluator()

	tests := map[string]struct {
		fen       string
		assertion func(t *testing.T, score Score)
	}{
		"empty material balance is draw": {
			fen: "4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			assertion: func(t *testing.T, score Score) {
				assert.Equal(t, DrawScore, score)
			},
		},
		"extra white queen is positive for white to move": {
			fen: "4k3/8/8/8/8/8/4Q3/4K3 w - - 0 1",
			assertion: func(t *testing.T, score Score) {
				assert.Greater(t, score, DrawScore)
			},
		},
		"extra white queen is negative for black to move": {
			fen: "4k3/8/8/8/8/8/4Q3/4K3 b - - 0 1",
			assertion: func(t *testing.T, score Score) {
				assert.Less(t, score, DrawScore)
			},
		},
		"centralized knight scores better than rim knight": {
			fen: "4k3/8/8/3N4/8/8/8/4K3 w - - 0 1",
			assertion: func(t *testing.T, center Score) {
				rimPos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/N7/4K3 w - - 0 1")
				assert.NoError(t, err)
				rim := evaluator.Evaluate(rimPos)
				assert.Greater(t, center, rim)
			},
		},
		"mirrored extra black rook is positive for black to move": {
			fen: "4k3/8/8/8/8/8/8/r3K3 b - - 0 1",
			assertion: func(t *testing.T, score Score) {
				assert.Greater(t, score, DrawScore)
			},
		},
		"centralized bishop scores better than corner bishop": {
			fen: "4k3/8/8/8/3B4/8/8/4K3 w - - 0 1",
			assertion: func(t *testing.T, center Score) {
				cornerPos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/8/B3K3 w - - 0 1")
				assert.NoError(t, err)
				corner := evaluator.Evaluate(cornerPos)
				assert.Greater(t, center, corner)
			},
		},
		"centralized rook scores better than corner rook": {
			fen: "4k3/8/8/8/3R4/8/8/4K3 w - - 0 1",
			assertion: func(t *testing.T, center Score) {
				cornerPos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/8/R3K3 w - - 0 1")
				assert.NoError(t, err)
				corner := evaluator.Evaluate(cornerPos)
				assert.Greater(t, center, corner)
			},
		},
		"undefended queen under multiple attacks is penalized": {
			fen: "4k3/8/8/8/3n4/2b5/4Q3/4K3 w - - 0 1",
			assertion: func(t *testing.T, exposed Score) {
				safePos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/4Q3/4K3 w - - 0 1")
				assert.NoError(t, err)
				safe := evaluator.Evaluate(safePos)
				assert.Less(t, exposed, safe)
			},
		},
		"king with intact pawn shield scores better than exposed king": {
			fen: "4k3/8/8/8/8/8/3PPP2/4K3 w - - 0 1",
			assertion: func(t *testing.T, shielded Score) {
				exposedPos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1")
				assert.NoError(t, err)
				exposed := evaluator.Evaluate(exposedPos)
				assert.Greater(t, shielded, exposed)
			},
		},
		"king under attack is penalized": {
			fen: "4k3/8/8/8/4r3/8/8/4K3 w - - 0 1",
			assertion: func(t *testing.T, attacked Score) {
				safePos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1")
				assert.NoError(t, err)
				safe := evaluator.Evaluate(safePos)
				assert.Less(t, attacked, safe)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pos, err := board.NewPositionFromFEN(tt.fen)
			assert.NoError(t, err)

			score := evaluator.Evaluate(pos)
			tt.assertion(t, score)
		})
	}
}
