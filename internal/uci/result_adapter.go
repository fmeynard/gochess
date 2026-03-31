package uci

import (
	"chessV2/internal/search"
	"time"
)

type searchResultLike interface {
	searchDepth() int
	searchNodes() uint64
	searchTime() time.Duration
	searchScore() int32
	bestMoveUCI() string
}

type resultAdapter struct {
	result search.Result
}

func adaptResult(result search.Result) resultAdapter {
	return resultAdapter{result: result}
}

func (r resultAdapter) searchDepth() int {
	return r.result.Stats.Depth
}

func (r resultAdapter) searchNodes() uint64 {
	return r.result.Stats.Nodes
}

func (r resultAdapter) searchTime() time.Duration {
	return r.result.Stats.Time
}

func (r resultAdapter) searchScore() int32 {
	return int32(r.result.Score)
}

func (r resultAdapter) bestMoveUCI() string {
	return r.result.BestMove.UCI()
}
