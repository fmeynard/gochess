package internal

const perftTTSize = 1 << 20

type perftTTEntry struct {
	keyWithDepth uint64
	count        uint64
}

type perftTT struct {
	entries []perftTTEntry
	mask    uint64
}

func newPerftTT() *perftTT {
	return &perftTT{
		entries: make([]perftTTEntry, perftTTSize),
		mask:    perftTTSize - 1,
	}
}

func (tt *perftTT) probe(zobrist uint64, depth int8) (uint64, bool) {
	combined := zobrist ^ (uint64(depth) << 56)
	e := &tt.entries[zobrist&tt.mask]
	if e.keyWithDepth == combined {
		return e.count, true
	}
	return 0, false
}

func (tt *perftTT) store(zobrist uint64, depth int8, count uint64) {
	tt.entries[zobrist&tt.mask] = perftTTEntry{
		keyWithDepth: zobrist ^ (uint64(depth) << 56),
		count:        count,
	}
}
