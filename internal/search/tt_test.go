package search

import (
	board "chessV2/internal/board"
	"chessV2/internal/eval"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchTTScoreRoundTripForMateScores(t *testing.T) {
	tests := []struct {
		name  string
		score eval.Score
		ply   int
	}{
		{name: "mate for side to move", score: eval.MateIn(3), ply: 5},
		{name: "mated line", score: eval.MatedIn(4), ply: 7},
		{name: "normal cp score", score: eval.Score(123), ply: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored := ttScoreForStorage(tt.score, tt.ply)
			loaded := ttScoreFromStored(stored, tt.ply)
			assert.Equal(t, tt.score, loaded)
		})
	}
}

func TestSearchTTProbeAndStore(t *testing.T) {
	tt := newSearchTT()
	key := uint64(42)
	move := board.NewMove(board.Piece(board.White|board.Knight), board.G1, board.F3, board.NormalMove)

	tt.store(key, 5, 3, eval.MateIn(2), ttBoundExact, move)

	entry, ok := tt.probe(key, 5, 3)
	assert.True(t, ok)
	assert.Equal(t, ttBoundExact, entry.bound)
	assert.Equal(t, move, entry.bestMove)
	assert.Equal(t, eval.MateIn(2), entry.score)

	_, ok = tt.probe(key, 6, 3)
	assert.False(t, ok)
}
