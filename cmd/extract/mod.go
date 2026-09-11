package extract

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tidjee-dev/mcae/internal/extractor"
	"github.com/tidjee-dev/mcae/internal/ui"
)

func newModCmd() *cobra.Command {
	var output string
	var force bool

	cmd := &cobra.Command{
		Use:   "mod <mod.jar>",
		Short: "Extract a Minecraft mod JAR",
		Long:  "Extract retained assets and data from a single Minecraft mod JAR. Output goes to ./extracted/<modname> unless --output is given.",
		Example: "  mcae extract mod <mod_jar_file>\n" +
			"  mcae extract mod <mod_jar_file> --output <output_dir>\n" +
			"  mcae extract mod <mod_jar_file> --force",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jarPath := args[0]
			quiet, _ := cmd.Flags().GetBool("quiet")

			// Validate before printing anything (no misleading "Extracting…" on failure).
			if st, err := os.Stat(jarPath); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("%w: %s", extractor.ErrModNotFound, jarPath)
				}
				return err
			} else if st.IsDir() || !strings.EqualFold(filepath.Ext(jarPath), ".jar") {
				return fmt.Errorf("%w (expected a .jar file): %s", extractor.ErrInvalidInput, jarPath)
			}

			if !quiet {
				cmd.Printf("%s\n", ui.ExtractLine(jarPath))
			}

			res, err := extractor.ExtractModWithForce(jarPath, output, force)
			if err != nil {
				return err
			}

			if quiet {
				fmt.Fprintln(cmd.OutOrStdout(), res.OutDir)
				return nil
			}
			cmd.Printf("%s\n", ui.ExtractDone(res.Assets, res.Data, res.OutDir))
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output base directory (default ./extracted/<modname>)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "remove the destination directory before extracting")

	return cmd
}
