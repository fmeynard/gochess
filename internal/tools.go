package internal

import (
	"fmt"
)

func SquareToIdx(square string) int {
	if len(square) != 2 {
		panic(fmt.Sprintf("invalid square identifier: %s", square))
	}

	file := square[0] - 'a'
	rank := square[1] - '1'

	if file < 0 || file > 8 {
		panic(fmt.Sprintf("invalid file identifier: %s", string(square[0])))
	}

	if rank < 0 || rank > 8 {
		panic(fmt.Sprintf("invalid rank identifier: %s", string(square[1])))
	}

	return int(rank*8 + file)
}

func IdxToSquare(idx int) string {
	if idx < 0 || idx > 63 {
		panic("idx out of range")
	}

	file := idx % 8
	rank := idx / 8

	return fmt.Sprintf("%c%d", 'a'+file, rank+1)
}

func IndexesToSquares(indexes []int) []string {
	squares := make([]string, len(indexes))
	for i, idx := range indexes {
		squares[i] = IdxToSquare(idx)
	}

	return squares
}

func SquaresToIndexes(squares []string) []int {
	indexes := make([]int, len(squares))
	for i, square := range squares {
		indexes[i] = SquareToIdx(square)
	}

	return indexes
}
