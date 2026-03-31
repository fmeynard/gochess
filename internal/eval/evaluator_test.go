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

