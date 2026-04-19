/*
Copyright © 2025 srz_zumix
*/
package gist

import (
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
			ctx := cmd.Context()
			srcClient, dstClient, err := newClientPair(ctx, src, dst, srcToken, dstToken)
			if err != nil {
				return err
			}

			gistIDs, err := resolveGistIDs(ctx, srcClient, args)
			if err != nil {
				return err
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
		},
	}

	f := cmd.Flags()
	f.StringVarP(&src, "src", "s", "", "Source GitHub host (default: current host from gh auth)")
	f.StringVarP(&dst, "dst", "d", "", "Destination GitHub host (default: current host from gh auth)")
	f.StringVar(&srcToken, "src-token", "", "Token for the source GitHub host")
	f.StringVar(&dstToken, "dst-token", "", "Token for the destination GitHub host")
	f.BoolVarP(&dryrun, "dryrun", "n", false, "Dry run: show what would be copied without making changes")

	return cmd
}
