package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/mcae/internal/inspect"
	"github.com/tidjee-dev/mcae/internal/ui"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <path>",
	Short: "Inspect a mod JAR or extracted directory",
	Long:  "Accepts a .jar file, an extracted directory, or a modpack directory (containing mods/). Prints one summary table with totals plus the namespace list.",
	Example: "  mcae inspect <mod_jar_file>\n" +
		"  mcae inspect <extracted_dir>\n" +
		"  mcae inspect <modpack_dir>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := inspect.Inspect(args[0])
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		if report.Modpack != nil {
			fmt.Fprintln(out, ui.Title("Minecraft Asset Summary"))
			fmt.Fprintf(out, "\nPath: %s (modpack, %d mods)\n\n", report.Modpack.Path, len(report.Modpack.Mods))
			rows := make([]ui.ModRow, 0, len(report.Modpack.Mods))
			for _, m := range report.Modpack.Mods {
				rows = append(rows, ui.ModRow{
					Name:       shortName(m.Path),
					Namespaces: len(m.Namespaces),
					Assets:     m.TotalAssets(),
					Data:       m.TotalData(),
				})
			}
			sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
			fmt.Fprintln(out, ui.RenderModpackTable(rows))
			assets, data := report.Modpack.Totals()
			fmt.Fprintf(out, "\nTotal: %d mods, %d assets, %d data files\n", len(rows), assets, data)
			return nil
		}

		s := report.Single
		ns := append([]string{}, s.Namespaces...)
		sort.Strings(ns)
		fmt.Fprintln(out, ui.RenderSummary(s.Path, ns, s.Assets, s.Data))
		return nil
	},
}

func shortName(path string) string {
	// Modpack summaries store the full JAR path; show the base for the table.
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
