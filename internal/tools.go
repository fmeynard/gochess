package internal

import (
	"fmt"
	"math/bits"
	"strings"
)

func SquareToIdx(square string) int8 {
	if len(square) != 2 {
		panic(fmt.Sprintf("invalid square identifier: %s", square))
	}

	file := square[0] - 'a'
	rank := square[1] - '1'

	if file < 0 || file > 7 {
		panic(fmt.Sprintf("invalid file identifier: %s", string(square[0])))
	}

	if rank < 0 || rank > 7 {
		panic(fmt.Sprintf("invalid rank identifier: %s", string(square[1])))
	}

	return int8(rank*8 + file)
}

func IdxToSquare(idx int8) string {
	if idx < 0 || idx > 63 {
		panic("idx out of range")
	}

	file := idx % 8
	rank := idx / 8

	return fmt.Sprintf("%c%d", 'a'+file, rank+1)
}

func RankAndFile(idx int8) (int8, int8) {
	return RankFromIdx(idx), FileFromIdx(idx)
}

func RankFromIdx(idx int8) int8 {
	return idx >> 3
}

func FileFromIdx(idx int8) int8 {
	return idx & 7
}

func absInt8(x int8) int8 {
	if x < 0 {
		return -x
	}
	return x
}

func leastSignificantOne(bb uint64) int8 {
	return int8(bits.TrailingZeros64(bb))
}

// mostSignificantBit returns the position of the highest set bit (most significant bit)
func mostSignificantBit(x uint64) int8 {
	return int8(63 - bits.LeadingZeros64(x))
}

func movesToUci(moves []Move) []string {
	var uciMoves []string
	for _, move := range moves {
		uciMoves = append(uciMoves, move.UCI())
	}

	return uciMoves
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

// isOnBoard checks if the given file and rank are within the bounds of the board.
func isOnBoard(file, rank int8) bool {
	return file >= 0 && file < 8 && rank >= 0 && rank < 8
}

func isSameLineOrRow(start, end, direction int8) bool {
	switch direction {
	case 1, -1: // Horizontal
		return start/8 == end/8
	case 8, -8: // Vertical
		return start%8 == end%8
	default: // Diagonals
		return absInt8(start%8-end%8) == absInt8(start/8-end/8)
	}
}
