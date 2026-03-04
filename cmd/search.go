package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ryo-imai-bit/confluence-cli/internal/api"
)

var (
	searchSpaceID string
	searchLimit   int
	searchFormat  string
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for pages by title",
	Long:  `Search for Confluence pages by title.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().StringVar(&searchSpaceID, "space-id", "", "Filter results by space ID")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 25, "Maximum number of results to return")
	searchCmd.Flags().StringVar(&searchFormat, "format", "table", "Output format (table, json)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	service, err := api.NewPageService()
	if err != nil {
		return err
	}

	pages, err := service.SearchPages(query, searchSpaceID, searchLimit)
	if err != nil {
		return err
	}

	if len(pages.Results) == 0 {
		fmt.Println("No pages found.")
		return nil
	}

	if searchFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pages)
	}

	// Table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tSPACE ID")
	fmt.Fprintln(w, "--\t-----\t------\t--------")
	for _, p := range pages.Results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.Title, p.Status, p.SpaceID)
	}
	return w.Flush()
}
