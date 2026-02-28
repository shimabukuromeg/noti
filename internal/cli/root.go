package cli

import (
	"github.com/shimabukuromeg/noti/internal/auth"
	"github.com/shimabukuromeg/noti/internal/config"
	"github.com/shimabukuromeg/noti/internal/notion"
	"github.com/spf13/cobra"
)

var (
	flagToken    string
	flagDatabase string
)

func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "noti",
		Short: "Sync local Markdown files with Notion",
	}

	root.PersistentFlags().StringVarP(&flagToken, "token", "t", "", "Notion integration token (overrides NOTION_TOKEN)")
	root.PersistentFlags().StringVarP(&flagDatabase, "database", "d", "", "Default database ID (overrides NOTI_DATABASE_ID)")

	root.AddCommand(
		newPushCmd(),
		newPullCmd(),
		newDeleteCmd(),
		newListCmd(),
		newOpenCmd(),
		newLoginCmd(),
		newLogoutCmd(),
		newVersionCmd(version),
	)

	return root
}

// loadConfig loads config and applies flag overrides.
// Token priority: --token flag > NOTION_TOKEN env > ~/.config/noti/token.json (OAuth)
func loadConfig() *config.Config {
	cfg := config.Load()
	if flagToken != "" {
		cfg.NotionToken = flagToken
	}
	if flagDatabase != "" {
		cfg.DatabaseID = flagDatabase
	}

	// Fall back to stored OAuth token
	if cfg.NotionToken == "" {
		if token, err := auth.LoadToken(); err == nil && token != nil {
			cfg.NotionToken = token.AccessToken
		}
	}

	return cfg
}

// newNotionClient creates a Notion client from the loaded config.
func newNotionClient(cfg *config.Config) *notion.Client {
	return notion.NewClient(cfg.NotionToken)
}
