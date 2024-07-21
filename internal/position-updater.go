package internal

type PositionUpdater struct {
	moveGenerator IMoveGenerator
}

func NewPositionUpdater(moveGenerator IMoveGenerator) *PositionUpdater {
	return &PositionUpdater{
		moveGenerator: moveGenerator,
	}
}

// updatePieceOnBoard
// update board and position masks
// Important: capture moves need to update opponent occupancy maks
func updatePieceOnBoard(p *Position, piece Piece, oldIdx int8, newIdx int8) {
	p.setPieceAt(oldIdx, NoPiece)
	p.setPieceAt(newIdx, piece)
}

func (updater *PositionUpdater) updateMovesAfterMove(pos *Position, move Move) {
	if updater.IsMoveAffectsKing(pos, move, pos.whiteKingIdx) {

		pos.whiteKingSafety = NotCalculated
		if IsKingInCheck(pos, White) {
			pos.whiteKingSafety = KingIsCheck
		} else {
			pos.whiteKingSafety = KingIsSafe
		}
	}

	if updater.IsMoveAffectsKing(pos, move, pos.blackKingIdx) {

		pos.blackKingSafety = NotCalculated
		if IsKingInCheck(pos, Black) {
			pos.blackKingSafety = KingIsCheck
		} else {
			pos.blackKingSafety = KingIsSafe
		}
	}

}

func (updater *PositionUpdater) MakeMove(pos *Position, move Move) MoveHistory {
	startPieceIdx := move.StartIdx()
	endPieceIdx := move.EndIdx()
	startPiece := pos.PieceAt(move.StartIdx())
	startPieceType := startPiece.Type()

	history := NewMoveHistory(pos)

	// update position
	updatePieceOnBoard(pos, startPiece, startPieceIdx, endPieceIdx)

	// King move -> update king pos and castleRights
	if startPieceType == King {
		if pos.activeColor == White {
			pos.whiteKingIdx = endPieceIdx
			pos.whiteCastleRights = NoCastle

			if startPieceIdx == E1 {
				if endPieceIdx == G1 {
					updatePieceOnBoard(pos, Piece(White|Rook), H1, F1)
				} else if endPieceIdx == C1 {
					updatePieceOnBoard(pos, Piece(White|Rook), A1, D1)
				}
			}
		} else {
			pos.blackKingIdx = endPieceIdx
			pos.blackCastleRights = NoCastle

			if startPieceIdx == E8 {
				if endPieceIdx == G8 {
					updatePieceOnBoard(pos, Piece(Black|Rook), H8, F8)
				} else if endPieceIdx == C8 {
					updatePieceOnBoard(pos, Piece(Black|Rook), A8, D8)
				}
			}
		}
	}

	// en passant
	if startPieceType == Pawn {
		if endPieceIdx == pos.enPassantIdx {
			var capturesPawnIdx int8
			if pos.activeColor == White {
				capturesPawnIdx = pos.enPassantIdx - 8
			} else {
				capturesPawnIdx = pos.enPassantIdx + 8
			}
			updatePieceOnBoard(pos, Piece(Pawn|pos.activeColor), startPieceIdx, capturesPawnIdx)
			updatePieceOnBoard(pos, Piece(Pawn|pos.activeColor), capturesPawnIdx, endPieceIdx)
		}

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

	updater.updateMovesAfterMove(pos, move)

	// change side ( important to do it last for previous updates )
	if pos.activeColor == White {
		pos.activeColor = Black
	} else {
		pos.activeColor = White
	}

	return history
}

func (updater *PositionUpdater) UnMakeMove(pos *Position, move Move, history MoveHistory) {
	// Restore the pieces to their original positions
	pos.board = history.board

	// Restore the state of the kings
	pos.whiteKingIdx = history.whiteKingIdx
	pos.blackKingIdx = history.blackKingIdx

	// Restore castling rights
	pos.whiteCastleRights = history.whiteCastleRights
	pos.blackCastleRights = history.blackCastleRights

	// reset king safety
	pos.blackKingSafety = history.blackKingSafety
	pos.whiteKingSafety = history.whiteKingSafety

	// reset occupancies masks
	pos.occupied = history.occupied
	pos.blackOccupied = history.blackOccupied
	pos.whiteOccupied = history.whiteOccupied

	pos.kingBoard = history.kingBoard
	pos.queenBoard = history.queenBoard
	pos.rookBoard = history.rookBoard
	pos.bishopBoard = history.bishopBoard
	pos.knightBoard = history.knightBoard
	pos.pawnBoard = history.pawnBoard

	// Restore en passant index
	pos.enPassantIdx = history.enPassantIdx
	// Restore the active color
	pos.activeColor = history.activeColor
}

func (updater *PositionUpdater) IsMoveAffectsKing(pos *Position, m Move, kingIdx int8) bool {
	// Convert indices to row and column
	startRank, startFile := RankAndFile(m.StartIdx())
	targetRank, targetFile := RankAndFile(m.EndIdx())
	kingRank, kingFile := RankAndFile(kingIdx)

	// Direct involvement
	if m.StartIdx() == kingIdx || m.EndIdx() == kingIdx {
		return true
	}

	// Check if the move is along the same rank, file, or diagonal as the king
	if startRank == kingRank || startFile == kingFile || absInt8(startRank-kingRank) == absInt8(startFile-kingFile) {
		// Now check if the move is actually on the path between the king and the moved piece
		if m.isOnLine(kingIdx, pos) {
			return true
		}
	}

	// Special checks for knights moves
	if m.knightAffectsKing(m.EndIdx(), kingIdx) {
		return true
	}

	if m.knightAffectsKing(m.StartIdx(), kingIdx) {
		return true
	}

	// Check the ending square as well, since moves can open up lines
	if targetRank == kingRank || targetFile == kingFile || absInt8(targetRank-kingRank) == absInt8(targetFile-kingFile) {
		if m.isOnLine(kingIdx, pos) {
			return true
		}
	}

	return false
}

func (m Move) knightAffectsKing(knightEndIdx, kingIdx int8) bool {
	for _, move := range KnightOffsets {
		if kingIdx == knightEndIdx+move {
			return true
		}
	}
	return false
}

func (m Move) isOnLine(kingIdx int8, pos *Position) bool {
	for _, dir := range QueenDirections {
		for i := int8(1); i < 8; i++ { // Check up to 7 squares away in each direction
			checkIdx := kingIdx + i*dir
			if checkIdx < 0 || checkIdx >= 64 {
				break
			}
			if checkIdx == m.StartIdx() || checkIdx == m.EndIdx() {
				return true
			}

			if pos.IsOccupied(checkIdx) {
				break
			}
		}
	}
	return false
}
