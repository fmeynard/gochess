package internal

import "math/bits"

type Engine struct {
	moveGenerator   *BitsBoardMoveGenerator
	positionUpdater moveApplier
	usePerftTricks  bool
}

const (
	MaxPerftPly   = 64
	MaxLegalMoves = 256
	MaxTargets    = 28
)

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
	if enabled {
		e.positionUpdater = NewPositionUpdater(e.moveGenerator)
		return
	}
	e.positionUpdater = NewPlainPositionUpdater(e.moveGenerator)
}

func (e *Engine) StartGame() {}

func (e *Engine) Move() {}

func (e *Engine) legalMovesInto(pos *Position, dst []Move) int {
	count := 0
	var targets [MaxTargets]int8
	initialColor := pos.activeColor
	inCheckColor := initialColor
	enPassantIdx := pos.enPassantIdx

	var kingIdx int8
	if pos.activeColor == White {
		kingIdx = pos.whiteKingIdx
	} else {
		kingIdx = pos.blackKingIdx
	}

	friendlyOcc := pos.OccupancyMask(pos.activeColor)
	enemyOcc := pos.OpponentOccupiedMask()
	enemyColor := pos.OpponentColor()

	info := computePositionAnalysis(pos, kingIdx, friendlyOcc, enemyOcc)
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
			pinRay = info.pinRayBySq[idx]
		}

		for i := 0; i < pseudoCount; i++ {
			targetIdx := targets[i]
			targetMask := uint64(1) << targetIdx
			targetPiece := pos.board[targetIdx]
			isEP := pieceType == Pawn && enPassantIdx != NoEnPassant && targetIdx == enPassantIdx

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
				flag := int8(NormalMove)
				if absInt8(targetIdx-idx) == 2 {
					flag = Castle
				} else if targetPiece != NoPiece {
					flag = Capture
				}
				dst[count] = Move{piece: piece, startIdx: idx, endIdx: targetIdx, flag: flag}
				count++
				continue
			}

			if info.inCheck && info.checkerCount == 1 {
				if !isEP && info.evasionMask&targetMask == 0 {
					continue
				}
			}

			if pieceType == Pawn && isPromotionSquare(pos.activeColor, targetIdx) {
				if info.inCheck && info.checkerCount == 1 && info.evasionMask&targetMask == 0 {
					continue
				}
				for _, flag := range promotionFlags {
					move := Move{piece: piece, startIdx: idx, endIdx: targetIdx, flag: flag}
					if !isPinned || pinRay&(1<<targetIdx) != 0 {
						dst[count] = move
						count++
					}
				}
				continue
			}

			flag := int8(NormalMove)
			if isEP {
				flag = EnPassant
			} else if pieceType == Pawn && absInt8(targetIdx-idx) == 16 {
				flag = PawnDoubleMove
			} else if targetPiece != NoPiece {
				flag = Capture
			}
			move := Move{piece: piece, startIdx: idx, endIdx: targetIdx, flag: flag}

			if !info.inCheck && !isPinned {
				if !isEP {
					dst[count] = move
					count++
					continue
				}
			}

			if !info.inCheck && isPinned {
				if pinRay&targetMask != 0 {
					if isEP {
						history := e.positionUpdater.MakeMove(pos, move)
						if !IsKingInCheck(pos, inCheckColor) {
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
				if isEP {
					history := e.positionUpdater.MakeMove(pos, move)
					if !IsKingInCheck(pos, inCheckColor) {
						dst[count] = move
						count++
					}
					e.positionUpdater.UnMakeMove(pos, history)
					continue
				}
				if isPinned && pinRay&targetMask == 0 {
					continue
				}
				dst[count] = move
				count++
				continue
			}

			history := e.positionUpdater.MakeMove(pos, move)
			if !IsKingInCheck(pos, inCheckColor) {
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
