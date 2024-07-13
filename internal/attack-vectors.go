package internal

import (
	"fmt"
	"strings"
)

func updateAttackVectors(pos *Position) {
	// Clear previous attacks
	pos.whiteAttacks = 0
	pos.blackAttacks = 0

	// Recalculate attacks for all pieces
	for i := int8(0); i < 64; i++ {
		piece := pos.board[i]
		if piece == NoPiece {
			continue
		}

		isWhite := piece.IsWhite()

		switch piece.Type() {
		case Pawn:
			updatePawnAttacks(pos, i, isWhite)
		case Knight:
			updateKnightAttacks(pos, i, isWhite)
		case Bishop:
			updateSlidingPieceAttacks(pos, i, BishopDirections, isWhite)
		case Rook:
			updateSlidingPieceAttacks(pos, i, RookDirections, isWhite)
		case Queen:
			updateSlidingPieceAttacks(pos, i, QueenDirections, isWhite)
		case King:
			updateKingAttacks(pos, i, isWhite)
		}
	}
}

// updatePawnAttacks updates the attack bitboards for pawns.
func updatePawnAttacks(pos *Position, idx int8, isWhite bool) {
	pieceFile := FileFromIdx(idx)
	if isWhite {
		if pieceFile != 0 { // Not on A-file
			pos.whiteAttacks |= 1 << (idx + 7)
		}
		if pieceFile != 7 { // Not on H-file
			pos.whiteAttacks |= 1 << (idx + 9)
		}
	} else {
		if pieceFile != 0 { // Not on A-file
			pos.blackAttacks |= 1 << (idx - 9)
		}
		if pieceFile != 7 { // Not on H-file
			pos.blackAttacks |= 1 << (idx - 7)
		}
	}
}

func updateKnightAttacks(pos *Position, idx int8, isWhite bool) {

	pieceRank, pieceFile := RankAndFile(idx)

	for _, move := range knightMoves2 {
		targetIdx := idx + move
		if targetIdx >= 0 && targetIdx < 64 {

			targetRank, targetFile := RankAndFile(targetIdx)

			if absInt8(targetRank-pieceRank) <= 2 && absInt8(targetFile-pieceFile) <= 2 {
				if isWhite {
					pos.whiteAttacks |= 1 << targetIdx
				} else {
					pos.blackAttacks |= 1 << targetIdx
				}
			}
		}
	}
}

// updateSlidingPieceAttacks updates the attack bitboards for sliding pieces.
func updateSlidingPieceAttacks(pos *Position, idx int8, directions []int8, isWhite bool) {

	currentRank, currentFile := RankAndFile(idx)
	for _, direction := range directions {
		targetIdx := idx
		for {
			targetIdx += direction
			if targetIdx < 0 || targetIdx >= 64 {
				break // Out of bounds
			}

			targetRank, targetFile := RankAndFile(targetIdx)

			if absInt8(targetRank-currentRank) > 1 && absInt8(targetFile-currentFile) > 1 {
				break // Move wraps around board edges
			}

			if isWhite {
				pos.whiteAttacks |= 1 << targetIdx
			} else {
				pos.blackAttacks |= 1 << targetIdx
			}

			if pos.board[targetIdx] != NoPiece {
				break // Stop if another piece is encountered
			}
		}
	}
}

// updateKingAttacks updates the attack bitboards for kings.
func updateKingAttacks(pos *Position, idx int8, isWhite bool) {
	//kingMoves := []int{
	//	-9, -8, -7, -1, 1, 7, 8, 9,
	//}

	currentRank, currentFile := RankAndFile(idx)
	for _, move := range kingMoves2 {
		targetIdx := idx + move
		if targetIdx >= 0 && targetIdx < 64 {
			targetRank, targetFile := RankAndFile(targetIdx)

			if absInt8(targetRank-currentRank) <= 1 && absInt8(targetFile-currentFile) <= 1 {
				if isWhite {
					pos.whiteAttacks |= 1 << targetIdx
				} else {
					pos.blackAttacks |= 1 << targetIdx
				}
			}
		}
	}
}

func isSquareAttacked(pos Position, squareIdx int, attackerColor int8) bool {
	if attackerColor == White {
		return (pos.whiteAttacks & (1 << squareIdx)) != 0
	}
	return (pos.blackAttacks & (1 << squareIdx)) != 0
}

// isKingInCheck checks if the king of the given color is in check.
func isKingInCheckByVector(pos Position, kingColor int8) bool {
	if kingColor == White {
		return isSquareAttacked(pos, int(pos.whiteKingIdx), Black)
	}
	return isSquareAttacked(pos, int(pos.blackKingIdx), White)
}

func draw(vector uint64) {
	for rank := 7; rank >= 0; rank-- {
		var currentLine []string
		for file := 0; file < 8; file++ {
			mask := uint64(1) << (rank*8 + file)
			if vector&mask != 0 {
				currentLine = append(currentLine, "1")
			} else {
				currentLine = append(currentLine, "0")
			}
		}

		fmt.Println("|", strings.Join(currentLine, " | "), "|")
	}
}
