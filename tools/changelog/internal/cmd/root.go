package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"changelog/internal/changelog"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "Changelog PoC CLI",
	}
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newGenerateBranchCmd())
	cmd.AddCommand(newGenerateMainCmd())
	return cmd
}

func newCreateCmd() *cobra.Command {
	var entriesDir string
	var repoRoot string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a changelog entry scaffold",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootAbs, err := filepath.Abs(repoRoot)
			if err != nil {
				return err
			}

			resolvedEntriesDir := changelog.ResolveEntriesDir(rootAbs, entriesDir)
			path, err := changelog.CreateEntryFile(resolvedEntriesDir, args[0])
			if err != nil {
				return err
			}

			fmt.Printf("created changelog entry scaffold at %s\n", path)
			return nil
		},
	}

	cmd.Flags().StringVar(&entriesDir, "entries-dir", "", "Directory containing changelog entries (defaults to .changelog under repo-root)")
	cmd.Flags().StringVar(&repoRoot, "repo-root", ".", "Path to git repository root")
	return cmd
}

func newValidateCmd() *cobra.Command {
	var entriesDir string
	var repoRoot string
	var allowedTypesCSV string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate changelog entry YAML files",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootAbs, err := filepath.Abs(repoRoot)
			if err != nil {
				return err
			}
			resolvedEntriesDir := changelog.ResolveEntriesDir(rootAbs, entriesDir)

			entries, err := changelog.LoadEntries(resolvedEntriesDir)
			if err != nil {
				return err
			}
			if err := changelog.ValidateEntryTypes(entries, changelog.ParseAllowedTypes(allowedTypesCSV)); err != nil {
				return err
			}

			fmt.Printf("validated %d entries in %s\n", len(entries), resolvedEntriesDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&entriesDir, "entries-dir", "", "Directory containing changelog entries (defaults to .changelog under repo-root)")
	cmd.Flags().StringVar(&repoRoot, "repo-root", ".", "Path to git repository root")
	cmd.Flags().StringVar(&allowedTypesCSV, "allowed-types", "", "Comma-separated allowed Type values (optional)")
	return cmd
}

func newGenerateBranchCmd() *cobra.Command {
	var entriesDir string
	var repoRoot string
	var version string
	var output string
	var allowedTypesCSV string

	cmd := &cobra.Command{
		Use:   "generate-branch",
		Short: "Generate changelog for current release branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootAbs, err := filepath.Abs(repoRoot)
			if err != nil {
				return err
			}
			resolvedEntriesDir := changelog.ResolveEntriesDir(rootAbs, entriesDir)

			resolvedVersion, err := changelog.ResolveVersion(rootAbs, version)
			if err != nil {
				return err
			}

			entries, err := changelog.LoadEntries(resolvedEntriesDir)
			if err != nil {
				return err
			}
			if err := changelog.ValidateEntryTypes(entries, changelog.ParseAllowedTypes(allowedTypesCSV)); err != nil {
				return err
			}

			content := changelog.RenderVersionSection(resolvedVersion, entries)
			if !filepath.IsAbs(output) {
				output = filepath.Join(rootAbs, output)
			}
			if err := changelog.EnsureParentDir(output); err != nil {
				return err
			}
			if err := os.WriteFile(output, []byte(content), 0o644); err != nil {
				return err
			}

			fmt.Printf("generated branch changelog for v%s at %s\n", resolvedVersion, output)
			return nil
		},
	}

	cmd.Flags().StringVar(&entriesDir, "entries-dir", "", "Directory containing changelog entries (defaults to .changelog under repo-root)")
	cmd.Flags().StringVar(&repoRoot, "repo-root", ".", "Path to git repository root")
	cmd.Flags().StringVar(&version, "version", "", "Release version (optional; inferred from branch if omitted)")
	cmd.Flags().StringVar(&output, "output", "CHANGELOG.md", "Output file path (absolute or relative to repo-root)")
	cmd.Flags().StringVar(&allowedTypesCSV, "allowed-types", "", "Comma-separated allowed Type values (optional)")
	return cmd
}

func newGenerateMainCmd() *cobra.Command {
	var entriesDir string
	var repoRoot string
	var targetBranch string
	var version string
	var output string
	var allowedTypesCSV string

	cmd := &cobra.Command{
		Use:   "generate-main",
		Short: "Generate accumulated changelog in target branch worktree",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootAbs, err := filepath.Abs(repoRoot)
			if err != nil {
				return err
			}
			resolvedEntriesDir := changelog.ResolveEntriesDir(rootAbs, entriesDir)

			resolvedVersion, err := changelog.ResolveVersion(rootAbs, version)
			if err != nil {
				return err
			}

			entries, err := changelog.LoadEntries(resolvedEntriesDir)
			if err != nil {
				return err
			}
			if err := changelog.ValidateEntryTypes(entries, changelog.ParseAllowedTypes(allowedTypesCSV)); err != nil {
				return err
			}

			worktreePath, cleanup, err := changelog.CreateTemporaryWorktree(rootAbs, targetBranch)
			if err != nil {
				return err
			}
			defer cleanup()

			if filepath.IsAbs(output) {
				return errors.New("--output must be a relative path for generate-main")
			}
			outputPath := filepath.Join(worktreePath, output)

			existing, err := os.ReadFile(outputPath)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}

			section := changelog.RenderVersionSection(resolvedVersion, entries)
			updated := changelog.UpsertVersionSection(string(existing), resolvedVersion, section)
			if err := changelog.EnsureParentDir(outputPath); err != nil {
				return err
			}
			if err := os.WriteFile(outputPath, []byte(updated), 0o644); err != nil {
				return err
			}

			fmt.Printf("generated accumulated changelog for v%s at %s (branch %s)\n", resolvedVersion, outputPath, targetBranch)
			return nil
		},
	}

	cmd.Flags().StringVar(&entriesDir, "entries-dir", "", "Directory containing changelog entries (defaults to .changelog under repo-root)")
	cmd.Flags().StringVar(&repoRoot, "repo-root", ".", "Path to git repository root")
	cmd.Flags().StringVar(&targetBranch, "target-branch", "main", "Branch where accumulated changelog will be written")
	cmd.Flags().StringVar(&version, "version", "", "Release version (optional; inferred from branch if omitted)")
	cmd.Flags().StringVar(&output, "output", "CHANGELOG.md", "Output file path inside target branch worktree")
	cmd.Flags().StringVar(&allowedTypesCSV, "allowed-types", "", "Comma-separated allowed Type values (optional)")
	return cmd
}
