package board

import "testing"

func benchmarkMakeUnmakeMove(b *testing.B, fen string, move Move) {
	pos, err := NewPositionFromFEN(fen)
	if err != nil {
		b.Fatal(err)
	}

	updater := NewPositionUpdater()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history := updater.MakeMove(pos, move)
		updater.UnMakeMove(pos, history)
	}
}

func BenchmarkPositionUpdaterMakeUnmakePawnQuiet(b *testing.B) {
	benchmarkMakeUnmakeMove(b, FenStartPos, NewMove(Piece(White|Pawn), E2, E3, NormalMove))
}

func BenchmarkPositionUpdaterMakeUnmakePawnDouble(b *testing.B) {
	benchmarkMakeUnmakeMove(b, FenStartPos, NewMove(Piece(White|Pawn), E2, E4, PawnDoubleMove))
}

func BenchmarkPositionUpdaterMakeUnmakePawnCapture(b *testing.B) {
	benchmarkMakeUnmakeMove(b, "4k3/8/8/3p4/4P3/8/8/4K3 w - - 0 1", NewMove(Piece(White|Pawn), E4, D5, Capture))
}

func BenchmarkPositionUpdaterMakeUnmakeKnightQuiet(b *testing.B) {
	benchmarkMakeUnmakeMove(b, FenStartPos, NewMove(Piece(White|Knight), G1, F3, NormalMove))
}

func BenchmarkPositionUpdaterMakeUnmakeRookQuiet(b *testing.B) {
	benchmarkMakeUnmakeMove(b, "4k3/8/8/8/8/8/R7/4K3 w - - 0 1", NewMove(Piece(White|Rook), A2, A7, NormalMove))
}

func BenchmarkPositionUpdaterMakeUnmakeKingQuiet(b *testing.B) {
	benchmarkMakeUnmakeMove(b, "4k3/8/8/8/8/8/8/4K3 w - - 0 1", NewMove(Piece(White|King), E1, E2, NormalMove))
}

func BenchmarkPositionUpdaterMakeUnmakeCastle(b *testing.B) {
	benchmarkMakeUnmakeMove(b, "4k2r/8/8/8/8/8/8/R3K2R w KQ - 0 1", NewMove(Piece(White|King), E1, G1, Castle))
}

func BenchmarkPositionUpdaterMakeUnmakePromotion(b *testing.B) {
	benchmarkMakeUnmakeMove(b, "4k3/3P4/8/8/8/8/8/4K3 w - - 0 1", NewMove(Piece(White|Pawn), D7, D8, QueenPromotion))
}
