package cli

import (
	"fmt"

	"github.com/shimabukuromeg/noti/internal/auth"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored Notion credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.DeleteToken(); err != nil {
				return fmt.Errorf("failed to remove credentials: %w", err)
			}
			fmt.Println("✓ Logged out. Token removed.")
			return nil
		},
	}
}
