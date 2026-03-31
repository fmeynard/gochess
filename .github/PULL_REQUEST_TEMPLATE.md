## Summary

- Issue:
- Scope:
- Why:

## Changes

- 

## Validation

```bash
GOCACHE=/home/fab/Projects/gochess/.codex-tmp/go-build-cache go test ./...
```

Results:

```text
paste results here
```

## Benchmark

Required if this PR touches `internal/board` or `internal/movegen`.

Command:

```bash
BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Results:

```text
paste results here
```

Assessment:

- [ ] Faster than current reference
- [ ] Roughly neutral
- [ ] Slower, but intentional
- [ ] Not applicable

## Risks / Follow-Ups

- 
