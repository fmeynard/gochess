package internal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_PositionAfterMove(t *testing.T) {
	type CheckArgs struct {
		idx   int8
		piece Piece
	}

	data := map[string]struct {
		fenPos string
		move   Move
		checks []CheckArgs
	}{
		"pawn move from init FEN": {
			fenPos: FenStartPos,
			move:   NewMove(Piece(Pawn|White), D2, D4, NormalMove),
			checks: []CheckArgs{
				{D2, NoPiece},
				{D4, Piece(Pawn | White)},
			},
		},
		"en passant capture": {
			fenPos: "rnbqkbnr/p1pppppp/8/Pp6/8/8/1PPPPPPP/RNBQKBNR w KQkq b6 0 1",
			move:   NewMove(Piece(Pawn|White), A5, B6, EnPassant),
			checks: []CheckArgs{
				{A5, NoPiece},
				{B5, NoPiece},
				{B6, Piece(Pawn | White)},
			},
		},
		"Queen side castle white": {
			fenPos: "rnbqkbnr/pppppppp/8/8/5B2/2NP4/PPPQPPPP/R3KBNR w KQkq - 0 1",
			move:   NewMove(Piece(King|White), E1, C1, Castle),
			checks: []CheckArgs{
				{E1, NoPiece},
				{A1, NoPiece},
				{C1, Piece(King | White)},
				{D1, Piece(Rook | White)},
			},
		},
		"King side castle black": {
			fenPos: "rnbqk2r/ppppnppp/2b1p3/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
			move:   NewMove(Piece(King|Black), E8, G8, Castle),
			checks: []CheckArgs{
				{E8, NoPiece},
				{H8, NoPiece},
				{G8, Piece(King | Black)},
				{F8, Piece(Rook | Black)},
			},
		},
		"Bishop capture white": {
			fenPos: "rnbqk1nr/pppp1ppp/4p3/8/8/bPP5/P2PPPPP/RNBQKBNR w KQkq - 0 1",
			move:   NewMove(Piece(Bishop|White), C1, A3, Capture),
			checks: []CheckArgs{
				{C1, NoPiece},
				{A3, Piece(Bishop | White)},
			},
		},
	}

	updater := NewPositionUpdater(NewBitsBoardMoveGenerator())
	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fenPos)
			if err != nil {
				t.Error(err)
			}

			updater.MakeMove(pos, d.move)
			for _, check := range d.checks {
				assert.Equal(t, check.piece, pos.PieceAt(check.idx), fmt.Sprintf("Wrong piece at idx : %d", check.idx))
			}
		})
	}
}

func TestUpdatePieceOnBoard(t *testing.T) {
	type MaskCheckArgs struct {
		idx     int8
		isEmpty bool
	}
	data := map[string]struct {
		fenPos              string
		move                Move
		occupiedChecks      []MaskCheckArgs
		whiteOccupiedChecks []MaskCheckArgs
		blackOccupiedChecks []MaskCheckArgs
	}{
		"White queen-side castle": {
			fenPos: "8/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			move:   NewMove(Piece(King|White), E1, C1, Castle),
			occupiedChecks: []MaskCheckArgs{
				{C1, false},
				{D1, false},
				{E1, true},
				{A1, true},
			},
		},
		"White king-side castle": {
			fenPos: "8/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			move:   NewMove(Piece(King|White), E1, G1, Castle),
			occupiedChecks: []MaskCheckArgs{
				{G1, false},
				{F1, false},
				{E1, true},
				{H1, true},
			},
		},
		"Black queen-side castle": {
			fenPos: "r3k2r/8/8/8/8/8/8/8 b KQkq - 0 1",
			move:   NewMove(Piece(King|Black), E8, C8, Castle),
			occupiedChecks: []MaskCheckArgs{
				{C8, false},
				{D8, false},
				{E8, true},
				{A8, true},
			},
		},
		"Black king-side castle": {
			fenPos: "r3k2r/8/8/8/8/8/8/8 b KQkq - 0 1",
			move:   NewMove(Piece(King|Black), E8, G8, Castle),
			occupiedChecks: []MaskCheckArgs{
				{G8, false},
				{F8, false},
				{E8, true},
				{H8, true},
			},
		},
		"White capture": {
			fenPos: "3r4/8/8/8/8/8/8/3R4 w - - 0 1",
			move:   NewMove(Piece(Rook|White), D1, D8, NormalMove),
			occupiedChecks: []MaskCheckArgs{
				{D1, true},
				{D8, false},
			},
			whiteOccupiedChecks: []MaskCheckArgs{
				{D1, true},
				{D8, false},
			},
			blackOccupiedChecks: []MaskCheckArgs{
				{D1, true},
				{D8, true},
			},
		},
		"Black capture": {
			fenPos: "3r4/8/8/8/8/8/8/3R4 b - - 0 1",
			move:   NewMove(Piece(Rook|Black), D8, D1, NormalMove),
			occupiedChecks: []MaskCheckArgs{
				{D1, false},
				{D8, true},
			},
			whiteOccupiedChecks: []MaskCheckArgs{
				{D1, true},
				{D8, true},
			},
			blackOccupiedChecks: []MaskCheckArgs{
				{D1, false},
				{D8, true},
			},
		},
		"White en-passant capture": {
			fenPos: "8/8/8/1Pp5/8/8/8/8 w KQkq c6 0 1",
			move:   NewMove(Piece(Pawn|White), B5, C6, EnPassant),
			occupiedChecks: []MaskCheckArgs{
				{B5, true},
				{C5, true},
				{C6, false},
			},
			whiteOccupiedChecks: []MaskCheckArgs{
				{B5, true},
				{C5, true},
				{C6, false},
			},
			blackOccupiedChecks: []MaskCheckArgs{
				{B5, true},
				{C5, true},
				{C6, true},
			},
		},
		"Black en-passant capture": {
			fenPos: "8/8/8/8/1Pp5/8/8/8 b KQkq b3 0 1",
			move:   NewMove(Piece(Pawn|Black), C4, B3, EnPassant),
			occupiedChecks: []MaskCheckArgs{
				{B4, true},
				{C4, true},
				{B3, false},
			},
			whiteOccupiedChecks: []MaskCheckArgs{
				{B4, true},
				{C4, true},
				{B3, true},
			},
			blackOccupiedChecks: []MaskCheckArgs{
				{B4, true},
				{C4, true},
				{B3, false},
			},
		},
	}

	updater := NewPositionUpdater(NewBitsBoardMoveGenerator())
	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			updater.MakeMove(pos, d.move)

			for _, check := range d.occupiedChecks {
				assert.Equal(t, !check.isEmpty, pos.occupied&(1<<check.idx) != 0)
			}

			for _, check := range d.whiteOccupiedChecks {
				assert.Equal(t, !check.isEmpty, pos.whiteOccupied&(1<<check.idx) != 0)
			}

			for _, check := range d.blackOccupiedChecks {
				assert.Equal(t, !check.isEmpty, pos.blackOccupied&(1<<check.idx) != 0)
			}
		})
	}
}

