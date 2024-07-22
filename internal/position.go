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
	occupied          uint64
	blackOccupied     uint64
	whiteOccupied     uint64
	whiteKingSafety   int8
	blackKingSafety   int8

	queenBoard  uint64
	kingBoard   uint64
	bishopBoard uint64
	rookBoard   uint64
	knightBoard uint64
	pawnBoard   uint64

	// is init
	isInit bool
	// @todo
	movesCache   [64][]int8
	whiteAttacks uint64
	blackAttacks uint64
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
		isInit:            false,
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
	hasWhiteKing := false
	hasBlackKing := false

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
				hasWhiteKing = true
				pos.whiteKingIdx = idx
			} else {
				hasBlackKing = true
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
	// Important to init again safeties: possible wrong states due to early calculation with partial board
	pos.whiteKingSafety = NotCalculated
	pos.blackKingSafety = NotCalculated
	if hasWhiteKing {
		if IsKingInCheck(pos, White) {
			pos.whiteKingSafety = KingIsCheck
		} else {
			pos.whiteKingSafety = KingIsSafe
		}
	}

	if hasBlackKing {
		if IsKingInCheck(pos, Black) {
			pos.blackKingSafety = KingIsCheck
		} else {
			pos.blackKingSafety = KingIsSafe
		}
	}

	// Half move clock

	// full move number

	pos.isInit = true

	return pos, nil
}

// @TODO
// update cache
// update attack vectors
// setPieceAt Reset all the bitboards then update only the relevant ones
func (p *Position) setPieceAt(idx int8, piece Piece) {
	prevPiece := p.board[idx]

	p.board[idx] = piece
	pieceMask := uint64(1 << idx)

	if prevPiece != NoPiece {
		p.kingBoard &= ^pieceMask
		p.queenBoard &= ^pieceMask
		p.rookBoard &= ^pieceMask
		p.bishopBoard &= ^pieceMask
		p.knightBoard &= ^pieceMask
		p.pawnBoard &= ^pieceMask

		p.whiteOccupied &= ^pieceMask
		p.blackOccupied &= ^pieceMask
	}

	p.occupied &= ^pieceMask

	if piece != NoPiece {
		if piece.Color() == White {
			p.whiteOccupied |= pieceMask
		} else {
			p.blackOccupied |= pieceMask
		}
		p.occupied |= uint64(1) << idx

		switch piece.Type() {
		case King:
			p.kingBoard |= pieceMask
		case Queen:
			p.queenBoard |= pieceMask
		case Rook:
			p.rookBoard |= pieceMask
		case Bishop:
			p.bishopBoard |= pieceMask
		case Knight:
			p.knightBoard |= pieceMask
		case Pawn:
			p.pawnBoard |= pieceMask
		}
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

func (p *Position) OpponentOccupiedMaskByPieceColor(pieceColor int8) uint64 {
	if pieceColor == White {
		return p.blackOccupied
	}

	return p.whiteOccupied
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
