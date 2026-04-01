package uci

import (
	"bufio"
	board "chessV2/internal/board"
	"chessV2/internal/engine"
	"chessV2/internal/search"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	engineName   = "GoChess"
	engineAuthor = "fmeynard"
)

type Server struct {
	mu           sync.Mutex
	writeMu      sync.Mutex
	engine       *engine.Engine
	position     *board.Position
	positionKeys []uint64
	activeSearch *activeSearch
}

type activeSearch struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
}

func NewServer(e *engine.Engine) (*Server, error) {
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	if err != nil {
		return nil, err
	}

	return &Server{
		engine:       e,
		position:     pos,
		positionKeys: []uint64{pos.ZobristKey()},
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
		s.stopSearch(true)
		fmt.Fprintln(out, "readyok")
	case "ucinewgame":
		s.stopSearch(true)
		s.engine.StartGame()
		return false, s.resetToStartPos()
	case "position":
		s.stopSearch(true)
		return false, s.handlePosition(fields[1:])
	case "go":
		return false, s.handleGo(fields[1:], out)
	case "stop":
		s.stopSearch(false)
		return false, nil
	case "quit":
		s.stopSearch(true)
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
		positionKeys, err := s.engine.ApplyUCIMovesWithPositionKeys(pos, args[i+1:])
		if err != nil {
			return err
		}
		s.position = pos
		s.positionKeys = positionKeys
		return nil
	}

	s.position = pos
	s.positionKeys = []uint64{pos.ZobristKey()}
	return nil
}

func (s *Server) handleGo(args []string, out io.Writer) error {
	limits, err := parseGoLimits(args)
	if err != nil {
		return err
	}

	s.stopSearch(true)

	snapshot, history := s.searchSnapshot()
	active := &activeSearch{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	s.mu.Lock()
	s.activeSearch = active
	s.mu.Unlock()

	go s.runSearch(active, snapshot, history, limits, out)
	return nil
}

func (s *Server) resetToStartPos() error {
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	if err != nil {
		return err
	}
	s.position = pos
	s.positionKeys = []uint64{pos.ZobristKey()}
	return nil
}

func (s *Server) searchSnapshot() (*board.Position, []uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	history := make([]uint64, len(s.positionKeys))
	copy(history, s.positionKeys)
	return s.position.Clone(), history
}

func (s *Server) runSearch(active *activeSearch, pos *board.Position, history []uint64, parsedLimits struct {
	Depth    int
	MoveTime time.Duration
}, out io.Writer) {
	defer close(active.done)

	result, err := s.engine.Search(pos, search.Limits{
		Depth:    parsedLimits.Depth,
		MoveTime: parsedLimits.MoveTime,
		Stop:     active.stop,
		History:  history,
	})

	s.mu.Lock()
	if s.activeSearch == active {
		s.activeSearch = nil
	}
	s.mu.Unlock()

	if err != nil {
		s.writef(out, "info string error %s\n", sanitizeInfo(err.Error()))
		return
	}

	result = s.ensureBestMove(pos, result)
	s.writeResult(out, result)
}

func (s *Server) stopSearch(wait bool) {
	s.mu.Lock()
	active := s.activeSearch
	s.mu.Unlock()

	if active == nil {
		return
	}

	active.once.Do(func() {
		close(active.stop)
	})

	if wait {
		<-active.done
	}
}

func (s *Server) writeResult(out io.Writer, result search.Result) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	writeInfo(out, adaptResult(result))
	bestMove := "0000"
	if result.BestMove != (board.Move{}) {
		bestMove = result.BestMove.UCI()
	}
	if bestMove == "" {
		bestMove = "0000"
	}
	fmt.Fprintf(out, "bestmove %s\n", bestMove)
}

func (s *Server) writef(out io.Writer, format string, args ...any) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	fmt.Fprintf(out, format, args...)
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

func (s *Server) ensureBestMove(pos *board.Position, result search.Result) search.Result {
	if result.BestMove != (board.Move{}) {
		return result
	}

	moves := s.engine.LegalMoves(pos)
	if len(moves) > 0 {
		result.BestMove = moves[0]
	}
	return result
}
