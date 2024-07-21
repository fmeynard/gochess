package internal

const (
	NoSlider     = 0
	BishopSlider = 1
	RookSlider   = 2
)

// verify if the square is attacked by pawn
func isSquareAttackedByPawn(pos *Position, idx int8, kingColor int8) bool {
	rank, file := RankAndFile(idx)
	pawnAttacks := [2][2]int8{{-1, 1}, {1, 1}}
	if kingColor != White {
		pawnAttacks = [2][2]int8{{-1, -1}, {1, -1}}
	}

	for _, attack := range pawnAttacks {
		newFile := file + attack[0]
		newRank := rank + attack[1]
		if !isOnBoard(newFile, newRank) {
			continue
		}

		endIdx := newRank*8 + newFile
		piece := pos.board[endIdx]
		if piece.Type() == Pawn && piece.Color() != kingColor {
			return true
		}
	}

	return false
}

// verify if the square is attacked by knight
func isSquareAttackedByKnight(pos *Position, idx int8, kingColor int8) bool {
	return (knightAttacksMask[idx] & pos.knightBoard & pos.OpponentOccupiedMaskByPieceColor(kingColor)) != 0
}

func isSquareAttackedBySlidingPiece(pos *Position, squareIdx int8, kingColor int8) bool {
	// Bitboards for enemy pieces
	var (
		enemyQueens uint64
		//enemyRooks   uint64
		enemyBishops uint64
		enemyColor   int8
	)

	if kingColor == White {
		enemyColor = Black
		enemyQueens = pos.queenBoard & pos.blackOccupied
		//enemyRooks = pos.rookBoard & pos.blackOccupied
		enemyBishops = pos.bishopBoard & pos.blackOccupied
	} else {
		enemyColor = White
		enemyQueens = pos.queenBoard & pos.whiteOccupied
		//enemyRooks = pos.rookBoard & pos.whiteOccupied
		enemyBishops = pos.bishopBoard & pos.whiteOccupied
	}

	if isRankAttackedByEnemy(pos, squareIdx, enemyColor) {
		return true
	}

	if isFileAttackedByEnemy(pos, squareIdx, enemyColor) {
		return true
	}

	// Check for threats along diagonals
	if diagonalAttacks(pos, squareIdx)&(enemyBishops|enemyQueens) != 0 {
		return true
	}

	return false
}

func diagonalAttacks(pos *Position, index int8) uint64 {
	row, col := index/8, index%8
	var attacks uint64

	// Trace each diagonal until blocked or end of board
	for _, dir := range BishopDirections {
		var boardLimit int8
		switch dir {
		case UpRight:
			boardLimit = min(row, 7-col)
		case UpLeft:
			boardLimit = min(row, col)
		case DownRight:
			boardLimit = min(7-row, 7-col)
		case DownLeft:
			boardLimit = min(7-row, col)
		}
		for step := int8(1); step <= boardLimit; step++ {
			targetIdx := index + dir*step
			if targetIdx < 0 || targetIdx >= 64 {
				break // Safety check for boundaries, should never actually trigger
			}
			attacks |= 1 << targetIdx

			// If this square is occupied, break, as it blocks further attacks in this direction
			if pos.occupied&(1<<targetIdx) != 0 {
				break
			}
		}
	}

	return attacks
}

func isRankAttackedByEnemy(pos *Position, index int8, enemyColor int8) bool {
	rank := index / 8
	rankBase := rank * 8
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
