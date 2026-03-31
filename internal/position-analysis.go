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

	processCandidates := func(candidates uint64) {
		for candidates != 0 {
			pinnerIdx := int8(bits.TrailingZeros64(candidates))
			pinnerMask := uint64(1) << pinnerIdx
			candidates &^= pinnerMask

			between := betweenMasks[kingIdx][pinnerIdx]
			blockers := between & pos.occupied
			if blockers == 0 {
				checkers |= pinnerMask
				blockMask |= between
				continue
			}
			if blockers&enemyOcc != 0 || bits.OnesCount64(blockers) != 1 {
				continue
			}
			pinned := blockers & friendlyOcc
			if pinned == 0 {
				continue
			}
			info.pinnedMask |= pinned
			info.pinRayBySq[bits.TrailingZeros64(pinned)] = between | pinnerMask
		}
	}

	processCandidates(orthogonalAttacksMask[kingIdx] & rookQueens)
	processCandidates(diagonalCombinedAttacksMask[kingIdx] & bishopQueens)

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
