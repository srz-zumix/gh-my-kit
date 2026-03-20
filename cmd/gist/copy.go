/*
Copyright © 2025 srz_zumix
*/
package gist

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

func NewCopyCmd() *cobra.Command {
	var (
		src      string
		dst      string
		srcToken string
		dstToken string
		dryrun   bool
	)

	cmd := &cobra.Command{
		Use:   "copy [gist-id...]",
		Short: "Copy gists from a source to a destination (latest content only)",
		Long: `Copy gists from a source GitHub host to a destination GitHub host.

Only the latest file content is copied; git history is not preserved.
Use "gist migrate" to preserve the full git history.

If no gist IDs are provided, all gists of the authenticated source user are
copied. Otherwise only the specified gist IDs are copied.

Examples:
  # Copy all gists from github.com to a GHES instance
  gh my-kit gist copy --dst ghes.example.com --dst-token <token>

  # Copy specific gists
  gh my-kit gist copy abc123 def456 --dst ghes.example.com --dst-token <token>

  # Copy between two GHES instances
  gh my-kit gist copy \
    --src src.example.com --src-token <src-token> \
    --dst dst.example.com --dst-token <dst-token>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCopy(args, src, dst, srcToken, dstToken, dryrun)
		},
	}

	cmd.Flags().StringVarP(&src, "src", "s", "", "Source GitHub host (default: current host from gh auth)")
	cmd.Flags().StringVarP(&dst, "dst", "d", "", "Destination GitHub host (default: current host from gh auth)")
	cmd.Flags().StringVar(&srcToken, "src-token", "", "Token for the source GitHub host")
	cmd.Flags().StringVar(&dstToken, "dst-token", "", "Token for the destination GitHub host")
	cmd.Flags().BoolVarP(&dryrun, "dryrun", "n", false, "Dry run: show what would be copied without making changes")

	return cmd
}

func runCopy(args []string, src, dst, srcToken, dstToken string, dryrun bool) error {
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

	var copied, failed int
	for _, id := range gistIDs {
		if dryrun {
			logger.Info("[dryrun] would copy", "id", id)
			copied++
			continue
		}
		created, err := gh.CopyGist(ctx, srcClient, dstClient, id)
		if err != nil {
			logger.Error("failed to copy gist", "id", id, "error", err)
			failed++
			continue
		}
		logger.Info("copied", "src", id, "dst", created.GetID())
		copied++
	}

	logger.Info("done", "copied", copied, "failed", failed)
	if failed > 0 {
		return fmt.Errorf("%d gist(s) failed to copy", failed)
	}
	return nil
}
