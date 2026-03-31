package internal

type MoveHistory struct {
	zobristKey          uint64
	whiteKingAffectMask uint64
	blackKingAffectMask uint64
	move                Move
	capturedPiece       Piece
	captureIdx          int8
	// packedState layout:
	// bits  0.. 5: previous white king square
	// bits  6..11: previous black king square
	// bits 12..18: previous en-passant square encoded as idx+1, 0 means "none"
	// bits 19..20: previous white castle rights
	// bits 21..22: previous black castle rights
	// bits 23..24: previous white king safety cache encoded via encodeKingSafety
	// bits 25..26: previous black king safety cache encoded via encodeKingSafety
	packedState uint32
}

const (
	metaWhiteKingShift   = 0
	metaBlackKingShift   = 6
	metaEnPassantShift   = 12
	metaWhiteCastleShift = 19
	metaBlackCastleShift = 21
	metaWhiteSafetyShift = 23
	metaBlackSafetyShift = 25
)

func packMoveHistoryMeta(pos *Position) uint32 {
	ep := uint32(pos.enPassantIdx + 1)
	return uint32(pos.whiteKingIdx)<<metaWhiteKingShift |
		uint32(pos.blackKingIdx)<<metaBlackKingShift |
		ep<<metaEnPassantShift |
		uint32(pos.whiteCastleRights)<<metaWhiteCastleShift |
		uint32(pos.blackCastleRights)<<metaBlackCastleShift |
		encodeKingSafety(pos.whiteKingSafety)<<metaWhiteSafetyShift |
		encodeKingSafety(pos.blackKingSafety)<<metaBlackSafetyShift
}

func (h MoveHistory) whiteKingIdx() int8 {
	return int8((h.packedState >> metaWhiteKingShift) & 0x3F)
}

func (h MoveHistory) blackKingIdx() int8 {
	return int8((h.packedState >> metaBlackKingShift) & 0x3F)
}

func (h MoveHistory) enPassantIdx() int8 {
	return int8((h.packedState>>metaEnPassantShift)&0x7F) - 1
}

func (h MoveHistory) whiteCastleRights() int8 {
	return int8((h.packedState >> metaWhiteCastleShift) & 0x3)
}

func (h MoveHistory) blackCastleRights() int8 {
	return int8((h.packedState >> metaBlackCastleShift) & 0x3)
}

func (h MoveHistory) whiteKingSafety() int8 {
	return decodeKingSafety((h.packedState >> metaWhiteSafetyShift) & 0x3)
}

func (h MoveHistory) blackKingSafety() int8 {
	return decodeKingSafety((h.packedState >> metaBlackSafetyShift) & 0x3)
}
