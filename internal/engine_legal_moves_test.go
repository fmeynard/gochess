package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngine_LegalMoves_Regressions(t *testing.T) {
	data := map[string]struct {
		fen           string
		expectedMoves []string
	}{
		"Start position": {
			fen: FenStartPos,
			expectedMoves: []string{
				"a2a3", "a2a4", "b2b3", "b2b4", "c2c3", "c2c4", "d2d3", "d2d4",
				"e2e3", "e2e4", "f2f3", "f2f4", "g2g3", "g2g4", "h2h3", "h2h4",
				"b1a3", "b1c3", "g1f3", "g1h3",
			},
		},
		"Perft position 2 exact root moves": {
			fen: "r3k2r/p1ppqpb1/bn2pnp1/2PN4/1p2P3/2N2Q1p/PPPB2PP/R3K2R w KQkq - 0 1",
			expectedMoves: []string{
				"a1b1", "a1c1", "a1d1", "e1c1", "e1d1", "e1f2", "h1f1", "h1g1",
				"a2a3", "a2a4", "b2b3", "c3a4", "c3b1", "c3b5", "c3d1", "c3e2",
				"c5b6", "c5c6", "d2c1", "d2e3", "d2f4", "d2g5", "d2h6", "d5b4",
				"d5b6", "d5c7", "d5e3", "d5e7", "d5f4", "d5f6", "e4e5", "f3d1",
				"f3d3", "f3e2", "f3e3", "f3f1", "f3f2", "f3f4", "f3f5", "f3f6",
				"f3g3", "f3g4", "f3h3", "f3h5", "g2g3", "g2g4", "g2h3",
			},
		},
		"Check evasions from perft position 4": {
			fen: "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
			expectedMoves: []string{
				"b4c5",
				"c4c5",
				"d2d4",
				"f1f2",
				"f3d4",
				"g1h1",
			},
		},
		"Perft position 3 exact root moves": {
			fen: "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
			expectedMoves: []string{
				"a5a4", "a5a6", "b4a4", "b4b1", "b4b2", "b4b3", "b4c4",
				"b4d4", "b4e4", "b4f4", "e2e3", "e2e4", "g2g3", "g2g4",
			},
		},
		"Perft position 5 exact root moves": {
			fen: "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
			expectedMoves: []string{
				"a2a3", "a2a4", "b1a3", "b1c3", "b1d2", "b2b3", "b2b4", "c1d2",
				"c1e3", "c1f4", "c1g5", "c1h6", "c2c3", "c4a6", "c4b3", "c4b5",
				"c4d3", "c4d5", "c4e6", "c4f7", "d1d2", "d1d3", "d1d4", "d1d5",
				"d1d6", "d7c8b", "d7c8n", "d7c8q", "d7c8r", "e1d2", "e1f1", "e1f2",
				"e1g1", "e2c3", "e2d4", "e2f4", "e2g1", "e2g3", "g2g3", "g2g4",
				"h1f1", "h1g1", "h2h3", "h2h4",
			},
		},
		"Perft position 6 exact root moves": {
			fen: "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/2NP1N2/PPP1QPPP/R4RK1 w - - 0 10",
			expectedMoves: []string{
				"a1b1", "a1c1", "a1d1", "a1e1", "a2a3", "a2a4", "b2b3", "b2b4",
				"c3a4", "c3b1", "c3b5", "c3d1", "c3d5", "c4a6", "c4b3", "c4b5",
				"c4d5", "c4e6", "c4f7", "d3d4", "e2d1", "e2d2", "e2e1", "e2e3",
				"f1b1", "f1c1", "f1d1", "f1e1", "f3d2", "f3d4", "f3e1", "f3e5",
				"f3h4", "g1h1", "g2g3", "g5c1", "g5d2", "g5e3", "g5f4", "g5f6",
				"g5h4", "g5h6", "h2h3", "h2h4",
			},
		},
		"Promotion root moves": {
			fen: "2r1k3/3P4/8/8/8/8/8/4K3 w - - 0 1",
			expectedMoves: []string{
				"d7c8b", "d7c8n", "d7c8q", "d7c8r",
				"d7d8b", "d7d8n", "d7d8q", "d7d8r",
				"d7e8b", "d7e8n", "d7e8q", "d7e8r",
				"e1d1", "e1d2", "e1e2", "e1f1", "e1f2",
			},
		},
		"Illegal en passant discovered check": {
			fen: "4r1k1/8/8/3pP3/8/8/8/4K3 w - d6 0 1",
			expectedMoves: []string{
				"e1d1", "e1d2", "e1e2", "e1f1", "e1f2", "e5e6",
			},
		},
		"Legal en passant available": {
			fen: "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1",
			expectedMoves: []string{
				"e1d1", "e1d2", "e1e2", "e1f1", "e1f2", "e5d6", "e5e6",
			},
		},
		"Cannot castle through check": {
			fen: "5rk1/8/8/8/8/8/8/R3K2R w KQ - 0 1",
			expectedMoves: []string{
				"a1a2", "a1a3", "a1a4", "a1a5", "a1a6", "a1a7", "a1a8", "a1b1", "a1c1", "a1d1",
				"e1c1", "e1d1", "e1d2", "e1e2",
				"h1f1", "h1g1", "h1h2", "h1h3", "h1h4", "h1h5", "h1h6", "h1h7", "h1h8",
			},
		},
		"Cannot castle while in check": {
			fen: "4r1k1/8/8/8/8/8/8/R3K2R w KQ - 0 1",
			expectedMoves: []string{
				"e1d1", "e1d2", "e1f1", "e1f2",
			},
		},
		"Double check only king moves": {
			fen: "4r2k/8/8/1b6/8/8/4K3/8 w - - 0 1",
			expectedMoves: []string{
				"e2d1",
				"e2d2",
				"e2f2",
				"e2f3",
			},
		},
		"Pinned knight cannot move": {
			fen: "4r2k/8/8/8/8/8/4N3/4K3 w - - 0 1",
			expectedMoves: []string{
				"e1d1",
				"e1d2",
				"e1f1",
				"e1f2",
			},
		},
	}

	engine := NewEngine()

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fen)
			if err != nil {
				t.Fatal(err)
			}

			assert.ElementsMatch(t, d.expectedMoves, movesToUci(engine.LegalMoves(pos)))
		})
	}
}

func TestEngine_LegalMoves_KnownPositionCounts(t *testing.T) {
	data := map[string]struct {
		fen           string
		expectedCount int
	}{
		"Perft position 2": {
			fen:           "r3k2r/p1ppqpb1/bn2pnp1/2PN4/1p2P3/2N2Q1p/PPPB2PP/R3K2R w KQkq - 0 1",
			expectedCount: 47,
		},
		"Perft position 3": {
			fen:           "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
			expectedCount: 14,
		},
		"Perft position 4": {
			fen:           "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
			expectedCount: 6,
		},
		"Perft position 5": {
			fen:           "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
			expectedCount: 44,
		},
		"Perft position 6": {
			fen:           "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/2NP1N2/PPP1QPPP/R4RK1 w - - 0 10",
			expectedCount: 44,
		},
	}

	engine := NewEngine()

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fen)
			if err != nil {
				t.Fatal(err)
			}

			assert.Len(t, engine.LegalMoves(pos), d.expectedCount)
		})
	}
}
