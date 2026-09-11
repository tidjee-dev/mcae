package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/mcae/cmd/extract"
	"github.com/tidjee-dev/mcae/internal/ui"
)

var (
	noColorFlag bool
	quietFlag   bool
)

var rootCmd = &cobra.Command{
	Use:   "mcae",
	Short: "Minecraft Companion Assets Extractor",
	Long:  "mcae extracts and inspects Minecraft mod, modpack and vanilla assets.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColorFlag || os.Getenv("NO_COLOR") != "" {
			ui.SetNoColor(true)
		}
	},
	// Keep quiet available for scripts; commands check it via helper.
}

// IsQuiet reports whether --quiet was set.
func IsQuiet() bool { return quietFlag }

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress informational output")
	rootCmd.AddCommand(extract.NewExtractCmd())
}
