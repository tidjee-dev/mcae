package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/mcae/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the mcae version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", version.Name, version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
