package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/shimabukuromeg/noti/internal/notion"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		limit     int
		jsonFlag  bool
		published bool
		tag       string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pages in the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()
			if err := cfg.ValidateToken(); err != nil {
				return err
			}
			if err := cfg.ValidateDatabase(); err != nil {
				return err
			}

			client := newNotionClient(cfg)

			opts := notion.QueryOptions{
				PageSize: limit,
			}
			if cmd.Flags().Changed("published") {
				opts.Published = &published
			}
			if tag != "" {
				opts.Tag = tag
			}

			result, err := client.QueryDatabase(cmd.Context(), cfg.DatabaseID, opts)
			if err != nil {
				return fmt.Errorf("failed to query database: %w", err)
			}

			if jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result.Results)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PAGE_ID\tTITLE\tDATE\tPUBLISHED")
			for _, page := range result.Results {
				pub := ""
				if page.IsPublished() {
					pub = "✓"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					page.ID,
					page.Title(),
					page.DateStr(),
					pub,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of pages to list")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&published, "published", false, "Filter by Published=true")
	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")

	return cmd
}
