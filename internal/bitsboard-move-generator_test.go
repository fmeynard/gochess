package internal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSliderPseudoLegalMoves(t *testing.T) {
	data := map[string]struct {
		fenString string
		piecePos  int8
		moves     []int8
	}{
		"Queen From middle empty board": {
			fenString: "8/8/8/3Q4/8/8/8/8 w - - 0 1",
			piecePos:  D5,
			moves: []int8{
				A8, B7, C6, E4, F3, G2, H1, // diagonal top left -> bottom right
				A2, B3, C4, E6, F7, G8, // diagonal bottom left -> top right
				A5, B5, C5, E5, F5, G5, H5, // horizontal
				D1, D2, D3, D4, D6, D7, D8, // vertical
			},
		},
		"Queen From corner": {
			fenString: "8/8/8/8/8/8/8/7Q b - - 0 1",
			piecePos:  H1,
			moves: []int8{
				A1, B1, C1, D1, E1, F1, G1, // horizontal
				H2, H3, H4, H5, H6, H7, H8, // vertical
				G2, F3, E4, D5, C6, B7, A8, // diagonal
			},
		},
		"Queen Blocked all directions": {
			fenString: "8/8/2ppppp1/2pPPPp1/2pPQPp1/2pPPPp1/2ppppp1/8 w - - 0 1",
		},
		"Queen Captures all directions": {
			fenString: "8/8/2PPPPP1/2PPPPP1/2PPqPP1/2PPPPP1/2PPPPP1/8 w - - 0 1",
			piecePos:  E4,
			moves:     []int8{D3, D4, D5, E3, E5, F3, F4, F5},
		},
		"Bishop From middle empty board": {
			fenString: "8/8/8/8/3B4/8/8/8 w - - 0 1",
			piecePos:  D4,
			moves:     []int8{A1, B2, C3, E5, F6, G7, H8, G1, F2, E3, C5, B6, A7},
		},
		"Bishop From corner": {
			fenString: "b7/8/8/8/8/8/8/8 b - - 0 1",
			piecePos:  A8,
			moves:     []int8{B7, C6, D5, E4, F3, G2, H1},
		},
		"Bishop Blocked all directions": {
			fenString: "8/8/8/2pPp3/2PbP3/2pPp3/8/8 w - - 0 1",
			piecePos:  D4,
		},
		"Bishop Captures all directions": {
			fenString: "8/8/1P3P2/2P1P3/3b4/2P1P3/1P3P2/8 b - - 0 1",
			piecePos:  D4,
			moves:     []int8{C3, C5, E3, E5},
		},
		"Rook A1 - Empty board": {
			fenString: "8/8/8/8/8/8/8/R7 w - - 0 1",
			piecePos:  A1,
			moves:     []int8{A2, A3, A4, A5, A6, A7, A8, B1, C1, D1, E1, F1, G1, H1},
		},
		"Rook E4 - Blocked left&down - 1 up&right- No captures": {
			fenString: "8/8/4P3/8/3PR1P1/4P3/8/8 w - - 0 1",
			piecePos:  E4,
			moves:     []int8{F4, E5},
		},
		"Rook E4 - Blocked left&down - 1 up- Capture G4": {
			fenString: "8/8/4P3/8/3PR1p1/4P3/8/8 w - - 0 1",
			piecePos:  E4,
			moves:     []int8{F4, G4, E5},
		},
		"Rook E4 - 1 down - Capture A5,G4": {
			fenString: "rn2kbnr/1pp1pppp/8/p7/R5bq/8/P1PPPPPP/1NBQKBNR w - - 0 1",
			piecePos:  A4,
			moves:     []int8{A3, A5, B4, C4, D4, E4, F4, G4},
		},
	}

	generator := NewBitsBoardMoveGenerator()

	fmt.Println(fmt.Sprintf("%d -> %b ", generator.rookMasks[A1], generator.rookMasks[A1]))
	draw(generator.rookMasks[A1])
	for tName, d := range data {
		t.Run(tName, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fenString)
			if err != nil {
				t.Fatal(err.Error())
			}

			moves := generator.SliderPseudoLegalMoves(pos, d.piecePos)
			assert.ElementsMatch(t, d.moves, moves, "Moves do not match")
		})
	}
}

func TestKnightPseudoLegalMoves(t *testing.T) {
	data := map[string]struct {
		fenString string
		piecePos  int8
		moves     []int8
	}{
		"All Directions - No Captures": {
			fenString: "8/8/8/8/4N3/8/8/8 w - - 0 1",
			piecePos:  E4,
			moves:     []int8{11, 13, 18, 22, 34, 38, 43, 45},
		},
		"Captures All Directions": {
			fenString: "8/8/3P1P2/2P3P1/4n3/2P3P1/3P1P2/8 b - - 0 1",
			piecePos:  E4,
			moves:     []int8{C3, C5, D2, D6, F2, F6, G3, G5},
		},
		"B1 : Edges check": {
			fenString: "8/8/8/8/8/8/8/1N6 w - - 0 1",
			piecePos:  B1,
			moves:     []int8{A3, C3, D2},
		},
		"G8 : Edges check": {
			fenString: "8/8/8/8/4N3/8/8/8 w - - 0 1",
			piecePos:  G8,
			moves:     []int8{E7, H6, F6},
		},
		"No cross-board capture": {
			fenString: "8/p7/8/7N/8/8/8/8 w - - 0 1",
			piecePos:  H5,
			moves:     []int8{G7, F6, F4, G3},
		},
	}

	generator := NewBitsBoardMoveGenerator()

	draw(generator.knightMasks[E4])

	fmt.Println("--")
	draw(generator.knightMasks[E4] & (1 << D2))

	for tName, d := range data {
		t.Run(tName, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fenString)
			if err != nil {
				t.Fatal(err.Error())
			}

			moves := generator.KnightPseudoLegalMoves(pos, d.piecePos)
			assert.ElementsMatch(t, d.moves, moves, "Moves do not match")
		})
	}
}

func BenchmarkSliderPseudoLegalMoves(b *testing.B) {
	// queen
	//pos, _ := NewPositionFromFEN("3p4/8/8/4b3/2pq3P/3P4/8/P5p1 b - - 0 1")
	// rook
	pos, _ := NewPositionFromFEN("3p4/8/8/8/3R3P/3p4/8/8 w - - 0 1")
	generator := NewBitsBoardMoveGenerator()

	b.ResetTimer()

	for i := 0; i < 100000000; i++ {
		generator.SliderPseudoLegalMoves(pos, D4)
	}
}

func BenchmarkKnightPseudoLegalMoves(b *testing.B) {
	pos, _ := NewPositionFromFEN("8/8/8/8/4N3/2P5/3p4/8 w - - 0 1")
	generator := NewBitsBoardMoveGenerator()

	b.ResetTimer()

	for i := 0; i < 100000000; i++ {
		generator.KnightPseudoLegalMoves(pos, E4)
	}
}