func Test_PositionAfterMoveCastleRights(t *testing.T) {
	data := map[string]struct {
		fenPos         string
		move           Move
		expectedRights int8
	}{
		"White king move -> no rights": {
			fenPos:         "rnbqkbnr/pppppppp/8/8/8/4P3/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			move:           NewMove(Piece(King|White), E1, E2, NormalMove),
			expectedRights: NoCastle,
		},
		"Black king move -> no rights": {
			fenPos:         "rnbqkbnr/ppppp1pp/5p2/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			move:           NewMove(Piece(King|Black), E8, F7, NormalMove),
			expectedRights: NoCastle,
		},
		"White QueenSide Rook Move -> KingSide rights": {
			fenPos:         "rnbqkbnr/pppppppp/8/8/8/P6P/1PPPPPP1/RNBQKBNR w KQkq - 0 1",
			move:           NewMove(Piece(Rook|White), A1, A2, NormalMove),
			expectedRights: KingSideCastle,
		},
		"White King Rook Move -> Queen rights": {
			fenPos:         "rnbqkbnr/pppppppp/8/8/8/P6P/1PPPPPP1/RNBQKBNR w KQkq - 0 1",
			move:           NewMove(Piece(Rook|White), H1, H2, NormalMove),
			expectedRights: QueenSideCastle,
		},
		"White QueenSide Rook Move -> No rights": {
			fenPos:         "rnbqkbnr/pppppppp/8/8/8/P6P/1PPPPPP1/RNBQKBNR w Qkq - 0 1",
			move:           NewMove(Piece(Rook|White), A1, A2, NormalMove),
			expectedRights: NoCastle,
		},
		"Black Rook Back to start pos -> No rights": {
			fenPos:         "rnbqkbnr/1pppppp1/p6p/8/8/P6P/1PPPPPP1/RNBQKBNR b - - 0 1",
			move:           NewMove(Piece(Rook|Black), A7, A8, NormalMove),
			expectedRights: NoCastle,
		},
		"Black King Back to start pos -> No rights": {
			fenPos:         "rnbq1bnr/1pppkpp1/p3p2p/8/8/P6P/1PPPPPP1/RNBQKBNR b - - 0 1",
			move:           NewMove(Piece(King|Black), E7, A8, NormalMove),
			expectedRights: NoCastle,
		},
	}

	updater := NewPositionUpdater(NewBitsBoardMoveGenerator())

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fenPos)
			if err != nil {
				t.Error(err)
			}

			initialPosColor := pos.activeColor

			updater.MakeMove(pos, d.move)

			var initialColorCastleRights int8
			if initialPosColor == White {
				initialColorCastleRights = pos.whiteCastleRights
			} else {
				initialColorCastleRights = pos.blackCastleRights
			}
			assert.Equal(t, d.expectedRights, initialColorCastleRights)
		})
	}
}

