package board

import (
	"errors"
	"fmt"
	"strconv"
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
	activeColor         int8
	whiteCastleRights   int8
	blackCastleRights   int8
	enPassantIdx        int8
	blackKingIdx        int8
	whiteKingIdx        int8
	occupied            uint64
	blackOccupied       uint64
	whiteOccupied       uint64
	whiteKingSafety     int8
	blackKingSafety     int8
	whiteKingAffectMask uint64
	blackKingAffectMask uint64
	zobristKey          uint64

	queenBoard  uint64
	kingBoard   uint64
	bishopBoard uint64
	rookBoard   uint64
	knightBoard uint64
	pawnBoard   uint64
	board       [64]Piece

	// is init
	isInit bool
	// @todo
	//movesCache   [64][]int8
}

func NewPosition() *Position {
	return &Position{
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

	// Initialize king caches after the board is fully built.
	pos.whiteKingSafety = NotCalculated
	pos.blackKingSafety = NotCalculated
	if hasWhiteKing {
		pos.whiteKingAffectMask = kingAffectMask(pos.whiteKingIdx)
	}

	if hasBlackKing {
		pos.blackKingAffectMask = kingAffectMask(pos.blackKingIdx)
	}

	// Half move clock

	// full move number

	pos.zobristKey = computeZobristKey(pos)
	pos.isInit = true

	return pos, nil
}

// @TODO
// update cache
// update attack vectors
// setPieceAt Reset all the bitboards then update only the relevant ones
func (p *Position) removePieceAt(idx int8, piece Piece) {
	if piece == NoPiece {
		return
	}

	pieceMask := uint64(1 << idx)
	p.occupied &^= pieceMask

	if piece.Color() == White {
		p.whiteOccupied &^= pieceMask
	} else {
		p.blackOccupied &^= pieceMask
	}

	switch piece.Type() {
	case King:
		p.kingBoard &^= pieceMask
	case Queen:
		p.queenBoard &^= pieceMask
	case Rook:
		p.rookBoard &^= pieceMask
	case Bishop:
		p.bishopBoard &^= pieceMask
	case Knight:
		p.knightBoard &^= pieceMask
	case Pawn:
		p.pawnBoard &^= pieceMask
	}

	p.board[idx] = NoPiece
}

func (p *Position) addPieceAt(idx int8, piece Piece) {
	if piece == NoPiece {
		return
	}

	pieceMask := uint64(1 << idx)
	p.occupied |= pieceMask

	if piece.Color() == White {
		p.whiteOccupied |= pieceMask
	} else {
		p.blackOccupied |= pieceMask
	}

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

	p.board[idx] = piece
}

func (p *Position) movePiece(piece Piece, fromIdx, toIdx int8) {
	if piece == NoPiece || fromIdx == toIdx {
		return
	}

	fromMask := uint64(1 << fromIdx)
	toMask := uint64(1 << toIdx)
	moveMask := fromMask | toMask

	p.occupied ^= moveMask

	if piece.Color() == White {
		p.whiteOccupied ^= moveMask
	} else {
		p.blackOccupied ^= moveMask
	}

	switch piece.Type() {
	case King:
		p.kingBoard ^= moveMask
	case Queen:
		p.queenBoard ^= moveMask
	case Rook:
		p.rookBoard ^= moveMask
	case Bishop:
		p.bishopBoard ^= moveMask
	case Knight:
		p.knightBoard ^= moveMask
	case Pawn:
		p.pawnBoard ^= moveMask
	}

	p.board[fromIdx] = NoPiece
	p.board[toIdx] = piece
}

func (p *Position) capturePiece(movingPiece, capturedPiece Piece, fromIdx, toIdx int8) {
	if movingPiece == NoPiece {
		return
	}

	fromMask := uint64(1 << fromIdx)
	toMask := uint64(1 << toIdx)

	p.occupied &^= fromMask
	p.occupied |= toMask

	if movingPiece.Color() == White {
		p.whiteOccupied &^= fromMask
		p.whiteOccupied |= toMask
		p.blackOccupied &^= toMask
	} else {
		p.blackOccupied &^= fromMask
		p.blackOccupied |= toMask
		p.whiteOccupied &^= toMask
	}

	switch capturedPiece.Type() {
	case King:
		p.kingBoard &^= toMask
	case Queen:
		p.queenBoard &^= toMask
	case Rook:
		p.rookBoard &^= toMask
	case Bishop:
		p.bishopBoard &^= toMask
	case Knight:
		p.knightBoard &^= toMask
	case Pawn:
		p.pawnBoard &^= toMask
	}

	switch movingPiece.Type() {
	case King:
		p.kingBoard &^= fromMask
		p.kingBoard |= toMask
	case Queen:
		p.queenBoard &^= fromMask
		p.queenBoard |= toMask
	case Rook:
		p.rookBoard &^= fromMask
		p.rookBoard |= toMask
	case Bishop:
		p.bishopBoard &^= fromMask
		p.bishopBoard |= toMask
	case Knight:
		p.knightBoard &^= fromMask
		p.knightBoard |= toMask
	case Pawn:
		p.pawnBoard &^= fromMask
		p.pawnBoard |= toMask
	}

	p.board[fromIdx] = NoPiece
	p.board[toIdx] = movingPiece
}

func (p *Position) setPieceAt(idx int8, piece Piece) {
	prevPiece := p.board[idx]
	if prevPiece == piece {
		return
	}

	p.removePieceAt(idx, prevPiece)
	p.addPieceAt(idx, piece)
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

func (p *Position) ActiveColor() int8 {
	return p.activeColor
}

func (p *Position) EnPassantIdx() int8 {
	return p.enPassantIdx
}

func (p *Position) WhiteKingIdx() int8 {
	return p.whiteKingIdx
}

func (p *Position) BlackKingIdx() int8 {
	return p.blackKingIdx
}

func (p *Position) Occupied() uint64 {
	return p.occupied
}

func (p *Position) WhiteOccupied() uint64 {
	return p.whiteOccupied
}

func (p *Position) BlackOccupied() uint64 {
	return p.blackOccupied
}

func (p *Position) PawnBoard() uint64 {
	return p.pawnBoard
}

func (p *Position) KnightBoard() uint64 {
	return p.knightBoard
}

func (p *Position) BishopBoard() uint64 {
	return p.bishopBoard
}

func (p *Position) RookBoard() uint64 {
	return p.rookBoard
}

func (p *Position) QueenBoard() uint64 {
	return p.queenBoard
}

func (p *Position) KingBoard() uint64 {
	return p.kingBoard
}

func (p *Position) WhiteCastleRights() int8 {
	return p.whiteCastleRights
}

func (p *Position) BlackCastleRights() int8 {
	return p.blackCastleRights
}

func (p *Position) ZobristKey() uint64 {
	return p.zobristKey
}

func (p *Position) KingSafety(color int8) int8 {
	if color == White {
		return p.whiteKingSafety
	}

	return p.blackKingSafety
}

func (p *Position) SetKingSafety(color, safety int8) {
	if color == White {
		p.whiteKingSafety = safety
		return
	}

	p.blackKingSafety = safety
}

func (p *Position) CastleRights() int8 {
	if p.activeColor == White {
		return p.whiteCastleRights
	}

	return p.blackCastleRights
}

func (p *Position) Clone() *Position {
	cloned := *p
	return &cloned
}

func (p *Position) FEN() string {
	var b strings.Builder
	for rank := int8(7); rank >= 0; rank-- {
		empty := 0
		for file := int8(0); file < 8; file++ {
			idx := rank*8 + file
			piece := p.board[idx]
			if piece == NoPiece {
				empty++
				continue
			}
			if empty > 0 {
				b.WriteString(strconv.Itoa(empty))
				empty = 0
			}
			b.WriteByte(pieceToFENChar(piece))
		}
		if empty > 0 {
			b.WriteString(strconv.Itoa(empty))
		}
		if rank > 0 {
			b.WriteByte('/')
		}
	}

	b.WriteByte(' ')
	if p.activeColor == White {
		b.WriteByte('w')
	} else {
		b.WriteByte('b')
	}

	b.WriteByte(' ')
	castle := castleRightsToFEN(p.whiteCastleRights, p.blackCastleRights)
	if castle == "" {
		b.WriteByte('-')
	} else {
		b.WriteString(castle)
	}

	b.WriteByte(' ')
	if p.enPassantIdx == NoEnPassant {
		b.WriteByte('-')
	} else {
		b.WriteString(IdxToSquare(p.enPassantIdx))
	}

	// Halfmove/fullmove are not tracked today. `0 1` is sufficient to
	// reproduce legal move generation and illegal-move diagnostics.
	b.WriteString(" 0 1")
	return b.String()
}

func pieceToFENChar(piece Piece) byte {
	var ch byte
	switch piece.Type() {
	case King:
		ch = 'k'
	case Queen:
		ch = 'q'
	case Rook:
		ch = 'r'
	case Bishop:
		ch = 'b'
	case Knight:
		ch = 'n'
	case Pawn:
		ch = 'p'
	default:
		return '?'
	}

	if piece.Color() == White {
		ch -= 'a' - 'A'
	}
	return ch
}

func castleRightsToFEN(whiteRights, blackRights int8) string {
	var b strings.Builder
	if whiteRights&KingSideCastle != 0 {
		b.WriteByte('K')
	}
	if whiteRights&QueenSideCastle != 0 {
		b.WriteByte('Q')
	}
	if blackRights&KingSideCastle != 0 {
		b.WriteByte('k')
	}
	if blackRights&QueenSideCastle != 0 {
		b.WriteByte('q')
	}
	return b.String()
}
