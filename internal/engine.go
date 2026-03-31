package internal

import "math/bits"

type Engine struct {
	moveGenerator   *BitsBoardMoveGenerator
	positionUpdater *PositionUpdater
	usePerftTricks  bool
}

const (
	MaxPerftPly   = 64
	MaxLegalMoves = 256
	MaxTargets    = 28
)

var promotionFlags = [4]int8{QueenPromotion, KnightPromotion, BishopPromotion, RookPromotion}

var rayDirections = [8]int8{West, East, South, North, SouthWest, SouthEast, NorthWest, NorthEast}

func NewEngine() *Engine {
	moveGenerator := NewBitsBoardMoveGenerator()
	positionUpdater := NewPositionUpdater(moveGenerator)

	return &Engine{
		moveGenerator:   moveGenerator,
		positionUpdater: positionUpdater,
		usePerftTricks:  true,
	}
}

func (e *Engine) SetPerftTricks(enabled bool) {
	e.usePerftTricks = enabled
	e.positionUpdater.SetTrackZobrist(enabled)
}

func (e *Engine) StartGame() {}

func (e *Engine) Move() {}

type posInfo struct {
	inCheck      bool
	checkerCount int
	evasionMask  uint64
	pinnedMask   uint64
	pinCount     int
	pinSquares   [8]int8
	pinRayMasks  [8]uint64
}

func computePosInfo(pos *Position, kingIdx int8, friendlyOcc, enemyOcc uint64) posInfo {
	var info posInfo
	checkers := uint64(0)
	blockMask := uint64(0)

	var pawnColorIdx int
	if pos.activeColor == White {
		pawnColorIdx = 1
	} else {
		pawnColorIdx = 0
	}
	checkers |= pawnAttacksBy[pawnColorIdx][kingIdx] & pos.pawnBoard & enemyOcc
	checkers |= knightAttacksMask[kingIdx] & pos.knightBoard & enemyOcc

	rookQueens := (pos.rookBoard | pos.queenBoard) & enemyOcc
	bishopQueens := (pos.bishopBoard | pos.queenBoard) & enemyOcc

	for dirIdx := 0; dirIdx < 8; dirIdx++ {
		var sliders uint64
		if dirIdx < 4 {
			sliders = rookQueens
		} else {
			sliders = bishopQueens
		}
		if sliders == 0 {
			continue
		}
		ray := sliderAttackMasks[kingIdx][dirIdx]
		if ray == 0 {
			continue
		}
		dir := rayDirections[dirIdx]
		first := firstBlockerOnRay(pos.occupied, ray, dir)
		if first == 0 {
			continue
		}
		if first&enemyOcc != 0 {
			if first&sliders != 0 {
				checkers |= first
				checkerIdx := int8(bits.TrailingZeros64(first))
				blockMask |= ray ^ sliderAttackMasks[checkerIdx][dirIdx]
			}
		} else if first&friendlyOcc != 0 {
			second := firstBlockerOnRay(pos.occupied&^first, ray, dir)
			if second != 0 && second&sliders != 0 {
				info.pinnedMask |= first
				info.pinSquares[info.pinCount] = int8(bits.TrailingZeros64(first))
				info.pinRayMasks[info.pinCount] = ray
				info.pinCount++
			}
		}
	}

	info.checkerCount = bits.OnesCount64(checkers)
	info.inCheck = info.checkerCount > 0
	if info.checkerCount == 1 {
		info.evasionMask = checkers | blockMask
	}
	return info
}

func (pi *posInfo) pinRayFor(sq int8) uint64 {
	for i := 0; i < pi.pinCount; i++ {
		if pi.pinSquares[i] == sq {
			return pi.pinRayMasks[i]
		}
	}
	return 0
}

func classifyMove(pos *Position, piece Piece, startIdx, targetIdx int8) int8 {
	if piece.Type() == King && absInt8(targetIdx-startIdx) == 2 {
		return Castle
	}
	if piece.Type() == Pawn {
		if pos.enPassantIdx != NoEnPassant && targetIdx == pos.enPassantIdx {
			return EnPassant
		}
		if absInt8(targetIdx-startIdx) == 16 {
			return PawnDoubleMove
		}
	}
	if pos.board[targetIdx] != NoPiece {
		return Capture
	}
	return NormalMove
}

