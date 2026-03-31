package movegen

import (
	board "chessV2/internal/board"
	"math/bits"
)

const (
	NoSlider     = 0
	BishopSlider = 1
	RookSlider   = 2
)

var pawnAttacksBy [2][64]uint64

func init() {
	for sq := int8(0); sq < 64; sq++ {
		rank, file := board.RankAndFile(sq)
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
		return (pawnAttacksBy[1][idx] & pos.PawnBoard() & pos.BlackOccupied()) != 0
	}
	return (pawnAttacksBy[0][idx] & pos.PawnBoard() & pos.WhiteOccupied()) != 0
}

// verify if the square is attacked by knight
func isSquareAttackedByKnight(pos *Position, idx int8, kingColor int8) bool {
	return (knightAttacksMask[idx] & pos.KnightBoard() & pos.OpponentOccupiedMaskByPieceColor(kingColor)) != 0
}

func isSquareAttackedBySlidingPiece(pos *Position, squareIdx int8, kingColor int8) bool {
	var enemyOcc uint64
	if kingColor == White {
		enemyOcc = pos.BlackOccupied()
	} else {
		enemyOcc = pos.WhiteOccupied()
	}

	rookQueenTargets := (pos.RookBoard() | pos.QueenBoard()) & enemyOcc
	if rookQueenTargets != 0 && rookAttacksMagic(squareIdx, pos.Occupied())&rookQueenTargets != 0 {
		return true
	}

	bishopQueenTargets := (pos.BishopBoard() | pos.QueenBoard()) & enemyOcc
	return bishopQueenTargets != 0 && bishopAttacksMagic(squareIdx, pos.Occupied())&bishopQueenTargets != 0
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
	firstBlocker := firstBlockerOnRay(pos.Occupied(), rayMask, direction)
	if firstBlocker == 0 {
		return false
	}

	return (targets & firstBlocker) != 0
}

// isDiagonallyAttacked determines if the position at index is diagonally attacked by the enemy.
func isDiagonallyAttacked(pos *Position, idx int8, enemyColor int8) bool {
	targets := (pos.QueenBoard() | pos.BishopBoard()) & pos.OccupancyMask(enemyColor)

	return scanRayForAttack(pos, idx, 4, SouthWest, targets) ||
		scanRayForAttack(pos, idx, 5, SouthEast, targets) ||
		scanRayForAttack(pos, idx, 6, NorthWest, targets) ||
		scanRayForAttack(pos, idx, 7, NorthEast, targets)
}

func isRankAttackedByEnemy(pos *Position, index int8, enemyColor int8) bool {
	targets := (pos.RookBoard() | pos.QueenBoard()) & pos.OccupancyMask(enemyColor)

	return scanRayForAttack(pos, index, 0, West, targets) ||
		scanRayForAttack(pos, index, 1, East, targets)
}

func isFileAttackedByEnemy(pos *Position, index int8, enemyColor int8) bool {
	targets := (pos.RookBoard() | pos.QueenBoard()) & pos.OccupancyMask(enemyColor)

	return scanRayForAttack(pos, index, 3, North, targets) ||
		scanRayForAttack(pos, index, 2, South, targets)
}

// verify if the square is attacked by king
func isSquareAttackedByKing(pos *Position, idx int8, kingColor int8) bool {
	return (kingAttacksMask[idx] & pos.KingBoard() & pos.OpponentOccupiedMaskByPieceColor(kingColor)) != 0
}

// isSquareAttacked checks if a square is attacked by pieces of attackerColor.
// occ is the occupancy used for sliding piece ray calculations.
func isSquareAttacked(pos *Position, sq int8, attackerColor int8, occ uint64) bool {
	attackerOcc := pos.BlackOccupied()
	pawnAttackMask := pawnAttacksBy[1][sq]
	if attackerColor == White {
		attackerOcc = pos.WhiteOccupied()
		pawnAttackMask = pawnAttacksBy[0][sq]
	}
	if pawnAttackMask&pos.PawnBoard()&attackerOcc != 0 {
		return true
	}
	if knightAttacksMask[sq]&pos.KnightBoard()&attackerOcc != 0 {
		return true
	}
	rookQueens := (pos.RookBoard() | pos.QueenBoard()) & attackerOcc
	if rookQueens != 0 && rookAttacksMagic(sq, occ)&rookQueens != 0 {
		return true
	}
	bishopQueens := (pos.BishopBoard() | pos.QueenBoard()) & attackerOcc
	if bishopQueens != 0 && bishopAttacksMagic(sq, occ)&bishopQueens != 0 {
		return true
	}
	return kingAttacksMask[sq]&pos.KingBoard()&attackerOcc != 0
}

// IsKingInCheck verifies if the king at the given index is "check".
func IsKingInCheck(pos *Position, kingColor int8) bool {
	var (
		kingIdx int8
	)
	if kingColor == White {
		kingIdx = pos.WhiteKingIdx()
		if pos.KingSafety(White) != NotCalculated {
			return pos.KingSafety(White) == KingIsCheck
		}
	} else {
		kingIdx = pos.BlackKingIdx()
		if pos.KingSafety(Black) != NotCalculated {
			return pos.KingSafety(Black) == KingIsCheck
		}
	}

	isAttacked := isSquareAttackedByPawn(pos, kingIdx, kingColor) ||
		isSquareAttackedByKnight(pos, kingIdx, kingColor) ||
		isSquareAttackedBySlidingPiece(pos, kingIdx, kingColor)

	if kingColor == White {
		if isAttacked {
			pos.SetKingSafety(White, KingIsCheck)
		} else {
			pos.SetKingSafety(White, KingIsSafe)
		}
	} else {
		if isAttacked {
			pos.SetKingSafety(Black, KingIsCheck)
		} else {
			pos.SetKingSafety(Black, KingIsSafe)
		}
	}

	return isAttacked
}
