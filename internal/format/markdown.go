package format

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/derek/seats-cli/internal/model"
)

// truncateID truncates an ID to at most 10 characters.
func truncateID(id string) string {
	if len(id) <= 10 {
		return id
	}
	return id[:10]
}

// formatNumber formats an integer with comma separators (e.g. 12500 → "12,500").
func formatNumber(n int) string {
	s := strconv.Itoa(n)
	if n < 0 {
		s = strconv.Itoa(-n)
	}
	var result []byte
	for i, ch := range s {
		pos := len(s) - i
		if i > 0 && pos%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(ch))
	}
	if n < 0 {
		return "-" + string(result)
	}
	return string(result)
}

// formatDuration converts total minutes into a human-readable string like "5h 30m".
func formatDuration(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}

// formatTime parses an RFC3339 timestamp and returns "15:04" (UTC).
func formatTime(isoTime string) string {
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		return isoTime
	}
	return t.UTC().Format("15:04")
}

// formatTaxes converts cents to a human-readable dollar string like "$5.60".
func formatTaxes(cents int, symbol string) string {
	dollars := cents / 100
	remainder := cents % 100
	return fmt.Sprintf("%s%d.%02d", symbol, dollars, remainder)
}

// formatCabinCell returns a formatted cell for an availability cabin column.
// It returns "—" if not available, or "35,000 (3)" otherwise.
func formatCabinCell(available bool, milesCost string, seats int) string {
	if !available {
		return "—"
	}
	miles := model.ParseMiles(milesCost)
	return fmt.Sprintf("%s (%d)", formatNumber(miles), seats)
}

// FormatSearchMarkdown renders a search results table as markdown.
func FormatSearchMarkdown(from, to, date string, rows []model.FlatRow, hasMore bool) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Award Search: %s -> %s\n\n", from, to)
	if date != "" {
		fmt.Fprintf(&sb, "**Date:** %s\n\n", date)
	}

	if len(rows) == 0 {
		sb.WriteString("No results found\n")
		return sb.String()
	}

	sb.WriteString("| ID | Date | Program | Cabin | Miles | Airline | Direct | Seats |\n")
	sb.WriteString("|----|------|---------|-------|-------|---------|--------|-------|\n")

	for _, r := range rows {
		direct := "no"
		if r.Direct {
			direct = "yes"
		}
		fmt.Fprintf(&sb, "| %s | %s | %s | %s | %s | %s | %s | %d |\n",
			truncateID(r.ID),
			r.Date,
			r.Program,
			r.Cabin,
			formatNumber(r.Miles),
			r.Airline,
			direct,
			r.Seats,
		)
	}

	sb.WriteString("\n")
	count := len(rows)
	fmt.Fprintf(&sb, "%d results", count)
	if hasMore {
		sb.WriteString(" (more available)")
	}
	sb.WriteString("\n")

	return sb.String()
}

// FormatLiveMarkdown renders live search results as markdown.
func FormatLiveMarkdown(from, to, date, program string, results []model.LiveResult) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Live Search: %s -> %s\n\n", from, to)
	if date != "" {
		fmt.Fprintf(&sb, "**Date:** %s  \n", date)
	}
	if program != "" {
		fmt.Fprintf(&sb, "**Program:** %s\n\n", program)
	}

	if len(results) == 0 {
		sb.WriteString("No results found\n")
		return sb.String()
	}

	sb.WriteString("| Flight | Cabin | Miles | Taxes | Stops | Duration | Departs | Arrives | Seats |\n")
	sb.WriteString("|--------|-------|-------|-------|-------|----------|---------|---------|-------|\n")

	for _, r := range results {
		fmt.Fprintf(&sb, "| %s | %s | %s | %s | %d | %s | %s | %s | %d |\n",
			r.FlightNumbers,
			r.Cabin,
			formatNumber(r.MileageCost),
			formatTaxes(r.TotalTaxes, r.TaxesCurrencySymbol),
			r.Stops,
			formatDuration(r.TotalDuration),
			formatTime(r.DepartsAt),
			formatTime(r.ArrivesAt),
			r.RemainingSeats,
		)
	}

	return sb.String()
}

