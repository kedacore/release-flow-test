# Release Process

This document describes how to create and manage releases, including hotfix branches.

## Overview

Release automation uses **GitHub's native release notes** (`gh api .../releases/generate-notes` + `.github/release.yml`) to maintain a **draft release** on each push to `main` or a `release/v*` branch. The existing draft for that branch is deleted and recreated with an up-to-date changelog on every push.

Each branch gets its **own independent draft**, isolated by target branch:

| Branch | Draft tracks | Version bump |
|---|---|---|
| `main` | Next minor release | `vX.Y+1.0` |
| `release/v1` | Next patch release for v1 | `vX.Y.Z+1` |

---

## Creating a New Major/Minor Release (e.g. v1.0.0)

1. **Merge all PRs** for the release into `main`. Each PR must have exactly one `kind/` label — this determines which section of the release notes it appears in.

2. **Create the release branch** from the current `main` HEAD:
   ```bash
   git checkout -b release/v1 main
   git push origin release/v1
   ```
   > **Important**: create the branch BEFORE publishing the release so that the release can target the branch.

3. **Open the draft release** at [GitHub Releases](https://github.com/kedacore/release-flow-test/releases). The release workflow will have created a draft targeting `release/v1` automatically on the push in step 2.

4. **Edit the draft**:
   - Verify the tag and title match the intended version (auto-calculated as next minor, e.g. `v1.0.0`).
   - Verify **Target** is set to `release/v1` (not `main`).
   - Fill in the intro section (upgrade notes, highlights, link to docs, next release date).
   - Review the generated changelog for accuracy.

5. **Publish the release**. GitHub creates the `v1.0.0` tag at the `release/v1` HEAD.

   From this point, `main` accumulates changes for the next release (v1.1.0 or v2.0.0), and `release/v1` is used only for hotfixes.

---

## Creating a Hotfix Release (e.g. v1.0.1)

1. **Open a PR targeting `main`** with the fix. Apply a `kind/bug` label (or whichever category applies).

2. **Merge the PR** into `main`.

3. **Cherry-pick the fix to `release/v1`** — you can use the cherry-pick bot by commenting on the merged PR:
   ```
   /cherry-pick release/v1
   ```
   The bot creates a cherry-pick PR targeting `release/v1` automatically. You can also do it manually:
   ```bash
   git checkout release/v1
   git cherry-pick <commit-sha>
   git push origin release/v1
   ```
   > Only members of the `keda-e2e-test-executors` team can trigger the bot.

4. The release workflow regenerates the draft for `release/v1` on push. **Open the draft** targeting `release/v1`.

5. **Edit and publish** the draft targeting `release/v1`. The version is auto-calculated as the next patch (e.g. `v1.0.1`) — verify it before publishing.

---

## Continuing Development on main (e.g. v1.1.0)

After v1.0.0 is published, all PRs merged into `main` are tracked in a new draft automatically. When ready to release v1.1.0, repeat the steps in [Creating a New Major/Minor Release](#creating-a-new-majorminor-release-eg-v100), setting the tag to `v1.1.0`.

---

## Cherry-pick Bot

The cherry-pick bot automates backporting merged PRs to release branches.

**Trigger**: comment `/cherry-pick release/vX` on a merged PR (replacing `X` with the target version).

**What it does**:
- Verifies the commenter is a member of the `keda-e2e-test-executors` GitHub team
- Creates a branch `cherry-pick-<PR>-to-release-vX` and opens a PR targeting `release/vX`
- Copies all `kind/*` labels from the original PR so the cherry-pick PR also passes the label check
- Adds a `cherry-pick:vX` label to the original PR for traceability
- Idempotent: re-running the command updates the existing cherry-pick PR

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
