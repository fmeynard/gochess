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

func (p Position) SliderPseudoLegalMoves(pieceIdx int) ([]int, []int) {
	var (
		directions    []int
		moves         []int
		capturesMoves []int
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
