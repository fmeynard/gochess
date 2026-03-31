package chess

type moveApplier interface {
	MakeMove(pos *Position, move Move) MoveHistory
	UnMakeMove(pos *Position, history MoveHistory)
}

type PlainPositionUpdater struct {
	moveGenerator *PseudoLegalMoveGenerator
}

func kingAffectMask(kingIdx int8) uint64 {
	return queenAttacksMask[kingIdx] | knightAttacksMask[kingIdx] | (uint64(1) << kingIdx)
}

func NewPositionUpdater(moveGenerator *PseudoLegalMoveGenerator) moveApplier {
	return NewZobristPositionUpdater(NewPlainPositionUpdater(moveGenerator))
}

func NewPlainPositionUpdater(moveGenerator *PseudoLegalMoveGenerator) *PlainPositionUpdater {
	return &PlainPositionUpdater{moveGenerator: moveGenerator}
}

func castleRookSquares(activeColor int8, kingEndIdx int8) (int8, int8) {
	if activeColor == White {
		if kingEndIdx == G1 {
			return H1, F1
		}
		return A1, D1
	}

	if kingEndIdx == G8 {
		return H8, F8
	}

	return A8, D8
}

func (updater *PlainPositionUpdater) invalidateKingSafetyCaches(pos *Position) {
	pos.whiteKingSafety = NotCalculated
	pos.blackKingSafety = NotCalculated
}

func xorPieceBoard(pos *Position, pieceType int8, mask uint64) {
	switch pieceType {
	case King:
		pos.kingBoard ^= mask
	case Queen:
		pos.queenBoard ^= mask
	case Rook:
		pos.rookBoard ^= mask
	case Bishop:
		pos.bishopBoard ^= mask
	case Knight:
		pos.knightBoard ^= mask
	case Pawn:
		pos.pawnBoard ^= mask
	}
}

func clearPieceBoard(pos *Position, pieceType int8, mask uint64) {
	switch pieceType {
	case King:
		pos.kingBoard &^= mask
	case Queen:
		pos.queenBoard &^= mask
	case Rook:
		pos.rookBoard &^= mask
	case Bishop:
		pos.bishopBoard &^= mask
	case Knight:
		pos.knightBoard &^= mask
	case Pawn:
		pos.pawnBoard &^= mask
	}
}

func setPieceBoard(pos *Position, pieceType int8, mask uint64) {
	switch pieceType {
	case King:
		pos.kingBoard |= mask
	case Queen:
		pos.queenBoard |= mask
	case Rook:
		pos.rookBoard |= mask
	case Bishop:
		pos.bishopBoard |= mask
	case Knight:
		pos.knightBoard |= mask
	case Pawn:
		pos.pawnBoard |= mask
	}
}

func promotionPieceType(flag int8) int8 {
	switch flag {
	case QueenPromotion:
		return Queen
	case KnightPromotion:
		return Knight
	case BishopPromotion:
		return Bishop
	default:
		return Rook
	}
}

