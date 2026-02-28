package cli

import (
	"fmt"

	"github.com/shimabukuromeg/noti/internal/auth"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var (
		clientID     string
		clientSecret string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Notion via OAuth",
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientID == "" || clientSecret == "" {
				return fmt.Errorf("both --client-id and --client-secret are required.\n\nGet them from https://www.notion.so/profile/integrations → your integration → Distribution tab")
			}

			oauth := auth.NewOAuthConfig(clientID, clientSecret)
			token, err := oauth.Login(cmd.Context())
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			fmt.Printf("✓ Logged in (workspace: %s)\n", token.WorkspaceID)
			fmt.Println("Token saved to ~/.config/noti/token.json")
			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth client ID (required)")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth client secret (required)")

	return cmd
}
