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

var availabilityCmd = &cobra.Command{
	Use:   "availability",
	Short: "Bulk program availability",
	Long:  "Retrieve all cached availability for a mileage program. Does not support --from/--to; use --origin-region/--dest-region or 'seats search' for airport-specific queries.",
	RunE:  runAvailability,
}

func init() {
	availabilityCmd.Flags().String("program", "", "Mileage program")
	availabilityCmd.Flags().String("transfer-partner", "", "Filter by transfer partner(s), comma-separated (e.g. chase,amex). Expands to matching programs; intersects with --program if both given.")
	availabilityCmd.Flags().String("cabin", "", "economy, premium, business, first")
	availabilityCmd.Flags().String("date", "", "Start date")
	availabilityCmd.Flags().String("end-date", "", "End date")
	availabilityCmd.Flags().String("origin-region", "", "North America, South America, Africa, Asia, Europe, Oceania")
	availabilityCmd.Flags().String("dest-region", "", "Destination region")
	availabilityCmd.Flags().String("sort", "", "Sort keys")
	availabilityCmd.Flags().Int("limit", 50, "Max results (10-1000)")
	availabilityCmd.Flags().Int("skip", 0, "Number of results to skip")

	// Note: --program is required unless --transfer-partner is specified; validated at runtime.

	rootCmd.AddCommand(availabilityCmd)
}

func runAvailability(cmd *cobra.Command, args []string) error {
	program, _ := cmd.Flags().GetString("program")
	transferPartner, _ := cmd.Flags().GetString("transfer-partner")
	cabin, _ := cmd.Flags().GetString("cabin")

	if transferPartner != "" {
		expanded, err := expandTransferPartners(transferPartner, program)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		program = expanded
	} else if program == "" {
		fmt.Fprintln(os.Stderr, "Error: required flag(s) \"program\" not set (or use --transfer-partner)")
		os.Exit(1)
	}
	date, _ := cmd.Flags().GetString("date")
	endDate, _ := cmd.Flags().GetString("end-date")
	originRegion, _ := cmd.Flags().GetString("origin-region")
	destRegion, _ := cmd.Flags().GetString("dest-region")
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
	resp, err := client.Availability(api.AvailabilityParams{
		Source:            program,
		Cabin:             cabin,
		StartDate:         date,
		EndDate:           endDate,
		OriginRegion:      originRegion,
		DestinationRegion: destRegion,
		Take:              limit,
		Skip:              skip,
	})
	if err != nil {
		return handleAPIError(err)
	}

	if sortStr != "" {
		var flatRows []model.FlatRow
		for _, a := range resp.Data {
			flatRows = append(flatRows, a.Flatten()...)
		}
		seatsSort.SortFlatRows(flatRows, sortKeys)
		if pretty {
			fmt.Print(format.PrettySearch(program, "(bulk)", date, flatRows, resp.HasMore))
		} else {
			fmt.Print(format.FormatSearchMarkdown(program, "(bulk)", date, flatRows, resp.HasMore))
		}
		return nil
	}

	if pretty {
		fmt.Print(format.PrettyAvailability(program, date, endDate, cabin, resp.Data, resp.HasMore))
	} else {
		fmt.Print(format.FormatAvailabilityMarkdown(program, date, endDate, cabin, resp.Data, resp.HasMore))
	}
	return nil
}
