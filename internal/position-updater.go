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

	if piece.Color() == White {
		p.whiteOccupied &= ^(uint64(1) << oldIdx)
		p.whiteOccupied |= uint64(1) << newIdx
		p.blackOccupied &= ^(uint64(1) << newIdx)
	} else {
		p.whiteOccupied &= ^(uint64(1) << newIdx)
		p.blackOccupied &= ^(uint64(1) << oldIdx)
		p.blackOccupied |= uint64(1) << newIdx
	}

	p.occupied &= ^(uint64(1) << oldIdx)
	p.occupied |= uint64(1) << newIdx

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

	history := MoveHistory{
		board:             pos.board,
		whiteKingIdx:      pos.whiteKingIdx,
		blackKingIdx:      pos.blackKingIdx,
		whiteCastleRights: pos.whiteCastleRights,
		blackCastleRights: pos.blackCastleRights,
		enPassantIdx:      pos.enPassantIdx,
		activeColor:       pos.activeColor,
		blackKingSafety:   pos.blackKingSafety,
		whiteKingSafety:   pos.whiteKingSafety,
		whiteOccupied:     pos.whiteOccupied,
		blackOccupied:     pos.blackOccupied,
		occupied:          pos.occupied,
	}

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
	if m.piece.Type() == Knight && m.knightAffectsKing(m.EndIdx(), kingIdx) {
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
	knightMoves := []int8{-17, -15, -10, -6, 6, 10, 15, 17} // Possible knight moves
	for _, move := range knightMoves {
		if kingIdx == knightEndIdx+move {
			return true
		}
	}
	return false
}

func (m Move) isOnLine(kingIdx int8, pos *Position) bool {
	direction := []int{1, -1, 8, -8, 7, -7, 9, -9} // directions representing all line movements
	for _, dir := range direction {
		for i := 1; i < 8; i++ { // Check up to 7 squares away in each direction
			checkIdx := kingIdx + int8(i*dir)
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

type MoveHistory struct {
	whiteKingIdx      int8
	blackKingIdx      int8
	whiteCastleRights int8
	blackCastleRights int8
	enPassantIdx      int8
	activeColor       int8
	whiteKingSafety   int8
	blackKingSafety   int8
	whiteOccupied     uint64
	blackOccupied     uint64
	occupied          uint64
	board             [64]Piece
}

//
//func (updater *PositionUpdater) updateMovesAfterMove(pos *Position, move Move) {
//	// Identify pieces potentially affected by the move
//	fromIdx := move.StartIdx()
//	toIdx := move.EndIdx()
//
//	for _, dir := range []int8{8, -8, 1, -1, 7, -7, 9, -9} {
//		for dist := int8(1); dist <= 7; dist++ { // Maximum board distance
//			affectedIdx := fromIdx + dir*dist
//			if affectedIdx < 0 || affectedIdx >= 64 || !isSameLine(fromIdx, affectedIdx, dir) {
//				break
//			}
//			piece := pos.board[affectedIdx]
//			if piece != NoPiece && canReach(piece.Type(), dir) {
//				//delete(pos.moveCache, affectedIdx)
//				//pos.generateMovesForPiece(affectedIdx)
//				pos.movesCache[affectedIdx] = nil
//				pos.generateMovesForPiece(affectedIdx)
//			}
//		}
//	}
//	pos.moveCache[toIdx] = pos.generateMovesForPiece(toIdx) // Recalculate moves for the moved piece
//}
