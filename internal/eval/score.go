package eval

// Score is expressed in centipawns from the side-to-move perspective.
// Positive values favor the side to move, negative values favor the opponent.
type Score int32

const (
	DrawScore Score = 0

	// MateScore is the base mate value used by search.
	// Future search code should offset it by ply so shorter mates are preferred.
	MateScore Score = 30000

	// InfinityScore is a practical search bound above any normal static score.
	InfinityScore Score = 32000
)

func MateIn(ply int) Score {
	return MateScore - Score(ply)
}

func MatedIn(ply int) Score {
	return -MateScore + Score(ply)
}

