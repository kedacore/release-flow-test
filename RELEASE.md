# Release Process

This document describes how to create and manage releases, including hotfix branches.

## Overview

Release automation uses **GitHub's native release notes** (`gh api .../releases/generate-notes` + `.github/release.yml`) to maintain a **draft release** on each push to `main` or a `release/v*` branch. The existing draft for that branch is deleted and recreated with an up-to-date changelog on every push.

Each branch gets its **own independent draft**, isolated by target branch:

| Branch | Draft tracks | Version bump |
|---|---|---|
| `main` | Next minor release | `vX.Y+1.0` |
| `release/vX.Y` | Next patch release for that minor line | `vX.Y.Z+1` |

---

## Creating a New Major/Minor Release (e.g. v1.0.0)

1. **Merge all PRs** for the release into `main`. Each PR must have exactly one `kind/` label — this determines which section of the release notes it appears in.

2. **Publish `vX.Y.0`** (for example, `v1.0.0`). The tag workflow automatically creates the release branch `release/vX.Y` from that tag commit.

3. **Open the draft release** at [GitHub Releases](https://github.com/kedacore/release-flow-test/releases). The release workflow will create a draft for the new release branch after it is created.

4. **Edit the draft**:
   - Verify the tag and title match the intended version (auto-calculated as next minor, e.g. `v1.0.0`).
   - Verify **Target** is set to the new release branch (for example `release/v1.0`) and not `main`.
   - Fill in the intro section (upgrade notes, highlights, link to docs, next release date).
   - Review the generated changelog for accuracy.

5. **Publish the release**. GitHub keeps the published tag and future hotfixes should be done on the generated release branch.

   From this point, `main` accumulates changes for the next release (v1.1.0 or v2.0.0), and `release/vX.Y` is used only for hotfixes.

---

## Creating a Hotfix Release (e.g. v1.0.1)

1. **Open a PR targeting `main`** with the fix. Apply a `kind/bug` label (or whichever category applies).

2. **Merge the PR** into `main`.

3. **Cherry-pick the fix to the matching release branch (`release/vX.Y`)** — you can use the cherry-pick bot by adding a trigger label to the merged PR:
   ```
   cherry-pick/v1.0
   ```
   The bot creates a cherry-pick PR targeting that release branch automatically and then swaps the trigger label for a confirmation label (`cherry-picked/v1.0`). You can also do it manually:
   ```bash
   git checkout release/v1.0
   git cherry-pick <commit-sha>
   git push origin release/v1.0
   ```


4. The release workflow regenerates the draft for the release branch on push. **Open the draft** targeting `release/vX.Y`.

5. **Edit and publish** the draft targeting `release/vX.Y`. The version is auto-calculated as the next patch (e.g. `v1.0.1`) — verify it before publishing.

---

## Continuing Development on main (e.g. v1.1.0)

After v1.0.0 is published, all PRs merged into `main` are tracked in a new draft automatically. When ready to release v1.1.0, repeat the steps in [Creating a New Major/Minor Release](#creating-a-new-majorminor-release-eg-v100), setting the tag to `v1.1.0`.

---

## Cherry-pick Bot

The cherry-pick bot automates backporting merged PRs to release branches.

**Trigger**: add label `cherry-pick/vX.Y` on a merged PR.

**What it does**:
- Creates a branch `cherry-pick-<PR>-to-release-vX-Y` and opens a PR targeting `release/vX.Y`
- Copies all `kind/*` labels from the original PR so the cherry-pick PR also passes the label check
- On success, replaces `cherry-pick/vX.Y` with `cherry-picked/vX.Y` on the original PR
- Idempotent: re-running the workflow updates the existing cherry-pick PR

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

The `Lint PR / Validate PR Metadata` check enforces this and blocks merge if no valid label is present.

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

The same `Lint PR / Validate PR Metadata` check enforces this title format and re-runs when a PR title is edited.
