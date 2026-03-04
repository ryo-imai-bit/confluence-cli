package label

import (
	"github.com/spf13/cobra"
)

// LabelCmd is the root command for label operations
var LabelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage Confluence labels",
	Long:  `Commands for listing labels and finding content by label.`,
}

func init() {
	LabelCmd.AddCommand(listCmd)
	LabelCmd.AddCommand(pagesCmd)
}
