package internal

import (
	"errors"
	"fmt"
	"strings"
)

const (
	NoCastle        = 0
	KingSideCastle  = 1
	QueenSideCastle = 2
	NoEnPassant     = -1
)

type Position struct {
	board             map[int]Piece
	activeColor       int8
	whiteCastleRights int8
	blackCastleRights int8
	enPassantSquares  int
	Yolo              int
}

func NewPosition() Position {
	board := make(map[int]Piece, 64)
	for i := 0; i < 64; i++ {
		board[i] = NoPiece
	}

	return Position{
		board:             board,
		activeColor:       White,
		whiteCastleRights: KingSideCastle | QueenSideCastle,
		blackCastleRights: KingSideCastle | QueenSideCastle,
		enPassantSquares:  NoEnPassant,
	}
}

func NewPositionFromFEN(fen string) (Position, error) {
	var FENCharToPiece = map[rune]Piece{
		'K': Piece(King | White),
		'Q': Piece(Queen | White),
		'R': Piece(Rook | White),
		'B': Piece(Bishop | White),
		'N': Piece(Knight | White),
		'P': Piece(Pawn | White),
		'k': Piece(King | Black),
		'q': Piece(Queen | Black),
		'r': Piece(Rook | Black),
		'b': Piece(Bishop | Black),
		'n': Piece(Knight | Black),
		'p': Piece(Pawn | Black),
	}

	pos := NewPosition()

	parts := strings.Split(fen, " ")
	if len(parts) < 4 {
		return pos, errors.New("invalid FEN")
	}

	// Board init
	var (
		rank = 7
		file = 0
	)

	for i, char := range parts[0] {
		if char == '/' {
			rank--
			file = 0
			continue
		}

		idx := rank*8 + file

		if char >= '1' && char <= '8' {
			file += int(char - '0')
			continue
		}

		piece, ok := FENCharToPiece[char]
		if !ok {
			return pos, errors.New(fmt.Sprintf("invalid FEN: invalid char '%c' at position %d", char, i))
		}
		pos.board[idx] = piece
		file++
	}

	// Next player turn
	switch parts[1][0] {
	case 'w':
		pos.activeColor = White
	case 'b':
		pos.activeColor = Black
	default:
		return pos, errors.New(fmt.Sprintf("invalid FEN: invalid color %s", string(parts[1][0])))
	}

	// Castle rights
	pos.blackCastleRights = NoCastle
	pos.whiteCastleRights = NoCastle
	if parts[2] != "-" {
		for _, char := range parts[2] {
			switch char {
			case 'k':
				pos.blackCastleRights |= KingSideCastle
			case 'q':
				pos.blackCastleRights |= QueenSideCastle
			case 'K':
				pos.whiteCastleRights |= KingSideCastle
			case 'Q':
				pos.whiteCastleRights |= QueenSideCastle
			default:
				return pos, errors.New(fmt.Sprintf("invalid FEN: castle rights %s", string(char)))
			}
		}
	}

	// En passant
	pos.enPassantSquares = NoEnPassant
	if parts[3] != "-" {
		pos.enPassantSquares = SquareToIdx(parts[3])
	}

	// Half move clock

	// full move number

	return pos, nil
}

func (p Position) PieceAt(idx int) Piece {
	return p.board[idx]
}

func (p Position) CanCastle(clr int8, castleRight int8) bool {
	var (
		kingPos      int
		rookPos      int
		emptyIdx     []int
		castleRights int8
	)

	if clr == White {
		castleRights = p.whiteCastleRights
		kingPos = 4
		switch castleRight {
		case QueenSideCastle:
			rookPos = 0
			emptyIdx = []int{1, 2}
		case KingSideCastle:
			rookPos = 7
			emptyIdx = []int{5, 6}
		}
	} else {
		castleRights = p.blackCastleRights
		kingPos = 60
		switch castleRight {
		case QueenSideCastle:
			rookPos = 56
			emptyIdx = []int{57, 58}
		case KingSideCastle:
			rookPos = 63
			emptyIdx = []int{61, 62}
		}
	}

	if (castleRights & castleRight) == 0 {
		return false
	}

	if (p.PieceAt(kingPos) != Piece(King|clr)) || (p.PieceAt(rookPos) != Piece(Rook|clr)) {
		return false
	}

	for _, idx := range emptyIdx {
		if p.PieceAt(idx) != NoPiece {
			return false
		}
	}

	return true
	// black
}
