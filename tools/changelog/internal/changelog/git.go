package changelog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var releaseBranchPattern = regexp.MustCompile(`^release/v([0-9]+\.[0-9]+(?:\.[0-9]+)?)$`)

func ResolveVersion(repoRoot, explicitVersion string) (string, error) {
	explicitVersion = strings.TrimSpace(explicitVersion)
	if explicitVersion != "" {
		return strings.TrimPrefix(explicitVersion, "v"), nil
	}

	branch, err := currentBranch(repoRoot)
	if err != nil {
		return "", err
	}

	matches := releaseBranchPattern.FindStringSubmatch(branch)
	if len(matches) != 2 {
		return "", fmt.Errorf("--version not provided and current branch %q is not release/vX.Y or release/vX.Y.Z", branch)
	}
	return matches[1], nil
}

func CreateTemporaryWorktree(repoRoot, targetBranch string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "changelog-worktree-")
	if err != nil {
		return "", nil, err
	}

	absTmp, err := filepath.Abs(tmpDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", nil, err
	}

	if _, err := runGit(repoRoot, "worktree", "add", "--quiet", absTmp, targetBranch); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", nil, fmt.Errorf("creating worktree for branch %q: %w", targetBranch, err)
	}

	cleanup := func() {
		_, _ = runGit(repoRoot, "worktree", "remove", "--force", absTmp)
		_ = os.RemoveAll(absTmp)
	}

	return absTmp, cleanup, nil
}

func currentBranch(repoRoot string) (string, error) {
	out, err := runGit(repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func runGit(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return string(out), nil
}
