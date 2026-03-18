// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/derek/seats-cli/internal/api"
	"github.com/spf13/cobra"
)

var pretty bool

var rootCmd = &cobra.Command{
	Use:   "seats",
	Short: "Search award flight availability via Seats.aero",
	Long:  "A CLI tool for searching award flight availability across 20+ mileage programs via the Seats.aero Partner API.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip API key check for programs command (no API call needed)
		if cmd.Name() == "programs" {
			return nil
		}
		if os.Getenv("SEATS_AERO_API_KEY") == "" {
			fmt.Fprintln(os.Stderr, "Error: SEATS_AERO_API_KEY environment variable not set.")
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "Styled terminal output instead of markdown")
}

func Execute() error {
	return rootCmd.Execute()
}

func handleAPIError(err error) error {
	apiErr, ok := err.(*api.APIError)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(4)
	}
	fmt.Fprintf(os.Stderr, "Error: %s\n", apiErr.Message)
	os.Exit(apiErr.ExitCode())
	return nil
}
