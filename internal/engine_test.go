package internal

import (
	"fmt"
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

func TestCustom(t *testing.T) {
	var fen string
	fen = "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"
	fen = "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"
	pos, _ := NewPositionFromFEN(fen)

	engine := NewEngine()

	assert.Equal(t, KingIsCheck, pos.whiteKingSafety)

	//testMove := NewMove(Piece(Rook|White), F1, F2, NormalMove)
	initialMoves := engine.LegalMoves(pos)

	assert.NotContains(t, movesToUci(initialMoves), "f1e1")
	fmt.Println("movesToUci(initialMoves)", movesToUci(initialMoves))
	//for _, testMove := range initialMoves {
	//
	//	history := engine.positionUpdater.MakeMove(pos, testMove)
	//	for _, testMove2 := range engine.LegalMoves(pos) {
	//
	//		history2 := engine.positionUpdater.MakeMove(pos, testMove)
	//		engine.positionUpdater.UnMakeMove(pos, testMove2, history2)
	//	}
	//
	//	engine.positionUpdater.UnMakeMove(pos, testMove, history)
	//}
	//newMoves := engine.LegalMoves(pos)
	//assert.NotContains(t, "f1e1", movesToUci(newMoves))
	//engine.positionUpdater.MakeMove(pos, NewMove(Piece(Rook|White), F1, E1, NormalMove))
	//engine.positionUpdater.MakeMove(pos, NewMove(Piece(Rook|White), A1, C1, NormalMove))

	//assert.Equal(t, true, IsKingInCheck(pos, White))

	fmt.Println("cachehits", cacheHits)
	fmt.Println("cachemiss", cacheMiss)
	//assert.Equal(t, NotCalculated, pos.whiteKingSafety)
}

func TestCustom2(t *testing.T) {
	var fen string
	fen = "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"
	pos, _ := NewPositionFromFEN(fen)

	assert.Equal(t, KingIsCheck, pos.whiteKingSafety)
	assert.Equal(t, KingIsCheck, pos.whiteKingSafety)

	fmt.Println("cachehits", cacheHits)
	fmt.Println("cachemiss", cacheMiss)

	positionUpdater := NewPositionUpdater(NewBitsBoardMoveGenerator())

	m := NewMove(Piece(Rook|White), A1, B1, NormalMove)

	positionUpdater.MakeMove(pos, m)

	assert.Equal(t, KingIsCheck, pos.whiteKingSafety)
	fmt.Println("affect", positionUpdater.IsMoveAffectsKing(pos, m, pos.whiteKingIdx))

	//testMove := NewMove(Piece(Rook|White), F1, F2, NormalMove)

	//assert.Equal(t, NotCalculated, pos.whiteKingSafety)
}

func TestCustom3(t *testing.T) {
	var fen string
	fen = "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"
	pos, _ := NewPositionFromFEN(fen)

	assert.Equal(t, KingIsCheck, pos.whiteKingSafety)
	IsKingInCheck(pos, White)
	IsKingInCheck(pos, White)
	assert.Equal(t, KingIsCheck, pos.whiteKingSafety)

	tryMoves := []Move{
		NewMove(Piece(Rook|White), A1, B1, NormalMove),
		NewMove(Piece(Rook|White), A1, C1, NormalMove),
		NewMove(Piece(Queen|White), D1, C2, NormalMove),
		NewMove(Piece(Queen|White), D1, B3, NormalMove),
		NewMove(Piece(Queen|White), D1, E2, NormalMove),
		NewMove(Piece(Queen|White), D1, C1, NormalMove),
		NewMove(Piece(Queen|White), D1, E1, NormalMove),
		NewMove(Piece(Rook|White), F1, F2, NormalMove),
	}

	positionUpdater := NewPositionUpdater(NewBitsBoardMoveGenerator())
	fmt.Println("cachehits", cacheHits)
	fmt.Println("cachemiss", cacheMiss)

	for _, move := range tryMoves {
		history := positionUpdater.MakeMove(pos, move)
		positionUpdater.UnMakeMove(pos, move, history)
	}

	assert.Equal(t, KingIsCheck, pos.whiteKingSafety)

	positionUpdater.MakeMove(pos, NewMove(Piece(Rook|White), F1, E1, NormalMove))

	assert.Equal(t, KingIsCheck, pos.whiteKingSafety)
	//
	//m := NewMove(Piece(Rook|White), C4, C5, NormalMove)
	//
	//positionUpdater.MakeMove(pos, m)
	//
	//assert.Equal(t, KingIsCheck, pos.whiteKingSafety)
	//fmt.Println("affect", positionUpdater.IsMoveAffectsKing(pos, m, pos.whiteKingIdx))

	//testMove := NewMove(Piece(Rook|White), F1, F2, NormalMove)

	//assert.Equal(t, NotCalculated, pos.whiteKingSafety)
}
