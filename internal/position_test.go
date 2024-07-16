package internal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewPositionFromFEN(t *testing.T) {

	type Check struct {
		square int8
		piece  Piece
	}

	data := []struct {
		fenPos string
		checks []Check
	}{
		{
			fenPos: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			checks: []Check{
				{A1, Piece(Rook | White)},
				{A8, Piece(Rook | Black)},
				{A4, NoPiece},
			},
		},
		{
			fenPos: "1k6/8/8/8/3Q4/8/8/K7 w KQkq - 0 1",
			checks: []Check{
				{B8, Piece(King | Black)},
				{A1, Piece(King | White)},
			},
		},
	}

	for _, d := range data {
		pos, err := NewPositionFromFEN(d.fenPos)
		if err != nil {
			t.Error(err)
		}

		for _, check := range d.checks {
			assert.Equal(t, check.piece, pos.PieceAt(check.square))
		}
	}
}

func TestPosition_CanCastle(t *testing.T) {
	data := []struct {
		fenPos             string
		color              int8
		canCastleQueenSide bool
		canCastleKingSide  bool
	}{
		{
			fenPos:             "8/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			color:              White,
			canCastleQueenSide: true,
			canCastleKingSide:  true,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K3 w KQkq - 0 1",
			color:              White,
			canCastleQueenSide: true,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "r3k2r/8/8/8/8/8/8/8 w KQkq - 0 1",
			color:              Black,
			canCastleQueenSide: true,
			canCastleKingSide:  true,
		},
		{
			fenPos:             "r3k3/8/8/8/8/8/8/8 w KQkq - 0 1",
			color:              Black,
			canCastleQueenSide: true,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K3 w - - 0 1",
			color:              White,
			canCastleQueenSide: false,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K3 w - - 0 1",
			color:              Black,
			canCastleQueenSide: false,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K2R w kq - 0 1",
			color:              White,
			canCastleQueenSide: false,
			canCastleKingSide:  false,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/R3K2R w KQ - 0 1",
			color:              White,
			canCastleQueenSide: true,
			canCastleKingSide:  true,
		},
		{
			fenPos:             "8/8/8/8/8/8/8/r3K2r w KQkq - 0 1",
			color:              White,
			canCastleQueenSide: false,
			canCastleKingSide:  false,
		},
	}

	for _, d := range data {
		pos, err := NewPositionFromFEN(d.fenPos)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, d.canCastleQueenSide, pos.CanCastle(d.color, QueenSideCastle))
		assert.Equal(t, d.canCastleKingSide, pos.CanCastle(d.color, KingSideCastle))
	}
}

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

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			initPos, err := NewPositionFromFEN(d.fenPos)
			if err != nil {
				t.Error(err)
			}

			newPos := initPos.PositionAfterMove(d.move)
			for _, check := range d.checks {
				assert.Equal(t, check.piece, newPos.PieceAt(check.idx), fmt.Sprintf("Wrong piece at idx : %d", check.idx))
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

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			initPos, _ := NewPositionFromFEN(d.fenPos)
			newPos := initPos.PositionAfterMove(d.move)

			for _, check := range d.occupiedChecks {
				assert.Equal(t, !check.isEmpty, newPos.occupied&(1<<check.idx) != 0)
			}

			for _, check := range d.whiteOccupiedChecks {
				assert.Equal(t, !check.isEmpty, newPos.whiteOccupied&(1<<check.idx) != 0)
			}

			for _, check := range d.blackOccupiedChecks {
				assert.Equal(t, !check.isEmpty, newPos.blackOccupied&(1<<check.idx) != 0)
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

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			initPos, err := NewPositionFromFEN(d.fenPos)
			if err != nil {
				t.Error(err)
			}

			newPos := initPos.PositionAfterMove(d.move)

			var actualRights int8
			if initPos.activeColor == White {
				actualRights = newPos.whiteCastleRights
			} else {
				actualRights = newPos.blackCastleRights
			}
			assert.Equal(t, d.expectedRights, actualRights)
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

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			pos = pos.PositionAfterMove(d.move)
			assert.Equal(t, d.enPassantIdxExpectedValue, pos.enPassantIdx)
		})
	}
}

func Test_PositionAfterMoveActiveColorUpdate(t *testing.T) {
	pos, _ := NewPositionFromFEN(FenStartPos)
	newPos := pos.PositionAfterMove(NewMove(Piece(Pawn|White), F2, F4, NormalMove))
	assert.Equal(t, Black, newPos.activeColor)
	newPos2 := newPos.PositionAfterMove(NewMove(Piece(Pawn|Black), G7, G6, NormalMove))
	assert.Equal(t, White, newPos2.activeColor)
}

func TestPosition_IsCheck(t *testing.T) {
	data := map[string]struct {
		fenPos  string
		isCheck bool
	}{
		"startPos white no check": {
			fenPos:  FenStartPos,
			isCheck: false,
		},
		"Check By Queen": {
			fenPos:  "rnb1kbnr/ppppqppp/8/5p2/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			isCheck: true,
		},
		"Black NOT inCheck (white is)": {
			fenPos:  "rnb1kbnr/ppppqppp/8/5p2/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1",
			isCheck: false,
		},
		"Check by Knight": {
			fenPos:  "rnb1kb1r/ppppqppp/8/3n1p2/5K2/8/PPPP1PPP/RNBQ1BNR w KQkq - 0 1",
			isCheck: true,
		},
		"Check by pawn": {
			fenPos:  "rnb1kb1r/ppppqppp/8/3n1p2/4K3/8/PPPP1PPP/RNBQ1BNR w KQkq - 0 1",
			isCheck: true,
		},
		"NOT Check by pawns": {
			fenPos:  "8/8/8/4p3/3pK3/8/8/8 w - - 0 1",
			isCheck: false,
		},
		"NOT Check by bishop same file": {
			fenPos:  "4b3/8/8/8/4K3/8/8/8 w - - 0 1",
			isCheck: false,
		},
		"NOT Check by bishop same rank": {
			fenPos:  "8/8/8/8/b3K3/8/8/8 w - - 0 1",
			isCheck: false,
		},
		"NOT Check by black pawn, pawnRank < king rank": {
			fenPos:  "8/8/8/8/4K3/3p4/8/8 w - - 0 1",
			isCheck: false,
		},
		"NOT Check by white pawn, pawnRank > king rank": {
			fenPos:  "8/8/8/3P4/4k3/8/8/8 b - - 0 1",
			isCheck: false,
		},
		"NOT Check by pawn same rank": {
			fenPos:  "rnb1k2r/ppppbppp/8/3np2q/4K3/8/PPPP1PPP/RNBQ1BNR w KQkq - 0 1",
			isCheck: false,
		},
		"Check By Rook same file": {
			fenPos:  "4r3/8/8/8/4K3/8/8/8 w - - 0 1",
			isCheck: true,
		},
		"Check By Rook same rank": {
			fenPos:  "8/8/8/8/R3k3/8/8/8 b - - 0 1",
			isCheck: true,
		},
		"NOT Check By Rook in diagonal": {
			fenPos:  "R7/8/8/8/4k3/8/8/8 b - - 0 1",
			isCheck: false,
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, _ := NewPositionFromFEN(d.fenPos)
			assert.Equal(t, d.isCheck, pos.IsCheck())
		})
	}
}
