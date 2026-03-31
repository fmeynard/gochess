package movegen

import board "chessV2/internal/board"

func PseudoLegalTargetsMask(pos *board.Position, piece board.Piece, idx int8) uint64 {
	ensureAttackTables()

	friendlyOcc := pos.OccupancyMask(piece.Color())
	switch piece.Type() {
	case board.Knight:
		return knightAttacksMask[idx] &^ friendlyOcc
	case board.Bishop:
		return bishopAttacksMagic(idx, pos.Occupied()) &^ friendlyOcc
	case board.Rook:
		return rookAttacksMagic(idx, pos.Occupied()) &^ friendlyOcc
	case board.Queen:
		return (rookAttacksMagic(idx, pos.Occupied()) | bishopAttacksMagic(idx, pos.Occupied())) &^ friendlyOcc
	default:
		return 0
	}
}
