package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputePositionAnalysis(t *testing.T) {
	data := map[string]struct {
		fen               string
		expectedInCheck   bool
		expectedCheckers  int
		expectedPinned    uint64
		expectedPinSquare int8
		expectedEvasion   uint64
	}{
		"Single rook check builds evasion mask": {
			fen:               "4k3/8/8/8/8/8/4r3/4K3 w - - 0 1",
			expectedInCheck:   true,
			expectedCheckers:  1,
			expectedPinned:    0,
			expectedPinSquare: NoEnPassant,
			expectedEvasion:   (uint64(1) << E2),
		},
		"Double check reports two checkers": {
			fen:               "4r2k/8/8/1b6/8/8/4K3/8 w - - 0 1",
			expectedInCheck:   true,
			expectedCheckers:  2,
			expectedPinned:    0,
			expectedPinSquare: NoEnPassant,
			expectedEvasion:   0,
		},
		"Pinned knight is recorded on pin ray": {
			fen:               "4r2k/8/8/8/8/8/4N3/4K3 w - - 0 1",
			expectedInCheck:   false,
			expectedCheckers:  0,
			expectedPinned:    uint64(1) << E2,
			expectedPinSquare: E2,
			expectedEvasion:   0,
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fen)
			if err != nil {
				t.Fatal(err)
			}

			kingIdx := pos.whiteKingIdx
			friendlyOcc := pos.whiteOccupied
			enemyOcc := pos.blackOccupied
			if pos.activeColor == Black {
				kingIdx = pos.blackKingIdx
				friendlyOcc = pos.blackOccupied
				enemyOcc = pos.whiteOccupied
			}

			var analysis positionAnalysis
			computePositionAnalysis(pos, kingIdx, friendlyOcc, enemyOcc, &analysis)
			assert.Equal(t, d.expectedInCheck, analysis.inCheck)
			assert.Equal(t, uint8(d.expectedCheckers), analysis.checkerCount)
			assert.Equal(t, d.expectedPinned, analysis.pinnedMask)
			assert.Equal(t, d.expectedEvasion, analysis.evasionMask)

			if d.expectedPinSquare != NoEnPassant {
				assert.NotZero(t, analysis.pinRayBySq[d.expectedPinSquare])
			}
		})
	}
}
