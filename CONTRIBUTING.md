# Contributing

## Default Workflow

The default change flow in this repository is:

1. Create or identify an issue.
2. Create a new branch for the change.
3. Implement and validate the change.
4. Open a pull request.
5. Merge only after review.

For external experiments or larger divergent work, a separate repo or fork is acceptable, but the normal in-repo path should still end in a PR.

## Branching

Recommended branch naming:

- `issue-123-short-topic`
- `perf-legal-movegen`
- `docs-board-movegen-layout`

Keep one branch focused on one issue or one coherent slice.

## Pull Requests

Every material change should go through a PR, even if the branch is short-lived.

PRs should include:

- a link to the issue
- a concise summary of what changed
- validation results
- benchmark results if the change touches `internal/board` or `internal/movegen`
- explicit risks or follow-ups if something remains open

## Required Validation

Minimum baseline for most PRs:

```bash
go test ./...
```

If the PR touches `internal/board` or `internal/movegen`, include:

```bash
BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

If a change only touches docs or repo metadata, benchmarking is not required.

## Benchmark Expectations

For board or move-generation changes, the PR should state:

- benchmark command used
- node count
- elapsed time
- whether the result is better, neutral, or worse than the current reference

If performance regresses intentionally, say so explicitly and explain why.

## Documentation Expectations

Update documentation when a change affects:

- package layout
- benchmark reference values
- contributor workflow
- architectural assumptions

Relevant docs live under `docs/`.
