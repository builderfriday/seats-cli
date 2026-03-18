package cmd

import (
	"fmt"
	"os"

	"github.com/derek/seats-cli/internal/api"
	"github.com/derek/seats-cli/internal/format"
	"github.com/derek/seats-cli/internal/model"
	seatsSort "github.com/derek/seats-cli/internal/sort"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search cached award availability",
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().String("from", "", "Origin airport(s), comma-separated")
	searchCmd.Flags().String("to", "", "Destination airport(s), comma-separated")
	searchCmd.Flags().String("date", "", "Start date (YYYY-MM-DD)")
	searchCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	searchCmd.Flags().String("cabin", "", "Cabin class: economy, premium, business, first (comma-separated)")
	searchCmd.Flags().String("program", "", "Mileage program filter (comma-separated)")
	searchCmd.Flags().String("carrier", "", "Airline filter (comma-separated)")
	searchCmd.Flags().Bool("direct", false, "Nonstop flights only")
	searchCmd.Flags().String("sort", "", "Sort keys: cabin, miles, stops, date, taxes, seats, airline (comma-separated)")
	searchCmd.Flags().Int("limit", 50, "Max results (10-1000)")
	searchCmd.Flags().Int("skip", 0, "Skip N results for pagination")

	_ = searchCmd.MarkFlagRequired("from")
	_ = searchCmd.MarkFlagRequired("to")

	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	date, _ := cmd.Flags().GetString("date")
	endDate, _ := cmd.Flags().GetString("end-date")
	cabin, _ := cmd.Flags().GetString("cabin")
	program, _ := cmd.Flags().GetString("program")
	carrier, _ := cmd.Flags().GetString("carrier")
	direct, _ := cmd.Flags().GetBool("direct")
	sortStr, _ := cmd.Flags().GetString("sort")
	limit, _ := cmd.Flags().GetInt("limit")
	skip, _ := cmd.Flags().GetInt("skip")

	if limit < 10 || limit > 1000 {
		fmt.Fprintln(os.Stderr, "Error: --limit must be between 10 and 1000")
		os.Exit(1)
	}

	sortKeys, err := seatsSort.ParseSortKeys(sortStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := api.NewClient(os.Getenv("SEATS_AERO_API_KEY"))

	resp, err := client.Search(api.SearchParams{
		OriginAirport:      from,
		DestinationAirport: to,
		StartDate:          date,
		EndDate:            endDate,
		Cabins:             cabin,
		Sources:            program,
		Carriers:           carrier,
		OnlyDirectFlights:  direct,
		Take:               limit,
		Skip:               skip,
	})
	if err != nil {
		return handleAPIError(err)
	}

	var rows []model.FlatRow
	for _, a := range resp.Data {
		rows = append(rows, a.Flatten()...)
	}

	seatsSort.SortFlatRows(rows, sortKeys)

	if pretty {
		fmt.Print(format.PrettySearch(from, to, date, rows, resp.HasMore))
	} else {
		fmt.Print(format.FormatSearchMarkdown(from, to, date, rows, resp.HasMore))
	}

	return nil
}
