package internal

import (
	"fmt"
	"sort"
)

// direction offsets
const (
	LEFT      int8 = -1
	RIGHT     int8 = 1
	UP        int8 = -8
	DOWN      int8 = 8
	UpLeft    int8 = -9
	UpRight   int8 = -7
	DownLeft  int8 = 7
	DownRight int8 = 9
)

var (
	QueenDirections  = []int8{LEFT, RIGHT, UP, DOWN, UpLeft, UpRight, DownLeft, DownRight}
	RookDirections   = []int8{UP, DOWN, LEFT, RIGHT}
	BishopDirections = []int8{UpLeft, UpRight, DownLeft, DownRight}
)

var kingMoves2 = []int8{
	-9, -8, -7, -1, 1, 7, 8, 9,
}

var knightMoves2 = []int8{
	-17, -15, -10, -6,
	6, 10, 15, 17,
}

var (
	queenDirections = [8]int8{8, -8, 1, -1, 9, -9, 7, -7}
	kingMoves       = [8][2]int8{{0, 1}, {1, 0}, {0, -1}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
)

func SliderPseudoLegalMoves(p Position, pieceIdx int8) ([]int8, []int8) {
	piece := p.PieceAt(pieceIdx)
	if piece == NoPiece {
		return nil, nil
	}

	return generateSliderPseudoLegalMoves(p, pieceIdx, piece)
}

func generateSliderPseudoLegalMoves(p Position, pieceIdx int8, piece Piece) ([]int8, []int8) {
	var (
		moves         = [45]int8{}
		capturesMoves = [45]int8{}
	)

	pieceType := piece.Type()
	pieceColor := piece.Color()

	var directions []int8
	if pieceType == Queen {
		directions = QueenDirections
	} else if pieceType == Bishop {
		directions = BishopDirections
	} else if pieceType == Rook {
		directions = RookDirections
	} else {
		panic("Invalid PieceType")
	}

	x := 0
	y := 0
	pieceFile := FileFromIdx(pieceIdx)
	for _, direction := range directions {

		// startPiece is on the edge -> skip related directions with horizontal moves
		if (pieceFile == 7 && (direction == RIGHT || direction == UpRight || direction == DownRight)) ||
			(pieceFile == 0 && (direction == LEFT || direction == UpLeft || direction == DownLeft)) {
			continue
		}

		for i := int8(1); i < 8; i++ { // start at 1 because 0 is current square
			targetIdx := pieceIdx + direction*i

			// current move+direction is out of the board
			// handle UP and DOWN
			if targetIdx < 0 || targetIdx > 63 {
				break
			}

			// target square is not empty -> stop current direction
			target := p.PieceAt(targetIdx)
			if target != NoPiece {
				if target.Color() == pieceColor {
					break
				}

				capturesMoves[y] = targetIdx
				y++
				moves[x] = targetIdx
				x++
				break
			}

			// add to the list
			moves[x] = targetIdx
			x++

			// target is on the edge -> no more moves in that direction
			targetFile := FileFromIdx(targetIdx)
			if targetFile == 7 && (direction == RIGHT || direction == UpRight || direction == DownRight) {
				break
			}

			if targetFile == 0 && (direction == LEFT || direction == UpLeft || direction == DownLeft) {
				break
			}
		}
	}

	return moves[:x], capturesMoves[:y]
}

func KnightPseudoLegalMoves(p Position, pieceIdx int8) ([]int8, []int8) {
	return generateKnightPseudoLegalMoves(p, pieceIdx, p.activeColor)
}

var knightMoves = [8][2]int8{
	{2, 1}, {1, 2}, {-1, 2}, {-2, 1},
	{-2, -1}, {-1, -2}, {1, -2}, {2, -1},
}

func generateKnightPseudoLegalMoves(p Position, startIdx int8, color int8) ([]int8, []int8) {
	var (
		moves         = make([]int8, 0, 8)
		capturesMoves = make([]int8, 0, 8)
	)

	startRank, startFile := RankAndFile(startIdx)

	for _, move := range knightMoves {
		newFile := startFile + move[0]
		if !(newFile >= 0 && newFile < 8) {
			continue
		}

		newRank := startRank + move[1]
		if !(newRank >= 0 && newRank < 8) {
			continue
		}

		endIdx := newRank*8 + newFile
		targetPiece := p.board[endIdx]

		if targetPiece == NoPiece {
			moves = append(moves, endIdx)
		} else if targetPiece.Color() != color {
			moves = append(moves, endIdx)
			capturesMoves = append(capturesMoves, endIdx)
		}
	}

	return moves, capturesMoves
}

func KingPseudoLegalMoves(p Position, startIdx int8) ([]int8, []int8) {
	var (
		moves         = make([]int8, 0, 8)
		capturesMoves = make([]int8, 0, 8)
	)

	startRank, startFile := RankAndFile(startIdx)

	for _, move := range kingMoves {
		newFile := startFile + move[0]
		if !(newFile >= 0 && newFile < 8) {
			continue
		}

		newRank := startRank + move[1]
		if !(newRank >= 0 && newRank < 8) {
			continue
		}

		endIdx := newRank*8 + newFile
		targetPiece := p.board[endIdx]
		if targetPiece == NoPiece {
			moves = append(moves, endIdx)
		} else if targetPiece.Color() != p.activeColor {
			moves = append(moves, endIdx)
			capturesMoves = append(capturesMoves, endIdx)
		}
	}

	var (
		castleRights int8
		kingStartIdx int8
	)
	if p.activeColor == White {
		castleRights = p.whiteCastleRights
		kingStartIdx = E1
	} else {
		castleRights = p.blackCastleRights
		kingStartIdx = E8
	}

	// early exit no castle
	if startIdx != kingStartIdx || castleRights == NoCastle {
		return moves, capturesMoves
	}

	// queen side
	var (
		queenPathIsClear bool
		queenRookIdx     int8
		queenCastleIdx   int8
	)
	if (castleRights & QueenSideCastle) != 0 {
		if p.activeColor == White {
			queenRookIdx = A1
			queenCastleIdx = C1
			queenPathIsClear = (p.PieceAt(B1) == NoPiece) && (p.PieceAt(C1) == NoPiece) && (p.PieceAt(D1) == NoPiece)
		} else {
			queenRookIdx = A8
			queenCastleIdx = C8
			queenPathIsClear = (p.PieceAt(B8) == NoPiece) && (p.PieceAt(C8) == NoPiece) && (p.PieceAt(D8) == NoPiece)
		}

		if queenPathIsClear && p.PieceAt(queenRookIdx).Type() == Rook {
			moves = append(moves, queenCastleIdx)
		}
	}

	// king side
	var (
		kingPathIsClear bool
		kingRookIdx     int8
		kingCastleIdx   int8
	)
	if (castleRights & KingSideCastle) != 0 {
		if p.activeColor == White {
			kingRookIdx = H1
			kingCastleIdx = G1
			kingPathIsClear = (p.PieceAt(G1) == NoPiece) && (p.PieceAt(F1) == NoPiece)
		} else {
			kingRookIdx = H8
			kingCastleIdx = G8
			kingPathIsClear = (p.PieceAt(G8) == NoPiece) && (p.PieceAt(F8) == NoPiece)
		}

		if kingPathIsClear && p.PieceAt(kingRookIdx).Type() == Rook {
			moves = append(moves, kingCastleIdx)
		}
	}

	return moves, capturesMoves
}

func PawnPseudoLegalMoves(p Position, pieceIdx int8) ([]int8, []int8) {
	var (
		moves         []int8
		capturesMoves []int8
	)

	piece := p.PieceAt(pieceIdx)
	pieceColor := piece.Color()

	direction := int8(1)
	if pieceColor == Black {
		direction = -1
	}

	rank, file := RankAndFile(pieceIdx)

	// 1 forward
	target1Idx := pieceIdx + (8 * direction)
	target1 := p.PieceAt(target1Idx)
	if target1 == NoPiece {
		moves = append(moves, target1Idx)

		// 2 forward
		if (pieceColor == White && rank == 1) || (pieceColor == Black && rank == 6) {
			target2Idx := pieceIdx + (16 * direction)
			target2 := p.PieceAt(target2Idx)
			if target2 == NoPiece {
				moves = append(moves, target2Idx)
			}
		}
	}

	// capture
	if file > 0 {
		leftTargetIdx := pieceIdx + (8 * direction) - 1
		leftTarget := p.PieceAt(leftTargetIdx)
		if (leftTarget != NoPiece && leftTarget.Color() != pieceColor) || leftTargetIdx == p.enPassantIdx {
			moves = append(moves, leftTargetIdx)
			capturesMoves = append(capturesMoves, leftTargetIdx)
		}
	}

	if file < 7 {
		rightTargetIdx := pieceIdx + (8 * direction) + 1
		rightTarget := p.PieceAt(rightTargetIdx)
		if rightTarget != NoPiece && rightTarget.Color() != pieceColor || rightTargetIdx == p.enPassantIdx {
			moves = append(moves, rightTargetIdx)
			capturesMoves = append(capturesMoves, rightTargetIdx)
		}
	}

	return moves, capturesMoves
}

// IsCheck
// Generate all captures moves from current king pos, but using opponent knight & queen
// if a captureMove is found it means that the current king is visible from these squares,
// so additional checks are required to verify if a capture is really possible
func (p Position) IsCheck() bool {
	return isKingInCheckByVector(p, p.activeColor)
}

func LegalMoves(pos Position) []Move {
	var moves []Move

	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece.Color() != pos.activeColor {
			continue
		}

		var pseudoLegalMoves []int8

		//fmt.Println("--- Piece legal moves ---")
		switch piece.Type() {
		case Pawn:
			pseudoLegalMoves, _ = PawnPseudoLegalMoves(pos, idx)
		case Rook:
			pseudoLegalMoves, _ = SliderPseudoLegalMoves(pos, idx)
		case Bishop:
			pseudoLegalMoves, _ = SliderPseudoLegalMoves(pos, idx)
		case Knight:
			pseudoLegalMoves, _ = KnightPseudoLegalMoves(pos, idx)
		case Queen:
			pseudoLegalMoves, _ = SliderPseudoLegalMoves(pos, idx)
		case King:
			pseudoLegalMoves, _ = KingPseudoLegalMoves(pos, idx)
		}

		for _, pseudoLegalMoveIdx := range pseudoLegalMoves {
			pseudoLegalMove := NewMove(piece, idx, pseudoLegalMoveIdx, NormalMove)
			newPos := pos.PositionAfterMove(pseudoLegalMove)
			if !isKingInCheckByVector(newPos, pos.activeColor) { // check is the new position that the initial color is not in check
				moves = append(moves, pseudoLegalMove)
			}
		}
	}

	return moves
}

func MoveGenerationTest(pos Position, depth int) int {
	if depth == 1 {
		return 1
	}

	posCount := 0

	legalMoves := LegalMoves(pos)

	//if depth != 2 {
	//	fmt.Println("Legal moves:", legalMoves)
	//}

	for _, move := range legalMoves {
		//if depth == 3 && !(move.piece.Type() == Knight && move.StartIdx() == B1) {
		//	continue
		//}
		newPos := pos.PositionAfterMove(move)
		nextDepth := depth - 1
		nextDepthResult := MoveGenerationTest(newPos, nextDepth)

		//if nextDepthResult != 20 && nextDepthResult != 1 {
		//	uciMove := move.UCI()
		//	fmt.Println(uciMove)
		//	panic("hello")
		//}

		posCount += nextDepthResult
	}

	return posCount
}

func PerftDivide(pos Position, depth int) {
	res := make(map[string]int)
	total := 0

	for _, move := range LegalMoves(pos) {
		res[move.UCI()] = MoveGenerationTest(pos.PositionAfterMove(move), depth)
		total += res[move.UCI()]
	}

	keys := make([]string, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Println(k, res[k])
	}

	fmt.Println("Nodes searched", total)
}
