package internal

const (
	NoSlider     = 0
	BishopSlider = 1
	RookSlider   = 2
)

// isOnBoard checks if the given file and rank are within the bounds of the board.
func isOnBoard(file, rank int8) bool {
	return file >= 0 && file < 8 && rank >= 0 && rank < 8
}

// verify if the square is attacked by pawn
func isSquareAttackedByPawn(pos *Position, idx int8, kingColor int8) bool {
	rank, file := RankAndFile(idx)
	pawnAttacks := [2][2]int8{{-1, 1}, {1, 1}}
	if kingColor != White {
		pawnAttacks = [2][2]int8{{-1, -1}, {1, -1}}
	}

	for _, attack := range pawnAttacks {
		newFile := file + attack[0]
		newRank := rank + attack[1]
		if !isOnBoard(newFile, newRank) {
			continue
		}

		endIdx := newRank*8 + newFile
		piece := pos.board[endIdx]
		if piece.Type() == Pawn && piece.Color() != kingColor {
			return true
		}
	}

	return false
}

// verify if the square is attacked by knight
func isSquareAttackedByKnight(pos *Position, idx int8, kingColor int8) bool {
	rank, file := RankAndFile(idx)
	for _, move := range knightMoves {
		newFile := file + move[0]
		newRank := rank + move[1]
		if isOnBoard(newFile, newRank) {
			endIdx := newRank*8 + newFile
			piece := pos.board[endIdx]
			if piece.Type() == Knight && piece.Color() != kingColor {
				return true
			}
		}
	}

	return false
}

func isSquareAttackedBySlidingPiece(pos *Position, pieceIdx int8, kingColor int8) bool {
	pieceRank, pieceFile := RankAndFile(pieceIdx)

	for _, dir := range queenDirections {
		for targetIdx := pieceIdx + dir; targetIdx >= 0 && targetIdx < 64; targetIdx += dir {
			targetRank, targetFile := RankAndFile(targetIdx)
			if (dir == LEFT || dir == RIGHT) && targetRank != pieceRank {
				break
			}

			if (dir == DOWN || dir == UP) && targetFile != pieceFile {
				break
			}

			isDiagonal := dir == DownRight || dir == UpLeft || dir == DownLeft || dir == UpRight
			if isDiagonal && !IsSameDiagonal(pieceRank, pieceFile, targetRank, targetFile) {
				break
			}

			targetPiece := pos.board[targetIdx]
			if targetPiece != NoPiece {
				if targetPiece.Color() != kingColor {
					targetPieceType := targetPiece.Type()

					if !isDiagonal && (targetPieceType == Rook || targetPieceType == Queen) {
						return true
					}
					if isDiagonal && (targetPieceType == Bishop || targetPieceType == Queen) {
						return true
					}
				}
				break
			}
		}
	}

	return false
}

// verify if the square is attacked by king
func isSquareAttackedByKing(pos *Position, idx int8) bool {
	rank, file := RankAndFile(idx)

	for _, move := range kingMoves {
		newFile := file + move[0]
		newRank := rank + move[1]
		if isOnBoard(newFile, newRank) {
			endIdx := newRank*8 + newFile
			piece := pos.board[endIdx]
			if piece.Type() == King && piece.Color() != pos.activeColor {
				return true
			}
		}
	}
	return false
}

var cacheHits = 0
var cacheMiss = 0

// IsKingInCheck verifies if the king at the given index is in check.
func IsKingInCheck(pos *Position, kingColor int8) bool {
	var (
		kingIdx int8
	)
	if kingColor == White {
		kingIdx = pos.whiteKingIdx
		if pos.whiteKingSafety != NotCalculated {
			return pos.whiteKingSafety == KingIsCheck
		}
	} else {
		kingIdx = pos.blackKingIdx
		if pos.blackKingSafety != NotCalculated {
			return pos.blackKingSafety == KingIsCheck
		}
	}

	isAttacked := false
	if isSquareAttackedByPawn(pos, kingIdx, kingColor) ||
		isSquareAttackedByKnight(pos, kingIdx, kingColor) ||
		isSquareAttackedBySlidingPiece(pos, kingIdx, kingColor) ||
		isSquareAttackedByKing(pos, kingIdx) {
		isAttacked = true
	}

	if kingColor == White {
		if isAttacked {
			pos.whiteKingSafety = KingIsCheck
		} else {
			pos.whiteKingSafety = KingIsSafe
		}
	} else {
		if isAttacked {
			pos.blackKingSafety = KingIsCheck
		} else {
			pos.blackKingSafety = KingIsSafe
		}
	}

	return isAttacked
}
