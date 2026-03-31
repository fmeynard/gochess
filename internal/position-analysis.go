package internal

import "math/bits"

type positionAnalysis struct {
	inCheck      bool
	checkerCount int
	evasionMask  uint64
	pinnedMask   uint64
	pinCount     int
	pinSquares   [8]int8
	pinRayMasks  [8]uint64
}

func computePositionAnalysis(pos *Position, kingIdx int8, friendlyOcc, enemyOcc uint64) positionAnalysis {
	var info positionAnalysis
	checkers := uint64(0)
	blockMask := uint64(0)

	var pawnColorIdx int
	if pos.activeColor == White {
		pawnColorIdx = 1
	} else {
		pawnColorIdx = 0
	}
	checkers |= pawnAttacksBy[pawnColorIdx][kingIdx] & pos.pawnBoard & enemyOcc
	checkers |= knightAttacksMask[kingIdx] & pos.knightBoard & enemyOcc

	rookQueens := (pos.rookBoard | pos.queenBoard) & enemyOcc
	bishopQueens := (pos.bishopBoard | pos.queenBoard) & enemyOcc

	for dirIdx := 0; dirIdx < 8; dirIdx++ {
		var sliders uint64
		if dirIdx < 4 {
			sliders = rookQueens
		} else {
			sliders = bishopQueens
		}
		if sliders == 0 {
			continue
		}
		ray := sliderAttackMasks[kingIdx][dirIdx]
		if ray == 0 {
			continue
		}
		dir := rayDirections[dirIdx]
		first := firstBlockerOnRay(pos.occupied, ray, dir)
		if first == 0 {
			continue
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
				info.pinSquares[info.pinCount] = int8(bits.TrailingZeros64(first))
				info.pinRayMasks[info.pinCount] = ray
				info.pinCount++
			}
		}
	}

	info.checkerCount = bits.OnesCount64(checkers)
	info.inCheck = info.checkerCount > 0
	if info.checkerCount == 1 {
		info.evasionMask = checkers | blockMask
	}
	return info
}

func (pa *positionAnalysis) pinRayFor(sq int8) uint64 {
	for i := 0; i < pa.pinCount; i++ {
		if pa.pinSquares[i] == sq {
			return pa.pinRayMasks[i]
		}
	}
	return 0
}
