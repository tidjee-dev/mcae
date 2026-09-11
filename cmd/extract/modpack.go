package extract

import (
	"github.com/spf13/cobra"

	"github.com/tidjee-dev/mcae/internal/extractor"
	"github.com/tidjee-dev/mcae/internal/ui"
)

func newModpackCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "modpack <modpack>",
		Short: "Extract all Minecraft mod JARs from a modpack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := extractor.ExtractModpack(args[0], output)
			if err != nil {
				return err
			}

			cmd.Printf("%s\n", ui.ModpackHeader(len(results)))
			for i, r := range results {
				cmd.Printf("%s %s\n", ui.ModpackStep(i+1, len(results), r.JarName), ui.StaticProgress(float64(i+1)/float64(len(results))))
				cmd.Printf("%s\n", ui.ExtractDone(r.Assets, r.Data, r.OutDir))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output directory (default <modpack>/extracted)")

	return cmd
}
