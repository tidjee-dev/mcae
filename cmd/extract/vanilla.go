package extract

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	vanilla "github.com/tidjee-dev/mcae/internal/minecraft/vanilla"
	"github.com/tidjee-dev/mcae/internal/ui"
	"github.com/tidjee-dev/mcae/internal/ui/live"
)

func newVanillaCmd() *cobra.Command {
	var output string
	var force bool

	cmd := &cobra.Command{
		Use:   "vanilla <version>",
		Short: "Extract vanilla Minecraft assets",
		Long:  "Resolve the version through Mojang's version manifest, download the client JAR to a temp file, and extract retained assets. Unofficial tool, not affiliated with Mojang.",
		Example: "  mcae extract vanilla <minecraft_version>\n" +
			"  mcae extract vanilla <minecraft_version> --output <output_dir>\n" +
			"  mcae extract vanilla <minecraft_version> --force",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			quiet, _ := cmd.Flags().GetBool("quiet")
			versionID := args[0]
			if versionID == "" {
				return fmt.Errorf("minecraft version is required")
			}
			start := time.Now()

			// Live TUI on interactive terminals.
			if ui.IsTTY() && !quiet && !ui.IsNoColor() {
				res, err := live.RunVanilla(versionID, output, force, "", cmd.OutOrStdout())
				if err != nil {
					return err
				}
				cmd.Printf("%s\n", ui.Muted(fmt.Sprintf("Vanilla %s: %d assets, %d data files in %s",
					res.Version, res.Assets, res.Data, time.Since(start).Round(time.Millisecond))))
				return nil
			}

			// Streaming fallback with throttled byte progress.
			if !quiet {
				cmd.Printf("Resolving Minecraft %s…\n", versionID)
			}
			lastPct := -1
			res, err := vanilla.ExtractVanillaWithForce(versionID, output, force, func(written, total int64) {
				if quiet || total <= 0 {
					return
				}
				pct := int(written * 100 / total)
				if pct/10 != lastPct/10 {
					lastPct = pct
					cmd.Printf("  downloading client: %s / %s (%d%%)\n",
						ui.FormatBytes(written), ui.FormatBytes(total), pct)
				}
			})
			if err != nil {
				return err
			}

			if quiet {
				fmt.Fprintln(cmd.OutOrStdout(), res.OutDir)
				return nil
			}
			cmd.Printf("%s %s\n", ui.ModpackHeader(1), ui.StaticProgress(1))
			cmd.Printf("%s\n", ui.ExtractDone(res.Assets, res.Data, res.OutDir))
			cmd.Printf("%s\n", ui.Muted(fmt.Sprintf("Done in %s", time.Since(start).Round(time.Millisecond))))
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output base directory (default ./extracted/vanilla/<version>)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "remove the destination directory before extracting")

	return cmd
}
