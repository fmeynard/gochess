package internal

import (
	"fmt"
	"math/bits"
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

func IsSameDiagonal(pieceRank, pieceFile, targetRank, targetFile int8) bool {
	return absInt8(pieceFile-targetFile) == absInt8(pieceRank-targetRank)
}

func leastSignificantOne(bb uint64) int8 {
	return int8(bits.TrailingZeros64(bb))
}

func isSameLine(fromIdx, toIdx, direction int8) bool {
	if direction == 1 || direction == -1 { // Horizontal movement
		return fromIdx/8 == toIdx/8
	}

	if direction == 8 || direction == -8 { // Vertical movement
		return fromIdx%8 == toIdx%8
	}

	return absInt8((fromIdx%8)-(toIdx%8)) == absInt8((fromIdx/8)-(toIdx/8))
}

func canReach(pieceType int8, direction int8) bool {
	switch pieceType {
	case Rook:
		return direction == 1 || direction == -1 || direction == 8 || direction == -8
	case Bishop:
		return direction == 7 || direction == -7 || direction == 9 || direction == -9
	case Queen:
		return true
	case King, Knight:
		return true // Kings and Knights move in all directions but limited distance
	default:
		return false
	}
}
