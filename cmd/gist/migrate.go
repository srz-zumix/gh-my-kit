/*
Copyright © 2025 srz_zumix
*/
package gist

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

func NewMigrateCmd() *cobra.Command {
	var (
		src      string
		dst      string
		srcToken string
		dstToken string
		dryrun   bool
	)

	cmd := &cobra.Command{
		Use:   "migrate [gist-id...]",
		Short: "Migrate gists preserving full git history",
		Long: `Migrate gists from a source GitHub host to a destination GitHub host,
preserving the full git history via git clone --mirror + git push --mirror.

If no gist IDs are provided, all gists of the authenticated source user are
migrated. Otherwise only the specified gist IDs are migrated.

Use "gist copy" instead if you only need the latest file content without history.

Examples:
  # Migrate all gists from github.com to a GHES instance
  gh my-kit gist migrate --dst ghes.example.com --dst-token <token>

  # Migrate specific gists
  gh my-kit gist migrate abc123 def456 --dst ghes.example.com --dst-token <token>

  # Migrate between two GHES instances
  gh my-kit gist migrate \
    --src src.example.com --src-token <src-token> \
    --dst dst.example.com --dst-token <dst-token>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(args, src, dst, srcToken, dstToken, dryrun)
		},
	}

	cmd.Flags().StringVarP(&src, "src", "s", "", "Source GitHub host (default: current host from gh auth)")
	cmd.Flags().StringVarP(&dst, "dst", "d", "", "Destination GitHub host (default: current host from gh auth)")
	cmd.Flags().StringVar(&srcToken, "src-token", "", "Token for the source GitHub host")
	cmd.Flags().StringVar(&dstToken, "dst-token", "", "Token for the destination GitHub host")
	cmd.Flags().BoolVarP(&dryrun, "dryrun", "n", false, "Dry run: show what would be migrated without making changes")

	return cmd
}

func newClientForHost(host, token string) (*gh.GitHubClient, error) {
	repo := repository.Repository{Host: host}
	if token != "" {
		return gh.NewGitHubClientWithToken(repo, token)
	}
	return gh.NewGitHubClientWithRepo(repo)
}

func runMigrate(args []string, src, dst, srcToken, dstToken string, dryrun bool) error {
	ctx := context.Background()

	srcClient, err := newClientForHost(src, srcToken)
	if err != nil {
		return fmt.Errorf("failed to create source client: %w", err)
	}

	dstClient, err := newClientForHost(dst, dstToken)
	if err != nil {
		return fmt.Errorf("failed to create destination client: %w", err)
	}

	srcUser, err := gh.GetLoginUser(ctx, srcClient)
	if err != nil {
		return fmt.Errorf("failed to get source user: %w", err)
	}
	dstUser, err := gh.GetLoginUser(ctx, dstClient)
	if err != nil {
		return fmt.Errorf("failed to get destination user: %w", err)
	}
	if srcClient.Host() == dstClient.Host() && srcUser.GetLogin() == dstUser.GetLogin() {
		return fmt.Errorf("source and destination user must be different (%s@%s)", srcUser.GetLogin(), srcClient.Host())
	}

	gistIDs := args
	if len(gistIDs) == 0 {
		gists, err := gh.ListGists(ctx, srcClient)
		if err != nil {
			return fmt.Errorf("failed to list gists: %w", err)
		}
		for _, g := range gists {
			gistIDs = append(gistIDs, g.GetID())
		}
	}

	var migrated, failed int
	for _, id := range gistIDs {
		if dryrun {
			logger.Info("[dryrun] would migrate", "id", id)
			migrated++
			continue
		}
		created, err := gh.MigrateGist(ctx, srcClient, dstClient, id)
		if err != nil {
			logger.Error("failed to migrate gist", "id", id, "error", err)
			failed++
			continue
		}
		logger.Info("migrated", "src", id, "dst", created.GetID())
		migrated++
	}

	logger.Info("done", "migrated", migrated, "failed", failed)
	if failed > 0 {
		return fmt.Errorf("%d gist(s) failed to migrate", failed)
	}
	return nil
}
