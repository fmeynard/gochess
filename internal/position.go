package internal

import (
	"errors"
	"fmt"
	"strings"
)

const (
	NoCastle             = 0
	KingSideCastle       = 1
	QueenSideCastle      = 2
	NoEnPassant     int8 = -1

	//king safety
	NotCalculated int8 = 0
	KingIsSafe    int8 = 8
	KingIsCheck   int8 = 16
)

const FenStartPos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
const EmptyBoard = "8/8/8/8/8/8/8/8 w - - 0 1"

type Position struct {
	board             [64]Piece
	activeColor       int8
	whiteCastleRights int8
	blackCastleRights int8
	enPassantIdx      int8
	blackKingIdx      int8
	whiteKingIdx      int8
	whiteAttacks      uint64
	blackAttacks      uint64
	occupied          uint64
	blackOccupied     uint64
	whiteOccupied     uint64
	movesCache        [64][]int8
	whiteKingSafety   int8
	blackKingSafety   int8
}

func NewPosition() *Position {
	board := [64]Piece{}

	for i := int8(0); i < 64; i++ {
		board[i] = NoPiece
	}

	return &Position{
		board:             board,
		activeColor:       White,
		whiteCastleRights: KingSideCastle | QueenSideCastle,
		blackCastleRights: KingSideCastle | QueenSideCastle,
		enPassantIdx:      NoEnPassant,
		occupied:          uint64(0),
		whiteOccupied:     uint64(0),
		blackOccupied:     uint64(0),
		whiteKingSafety:   NotCalculated,
		blackKingSafety:   NotCalculated,
	}
}

func NewPositionFromFEN(fen string) (*Position, error) {
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
		return nil, errors.New("invalid FEN")
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

		idx := int8(rank*8 + file)

		if char >= '1' && char <= '8' {
			file += int(char - '0')
			continue
		}

		piece, ok := FENCharToPiece[char]
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid FEN: invalid char '%c' at position %d", char, i))
		}
		pos.setPieceAt(idx, piece)
		if piece.Type() == King {
			if piece.Color() == White {
				pos.whiteKingIdx = idx
			} else {
				pos.blackKingIdx = idx
			}
		}
		file++
	}

	// Next player turn
	switch parts[1][0] {
	case 'w':
		pos.activeColor = White
	case 'b':
		pos.activeColor = Black
	default:
		return nil, errors.New(fmt.Sprintf("invalid FEN: invalid color %s", string(parts[1][0])))
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
				return nil, errors.New(fmt.Sprintf("invalid FEN: castle rights %s", string(char)))
			}
		}
	}

	// En passant
	pos.enPassantIdx = NoEnPassant
	if parts[3] != "-" {
		pos.enPassantIdx = SquareToIdx(parts[3])
	}

	// Init king safety
	if IsKingInCheck(pos, White) {
		pos.whiteKingSafety = KingIsCheck
	} else {
		pos.whiteKingSafety = KingIsSafe
	}

	if IsKingInCheck(pos, Black) {
		pos.blackKingSafety = KingIsCheck
	} else {
		pos.blackKingSafety = KingIsSafe
	}

	// Half move clock

	// full move number

	return pos, nil
}

// @TODO
// update cache
// update attack vectors
func (p *Position) setPieceAt(idx int8, piece Piece) {
	p.board[idx] = piece

	if piece == NoPiece {
		if p.activeColor == White {
			p.whiteOccupied &= ^(uint64(1) << idx)
		} else {
			p.blackOccupied &= ^(uint64(1) << idx)
		}
		p.occupied &= ^(uint64(1) << idx)
	} else {
		if piece.Color() == White {
			p.blackOccupied &= ^(uint64(1) << idx)
			p.whiteOccupied |= uint64(1) << idx
		} else {
			p.whiteOccupied &= ^(uint64(1) << idx)
			p.blackOccupied |= uint64(1) << idx
		}
		p.occupied |= uint64(1) << idx
	}
}

func (p *Position) OpponentColor() int8 {
	if p.activeColor == White {
		return Black
	}
	return White
}

func (p *Position) PieceAt(idx int8) Piece {
	return p.board[idx]
}

func (p *Position) IsOccupied(idx int8) bool {
	return (p.occupied & (1 << idx)) != 0
}

func (p *Position) IsColorOccupied(color, idx int8) bool {
	if color == White {
		return (p.whiteOccupied & (1 << idx)) != 0
	}

	return (p.blackOccupied & (1 << idx)) != 0
}

func (p *Position) OccupancyMask(color int8) uint64 {
	if color == White {
		return p.whiteOccupied
	}

	return p.blackOccupied
}

func (p *Position) OpponentOccupiedMask() uint64 {
	if p.activeColor == White {
		return p.blackOccupied
	}

	return p.whiteOccupied
}

func (p *Position) CastleRights() int8 {
	if p.activeColor == White {
		return p.whiteCastleRights
	}

	return p.blackCastleRights
}

func (p *Position) IsCheck() bool {
	return IsKingInCheck(p, p.activeColor)
}
