// Package cmd wires the mcae Cobra command tree (extract, inspect, version).
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/mcae/cmd/extract"
	"github.com/tidjee-dev/mcae/internal/ui"
	"github.com/tidjee-dev/mcae/internal/version"
)

var (
	noColorFlag bool
	quietFlag   bool
)

var rootCmd = &cobra.Command{
	Use:   "mcae",
	Short: "Minecraft Companion Assets Extractor",
	Long:  "mcae extracts and inspects Minecraft mod, modpack and vanilla assets.",
	Example: "  mcae extract mod <mod_jar_file>\n" +
		"  mcae extract modpack <modpack_dir>\n" +
		"  mcae extract vanilla <minecraft_version>\n" +
		"  mcae inspect <extracted_dir>",
	Version:       version.Version,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColorFlag || os.Getenv("NO_COLOR") != "" {
			ui.SetNoColor(true)
		}
	},
}

// IsQuiet reports whether --quiet was set.
func IsQuiet() bool { return quietFlag }

// Execute runs the mcae command tree, printing failures once to stderr.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorLine(err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress informational output")
	rootCmd.AddCommand(extract.NewExtractCmd())
}
