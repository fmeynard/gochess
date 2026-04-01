package search

import (
	board "chessV2/internal/board"
	"chessV2/internal/eval"
	"chessV2/internal/movegen"
	"testing"
	"time"

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
			fen:   "7k/5KQ1/8/8/8/8/8/8 w - - 0 1",
			depth: 1,
			assertScore: func(t *testing.T, score eval.Score) {
				assert.Greater(t, score, eval.Score(29000))
			},
		},
		"winning queen capture is preferred": {
			fen:   "3qk3/8/8/8/8/8/3Q4/4K3 w - - 0 1",
			depth: 1,
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

func TestAlphaBetaSearcherSearchWithMoveTime(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	assert.NoError(t, err)

	result, err := searcher.Search(pos, Limits{MoveTime: 20 * time.Millisecond})
	assert.NoError(t, err)
	assert.NotEqual(t, board.Move{}, result.BestMove)
	assert.GreaterOrEqual(t, result.Stats.Depth, 0)
	assert.GreaterOrEqual(t, result.Stats.Time, time.Duration(0))
}

func TestAlphaBetaSearcherSearchReturnsFallbackMoveWhenStoppedImmediately(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	assert.NoError(t, err)

	stop := make(chan struct{})
	close(stop)

	result, err := searcher.Search(pos, Limits{Depth: 3, Stop: stop})
	assert.NoError(t, err)
	assert.NotEqual(t, board.Move{}, result.BestMove)
}

func TestAlphaBetaSearcherSearchWithDepthAndMoveTime(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN("3qk3/8/8/8/8/8/3Q4/4K3 w - - 0 1")
	assert.NoError(t, err)

	result, err := searcher.Search(pos, Limits{Depth: 2, MoveTime: 50 * time.Millisecond})
	assert.NoError(t, err)
	assert.NotEqual(t, board.Move{}, result.BestMove)
	assert.GreaterOrEqual(t, result.Stats.Depth, 0)
	assert.LessOrEqual(t, result.Stats.Depth, 2)
}

func TestAlphaBetaSearcherQuiescenceAvoidsPoisonedPawn(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN("4k3/3p4/8/8/8/8/3Q4/4K3 w - - 0 1")
	assert.NoError(t, err)

	result, err := searcher.Search(pos, Limits{Depth: 1})
	assert.NoError(t, err)
	assert.NotEqual(t, "d2d7", result.BestMove.UCI())
	assert.Greater(t, result.Stats.QuiescenceNodes, uint64(0))
}

func TestAlphaBetaSearcherQuiescenceAvoidsPoisonedQueenCapture(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN("3rk3/3p4/8/8/8/8/3Q4/4K3 w - - 0 1")
	assert.NoError(t, err)

	result, err := searcher.Search(pos, Limits{Depth: 1})
	assert.NoError(t, err)
	assert.NotEqual(t, "d2d7", result.BestMove.UCI())
	assert.Greater(t, result.Stats.QuiescenceNodes, uint64(0))
}

func TestSEELiteRejectsPoisonedQueenCapture(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN("3rk3/3p4/8/8/8/8/3Q4/4K3 w - - 0 1")
	assert.NoError(t, err)

	move := board.NewMove(board.Piece(board.White|board.Queen), board.D2, board.D7, board.Capture)
	assert.Less(t, searcher.seeLite(pos, move), 0)
}

func TestSEELiteKeepsWinningQueenCapturePositive(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/3r4/3QK3 w - - 0 1")
	assert.NoError(t, err)

	move := board.NewMove(board.Piece(board.White|board.Queen), board.D1, board.D2, board.Capture)
	assert.Greater(t, searcher.seeLite(pos, move), 0)
}

func TestRepetitionTrackerDetectsThreefold(t *testing.T) {
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	assert.NoError(t, err)

	history := []uint64{pos.ZobristKey()}
	moveGenerator := movegen.NewPseudoLegalMoveGenerator()
	updater := board.NewPositionUpdater()

	for _, uci := range []string{"g1f3", "g8f6", "f3g1", "f6g8", "g1f3", "g8f6", "f3g1", "f6g8"} {
		var moves [256]board.Move
		moveCount := moveGenerator.LegalMovesInto(pos, updater, moves[:])
		var selected board.Move
		for i := 0; i < moveCount; i++ {
			if moves[i].UCI() == uci {
				selected = moves[i]
				break
			}
		}

		assert.NotEqual(t, board.Move{}, selected)
		updater.MakeMove(pos, selected)
		history = append(history, pos.ZobristKey())
	}

	tracker := newRepetitionTracker(pos, history)
	assert.True(t, tracker.isThreefold())
}

func TestRepetitionScoreDiscouragesDrawWhenAhead(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)

	betterPos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/4Q3/4K3 w - - 0 1")
	assert.NoError(t, err)
	assert.Less(t, searcher.repetitionScore(betterPos), eval.DrawScore)

	worsePos, err := board.NewPositionFromFEN("4k3/8/8/8/8/8/4q3/4K3 w - - 0 1")
	assert.NoError(t, err)
	assert.Greater(t, searcher.repetitionScore(worsePos), eval.DrawScore)
}

