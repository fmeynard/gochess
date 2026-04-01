package eval

import (
	board "chessV2/internal/board"
	"chessV2/internal/movegen"
	"math/bits"
)

const maxGamePhase = 24

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

type phaseScore struct {
	mg Score
	eg Score
}

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

var protectedPieceWeights = [7]Score{
	0,  // no piece
	0,  // king
	10, // queen
	2,  // pawn
	5,  // knight
	5,  // bishop
	8,  // rook
}

const (
	kingRingAttackWeightMG Score = 10
	kingRingAttackWeightEG Score = 3
	kingInCheckPenaltyMG   Score = 45
	kingInCheckPenaltyEG   Score = 20
	missingShieldPenaltyMG Score = 12
	missingShieldPenaltyEG Score = 2
	isolatedPawnPenaltyMG  Score = 12
	isolatedPawnPenaltyEG  Score = 7
	doubledPawnPenaltyMG   Score = 10
	doubledPawnPenaltyEG   Score = 8
)

var passedPawnBonusMG = [8]Score{0, 0, 8, 14, 24, 40, 70, 0}
var passedPawnBonusEG = [8]Score{0, 0, 16, 28, 48, 80, 130, 0}

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

var phaseWeights = [7]int{
	0, // no piece
	0, // king
	4, // queen
	0, // pawn
	1, // knight
	1, // bishop
	2, // rook
}

