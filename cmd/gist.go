/*
Copyright © 2025 srz_zumix
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-my-kit/cmd/gist"
)

func NewGistCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "gist",
		Short: "Manage gists",
		Long:  `Commands for managing GitHub Gists.`,
	}

	cmd.AddCommand(gist.NewConvertCmd())
	cmd.AddCommand(gist.NewCopyCmd())
	cmd.AddCommand(gist.NewMigrateCmd())

	return cmd
}

func init() {
	rootCmd.AddCommand(NewGistCmd())
}
