package internal

import (
	"fmt"
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

func TestCustom(t *testing.T) {
	var fen string
	fen = "r1bqkbnr/1ppppppp/8/p7/8/P2KP3/1PPP1PPP/RNB1QBnR b KQkq - 0 1"
	pos, _ := NewPositionFromFEN(fen)
	engine := NewEngine()

	fmt.Println(movesToUci(engine.LegalMoves(pos)))
}
