package internal

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

func scanRayForAttack(pos *Position, startIdx int8, direction int8, enemyColor int8, targets uint64) bool {
	for step := int8(1); step < 8; step++ {
		nextIdx := startIdx + step*direction
		if nextIdx < 0 || nextIdx >= 64 || !isSameLineOrRow(startIdx, nextIdx, direction) {
			return false
		}

		piece := pos.PieceAt(nextIdx)
		if piece == NoPiece {
			continue
		}

		if piece.Color() != enemyColor {
			return false
		}

		return (targets & (uint64(1) << nextIdx)) != 0
	}

	return false
}

// isDiagonallyAttacked determines if the position at index is diagonally attacked by the enemy.
func isDiagonallyAttacked(pos *Position, idx int8, enemyColor int8) bool {
	targets := (pos.queenBoard | pos.bishopBoard) & pos.OccupancyMask(enemyColor)

	return scanRayForAttack(pos, idx, SouthWest, enemyColor, targets) ||
		scanRayForAttack(pos, idx, SouthEast, enemyColor, targets) ||
		scanRayForAttack(pos, idx, NorthWest, enemyColor, targets) ||
		scanRayForAttack(pos, idx, NorthEast, enemyColor, targets)
}

func isRankAttackedByEnemy(pos *Position, index int8, enemyColor int8) bool {
	targets := (pos.rookBoard | pos.queenBoard) & pos.OccupancyMask(enemyColor)

	return scanRayForAttack(pos, index, West, enemyColor, targets) ||
		scanRayForAttack(pos, index, East, enemyColor, targets)
}

func isFileAttackedByEnemy(pos *Position, index int8, enemyColor int8) bool {
	targets := (pos.rookBoard | pos.queenBoard) & pos.OccupancyMask(enemyColor)

	return scanRayForAttack(pos, index, North, enemyColor, targets) ||
		scanRayForAttack(pos, index, South, enemyColor, targets)
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