func (updater *PlainPositionUpdater) MakeMove(pos *Position, move Move) MoveHistory {
	startPieceIdx := move.startIdx
	endPieceIdx := move.endIdx
	startPiece := move.piece
	startColor := pos.activeColor
	startPieceType := startPiece.Type()
	flag := move.flag
	isEnPassant := flag == EnPassant
	isPromotion := flag >= QueenPromotion && flag <= RookPromotion
	isCastle := flag == Castle

	captureIdx := endPieceIdx
	capturedPiece := pos.board[endPieceIdx]
	if isEnPassant {
		if startColor == White {
			captureIdx = endPieceIdx - 8
		} else {
			captureIdx = endPieceIdx + 8
		}
		capturedPiece = pos.board[captureIdx]
	}
	isCapture := capturedPiece != NoPiece

	history := MoveHistory{
		move:          move,
		capturedPiece: capturedPiece,
		captureIdx:    captureIdx,
		packedState: uint32(pos.whiteKingIdx)<<metaWhiteKingShift |
			uint32(pos.blackKingIdx)<<metaBlackKingShift |
			uint32(pos.enPassantIdx+1)<<metaEnPassantShift |
			uint32(pos.whiteCastleRights)<<metaWhiteCastleShift |
			uint32(pos.blackCastleRights)<<metaBlackCastleShift |
			(uint32(pos.whiteKingSafety)>>3)<<metaWhiteSafetyShift |
			(uint32(pos.blackKingSafety)>>3)<<metaBlackSafetyShift,
		whiteKingAffectMask: pos.whiteKingAffectMask,
		blackKingAffectMask: pos.blackKingAffectMask,
	}

	if (flag == NormalMove || flag == PawnDoubleMove) && !isCapture {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		moveMask := fromMask | toMask

		pos.occupied ^= moveMask
		if startColor == White {
			pos.whiteOccupied ^= moveMask
		} else {
			pos.blackOccupied ^= moveMask
		}

		if startPieceType == Pawn {
			pos.pawnBoard ^= moveMask
		} else if startPieceType == Rook {
			pos.rookBoard ^= moveMask
		} else if startPieceType == King {
			pos.kingBoard ^= moveMask
		} else {
			xorPieceBoard(pos, startPieceType, moveMask)
		}

		pos.board[startPieceIdx] = NoPiece
		pos.board[endPieceIdx] = startPiece
	} else if flag == Capture || ((flag == NormalMove || flag == PawnDoubleMove) && isCapture) {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		moveMask := fromMask | toMask

		pos.occupied &^= fromMask
		if startColor == White {
			pos.whiteOccupied ^= moveMask
			pos.blackOccupied &^= toMask
		} else {
			pos.blackOccupied ^= moveMask
			pos.whiteOccupied &^= toMask
		}

		if capturedPiece.Type() == Pawn {
			pos.pawnBoard &^= toMask
		} else if capturedPiece.Type() == Rook {
			pos.rookBoard &^= toMask
		} else if capturedPiece.Type() == King {
			pos.kingBoard &^= toMask
		} else {
			clearPieceBoard(pos, capturedPiece.Type(), toMask)
		}

		if startPieceType == Pawn {
			pos.pawnBoard ^= moveMask
		} else if startPieceType == Rook {
			pos.rookBoard ^= moveMask
		} else if startPieceType == King {
			pos.kingBoard ^= moveMask
		} else {
			xorPieceBoard(pos, startPieceType, moveMask)
		}

		pos.board[startPieceIdx] = NoPiece
		pos.board[endPieceIdx] = startPiece
	} else if isEnPassant {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		captureMask := uint64(1 << captureIdx)
		moveMask := fromMask | toMask

		pos.occupied ^= moveMask
		pos.occupied &^= captureMask
		if startColor == White {
			pos.whiteOccupied ^= moveMask
			pos.blackOccupied &^= captureMask
		} else {
			pos.blackOccupied ^= moveMask
			pos.whiteOccupied &^= captureMask
		}

		clearPieceBoard(pos, Pawn, captureMask)
		xorPieceBoard(pos, Pawn, moveMask)

		pos.board[startPieceIdx] = NoPiece
		pos.board[captureIdx] = NoPiece
		pos.board[endPieceIdx] = startPiece
	} else if isPromotion {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		promotedPieceType := promotionPieceType(flag)

		pos.occupied &^= fromMask
		pos.occupied |= toMask
		if startColor == White {
			pos.whiteOccupied &^= fromMask
			pos.whiteOccupied |= toMask
			if capturedPiece != NoPiece {
				pos.blackOccupied &^= toMask
			}
		} else {
			pos.blackOccupied &^= fromMask
			pos.blackOccupied |= toMask
			if capturedPiece != NoPiece {
				pos.whiteOccupied &^= toMask
			}
		}

		clearPieceBoard(pos, Pawn, fromMask)
		if capturedPiece != NoPiece {
			clearPieceBoard(pos, capturedPiece.Type(), toMask)
		}
		setPieceBoard(pos, promotedPieceType, toMask)

		pos.board[startPieceIdx] = NoPiece
		pos.board[endPieceIdx] = Piece(startColor | promotedPieceType)
	} else if isCastle {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		rookStartIdx, rookEndIdx := castleRookSquares(startColor, endPieceIdx)
		rookFromMask := uint64(1 << rookStartIdx)
		rookToMask := uint64(1 << rookEndIdx)
		kingMoveMask := fromMask | toMask
		rookMoveMask := rookFromMask | rookToMask

		pos.occupied ^= kingMoveMask | rookMoveMask
		if startColor == White {
			pos.whiteOccupied ^= kingMoveMask | rookMoveMask
		} else {
			pos.blackOccupied ^= kingMoveMask | rookMoveMask
		}
		xorPieceBoard(pos, King, kingMoveMask)
		xorPieceBoard(pos, Rook, rookMoveMask)

		pos.board[startPieceIdx] = NoPiece
		pos.board[endPieceIdx] = startPiece
		pos.board[rookStartIdx] = NoPiece
		pos.board[rookEndIdx] = Piece(startColor | Rook)
	}

	// King move -> update king pos and castleRights
	if startPieceType == King {
		if startColor == White {
			pos.whiteKingIdx = endPieceIdx
			pos.whiteKingAffectMask = kingAffectMask(endPieceIdx)
			pos.whiteCastleRights = NoCastle

		} else {
			pos.blackKingIdx = endPieceIdx
			pos.blackKingAffectMask = kingAffectMask(endPieceIdx)
			pos.blackCastleRights = NoCastle

		}
	}

	// en passant
	if startPieceType == Pawn {
		diffIdx := endPieceIdx - startPieceIdx
		if startColor == White && diffIdx == 16 {
			pos.enPassantIdx = startPieceIdx + 8
		} else if startColor == Black && diffIdx == -16 {
			pos.enPassantIdx = startPieceIdx - 8
		} else {
			pos.enPassantIdx = NoEnPassant
		}
	} else {
		pos.enPassantIdx = NoEnPassant
	}

	// knight
	if startPieceType == Rook {
		if startColor == White {
			if startPieceIdx == A1 {
				pos.whiteCastleRights = pos.whiteCastleRights &^ QueenSideCastle
			} else if startPieceIdx == H1 {
				pos.whiteCastleRights = pos.whiteCastleRights &^ KingSideCastle
			}
		} else {
			if startPieceIdx == A8 {
				pos.blackCastleRights = pos.blackCastleRights &^ QueenSideCastle
			} else if startPieceIdx == H8 {
				pos.blackCastleRights = pos.blackCastleRights &^ KingSideCastle
			}
		}
	}

	updater.invalidateKingSafetyCaches(pos)

	// change side ( important to do it last for previous updates )
	if pos.activeColor == White {
		pos.activeColor = Black
	} else {
		pos.activeColor = White
	}

	return history
}

