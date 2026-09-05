# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`tick` is a small Go CLI (`github.com/cgalvisleon/tick`) for tracking a project's tasks with git-like ergonomics: `init`, `config`, `remote add/remove`, `push`/`pull`. Each project's data lives in a single SQLite file at `.tick/tick.db` in the project root, found by walking up from the cwd the same way git finds `.git`.

This directory is part of the `cgalvisleon` multi-repo workspace and is tied into the root `go.work` alongside `et` and `josefina` — see `/Users/cesargalvisleon/Projects/cgalvisleon/CLAUDE.md`. In particular, `tick` depends on `github.com/cgalvisleon/et` (the shared library) for its SQL layer (`et/jsql`) and terminal color helpers (`et/color`); since `et` is in the same workspace, local `et` changes are picked up live without bumping the `go.mod` version pin.

## Commands

Requires Go 1.25+.

```sh
go build -o tick ./cmd      # build the binary (entry point lives in cmd/, not the repo root)
go run ./cmd <args>              # run without building, e.g. go run ./cmd task
gofmt -w .                      # format (standard across the cgalvisleon workspace)
go vet ./...                    # static checks
```

`command.sh` wraps build/install: `./command.sh --build|--b`, `./command.sh --install|--i` (builds then `sudo cp`s the binary to `/usr/local/bin`), `./command.sh --help|--h`.

There are no tests in this repo yet.

Typical manual smoke test, from a scratch directory:

```sh
go run ./cmd init
go run ./cmd config user.name "Cesar"
go run ./cmd project code:P1 name:"My project"
go run ./cmd task code:T1 name:"First task" type:feature
go run ./cmd status code:T1 status:in_process description:"started"
go run ./cmd task code:T1 status
```

## Architecture

**Entry point (`cmd/main.go`)** — package `main`, just calls `tick.Root.Execute()` from `pkg/tick`. Kept separate from the `pkg/tick` Cobra package so `go build ./cmd` produces the binary while `pkg/tick` stays importable as a library package.

**Command layer (`pkg/tick/`)** — one Cobra command per file, all registered onto `tick.Root` (defined in `root.go`) via `init()`. `root.go` also defines `openStore()`, the shared helper every command except `init` uses to locate the nearest `.tick/` (via `internal/findroot`) and open its database.

- `args.go` — `parseKV` implements the `key:value` argument style used throughout (`project code:P1 name:Foo`, `task ID:x status`): splits args into a `map[string]string` plus a slice of bare (non key:value) words, preserving bare-word order. Commands then branch on the bare words (e.g. `task ID:x tag ...` vs `task ID:x status`).
- `project.go` / `task.go` — the two "resource" commands. Each follows the same shape: no args → list/show; `tag`/`tag remove` subcommand → manage tags; otherwise the kv pairs are applied as a create (task only, requires `code:`) or update.
- `status.go` — appends a status-history entry to a task and re-renders its history; status values are normalized via `store.NormalizeStatus` (accepts `in process`/`in-process`/`in_process` etc.).
- `remote.go` — remotes are just named filesystem paths stored in the project db (no network protocol).
- `push.go` / `pull.go` — **not a merge**: push/pull do a full raw-file copy of `tick.db` between the local project and a remote's path, overwriting the destination entirely. `sync.go` holds the shared helpers for this: `resolveRemote` (defaults to `"origin"`), `copyFile`, and `removeWALSidecars` (deletes stale `-wal`/`-shm` files before overwriting a db file, since a leftover WAL checksums against the old main file and can make a freshly-copied db look corrupt). `push` calls `db.Checkpoint()` first (`PRAGMA wal_checkpoint(TRUNCATE)`) so the copy captures every committed write.
- `config.go` — reads/writes arbitrary `key`/`value` pairs (`user.name`, `user.email`, `token`, ...) in the project's `.tick/tick.db` via `store.Config`. Since it's per-project (not global), it goes through `openStore()` like every other command; there's no longer a `~/.tick/config` file. `token` is a reserved key for a future `login` command to persist an auth token the same way any other setting is stored — no dedicated schema needed since `Config` is already generic key/value.

**`internal/findroot`** — walks up from cwd looking for a `.tick` directory, mirroring git's `.git` discovery.

**`internal/store`** — the persistence layer, built on `et/jsql` (SQLite driver `et/jsql/drivers/sqlite`). `db.go`'s `Open` connects and defines all four models (`Project`, `Task`, `Remote`, `Config`), each following the same `jsql.Def` → `db.Define` → `store.Init()` pattern seen in `project.go`/`task.go`/`remote.go`/`config.go`. Key model details:

- **Project**: exactly one row per `.tick/` — there's no lookup by id, `Get()` just fetches the sole row. Has a separate `project_tags` table (name/value pairs).
- **Task**: has its own `task_tags` table plus `task_status_history` (append-only log of status/description/percent/created_at). `SetStatus` implements pause-aware time tracking: `actual_start` is set on the first transition into `in_process`; `stop`/`await` are "pausing" statuses that get subtracted from the elapsed time via `paused_minutes` when the task leaves them; `actual_end`/`actual_minutes` are finalized on `done`. `AveragesByType` computes per-type average duration over `done` tasks on read (not cached), so it can't drift from the history.
- **Remote**: simple name→path table, unique on name.
- **Config**: simple name→value table, unique on name; backs the `config` command.

**`internal/ui`** — renders status pills and percent bars using `et/color` ANSI helpers, shared by `task.go`/`status.go`/`project.go` so all colored output stays consistent (pending=white, in_process=cyan, stop=red, await=yellow, done=green).

## Conventions to follow

- New commands needing project data should go through `openStore()` in `root.go`, not reimplement root/db discovery.
- Field-setting commands parse args with `parseKV` and treat bare words as subcommand routing (`tag`, `status`), matching the existing `project`/`task` commands — keep new commands consistent with this style rather than inventing flag-based syntax.
- Any code that overwrites a `tick.db` file wholesale (as `push`/`pull` do) must call `removeWALSidecars` first and, when writing from the live/local db, checkpoint via `db.Checkpoint()` beforehand.
