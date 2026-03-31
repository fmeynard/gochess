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
	startPieceType := startPiece.Type()
	isEnPassant := isEnPassantMove(pos, move)

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
		move:                move,
		capturedPiece:       capturedPiece,
		captureIdx:          captureIdx,
		packedState:         packMoveHistoryMeta(pos),
		whiteKingAffectMask: pos.whiteKingAffectMask,
		blackKingAffectMask: pos.blackKingAffectMask,
	}

	if isEnPassant {
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
		if pos.activeColor == White {
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
		if pos.activeColor == White && diffIdx == 16 {
			pos.enPassantIdx = startPieceIdx + 8
		} else if pos.activeColor == Black && diffIdx == -16 {
			pos.enPassantIdx = startPieceIdx - 8
		} else {
			pos.enPassantIdx = NoEnPassant
		}
	} else {
		pos.enPassantIdx = NoEnPassant
	}

	// knight
	if startPieceType == Rook {
		if pos.activeColor == White {
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

	pos.activeColor = move.piece.Color()
	pos.whiteKingIdx = int8((packedState >> metaWhiteKingShift) & 0x3F)
	pos.blackKingIdx = int8((packedState >> metaBlackKingShift) & 0x3F)
	pos.whiteKingAffectMask = history.whiteKingAffectMask
	pos.blackKingAffectMask = history.blackKingAffectMask
	pos.whiteCastleRights = int8((packedState >> metaWhiteCastleShift) & 0x3)
	pos.blackCastleRights = int8((packedState >> metaBlackCastleShift) & 0x3)
	pos.whiteKingSafety = decodeKingSafety((packedState >> metaWhiteSafetyShift) & 0x3)
	pos.blackKingSafety = decodeKingSafety((packedState >> metaBlackSafetyShift) & 0x3)
	pos.enPassantIdx = int8((packedState>>metaEnPassantShift)&0x7F) - 1

	if move.flag == QueenPromotion || move.flag == KnightPromotion || move.flag == BishopPromotion || move.flag == RookPromotion {
		pos.removePieceAt(endPieceIdx, pos.board[endPieceIdx])
		pos.addPieceAt(startPieceIdx, move.piece)
	} else {
		pos.movePiece(move.piece, endPieceIdx, startPieceIdx)
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
