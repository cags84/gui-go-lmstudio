# Feature: lmstudio-desktop-controller

Repository-relative locator: `odd/tasks/lmstudio-desktop-controller.md`
Engram mirror topic: `odd/lmstudio-desktop-controller/tasks` (project `gui-go-lmstudio`)
Plan of record: `~/.claude/plans/pasted-content-id-4abf-crear-una-tidy-parnas.md` (approved 2026-09-20)

## Objective

Ship a cross-platform desktop app (macOS/Windows/Linux) in Go + Wails v3 + React/TS that controls the LM Studio local server through the `lms` CLI: start/stop, choose localhost vs LAN exposure, CORS, list/load/unload models, stream live logs, and show system info.

## Problem and why

Today the LM Studio server is driven from a terminal (`lms server start --port --bind --cors`, `lms ps`, `lms log stream`). The user wants the same power with a GUI that also visualizes what is happening, published as a public MIT repo (`cags84/gui-go-lmstudio`).

## Scope (authorized)

- This repository (Go module `github.com/cags84/gui-go-lmstudio`).
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
| T2 | `internal/lms`: Runner interface + exec impl, Client `ServerStatus`/`ListModels`/`LoadedModels` with JSON fixtures, FakeRunner; integration test skipped with `-short` | delegated writer | [x] |
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

- Native review (RDD is **on**, decided by global): `gentle-ai review assess` over `ebb3c8f..HEAD` returned **risk high**, `review_due=true`, reason `high_risk`, on two signals — process spawning in `build/ios/scripts/deps/install_deps.go` and shell scripting in `.github/workflows/ci.yml`. Consent was relayed to the user, who chose **granted**.
  - Outcome: **unavailable — terminal stop**. Three lenses (`review-risk`, `review-resilience`, `review-readability`) were admitted. `review-reliability` failed twice: first the model provider rejected the request (API safeguard error), then the reviewer payload came back as truncated, incomplete JSON. The refuter then ran and the transaction reached `correction_required`, but the next bound status returned `stop` with `corrupted_or_unverifiable_authority`, which the contract defines as terminal.
  - No correction ledger was ever readable, so no finding was applied or dismissed. No review receipt exists for this candidate, and none is claimed.
  - The failures began at the model provider, not in a Gentle AI contract, so no provider-defect report was filed.
  - Delivery therefore follows ordinary repository policy. The independent evidence for this work unit is the clean-clone reproduction of every CI step recorded above, not a review receipt.

### T1 remainder — done
- Published: https://github.com/cags84/gui-go-lmstudio (public, MIT), `main` tracking `origin/main`.
- Before publishing, tracked files were scanned for secrets, the maintainer's email and absolute home paths. One hit, an absolute project path in this document, was replaced with a repository-relative reference.
- **The first CI run failed on two of three runners**, which local checks could not have caught:
  - *Ubuntu*: the workflow installed `libwebkit2gtk-4.1-dev`, but Wails v3 builds on GTK4 + WebKitGTK 6.0 by default, so `pkg-config` could not resolve `gtk4` / `webkitgtk-6.0` and `go vet` failed. Fixed to `build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev`. The README carried the same wrong instruction and now documents the default GTK4 stack plus the legacy `-tags gtk3` path.
  - *Windows*: Go sources were checked out with CRLF, so `gofmt -l .` flagged every file in the repo, `main.go` included. Fixed with a `.gitattributes` pinning `*.go` to `eol=lf`.
  - Fix commit `452382f`. **Re-run 35519034211: ubuntu-latest, macos-latest and windows-latest all green.**
- Lesson recorded: a green local macOS check is not evidence about the other two runners. Cross-platform claims need the matrix run.

### T2 — `internal/lms` boundary
- Status: **done** (2026-09-20). Route: delegated writer (writer trigger: 9 new non-trivial files).
- Delivered: `runner.go` (Runner interface, `ExecRunner`, `RunError` carrying exit code and captured output, `ErrNotInstalled` sentinel, binary resolution PATH → `~/.lmstudio/bin/lms[.exe]`), `types.go` (domain types, `Model.IsRemote`), `wire.go` (JSON mirrors, epoch ms → `time.Time`/`time.Duration`), `client.go` (`ServerStatus`, `ListModels` with a typed filter enum, `LoadedModels`), plus `FakeRunner` and the six real fixtures.
- TDD: **strict, honored.** The writer reported an observed compile-failure RED for each of the three production groups before implementing. The integration test was added last against already-implemented code, so no RED applies to it; that was reported plainly rather than dressed up.
- Verification, re-run independently by the parent, not taken on the writer's word:
  - `go vet ./...` → PASS, no output
  - `go test ./internal/lms/ -v -count=1` → **39 cases PASS, 0 FAIL**
  - `go test ./... -short` → PASS; `TestServerStatus_Integration` SKIP confirmed under `-short` and PASS without it
  - `gofmt -l .` → empty
