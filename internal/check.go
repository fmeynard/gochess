package internal

import "math/bits"

const (
	NoSlider     = 0
	BishopSlider = 1
	RookSlider   = 2
)

func isSquareAttackedByPawn(pos *Position, idx int8, kingColor int8) bool {
	rank := RankFromIdx(idx)

	var (
		mask         uint64
		opponentMask uint64
	)

	if kingColor == Black {
		rank--
		if rank == 0 || rank == -1 {
			return false
		}
		mask |= 1 << (idx - 7)
		mask |= 1 << (idx - 9)
		opponentMask = pos.OccupancyMask(White)
	} else {
		rank++
		if rank == 8 || rank == 9 {
			return false
		}
		mask |= 1 << (idx + 7)
		mask |= 1 << (idx + 9)
		opponentMask = pos.OccupancyMask(Black)
	}

	rankMask := uint64(0xFF) << (rank * 8)

	return (mask & pos.pawnBoard & rankMask & opponentMask) != 0
}

// verify if the square is attacked by knight
func isSquareAttackedByKnight(pos *Position, idx int8, kingColor int8) bool {
	return (knightAttacksMask[idx] & pos.knightBoard & pos.OpponentOccupiedMaskByPieceColor(kingColor)) != 0
}

func isSquareAttackedBySlidingPiece(pos *Position, squareIdx int8, kingColor int8) bool {
	// Bitboards for enemy pieces
	var (
		enemyColor int8
	)

	if kingColor == White {
		enemyColor = Black
	} else {
		enemyColor = White
	}

	if isRankAttackedByEnemy(pos, squareIdx, enemyColor) {
		return true
	}

	if isFileAttackedByEnemy(pos, squareIdx, enemyColor) {
		return true
	}

	if isDiagonallyAttacked(pos, squareIdx, enemyColor) {
		return true
	}

	return false
}

func firstBlockerOnRay(occupied, ray uint64, direction int8) uint64 {
	blockers := occupied & ray
	if blockers == 0 {
		return 0
	}

	switch direction {
	case East, North, NorthEast, NorthWest:
		return uint64(1) << bits.TrailingZeros64(blockers)
	default:
		return uint64(1) << (63 - bits.LeadingZeros64(blockers))
	}
}

func scanRayForAttack(pos *Position, startIdx int8, dirIdx int, direction int8, targets uint64) bool {
	rayMask := sliderAttackMasks[startIdx][dirIdx]
	firstBlocker := firstBlockerOnRay(pos.occupied, rayMask, direction)
	if firstBlocker == 0 {
		return false
	}

	return (targets & firstBlocker) != 0
}

// isDiagonallyAttacked determines if the position at index is diagonally attacked by the enemy.
func isDiagonallyAttacked(pos *Position, idx int8, enemyColor int8) bool {
	targets := (pos.queenBoard | pos.bishopBoard) & pos.OccupancyMask(enemyColor)

	return scanRayForAttack(pos, idx, 4, SouthWest, targets) ||
		scanRayForAttack(pos, idx, 5, SouthEast, targets) ||
		scanRayForAttack(pos, idx, 6, NorthWest, targets) ||
		scanRayForAttack(pos, idx, 7, NorthEast, targets)
}

func isRankAttackedByEnemy(pos *Position, index int8, enemyColor int8) bool {
	targets := (pos.rookBoard | pos.queenBoard) & pos.OccupancyMask(enemyColor)

	return scanRayForAttack(pos, index, 0, West, targets) ||
		scanRayForAttack(pos, index, 1, East, targets)
}

func isFileAttackedByEnemy(pos *Position, index int8, enemyColor int8) bool {
	targets := (pos.rookBoard | pos.queenBoard) & pos.OccupancyMask(enemyColor)

	return scanRayForAttack(pos, index, 3, North, targets) ||
		scanRayForAttack(pos, index, 2, South, targets)
}

// verify if the square is attacked by king
func isSquareAttackedByKing(pos *Position, idx int8, kingColor int8) bool {
	return (kingAttacksMask[idx] & pos.kingBoard & pos.OpponentOccupiedMaskByPieceColor(kingColor)) != 0
}

// IsKingInCheck verifies if the king at the given index is "check".
func IsKingInCheck(pos *Position, kingColor int8) bool {
	var (
		kingIdx int8
	)
	if kingColor == White {
		kingIdx = pos.whiteKingIdx
		if pos.whiteKingSafety != NotCalculated {
			return pos.whiteKingSafety == KingIsCheck
		}
	} else {
		kingIdx = pos.blackKingIdx
		if pos.blackKingSafety != NotCalculated {
			return pos.blackKingSafety == KingIsCheck
		}
	}

	isAttacked := false
	if isSquareAttackedByPawn(pos, kingIdx, kingColor) ||
		isSquareAttackedByKnight(pos, kingIdx, kingColor) ||
		isSquareAttackedBySlidingPiece(pos, kingIdx, kingColor) ||
		isSquareAttackedByKing(pos, kingIdx, kingColor) {
		isAttacked = true
	}

	if kingColor == White {
		if isAttacked {
			pos.whiteKingSafety = KingIsCheck
		} else {
			pos.whiteKingSafety = KingIsSafe
		}
	} else {
		if isAttacked {
			pos.blackKingSafety = KingIsCheck
		} else {
			pos.blackKingSafety = KingIsSafe
		}
	}

	return isAttacked
}
