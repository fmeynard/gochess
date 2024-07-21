package internal

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

func SquareToIdx(square string) int8 {
	if len(square) != 2 {
		panic(fmt.Sprintf("invalid square identifier: %s", square))
	}

	file := square[0] - 'a'
	rank := square[1] - '1'

	if file < 0 || file > 7 {
		panic(fmt.Sprintf("invalid file identifier: %s", string(square[0])))
	}

	if rank < 0 || rank > 7 {
		panic(fmt.Sprintf("invalid rank identifier: %s", string(square[1])))
	}

	return int8(rank*8 + file)
}

func IdxToSquare(idx int8) string {
	if idx < 0 || idx > 63 {
		panic("idx out of range")
	}

	file := idx % 8
	rank := idx / 8

	return fmt.Sprintf("%c%d", 'a'+file, rank+1)
}

func RankAndFile(idx int8) (int8, int8) {
	return RankFromIdx(idx), FileFromIdx(idx)
}

func RankFromIdx(idx int8) int8 {
	return idx >> 3
}

func FileFromIdx(idx int8) int8 {
	return idx & 7
}

func absInt8(x int8) int8 {
	if x < 0 {
		return -x
	}
	return x
}

func leastSignificantOne(bb uint64) int8 {
	return int8(bits.TrailingZeros64(bb))
}

// mostSignificantBit returns the position of the highest set bit (most significant bit)
func mostSignificantBit(x uint64) int8 {
	return int8(63 - bits.LeadingZeros64(x))
}

func movesToUci(moves []Move) []string {
	var uciMoves []string
	for _, move := range moves {
		uciMoves = append(uciMoves, move.UCI())
	}

	return uciMoves
}

func draw(vector uint64) {
	for rank := 7; rank >= 0; rank-- {
		var currentLine []string
		for file := 0; file < 8; file++ {
			mask := uint64(1) << (rank*8 + file)
			if vector&mask != 0 {
				currentLine = append(currentLine, "1")
			} else {
				currentLine = append(currentLine, "0")
			}
		}

		fmt.Println("|", strings.Join(currentLine, " | "), "|")
	}
}

// isOnBoard checks if the given file and rank are within the bounds of the board.
func isOnBoard(file, rank int8) bool {
	return file >= 0 && file < 8 && rank >= 0 && rank < 8
}

func positionToFEN(pos *Position) string {
	var fen string
	emptyCount := 0

	// Generate the piece placement data
	for row := 7; row >= 0; row-- { // FEN starts from the 8th rank down to the 1st
		for col := 0; col < 8; col++ {
			index := row*8 + col
			piece := pos.board[index]

			if piece == NoPiece {
				emptyCount++
			} else {
				if emptyCount > 0 {
					fen += strconv.Itoa(emptyCount)
					emptyCount = 0
				}
				fen += pieceToFENChar(piece)
			}
		}
		if emptyCount > 0 {
			fen += strconv.Itoa(emptyCount)
			emptyCount = 0
		}
		if row > 0 {
			fen += "/"
		}
	}

	// Active color
	if pos.activeColor == White {
		fen += " w "
	} else {
		fen += " b "
	}

	// Castling availability
	castle := ""
	if pos.whiteCastleRights&KingSideCastle != 0 {
		castle += "K"
	}
	if pos.whiteCastleRights&QueenSideCastle != 0 {
		castle += "Q"
	}
	if pos.blackCastleRights&KingSideCastle != 0 {
		castle += "k"
	}
	if pos.blackCastleRights&QueenSideCastle != 0 {
		castle += "q"
	}
	if castle == "" {
		castle = "-"
	}
	fen += castle + " "

	// En passant target square
	if pos.enPassantIdx != NoEnPassant {
		fen += indexToFENPosition(pos.enPassantIdx) + " "
	} else {
		fen += "- "
	}

	// Half-move clock
	fen += "0 "

	// Full-move number
	fen += "1"

	return fen
}

func pieceToFENChar(piece Piece) string {
	if piece == NoPiece {
		return ""
	}

	var pieceStr string
	switch piece.Type() {
	case Pawn:
		pieceStr = "p"
	case Knight:
		pieceStr = "n"
	case Bishop:
		pieceStr = "b"
	case Rook:
		pieceStr = "r"
	case Queen:
		pieceStr = "q"
	case King:
		pieceStr = "k"
	default:
		pieceStr = ""
	}

	if piece.Color() == White {
		return strings.ToUpper(pieceStr)
	}

	return pieceStr
}

func indexToFENPosition(index int8) string {
	// Convert board index into FEN position notation, e.g., e2, h7, etc.
	file := index % 8
	rank := index / 8
	return strconv.Itoa(int('a'+file)) + strconv.Itoa(int(rank+1))
}

func isSameLineOrRow(start, end, direction int8) bool {
	switch direction {
	case 1, -1: // Horizontal
		return start/8 == end/8
	case 8, -8: // Vertical
		return start%8 == end%8
	default: // Diagonals
		return absInt8(start%8-end%8) == absInt8(start/8-end/8)
	}
}
