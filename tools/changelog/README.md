# Changelog PoC Tool

This directory contains a PoC changelog tool inspired by OTel chloggen workflows.

## Entry files

The tool expects YAML files in a directory passed by flag.

- File name can be any value, as long as extension is `.yaml` or `.yml`
- Required fields:
  - `Description`
  - `Type`
- At least one reference is required:
  - `Issues`, or
  - `PRs` / `PullRequests`
- Optional fields:
  - `Component` (single string)
  - `Subcomponent` (single string)
  - `Issues` (list of strings)
  - `PRs` (list of strings)

Example:

```yaml
Description: Add new scaler behavior
Type: feat
Component: scaler
Subcomponent: operator
Issues:
  - "#123"
```

## Commands

Create a new entry scaffold in `.changelog` by default:

```bash
go run . create 1928
```

Or target another directory explicitly:

```bash
go run . create 1928 --entries-dir /path/to/entries --repo-root /path/to/repo
```

Run from this directory:

```bash
go run . validate
```

```bash
go run . generate-branch --repo-root /path/to/repo
```

```bash
go run . generate-main --repo-root /path/to/repo --target-branch main
```

Notes:

- `generate-branch` writes a changelog for the current branch/version.
- `generate-main` uses a temporary `git worktree` for the target branch and writes there.
- Both commands can receive `--version` to override branch-based version inference.

## Makefile helpers

From repository root:

```bash
make changelog-create CHANGELOG_ENTRY_NAME=1928
```

```bash
make changelog-validate CHANGELOG_ENTRIES_DIR=./path/to/entries
```

```bash
make changelog-generate-branch CHANGELOG_ENTRIES_DIR=./path/to/entries CHANGELOG_OUTPUT=CHANGELOG.md
```

```bash
make changelog-generate-main CHANGELOG_ENTRIES_DIR=./path/to/entries CHANGELOG_TARGET_BRANCH=main CHANGELOG_OUTPUT=CHANGELOG.md
```
