package internal

import (
	"fmt"
)

// direction offsets
const (
	LEFT      = -1
	RIGHT     = 1
	UP        = -8
	DOWN      = 8
	UpLeft    = -9
	UpRight   = -7
	DownLeft  = 7
	DownRight = 9
)

func SliderPseudoLegalMoves(p Position, pieceIdx int8) ([]int8, []int8) {
	var (
		directions    []int
		moves         []int8
		capturesMoves []int8
	)

	piece := p.PieceAt(pieceIdx)
	if piece == NoPiece {
		return nil, nil
	}

	pieceColor := piece.Color()

	if piece.IsType(Bishop) {
		directions = []int{UpLeft, UpRight, DownLeft, DownRight}
	} else if piece.IsType(Rook) {
		directions = []int{UP, DOWN, LEFT, RIGHT}
	} else if piece.IsType(Queen) {
		directions = []int{LEFT, RIGHT, UP, DOWN, UpLeft, UpRight, DownLeft, DownRight}
	} else {
		panic(fmt.Sprintf("PieceType is not a slider : %d", piece))
	}

	for _, direction := range directions {
		for i := 1; i < 8; i++ { // start at 1 because 0 is current square
			targetIdx := pieceIdx + int8(direction*i)

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

		targetFile := targetIdx % 8
		targetRank := targetIdx / 8

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
