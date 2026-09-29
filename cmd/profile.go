/*
Copyright © 2025 srz_zumix
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-my-kit/cmd/profile"
)

func NewProfileCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "profile",
		Short: "Manage GitHub user profiles",
		Long:  `Commands for managing GitHub user profiles.`,
	}

	cmd.AddCommand(profile.NewCopyCmd())

	return cmd
}

func init() {
	rootCmd.AddCommand(NewProfileCmd())
}
