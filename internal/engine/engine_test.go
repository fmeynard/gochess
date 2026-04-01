package engine

import (
	. "chessV2/internal/board"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkPerftDivideB(b *testing.B) {
	pos, _ := NewPositionFromFEN("rnbqkbnr/2pppppp/8/pp6/2P5/N7/PP1PPPPP/R1BQKBNR w KQkq - 0 3")
	engine := NewEngine()
	engine.PerftDivide(pos, 2)
}

func TestPerftDivideB(t *testing.T) {
	engine := NewEngine()
	pos, _ := NewPositionFromFEN(FenStartPos)
	engine.positionUpdater.MakeMove(pos, NewMove(Piece(White|Pawn), E2, E3, NormalMove))
	engine.positionUpdater.MakeMove(pos, NewMove(Piece(Black|Pawn), A7, A6, NormalMove))
	//pos = pos.PositionAfterMove(NewMove(Piece(White|Rook), F1, B5, NormalMove))

	engine.PerftDivide(pos, 5)
}

func BenchmarkEngine_LegalMoves(b *testing.B) {
	pos, _ := NewPositionFromFEN(FenStartPos)
	engine := NewEngine()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine.LegalMoves(pos)
	}
}

func TestApplyUCIMovesRegression_CurrentIllegalMoveCases(t *testing.T) {
	engine := NewEngine()

	tests := map[string]struct {
		moves         []string
		expectedFEN   string
		expectedLegal []string
	}{
		"game 38 sequence reconstructs expected black position": {
			moves: []string{
				"b1a3", "a7a5", "g1f3", "d7d5", "a1b1", "a5a4", "b1a1", "e7e6",
				"a1b1", "b8c6", "b1a1", "e6e5", "f3e5", "c6e5", "a1b1", "d5d4",
				"b1a1", "f8a3", "b2a3", "e8f8", "a1b1", "f8e8", "b1b7",
			},
			expectedFEN: "r1bqk1nr/1Rp2ppp/8/4n3/p2p4/P7/P1PPPPPP/2BQKB1R b K - 0 1",
			expectedLegal: []string{
				"e8d7", "e8e7", "e8f8", "e5d3", "e5f3", "d8d7",
			},
		},
		"game 49 sequence reconstructs expected white position": {
			moves: []string{
				"b1c3", "a7a5", "g1f3", "b8c6", "d2d4", "c6b4", "a1b1", "b4a2",
				"c3a2", "d7d6", "b1a1", "c8g4", "c1e3", "g4f3",
			},
			expectedFEN: "r2qkbnr/1pp1pppp/3p4/p7/3P4/4Bb2/NPP1PPPP/R2QKB1R w Kkq - 0 1",
			expectedLegal: []string{
				"e1d2", "d4d5", "e3f4", "a1b1",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pos, err := NewPositionFromFEN(FenStartPos)
			assert.NoError(t, err)

			err = engine.ApplyUCIMoves(pos, tt.moves)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFEN, pos.FEN())

			legal := movesToUCI(engine.LegalMoves(pos))
			for _, expected := range tt.expectedLegal {
				assert.Contains(t, legal, expected)
			}
		})
	}
}
