package format

import (
	"strings"
	"testing"

	"github.com/derek/seats-cli/internal/model"
)

func TestFormatSearchMarkdown(t *testing.T) {
	rows := []model.FlatRow{
		{
			ID:      "abc123",
			Date:    "2026-04-03",
			Program: "united",
			Cabin:   "economy",
			Miles:   12500,
			Airline: "UA",
			Direct:  true,
			Seats:   3,
		},
	}

	result := FormatSearchMarkdown("SFO", "BOS", "2026-04-03", rows, false)

	checks := []string{
		"## Award Search: SFO -> BOS",
		"abc123",
		"12,500",
		"united",
		"| yes |",
		"1 results",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, result)
		}
	}
}

func TestFormatSearchMarkdownNoResults(t *testing.T) {
	result := FormatSearchMarkdown("SFO", "BOS", "2026-04-03", nil, false)

	if !strings.Contains(result, "No results found") {
		t.Errorf("expected output to contain 'No results found', got:\n%s", result)
	}
}

func TestFormatProgramsMarkdown(t *testing.T) {
	result := FormatProgramsMarkdown()

	checks := []string{
		"## Available Programs",
		"united",
		"United MileagePlus",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, result)
		}
	}
}

func TestFormatLiveMarkdown(t *testing.T) {
	results := []model.LiveResult{
		{
			Trip: model.Trip{
				FlightNumbers:       "UA123",
				Cabin:               "economy",
				MileageCost:         12500,
				TotalTaxes:          560,
				TaxesCurrencySymbol: "$",
				Stops:               0,
				TotalDuration:       330,
				DepartsAt:           "2026-04-03T08:00:00Z",
				ArrivesAt:           "2026-04-03T16:30:00Z",
				RemainingSeats:      3,
			},
		},
	}

	result := FormatLiveMarkdown("SFO", "BOS", "2026-04-03", "united", results)

	checks := []string{
		"## Live Search: SFO -> BOS",
		"UA123",
		"12,500",
		"5h 30m",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, result)
		}
	}
}

func TestFormatTripMarkdown(t *testing.T) {
	resp := &model.TripResponse{
		Data: []model.Trip{
			{
				Cabin:               "economy",
				MileageCost:         12500,
				TotalTaxes:          560,
				TaxesCurrency:       "USD",
				TaxesCurrencySymbol: "$",
				TotalDuration:       330,
				Stops:               0,
				AvailabilitySegments: []model.Segment{
					{
						FlightNumber:       "UA123",
						OriginAirport:      "SFO",
						DestinationAirport: "BOS",
						DepartsAt:          "2026-04-03T08:00:00Z",
						ArrivesAt:          "2026-04-03T16:30:00Z",
						AircraftName:       "Boeing 737-900",
						Order:              1,
					},
				},
			},
		},
		BookingLinks: []model.BookingLink{
			{
				Label:   "United MileagePlus",
				Link:    "https://united.com",
				Primary: true,
			},
		},
	}

	result := FormatTripMarkdown("abc123", resp)

	checks := []string{
		"## Trip Details",
		"economy",
		"12,500",
		"Booking Links",
		"United MileagePlus",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, result)
		}
	}
}

func TestFormatRoutesMarkdown(t *testing.T) {
	routes := []model.Route{
		{
			OriginAirport:      "SFO",
			DestinationAirport: "NRT",
			OriginRegion:       "North America",
			DestinationRegion:  "Asia",
			Distance:           5130,
			NumDaysOut:         60,
		},
	}

	result := FormatRoutesMarkdown("aeroplan", routes)

	checks := []string{
		"## Routes: aeroplan",
		"SFO",
		"5,130",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, result)
		}
	}
}

func TestFormatAvailabilityMarkdown(t *testing.T) {
	data := []model.Availability{
		{
			ID:   "abc123",
			Date: "2026-04-03",
			Route: model.Route{
				OriginAirport:      "DFW",
				DestinationAirport: "NRT",
			},
			YAvailable:      true,
			JAvailable:      true,
			FAvailable:      false,
			YMileageCost:    "35000",
			JMileageCost:    "70000",
			YRemainingSeats: 3,
			JRemainingSeats: 2,
			YDirect:         true,
			JDirect:         true,
			Source:          "american",
		},
	}

	result := FormatAvailabilityMarkdown("american", "2026-04-01", "2026-04-30", "", data, false)

	checks := []string{
		"## Bulk Availability: american",
		"DFW -> NRT",
		"35,000 (3)",
		"70,000 (2)",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, result)
		}
	}
}
