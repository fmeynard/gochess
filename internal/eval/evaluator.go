package eval

import (
	board "chessV2/internal/board"
	"chessV2/internal/movegen"
	"math/bits"
)

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

var mobilityWeights = [7]Score{
	0, // no piece
	0, // king
	1, // queen
	0, // pawn
	4, // knight
	5, // bishop
	2, // rook
}

var exposedPieceWeights = [7]Score{
	0,  // no piece
	0,  // king
	32, // queen
	4,  // pawn
	12, // knight
	12, // bishop
	20, // rook
}

const (
	kingRingAttackWeight Score = 8
	kingInCheckPenalty   Score = 40
	missingShieldPenalty Score = 10
)

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

	whiteScore += mobilityScore(pos, board.White)
	blackScore += mobilityScore(pos, board.Black)
	whiteScore -= exposedPiecesPenalty(pos, board.White)
	blackScore -= exposedPiecesPenalty(pos, board.Black)
	whiteScore -= kingSafetyPenalty(pos, board.White)
	blackScore -= kingSafetyPenalty(pos, board.Black)

	if pos.ActiveColor() == board.White {
		return whiteScore - blackScore
	}
	return blackScore - whiteScore
}

func mobilityScore(pos *board.Position, color int8) Score {
	score := DrawScore
	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece || piece.Color() != color {
			continue
		}

		weight := mobilityWeights[piece.Type()]
		if weight == 0 {
			continue
		}

		targets := movegen.PseudoLegalTargetsMask(pos, piece, idx)
		score += Score(bits.OnesCount64(targets)) * weight
	}
	return score
}

func exposedPiecesPenalty(pos *board.Position, color int8) Score {
	enemyColor := board.White
	if color == board.White {
		enemyColor = board.Black
	}

	penalty := DrawScore
	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece || piece.Color() != color || piece.Type() == board.King {
			continue
		}

		attackers := attackCountOnSquare(pos, enemyColor, idx)
		if attackers == 0 {
			continue
		}

		defenders := attackCountOnSquare(pos, color, idx)
		base := exposedPieceWeights[piece.Type()] * Score(attackers)
		if defenders == 0 {
			base *= 2
		} else if attackers > defenders {
			base += exposedPieceWeights[piece.Type()] * Score(attackers-defenders)
		}
		penalty += base
	}

	return penalty
}

func kingSafetyPenalty(pos *board.Position, color int8) Score {
	kingIdx := pos.BlackKingIdx()
	enemyColor := board.White
	if color == board.White {
		kingIdx = pos.WhiteKingIdx()
		enemyColor = board.Black
	}

	penalty := DrawScore
	if movegen.IsKingInCheck(pos, color) {
		penalty += kingInCheckPenalty
	}

	ring := movegen.KingRingMask(kingIdx)
	enemyAttacks := attackMap(pos, enemyColor)
	penalty += Score(bits.OnesCount64(ring&enemyAttacks)) * kingRingAttackWeight
	penalty += pawnShieldPenalty(pos, color, kingIdx)
	return penalty
}

func attackMap(pos *board.Position, color int8) uint64 {
	mask := uint64(0)
	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece || piece.Color() != color {
			continue
		}
		mask |= movegen.PieceAttackMask(pos, piece, idx)
	}
	return mask
}

func attackCountOnSquare(pos *board.Position, color, square int8) int {
	count := 0
	squareMask := uint64(1) << square
	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece || piece.Color() != color {
			continue
		}
		if movegen.PieceAttackMask(pos, piece, idx)&squareMask != 0 {
			count++
		}
	}
	return count
}

func pawnShieldPenalty(pos *board.Position, color, kingIdx int8) Score {
	file := board.FileFromIdx(kingIdx)
	rank := board.RankFromIdx(kingIdx)
	penalty := DrawScore

	for df := int8(-1); df <= 1; df++ {
		targetFile := file + df
		if targetFile < 0 || targetFile > 7 {
			continue
		}

		shieldRank := rank + 1
		if color == board.Black {
			shieldRank = rank - 1
		}
		if shieldRank < 0 || shieldRank > 7 {
			penalty += missingShieldPenalty
			continue
		}

		shieldIdx := shieldRank*8 + targetFile
		expectedPawn := board.Piece(color | board.Pawn)
		if pos.PieceAt(shieldIdx) != expectedPawn {
			penalty += missingShieldPenalty
		}
	}

	return penalty
}
