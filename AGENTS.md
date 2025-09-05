# Repository Guidelines

## Project Structure & Module Organization
- `cmd/pyqol/`: CLI entry (`main.go`).
- `internal/`: private packages (parser, analyzer, config, reporter, version).
- `service/`, `app/`: use cases, formatters, services.
- `domain/`: core entities, interfaces, errors.
- `e2e/`, `integration/`, `testdata/`: test suites and fixtures.
- `docs/`: architecture, standards, testing, branching.

## Build, Test, and Development Commands
- `make build`: build binary to `./pyqol` (injects version info).
- `make test`: run all tests verbosely (`go test ./...`).
- `make coverage`: create `coverage.out` + `coverage.html`.
- `make bench`: run benchmarks with `-benchmem`.
- `make fmt` / `make lint`: format with `gofmt`; run `go vet` + `golangci-lint`.
- `make dev`: hot reload with `air` (auto-installs if missing).
- Manual: `go build -o pyqol ./cmd/pyqol`.

## Coding Style & Naming Conventions
- Language: Go 1.22+; format with `gofmt` / `go fmt`.
- Packages/files: lowercase single word; tests end with `_test.go`.
- Exports: PascalCase; locals: camelCase; avoid one-letter names.
- Errors: lowercase messages, no trailing punctuation; wrap with context
  (e.g., `fmt.Errorf("parse file: %w", err)`).
- See `docs/CODING_STANDARDS.md` for imports grouping and examples.

## Testing Guidelines
- Framework: standard `go test`; table-driven tests preferred.
- Run all: `go test ./...`; coverage: `go test -cover ./...` or `make coverage`.
- Benchmarks: `make bench` or `go test -bench=. ./internal/analyzer`.
- Build tags: `//go:build integration` or `e2e`; run with
  `go test -tags=integration ./...`.
- Name tests `*_test.go`; keep fixtures in `testdata/`.

## Commit & Pull Request Guidelines
- Commits: Conventional commits (no scope). Example: `feat: add analyzer rule`.
- Subject ≤ 50 chars, imperative, lowercase, no period.
- Branches: `feature/issue-123-short-desc`, `fix/issue-456-bug`, `docs/...`.
- PRs: include description, checklist, linked issue (e.g., `Closes #123`), and
  screenshots for UX/output changes; ensure `make test`, `make lint`, coverage pass.

## Security & Configuration Tips
- Config discovery: nearest `.pyqol.yaml` → XDG `~/.config/pyqol/` → `~/.pyqol.yaml`.
- Example: set output dir via `output.directory: "reports"`.
- Do not commit secrets; keep large assets out of repo.

