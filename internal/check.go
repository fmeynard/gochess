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

func calculateActualAttacks(pos *Position, idx int8, dir int) uint64 {
	attackMask := diagonalAttacksMask[idx][dir]
	blocker := attackMask & pos.occupied

	lsb, msb := leastSignificantOne(blocker), mostSignificantBit(blocker)
	if msb == lsb || blocker == 0 {
		return attackMask
	}

	switch dir {
	case 0: //SouthWest
		return attackMask & ^((1 << (msb + 1)) - 1)
	case 1: //SouthEast
		return attackMask & ^((1 << (lsb + 1)) - 1)
	case 2: //NorthWest
		return attackMask&(1<<msb) - 1
	case 3: //NorthEast
		return attackMask&(1<<lsb) - 1
	}

	panic("Unexpected direction value")
}

// isDiagonallyAttacked determines if the position at index is diagonally attacked by the enemy.
func isDiagonallyAttacked(pos *Position, idx int8, enemyColor int8) bool {
	enemyDiagonalSliders := pos.OccupancyMask(enemyColor) & (pos.queenBoard | pos.bishopBoard)

	if enemyDiagonalSliders&diagonalCombinedAttacksMask[idx] == 0 {
		return false
	}

	for dir := 0; dir < 4; dir++ {
		if calculateActualAttacks(pos, idx, dir)&enemyDiagonalSliders != 0 {
			return true
		}
	}
	return false
}

func isRankAttackedByEnemy(pos *Position, index int8, enemyColor int8) bool {
	rankBase := RankFromIdx(index) * 8
	rankMask := uint64(0xFF) << (rankBase) // Mask for the entire rank

	// Combine enemy rooks and queens into one bitboard
	enemyRookOrQueen := (pos.rookBoard | pos.queenBoard) & pos.OccupancyMask(enemyColor) & rankMask

	if enemyRookOrQueen&rankMask == 0 {
		return false
	}

	rankOccupied := pos.occupied & rankMask

	msb := mostSignificantBit(enemyRookOrQueen)
	if index > msb {
		leftMask := (uint64(1) << index) - 1
		blockersLeft := rankOccupied & leftMask
		if msb >= mostSignificantBit(blockersLeft) {
			return true
		}
	}

	lsb := leastSignificantOne(enemyRookOrQueen)
	if index < lsb {
		rightMask := ^((uint64(1) << (index + 1)) - 1)
		blockersRight := rankOccupied & rightMask

		if lsb <= leastSignificantOne(blockersRight) {
			return true
		}
	}

	return false
}

func isFileAttackedByEnemy(pos *Position, index int8, enemyColor int8) bool {
	file := index % 8
	fileMask := uint64(0x0101010101010101) << file // Mask for the entire rank

	enemyRookOrQueen := (pos.rookBoard | pos.queenBoard) & pos.OccupancyMask(enemyColor) & fileMask

	if fileMask&enemyRookOrQueen == 0 {
		return false
	}

	fileOccupied := pos.occupied & fileMask

	topMask := uint64(0xFFFFFFFFFFFFFFFF) << (index + 8)

	msb := mostSignificantBit(enemyRookOrQueen)
	lsb := leastSignificantOne(enemyRookOrQueen & topMask)
	if index < msb {
		blockersTop := fileOccupied & topMask
		if lsb <= leastSignificantOne(blockersTop) {
			return true
		}
	}

	if index > msb {

		bottomMask := uint64(0xFFFFFFFFFFFFFFFF) >> (64 - index)
		blockersBottom := fileOccupied & bottomMask

		if msb >= mostSignificantBit(blockersBottom) {
			return true
		}
	}

	return false
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
