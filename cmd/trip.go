package cmd

import (
	"fmt"
	"os"

	"github.com/derek/seats-cli/internal/api"
	"github.com/derek/seats-cli/internal/format"
	"github.com/spf13/cobra"
)

var tripCmd = &cobra.Command{
	Use:   "trip <availability-id>",
	Short: "Get trip details and booking links",
	Args:  cobra.ExactArgs(1),
	RunE:  runTrip,
}

func init() {
	rootCmd.AddCommand(tripCmd)
}

func runTrip(cmd *cobra.Command, args []string) error {
	id := args[0]

	client := api.NewClient(os.Getenv("SEATS_AERO_API_KEY"))
	resp, err := client.Trip(id)
	if err != nil {
		return handleAPIError(err)
	}

	fmt.Print(format.FormatTripMarkdown(id, resp))
	return nil
}
