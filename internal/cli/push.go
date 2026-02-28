package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shimabukuromeg/noti/internal/markdown"
	"github.com/shimabukuromeg/noti/internal/notion"
	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	var (
		database string
		pageID   string
		force    bool
	)

	cmd := &cobra.Command{
		Use:   "push <file.md>",
		Short: "Push a Markdown file to Notion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			cfg := loadConfig()
			if err := cfg.ValidateToken(); err != nil {
				return err
			}

			// Read and parse file
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			parsed, err := markdown.Parse(string(data))
			if err != nil {
				return fmt.Errorf("failed to parse markdown: %w", err)
			}

			client := newNotionClient(cfg)
			ctx := cmd.Context()

			// Build page properties from frontmatter
			props := notion.PageProperties{
				Title:     parsed.Frontmatter.Title,
				Slug:      parsed.Frontmatter.Slug,
				Date:      parsed.Frontmatter.Date,
				Tags:      parsed.Frontmatter.Tags,
				Excerpt:   parsed.Frontmatter.Excerpt,
				Published: parsed.Frontmatter.Published,
			}

			// Use filename as title if no title in frontmatter
			if props.Title == "" {
				base := filepath.Base(filePath)
				props.Title = strings.TrimSuffix(base, filepath.Ext(base))
			}

			// Resolve page ID: --page-id flag > frontmatter notion_id
			targetPageID := pageID
			if targetPageID == "" {
				targetPageID = parsed.Frontmatter.NotionID
			}

			if targetPageID != "" {
				// Update existing page
				if !force {
					// Conflict detection: compare last_edited_time
					page, err := client.GetPage(ctx, targetPageID)
					if err != nil {
						return fmt.Errorf("failed to get page: %w", err)
					}

					fileInfo, err := os.Stat(filePath)
					if err != nil {
						return fmt.Errorf("failed to stat file: %w", err)
					}

					if page.LastEditedTime.After(fileInfo.ModTime()) {
						return fmt.Errorf("conflict: Notion page was edited more recently than local file (Notion: %s, local: %s). Use --force to overwrite",
							page.LastEditedTime.Format("2006-01-02 15:04:05"),
							fileInfo.ModTime().Format("2006-01-02 15:04:05"))
					}
				}

				// Update properties
				if err := client.UpdatePageProperties(ctx, targetPageID, props); err != nil {
					return fmt.Errorf("failed to update page properties: %w", err)
				}

				// Replace markdown content
				if _, err := client.ReplaceMarkdown(ctx, targetPageID, parsed.Body); err != nil {
					return fmt.Errorf("failed to replace markdown: %w", err)
				}

				fmt.Fprintf(os.Stderr, "✓ Pushed \"%s\" → https://www.notion.so/%s\n", props.Title, strings.ReplaceAll(targetPageID, "-", ""))
				return nil
			}

			// Create new page
			dbID := database
			if dbID == "" {
				dbID = cfg.DatabaseID
			}
			if dbID == "" {
				return fmt.Errorf("database ID required for creating new pages: use --database flag or set NOTI_DATABASE_ID")
			}

			page, err := client.CreatePage(ctx, dbID, props)
			if err != nil {
				return fmt.Errorf("failed to create page: %w", err)
			}

			// Insert markdown content
			if _, err := client.InsertMarkdown(ctx, page.ID, parsed.Body); err != nil {
				return fmt.Errorf("failed to insert markdown: %w", err)
			}

			// Write back notion_id to frontmatter
			absPath, err := filepath.Abs(filePath)
			if err != nil {
				absPath = filePath
			}
			if err := markdown.UpdateNotionID(absPath, page.ID); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to write notion_id to file: %v\n", err)
			}

			fmt.Fprintf(os.Stderr, "✓ Pushed \"%s\" → %s\n", props.Title, page.URL)
			return nil
		},
	}

	cmd.Flags().StringVar(&database, "database", "", "Database ID for new page creation")
	cmd.Flags().StringVar(&pageID, "page-id", "", "Explicit page ID to update")
	cmd.Flags().BoolVar(&force, "force", false, "Skip conflict warning and force overwrite")

	return cmd
}
