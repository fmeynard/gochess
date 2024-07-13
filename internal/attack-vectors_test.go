package internal

import "testing"

func TestDraw(t *testing.T) {
	pos, _ := NewPositionFromFEN("8/8/8/8/8/8/8/4N3 b - - 0 1")
	updateAttackVectors(&pos)

	draw(pos.whiteAttacks)
}
