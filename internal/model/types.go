package model

import (
	"unicode"
)

// SearchResponse is the top-level response from the /availability endpoint.
type SearchResponse struct {
	Data    []Availability `json:"data"`
	Count   int            `json:"count"`
	HasMore bool           `json:"hasMore"`
	Cursor  int            `json:"cursor"`
}

// Availability represents a single cached availability record for one route/date/program combination.
type Availability struct {
	ID      string `json:"ID"`
	RouteID string `json:"RouteID"`
	Route   Route  `json:"Route"`
	Date    string `json:"Date"`
	Source  string `json:"Source"`

	// Per-cabin availability fields.
	YAvailable      bool   `json:"YAvailable"`
	YMileageCost    string `json:"YMileageCost"`
	YRemainingSeats int    `json:"YRemainingSeats"`
	YAirlines       string `json:"YAirlines"`
	YDirect         bool   `json:"YDirect"`

	WAvailable      bool   `json:"WAvailable"`
	WMileageCost    string `json:"WMileageCost"`
	WRemainingSeats int    `json:"WRemainingSeats"`
	WAirlines       string `json:"WAirlines"`
	WDirect         bool   `json:"WDirect"`

	JAvailable      bool   `json:"JAvailable"`
	JMileageCost    string `json:"JMileageCost"`
	JRemainingSeats int    `json:"JRemainingSeats"`
	JAirlines       string `json:"JAirlines"`
	JDirect         bool   `json:"JDirect"`

	FAvailable      bool   `json:"FAvailable"`
	FMileageCost    string `json:"FMileageCost"`
	FRemainingSeats int    `json:"FRemainingSeats"`
	FAirlines       string `json:"FAirlines"`
	FDirect         bool   `json:"FDirect"`

	CreatedAt string `json:"CreatedAt"`
	UpdatedAt string `json:"UpdatedAt"`
}

// Flatten converts the per-cabin fields of an Availability into a slice of FlatRow,
// one row per cabin that is available. Cabins that are not available are skipped.
func (a Availability) Flatten() []FlatRow {
	type cabinDef struct {
		code      string
		name      string
		available bool
		cost      string
		seats     int
		airline   string
		direct    bool
	}

	cabins := []cabinDef{
		{"Y", "economy", a.YAvailable, a.YMileageCost, a.YRemainingSeats, a.YAirlines, a.YDirect},
		{"W", "premium", a.WAvailable, a.WMileageCost, a.WRemainingSeats, a.WAirlines, a.WDirect},
		{"J", "business", a.JAvailable, a.JMileageCost, a.JRemainingSeats, a.JAirlines, a.JDirect},
		{"F", "first", a.FAvailable, a.FMileageCost, a.FRemainingSeats, a.FAirlines, a.FDirect},
	}

	var rows []FlatRow
	for _, c := range cabins {
		if !c.available {
			continue
		}

		stops := -1 // unknown for cached search
		if c.direct {
			stops = 0
		}

		rows = append(rows, FlatRow{
			ID:      a.ID,
			Date:    a.Date,
			Program: a.Source,
			Cabin:   c.name,
			Miles:   ParseMiles(c.cost),
			Taxes:   0, // not available in cached search
			Airline: c.airline,
			Direct:  c.direct,
			Stops:   stops,
			Seats:   c.seats,
			Route:   a.Route,
		})
	}
	return rows
}

// Route represents an origin-to-destination route.
type Route struct {
	ID                 string `json:"ID"`
	OriginAirport      string `json:"OriginAirport"`
	OriginRegion       string `json:"OriginRegion"`
	DestinationAirport string `json:"DestinationAirport"`
	DestinationRegion  string `json:"DestinationRegion"`
	NumDaysOut         int    `json:"NumDaysOut"`
	Distance           int    `json:"Distance"`
	Source             string `json:"Source"`
}

// TripResponse is the top-level response from the /trips endpoint.
type TripResponse struct {
	Data                   []Trip        `json:"data"`
	OriginCoordinates      *Coordinates  `json:"originCoordinates"`
	DestinationCoordinates *Coordinates  `json:"destinationCoordinates"`
	BookingLinks           []BookingLink `json:"bookingLinks"`
}

// Trip represents a full itinerary with one or more segments.
type Trip struct {
	ID                   string    `json:"ID"`
	RouteID              string    `json:"RouteID"`
	AvailabilityID       string    `json:"AvailabilityID"`
	AvailabilitySegments []Segment `json:"AvailabilitySegments"`
	TotalDuration        int       `json:"TotalDuration"`
	Stops                int       `json:"Stops"`
	Carriers             string    `json:"Carriers"`
	RemainingSeats       int       `json:"RemainingSeats"`
	MileageCost          int       `json:"MileageCost"`
	TotalTaxes           int       `json:"TotalTaxes"`
	TaxesCurrency        string    `json:"TaxesCurrency"`
	TaxesCurrencySymbol  string    `json:"TaxesCurrencySymbol"`
	AllianceCost         int       `json:"AllianceCost"`
	FlightNumbers        string    `json:"FlightNumbers"`
	DepartsAt            string    `json:"DepartsAt"`
	ArrivesAt            string    `json:"ArrivesAt"`
	Cabin                string    `json:"Cabin"`
	Source               string    `json:"Source"`
	CreatedAt            string    `json:"CreatedAt"`
	UpdatedAt            string    `json:"UpdatedAt"`
}

// Segment represents a single flight leg within a trip.
type Segment struct {
	ID                 string `json:"ID"`
	FlightNumber       string `json:"FlightNumber"`
	Distance           int    `json:"Distance"`
	FareClass          string `json:"FareClass"`
	AircraftName       string `json:"AircraftName"`
	AircraftCode       string `json:"AircraftCode"`
	OriginAirport      string `json:"OriginAirport"`
	DestinationAirport string `json:"DestinationAirport"`
	DepartsAt          string `json:"DepartsAt"`
	ArrivesAt          string `json:"ArrivesAt"`
	Source             string `json:"Source"`
	Cabin              string `json:"Cabin"`
	Order              int    `json:"Order"`
}

// Coordinates holds a geographic lat/lon pair.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// BookingLink is a link to a booking page with a human-readable label.
type BookingLink struct {
	Label   string `json:"label"`
	Link    string `json:"link"`
	Primary bool   `json:"primary"`
}

// LiveResponse is the top-level response from the /live endpoint.
type LiveResponse struct {
	Results []LiveResult `json:"results"`
}

// LiveResult is a Trip with an additional Filtered flag.
type LiveResult struct {
	Trip
	Filtered bool `json:"Filtered"`
}

// FlatRow is a denormalized, display-ready representation of one cabin option.
type FlatRow struct {
	ID      string
	Date    string
	Program string
	Cabin   string
	Miles   int
	Taxes   int
	Airline string
	Direct  bool
	Stops   int
	Seats   int
	Route   Route
}

// ParseMiles parses a mileage string that may contain formatting characters (e.g. "12,500")
// and returns the integer value. Non-digit characters are ignored.
func ParseMiles(s string) int {
	n := 0
	for _, ch := range s {
		if unicode.IsDigit(ch) {
			n = n*10 + int(ch-'0')
		}
	}
	return n
}
