package search

import (
	"errors"
	board "chessV2/internal/board"
	"chessV2/internal/eval"
	"chessV2/internal/movegen"
	"time"
)

var ErrNotImplemented = errors.New("search not implemented")
var ErrInvalidLimits = errors.New("invalid search limits")

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
	if limits.Depth <= 0 {
		return Result{}, ErrInvalidLimits
	}

	start := time.Now()
	stats := Stats{Depth: limits.Depth}

	var moves [256]board.Move
	moveCount := s.moveGenerator.LegalMovesInto(pos, s.positionUpdater, moves[:])
	if moveCount == 0 {
		return Result{
			Score: terminalScore(pos, 0),
			Stats: Stats{
				Nodes: 1,
				Depth: limits.Depth,
				Time:  time.Since(start),
			},
		}, nil
	}

	bestMove := moves[0]
	bestScore := -eval.InfinityScore
	alpha := -eval.InfinityScore
	beta := eval.InfinityScore

	for i := 0; i < moveCount; i++ {
		move := moves[i]
		history := s.positionUpdater.MakeMove(pos, move)
		score := -s.negamax(pos, limits.Depth-1, 1, -beta, -alpha, &stats)
		s.positionUpdater.UnMakeMove(pos, history)

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		if score > alpha {
			alpha = score
		}
	}

	stats.Time = time.Since(start)
	return Result{
		BestMove: bestMove,
		Score:    bestScore,
		Stats:    stats,
	}, nil
}

func (s *AlphaBetaSearcher) NewGame() {}

func (s *AlphaBetaSearcher) negamax(pos *board.Position, depth int, ply int, alpha eval.Score, beta eval.Score, stats *Stats) eval.Score {
	stats.Nodes++

	var moves [256]board.Move
	moveCount := s.moveGenerator.LegalMovesInto(pos, s.positionUpdater, moves[:])
	if moveCount == 0 {
		return terminalScore(pos, ply)
	}

	if depth == 0 {
		return s.evaluator.Evaluate(pos)
	}

	bestScore := -eval.InfinityScore
	for i := 0; i < moveCount; i++ {
		move := moves[i]
		history := s.positionUpdater.MakeMove(pos, move)
		score := -s.negamax(pos, depth-1, ply+1, -beta, -alpha, stats)
		s.positionUpdater.UnMakeMove(pos, history)

		if score > bestScore {
			bestScore = score
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break
		}
	}

	return bestScore
}

func terminalScore(pos *board.Position, ply int) eval.Score {
	if movegen.IsKingInCheck(pos, pos.ActiveColor()) {
		return eval.MatedIn(ply)
	}
	return eval.DrawScore
}

