package internal

var (
	zobristPieceTable [32][64]uint64
	zobristCastle     [16]uint64
	zobristEnPassant  [65]uint64
	zobristSideToMove uint64
)

func init() {
	var seed uint64 = 0x9e3779b97f4a7c15
	next := func() uint64 {
		seed += 0x9e3779b97f4a7c15
		z := seed
		z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
		z = (z ^ (z >> 27)) * 0x94d049bb133111eb
		return z ^ (z >> 31)
	}

	for p := range zobristPieceTable {
		for sq := range zobristPieceTable[p] {
			zobristPieceTable[p][sq] = next()
		}
	}
	for i := range zobristCastle {
		zobristCastle[i] = next()
	}
	for i := range zobristEnPassant {
		zobristEnPassant[i] = next()
	}
	zobristSideToMove = next()
}

func zobristPieceKey(piece Piece, idx int8) uint64 {
	if piece == NoPiece || idx < 0 || idx >= 64 {
		return 0
	}
	return zobristPieceTable[int(piece)][idx]
}

func zobristCastleKey(whiteRights, blackRights int8) uint64 {
	key := int((whiteRights & 0x3) | ((blackRights & 0x3) << 2))
	return zobristCastle[key]
}

func zobristEPKey(epIdx int8) uint64 {
	if epIdx == NoEnPassant {
		return zobristEnPassant[64]
	}
	return zobristEnPassant[epIdx]
}

func computeZobristKey(pos *Position) uint64 {
	var key uint64

	for idx, piece := range pos.board {
		if piece != NoPiece {
			key ^= zobristPieceKey(piece, int8(idx))
		}
	}

	key ^= zobristCastleKey(pos.whiteCastleRights, pos.blackCastleRights)
	key ^= zobristEPKey(pos.enPassantIdx)
	if pos.activeColor == Black {
		key ^= zobristSideToMove
	}

	return key
}
