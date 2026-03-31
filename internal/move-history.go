package internal

type MoveHistory struct {
	zobristKey          uint64
	whiteKingAffectMask uint64
	blackKingAffectMask uint64
	move                Move
	capturedPiece       Piece
	captureIdx          int8
	meta                uint32
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

func encodeKingSafety(v int8) uint32 {
	switch v {
	case KingIsSafe:
		return 1
	case KingIsCheck:
		return 2
	default:
		return 0
	}
}

func decodeKingSafety(v uint32) int8 {
	switch v {
	case 1:
		return KingIsSafe
	case 2:
		return KingIsCheck
	default:
		return NotCalculated
	}
}

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
	return int8((h.meta >> metaWhiteKingShift) & 0x3F)
}

func (h MoveHistory) blackKingIdx() int8 {
	return int8((h.meta >> metaBlackKingShift) & 0x3F)
}

func (h MoveHistory) enPassantIdx() int8 {
	return int8((h.meta>>metaEnPassantShift)&0x7F) - 1
}

func (h MoveHistory) whiteCastleRights() int8 {
	return int8((h.meta >> metaWhiteCastleShift) & 0x3)
}

func (h MoveHistory) blackCastleRights() int8 {
	return int8((h.meta >> metaBlackCastleShift) & 0x3)
}

func (h MoveHistory) whiteKingSafety() int8 {
	return decodeKingSafety((h.meta >> metaWhiteSafetyShift) & 0x3)
}

func (h MoveHistory) blackKingSafety() int8 {
	return decodeKingSafety((h.meta >> metaBlackSafetyShift) & 0x3)
}
