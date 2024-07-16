package internal

import (
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
			fenString: "8/8/2PPPPP1/2PPPPP1/2PPqPP1/2PPPPP1/2PPPPP1/8 b - - 0 1",
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
			fenString: "8/8/8/2pPp3/2PbP3/2pPp3/8/8 b - - 0 1",
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

	for tName, d := range data {
		t.Run(tName, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fenString)
			if err != nil {
				t.Fatal(err.Error())
			}

			moves := generator.SliderPseudoLegalMoves(&pos, d.piecePos, pos.board[d.piecePos].Type())
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

	for tName, d := range data {
		t.Run(tName, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fenString)
			if err != nil {
				t.Fatal(err.Error())
			}

			moves := generator.KnightPseudoLegalMoves(&pos, d.piecePos)
			assert.ElementsMatch(t, d.moves, moves, "Moves do not match")
		})
	}
}

func TestKingPseudoLegalMoves(t *testing.T) {
	data := map[string]struct {
		fenString string
		piecePos  int8
		moves     []int8
	}{
		"Start pos no moves white": {
			fenString: FenStartPos,
			piecePos:  E1,
		},
		"All directions white - no capture": {
			fenString: "8/8/8/8/8/8/2k5/8 w - - 0 1",
			piecePos:  C2,
			moves:     []int8{B1, C1, D1, B2, D2, B3, C3, D3},
		},
		"All directions capture": {
			fenString: "8/1QRB4/1PkP4/1BRQ4/8/8/8/8 b - - 0 1",
			piecePos:  C6,
			moves:     []int8{B7, C7, D7, B6, D6, B5, C5, D5},
		},
		"All directions no capture same color": {
			fenString: "8/1QRB4/1PKP4/1BRQ4/8/8/8/8 w - - 0 1",
			piecePos:  C6,
		},
		"QueenSide castle white": {
			fenString: "rnbqkbnr/pppppppp/8/8/8/B1NP4/PPPQPPPP/R3KBNR w KQkq - 0 1",
			piecePos:  E1,
			moves:     []int8{C1, D1},
		},
		"QueenSide castle white - Only kingSide allowed": {
			fenString: "rnbqkbnr/pppppppp/8/8/8/B1NP4/PPPQPPPP/R3KBNR w Kkq - 0 1",
			piecePos:  E1,
			moves:     []int8{D1},
		},
		"QueenSide castle white - no rights": {
			fenString: "rnbqkbnr/pppppppp/8/8/8/B1NP4/PPPQPPPP/R3KBNR w kq - 0 1",
			piecePos:  E1,
			moves:     []int8{D1},
		},
		"KingSide castle black": {
			fenString: "rnbqk2r/ppppnppp/3bp3/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
			piecePos:  E8,
			moves:     []int8{F8, G8},
		},
		"KingSide castle black - Only queen allowed": {
			fenString: "rnbqk2r/ppppnppp/3bp3/8/8/8/PPPPPPPP/RNBQKBNR b KQq - 0 1",
			piecePos:  E8,
			moves:     []int8{F8},
		},
		"KingSide castle black - ro rights": {
			fenString: "rnbqk2r/ppppnppp/3bp3/8/8/8/PPPPPPPP/RNBQKBNR b - - 0 1",
			piecePos:  E8,
			moves:     []int8{F8},
		},
	}

	generator := NewBitsBoardMoveGenerator()

	for tName, d := range data {
		t.Run(tName, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fenString)
			if err != nil {
				t.Fatal(err.Error())
			}

			moves := generator.KingPseudoLegalMoves(&pos, d.piecePos)
			assert.ElementsMatch(t, d.moves, moves, "Moves do not match")
		})
	}
}

