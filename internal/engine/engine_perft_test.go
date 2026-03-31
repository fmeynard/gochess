package engine

import (
	. "chessV2/internal/board"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngine_MoveGenerationTest_Regressions(t *testing.T) {
	// MoveGenerationTest() uses an internal depth convention where:
	// depth=2 matches standard perft(1), depth=3 matches perft(2), etc.
	data := map[string]struct {
		fen      string
		depth    int
		expected uint64
	}{
		"Start position depth 2": {
			fen:      FenStartPos,
			depth:    2,
			expected: 20,
		},
		"Start position depth 3": {
			fen:      FenStartPos,
			depth:    3,
			expected: 400,
		},
		"Start position depth 4": {
			fen:      FenStartPos,
			depth:    4,
			expected: 8902,
		},
		"Perft position 2 depth 3": {
			fen:      "r3k2r/p1ppqpb1/bn2pnp1/2PN4/1p2P3/2N2Q1p/PPPB2PP/R3K2R w KQkq - 0 1",
			depth:    3,
			expected: 1955,
		},
		"Perft position 3 depth 4": {
			fen:      "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
			depth:    4,
			expected: 2812,
		},
		"Perft position 4 depth 3": {
			fen:      "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
			depth:    3,
			expected: 264,
		},
		"Promotion position depth 3": {
			fen:      "2r1k3/3P4/8/8/8/8/8/4K3 w - - 0 1",
			depth:    3,
			expected: 73,
		},
	}

	engine := NewEngine()

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fen)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, d.expected, engine.MoveGenerationTest(pos, d.depth))
		})
	}
}

func TestEngine_MoveGenerationTest_KingCaptureIsNotCounted(t *testing.T) {
	engine := NewEngine()
	pos, err := NewPositionFromFEN("2r1k3/3P4/8/8/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(73), engine.MoveGenerationTest(pos, 3))
}
