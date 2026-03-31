package uci

import (
	"bufio"
	board "chessV2/internal/board"
	"chessV2/internal/engine"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	engineName   = "GoChess"
	engineAuthor = "fmeynard"
)

type Server struct {
	engine   *engine.Engine
	position *board.Position
}

func NewServer(e *engine.Engine) (*Server, error) {
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	if err != nil {
		return nil, err
	}

	return &Server{
		engine:   e,
		position: pos,
	}, nil
}

func (s *Server) Run(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		quit, err := s.handleCommand(line, out)
		if err != nil {
			fmt.Fprintf(out, "info string error %s\n", sanitizeInfo(err.Error()))
		}
		if quit {
			return nil
		}
	}

	return scanner.Err()
}

func (s *Server) handleCommand(line string, out io.Writer) (bool, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false, nil
	}

	switch fields[0] {
	case "uci":
		fmt.Fprintf(out, "id name %s\n", engineName)
		fmt.Fprintf(out, "id author %s\n", engineAuthor)
		fmt.Fprintln(out, "uciok")
	case "isready":
		fmt.Fprintln(out, "readyok")
	case "ucinewgame":
		s.engine.StartGame()
		return false, s.resetToStartPos()
	case "position":
		return false, s.handlePosition(fields[1:])
	case "go":
		return false, s.handleGo(fields[1:], out)
	case "stop":
		return false, nil
	case "quit":
		return true, nil
	}

	return false, nil
}

func (s *Server) handlePosition(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing position arguments")
	}

	var pos *board.Position
	var err error
	i := 0

	switch args[0] {
	case "startpos":
		pos, err = board.NewPositionFromFEN(board.FenStartPos)
		if err != nil {
			return err
		}
		i = 1
	case "fen":
		if len(args) < 7 {
			return fmt.Errorf("invalid fen position command")
		}
		fen := strings.Join(args[1:7], " ")
		pos, err = board.NewPositionFromFEN(fen)
		if err != nil {
			return err
		}
		i = 7
	default:
		return fmt.Errorf("unsupported position command")
	}

	if i < len(args) {
		if args[i] != "moves" {
			return fmt.Errorf("unexpected token in position command: %s", args[i])
		}
		if err := s.engine.ApplyUCIMoves(pos, args[i+1:]); err != nil {
			return err
		}
	}

	s.position = pos
	return nil
}

func (s *Server) handleGo(args []string, out io.Writer) error {
	limits, err := parseGoLimits(args)
	if err != nil {
		return err
	}

	if limits.MoveTime > 0 {
		result, err := s.engine.SearchTime(s.position, limits.MoveTime)
		if err != nil {
			return err
		}
		writeInfo(out, adaptResult(result))
		fmt.Fprintf(out, "bestmove %s\n", result.BestMove.UCI())
		return nil
	}

	result, err := s.engine.SearchDepth(s.position, limits.Depth)
	if err != nil {
		return err
	}
	writeInfo(out, adaptResult(result))
	fmt.Fprintf(out, "bestmove %s\n", result.BestMove.UCI())
	return nil
}

func (s *Server) resetToStartPos() error {
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	if err != nil {
		return err
	}
	s.position = pos
	return nil
}

func parseGoLimits(args []string) (limits struct {
	Depth    int
	MoveTime time.Duration
}, err error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "depth":
			if i+1 >= len(args) {
				return limits, fmt.Errorf("missing go depth value")
			}
			depth, convErr := strconv.Atoi(args[i+1])
			if convErr != nil {
				return limits, fmt.Errorf("invalid go depth value")
			}
			limits.Depth = depth
			i++
		case "movetime":
			if i+1 >= len(args) {
				return limits, fmt.Errorf("missing go movetime value")
			}
			ms, convErr := strconv.Atoi(args[i+1])
			if convErr != nil {
				return limits, fmt.Errorf("invalid go movetime value")
			}
			limits.MoveTime = time.Duration(ms) * time.Millisecond
			i++
		}
	}

	if limits.MoveTime <= 0 && limits.Depth <= 0 {
		limits.Depth = 1
	}

	return limits, nil
}

func writeInfo(out io.Writer, result searchResultLike) {
	timeMs := result.searchTime().Milliseconds()
	score := result.searchScore()
	if score > 29000 || score < -29000 {
		matePly := int((30000 - absScore(score) + 1) / 2)
		if score < 0 {
			matePly = -matePly
		}
		fmt.Fprintf(out, "info depth %d nodes %d time %d score mate %d pv %s\n", result.searchDepth(), result.searchNodes(), timeMs, matePly, result.bestMoveUCI())
		return
	}

	fmt.Fprintf(out, "info depth %d nodes %d time %d score cp %d pv %s\n", result.searchDepth(), result.searchNodes(), timeMs, score, result.bestMoveUCI())
}

func sanitizeInfo(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

func absScore(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
