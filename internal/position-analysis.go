package internal

import "math/bits"

type positionAnalysis struct {
	inCheck      bool
	checkerCount int
	evasionMask  uint64
	pinnedMask   uint64
	pinRayBySq   [64]uint64
}

func computePositionAnalysis(pos *Position, kingIdx int8, friendlyOcc, enemyOcc uint64) positionAnalysis {
	var info positionAnalysis
	checkers := uint64(0)
	blockMask := uint64(0)

	pawnAttackMask := pawnAttacksBy[0][kingIdx]
	if pos.activeColor == White {
		pawnAttackMask = pawnAttacksBy[1][kingIdx]
	}
	checkers |= pawnAttackMask & pos.pawnBoard & enemyOcc
	checkers |= knightAttacksMask[kingIdx] & pos.knightBoard & enemyOcc

	rookQueens := (pos.rookBoard | pos.queenBoard) & enemyOcc
	bishopQueens := (pos.bishopBoard | pos.queenBoard) & enemyOcc

	processDir := func(dirIdx int, sliders uint64, dir int8) {
		if sliders == 0 {
			return
		}
		ray := sliderAttackMasks[kingIdx][dirIdx]
		if ray == 0 {
			return
		}
		first := firstBlockerOnRay(pos.occupied, ray, dir)
		if first == 0 {
			return
		}
		if first&enemyOcc != 0 {
			if first&sliders != 0 {
				checkers |= first
				checkerIdx := int8(bits.TrailingZeros64(first))
				blockMask |= ray ^ sliderAttackMasks[checkerIdx][dirIdx]
			}
		} else if first&friendlyOcc != 0 {
			second := firstBlockerOnRay(pos.occupied&^first, ray, dir)
			if second != 0 && second&sliders != 0 {
				info.pinnedMask |= first
				info.pinRayBySq[bits.TrailingZeros64(first)] = ray
			}
		}
	}

	processDir(0, rookQueens, West)
	processDir(1, rookQueens, East)
	processDir(2, rookQueens, South)
	processDir(3, rookQueens, North)
	processDir(4, bishopQueens, SouthWest)
	processDir(5, bishopQueens, SouthEast)
	processDir(6, bishopQueens, NorthWest)
	processDir(7, bishopQueens, NorthEast)

	info.checkerCount = bits.OnesCount64(checkers)
	info.inCheck = info.checkerCount > 0
	if info.checkerCount == 1 {
		info.evasionMask = checkers | blockMask
	}
	return info
}

func (pa *positionAnalysis) pinRayFor(sq int8) uint64 {
	return pa.pinRayBySq[sq]
}
