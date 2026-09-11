package extract

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/mcae/internal/extractor"
	"github.com/tidjee-dev/mcae/internal/ui"
	"github.com/tidjee-dev/mcae/internal/ui/live"
)

func newModpackCmd() *cobra.Command {
	var output string
	var force bool

	cmd := &cobra.Command{
		Use:   "modpack <modpack>",
		Short: "Extract all Minecraft mod JARs from a modpack",
		Long:  "Find .jar files directly inside <modpack>/mods and extract each into its own directory under <modpack>/extracted unless --output is given.",
		Example: "  mcae extract modpack ./data/modpacks/cuboid-outpost-luxury\n" +
			"  mcae extract modpack ./data/modpacks/cuboid-outpost-luxury --output ./data/extracted\n" +
			"  mcae extract modpack ./data/modpacks/cuboid-outpost-luxury --force",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			quiet, _ := cmd.Flags().GetBool("quiet")
			start := time.Now()

			// Resolve the JAR list first so errors surface before any output.
			modsDir, outputDir, jars, err := extractor.ListModJars(args[0], output)
			if err != nil {
				return err
			}

			reused := 0
			if !force {
				for _, j := range jars {
					if st, err := os.Stat(filepath.Join(outputDir, extractor.JarModName(j))); err == nil && st.IsDir() {
						reused++
					}
				}
			}

			// Live TUI on interactive terminals.
			if ui.IsTTY() && !quiet && !ui.IsNoColor() {
				if err := os.MkdirAll(outputDir, 0o755); err != nil {
					return fmt.Errorf("create output directory: %w", err)
				}
				results, err := live.RunModpack(jars, modsDir, outputDir, force, cmd.OutOrStdout())
				if err != nil {
					return err
				}
				totalA, totalD := 0, 0
				for _, r := range results {
					totalA += r.Assets
					totalD += r.Data
				}
				cmd.Printf("%s\n", ui.Muted(fmt.Sprintf("Extracted %d mods: %d assets, %d data files in %s",
					len(results), totalA, totalD, time.Since(start).Round(time.Millisecond))))
				return nil
			}

			// Streaming fallback: print each mod as it finishes (stable for pipes).
			if !quiet {
				cmd.Printf("%s\n", ui.ModpackHeader(len(jars)))
				if reused > 0 {
					cmd.Printf("%s\n", ui.Muted(fmt.Sprintf("Reusing %d existing directories (use --force to clean)", reused)))
				}
			}
			totalA, totalD := 0, 0
			results, err := extractor.ExtractModpackWithForce(args[0], output, force,
				func(i, total int, res extractor.Result) {
					totalA += res.Assets
					totalD += res.Data
					if quiet {
						return
					}
					cmd.Printf("%s %s\n", ui.ModpackStep(i+1, total, res.JarName),
						ui.StaticProgress(float64(i+1)/float64(total)))
					cmd.Printf("%s\n", ui.ExtractDone(res.Assets, res.Data, res.OutDir))
				})
			if err != nil {
				return err
			}
			if quiet {
				for _, r := range results {
					fmt.Fprintln(cmd.OutOrStdout(), r.OutDir)
				}
				return nil
			}
			cmd.Printf("%s\n", ui.Muted(fmt.Sprintf("Extracted %d mods: %d assets, %d data files in %s",
				len(results), totalA, totalD, time.Since(start).Round(time.Millisecond))))
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output directory (default <modpack>/extracted)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "remove each mod destination before extracting")

	return cmd
}
