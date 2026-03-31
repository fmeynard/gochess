package chess

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKingSafetyEncodingRoundTrip(t *testing.T) {
	assert.Equal(t, NotCalculated, decodeKingSafety(encodeKingSafety(NotCalculated)))
	assert.Equal(t, KingIsSafe, decodeKingSafety(encodeKingSafety(KingIsSafe)))
	assert.Equal(t, KingIsCheck, decodeKingSafety(encodeKingSafety(KingIsCheck)))
}

func TestMoveHistoryPackedStateRoundTrip(t *testing.T) {
	pos, err := NewPositionFromFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq e3 0 1")
	if err != nil {
		t.Fatal(err)
	}
	pos.whiteKingSafety = KingIsSafe
	pos.blackKingSafety = KingIsCheck

	h := MoveHistory{
		packedState: packMoveHistoryMeta(pos),
	}

	assert.Equal(t, pos.whiteKingIdx, h.whiteKingIdx())
	assert.Equal(t, pos.blackKingIdx, h.blackKingIdx())
	assert.Equal(t, pos.enPassantIdx, h.enPassantIdx())
	assert.Equal(t, pos.whiteCastleRights, h.whiteCastleRights())
	assert.Equal(t, pos.blackCastleRights, h.blackCastleRights())
	assert.Equal(t, pos.whiteKingSafety, h.whiteKingSafety())
	assert.Equal(t, pos.blackKingSafety, h.blackKingSafety())
}
