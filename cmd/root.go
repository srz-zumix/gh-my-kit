/*
Copyright © 2025 srz_zumix
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-my-kit/version"
)

var rootCmd = &cobra.Command{
	Use:     "gh-my-kit",
	Short:   "My personal GitHub CLI extension kit",
	Long:    `gh-my-kit is my personal GitHub CLI extension kit.`,
	Version: version.Version,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
