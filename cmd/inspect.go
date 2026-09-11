package cmd

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/mcae/internal/inspect"
	"github.com/tidjee-dev/mcae/internal/ui"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <path>",
	Short: "Inspect a mod JAR or extracted directory",
	Long:  "Accepts a .jar file or an extracted directory. Detects modpack directories (containing mods/).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := inspect.Inspect(args[0])
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		if report.Modpack != nil {
			fmt.Fprintln(out, ui.Title("Minecraft Asset Summary"))
			fmt.Fprintln(out)
			fmt.Fprintf(out, "Path: %s (modpack)\n\n", report.Modpack.Path)
			names := make([]string, 0, len(report.Modpack.Mods))
			assets := make([]int, 0, len(report.Modpack.Mods))
			data := make([]int, 0, len(report.Modpack.Mods))
			for _, m := range report.Modpack.Mods {
				a := 0
				for _, v := range m.Assets {
					a += v
				}
				d := 0
				for _, v := range m.Data {
					d += v
				}
				names = append(names, filepath.Base(m.Path))
				assets = append(assets, a)
				data = append(data, d)
			}
			fmt.Fprintln(out, ui.ModpackTable(names, assets, data))
			return nil
		}

		s := report.Single
		fmt.Fprintln(out, ui.Title("Minecraft Asset Summary"))
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Path: %s\n\n", s.Path)

		ns := append([]string{}, s.Namespaces...)
		sort.Strings(ns)
		fmt.Fprintf(out, "Assets\n  Namespaces: %d\n\n", len(ns))
		fmt.Fprintln(out, ui.InspectTable(s.Assets, nil))
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Data")
		fmt.Fprintln(out, ui.InspectTable(nil, s.Data))
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Namespaces")
		if len(ns) == 0 {
			fmt.Fprintln(out, "  (none)")
			return nil
		}
		for _, n := range ns {
			fmt.Fprintf(out, "  - %s\n", n)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
