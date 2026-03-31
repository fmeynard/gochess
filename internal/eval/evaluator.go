package eval

import board "chessV2/internal/board"

type Evaluator interface {
	Evaluate(pos *board.Position) Score
}

type ZeroEvaluator struct{}

func NewZeroEvaluator() *ZeroEvaluator {
	return &ZeroEvaluator{}
}

func (e *ZeroEvaluator) Evaluate(pos *board.Position) Score {
	return DrawScore
}

type StaticEvaluator struct{}

func NewStaticEvaluator() *StaticEvaluator {
	return &StaticEvaluator{}
}

var pieceValues = [7]Score{
	0,   // no piece
	0,   // king
	900, // queen
	100, // pawn
	320, // knight
	330, // bishop
	500, // rook
}

var pawnTable = [64]Score{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTable = [64]Score{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopTable = [64]Score{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookTable = [64]Score{
	0, 0, 0, 5, 5, 0, 0, 0,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	5, 10, 10, 10, 10, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var queenTable = [64]Score{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingTable = [64]Score{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}

func mirrorForBlack(idx int8) int8 {
	return idx ^ 56
}

func pieceSquareValue(pieceType int8, idx int8, isWhite bool) Score {
	tableIdx := idx
	if !isWhite {
		tableIdx = mirrorForBlack(idx)
	}

	switch pieceType {
	case board.Pawn:
		return pawnTable[tableIdx]
	case board.Knight:
		return knightTable[tableIdx]
	case board.Bishop:
		return bishopTable[tableIdx]
	case board.Rook:
		return rookTable[tableIdx]
	case board.Queen:
		return queenTable[tableIdx]
	case board.King:
		return kingTable[tableIdx]
	default:
		return 0
	}
}

func (e *StaticEvaluator) Evaluate(pos *board.Position) Score {
	whiteScore := DrawScore
	blackScore := DrawScore

	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece {
			continue
		}

		pieceType := piece.Type()
		score := pieceValues[pieceType] + pieceSquareValue(pieceType, idx, piece.IsWhite())
		if piece.IsWhite() {
			whiteScore += score
		} else {
			blackScore += score
		}
	}

	if pos.ActiveColor() == board.White {
		return whiteScore - blackScore
	}
	return blackScore - whiteScore
}
