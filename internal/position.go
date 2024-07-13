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
)

const FenStartPos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

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
}

func NewPosition() Position {
	board := [64]Piece{}

	for i := int8(0); i < 64; i++ {
		board[i] = NoPiece
	}

	return Position{
		board:             board,
		activeColor:       White,
		whiteCastleRights: KingSideCastle | QueenSideCastle,
		blackCastleRights: KingSideCastle | QueenSideCastle,
		enPassantIdx:      NoEnPassant,
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

		idx := int8(rank*8 + file)

		if char >= '1' && char <= '8' {
			file += int(char - '0')
			continue
		}

		piece, ok := FENCharToPiece[char]
		if !ok {
			return pos, errors.New(fmt.Sprintf("invalid FEN: invalid char '%c' at position %d", char, i))
		}
		pos.board[idx] = piece
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
	pos.enPassantIdx = NoEnPassant
	if parts[3] != "-" {
		pos.enPassantIdx = SquareToIdx(parts[3])
	}

	// Half move clock

	// full move number

	updateAttackVectors(&pos)

	return pos, nil
}

func (p Position) PieceAt(idx int8) Piece {
	return p.board[idx]
}

func (p Position) CanCastle(clr int8, castleRight int8) bool {
	var (
		kingPos      int8
		rookPos      int8
		emptyIdx     []int8
		castleRights int8
	)

	if clr == White {
		castleRights = p.whiteCastleRights
		kingPos = 4
		switch castleRight {
		case QueenSideCastle:
			rookPos = 0
			emptyIdx = []int8{1, 2}
		case KingSideCastle:
			rookPos = 7
			emptyIdx = []int8{5, 6}
		}
	} else {
		castleRights = p.blackCastleRights
		kingPos = 60
		switch castleRight {
		case QueenSideCastle:
			rookPos = 56
			emptyIdx = []int8{57, 58}
		case KingSideCastle:
			rookPos = 63
			emptyIdx = []int8{61, 62}
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
}

func updatePieceOnBoard(p *Position, piece Piece, oldIdx int8, newIdx int8) {
	p.board[oldIdx] = NoPiece
	p.board[newIdx] = piece

}

func (p Position) PositionAfterMove(move Move) Position {
	newPos := p

	startPieceIdx := move.StartIdx()
	endPieceIdx := move.EndIdx()
	startPiece := p.PieceAt(move.StartIdx())
	startPieceType := startPiece.Type()

	// update position
	newPos.board[startPieceIdx] = NoPiece
	newPos.board[endPieceIdx] = startPiece
	//updatePieceOnBoard(&newPos, startPiece, startPieceIdx, endPieceIdx)

	// King move -> update king pos and castleRights
	if startPieceType == King {
		if p.activeColor == White {
			newPos.whiteKingIdx = endPieceIdx
			newPos.whiteCastleRights = NoCastle

			if startPieceIdx == E1 {
				if endPieceIdx == G1 {
					newPos.board[H1] = NoPiece
					newPos.board[F1] = Piece(White | Rook)
				} else if endPieceIdx == C1 {
					newPos.board[A1] = NoPiece
					newPos.board[D1] = Piece(White | Rook)
				}
			}

		} else {
			newPos.blackKingIdx = endPieceIdx
			newPos.blackCastleRights = NoCastle

			if startPieceIdx == E8 {
				if endPieceIdx == G8 {
					newPos.board[H8] = NoPiece
					newPos.board[F8] = Piece(Black | Rook)
				} else if endPieceIdx == C8 {
					newPos.board[A8] = NoPiece
					newPos.board[D8] = Piece(Black | Rook)
				}
			}
		}
	}

	// en passant
	if startPieceType == Pawn {
		if endPieceIdx == p.enPassantIdx {
			var capturesPawnIdx int8
			if startPiece.Color() == White {
				capturesPawnIdx = p.enPassantIdx - 8
			} else {
				capturesPawnIdx = p.enPassantIdx + 8
			}
			newPos.board[capturesPawnIdx] = NoPiece
		}

		diffIdx := endPieceIdx - startPieceIdx
		if p.activeColor == White && diffIdx == 16 {
			newPos.enPassantIdx = startPieceIdx + 8
		} else if p.activeColor == Black && diffIdx == -16 {
			newPos.enPassantIdx = startPieceIdx - 8
		} else {
			newPos.enPassantIdx = NoEnPassant
		}
	}

	// knight
	if startPieceType == Rook {
		if p.activeColor == White {
			if startPieceIdx == A1 {
				newPos.whiteCastleRights = newPos.whiteCastleRights &^ QueenSideCastle
			} else if startPieceIdx == H1 {
				newPos.whiteCastleRights = newPos.whiteCastleRights &^ KingSideCastle
			}
		} else {
			if startPieceIdx == A8 {
				newPos.blackCastleRights = newPos.blackCastleRights &^ QueenSideCastle
			} else if startPieceIdx == H8 {
				newPos.blackCastleRights = newPos.blackCastleRights &^ KingSideCastle
			}
		}
	}

	// change side ( important to do it last for previous updates )
	if p.activeColor == White {
		newPos.activeColor = Black
	} else {
		newPos.activeColor = White
	}

	updateAttackVectors(&newPos)

	return newPos
}
