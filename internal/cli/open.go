package cli

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func newOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <page-id>",
		Short: "Open a Notion page in the browser",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pageID := args[0]
			// Normalize page ID: remove hyphens so both formats work
			normalized := strings.ReplaceAll(pageID, "-", "")
			url := "https://www.notion.so/" + normalized

			if err := openBrowser(url); err != nil {
				return fmt.Errorf("failed to open browser: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Opening %s\n", url)
			return nil
		},
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
