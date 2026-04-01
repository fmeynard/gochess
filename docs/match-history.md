# Match History

## Version Summaries

| Version | Commit | Summary |
| --- | --- | --- |
| [`score-v3`](#score-v3) | `f3719bf` | Phase-aware middlegame/endgame eval, passed pawns, protection-aware piece safety, pawn structure, repetition contempt. |
| [`score-v2`](#score-v2) | `8a3e0a4` | Exposed-piece penalties and king-safety evaluation on top of mobility scoring. |
| [`score-v1`](#score-v1) | `64d4800` | Mobility-based static evaluation, plus the first stable matchable scoring tag. |
| [`score-v0`](#score-v0) | `22d7688` | Baseline static evaluation before mobility and later positional heuristics. |

### score-v3

- Commit: `f3719bf`
- Highlights:
  - phase-aware PST interpolation
  - endgame king activity
  - passed pawn bonuses
  - protection-aware piece safety
  - isolated and doubled pawn penalties
  - repetition contempt when ahead

### score-v2

- Commit: `8a3e0a4`
- Highlights:
  - exposed-piece malus weighted by attacked piece value
  - basic king safety with king-ring attacks and pawn shield

### score-v1

- Commit: `64d4800`
- Highlights:
  - mobility bonuses for active minor and major pieces

### score-v0

- Commit: `22d7688`
- Highlights:
  - baseline material plus piece-square tables

## Results

| Date (UTC) | Current | Opponent | Movetime | Games | Score | W/D/L | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-04-01T13:19:29Z | [`score-v3`](#score-v3) (`f3719bf`) | [`score-v2`](#score-v2) (`8a3e0a4`) | 300ms | 50 | 36.0/50 | 26/20/4 | strong gain over `score-v2`; white side especially dominant at 17/8/0 |
| 2026-04-01T01:29:26Z | [`score-v2`](#score-v2) (`8a3e0a4`) | [`score-v1`](#score-v1) (`64d4800`) | 300ms | 50 | 28.0/50 | 12/32/6 | exposed-piece and king-safety evaluation vs previous score version |
| 2026-04-01T01:11:56Z | [`score-v1`](#score-v1) (`64d4800`) | [`score-v0`](#score-v0) (`22d7688`) | 300ms | 50 | 29.5/50 | 16/27/7 | current only illegal moves fixed; opponent still emits illegal moves |
| 2026-03-31T22:33:13Z | `self` (`691a34a`) | `self` (`691a34a`) | 50ms | 2 | 1.0/2 | 0/2/0 | sample self-play smoke run |
