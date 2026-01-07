package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewCleanCommand creates the clean command
func NewCleanCommand() *cobra.Command {
	var all bool
	var force bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean ritual cache",
		Long: `Clean the ritual cache directory.

This removes cached ritual files and forces re-extraction of embedded rituals
on next use. Useful when rituals appear outdated after upgrading toutā.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			cachePath := filepath.Join(homeDir, ".toutago", "ritual-cache")
			
			// Check if cache exists
			if _, err := os.Stat(cachePath); os.IsNotExist(err) {
				fmt.Println("Cache is already clean")
				return nil
			}

			// Confirm unless force flag is set
			if !force {
				fmt.Printf("This will delete all cached rituals in: %s\n", cachePath)
				fmt.Print("Are you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			// Remove cache
			if err := os.RemoveAll(cachePath); err != nil {
				return fmt.Errorf("failed to clean cache: %w", err)
			}

			fmt.Println("✓ Cache cleaned successfully")
			fmt.Println("Rituals will be re-extracted on next use")
			
			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Clean all ritual data (not just cache)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
