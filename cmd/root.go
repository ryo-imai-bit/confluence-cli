package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ryo-imai-bit/confluence-cli/cmd/page"
)

var rootCmd = &cobra.Command{
	Use:   "confluence",
	Short: "CLI tool for Atlassian Confluence",
	Long:  `A command-line interface for managing Confluence pages and content.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(page.PageCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	// Config validation is done at command execution time, not initialization
	// This allows help and config commands to work without valid credentials
}
