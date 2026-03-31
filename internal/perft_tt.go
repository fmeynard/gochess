package internal

type perftTTKey struct {
	zobrist uint64
	depth   int8
}

type perftTT struct {
	data map[perftTTKey]uint64
}

func newPerftTT() *perftTT {
	return &perftTT{data: make(map[perftTTKey]uint64, 1<<16)}
}

func (tt *perftTT) probe(zobrist uint64, depth int8) (uint64, bool) {
	v, ok := tt.data[perftTTKey{zobrist: zobrist, depth: depth}]
	return v, ok
}

func (tt *perftTT) store(zobrist uint64, depth int8, count uint64) {
	tt.data[perftTTKey{zobrist: zobrist, depth: depth}] = count
}
