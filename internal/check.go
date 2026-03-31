package internal

import "math/bits"

const (
	NoSlider     = 0
	BishopSlider = 1
	RookSlider   = 2
)

var pawnAttacksBy [2][64]uint64

func init() {
	for sq := int8(0); sq < 64; sq++ {
		rank, file := RankAndFile(sq)
		if rank > 0 {
			if file > 0 {
				pawnAttacksBy[0][sq] |= 1 << (sq - 9)
			}
			if file < 7 {
				pawnAttacksBy[0][sq] |= 1 << (sq - 7)
			}
		}
		if rank < 7 {
			if file > 0 {
				pawnAttacksBy[1][sq] |= 1 << (sq + 7)
			}
			if file < 7 {
				pawnAttacksBy[1][sq] |= 1 << (sq + 9)
			}
		}
	}
}

func isSquareAttackedByPawn(pos *Position, idx int8, kingColor int8) bool {
	if kingColor == White {
		return (pawnAttacksBy[1][idx] & pos.pawnBoard & pos.blackOccupied) != 0
	}
	return (pawnAttacksBy[0][idx] & pos.pawnBoard & pos.whiteOccupied) != 0
}

// verify if the square is attacked by knight
func isSquareAttackedByKnight(pos *Position, idx int8, kingColor int8) bool {
	return (knightAttacksMask[idx] & pos.knightBoard & pos.OpponentOccupiedMaskByPieceColor(kingColor)) != 0
}

func isSquareAttackedBySlidingPiece(pos *Position, squareIdx int8, kingColor int8) bool {
	var enemyOcc uint64
	if kingColor == White {
		enemyOcc = pos.blackOccupied
	} else {
		enemyOcc = pos.whiteOccupied
	}

	rookQueenTargets := (pos.rookBoard | pos.queenBoard) & enemyOcc
	if rookQueenTargets != 0 && rookAttacksMagic(squareIdx, pos.occupied)&rookQueenTargets != 0 {
		return true
	}

	bishopQueenTargets := (pos.bishopBoard | pos.queenBoard) & enemyOcc
	return bishopQueenTargets != 0 && bishopAttacksMagic(squareIdx, pos.occupied)&bishopQueenTargets != 0
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

// isSquareAttacked checks if a square is attacked by pieces of attackerColor.
// occ is the occupancy used for sliding piece ray calculations.
func isSquareAttacked(pos *Position, sq int8, attackerColor int8, occ uint64) bool {
	attackerOcc := pos.blackOccupied
	pawnAttackMask := pawnAttacksBy[1][sq]
	if attackerColor == White {
		attackerOcc = pos.whiteOccupied
		pawnAttackMask = pawnAttacksBy[0][sq]
	}
	if pawnAttackMask&pos.pawnBoard&attackerOcc != 0 {
		return true
	}
	if knightAttacksMask[sq]&pos.knightBoard&attackerOcc != 0 {
		return true
	}
	rookQueens := (pos.rookBoard | pos.queenBoard) & attackerOcc
	if rookQueens != 0 && rookAttacksMagic(sq, occ)&rookQueens != 0 {
		return true
	}
	bishopQueens := (pos.bishopBoard | pos.queenBoard) & attackerOcc
	if bishopQueens != 0 && bishopAttacksMagic(sq, occ)&bishopQueens != 0 {
		return true
	}
	return kingAttacksMask[sq]&pos.kingBoard&attackerOcc != 0
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

	isAttacked := isSquareAttackedByPawn(pos, kingIdx, kingColor) ||
		isSquareAttackedByKnight(pos, kingIdx, kingColor) ||
		isSquareAttackedBySlidingPiece(pos, kingIdx, kingColor)

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