func (updater *PlainPositionUpdater) UnMakeMove(pos *Position, history MoveHistory) {
	move := history.move
	startPieceIdx := move.startIdx
	endPieceIdx := move.endIdx
	packedState := history.packedState
	movePiece := move.piece
	movePieceType := movePiece.Type()
	flag := move.flag
	isPromotion := flag >= QueenPromotion && flag <= RookPromotion
	isCastle := flag == Castle

	pos.activeColor = movePiece.Color()
	pos.whiteKingIdx = int8((packedState >> metaWhiteKingShift) & 0x3F)
	pos.blackKingIdx = int8((packedState >> metaBlackKingShift) & 0x3F)
	pos.whiteKingAffectMask = history.whiteKingAffectMask
	pos.blackKingAffectMask = history.blackKingAffectMask
	pos.whiteCastleRights = int8((packedState >> metaWhiteCastleShift) & 0x3)
	pos.blackCastleRights = int8((packedState >> metaBlackCastleShift) & 0x3)
	pos.whiteKingSafety = int8((packedState>>metaWhiteSafetyShift)&0x3) << 3
	pos.blackKingSafety = int8((packedState>>metaBlackSafetyShift)&0x3) << 3
	pos.enPassantIdx = int8((packedState>>metaEnPassantShift)&0x7F) - 1

	if (flag == NormalMove || flag == PawnDoubleMove) && history.capturedPiece == NoPiece {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		moveMask := fromMask | toMask

		pos.occupied ^= moveMask
		if pos.activeColor == White {
			pos.whiteOccupied ^= moveMask
		} else {
			pos.blackOccupied ^= moveMask
		}

		if movePieceType == Pawn {
			pos.pawnBoard ^= moveMask
		} else if movePieceType == Rook {
			pos.rookBoard ^= moveMask
		} else if movePieceType == King {
			pos.kingBoard ^= moveMask
		} else {
			xorPieceBoard(pos, movePieceType, moveMask)
		}

		pos.board[endPieceIdx] = NoPiece
		pos.board[startPieceIdx] = movePiece
	} else if flag == Capture || ((flag == NormalMove || flag == PawnDoubleMove) && history.capturedPiece != NoPiece) {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		moveMask := fromMask | toMask

		pos.occupied |= fromMask
		if pos.activeColor == White {
			pos.whiteOccupied ^= moveMask
			pos.blackOccupied |= toMask
		} else {
			pos.blackOccupied ^= moveMask
			pos.whiteOccupied |= toMask
		}

		if movePieceType == Pawn {
			pos.pawnBoard ^= moveMask
		} else if movePieceType == Rook {
			pos.rookBoard ^= moveMask
		} else if movePieceType == King {
			pos.kingBoard ^= moveMask
		} else {
			xorPieceBoard(pos, movePieceType, moveMask)
		}

		if history.capturedPiece.Type() == Pawn {
			pos.pawnBoard |= toMask
		} else if history.capturedPiece.Type() == Rook {
			pos.rookBoard |= toMask
		} else if history.capturedPiece.Type() == King {
			pos.kingBoard |= toMask
		} else {
			setPieceBoard(pos, history.capturedPiece.Type(), toMask)
		}

		pos.board[startPieceIdx] = movePiece
		pos.board[endPieceIdx] = history.capturedPiece
	} else if flag == EnPassant {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		captureMask := uint64(1 << history.captureIdx)
		moveMask := fromMask | toMask

		pos.occupied ^= moveMask
		pos.occupied |= captureMask
		if pos.activeColor == White {
			pos.whiteOccupied ^= moveMask
			pos.blackOccupied |= captureMask
		} else {
			pos.blackOccupied ^= moveMask
			pos.whiteOccupied |= captureMask
		}

		xorPieceBoard(pos, Pawn, moveMask)
		setPieceBoard(pos, Pawn, captureMask)

		pos.board[startPieceIdx] = movePiece
		pos.board[endPieceIdx] = NoPiece
		pos.board[history.captureIdx] = history.capturedPiece
	} else if isPromotion {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		promotedPieceType := promotionPieceType(flag)

		pos.occupied |= fromMask
		if history.capturedPiece == NoPiece {
			pos.occupied &^= toMask
		}
		if pos.activeColor == White {
			pos.whiteOccupied |= fromMask
			pos.whiteOccupied &^= toMask
			if history.capturedPiece != NoPiece {
				pos.blackOccupied |= toMask
			}
		} else {
			pos.blackOccupied |= fromMask
			pos.blackOccupied &^= toMask
			if history.capturedPiece != NoPiece {
				pos.whiteOccupied |= toMask
			}
		}

		clearPieceBoard(pos, promotedPieceType, toMask)
		setPieceBoard(pos, Pawn, fromMask)
		if history.capturedPiece != NoPiece {
			setPieceBoard(pos, history.capturedPiece.Type(), toMask)
		}

		pos.board[startPieceIdx] = movePiece
		pos.board[endPieceIdx] = history.capturedPiece
	} else if isCastle {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
		rookFromMask := uint64(1 << rookStartIdx)
		rookToMask := uint64(1 << rookEndIdx)
		kingMoveMask := fromMask | toMask
		rookMoveMask := rookFromMask | rookToMask

		pos.occupied ^= kingMoveMask | rookMoveMask
		if pos.activeColor == White {
			pos.whiteOccupied ^= kingMoveMask | rookMoveMask
		} else {
			pos.blackOccupied ^= kingMoveMask | rookMoveMask
		}
		xorPieceBoard(pos, King, kingMoveMask)
		xorPieceBoard(pos, Rook, rookMoveMask)

		pos.board[startPieceIdx] = movePiece
		pos.board[endPieceIdx] = NoPiece
		pos.board[rookStartIdx] = Piece(pos.activeColor | Rook)
		pos.board[rookEndIdx] = NoPiece
	}
}
