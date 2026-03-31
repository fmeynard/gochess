package internal

type moveApplier interface {
	MakeMove(pos *Position, move Move) MoveHistory
	UnMakeMove(pos *Position, history MoveHistory)
	IsMoveAffectsKing(pos *Position, m Move, kingColor int8) bool
}

type PlainPositionUpdater struct {
	moveGenerator *BitsBoardMoveGenerator
}

func kingAffectMask(kingIdx int8) uint64 {
	return queenAttacksMask[kingIdx] | knightAttacksMask[kingIdx] | (uint64(1) << kingIdx)
}

func NewPositionUpdater(moveGenerator *BitsBoardMoveGenerator) moveApplier {
	return NewZobristPositionUpdater(NewPlainPositionUpdater(moveGenerator))
}

func NewPlainPositionUpdater(moveGenerator *BitsBoardMoveGenerator) *PlainPositionUpdater {
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

func (updater *PlainPositionUpdater) MakeMove(pos *Position, move Move) MoveHistory {
	startPieceIdx := move.startIdx
	endPieceIdx := move.endIdx
	startPiece := move.piece
	startColor := pos.activeColor
	startPieceType := startPiece.Type()
	isEnPassant := isEnPassantMove(pos, move)
	isPromotion := move.flag == QueenPromotion || move.flag == KnightPromotion || move.flag == BishopPromotion || move.flag == RookPromotion

	captureIdx := endPieceIdx
	capturedPiece := pos.board[endPieceIdx]
	if isEnPassant {
		if pos.activeColor == White {
			captureIdx = endPieceIdx - 8
		} else {
			captureIdx = endPieceIdx + 8
		}
		capturedPiece = pos.board[captureIdx]
	}

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

	if !isEnPassant && !isPromotion && capturedPiece == NoPiece {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		moveMask := fromMask | toMask

		pos.occupied ^= moveMask
		if startColor == White {
			pos.whiteOccupied ^= moveMask
		} else {
			pos.blackOccupied ^= moveMask
		}

		switch startPieceType {
		case King:
			pos.kingBoard ^= moveMask
		case Queen:
			pos.queenBoard ^= moveMask
		case Rook:
			pos.rookBoard ^= moveMask
		case Bishop:
			pos.bishopBoard ^= moveMask
		case Knight:
			pos.knightBoard ^= moveMask
		case Pawn:
			pos.pawnBoard ^= moveMask
		}

		pos.board[startPieceIdx] = NoPiece
		pos.board[endPieceIdx] = startPiece
	} else if !isEnPassant && !isPromotion && capturedPiece != NoPiece {
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

		switch capturedPiece.Type() {
		case King:
			pos.kingBoard &^= toMask
		case Queen:
			pos.queenBoard &^= toMask
		case Rook:
			pos.rookBoard &^= toMask
		case Bishop:
			pos.bishopBoard &^= toMask
		case Knight:
			pos.knightBoard &^= toMask
		case Pawn:
			pos.pawnBoard &^= toMask
		}

		switch startPieceType {
		case King:
			pos.kingBoard ^= moveMask
		case Queen:
			pos.queenBoard ^= moveMask
		case Rook:
			pos.rookBoard ^= moveMask
		case Bishop:
			pos.bishopBoard ^= moveMask
		case Knight:
			pos.knightBoard ^= moveMask
		case Pawn:
			pos.pawnBoard ^= moveMask
		}

		pos.board[startPieceIdx] = NoPiece
		pos.board[endPieceIdx] = startPiece
	} else if isEnPassant {
		pos.removePieceAt(startPieceIdx, startPiece)
		pos.removePieceAt(captureIdx, capturedPiece)
		pos.addPieceAt(endPieceIdx, startPiece)
	} else if capturedPiece != NoPiece {
		pos.capturePiece(startPiece, capturedPiece, startPieceIdx, endPieceIdx)
	} else {
		pos.movePiece(startPiece, startPieceIdx, endPieceIdx)
	}

	switch move.flag {
	case QueenPromotion:
		pos.removePieceAt(endPieceIdx, startPiece)
		pos.addPieceAt(endPieceIdx, Piece(pos.activeColor|Queen))
	case KnightPromotion:
		pos.removePieceAt(endPieceIdx, startPiece)
		pos.addPieceAt(endPieceIdx, Piece(pos.activeColor|Knight))
	case BishopPromotion:
		pos.removePieceAt(endPieceIdx, startPiece)
		pos.addPieceAt(endPieceIdx, Piece(pos.activeColor|Bishop))
	case RookPromotion:
		pos.removePieceAt(endPieceIdx, startPiece)
		pos.addPieceAt(endPieceIdx, Piece(pos.activeColor|Rook))
	}

	// King move -> update king pos and castleRights
	if startPieceType == King {
		if startColor == White {
			pos.whiteKingIdx = endPieceIdx
			pos.whiteKingAffectMask = kingAffectMask(endPieceIdx)
			pos.whiteCastleRights = NoCastle

			if isCastleMove(move) {
				rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
				pos.movePiece(Piece(White|Rook), rookStartIdx, rookEndIdx)
			}
		} else {
			pos.blackKingIdx = endPieceIdx
			pos.blackKingAffectMask = kingAffectMask(endPieceIdx)
			pos.blackCastleRights = NoCastle

			if isCastleMove(move) {
				rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
				pos.movePiece(Piece(Black|Rook), rookStartIdx, rookEndIdx)
			}
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
	isPromotion := move.flag == QueenPromotion || move.flag == KnightPromotion || move.flag == BishopPromotion || move.flag == RookPromotion

	pos.activeColor = movePiece.Color()
	pos.whiteKingIdx = int8((packedState >> metaWhiteKingShift) & 0x3F)
	pos.blackKingIdx = int8((packedState >> metaBlackKingShift) & 0x3F)
	pos.whiteKingAffectMask = history.whiteKingAffectMask
	pos.blackKingAffectMask = history.blackKingAffectMask
	pos.whiteCastleRights = int8((packedState >> metaWhiteCastleShift) & 0x3)
	pos.blackCastleRights = int8((packedState >> metaBlackCastleShift) & 0x3)
	pos.whiteKingSafety = decodeKingSafety((packedState >> metaWhiteSafetyShift) & 0x3)
	pos.blackKingSafety = decodeKingSafety((packedState >> metaBlackSafetyShift) & 0x3)
	pos.enPassantIdx = int8((packedState>>metaEnPassantShift)&0x7F) - 1

	if !isPromotion && move.flag != EnPassant && !isCastleMove(move) && history.capturedPiece == NoPiece {
		fromMask := uint64(1 << startPieceIdx)
		toMask := uint64(1 << endPieceIdx)
		moveMask := fromMask | toMask

		pos.occupied ^= moveMask
		if pos.activeColor == White {
			pos.whiteOccupied ^= moveMask
		} else {
			pos.blackOccupied ^= moveMask
		}

		switch movePieceType {
		case King:
			pos.kingBoard ^= moveMask
		case Queen:
			pos.queenBoard ^= moveMask
		case Rook:
			pos.rookBoard ^= moveMask
		case Bishop:
			pos.bishopBoard ^= moveMask
		case Knight:
			pos.knightBoard ^= moveMask
		case Pawn:
			pos.pawnBoard ^= moveMask
		}

		pos.board[endPieceIdx] = NoPiece
		pos.board[startPieceIdx] = movePiece
	} else if !isPromotion && move.flag != EnPassant && !isCastleMove(move) && history.capturedPiece != NoPiece {
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

		switch movePieceType {
		case King:
			pos.kingBoard ^= moveMask
		case Queen:
			pos.queenBoard ^= moveMask
		case Rook:
			pos.rookBoard ^= moveMask
		case Bishop:
			pos.bishopBoard ^= moveMask
		case Knight:
			pos.knightBoard ^= moveMask
		case Pawn:
			pos.pawnBoard ^= moveMask
		}

		switch history.capturedPiece.Type() {
		case King:
			pos.kingBoard |= toMask
		case Queen:
			pos.queenBoard |= toMask
		case Rook:
			pos.rookBoard |= toMask
		case Bishop:
			pos.bishopBoard |= toMask
		case Knight:
			pos.knightBoard |= toMask
		case Pawn:
			pos.pawnBoard |= toMask
		}

		pos.board[startPieceIdx] = movePiece
		pos.board[endPieceIdx] = history.capturedPiece
	} else if isPromotion {
		pos.removePieceAt(endPieceIdx, pos.board[endPieceIdx])
		pos.addPieceAt(startPieceIdx, movePiece)
	} else {
		pos.movePiece(movePiece, endPieceIdx, startPieceIdx)
	}

	if isCastleMove(move) {
		rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
		pos.movePiece(Piece(pos.activeColor|Rook), rookEndIdx, rookStartIdx)
	}

	if history.capturedPiece != NoPiece {
		pos.addPieceAt(history.captureIdx, history.capturedPiece)
	}
}

// IsMoveAffectsKing
// The goal here is to trigger king safety recalculation as less as possible,
// but it's a balance, if detection is less performant than the recalculation it's better to recalculate
func (updater *PlainPositionUpdater) IsMoveAffectsKing(pos *Position, m Move, kingColor int8) bool {
	var kingAffectMask uint64
	if kingColor == White {
		kingAffectMask = pos.whiteKingAffectMask
	} else {
		kingAffectMask = pos.blackKingAffectMask
	}

	movePiecesMask := uint64(1<<m.startIdx | 1<<m.endIdx)
	return (kingAffectMask & movePiecesMask) != 0
}
