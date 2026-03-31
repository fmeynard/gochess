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

func appendMovesFromMask(pos *Position, piece Piece, startIdx int8, targets uint64, dst []Move, count int) int {
	for targets != 0 {
		targetIdx := int8(bits.TrailingZeros64(targets))
		targets &^= 1 << targetIdx

		flag := int8(NormalMove)
		if piece.Type() == Pawn && absInt8(targetIdx-startIdx) == 16 {
			flag = PawnDoubleMove
		} else if pos.board[targetIdx] != NoPiece {
			flag = Capture
		}

		dst[count] = Move{piece: piece, startIdx: startIdx, endIdx: targetIdx, flag: flag}
		count++
	}

	return count
}

func appendPromotionMoves(piece Piece, startIdx int8, targets uint64, dst []Move, count int) int {
	for targets != 0 {
		targetIdx := int8(bits.TrailingZeros64(targets))
		targets &^= 1 << targetIdx
		for _, flag := range promotionFlags {
			dst[count] = Move{piece: piece, startIdx: startIdx, endIdx: targetIdx, flag: flag}
			count++
		}
	}
	return count
}

func sliderTargetsMask(pos *Position, idx int8, pieceType int8, friendlyOcc uint64) uint64 {
	var attacks uint64
	var startDir, endDir int

	switch pieceType {
	case Bishop:
		startDir, endDir = 4, 8
	case Rook:
		startDir, endDir = 0, 4
	default:
		startDir, endDir = 0, 8
	}

	for dirIdx := startDir; dirIdx < endDir; dirIdx++ {
		ray := sliderAttackMasks[idx][dirIdx]
		if ray == 0 {
			continue
		}

		blocker := firstBlockerOnRay(pos.occupied, ray, rayDirections[dirIdx])
		if blocker == 0 {
			attacks |= ray
			continue
		}

		blockerIdx := int8(bits.TrailingZeros64(blocker))
		segment := ray ^ sliderAttackMasks[blockerIdx][dirIdx]
		if blocker&friendlyOcc != 0 {
			segment &^= blocker
		}
		attacks |= segment
	}

	return attacks
}

func (e *Engine) appendKingMoves(pos *Position, piece Piece, kingIdx int8, enemyColor int8, info positionAnalysis, dst []Move, count int) int {
	friendlyOcc := pos.OccupancyMask(pos.activeColor)
	targets := kingAttacksMask[kingIdx] &^ friendlyOcc
	occWithoutKing := pos.occupied &^ (uint64(1) << kingIdx)

	for targets != 0 {
		targetIdx := int8(bits.TrailingZeros64(targets))
		targets &^= 1 << targetIdx

		occ := occWithoutKing &^ (uint64(1) << targetIdx)
		if isSquareAttacked(pos, targetIdx, enemyColor, occ) {
			continue
		}

		flag := int8(NormalMove)
		if pos.board[targetIdx] != NoPiece {
			flag = Capture
		}
		dst[count] = Move{piece: piece, startIdx: kingIdx, endIdx: targetIdx, flag: flag}
		count++
	}

	if info.inCheck {
		return count
	}

	castleRights := pos.CastleRights()
	if pos.activeColor == White && kingIdx == E1 {
		if castleRights&KingSideCastle != 0 &&
			pos.board[F1] == NoPiece && pos.board[G1] == NoPiece &&
			!isSquareAttacked(pos, F1, enemyColor, occWithoutKing) &&
			!isSquareAttacked(pos, G1, enemyColor, occWithoutKing&^(uint64(1)<<G1)) {
			dst[count] = Move{piece: piece, startIdx: E1, endIdx: G1, flag: Castle}
			count++
		}
		if castleRights&QueenSideCastle != 0 &&
			pos.board[B1] == NoPiece && pos.board[C1] == NoPiece && pos.board[D1] == NoPiece &&
			!isSquareAttacked(pos, D1, enemyColor, occWithoutKing) &&
			!isSquareAttacked(pos, C1, enemyColor, occWithoutKing&^(uint64(1)<<C1)) {
			dst[count] = Move{piece: piece, startIdx: E1, endIdx: C1, flag: Castle}
			count++
		}
		return count
	}

	if pos.activeColor == Black && kingIdx == E8 {
		if castleRights&KingSideCastle != 0 &&
			pos.board[F8] == NoPiece && pos.board[G8] == NoPiece &&
			!isSquareAttacked(pos, F8, enemyColor, occWithoutKing) &&
			!isSquareAttacked(pos, G8, enemyColor, occWithoutKing&^(uint64(1)<<G8)) {
			dst[count] = Move{piece: piece, startIdx: E8, endIdx: G8, flag: Castle}
			count++
		}
		if castleRights&QueenSideCastle != 0 &&
			pos.board[B8] == NoPiece && pos.board[C8] == NoPiece && pos.board[D8] == NoPiece &&
			!isSquareAttacked(pos, D8, enemyColor, occWithoutKing) &&
			!isSquareAttacked(pos, C8, enemyColor, occWithoutKing&^(uint64(1)<<C8)) {
			dst[count] = Move{piece: piece, startIdx: E8, endIdx: C8, flag: Castle}
			count++
		}
	}

	return count
}

