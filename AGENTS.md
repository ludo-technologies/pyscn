# Repository Guidelines

## Project Structure & Modules
- `cmd/pyqol/`: CLI entry (`main.go`).
- `internal/`: private packages (parser, analyzer, config, reporter, version).
- `service/` and `app/`: use cases, formatters, and services.
- `domain/`: core entities, interfaces, and errors.
- `e2e/`, `integration/`, `testdata/`: test suites and fixtures.
- `docs/`: architecture, standards, testing, branching.

## Build, Test, and Development
- `make build`: build binary to `./pyqol` (injects version info).
- `make test`: run all tests verbosely.
- `make coverage`: generate `coverage.html` and `coverage.out`.
- `make bench`: run benchmarks with `-benchmem`.
- `make fmt` / `make lint`: format with gofmt and run vet + golangci-lint.
- `make dev`: hot reload with `air` (installs if missing).
- Manual build: `go build -o pyqol ./cmd/pyqol`.

## Coding Style & Naming
- Language: Go 1.22+; format with `gofmt` and `go fmt`.
- Packages: lowercase single word (`internal/parser`).
- Files: lowercase with underscores; tests end with `_test.go`.
- Exported: PascalCase types/functions; locals camelCase; avoid one-letter names.
- Errors: lowercase messages, no trailing punctuation; wrap with context (`fmt.Errorf("parse file: %w", err)`).
- See `docs/CODING_STANDARDS.md` for details (imports grouping, tests, perf).

## Testing Guidelines
- Framework: standard `go test` with table-driven tests.
- Suites: unit (`*_test.go` in packages), `integration/`, `e2e/`.
- Targets: `go test ./...` for all; coverage: `go test -cover ./...` or `make coverage`.
- Benchmarks: `make bench` or `go test -bench=. ./internal/analyzer`.
- Build tags (when applicable): `//go:build integration` or `e2e`; run with `go test -tags=integration ./...`.

## Commit & Pull Requests
- Conventional commits (no scope): `feat: ...`, `fix: ...`, `docs: ...`, `refactor: ...`, `test: ...`, `perf: ...`, `style: ...`, `chore: ...`.
- Subject ≤ 50 chars, imperative, lowercase, no period. See `docs/COMMIT_CONVENTION.md`.
- Branches: `feature/issue-123-short-desc`, `fix/issue-456-bug`, `docs/...` (see `docs/BRANCHING.md`). No direct commits to `main`; squash-merge PRs.
- PRs: include description, checklist, and linked issue (`Closes #123`). Attach screenshots for UX/output changes. Ensure `make test`, `make lint`, and coverage pass.

## Security & Configuration
- Config discovery: nearest `.pyqol.yaml` → XDG (`~/.config/pyqol/`) → home (`~/.pyqol.yaml`).
- Example: set reports directory `output.directory: "reports"` (see `.pyqol.yaml` and `pyqol.yaml.example`).
- Do not commit secrets; keep fixtures in `testdata/` and large assets out of repo.