func (e *Engine) legalMovesInto(pos *Position, dst []Move) int {
	count := 0
	var targets [MaxTargets]int8
	initialColor := pos.activeColor

	var kingIdx int8
	if pos.activeColor == White {
		kingIdx = pos.whiteKingIdx
	} else {
		kingIdx = pos.blackKingIdx
	}

	friendlyOcc := pos.OccupancyMask(pos.activeColor)
	enemyOcc := pos.OpponentOccupiedMask()
	enemyColor := pos.OpponentColor()

	info := computePosInfo(pos, kingIdx, friendlyOcc, enemyOcc)
	occWithoutKing := pos.occupied &^ (1 << kingIdx)

	piecesMask := friendlyOcc
	for piecesMask != 0 {
		idx := int8(bits.TrailingZeros64(piecesMask))

		piece := pos.board[idx]
		pieceType := piece.Type()
		var pseudoCount int

		if info.checkerCount >= 2 && pieceType != King {
			piecesMask &^= 1 << idx
			continue
		}

		switch pieceType {
		case Pawn:
			pseudoCount, _ = e.moveGenerator.PawnPseudoLegalMovesInto(pos, idx, targets[:])
		case Knight:
			pseudoCount = e.moveGenerator.KnightPseudoLegalMovesInto(pos, idx, targets[:])
		case Bishop:
			pseudoCount = e.moveGenerator.SliderPseudoLegalMovesInto(pos, idx, Bishop, targets[:])
		case Rook:
			pseudoCount = e.moveGenerator.SliderPseudoLegalMovesInto(pos, idx, Rook, targets[:])
		case Queen:
			pseudoCount = e.moveGenerator.SliderPseudoLegalMovesInto(pos, idx, Queen, targets[:])
		case King:
			pseudoCount = e.moveGenerator.KingPseudoLegalMovesInto(pos, idx, targets[:])
		}

		isPinned := info.pinnedMask&(1<<idx) != 0
		var pinRay uint64
		if isPinned {
			pinRay = info.pinRayFor(idx)
		}

		for i := 0; i < pseudoCount; i++ {
			targetIdx := targets[i]

			if pieceType == King {
				if absInt8(targetIdx-idx) == 2 {
					if info.inCheck {
						continue
					}
					step := int8(1)
					if targetIdx < idx {
						step = -1
					}
					if isSquareAttacked(pos, idx+step, enemyColor, occWithoutKing) {
						continue
					}
				}
				occ := occWithoutKing &^ (1 << targetIdx)
				if isSquareAttacked(pos, targetIdx, enemyColor, occ) {
					continue
				}
				dst[count] = NewMove(piece, idx, targetIdx, classifyMove(pos, piece, idx, targetIdx))
				count++
				continue
			}

			if info.inCheck && info.checkerCount == 1 {
				isEP := pieceType == Pawn && pos.enPassantIdx != NoEnPassant && targetIdx == pos.enPassantIdx
				if !isEP && info.evasionMask&(1<<targetIdx) == 0 {
					continue
				}
			}

			if pieceType == Pawn && isPromotionSquare(pos.activeColor, targetIdx) {
				if info.inCheck && info.checkerCount == 1 && info.evasionMask&(1<<targetIdx) == 0 {
					continue
				}
				for _, flag := range promotionFlags {
					move := NewMove(piece, idx, targetIdx, flag)
					if !isPinned || pinRay&(1<<targetIdx) != 0 {
						dst[count] = move
						count++
					}
				}
				continue
			}

			move := NewMove(piece, idx, targetIdx, classifyMove(pos, piece, idx, targetIdx))

			if !info.inCheck && !isPinned {
				isEP := pieceType == Pawn && pos.enPassantIdx != NoEnPassant && targetIdx == pos.enPassantIdx
				if !isEP {
					dst[count] = move
					count++
					continue
				}
			}

			if !info.inCheck && isPinned {
				if pinRay&(1<<targetIdx) != 0 {
					isEP := pieceType == Pawn && pos.enPassantIdx != NoEnPassant && targetIdx == pos.enPassantIdx
					if isEP {
						history := e.positionUpdater.MakeMove(pos, move)
						if !IsKingInCheck(pos, initialColor) {
							dst[count] = move
							count++
						}
						e.positionUpdater.UnMakeMove(pos, history)
						continue
					}
					dst[count] = move
					count++
				}
				continue
			}

			if info.inCheck && info.checkerCount == 1 {
				isEP := pieceType == Pawn && pos.enPassantIdx != NoEnPassant && targetIdx == pos.enPassantIdx
				if isEP {
					history := e.positionUpdater.MakeMove(pos, move)
					if !IsKingInCheck(pos, initialColor) {
						dst[count] = move
						count++
					}
					e.positionUpdater.UnMakeMove(pos, history)
					continue
				}
				if isPinned && pinRay&(1<<targetIdx) == 0 {
					continue
				}
				dst[count] = move
				count++
				continue
			}

			history := e.positionUpdater.MakeMove(pos, move)
			if !IsKingInCheck(pos, initialColor) {
				dst[count] = move
				count++
			}
			e.positionUpdater.UnMakeMove(pos, history)
		}

		piecesMask &^= 1 << idx
	}

	return count
}

