package page

import (
	"github.com/spf13/cobra"
)

// PageCmd represents the page command group
var PageCmd = &cobra.Command{
	Use:   "page",
	Short: "Manage Confluence pages",
	Long:  `Commands for listing, viewing, creating, updating, and deleting Confluence pages.`,
}

func init() {
	PageCmd.AddCommand(listCmd)
	PageCmd.AddCommand(getCmd)
	PageCmd.AddCommand(createCmd)
	PageCmd.AddCommand(updateCmd)
	PageCmd.AddCommand(deleteCmd)
}