func TestPawnPseudoLegalMoves(t *testing.T) {
	data := map[string]struct {
		fenString string
		piecePos  int8
		moves     []int8
	}{
		"start pos white": {
			fenString: "8/8/8/8/8/8/P7/8 w - - 0 1",
			piecePos:  A2,
			moves:     []int8{A3, A4},
		},
		"start pos black": {
			fenString: "8/p7/8/8/8/8/8/8 b - - 0 1",
			piecePos:  A7,
			moves:     []int8{A6, A5},
		},
		"no capture left white": {
			fenString: "8/8/8/pp5p/P7/8/8/8 w - - 0 1",
			piecePos:  A4,
			moves:     []int8{B5},
		},
		"no capture left black": {
			fenString: "8/8/8/p7/1P5P/8/8/8 b - - 0 1",
			piecePos:  A5,
			moves:     []int8{A4, B4},
		},
		"both capture white": {
			fenString: "8/8/8/8/8/ppp4p/1P6/8 w - - 0 1",
			piecePos:  B2,
			moves:     []int8{A3, C3},
		},
		"both capture black": {
			fenString: "8/8/8/8/3PPP2/3PpP2/PPPPPPPP/8 b - - 0 1",
			piecePos:  E3,
			moves:     []int8{D2, F2},
		},
		"enPassant left white": {
			fenString: "8/8/1p6/pPp5/8/8/8/8 w - a6 0 1",
			piecePos:  B5,
			moves:     []int8{A6},
		},
		"enPassant right white": {
			fenString: "8/8/1p6/pPp5/8/8/8/8 w - c6 0 1",
			piecePos:  B5,
			moves:     []int8{C6},
		},
		"no enPassant white": {
			fenString: "8/8/1p6/pPp5/8/8/8/8 w - - 0 1",
			piecePos:  B5,
		},
		"enPassant left black": {
			fenString: "8/8/8/8/PpP5/1P6/8/8 b - a3 0 1",
			piecePos:  B4,
			moves:     []int8{A3},
		},
		"No-EnPassant cross-board black": {
			fenString: "8/8/8/8/PpP4p/1P6/8/8 b - a3 0 1",
			piecePos:  H4,
			moves:     []int8{H3},
		},
		"Knight blocking double pawn move": {
			fenString: "8/8/8/8/8/5N2/5P2/8 w - - 0 1",
			piecePos:  F2,
		},
		"En passant only right": {
			fenString: "rnbqkbnr/p1p1ppp1/7p/1pPp4/8/8/PP1PPPPP/RNBQKBNR w KQkq d6 0 1",
			piecePos:  C5,
			moves:     []int8{C6, D6},
		},
	}

	generator := NewBitsBoardMoveGenerator()

	for tName, d := range data {
		t.Run(tName, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fenString)
			if err != nil {
				t.Fatal(err.Error())
			}

			moves, _ := generator.PawnPseudoLegalMoves(&pos, d.piecePos)
			assert.ElementsMatch(t, d.moves, moves, "Moves do not match")
		})
	}
}

func BenchmarkRookPseudoLegalMoves(b *testing.B) {
	pos, _ := NewPositionFromFEN("3p4/8/8/8/3R3P/3p4/8/8 w - - 0 1")
	generator := NewBitsBoardMoveGenerator()

	b.ResetTimer()

	for i := 0; i < 100000000; i++ {
		generator.SliderPseudoLegalMoves(&pos, D4, Rook)
	}
}

func BenchmarkQueenPseudoLegalMoves(b *testing.B) {
	// queen
	pos, _ := NewPositionFromFEN("3p4/8/8/4b3/2pq3P/3P4/8/P5p1 b - - 0 1")
	generator := NewBitsBoardMoveGenerator()

	b.ResetTimer()

	for i := 0; i < 100000000; i++ {
		generator.SliderPseudoLegalMoves(&pos, D4, Queen)
	}
}

func BenchmarkKingPseudoLegalMoves(b *testing.B) {
	pos, _ := NewPositionFromFEN("8/8/8/3P4/3K4/2p5/8/8 w KQkq - 0 1")
	generator := NewBitsBoardMoveGenerator()

	b.ResetTimer()

	for i := 0; i < 100000000; i++ {
		generator.KingPseudoLegalMoves(&pos, D4)
	}
}

func BenchmarkKnightPseudoLegalMoves(b *testing.B) {
	pos, _ := NewPositionFromFEN("8/8/8/8/4N3/2P5/3p4/8 w - - 0 1")
	generator := NewBitsBoardMoveGenerator()

	b.ResetTimer()

	for i := 0; i < 100000000; i++ {
		generator.KnightPseudoLegalMoves(&pos, E4)
	}
}

func BenchmarkPawnPseudoLegalMoves(b *testing.B) {
	pos, _ := NewPositionFromFEN("rnbqkbnr/p1p1ppp1/7p/1pPp4/8/8/PP1PPPPP/RNBQKBNR w KQkq d6 0 1")
	generator := NewBitsBoardMoveGenerator()

	b.ResetTimer()

	for i := 0; i < 100000000; i++ {
		generator.PawnPseudoLegalMoves(&pos, C5)
	}
}

func BenchmarkPerftDivideB(b *testing.B) {
	pos, _ := NewPositionFromFEN("rnbqkbnr/2pppppp/8/pp6/2P5/N7/PP1PPPPP/R1BQKBNR w KQkq - 0 3")
	generator := NewBitsBoardMoveGenerator()
	generator.PerftDivide(pos, 2)
}

func TestPerftDivideB(t *testing.T) {
	pos, _ := NewPositionFromFEN(FenStartPos)
	pos = pos.PositionAfterMove(NewMove(Piece(White|Pawn), E2, E3, NormalMove))
	pos = pos.PositionAfterMove(NewMove(Piece(Black|Pawn), A7, A6, NormalMove))
	//pos = pos.PositionAfterMove(NewMove(Piece(White|Rook), F1, B5, NormalMove))

	draw(pos.occupied)

	generator := NewBitsBoardMoveGenerator()
	generator.PerftDivide(pos, 4)
}
