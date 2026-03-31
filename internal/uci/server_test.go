package uci

import (
	"bytes"
	board "chessV2/internal/board"
	"chessV2/internal/engine"
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