// FormatAvailabilityMarkdown renders bulk availability data as markdown.
func FormatAvailabilityMarkdown(program, startDate, endDate, cabin string, data []model.Availability, hasMore bool) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Bulk Availability: %s\n\n", program)
	if startDate != "" || endDate != "" {
		fmt.Fprintf(&sb, "**Dates:** %s to %s\n\n", startDate, endDate)
	}
	if cabin != "" {
		fmt.Fprintf(&sb, "**Cabin:** %s\n\n", cabin)
	}

	if len(data) == 0 {
		sb.WriteString("No results found\n")
		return sb.String()
	}

	sb.WriteString("| ID | Date | Route | Economy | Business | First | Direct |\n")
	sb.WriteString("|----|------|-------|---------|----------|-------|--------|\n")

	for _, a := range data {
		route := fmt.Sprintf("%s -> %s", a.Route.OriginAirport, a.Route.DestinationAirport)
		direct := "no"
		if a.YDirect || a.JDirect || a.FDirect {
			direct = "yes"
		}
		fmt.Fprintf(&sb, "| %s | %s | %s | %s | %s | %s | %s |\n",
			truncateID(a.ID),
			a.Date,
			route,
			formatCabinCell(a.YAvailable, a.YMileageCost, a.YRemainingSeats),
			formatCabinCell(a.JAvailable, a.JMileageCost, a.JRemainingSeats),
			formatCabinCell(a.FAvailable, a.FMileageCost, a.FRemainingSeats),
			direct,
		)
	}

	sb.WriteString("\n")
	count := len(data)
	fmt.Fprintf(&sb, "%d results", count)
	if hasMore {
		sb.WriteString(" (more available)")
	}
	sb.WriteString("\n")

	return sb.String()
}

// FormatTripMarkdown renders detailed trip information as markdown.
func FormatTripMarkdown(id string, resp *model.TripResponse) string {
	var sb strings.Builder

	sb.WriteString("## Trip Details\n\n")
	if id != "" {
		fmt.Fprintf(&sb, "**Availability ID:** %s\n\n", id)
	}

	for i, trip := range resp.Data {
		fmt.Fprintf(&sb, "### Option %d: %s — %s miles\n\n", i+1, trip.Cabin, formatNumber(trip.MileageCost))
		fmt.Fprintf(&sb, "**Taxes:** %s  \n", formatTaxes(trip.TotalTaxes, trip.TaxesCurrencySymbol))
		fmt.Fprintf(&sb, "**Duration:** %s  \n", formatDuration(trip.TotalDuration))
		fmt.Fprintf(&sb, "**Stops:** %d\n\n", trip.Stops)

		if len(trip.AvailabilitySegments) > 0 {
			sb.WriteString("| # | Flight | From | To | Departs | Arrives | Aircraft |\n")
			sb.WriteString("|---|--------|------|----|---------|---------|----------|\n")
			for _, seg := range trip.AvailabilitySegments {
				fmt.Fprintf(&sb, "| %d | %s | %s | %s | %s | %s | %s |\n",
					seg.Order,
					seg.FlightNumber,
					seg.OriginAirport,
					seg.DestinationAirport,
					formatTime(seg.DepartsAt),
					formatTime(seg.ArrivesAt),
					seg.AircraftName,
				)
			}
			sb.WriteString("\n")
		}
	}

	if len(resp.BookingLinks) > 0 {
		sb.WriteString("## Booking Links\n\n")
		for _, link := range resp.BookingLinks {
			label := link.Label
			if link.Primary {
				label += " ⭐"
			}
			fmt.Fprintf(&sb, "- [%s](%s)\n", label, link.Link)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatRoutesMarkdown renders available routes as markdown.
func FormatRoutesMarkdown(program string, routes []model.Route) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Routes: %s\n\n", program)

	if len(routes) == 0 {
		sb.WriteString("No routes found\n")
		return sb.String()
	}

	sb.WriteString("| Origin | Destination | Origin Region | Dest Region | Distance | Days Out |\n")
	sb.WriteString("|--------|-------------|---------------|-------------|----------|----------|\n")

	for _, r := range routes {
		fmt.Fprintf(&sb, "| %s | %s | %s | %s | %s | %d |\n",
			r.OriginAirport,
			r.DestinationAirport,
			r.OriginRegion,
			r.DestinationRegion,
			formatNumber(r.Distance),
			r.NumDaysOut,
		)
	}

	return sb.String()
}

// FormatProgramsMarkdown renders the list of supported loyalty programs as markdown.
func FormatProgramsMarkdown() string {
	var sb strings.Builder

	sb.WriteString("## Available Programs\n\n")
	sb.WriteString("| Program | Name |\n")
	sb.WriteString("|---------|------|\n")

	for _, p := range model.Programs {
		fmt.Fprintf(&sb, "| %s | %s |\n", p.ID, p.Name)
	}

	return sb.String()
}