- Size: 1157 authored lines (405 production, 752 test), over the ~400-line advisory heuristic. Accepted: the required behavior list is genuinely large and trimming tests to hit a number is explicitly forbidden.
- Rollback boundary: `rm -rf internal/` removes this unit entirely; no other file, `go.mod` or `go.sum` was touched.

### Carried into T5 (found by experiment, not yet implemented)
- `Runner.Stream` relies on `exec.CommandContext`, which kills the child with SIGKILL on cancel. `lms log stream` loses its buffered tail under SIGKILL but flushes cleanly under SIGINT. T5 must set `cmd.Cancel` to send `os.Interrupt` with a `WaitDelay` fallback, and handle Windows separately, where `os.Interrupt` is unsupported.

### T3..T7
- Not started.

## Forecast

Authored changed lines (generated template and lockfiles excluded): ~2 000 across T2..T7, ~250–400 per milestone. Each milestone is its own PR into `main` (`stacked-to-main`).

## Next step

Create the public GitHub repository and push, then start T2 (`internal/lms` boundary) with the fixtures already staged in `internal/lms/testdata/`.

## Discoveries (affect the design; recorded 2026-09-20)

- **Remote models exist, and the only safe signal is `deviceIdentifier`.** A non-null `deviceIdentifier` means the model is served by a remote LM Link device; 7 of 11 entries on this machine are remote, so this is the common case. **Do not sniff the `<deviceHash>:` prefix.** It is inconsistent across commands: in `ls` a remote entry prefixes both `path` and `indexedModelIdentifier`, but in `ps` the same remote model has an unprefixed `path` and a prefixed `indexedModelIdentifier`. Prefix-sniffing silently misclassifies loaded remote models as local. The model table must badge local vs remote or it will look like duplicates.
- **Optional fields.** `paramsString`, `variants`, `selectedVariant`, `vision`, `trainedForToolUse` are absent on embedding entries; `deviceIdentifier` and `ttlMs` are nullable. Go structs need pointers / `omitempty`, not zero values.
- **`lms ps --json`** adds `identifier`, `ttlMs`, `lastUsedTime`, `contextLength`, `status`, `queued`, `parallel` on top of the `ls` fields, and it wakes the daemon as a side effect — an empty array is not proof that nothing is loaded.
- **REST is an enrichment, not a replacement.** While running, the server answers `GET /api/v1/models` (snake_case, includes `loaded_instances` and a `capabilities` block richer than the CLI) and the older `GET /api/v0/models` (`state`, `loaded_context_length`). Neither is reachable while the server is stopped, so lifecycle stays on `lms server status --json`. The two vocabularies disagree (`v1: "llm"` vs `v0: "vlm"` for the same vision model) and must not be mixed.
- **`go vet` / `go test` fail before the frontend is built**, because `main.go` embeds `all:frontend/dist`. Any CI job or local check must run `npm run build --prefix frontend` first.

## Fixtures captured

Real CLI output saved for `internal/lms/testdata/` in T2: `server_status_stopped.json`, `server_status_running.json`, `ps_empty.json`, `ps_one_llm.json`, `ls_all.json` (11 models, both types, local + remote), `ls_embedding.json`.

## Server exposure: how the UI can tell the truth (settled 2026-09-20)

`lms server status --json` returns only `{"running","port"}`. **There is no way to read the bind address back from the CLI**, and remembering what the app passed to `--bind` is wrong, because LM Studio's own GUI or a terminal can start the server independently.

Working detection, pure stdlib and cross-platform: dial the machine's own LAN address on the server port. Verified against a localhost-only server on port 1235:

| Target | `net.DialTimeout` result |
|---|---|
| `127.0.0.1:1235` | connects |
| `192.168.3.100:1235` (this host's LAN IP) | refused |

A server bound to `0.0.0.0` accepts both. T3 uses this probe for the exposure badge instead of trusting remembered settings.

Related: `lms server start` against an already-running server exits 0 and reprints its success line, so a Start button is safe to press twice.

## Log stream: verified behavior for T5

- Use `-s server`. The **default source is `model`**, which emitted nothing even for a real chat completion, because model input/output logging depends on LM Studio's own prompt-logging setting.
- The stream prints a plain-text banner (`Streaming logs from LM Studio`) and a blank line **before** any JSON. The NDJSON reader must skip non-JSON lines rather than error on them.
- Line schema: `{"timestamp":<epoch ms>,"data":{"type":"server.log","content":"<text>","level":"debug"|"info"}}`.
- It is genuinely **live** when piped, not block-buffered: a request fired at 10:27:31.731 produced its first log line at 10:27:31.750, about 19 ms later. A `bufio.Scanner` over the stdout pipe is enough; no PTY is needed.
- Fixture captured: `log_stream_server.ndjson` (7 JSON lines plus the banner), held in the session scratchpad until T5 needs it.
