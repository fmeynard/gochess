package movegen

import "math/bits"

type positionAnalysis struct {
	inCheck      bool
	checkerCount uint8
	evasionMask  uint64
	pinnedMask   uint64
	pinRayBySq   [64]uint64
}

func computePositionAnalysis(pos *Position, kingIdx int8, friendlyOcc, enemyOcc uint64, info *positionAnalysis) {
	*info = positionAnalysis{}
	checkers := uint64(0)
	blockMask := uint64(0)

	pawnAttackMask := pawnAttacksBy[0][kingIdx]
	if pos.ActiveColor() == White {
		pawnAttackMask = pawnAttacksBy[1][kingIdx]
	}
	checkers |= pawnAttackMask & pos.PawnBoard() & enemyOcc
	checkers |= knightAttacksMask[kingIdx] & pos.KnightBoard() & enemyOcc

	rookQueens := (pos.RookBoard() | pos.QueenBoard()) & enemyOcc
	bishopQueens := (pos.BishopBoard() | pos.QueenBoard()) & enemyOcc

	processCandidates := func(candidates uint64) {
		for candidates != 0 {
			pinnerIdx := int8(bits.TrailingZeros64(candidates))
			pinnerMask := uint64(1) << pinnerIdx
			candidates &^= pinnerMask

			between := betweenMasks[kingIdx][pinnerIdx]
			blockers := between & pos.Occupied()
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

	info.checkerCount = uint8(bits.OnesCount64(checkers))
	info.inCheck = info.checkerCount > 0
	if info.checkerCount == 1 {
		info.evasionMask = checkers | blockMask
	}
}
