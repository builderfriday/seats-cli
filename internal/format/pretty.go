package format

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/derek/seats-cli/internal/model"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).MarginBottom(1)
)

// newTable creates a lipgloss table with NormalBorder, dim border color, styled headers.
func newTable(headers []string, rows [][]string) string {
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		Headers(headers...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	for _, r := range rows {
		t.Row(r...)
	}

	return t.String()
}

// PrettySearch renders a search results table with styled terminal output.
func PrettySearch(from, to, date string, rows []model.FlatRow, hasMore bool) string {
	var sb strings.Builder

	title := fmt.Sprintf("Award Search: %s → %s", from, to)
	if date != "" {
		title += fmt.Sprintf("  |  Date: %s", date)
	}
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n")

	if len(rows) == 0 {
		sb.WriteString("No results found\n")
		return sb.String()
	}

	headers := []string{"ID", "Date", "Program", "Cabin", "Miles", "Airline", "Direct", "Seats"}
	tableRows := make([][]string, 0, len(rows))
	for _, r := range rows {
		direct := "✗"
		if r.Direct {
			direct = "✓"
		}
		tableRows = append(tableRows, []string{
			truncateID(r.ID),
			r.Date,
			r.Program,
			r.Cabin,
			formatNumber(r.Miles),
			r.Airline,
			direct,
			fmt.Sprintf("%d", r.Seats),
		})
	}

	sb.WriteString(newTable(headers, tableRows))
	sb.WriteString("\n")

	count := len(rows)
	fmt.Fprintf(&sb, "%d results", count)
	if hasMore {
		sb.WriteString(" (more available)")
	}
	sb.WriteString("\n")

	return sb.String()
}

// PrettyLive renders live search results with styled terminal output.
func PrettyLive(from, to, date, program string, results []model.LiveResult) string {
	var sb strings.Builder

	title := fmt.Sprintf("Live Search: %s → %s", from, to)
	if date != "" {
		title += fmt.Sprintf("  |  Date: %s", date)
	}
	if program != "" {
		title += fmt.Sprintf("  |  Program: %s", program)
	}
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n")

	if len(results) == 0 {
		sb.WriteString("No results found\n")
		return sb.String()
	}

	headers := []string{"Flight", "Cabin", "Miles", "Taxes", "Stops", "Duration", "Departs", "Arrives", "Seats"}
	tableRows := make([][]string, 0, len(results))
	for _, r := range results {
		tableRows = append(tableRows, []string{
			r.FlightNumbers,
			r.Cabin,
			formatNumber(r.MileageCost),
			formatTaxes(r.TotalTaxes, r.TaxesCurrencySymbol),
			fmt.Sprintf("%d", r.Stops),
			formatDuration(r.TotalDuration),
			formatTime(r.DepartsAt),
			formatTime(r.ArrivesAt),
			fmt.Sprintf("%d", r.RemainingSeats),
		})
	}

	sb.WriteString(newTable(headers, tableRows))
	sb.WriteString("\n")

	return sb.String()
}

// PrettyAvailability renders bulk availability data with styled terminal output.
func PrettyAvailability(program, startDate, endDate, cabin string, data []model.Availability, hasMore bool) string {
	var sb strings.Builder

	title := fmt.Sprintf("Bulk Availability: %s", program)
	if startDate != "" || endDate != "" {
		title += fmt.Sprintf("  |  Dates: %s to %s", startDate, endDate)
	}
	if cabin != "" {
		title += fmt.Sprintf("  |  Cabin: %s", cabin)
	}
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n")

	if len(data) == 0 {
		sb.WriteString("No results found\n")
		return sb.String()
	}

	headers := []string{"ID", "Date", "Route", "Economy", "Business", "First", "Direct"}
	tableRows := make([][]string, 0, len(data))
	for _, a := range data {
		route := fmt.Sprintf("%s → %s", a.Route.OriginAirport, a.Route.DestinationAirport)
		direct := "✗"
		if a.YDirect || a.JDirect || a.FDirect {
			direct = "✓"
		}
		tableRows = append(tableRows, []string{
			truncateID(a.ID),
			a.Date,
			route,
			formatCabinCell(a.YAvailable, a.YMileageCost, a.YRemainingSeats),
			formatCabinCell(a.JAvailable, a.JMileageCost, a.JRemainingSeats),
			formatCabinCell(a.FAvailable, a.FMileageCost, a.FRemainingSeats),
			direct,
		})
	}

	sb.WriteString(newTable(headers, tableRows))
	sb.WriteString("\n")

	count := len(data)
	fmt.Fprintf(&sb, "%d results", count)
	if hasMore {
		sb.WriteString(" (more available)")
	}
	sb.WriteString("\n")

	return sb.String()
}

// PrettyTrip renders detailed trip information with styled terminal output.
func PrettyTrip(id string, resp *model.TripResponse) string {
	var sb strings.Builder

	title := "Trip Details"
	if id != "" {
		title += fmt.Sprintf("  |  ID: %s", id)
	}
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n")

	for i, trip := range resp.Data {
		optTitle := fmt.Sprintf("Option %d: %s — %s miles", i+1, trip.Cabin, formatNumber(trip.MileageCost))
		sb.WriteString(headerStyle.Render(optTitle))
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "Taxes: %s  |  Duration: %s  |  Stops: %d\n\n",
			formatTaxes(trip.TotalTaxes, trip.TaxesCurrencySymbol),
			formatDuration(trip.TotalDuration),
			trip.Stops,
		)

		if len(trip.AvailabilitySegments) > 0 {
			headers := []string{"#", "Flight", "From", "To", "Departs", "Arrives", "Aircraft"}
			tableRows := make([][]string, 0, len(trip.AvailabilitySegments))
			for _, seg := range trip.AvailabilitySegments {
				tableRows = append(tableRows, []string{
					fmt.Sprintf("%d", seg.Order),
					seg.FlightNumber,
					seg.OriginAirport,
					seg.DestinationAirport,
					formatTime(seg.DepartsAt),
					formatTime(seg.ArrivesAt),
					seg.AircraftName,
				})
			}
			sb.WriteString(newTable(headers, tableRows))
			sb.WriteString("\n")
		}
	}

	if len(resp.BookingLinks) > 0 {
		sb.WriteString(headerStyle.Render("Booking Links"))
		sb.WriteString("\n")
		for _, link := range resp.BookingLinks {
			label := link.Label
			if link.Primary {
				label += " ⭐"
			}
			fmt.Fprintf(&sb, "  %s: %s\n", label, link.Link)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// PrettyRoutes renders available routes with styled terminal output.
func PrettyRoutes(program string, routes []model.Route) string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render(fmt.Sprintf("Routes: %s", program)))
	sb.WriteString("\n")

	if len(routes) == 0 {
		sb.WriteString("No routes found\n")
		return sb.String()
	}

	headers := []string{"Origin", "Destination", "Origin Region", "Dest Region", "Distance", "Days Out"}
	tableRows := make([][]string, 0, len(routes))
	for _, r := range routes {
		tableRows = append(tableRows, []string{
			r.OriginAirport,
			r.DestinationAirport,
			r.OriginRegion,
			r.DestinationRegion,
			formatNumber(r.Distance),
			fmt.Sprintf("%d", r.NumDaysOut),
		})
	}

	sb.WriteString(newTable(headers, tableRows))
	sb.WriteString("\n")

	fmt.Fprintf(&sb, "%d routes\n", len(routes))

	return sb.String()
}

// PrettyPrograms renders the list of supported loyalty programs with styled terminal output.
func PrettyPrograms() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Available Programs"))
	sb.WriteString("\n")

	headers := []string{"Program", "Name"}
	tableRows := make([][]string, 0, len(model.Programs))
	for _, p := range model.Programs {
		tableRows = append(tableRows, []string{p.ID, p.Name})
	}

	sb.WriteString(newTable(headers, tableRows))
	sb.WriteString("\n")

	fmt.Fprintf(&sb, "%d programs\n", len(model.Programs))

	return sb.String()
}
