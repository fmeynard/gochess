package internal

// direction offsets
const (
	LEFT      int8 = -1
	RIGHT     int8 = 1
	UP        int8 = -8
	DOWN      int8 = 8
	UpLeft    int8 = -9
	UpRight   int8 = -7
	DownLeft  int8 = 7
	DownRight int8 = 9
)

func SliderPseudoLegalMoves(p Position, pieceIdx int8) ([]int8, []int8) {
	piece := p.PieceAt(pieceIdx)
	if piece == NoPiece {
		return nil, nil
	}

	return generateSliderPseudoLegalMoves(p, pieceIdx, piece)
}

func generateSliderPseudoLegalMoves(p Position, pieceIdx int8, piece Piece) ([]int8, []int8) {
	var (
		moves         []int8
		capturesMoves []int8
	)

	directions, maxMoves := piece.PossibleDirectionsAndMaxMoves()

	for _, direction := range directions {
		for i := int8(1); i <= maxMoves; i++ { // start at 1 because 0 is current square
			targetIdx := pieceIdx + direction*i

			// current move+direction is out of the board
			// handle UP and DOWN
			if targetIdx < 0 || targetIdx > 63 {
				break
			}

			// horizontal+diagonal checks
			file := pieceIdx % 8
			if file == 7 && (direction == RIGHT || direction == UpRight || direction == DownRight) {
				break
			}

			if file == 0 && (direction == LEFT || direction == UpLeft || direction == DownLeft) {
				break
			}

			// target square is not empty -> stop
			target := p.PieceAt(targetIdx)
			if target != NoPiece {
				if target.Color() == piece.Color() {
					break
				}

				capturesMoves = append(capturesMoves, targetIdx)
				moves = append(moves, targetIdx)
				break
			}

			// add to the list
			moves = append(moves, targetIdx)
		}
	}

	return moves, capturesMoves
}

func KnightPseudoLegalMoves(p Position, pieceIdx int8) ([]int8, []int8) {
	return generateKnightPseudoLegalMoves(p, pieceIdx, p.PieceAt(pieceIdx))
}

func generateKnightPseudoLegalMoves(p Position, pieceIdx int8, piece Piece) ([]int8, []int8) {
	var (
		moves         []int8
		capturesMoves []int8
	)
	offsets := []int8{-17, -15, -10, -6, 6, 10, 15, 17}
	pieceRank, pieceFile := RankAndFile(pieceIdx)
	for _, offset := range offsets {
		targetIdx := pieceIdx + offset
		if targetIdx < 0 || targetIdx > 63 {
			continue
		}

		targetRank, targetFile := RankAndFile(targetIdx)

		var (
			rankDiff int8
			fileDiff int8
		)
		if pieceRank > targetRank {
			rankDiff = pieceRank - targetRank
		} else {
			rankDiff = targetRank - pieceRank
		}

		if pieceFile > targetFile {
			fileDiff = pieceFile - targetFile
		} else {
			fileDiff = targetFile - pieceFile
		}

		combinedDiff := fileDiff + rankDiff
		if combinedDiff != -3 && combinedDiff != 3 {
			continue
		}

		target := p.PieceAt(targetIdx)
		if target != NoPiece {
			if target.Color() != piece.Color() {
				moves = append(moves, targetIdx)
				capturesMoves = append(capturesMoves, targetIdx)
			}
		} else {
			moves = append(moves, targetIdx)
		}
	}

	return moves, capturesMoves
}

func KingPseudoLegalMoves(p Position, pieceIdx int8) ([]int8, []int8) {
	moves, capturesMoves := SliderPseudoLegalMoves(p, pieceIdx)

	piece := p.PieceAt(pieceIdx)
	pieceColor := piece.Color()

	var (
		castleRights int8
		kingStartIdx int8
	)
	if pieceColor == White {
		castleRights = p.whiteCastleRights
		kingStartIdx = E1
	} else {
		castleRights = p.blackCastleRights
		kingStartIdx = E8
	}

	// early exit no castle
	if pieceIdx != kingStartIdx || castleRights == NoCastle {
		return moves, capturesMoves
	}

	// queen side
	var (
		queenPathIsClear bool
		queenRookIdx     int8
		queenCastleIdx   int8
	)
	if (castleRights & QueenSideCastle) != 0 {
		if piece.Color() == White {
			queenRookIdx = A1
			queenCastleIdx = B1
			queenPathIsClear = (p.PieceAt(B1) == NoPiece) && (p.PieceAt(C1) == NoPiece) && (p.PieceAt(D1) == NoPiece)
		} else {
			queenRookIdx = A8
			queenCastleIdx = B8
			queenPathIsClear = (p.PieceAt(B8) == NoPiece) && (p.PieceAt(C8) == NoPiece) && (p.PieceAt(D8) == NoPiece)
		}

		if queenPathIsClear && p.PieceAt(queenRookIdx).Type() == Rook {
			moves = append(moves, queenCastleIdx)
		}
	}

	// king side
	var (
		kingPathIsClear bool
		kingRookIdx     int8
		kingCastleIdx   int8
	)
	if (castleRights & KingSideCastle) != 0 {
		if piece.Color() == White {
			kingRookIdx = H1
			kingCastleIdx = G1
			kingPathIsClear = (p.PieceAt(G1) == NoPiece) && (p.PieceAt(F1) == NoPiece)
		} else {
			kingRookIdx = H8
			kingCastleIdx = G8
			kingPathIsClear = (p.PieceAt(G8) == NoPiece) && (p.PieceAt(F8) == NoPiece)
		}

		if kingPathIsClear && p.PieceAt(kingRookIdx).Type() == Rook {
			moves = append(moves, kingCastleIdx)
		}
	}

	return moves, capturesMoves
}

func PawnPseudoLegalMoves(p Position, pieceIdx int8) ([]int8, []int8) {
	var (
		moves         []int8
		capturesMoves []int8
	)

	piece := p.PieceAt(pieceIdx)
	pieceColor := piece.Color()

	direction := int8(1)
	if pieceColor == Black {
		direction = -1
	}

	rank, file := RankAndFile(pieceIdx)

	// 1 forward
	target1Idx := pieceIdx + (8 * direction)
	target1 := p.PieceAt(target1Idx)
	if target1 == NoPiece {
		moves = append(moves, target1Idx)
	}

	// 2 forward
	if ((pieceColor == White && rank == 1) || (pieceColor == Black && rank == 6)) && target1 == NoPiece {
		target2Idx := pieceIdx + (16 * direction)
		target2 := p.PieceAt(target2Idx)
		if target2 == NoPiece {
			moves = append(moves, target2Idx)
		}
	}

	// capture
	if file > 0 {
		leftTargetIdx := pieceIdx + (8 * direction) - 1
		leftTarget := p.PieceAt(leftTargetIdx)
		if (leftTarget != NoPiece && leftTarget.Color() != pieceColor) || leftTargetIdx == p.enPassantIdx {
			moves = append(moves, leftTargetIdx)
			capturesMoves = append(capturesMoves, leftTargetIdx)
		}
	}

	if file < 7 {
		rightTargetIdx := pieceIdx + (8 * direction) + 1
		rightTarget := p.PieceAt(rightTargetIdx)
		if rightTarget != NoPiece && rightTarget.Color() != pieceColor || rightTargetIdx == p.enPassantIdx {
			moves = append(moves, rightTargetIdx)
			capturesMoves = append(capturesMoves, rightTargetIdx)
		}
	}

	return moves, capturesMoves
}
