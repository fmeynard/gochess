package search

import (
	"errors"
	board "chessV2/internal/board"
	"chessV2/internal/eval"
	"chessV2/internal/movegen"
	"time"
)

var ErrInvalidLimits = errors.New("invalid search limits")
var errSearchTimeout = errors.New("search timeout")

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
	if limits.Depth <= 0 && limits.MoveTime <= 0 {
		return Result{}, ErrInvalidLimits
	}

	if limits.MoveTime > 0 {
		return s.searchIterative(pos, limits)
	}
	return s.searchDepth(pos, limits.Depth, time.Time{})
}

func (s *AlphaBetaSearcher) searchIterative(pos *board.Position, limits Limits) (Result, error) {
	start := time.Now()
	deadline := start.Add(limits.MoveTime)
	maxDepth := limits.Depth
	if maxDepth <= 0 {
		maxDepth = 64
	}

	var lastComplete Result
	var haveComplete bool

	for depth := 1; depth <= maxDepth; depth++ {
		result, err := s.searchDepth(pos, depth, deadline)
		if err != nil {
			if errors.Is(err, errSearchTimeout) {
				if haveComplete {
					lastComplete.Stats.Time = time.Since(start)
					return lastComplete, nil
				}

				var fallbackMoves [256]board.Move
				moveCount := s.moveGenerator.LegalMovesInto(pos, s.positionUpdater, fallbackMoves[:])
				if moveCount == 0 {
					return Result{
						Score: terminalScore(pos, 0),
						Stats: Stats{
							Nodes: 1,
							Time:  time.Since(start),
						},
					}, nil
				}

				return Result{
					BestMove: fallbackMoves[0],
					Score:    eval.DrawScore,
					Stats: Stats{
						Depth: 0,
						Time:  time.Since(start),
					},
				}, nil
			}
			return Result{}, err
		}

		lastComplete = result
		haveComplete = true
		if time.Now().After(deadline) {
			break
		}
	}

	lastComplete.Stats.Time = time.Since(start)
	return lastComplete, nil
}

func (s *AlphaBetaSearcher) searchDepth(pos *board.Position, depth int, deadline time.Time) (Result, error) {
	if depth <= 0 {
		return Result{}, ErrInvalidLimits
	}

	start := time.Now()
	stats := Stats{Depth: depth}

	var moves [256]board.Move
	moveCount := s.moveGenerator.LegalMovesInto(pos, s.positionUpdater, moves[:])
	if moveCount == 0 {
		return Result{
			Score: terminalScore(pos, 0),
			Stats: Stats{
				Nodes: 1,
				Depth: depth,
				Time:  time.Since(start),
			},
		}, nil
	}

	bestMove := moves[0]
	bestScore := -eval.InfinityScore
	alpha := -eval.InfinityScore
	beta := eval.InfinityScore

	for i := 0; i < moveCount; i++ {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return Result{}, errSearchTimeout
		}

		move := moves[i]
		history := s.positionUpdater.MakeMove(pos, move)
		score, err := s.negamax(pos, depth-1, 1, -beta, -alpha, &stats, deadline)
		s.positionUpdater.UnMakeMove(pos, history)
		if err != nil {
			return Result{}, err
		}
		score = -score

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

func (s *AlphaBetaSearcher) negamax(pos *board.Position, depth int, ply int, alpha eval.Score, beta eval.Score, stats *Stats, deadline time.Time) (eval.Score, error) {
	if !deadline.IsZero() && time.Now().After(deadline) {
		return 0, errSearchTimeout
	}

	stats.Nodes++

	var moves [256]board.Move
	moveCount := s.moveGenerator.LegalMovesInto(pos, s.positionUpdater, moves[:])
	if moveCount == 0 {
		return terminalScore(pos, ply), nil
	}

	if depth == 0 {
		return s.evaluator.Evaluate(pos), nil
	}

	bestScore := -eval.InfinityScore
	for i := 0; i < moveCount; i++ {
		move := moves[i]
		history := s.positionUpdater.MakeMove(pos, move)
		score, err := s.negamax(pos, depth-1, ply+1, -beta, -alpha, stats, deadline)
		s.positionUpdater.UnMakeMove(pos, history)
		if err != nil {
			return 0, err
		}
		score = -score

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

	return bestScore, nil
}

func terminalScore(pos *board.Position, ply int) eval.Score {
	if movegen.IsKingInCheck(pos, pos.ActiveColor()) {
		return eval.MatedIn(ply)
	}
	return eval.DrawScore
}
