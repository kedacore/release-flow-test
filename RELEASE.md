# Release Process

This document describes how to create and manage releases, including hotfix branches.

## Overview

Release Drafter automatically maintains a **draft release** that accumulates all merged PRs since the last published release. When a PR is merged into `main` or a `release/v*` branch, the corresponding draft is updated.

Each branch gets its **own independent draft**, isolated by `filter-by-commitish`. The suggested version is calculated automatically on each push:

| Branch | Draft tracks | Version bump |
|---|---|---|
| `main` | Next minor release | `vX.Y+1.0` |
| `release/v1` | Next patch release for v1 | `vX.Y.Z+1` |

---

## Creating a New Major/Minor Release (e.g. v1.0.0)

1. **Merge all PRs** for the release into `main`. Each PR must have exactly one `kind/` label — this determines which section of the release notes it appears in.

2. **Open the draft release** at [GitHub Releases](https://github.com/kedacore/release-flow-test/releases) (marked as *Draft*, targeting `main`).

3. **Edit the draft**:
   - Verify the tag and title match the intended version (auto-calculated as next minor, e.g. `v1.1.0`).
   - Fill in the intro section (upgrade notes, highlights, link to docs, next release date).
   - Review the generated changelog for accuracy.

4. **Publish the release**. GitHub creates the `v1.0.0` tag at the current `main` HEAD.

5. **Create the release branch** from the published tag:
   ```bash
   git checkout -b release/v1 v1.0.0
   git push origin release/v1
   ```

   From this point, `main` accumulates changes for the next release (v1.1.0 or v2.0.0), and `release/v1` is used only for hotfixes.

---

## Creating a Hotfix Release (e.g. v1.0.1)

1. **Open a PR targeting `main`** with the fix. Apply a `kind/bug` label (or whichever category applies).

2. **Merge the PR** into `main`.

3. **Cherry-pick the fix to `release/v1`**:
   ```bash
   git checkout release/v1
   git cherry-pick <commit-sha>
   git push origin release/v1
   ```

4. Release Drafter updates the draft for `release/v1` on push. **Open the draft** targeting `release/v1`.

5. **Edit and publish** the draft targeting `release/v1`. The version is auto-calculated as the next patch (e.g. `v1.0.1`) — verify it before publishing.

---

## Continuing Development on main (e.g. v1.1.0)

After v1.0.0 is published, all PRs merged into `main` are tracked in a new draft automatically. When ready to release v1.1.0, repeat the steps in [Creating a New Major/Minor Release](#creating-a-new-majorminor-release-eg-v100), setting the tag to `v1.1.0`.

---

## PR Requirements

Every PR must have **exactly one** of the following labels before it can be merged. The label controls which section the PR appears in within the release notes:

| Label | Release notes section |
|---|---|
| `kind/breaking-change` | Breaking Changes |
| `kind/feature` | New |
| `kind/new-scaler` | New |
| `kind/improvement` | Improvements |
| `kind/enhancement` | Improvements |
| `kind/bug` | Fixes |
| `kind/deprecation` | Deprecations |
| `kind/chore` | Other |
| `kind/documentation` | Other |
| `kind/dependencies` | Other |
| `kind/ci` | Other |
| `skip-changelog` | *(excluded from release notes)* |

The `Lint PR / Validate PR Labels` check enforces this and blocks merge if no valid label is present.

---

## PR Title Convention

To produce release notes consistent with KEDA's changelog format, PR titles should follow this pattern:

```
**ComponentName**: Brief description of the change
```

Examples:
```
**General**: Fix nil pointer dereference in fallback
**Kafka Scaler**: Add support for SASL/OAuth bearer authentication
**CI**: Replace stale bot with GitHub Actions stale action
```

This makes each entry in the release notes render as:

```
- **Kafka Scaler**: Add support for SASL/OAuth bearer authentication (#42)
```