func (e *Engine) LegalMoves(pos *Position) []Move {
	var buf [MaxLegalMoves]Move
	count := e.legalMovesInto(pos, buf[:])
	moves := make([]Move, count)
	copy(moves, buf[:count])
	return moves
}

func (e *Engine) PerftDivide(pos *Position, depth int) (map[string]uint64, uint64) {
	var moveBuffers [MaxPerftPly][MaxLegalMoves]Move
	var tt *perftTT
	if e.usePerftTricks {
		tt = newPerftTT()
	}
	res := make(map[string]uint64)
	total := uint64(0)

	rootMoves := moveBuffers[0][:]
	rootCount := e.legalMovesInto(pos, rootMoves)
	for i := 0; i < rootCount; i++ {
		move := rootMoves[i]
		history := e.positionUpdater.MakeMove(pos, move)
		uci := move.UCI()
		res[uci] = e.moveGenerationTestWithBuffers(pos, depth, 1, &moveBuffers, tt)
		total += res[uci]
		e.positionUpdater.UnMakeMove(pos, history)
	}

	return res, total
}

func (e *Engine) MoveGenerationTest(pos *Position, depth int) uint64 {
	var moveBuffers [MaxPerftPly][MaxLegalMoves]Move
	var tt *perftTT
	if e.usePerftTricks {
		tt = newPerftTT()
	}
	return e.moveGenerationTestWithBuffers(pos, depth, 0, &moveBuffers, tt)
}

func (e *Engine) moveGenerationTestWithBuffers(pos *Position, depth int, ply int, moveBuffers *[MaxPerftPly][MaxLegalMoves]Move, tt *perftTT) uint64 {
	if depth == 1 {
		return uint64(1)
	}
	if tt != nil {
		if count, ok := tt.probe(pos.zobristKey, int8(depth)); ok {
			return count
		}
	}
	if e.usePerftTricks && depth == 2 {
		result := uint64(e.legalMovesInto(pos, moveBuffers[ply][:]))
		if tt != nil {
			tt.store(pos.zobristKey, int8(depth), result)
		}
		return result
	}

	moves := moveBuffers[ply][:]
	moveCount := e.legalMovesInto(pos, moves)
	posCount := uint64(0)
	for i := 0; i < moveCount; i++ {
		move := moves[i]
		history := e.positionUpdater.MakeMove(pos, move)

		nextDepth := depth - 1
		nextDepthResult := e.moveGenerationTestWithBuffers(pos, nextDepth, ply+1, moveBuffers, tt)
		e.positionUpdater.UnMakeMove(pos, history)

		posCount += nextDepthResult
	}

	if tt != nil {
		tt.store(pos.zobristKey, int8(depth), posCount)
	}
	return posCount
}
