/*
Copyright © 2025 srz_zumix
*/
package profile

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-my-kit/cmd/common"
	"github.com/srz-zumix/gh-my-kit/pkg/profile"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

func NewCopyCmd() *cobra.Command {
	var (
		src      string
		dst      string
		srcToken string
		dstToken string
		fields   []string
		dryrun   bool
	)

	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy a user profile from a source to a destination host",
		Long: `Copy the authenticated source user's profile to the authenticated
destination user.

Fields left empty on the source are skipped, preserving the destination's
existing value. Social accounts present on the source but missing on the
destination are added; existing destination social accounts are left as-is.

Examples:
  # Copy all profile fields and social accounts to a GHES instance
  gh my-kit profile copy --dst ghes.example.com --dst-token <token>

  # Copy only name and bio
  gh my-kit profile copy --dst ghes.example.com --dst-token <token> --fields name,bio

  # Copy between two GHES instances
  gh my-kit profile copy \
    --src src.example.com --src-token <src-token> \
    --dst dst.example.com --dst-token <dst-token>

  # Dry run: show what would be copied without making changes
  gh my-kit profile copy --dst ghes.example.com --dst-token <token> --dryrun`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			parsedFields, err := profile.ParseFields(fields)
			if err != nil {
				return fmt.Errorf("failed to parse fields: %w", err)
			}

			srcClient, dstClient, err := common.NewClientPair(ctx, src, dst, srcToken, dstToken)
			if err != nil {
				return err
			}

			srcUser, err := gh.GetLoginUser(ctx, srcClient)
			if err != nil {
				return fmt.Errorf("failed to get source user profile: %w", err)
			}

			req, changed := profile.BuildUpdateRequest(srcUser, parsedFields)
			if dryrun {
				logger.Info("[dryrun] would update profile fields", "fields", changed)
			} else if len(changed) > 0 {
				if _, err := gh.UpdateLoginUser(ctx, dstClient, req); err != nil {
					return fmt.Errorf("failed to update destination profile: %w", err)
				}
				logger.Info("updated profile fields", "fields", changed)
			}

			if hasField(parsedFields, profile.FieldSocialAccounts) {
				srcAccounts, err := gh.ListLoginUserSocialAccounts(ctx, srcClient)
				if err != nil {
					return fmt.Errorf("failed to list source social accounts: %w", err)
				}
				dstAccounts, err := gh.ListLoginUserSocialAccounts(ctx, dstClient)
				if err != nil {
					return fmt.Errorf("failed to list destination social accounts: %w", err)
				}
				missing := profile.MissingSocialAccounts(srcAccounts, dstAccounts)
				if dryrun {
					logger.Info("[dryrun] would add social accounts", "urls", missing)
				} else if len(missing) > 0 {
					if _, err := gh.AddLoginUserSocialAccounts(ctx, dstClient, missing); err != nil {
						return fmt.Errorf("failed to add destination social accounts: %w", err)
					}
					logger.Info("added social accounts", "urls", missing)
				}
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&src, "src", "s", "", "Source GitHub host (default: current host from gh auth)")
	f.StringVarP(&dst, "dst", "d", "", "Destination GitHub host (default: current host from gh auth)")
	f.StringVar(&srcToken, "src-token", "", "Token for the source GitHub host")
	f.StringVar(&dstToken, "dst-token", "", "Token for the destination GitHub host")
	f.StringSliceVar(&fields, "fields", profile.AllFields(), "Comma-separated list of fields to copy")
	f.BoolVarP(&dryrun, "dryrun", "n", false, "Dry run: show what would be copied without making changes")

	return cmd
}

func hasField(fields []string, field profile.Field) bool {
	for _, f := range fields {
		if f == string(field) {
			return true
		}
	}
	return false
}
