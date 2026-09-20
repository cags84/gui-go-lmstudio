# Feature: lmstudio-desktop-controller

Repository-relative locator: `odd/tasks/lmstudio-desktop-controller.md`
Engram mirror topic: `odd/lmstudio-desktop-controller/tasks` (project `gui-go-lmstudio`)
Plan of record: `~/.claude/plans/pasted-content-id-4abf-crear-una-tidy-parnas.md` (approved 2026-09-20)

## Objective

Ship a cross-platform desktop app (macOS/Windows/Linux) in Go + Wails v3 + React/TS that controls the LM Studio local server through the `lms` CLI: start/stop, choose localhost vs LAN exposure, CORS, list/load/unload models, stream live logs, and show system info.

## Problem and why

Today the LM Studio server is driven from a terminal (`lms server start --port --bind --cors`, `lms ps`, `lms log stream`). The user wants the same power with a GUI that also visualizes what is happening, published as a public MIT repo (`cags84/gui-go-lmstudio`).

## Scope (authorized)

- New project under `/Users/carlos/Code/Claude/Projects/gui-go-lmstudio` (module `github.com/cags84/gui-go-lmstudio`).
- Everything listed in the milestones below. Out of scope (roadmap only): model download UI, REST polling, LM Link, chat.

## Constraints

- Artifacts in English; conversation in Spanish. Conventional commits, no AI attribution.
- Every `lms` interaction goes through `internal/lms.Runner`; unit tests never spawn the real binary.
- TDD mode: **strict** (source: session configuration "Strict TDD Mode: enabled"). Runners: Go `go test ./...`; frontend `npm run test` (vitest) for pure logic. Observed RED before GREEN on every behavior.
- Per-task size heuristic ~400 authored changed lines (advisory, not a cap).
- Delivery strategy: `ask-on-risk`; chain strategy fixed by the approved plan as `stacked-to-main` (one PR per milestone, sequential into `main`).
- Receipt-driven review (RDD): see status recorded per task below.

## Checklist

| ID | Task | Route | Status |
|---|---|---|---|
| T1 | Bootstrap: `wails3 init` (react), module path, MIT, README, .gitignore, CI (vet + test + gofmt on 3 OSes), git init, commits, `gh repo create --public --push` | inline (generator) + delegated writer for README/CI/CONTRIBUTING | [x] |
| T2 | `internal/lms`: Runner interface + exec impl, Client `ServerStatus`/`PS`/`LS` with JSON fixtures, FakeRunner; integration test skipped with `-short` | delegated writer | [ ] |
| T3 | Server lifecycle: `ServerStart{Port,Bind,CORS}`, `ServerStop`, Wails service, React status card + controls, Local/Network warning | delegated writer | [ ] |
| T4 | Models: on-disk table (filters), loaded list + unload, load dialog (gpu, context, ttl, identifier) | delegated writer | [ ] |
| T5 | Live logs: `lms log stream --json --stats` supervisor, ring buffer, `logs:entry` events, React log view with filters | delegated writer | [ ] |
| T6 | System panel + tray: lms detection/version, daemon status, runtime engines, systray quick actions | delegated writer | [ ] |
| T7 | Packaging: `wails3 package` in CI on tag, README install notes + screenshots | delegated writer | [ ] |

## Acceptance criteria

- `go test ./... -short` green locally and in CI on ubuntu/macos/windows.
- `wails3 dev` opens the app; start/stop/bind/cors reflected by `lms server status --json`.
- Loaded models match `lms ps --json`; unload clears them.
- Server log entries appear live after a `curl localhost:1234/v1/models`.
- `wails3 package` yields a runnable macOS binary; CI uploads Windows/Linux artifacts.

## Applicable checks per task

- Go: `go vet ./...`, `go test ./... -short` (focused package first).
- Frontend: `npm run build`, `npm run test` when tests exist.
- Runtime harness: `wails3 dev` scenario named in the task evidence, or explicit `N/A` with reason.

## Progress and evidence

