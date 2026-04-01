package search

import (
	board "chessV2/internal/board"
	"chessV2/internal/eval"
)

const searchTTSize = 1 << 20
const searchMatePlyWindow = 256

type ttBound uint8

const (
	ttBoundExact ttBound = iota + 1
	ttBoundLower
	ttBoundUpper
)

type ttEntry struct {
	key      uint64
	depth    int16
	score    eval.Score
	bound    ttBound
	bestMove board.Move
}

type searchTT struct {
	entries []ttEntry
	mask    uint64
}

func newSearchTT() *searchTT {
	return &searchTT{
		entries: make([]ttEntry, searchTTSize),
		mask:    searchTTSize - 1,
	}
}

func (tt *searchTT) clear() {
	clear(tt.entries)
}

func (tt *searchTT) probe(key uint64, depth int, ply int) (ttEntry, bool) {
	entry := tt.entries[key&tt.mask]
	if entry.key != key || int(entry.depth) < depth {
		return ttEntry{}, false
	}
	entry.score = ttScoreFromStored(entry.score, ply)
	return entry, true
}

func (tt *searchTT) store(key uint64, depth int, ply int, score eval.Score, bound ttBound, bestMove board.Move) {
	index := key & tt.mask
	current := tt.entries[index]
	if current.key == key && int(current.depth) > depth && bestMove == (board.Move{}) {
		return
	}

	tt.entries[index] = ttEntry{
		key:      key,
		depth:    int16(depth),
		score:    ttScoreForStorage(score, ply),
		bound:    bound,
		bestMove: bestMove,
	}
}

func ttScoreForStorage(score eval.Score, ply int) eval.Score {
	if score >= eval.MateScore-searchMatePlyWindow {
		return score + eval.Score(ply)
	}
	if score <= -eval.MateScore+searchMatePlyWindow {
		return score - eval.Score(ply)
	}
	return score
}

func ttScoreFromStored(score eval.Score, ply int) eval.Score {
	if score >= eval.MateScore-searchMatePlyWindow {
		return score - eval.Score(ply)
	}
	if score <= -eval.MateScore+searchMatePlyWindow {
		return score + eval.Score(ply)
	}
	return score
}
