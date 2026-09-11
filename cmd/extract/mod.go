package extract

import (
	"github.com/spf13/cobra"

	"github.com/tidjee-dev/mcae/internal/extractor"
	"github.com/tidjee-dev/mcae/internal/ui"
)

func newModCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "mod <mod.jar>",
		Short: "Extract a Minecraft mod JAR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jarPath := args[0]
			cmd.Printf("%s\n", ui.ExtractLine(jarPath))

			res, err := extractor.ExtractMod(jarPath, output)
			if err != nil {
				return err
			}

			cmd.Printf("%s\n", ui.ExtractDone(res.Assets, res.Data, res.OutDir))
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output base directory (default ./extracted/<modname>)")

	return cmd
}
