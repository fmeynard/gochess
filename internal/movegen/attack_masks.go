package movegen

import board "chessV2/internal/board"

func PieceAttackMask(pos *board.Position, piece board.Piece, idx int8) uint64 {
	ensureAttackTables()

	switch piece.Type() {
	case board.Pawn:
		return pawnAttackMask(piece.Color(), idx)
	case board.Knight:
		return knightAttacksMask[idx]
	case board.Bishop:
		return bishopAttacksMagic(idx, pos.Occupied())
	case board.Rook:
		return rookAttacksMagic(idx, pos.Occupied())
	case board.Queen:
		return rookAttacksMagic(idx, pos.Occupied()) | bishopAttacksMagic(idx, pos.Occupied())
	case board.King:
		return kingAttacksMask[idx]
	default:
		return 0
	}
}

func KingRingMask(idx int8) uint64 {
	ensureAttackTables()
	return kingAttacksMask[idx]
}

func pawnAttackMask(color, idx int8) uint64 {
	file := board.FileFromIdx(idx)
	rank := board.RankFromIdx(idx)
	mask := uint64(0)

	if color == board.White {
		if file > 0 && rank < 7 {
			mask |= uint64(1) << ((rank+1)*8 + (file - 1))
		}
		if file < 7 && rank < 7 {
			mask |= uint64(1) << ((rank+1)*8 + (file + 1))
		}
		return mask
	}

	if file > 0 && rank > 0 {
		mask |= uint64(1) << ((rank-1)*8 + (file - 1))
	}
	if file < 7 && rank > 0 {
		mask |= uint64(1) << ((rank-1)*8 + (file + 1))
	}
	return mask
}
