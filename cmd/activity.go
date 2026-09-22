/*
Copyright © 2025 srz_zumix
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-my-kit/pkg/activity"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
	"github.com/srz-zumix/go-gh-extension/pkg/render"
)

func NewActivityCmd() *cobra.Command {
	var (
		host      string
		output    string
		mode      string
		period    string
		since     string
		until     string
		include   []string
		exclude   []string
		skipEmpty bool
		force     bool
		exporter  cmdutil.Exporter
	)

	cmd := &cobra.Command{
		Use:   "activity [user]",
		Short: "Dump a user's GitHub activity",
		Long: `Collect a GitHub user's activity (profile, follows, organizations, events,
contributions, pull requests, issues, reviews, comments, repositories, starred
and watched repositories, gists, packages, projects, discussions, and
notifications) and write it to a directory as Markdown and JSON files.

If no user is given, the authenticated user is used. Notifications, private
events, and watched repositories are only available for the authenticated
user and are skipped with a warning otherwise.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			logger.Info("connecting to GitHub", "host", host)
			g, err := gh.NewGitHubClientWithRepo(repository.Repository{Host: host})
			if err != nil {
				return fmt.Errorf("failed to create GitHub client: %w", err)
			}

			username := ""
			if len(args) > 0 {
				username = args[0]
			}
			self := username == ""

			loginUser, err := gh.GetLoginUser(ctx, g)
			if err != nil {
				return fmt.Errorf("failed to get authenticated user: %w", err)
			}
			if username == "" {
				username = loginUser.GetLogin()
			} else if username == loginUser.GetLogin() {
				self = true
			}

			since, until, err := activity.ResolvePeriod(period, since, until, time.Now())
			if err != nil {
				return fmt.Errorf("failed to resolve period: %w", err)
			}

			kinds, err := activity.ResolveKinds(include, exclude)
			if err != nil {
				return fmt.Errorf("failed to resolve activity kinds: %w", err)
			}

			opts := activity.Options{
				Username: username,
				Self:     self,
				Since:    since,
				Until:    until,
				Kinds:    kinds,
			}

			dir := output
			if dir == "" {
				dir = fmt.Sprintf("./activity-%s-%s", username, time.Now().Format("20060102"))
			}
			if exporter == nil {
				if err := activity.CheckOutputDir(dir, force); err != nil {
					return fmt.Errorf("cannot write activity output to '%s': %w", dir, err)
				}
			}

			logger.Info("starting activity dump", "user", username, "self", self, "mode", mode)

			result, err := activity.Collect(ctx, g, opts)
			if err != nil {
				return fmt.Errorf("failed to collect activity for '%s': %w", username, err)
			}

			if exporter != nil {
				logger.Info("rendering result as JSON")
				return render.NewRenderer(exporter).RenderExportedData(result)
			}

			writeOpts := activity.WriteOptions{
				Dir:       dir,
				Mode:      activity.Mode(mode),
				Kinds:     kinds,
				SkipEmpty: skipEmpty,
				Force:     force,
			}
			if err := activity.Write(result, writeOpts); err != nil {
				return fmt.Errorf("failed to write activity output to '%s': %w", dir, err)
			}
			logger.Info("activity dump complete", "dir", dir)

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&host, "host", "", "GitHub host to query (default: current host from gh auth)")
	f.StringVarP(&output, "output", "o", "", "Output directory (default: ./activity-<user>-<YYYYMMDD>)")
	cmdutil.StringEnumFlag(cmd, &mode, "mode", "", string(activity.ModeDetail), activity.ModeList, "Output detail level")
	f.StringVar(&period, "period", "", "Period to collect: <N>d|w|m|y (e.g. 30d, 4w, 6m, 1y) or a fiscal year FY<YY>[H1|H2|Q1..Q4] (e.g. FY26, FY26H1, FY26Q1); default 30d; mutually exclusive with --since/--until")
	f.StringVar(&since, "since", "", "Collect activity since this date (RFC3339 or YYYY-MM-DD; mutually exclusive with --period)")
	f.StringVar(&until, "until", "", "Collect activity until this date (RFC3339 or YYYY-MM-DD; mutually exclusive with --period)")
	f.StringSliceVar(&include, "include", nil, "Only collect these activity kinds (default: all; mutually exclusive with --exclude)")
	f.StringSliceVar(&exclude, "exclude", nil, "Exclude these activity kinds (mutually exclusive with --include)")
	f.BoolVar(&skipEmpty, "skip-empty", false, "Do not write output files for activity kinds with no collected data")
	f.BoolVar(&force, "force", false, "Empty the output directory before writing instead of failing when it is not empty")
	cmdutil.AddFormatFlags(cmd, &exporter)

	return cmd
}

func init() {
	rootCmd.AddCommand(NewActivityCmd())
}