var pawnTableMG = [64]Score{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var pawnTableEG = [64]Score{
	0, 0, 0, 0, 0, 0, 0, 0,
	70, 70, 70, 70, 70, 70, 70, 70,
	20, 24, 28, 34, 34, 28, 24, 20,
	12, 16, 20, 28, 28, 20, 16, 12,
	8, 12, 16, 24, 24, 16, 12, 8,
	5, 8, 10, 12, 12, 10, 8, 5,
	5, 6, 6, -8, -8, 6, 6, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTableMG = [64]Score{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var knightTableEG = [64]Score{
	-40, -25, -20, -20, -20, -20, -25, -40,
	-25, -10, 0, 0, 0, 0, -10, -25,
	-20, 0, 10, 15, 15, 10, 0, -20,
	-20, 5, 15, 20, 20, 15, 5, -20,
	-20, 0, 15, 20, 20, 15, 0, -20,
	-20, 5, 10, 15, 15, 10, 5, -20,
	-25, -10, 0, 5, 5, 0, -10, -25,
	-40, -25, -20, -20, -20, -20, -25, -40,
}

var bishopTableMG = [64]Score{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var bishopTableEG = [64]Score{
	-15, -8, -8, -8, -8, -8, -8, -15,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 0, 6, 10, 10, 6, 0, -8,
	-8, 6, 8, 12, 12, 8, 6, -8,
	-8, 0, 12, 12, 12, 12, 0, -8,
	-8, 10, 10, 10, 10, 10, 10, -8,
	-8, 5, 0, 0, 0, 0, 5, -8,
	-15, -8, -8, -8, -8, -8, -8, -15,
}

var rookTableMG = [64]Score{
	0, 0, 0, 5, 5, 0, 0, 0,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	5, 10, 10, 10, 10, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var rookTableEG = [64]Score{
	0, 2, 4, 6, 6, 4, 2, 0,
	0, 4, 6, 8, 8, 6, 4, 0,
	0, 4, 6, 8, 8, 6, 4, 0,
	0, 4, 6, 8, 8, 6, 4, 0,
	0, 4, 6, 8, 8, 6, 4, 0,
	0, 4, 6, 8, 8, 6, 4, 0,
	2, 6, 8, 10, 10, 8, 6, 2,
	0, 2, 4, 6, 6, 4, 2, 0,
}

var queenTableMG = [64]Score{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var queenTableEG = [64]Score{
	-10, -6, -6, -2, -2, -6, -6, -10,
	-6, 0, 0, 0, 0, 0, 0, -6,
	-6, 0, 4, 4, 4, 4, 0, -6,
	-2, 0, 4, 6, 6, 4, 0, -2,
	-2, 0, 4, 6, 6, 4, 0, -2,
	-6, 4, 4, 4, 4, 4, 0, -6,
	-6, 0, 4, 0, 0, 0, 0, -6,
	-10, -6, -6, -2, -2, -6, -6, -10,
}

var kingTableMG = [64]Score{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}

var kingTableEG = [64]Score{
	-50, -30, -20, -20, -20, -20, -30, -50,
	-30, -10, 0, 0, 0, 0, -10, -30,
	-20, 0, 10, 15, 15, 10, 0, -20,
	-20, 0, 15, 20, 20, 15, 0, -20,
	-20, 0, 15, 20, 20, 15, 0, -20,
	-20, 0, 10, 15, 15, 10, 0, -20,
	-30, -10, 0, 0, 0, 0, -10, -30,
	-50, -30, -20, -20, -20, -20, -30, -50,
}

func mirrorForBlack(idx int8) int8 {
	return idx ^ 56
}

func phaseBlend(mg, eg Score, phase int) Score {
	return (mg*Score(phase) + eg*Score(maxGamePhase-phase)) / Score(maxGamePhase)
}

func pieceSquareValue(pieceType int8, idx int8, isWhite bool, phase int) Score {
	tableIdx := idx
	if !isWhite {
		tableIdx = mirrorForBlack(idx)
	}

	var mg, eg Score
	switch pieceType {
	case board.Pawn:
		mg = pawnTableMG[tableIdx]
		eg = pawnTableEG[tableIdx]
	case board.Knight:
		mg = knightTableMG[tableIdx]
		eg = knightTableEG[tableIdx]
	case board.Bishop:
		mg = bishopTableMG[tableIdx]
		eg = bishopTableEG[tableIdx]
	case board.Rook:
		mg = rookTableMG[tableIdx]
		eg = rookTableEG[tableIdx]
	case board.Queen:
		mg = queenTableMG[tableIdx]
		eg = queenTableEG[tableIdx]
	case board.King:
		mg = kingTableMG[tableIdx]
		eg = kingTableEG[tableIdx]
	default:
		return 0
	}

	return phaseBlend(mg, eg, phase)
}

func (e *StaticEvaluator) Evaluate(pos *board.Position) Score {
	phase := gamePhase(pos)

	whiteScore := DrawScore
	blackScore := DrawScore

	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece {
			continue
		}

		pieceType := piece.Type()
		score := pieceValues[pieceType] + pieceSquareValue(pieceType, idx, piece.IsWhite(), phase)
		if piece.IsWhite() {
			whiteScore += score
		} else {
			blackScore += score
		}
	}

	whiteScore += mobilityScore(pos, board.White)
	blackScore += mobilityScore(pos, board.Black)

	whiteScore += pieceSafetyScore(pos, board.White)
	blackScore += pieceSafetyScore(pos, board.Black)

	whiteScore -= kingSafetyPenalty(pos, board.White, phase)
	blackScore -= kingSafetyPenalty(pos, board.Black, phase)

	whiteScore += passedPawnScore(pos, board.White, phase)
	blackScore += passedPawnScore(pos, board.Black, phase)

	whiteScore -= pawnStructurePenalty(pos, board.White, phase)
	blackScore -= pawnStructurePenalty(pos, board.Black, phase)

	if pos.ActiveColor() == board.White {
		return whiteScore - blackScore
	}
	return blackScore - whiteScore
}

func gamePhase(pos *board.Position) int {
	phase := 0
	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece {
			continue
		}
		phase += phaseWeights[piece.Type()]
	}
	if phase > maxGamePhase {
		return maxGamePhase
	}
	return phase
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

func pieceSafetyScore(pos *board.Position, color int8) Score {
	enemyColor := board.White
	if color == board.White {
		enemyColor = board.Black
	}

	score := DrawScore
	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece || piece.Color() != color || piece.Type() == board.King {
			continue
		}

		attackers := attackCountOnSquare(pos, enemyColor, idx)
		defenders := attackCountOnSquare(pos, color, idx)

		if defenders > 0 {
			bonus := protectedPieceWeights[piece.Type()]
			if defenders > 1 {
				bonus += protectedPieceWeights[piece.Type()] / 2
			}
			score += bonus
		}

		if attackers == 0 {
			continue
		}

		penalty := exposedPieceWeights[piece.Type()] * Score(attackers)
		if defenders == 0 {
			penalty += exposedPieceWeights[piece.Type()]
		} else if attackers > defenders {
			penalty += exposedPieceWeights[piece.Type()] * Score(attackers-defenders)
		}
		score -= penalty
	}

	return score
}

func kingSafetyPenalty(pos *board.Position, color int8, phase int) Score {
	kingIdx := pos.BlackKingIdx()
	enemyColor := board.White
	if color == board.White {
		kingIdx = pos.WhiteKingIdx()
		enemyColor = board.Black
	}

	penalty := DrawScore
	if movegen.IsKingInCheck(pos, color) {
		penalty += phaseBlend(kingInCheckPenaltyMG, kingInCheckPenaltyEG, phase)
	}

	ring := movegen.KingRingMask(kingIdx)
	enemyAttacks := attackMap(pos, enemyColor)
	penalty += Score(bits.OnesCount64(ring&enemyAttacks)) * phaseBlend(kingRingAttackWeightMG, kingRingAttackWeightEG, phase)
	penalty += pawnShieldPenalty(pos, color, kingIdx, phase)
	return penalty
}

func passedPawnScore(pos *board.Position, color int8, phase int) Score {
	score := DrawScore
	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece || piece.Color() != color || piece.Type() != board.Pawn {
			continue
		}
		if !isPassedPawn(pos, color, idx) {
			continue
		}

		rank := board.RankFromIdx(idx)
		progress := rank
		if color == board.Black {
			progress = 7 - rank
		}

		bonus := phaseBlend(passedPawnBonusMG[progress], passedPawnBonusEG[progress], phase)
		if attackCountOnSquare(pos, color, idx) > 0 {
			bonus += bonus / 4
		}
		score += bonus
	}
	return score
}

func pawnStructurePenalty(pos *board.Position, color int8, phase int) Score {
	var fileCounts [8]int
	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece || piece.Color() != color || piece.Type() != board.Pawn {
			continue
		}
		fileCounts[board.FileFromIdx(idx)]++
	}

	penalty := DrawScore
	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece == board.NoPiece || piece.Color() != color || piece.Type() != board.Pawn {
			continue
		}

		file := board.FileFromIdx(idx)
		leftCount := 0
		rightCount := 0
		if file > 0 {
			leftCount = fileCounts[file-1]
		}
		if file < 7 {
			rightCount = fileCounts[file+1]
		}
		if leftCount == 0 && rightCount == 0 {
			penalty += phaseBlend(isolatedPawnPenaltyMG, isolatedPawnPenaltyEG, phase)
		}
	}

	for file := 0; file < 8; file++ {
		if fileCounts[file] > 1 {
			extras := fileCounts[file] - 1
			penalty += phaseBlend(doubledPawnPenaltyMG, doubledPawnPenaltyEG, phase) * Score(extras)
		}
	}

	return penalty
}

func isPassedPawn(pos *board.Position, color, idx int8) bool {
	file := board.FileFromIdx(idx)
	rank := board.RankFromIdx(idx)
	enemyPawn := board.Piece((color ^ 24) | board.Pawn)

	fileStart := maxInt8(0, file-1)
	fileEnd := minInt8(7, file+1)

	if color == board.White {
		for targetRank := rank + 1; targetRank <= 7; targetRank++ {
			for targetFile := fileStart; targetFile <= fileEnd; targetFile++ {
				if pos.PieceAt(targetRank*8+targetFile) == enemyPawn {
					return false
				}
			}
		}
		return true
	}

	for targetRank := rank - 1; targetRank >= 0; targetRank-- {
		for targetFile := fileStart; targetFile <= fileEnd; targetFile++ {
			if pos.PieceAt(targetRank*8+targetFile) == enemyPawn {
				return false
			}
		}
	}
	return true
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

func pawnShieldPenalty(pos *board.Position, color, kingIdx int8, phase int) Score {
	file := board.FileFromIdx(kingIdx)
	rank := board.RankFromIdx(kingIdx)
	penalty := DrawScore
	missingPenalty := phaseBlend(missingShieldPenaltyMG, missingShieldPenaltyEG, phase)

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
			penalty += missingPenalty
			continue
		}

		shieldIdx := shieldRank*8 + targetFile
		expectedPawn := board.Piece(color | board.Pawn)
		if pos.PieceAt(shieldIdx) != expectedPawn {
			penalty += missingPenalty
		}
	}

	return penalty
}

func minInt8(a, b int8) int8 {
	if a < b {
		return a
	}
	return b
}

func maxInt8(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}
