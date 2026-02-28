package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <page-id>",
		Short: "Archive a Notion page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pageID := args[0]

			cfg := loadConfig()
			if err := cfg.ValidateToken(); err != nil {
				return err
			}

			client := newNotionClient(cfg)
			ctx := cmd.Context()

			// Get page title for confirmation
			page, err := client.GetPage(ctx, pageID)
			if err != nil {
				return fmt.Errorf("failed to get page: %w", err)
			}

			title := page.Title()

			if !force {
				fmt.Fprintf(os.Stderr, "Archive \"%s\"? [y/N] ", title)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Fprintln(os.Stderr, "Cancelled.")
					return nil
				}
			}

			if err := client.ArchivePage(ctx, pageID); err != nil {
				return fmt.Errorf("failed to archive page: %w", err)
			}

			fmt.Fprintf(os.Stderr, "✓ Archived \"%s\"\n", title)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