func TestOrderMovesPrefersCaptures(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN("4k3/8/8/3p4/4P3/8/8/4K3 w - - 0 1")
	assert.NoError(t, err)

	moves := []board.Move{
		board.NewMove(board.Piece(board.White|board.Pawn), board.E4, board.E5, board.NormalMove),
		board.NewMove(board.Piece(board.White|board.Pawn), board.E4, board.D5, board.Capture),
	}

	searcher.orderMoves(pos, moves, 0, board.Move{})
	assert.Equal(t, "e4d5", moves[0].UCI())
}

func TestOrderMovesPrefersTTMove(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	assert.NoError(t, err)

	moves := []board.Move{
		board.NewMove(board.Piece(board.White|board.Knight), board.B1, board.C3, board.NormalMove),
		board.NewMove(board.Piece(board.White|board.Knight), board.G1, board.F3, board.NormalMove),
	}

	searcher.orderMoves(pos, moves, 0, moves[1])
	assert.Equal(t, "g1f3", moves[0].UCI())
}

func TestOrderMovesPrefersKillerMove(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	assert.NoError(t, err)

	killer := board.NewMove(board.Piece(board.White|board.Knight), board.G1, board.F3, board.NormalMove)
	searcher.recordKiller(3, killer)

	moves := []board.Move{
		board.NewMove(board.Piece(board.White|board.Knight), board.B1, board.C3, board.NormalMove),
		killer,
	}

	searcher.orderMoves(pos, moves, 3, board.Move{})
	assert.Equal(t, "g1f3", moves[0].UCI())
}

func TestOrderMovesPrefersHistoryMove(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	assert.NoError(t, err)

	historyMove := board.NewMove(board.Piece(board.White|board.Knight), board.G1, board.F3, board.NormalMove)
	searcher.recordHistory(historyMove, 6)

	moves := []board.Move{
		board.NewMove(board.Piece(board.White|board.Knight), board.B1, board.C3, board.NormalMove),
		historyMove,
	}

	searcher.orderMoves(pos, moves, 2, board.Move{})
	assert.Equal(t, "g1f3", moves[0].UCI())
}

func TestAlphaBetaSearcherSearchWithMoveTimeRegressionGame45NeverReturnsZeroMove(t *testing.T) {
	searcher := NewAlphaBetaSearcher(
		movegen.NewPseudoLegalMoveGenerator(),
		board.NewPositionUpdater(),
		eval.NewStaticEvaluator(),
	)
	pos, err := board.NewPositionFromFEN("r1bqkbnr/1ppppppp/8/n7/8/P1N1PN2/P1PP1PPP/R1BQKBR1 w Qkq - 0 1")
	assert.NoError(t, err)

	legal := make(map[string]struct{})
	for _, move := range engineLegalMovesForTest(pos) {
		legal[move.UCI()] = struct{}{}
	}

	for i := 0; i < 50; i++ {
		searchPos := pos.Clone()
		beforeFEN := searchPos.FEN()
		result, err := searcher.Search(searchPos, Limits{MoveTime: 250 * time.Millisecond})
		assert.NoError(t, err)
		assert.Equal(t, beforeFEN, searchPos.FEN(), "iteration %d mutated root position", i)
		assert.NotEqual(t, board.Move{}, result.BestMove)
		assert.NotEqual(t, "0000", result.BestMove.UCI())
		_, ok := legal[result.BestMove.UCI()]
		assert.True(t, ok, "iteration %d returned illegal move %s", i, result.BestMove.UCI())
	}
}

func engineLegalMovesForTest(pos *board.Position) []board.Move {
	moveGenerator := movegen.NewPseudoLegalMoveGenerator()
	updater := board.NewPositionUpdater()
	var buf [256]board.Move
	count := moveGenerator.LegalMovesInto(pos, updater, buf[:])
	moves := make([]board.Move, count)
	copy(moves, buf[:count])
	return moves
}
