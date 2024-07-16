package internal

import "testing"

func BenchmarkPerftDivideB(b *testing.B) {
	pos, _ := NewPositionFromFEN("rnbqkbnr/2pppppp/8/pp6/2P5/N7/PP1PPPPP/R1BQKBNR w KQkq - 0 3")
	engine := NewEngine()
	engine.PerftDivide(&pos, 2)
}

func TestPerftDivideB(t *testing.T) {
	engine := NewEngine()
	initPos, _ := NewPositionFromFEN(FenStartPos)
	pos := engine.positionUpdater.PositionAfterMove(&initPos, NewMove(Piece(White|Pawn), E2, E3, NormalMove))
	pos = engine.positionUpdater.PositionAfterMove(pos, NewMove(Piece(Black|Pawn), A7, A6, NormalMove))
	//pos = pos.PositionAfterMove(NewMove(Piece(White|Rook), F1, B5, NormalMove))

	draw(pos.occupied)

	engine.PerftDivide(pos, 4)
}
