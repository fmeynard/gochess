package uci

import (
	"bytes"
	board "chessV2/internal/board"
	"chessV2/internal/engine"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerBasicHandshake(t *testing.T) {
	e := engine.NewEngine()
	server, err := NewServer(e)
	assert.NoError(t, err)

	var out bytes.Buffer
	err = server.Run(strings.NewReader("uci\nisready\nquit\n"), &out)
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "id name GoChess")
	assert.Contains(t, output, "uciok")
	assert.Contains(t, output, "readyok")
}

func TestServerPositionAndGoDepth(t *testing.T) {
	e := engine.NewEngine()
	server, err := NewServer(e)
	assert.NoError(t, err)

	var out bytes.Buffer
	input := "position startpos moves e2e4 e7e5\ngo depth 1\nquit\n"
	err = server.Run(strings.NewReader(input), &out)
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "info depth 1")
	assert.Contains(t, output, "bestmove ")
}

func TestServerPositionFenAndMoveTime(t *testing.T) {
	e := engine.NewEngine()
	server, err := NewServer(e)
	assert.NoError(t, err)

	var out bytes.Buffer
	input := "position fen 7k/5KQ1/8/8/8/8/8/8 w - - 0 1\ngo movetime 10\nquit\n"
	err = server.Run(strings.NewReader(input), &out)
	assert.NoError(t, err)
	assert.Contains(t, out.String(), "bestmove ")
}

func TestServerStopCancelsRunningSearch(t *testing.T) {
	e := engine.NewEngine()
	server, err := NewServer(e)
	assert.NoError(t, err)

	var out bytes.Buffer
	input := "position startpos\ngo movetime 1000\nstop\nquit\n"
	err = server.Run(strings.NewReader(input), &out)
	assert.NoError(t, err)
	assert.Contains(t, out.String(), "bestmove ")
}

func TestParseGoLimitsDefaultsDepthOne(t *testing.T) {
	limits, err := parseGoLimits(nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, limits.Depth)
}

func TestEngineApplyUCIMoves(t *testing.T) {
	e := engine.NewEngine()
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	assert.NoError(t, err)

	err = e.ApplyUCIMoves(pos, []string{"e2e4", "e7e5"})
	assert.NoError(t, err)
	assert.Equal(t, board.Piece(board.Pawn|board.White), pos.PieceAt(board.E4))
	assert.Equal(t, board.Piece(board.Pawn|board.Black), pos.PieceAt(board.E5))
}

func TestServerGoMoveTimeReturnsLegalBestMoveOnRegressionPositions(t *testing.T) {
	tests := map[string][]string{
		"game 38 position": {
			"b1a3", "a7a5", "g1f3", "d7d5", "a1b1", "a5a4", "b1a1", "e7e6",
			"a1b1", "b8c6", "b1a1", "e6e5", "f3e5", "c6e5", "a1b1", "d5d4",
			"b1a1", "f8a3", "b2a3", "e8f8", "a1b1", "f8e8", "b1b7",
		},
		"game 49 position": {
			"b1c3", "a7a5", "g1f3", "b8c6", "d2d4", "c6b4", "a1b1", "b4a2",
			"c3a2", "d7d6", "b1a1", "c8g4", "c1e3", "g4f3",
		},
	}

	for name, moves := range tests {
		t.Run(name, func(t *testing.T) {
			e := engine.NewEngine()
			server, err := NewServer(e)
			assert.NoError(t, err)

			pos, err := board.NewPositionFromFEN(board.FenStartPos)
			assert.NoError(t, err)
			assert.NoError(t, e.ApplyUCIMoves(pos, moves))
			legalMoves := movesToSet(e.LegalMoves(pos))

			var out bytes.Buffer
			input := fmt.Sprintf("position startpos moves %s\ngo movetime 300\nquit\n", strings.Join(moves, " "))
			err = server.Run(strings.NewReader(input), &out)
			assert.NoError(t, err)

			bestMove := parseBestMove(out.String())
			assert.NotEmpty(t, bestMove)
			assert.NotEqual(t, "0000", bestMove)
			assert.Contains(t, legalMoves, bestMove)
		})
	}
}

func parseBestMove(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "bestmove ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}
	return ""
}

func movesToSet(moves []board.Move) map[string]struct{} {
	set := make(map[string]struct{}, len(moves))
	for _, move := range moves {
		set[move.UCI()] = struct{}{}
	}
	return set
}
