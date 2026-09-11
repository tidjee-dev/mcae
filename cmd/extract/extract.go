// Package extract implements the mcae extract command group (mod, modpack, vanilla).
package extract

import (
	"github.com/spf13/cobra"
)

// NewExtractCmd returns the `extract` parent command group.
func NewExtractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract Minecraft assets",
		Long:  "Extract Minecraft assets from mods, modpacks and vanilla clients.",
	}

	cmd.AddCommand(newModCmd())
	cmd.AddCommand(newModpackCmd())
	cmd.AddCommand(newVanillaCmd())

	return cmd
}
