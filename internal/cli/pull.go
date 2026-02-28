package cli

import (
	"fmt"
	"os"

	"github.com/shimabukuromeg/noti/internal/markdown"
	"github.com/spf13/cobra"
)

func newPullCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "pull <page-id>",
		Short: "Pull a Notion page as Markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pageID := args[0]

			cfg := loadConfig()
			if err := cfg.ValidateToken(); err != nil {
				return err
			}

			client := newNotionClient(cfg)
			ctx := cmd.Context()

			// Get page properties
			page, err := client.GetPage(ctx, pageID)
			if err != nil {
				return fmt.Errorf("failed to get page: %w", err)
			}

			// Get markdown content
			md, err := client.RetrieveMarkdown(ctx, pageID)
			if err != nil {
				return fmt.Errorf("failed to retrieve markdown: %w", err)
			}

			// Build frontmatter from page properties
			fm := markdown.Frontmatter{
				Title:    page.Title(),
				Slug:     page.Slug(),
				Date:     page.DateStr(),
				Tags:     page.Tags(),
				Excerpt:  page.ExcerptStr(),
				Published: page.IsPublished(),
				NotionID: page.ID,
			}

			result := markdown.Render(fm, md.Markdown)

			if output != "" {
				if err := os.WriteFile(output, []byte(result), 0644); err != nil {
					return fmt.Errorf("failed to write file: %w", err)
				}
				fmt.Fprintf(os.Stderr, "✓ Pulled \"%s\" → %s\n", page.Title(), output)
				return nil
			}

			fmt.Print(result)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")

	return cmd
}
