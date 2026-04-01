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

## Scoring Version Workflow

For a new scoring version, use the normal issue -> branch -> PR flow, then do the versioning and comparison only after the PR is merged.

Recommended sequence:

1. Create one or more issues for the scoring ideas.
2. Implement them on a dedicated branch.
3. Open one PR for the scoring slice.
4. Merge the PR to `main`.
5. Tag the merged `main` commit as the next `score-vN`.
6. Run a match against the previous score tag, for example:

```bash
make runner MODE=plain GAMES=50 MOVETIME=300 OPPONENT_TAG=score-v2
```

7. Update [`docs/match-history.md`](/home/fab/Projects/gochess/docs/match-history.md):
   - add the new version to the version summary table
   - add the new match result at the top of the results table
   - keep both tables in reverse chronological order

Do not tag a scoring version from an in-flight branch unless the tag is explicitly meant to be experimental.

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
