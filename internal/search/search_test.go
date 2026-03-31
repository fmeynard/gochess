package search

import (
	board "chessV2/internal/board"
	"chessV2/internal/eval"
	"chessV2/internal/movegen"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlphaBetaSearcherSearch(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)

	tests := map[string]struct {
		fen          string
		depth        int
		expectedMove string
		assertScore  func(t *testing.T, score eval.Score)
	}{
		"mate in one is found": {
			fen:          "7k/5KQ1/8/8/8/8/8/8 w - - 0 1",
			depth:        1,
			assertScore: func(t *testing.T, score eval.Score) {
				assert.Greater(t, score, eval.Score(29000))
			},
		},
		"winning queen capture is preferred": {
			fen:          "3qk3/8/8/8/8/8/3Q4/4K3 w - - 0 1",
			depth:        1,
			expectedMove: "d2d8",
			assertScore: func(t *testing.T, score eval.Score) {
				assert.Greater(t, score, eval.DrawScore)
			},
		},
		"stalemate position returns draw score": {
			fen:   "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1",
			depth: 1,
			assertScore: func(t *testing.T, score eval.Score) {
				assert.Equal(t, eval.DrawScore, score)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pos, err := board.NewPositionFromFEN(tt.fen)
			assert.NoError(t, err)

			result, err := searcher.Search(pos, Limits{Depth: tt.depth})
			assert.NoError(t, err)

			tt.assertScore(t, result.Score)

			if tt.expectedMove != "" {
				assert.Equal(t, tt.expectedMove, result.BestMove.UCI())
			}
		})
	}
}

func TestAlphaBetaSearcherSearchRejectsInvalidDepth(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	assert.NoError(t, err)

	_, err = searcher.Search(pos, Limits{Depth: 0})
	assert.ErrorIs(t, err, ErrInvalidLimits)
}