### T1 — Bootstrap
- Status: **done** (2026-09-20), pending only the public repo creation.
- Route: generator run inline; docs/CI/metadata delegated to one bounded writer (writer trigger: 7 non-trivial files).
- Commits on `main`:
  - `ebb3c8f` chore: scaffold Wails v3 project with React and TypeScript
  - `460972f` chore: set application identity for LM Studio Controller
  - `bf02c5b` build: track frontend lockfile and generated Wails bindings
  - `68de6b8` docs: add README, MIT license and contributing guide
  - `04d4725` ci: add cross-platform build and test workflow
- TDD: **N/A** for this task — documentation, CI configuration and metadata, no new behavior.
- Verification, run against a **fresh `git clone` of HEAD** (not the dirty worktree), reproducing every CI step:
  - `npm ci --prefix frontend` → PASS
  - `npm run build --prefix frontend` → PASS, 46 modules, dist/index.html + dist/assets/index-*.js
  - `go vet ./...` → PASS, no output
  - `go test ./... -short` → PASS, `no test files` in all 4 packages (expected at this stage)
  - `gofmt -l .` → PASS, empty
  - `wails3 build` (dirty worktree, earlier) → PASS, `bin/lm-studio-controller`, 9.8 MB
- Defects found and fixed before the commit, each verified rather than assumed:
  - A clean clone had no `frontend/package-lock.json` (reproduced: `npm ci` aborts), so CI could never have run. Now tracked.
  - `frontend/bindings/` was untracked while `App.tsx` imports from it, so `tsc` would fail on a clean clone. Now tracked.
  - Three scaffold files under `build/ios/` lacked trailing newlines and failed the new gofmt gate. Formatted.
- Rollback boundary: `git revert` of the four authored commits restores the untouched `wails3 init` scaffold at `ebb3c8f`. Nothing outside this repository was changed.
- Known environmental noise: macOS linker prints `object file ... built for newer 'macOS' version (13.0) than being linked (12.0)` during `wails3 build`; benign, present on the untouched scaffold too.

### T1 remainder
- `gh repo create cags84/gui-go-lmstudio --public --source . --push` — name confirmed free on GitHub (`gh repo view` returns "Could not resolve"). Not yet run.

### T2..T7
- Not started.

## Forecast

Authored changed lines (generated template and lockfiles excluded): ~2 000 across T2..T7, ~250–400 per milestone. Each milestone is its own PR into `main` (`stacked-to-main`).

## Next step

Create the public GitHub repository and push, then start T2 (`internal/lms` boundary) with the fixtures already staged in `internal/lms/testdata/`.

## Discoveries (affect the design; recorded 2026-09-20)

- **Remote models exist.** `lms ls --json` returns a non-null `deviceIdentifier` for models served by a remote LM Link device, and then prefixes `path`/`indexedModelIdentifier` with `<deviceHash>:`. 7 of 11 entries on this machine were remote. The model table must group or badge local vs remote or it will look like duplicates.
- **Optional fields.** `paramsString`, `variants`, `selectedVariant`, `vision`, `trainedForToolUse` are absent on embedding entries; `deviceIdentifier` and `ttlMs` are nullable. Go structs need pointers / `omitempty`, not zero values.
- **`lms ps --json`** adds `identifier`, `ttlMs`, `lastUsedTime`, `contextLength`, `status`, `queued`, `parallel` on top of the `ls` fields, and it wakes the daemon as a side effect — an empty array is not proof that nothing is loaded.
- **REST is an enrichment, not a replacement.** While running, the server answers `GET /api/v1/models` (snake_case, includes `loaded_instances` and a `capabilities` block richer than the CLI) and the older `GET /api/v0/models` (`state`, `loaded_context_length`). Neither is reachable while the server is stopped, so lifecycle stays on `lms server status --json`. The two vocabularies disagree (`v1: "llm"` vs `v0: "vlm"` for the same vision model) and must not be mixed.
- **`go vet` / `go test` fail before the frontend is built**, because `main.go` embeds `all:frontend/dist`. Any CI job or local check must run `npm run build --prefix frontend` first.

## Fixtures captured

Real CLI output saved for `internal/lms/testdata/` in T2: `server_status_stopped.json`, `server_status_running.json`, `ps_empty.json`, `ps_one_llm.json`, `ls_all.json` (11 models, both types, local + remote), `ls_embedding.json`.