func (e *Engine) appendPawnMoves(pos *Position, piece Piece, idx int8, info positionAnalysis, pinRay uint64, isPinned bool, inCheckColor int8, dst []Move, count int) int {
	var quietTargets uint64
	var captureTargets uint64
	var promotionTargets uint64
	var epTarget uint64

	enemyOcc := pos.OpponentOccupiedMask()
	if pos.activeColor == White {
		oneStep := idx + 8
		if oneStep < 64 && pos.board[oneStep] == NoPiece {
			quietTargets |= uint64(1) << oneStep
			if idx >= A2 && idx <= H2 && pos.board[idx+16] == NoPiece {
				quietTargets |= uint64(1) << (idx + 16)
			}
		}
		captureTargets = e.moveGenerator.whitePawnCapturesMasks[idx] & enemyOcc
		if pos.enPassantIdx != NoEnPassant && e.moveGenerator.whitePawnCapturesMasks[idx]&(uint64(1)<<pos.enPassantIdx) != 0 {
			epTarget = uint64(1) << pos.enPassantIdx
		}
	} else {
		oneStep := idx - 8
		if oneStep >= 0 && pos.board[oneStep] == NoPiece {
			quietTargets |= uint64(1) << oneStep
			if idx >= A7 && idx <= H7 && pos.board[idx-16] == NoPiece {
				quietTargets |= uint64(1) << (idx - 16)
			}
		}
		captureTargets = e.moveGenerator.blackPawnCapturesMasks[idx] & enemyOcc
		if pos.enPassantIdx != NoEnPassant && e.moveGenerator.blackPawnCapturesMasks[idx]&(uint64(1)<<pos.enPassantIdx) != 0 {
			epTarget = uint64(1) << pos.enPassantIdx
		}
	}

	allTargets := quietTargets | captureTargets
	if info.checkerCount == 1 {
		allTargets &= info.evasionMask
	}
	if isPinned {
		allTargets &= pinRay
	}

	if pos.activeColor == White {
		promotionTargets = allTargets & (uint64(0xFF) << 56)
	} else {
		promotionTargets = allTargets & uint64(0xFF)
	}
	allTargets &^= promotionTargets

	count = appendMovesFromMask(pos, piece, idx, allTargets, dst, count)
	count = appendPromotionMoves(piece, idx, promotionTargets, dst, count)

	if epTarget == 0 {
		return count
	}
	if isPinned && pinRay&epTarget == 0 {
		return count
	}

	move := Move{piece: piece, startIdx: idx, endIdx: int8(bits.TrailingZeros64(epTarget)), flag: EnPassant}
	history := e.positionUpdater.MakeMove(pos, move)
	if !IsKingInCheck(pos, inCheckColor) {
		dst[count] = move
		count++
	}
	e.positionUpdater.UnMakeMove(pos, history)
	return count
}

func (e *Engine) legalMovesInto(pos *Position, dst []Move) int {
	count := 0
	inCheckColor := pos.activeColor

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

	kingPiece := pos.board[kingIdx]
	if kingPiece.Type() == King && kingPiece.Color() == pos.activeColor {
		count = e.appendKingMoves(pos, kingPiece, kingIdx, enemyColor, info, dst, count)
	}

	if info.checkerCount >= 2 {
		return count
	}

	piecesMask := friendlyOcc &^ (uint64(1) << kingIdx)
	for piecesMask != 0 {
		idx := int8(bits.TrailingZeros64(piecesMask))
		piecesMask &^= 1 << idx

		piece := pos.board[idx]
		pieceType := piece.Type()
		isPinned := info.pinnedMask&(1<<idx) != 0
		pinRay := info.pinRayBySq[idx]

		switch pieceType {
		case Pawn:
			count = e.appendPawnMoves(pos, piece, idx, info, pinRay, isPinned, inCheckColor, dst, count)
		case Knight:
			if isPinned {
				continue
			}
			targets := knightAttacksMask[idx] &^ friendlyOcc
			if info.checkerCount == 1 {
				targets &= info.evasionMask
			}
			count = appendMovesFromMask(pos, piece, idx, targets, dst, count)
		case Bishop, Rook, Queen:
			targets := sliderTargetsMask(pos, idx, pieceType, friendlyOcc)
			if info.checkerCount == 1 {
				targets &= info.evasionMask
			}
			if isPinned {
				targets &= pinRay
			}
			count = appendMovesFromMask(pos, piece, idx, targets, dst, count)
		}
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
