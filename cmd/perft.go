package main

import (
	"chessV2/internal"
	"fmt"
	"os"
	"sort"
	"strconv"
)

func main() {
	fenPos := internal.FenStartPos
	if len(os.Args) != 3 {
		panic("Invalid number of arguments")
	}

	fenPos = os.Args[1]
	depth, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	engine := internal.NewEngine()
	pos, _ := internal.NewPositionFromFEN(fenPos)
	res, nodesCount := engine.PerftDivide(&pos, depth)

	keys := make([]string, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Println(fmt.Sprintf("%s:", k), res[k])
	}

	fmt.Println()

	fmt.Println("Nodes searched:", nodesCount)
}
