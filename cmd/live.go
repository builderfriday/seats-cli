package cmd

import (
	"fmt"
	"os"

	"github.com/derek/seats-cli/internal/api"
	"github.com/derek/seats-cli/internal/format"
	"github.com/spf13/cobra"
)

var liveCmd = &cobra.Command{
	Use:   "live",
	Short: "Real-time award search",
	RunE:  runLive,
}

func init() {
	liveCmd.Flags().String("from", "", "Origin airport code")
	liveCmd.Flags().String("to", "", "Destination airport code")
	liveCmd.Flags().String("date", "", "Departure date YYYY-MM-DD")
	liveCmd.Flags().String("program", "", "Mileage program (single value)")
	liveCmd.Flags().Int("seats", 1, "Passenger count (1-9)")

	liveCmd.MarkFlagRequired("from")
	liveCmd.MarkFlagRequired("to")
	liveCmd.MarkFlagRequired("date")
	liveCmd.MarkFlagRequired("program")

	rootCmd.AddCommand(liveCmd)
}

func runLive(cmd *cobra.Command, args []string) error {
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	date, _ := cmd.Flags().GetString("date")
	program, _ := cmd.Flags().GetString("program")
	seats, _ := cmd.Flags().GetInt("seats")

	if seats < 1 || seats > 9 {
		fmt.Fprintln(os.Stderr, "Error: --seats must be between 1 and 9")
		os.Exit(1)
	}

	client := api.NewClient(os.Getenv("SEATS_AERO_API_KEY"))

	resp, err := client.Live(api.LiveParams{
		OriginAirport:      from,
		DestinationAirport: to,
		DepartureDate:      date,
		Source:             program,
		SeatCount:          seats,
	})
	if err != nil {
		return handleAPIError(err)
	}

	if pretty {
		fmt.Print(format.PrettyLive(from, to, date, program, resp.Results))
	} else {
		fmt.Print(format.FormatLiveMarkdown(from, to, date, program, resp.Results))
	}
	return nil
}
