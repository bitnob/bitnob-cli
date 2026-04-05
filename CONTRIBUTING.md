# Contributing to Bitnob CLI

Thanks for contributing to Bitnob CLI.

This guide covers how to set up locally, make safe changes, and open PRs that are easy to review.

## Development Setup

Prerequisites:

- Go 1.24+
- Git
- A Bitnob client ID and secret key (for integration/manual checks)

Clone and run:

```bash
git clone git@github.com:bitnob/bitnob-cli.git
cd bitnob-cli
go mod download
go run ./cmd/bitnob version
```

## Local Workflow

1. Create a branch from `main`.
2. Make focused changes (one concern per PR when possible).
3. Run formatting and tests before pushing.
4. Update docs when behavior or flags change.

Recommended commands:

```bash
make fmt
make test
go test ./...
```

Build and run locally:

```bash
make build
./bitnob version
```

## Testing Guidance

Before opening a PR:

- Run `go test ./...` and ensure all tests pass.
- Add or update tests for command behavior, API/client behavior, or config/auth behavior you changed.
- For staging checks, use the smoke script:

```bash
BITNOB_CLIENT_ID=your-client-id \
BITNOB_SECRET_KEY=your-secret-key \
./scripts/staging_smoke.sh
```

Use `RUN_WRITE=1` only when you intentionally want write-path checks.

## Code Expectations

- Keep changes backward compatible unless a breaking change is explicitly planned.
- Prefer small, composable functions over large command handlers.
- Return clear errors with actionable context.
- Avoid introducing new dependencies without strong need.
- Keep security-sensitive data out of logs and command output.

## CLI and DX Standards

When adding or changing commands:

- Keep command names and flags consistent with existing patterns.
- Provide helpful `--help` text and examples.
- Preserve machine-friendly output for scripts when possible.
- Ensure failure messages explain how users can recover.

## Documentation

Update [README.md](README.md) when you change:

- commands or flags
- install/upgrade instructions
- release behavior
- config/auth behavior

## Pull Requests

PRs should include:

- clear summary of what changed and why
- linked issue/context (if available)
- test evidence (`go test ./...` output summary)
- docs updates where relevant

Review checklist:

- tests pass
- no secret leakage
- no unrelated file churn
- release/install docs still accurate

## Commit Messages

Use short, imperative messages. Example:

- `Add config restore command`
- `Fix keyring fallback warning`
- `Update README install instructions`

## Release Notes

Releases are tag-driven via GitHub Actions. If your change impacts users directly, include release-note-ready text in your PR description so maintainers can summarize it quickly.
