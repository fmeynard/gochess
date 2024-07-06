package internal

import (
	"fmt"
)

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
	var (
		directions    []int8
		moves         []int8
		capturesMoves []int8
		maxMoves      int8
	)

	piece := p.PieceAt(pieceIdx)
	if piece == NoPiece {
		return nil, nil
	}

	pieceColor := piece.Color()
	pieceType := piece.Type()

	maxMoves = 7
	if pieceType == Bishop {
		directions = []int8{UpLeft, UpRight, DownLeft, DownRight}
	} else if pieceType == Rook {
		directions = []int8{UP, DOWN, LEFT, RIGHT}
	} else if pieceType == Queen || pieceType == King {
		directions = []int8{LEFT, RIGHT, UP, DOWN, UpLeft, UpRight, DownLeft, DownRight}
	} else {
		panic(fmt.Sprintf("PieceType is not a slider : %d", piece))
	}

	if pieceType == King {
		maxMoves = 1
	}

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
				if target.Color() == pieceColor {
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
	var (
		moves         []int8
		capturesMoves []int8
	)
	offsets := []int8{-17, -15, -10, -6, 6, 10, 15, 17}
	piece := p.PieceAt(pieceIdx)
	pieceFile := pieceIdx % 8
	pieceRank := pieceIdx / 8
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
