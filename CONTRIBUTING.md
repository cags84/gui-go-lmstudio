# Contributing

## Commits

- Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `chore:`, etc.).
- No AI attribution or `Co-Authored-By` trailers.
- One work unit per commit: the behavior change, its tests, and any doc updates belong together.

## Workflow

Write a failing test before writing the behavior it covers (test-first). Keep commits small and reviewable rather than batching unrelated changes.

## Where code lives

| Location | Purpose |
|---|---|
| `internal/<feature>/` | Go application features (server control, models, logs, etc.) |
| `internal/lms/` | The `lms` CLI boundary — process invocation, JSON/NDJSON parsing |
| `frontend/src/` | React UI |

## Before opening a PR

Run these from the repository root:

```sh
npm --prefix frontend run build
go vet ./...
go test ./... -short
gofmt -l .
```

`gofmt -l .` must print nothing. Fix any files it lists with `gofmt -w`.
