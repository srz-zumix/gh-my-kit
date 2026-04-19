/*
Copyright © 2025 srz_zumix
*/
package gist

import (
	"fmt"
	"strings"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-my-kit/pkg/gist"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

func NewConvertCmd() *cobra.Command {
	var (
		src            string
		srcToken       string
		dst            string
		dstToken       string
		name           string
		owner          string
		visibility     string
		dryrun         bool
		noRenameBranch bool
	)

	cmd := &cobra.Command{
		Use:   "convert <gist-id...>",
		Short: "Convert gists to regular repositories",
		Long: `Convert one or more gists to regular GitHub repositories,
preserving the full git history via git clone --mirror + git push --mirror.

By default the repository name is derived from the gist description.
If the description is empty, the gist ID is used as the repository name.
The repository visibility defaults to the gist's own visibility (public/private).

Examples:
  # Convert a gist to a repository (name derived from gist description)
  gh my-kit gist convert abc123

  # Convert a gist with a specific repository name
  gh my-kit gist convert abc123 --name my-repo

  # Convert a gist and create the repository under an organization
  gh my-kit gist convert abc123 --owner my-org

  # Convert multiple gists (names derived automatically)
  gh my-kit gist convert abc123 def456

  # Convert a gist to a repository on a GHES instance
  gh my-kit gist convert abc123 --dst ghes.example.com --dst-token <token>

  # Dry run: show what would be created without making changes
  gh my-kit gist convert abc123 --dryrun

  # Convert without renaming default branch from master to main
  gh my-kit gist convert abc123 --no-rename-branch`,

		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if name != "" && len(args) > 1 {
				return fmt.Errorf("--name can only be used with a single gist ID")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if strings.Contains(owner, "/") {
				return fmt.Errorf("--owner must be a plain owner name without '/': %q", owner)
			}
			srcClient, err := newClientForHost(src, srcToken)
			if err != nil {
				return fmt.Errorf("failed to create source client: %w", err)
			}
			dstClient, err := newClientForHost(dst, dstToken)
			if err != nil {
				return fmt.Errorf("failed to create destination client: %w", err)
			}

			var converted, failed int
			for _, id := range args {
				repoName := name
				if dryrun {
					if repoName == "" {
						repoName = "(derived from gist)"
					}
					logger.Info("[dryrun] would convert", "gist", id, "repo", repoName)
					converted++
					continue
				}

				opts := gist.ConvertGistToRepoOptions{
					RepoName:           repoName,
					OrgName:            owner,
					Visibility:         visibility,
					RenameMasterToMain: !noRenameBranch,
				}
				repo, err := gist.ConvertGistToRepo(ctx, srcClient, dstClient, id, opts)
				if err != nil {
					logger.Error("failed to convert gist", "gist", id, "error", err)
					failed++
					continue
				}
				logger.Info("converted", "gist", id, "repo", repo.GetFullName(), "url", repo.GetHTMLURL())
				converted++
			}

			logger.Info("done", "converted", converted, "failed", failed)
			if failed > 0 {
				return fmt.Errorf("%d gist(s) failed to convert", failed)
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&src, "src", "s", "", "Source GitHub host where the gist resides (default: current host from gh auth)")
	f.StringVar(&srcToken, "src-token", "", "Token for the source GitHub host")
	f.StringVarP(&dst, "dst", "d", "", "Destination GitHub host where the repository will be created (default: current host from gh auth)")
	f.StringVar(&dstToken, "dst-token", "", "Token for the destination GitHub host")
	f.StringVar(&name, "name", "", "Repository name (only valid with a single gist ID; default: derived from gist description)")
	f.StringVarP(&owner, "owner", "o", "", "Organization to create the repository under (default: authenticated user)")
	cmdutil.StringEnumFlag(cmd, &visibility, "visibility", "v", "", gh.RepoVisibilityList, "Visibility of the created repository (default: inherit from gist)")
	f.BoolVarP(&dryrun, "dryrun", "n", false, "Dry run: show what would be converted without making changes")
	f.BoolVar(&noRenameBranch, "no-rename-branch", false, "Disable renaming the default branch from master to main")

	return cmd
}
