package search

import (
	"errors"
	board "chessV2/internal/board"
	"chessV2/internal/eval"
	"chessV2/internal/movegen"
	"time"
)

var ErrNotImplemented = errors.New("search not implemented")

type Limits struct {
	Depth    int
	MoveTime time.Duration
}

type Stats struct {
	Nodes uint64
	Depth int
	Time  time.Duration
}

type Result struct {
	BestMove board.Move
	Score    eval.Score
	Stats    Stats
}

type Searcher interface {
	Search(pos *board.Position, limits Limits) (Result, error)
	NewGame()
}

type AlphaBetaSearcher struct {
	moveGenerator   *movegen.PseudoLegalMoveGenerator
	positionUpdater board.MoveApplier
	evaluator       eval.Evaluator
}

func NewAlphaBetaSearcher(moveGenerator *movegen.PseudoLegalMoveGenerator, positionUpdater board.MoveApplier, evaluator eval.Evaluator) *AlphaBetaSearcher {
	return &AlphaBetaSearcher{
		moveGenerator:   moveGenerator,
		positionUpdater: positionUpdater,
		evaluator:       evaluator,
	}
}

func (s *AlphaBetaSearcher) Search(pos *board.Position, limits Limits) (Result, error) {
	return Result{}, ErrNotImplemented
}

func (s *AlphaBetaSearcher) NewGame() {}