func Test_PositionAfterEnPassantUpdate(t *testing.T) {
	var data = map[string]struct {
		fenPos                    string
		move                      Move
		enPassantIdxExpectedValue int8
	}{
		"white double pawn move": {
			fenPos:                    FenStartPos,
			move:                      NewMove(Piece(Pawn|White), F2, F4, NormalMove),
			enPassantIdxExpectedValue: F3,
		},
		"black double pawn move": {
			fenPos:                    "rnbqkbnr/pppppppp/8/8/8/P7/1PPPPPPP/RNBQKBNR b KQkq - 0 1",
			move:                      NewMove(Piece(Pawn|Black), F7, F5, NormalMove),
			enPassantIdxExpectedValue: F6,
		},
		"reset after pawn move": {
			fenPos:                    "rnbqkbnr/pppppp1p/8/6p1/8/P7/1PPPPPPP/RNBQKBNR w KQkq g6 0 1",
			move:                      NewMove(Piece(Pawn|White), B2, B3, NormalMove),
			enPassantIdxExpectedValue: NoEnPassant,
		},
		"reset after non-pawn move": {
			fenPos:                    "rnbqkbnr/pppppp1p/8/6p1/8/P7/1PPPPPPP/RNBQKBNR w KQkq g6 0 1",
			move:                      NewMove(Piece(Rook|White), A1, A2, NormalMove),
			enPassantIdxExpectedValue: NoEnPassant,
		},
		"reset after en-passant capture": {
			fenPos:                    "rnbqkbnr/pppppp1p/8/8/P4Pp1/8/1PPPP1PP/RNBQKBNR b KQkq f3 0 1",
			move:                      NewMove(Piece(Pawn|Black), G4, G3, EnPassant),
			enPassantIdxExpectedValue: NoEnPassant,
		},
	}

	updater := NewPositionUpdater(NewBitsBoardMoveGenerator())
	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			updater.MakeMove(pos, d.move)
			assert.Equal(t, d.enPassantIdxExpectedValue, pos.enPassantIdx)
		})
	}
}

func Test_PositionAfterMoveActiveColorUpdate(t *testing.T) {

	updater := NewPositionUpdater(NewBitsBoardMoveGenerator())

	pos, _ := NewPositionFromFEN(FenStartPos)
	updater.MakeMove(pos, NewMove(Piece(Pawn|White), F2, F4, NormalMove))
	assert.Equal(t, Black, pos.activeColor)
	updater.MakeMove(pos, NewMove(Piece(Pawn|Black), G7, G6, NormalMove))
	assert.Equal(t, White, pos.activeColor)
}

func TestPositionUpdater_UnMakeMove(t *testing.T) {

	engine := NewEngine()
	updater := NewPositionUpdater(NewBitsBoardMoveGenerator())
	t.Run("Simple pawn move", func(t *testing.T) {
		pos, _ := NewPositionFromFEN(FenStartPos)
		move := NewMove(Piece(White|Pawn), A2, A4, NormalMove)
		history := updater.MakeMove(pos, move)
		assert.Equal(t, Black, pos.activeColor)
		assert.Equal(t, NoPiece, pos.PieceAt(A2))
		assert.Equal(t, Piece(White|Pawn), pos.PieceAt(A4))

		updater.UnMakeMove(pos, move, history)
		assert.Equal(t, White, pos.activeColor)
		assert.Equal(t, NoPiece, pos.PieceAt(A4))
		assert.Equal(t, Piece(White|Pawn), pos.PieceAt(A2))
	})

	t.Run("Test Position 4", func(t *testing.T) {
		pos, _ := NewPositionFromFEN("r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1")
		legalMoves := engine.LegalMoves(pos)
		assert.Len(t, legalMoves, 6)

		var (
			whiteCastleRights = pos.whiteCastleRights
			blackCastleRights = pos.blackCastleRights
			occupied          = pos.occupied
			blackOccupied     = pos.blackOccupied
			whiteOccupied     = pos.whiteOccupied
			board             = pos.board
			enPassantIdx      = pos.enPassantIdx
			activeColor       = pos.activeColor
		)

		for _, move := range legalMoves {
			history := engine.positionUpdater.MakeMove(pos, move)
			engine.positionUpdater.UnMakeMove(pos, move, history)
		}

		assert.Equal(t, occupied, pos.occupied)
		assert.Equal(t, whiteOccupied, pos.whiteOccupied)
		assert.Equal(t, blackOccupied, pos.blackOccupied)
		assert.Equal(t, whiteCastleRights, pos.whiteCastleRights)
		assert.Equal(t, blackCastleRights, pos.blackCastleRights)
		assert.ElementsMatch(t, board, pos.board)
		assert.Equal(t, enPassantIdx, pos.enPassantIdx)
		assert.Equal(t, activeColor, pos.activeColor)
	})
}
