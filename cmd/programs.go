package cmd

import (
	"fmt"

	"github.com/derek/seats-cli/internal/format"
	"github.com/spf13/cobra"
)

var programsCmd = &cobra.Command{
	Use:   "programs",
	Short: "List available mileage programs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if pretty {
			fmt.Print(format.PrettyPrograms())
		} else {
			fmt.Print(format.FormatProgramsMarkdown())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(programsCmd)
}
