package internal

import (
	"fmt"
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
	if idx < 0 || idx > 63 {
		panic("idx out of range")
	}

	return RankFromIdx(idx), FileFromIdx(idx)
}

func RankFromIdx(idx int8) int8 {
	return idx / 8
}

func FileFromIdx(idx int8) int8 {
	return idx % 8
}
