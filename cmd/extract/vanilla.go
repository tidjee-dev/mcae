package extract

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	vanilla "github.com/tidjee-dev/mcae/internal/minecraft/vanilla"
	"github.com/tidjee-dev/mcae/internal/ui"
)

func newVanillaCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "vanilla <version>",
		Short: "Extract vanilla Minecraft assets",
		Example: "  mcae extract vanilla 1.21.8\n" +
			"  mcae extract vanilla 1.21.8 --output ./data/minecraft",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			versionID := args[0]
			start := time.Now()
			frame := ui.SpinnerFrame(0)
			cmd.Printf("Resolving Minecraft %s %s\n", versionID, frame)

			res, err := vanilla.ExtractVanilla(versionID, output, func(written, total int64) {
				if total > 0 {
					// Periodic progress is intentionally minimal; the final
					// bar below shows completion. Live updating is skipped
					// to keep piped output stable.
					_ = written
				}
			})
			if err != nil {
				return err
			}

			cmd.Printf("%s %s\n", ui.ModpackHeader(1), ui.StaticProgress(1))
			cmd.Printf("%s\n", ui.ExtractDone(res.Assets, res.Data, res.OutDir))
			cmd.Printf("%s\n", ui.Muted(fmt.Sprintf("Done in %s", time.Since(start).Round(time.Millisecond))))
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output base directory (default ./extracted/vanilla/<version>)")

	return cmd
}
