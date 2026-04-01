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
		"central king is rewarded in endgame": {
			fen: "4k3/8/8/3K4/8/8/8/8 w - - 0 1",
			assertion: func(t *testing.T, center Score) {
				edgePos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/8/K7 w - - 0 1")
				assert.NoError(t, err)
				edge := evaluator.Evaluate(edgePos)
				assert.Greater(t, center, edge)
			},
		},
		"advanced passed pawn scores better than blocked pawn": {
			fen: "4k3/8/4P3/8/8/8/8/4K3 w - - 0 1",
			assertion: func(t *testing.T, passed Score) {
				blockedPos, err := board.NewPositionFromFEN("4k3/4p3/4P3/8/8/8/8/4K3 w - - 0 1")
				assert.NoError(t, err)
				blocked := evaluator.Evaluate(blockedPos)
				assert.Greater(t, passed, blocked)
			},
		},
		"protected queen scores better than unprotected queen": {
			fen: "4k3/8/8/8/8/8/4P3/4QK2 w - - 0 1",
			assertion: func(t *testing.T, protected Score) {
				unprotectedPos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/8/4QK2 w - - 0 1")
				assert.NoError(t, err)
				unprotected := evaluator.Evaluate(unprotectedPos)
				assert.Greater(t, protected, unprotected)
			},
		},
		"bishop attacked by cheaper piece with bad defense is heavily penalized": {
			fen: "4k3/8/8/8/8/3p4/4B3/3QK3 w - - 0 1",
			assertion: func(t *testing.T, exposed Score) {
				safePos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/4B3/3QK3 w - - 0 1")
				assert.NoError(t, err)
				safe := evaluator.Evaluate(safePos)
				assert.Less(t, exposed, safe)
			},
		},
		"healthy pawn chain scores better than isolated doubled pawns": {
			fen: "4k3/8/8/8/8/2P5/3PP3/4K3 w - - 0 1",
			assertion: func(t *testing.T, healthy Score) {
				weakPos, err := board.NewPositionFromFEN("4k3/8/8/8/8/3P4/3P4/4K3 w - - 0 1")
				assert.NoError(t, err)
				weak := evaluator.Evaluate(weakPos)
				assert.Greater(t, healthy, weak)
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
