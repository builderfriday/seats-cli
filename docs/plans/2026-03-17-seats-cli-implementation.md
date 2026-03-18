# seats CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI tool that wraps the Seats.aero Partner API for award flight searching, with markdown output by default and pretty terminal rendering via Charm.

**Architecture:** Thin API wrapper — 6 subcommands mapping 1:1 to API endpoints. Bottom-up build: models → API client → formatters → sort engine → commands → main. Each layer is independently testable.

**Tech Stack:** Go, Cobra (CLI framework), Lip Gloss + Charm Table (terminal styling), standard library net/http for API calls.

**Spec:** `docs/specs/2026-03-17-seats-cli-design.md`

---

## File Structure

```
seats-cli/
├── main.go                        # Entry point — calls cmd.Execute()
├── cmd/
│   ├── root.go                    # Root command, --pretty global flag, env var validation
│   ├── search.go                  # seats search — cached award search
│   ├── live.go                    # seats live — real-time search
│   ├── availability.go            # seats availability — bulk program availability
│   ├── routes.go                  # seats routes — list program routes
│   ├── trip.go                    # seats trip <id> — trip details + booking links
│   └── programs.go                # seats programs — list mileage programs
├── internal/
│   ├── api/
│   │   └── client.go              # HTTP client, auth header, all endpoint methods, error mapping
│   ├── model/
│   │   ├── types.go               # API response structs (search, availability, trip, route)
│   │   └── programs.go            # Hardcoded program list
│   ├── format/
│   │   ├── markdown.go            # Markdown table rendering for each output type
│   │   └── pretty.go              # Lip Gloss styled table rendering
│   └── sort/
│       └── sort.go                # Multi-key sort parsing and execution
├── go.mod
└── go.sum
```

---

### Task 1: Project Scaffolding

**Files:**
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `go.mod`

- [ ] **Step 1: Initialize Go module**

```bash
cd ~/repo/seats-cli
go mod init github.com/derek/seats-cli
```

- [ ] **Step 2: Install dependencies**

```bash
cd ~/repo/seats-cli
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/lipgloss/table@latest
```

- [ ] **Step 3: Create main.go**

```go
// main.go
package main

import (
	"github.com/derek/seats-cli/cmd"
)

func main() {
	// cmd.Execute() handles os.Exit() with proper exit codes internally.
	// Cobra prints usage errors to stderr and exits with code 1.
	cmd.Execute()
}
```

- [ ] **Step 4: Create cmd/root.go with global flags**

```go
// cmd/root.go
package cmd

import (
	"fmt"
	"os"

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
```

- [ ] **Step 5: Verify it compiles**

```bash
cd ~/repo/seats-cli && go build ./...
```
Expected: clean build, no errors.

- [ ] **Step 6: Commit**

```bash
cd ~/repo/seats-cli
git init
git add main.go cmd/root.go go.mod go.sum
git commit -m "feat: scaffold seats CLI with cobra root command"
```

---

### Task 2: API Response Models

**Files:**
- Create: `internal/model/types.go`
- Create: `internal/model/programs.go`
- Create: `internal/model/types_test.go`

- [ ] **Step 1: Write test for JSON unmarshaling of search response**

```go
// internal/model/types_test.go
package model

import (
	"encoding/json"
	"testing"
)

func TestSearchResponseUnmarshal(t *testing.T) {
	raw := `{
		"data": [{
			"ID": "abc123",
			"RouteID": "route1",
			"Route": {
				"ID": "route1",
				"OriginAirport": "SFO",
				"DestinationAirport": "BOS",
				"OriginRegion": "North America",
				"DestinationRegion": "North America",
				"NumDaysOut": 30,
				"Distance": 2700,
				"Source": "united"
			},
			"Date": "2026-04-03",
			"YAvailable": true,
			"JAvailable": true,
			"FAvailable": false,
			"WAvailable": false,
			"YMileageCost": "12500",
			"JMileageCost": "35000",
			"FMileageCost": "0",
			"WMileageCost": "0",
			"YRemainingSeats": 3,
			"JRemainingSeats": 2,
			"FRemainingSeats": 0,
			"WRemainingSeats": 0,
			"YAirlines": "UA",
			"JAirlines": "UA",
			"FAirlines": "",
			"WAirlines": "",
			"YDirect": true,
			"JDirect": true,
			"FDirect": false,
			"WDirect": false,
			"Source": "united"
		}],
		"count": 1,
		"hasMore": false,
		"cursor": 0
	}`

	var resp SearchResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Data))
	}
	r := resp.Data[0]
	if r.ID != "abc123" {
		t.Errorf("expected ID abc123, got %s", r.ID)
	}
	if r.Route.OriginAirport != "SFO" {
		t.Errorf("expected origin SFO, got %s", r.Route.OriginAirport)
	}
	if !r.YAvailable {
		t.Error("expected YAvailable true")
	}
	if r.YMileageCost != "12500" {
		t.Errorf("expected YMileageCost 12500, got %s", r.YMileageCost)
	}
	if r.YRemainingSeats != 3 {
		t.Errorf("expected YRemainingSeats 3, got %d", r.YRemainingSeats)
	}
}

func TestTripResponseUnmarshal(t *testing.T) {
	raw := `{
		"data": [{
			"ID": "trip1",
			"RouteID": "route1",
			"AvailabilityID": "abc123",
			"AvailabilitySegments": [{
				"ID": "seg1",
				"FlightNumber": "UA123",
				"OriginAirport": "SFO",
				"DestinationAirport": "BOS",
				"DepartsAt": "2026-04-03T08:00:00Z",
				"ArrivesAt": "2026-04-03T16:30:00Z",
				"AircraftName": "Boeing 737-900",
				"AircraftCode": "739",
				"FareClass": "Y",
				"Distance": 2700,
				"Source": "united",
				"Order": 1
			}],
			"TotalDuration": 330,
			"Stops": 0,
			"Carriers": "UA",
			"RemainingSeats": 3,
			"MileageCost": 12500,
			"TotalTaxes": 560,
			"TaxesCurrency": "USD",
			"TaxesCurrencySymbol": "$",
			"FlightNumbers": "UA123",
			"DepartsAt": "2026-04-03T08:00:00Z",
			"ArrivesAt": "2026-04-03T16:30:00Z",
			"Cabin": "economy",
			"Source": "united"
		}],
		"booking_links": [{
			"label": "United MileagePlus",
			"link": "https://united.com/book",
			"primary": true
		}]
	}`

	var resp TripResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 trip, got %d", len(resp.Data))
	}
	trip := resp.Data[0]
	if trip.TotalDuration != 330 {
		t.Errorf("expected duration 330, got %d", trip.TotalDuration)
	}
	if trip.MileageCost != 12500 {
		t.Errorf("expected mileage 12500, got %d", trip.MileageCost)
	}
	if len(trip.AvailabilitySegments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(trip.AvailabilitySegments))
	}
	if trip.AvailabilitySegments[0].FlightNumber != "UA123" {
		t.Errorf("expected flight UA123, got %s", trip.AvailabilitySegments[0].FlightNumber)
	}
	if len(resp.BookingLinks) != 1 {
		t.Fatalf("expected 1 booking link, got %d", len(resp.BookingLinks))
	}
	if !resp.BookingLinks[0].Primary {
		t.Error("expected primary booking link")
	}
}

func TestLiveResponseUnmarshal(t *testing.T) {
	raw := `{
		"results": [{
			"ID": "live1",
			"AvailabilitySegments": [{
				"FlightNumber": "UA123",
				"OriginAirport": "SFO",
				"DestinationAirport": "BOS",
				"DepartsAt": "2026-04-03T08:00:00Z",
				"ArrivesAt": "2026-04-03T16:30:00Z",
				"AircraftCode": "739",
				"AircraftName": "Boeing 737-900",
				"Source": "united",
				"Cabin": "economy"
			}],
			"TotalDuration": 330,
			"Stops": 0,
			"Carriers": "UA",
			"RemainingSeats": 3,
			"MileageCost": 12500,
			"TotalTaxes": 560,
			"TaxesCurrency": "USD",
			"TaxesCurrencySymbol": "$",
			"FlightNumbers": "UA123",
			"DepartsAt": "2026-04-03T08:00:00Z",
			"ArrivesAt": "2026-04-03T16:30:00Z",
			"Cabin": "economy",
			"Source": "united",
			"Filtered": false
		}]
	}`

	var resp LiveResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].MileageCost != 12500 {
		t.Errorf("expected 12500, got %d", resp.Results[0].MileageCost)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/repo/seats-cli && go test ./internal/model/ -v
```
Expected: FAIL — types not defined yet.

- [ ] **Step 3: Implement model types**

```go
// internal/model/types.go
package model

// SearchResponse is the response from GET /partnerapi/search and GET /partnerapi/availability.
type SearchResponse struct {
	Data    []Availability `json:"data"`
	Count   int            `json:"count"`
	HasMore bool           `json:"hasMore"`
	Cursor  int            `json:"cursor"`
}

// Availability represents a single availability record with per-cabin data.
type Availability struct {
	ID      string `json:"ID"`
	RouteID string `json:"RouteID"`
	Route   Route  `json:"Route"`
	Date    string `json:"Date"`
	Source  string `json:"Source"`

	// Per-cabin availability
	YAvailable bool `json:"YAvailable"`
	WAvailable bool `json:"WAvailable"`
	JAvailable bool `json:"JAvailable"`
	FAvailable bool `json:"FAvailable"`

	// Per-cabin mileage costs (string from API)
	YMileageCost string `json:"YMileageCost"`
	WMileageCost string `json:"WMileageCost"`
	JMileageCost string `json:"JMileageCost"`
	FMileageCost string `json:"FMileageCost"`

	// Per-cabin remaining seats
	YRemainingSeats int `json:"YRemainingSeats"`
	WRemainingSeats int `json:"WRemainingSeats"`
	JRemainingSeats int `json:"JRemainingSeats"`
	FRemainingSeats int `json:"FRemainingSeats"`

	// Per-cabin airlines
	YAirlines string `json:"YAirlines"`
	WAirlines string `json:"WAirlines"`
	JAirlines string `json:"JAirlines"`
	FAirlines string `json:"FAirlines"`

	// Per-cabin direct flight flags
	YDirect bool `json:"YDirect"`
	WDirect bool `json:"WDirect"`
	JDirect bool `json:"JDirect"`
	FDirect bool `json:"FDirect"`

	CreatedAt string `json:"CreatedAt"`
	UpdatedAt string `json:"UpdatedAt"`
}

// Route represents an airline route.
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

// TripResponse is the response from GET /partnerapi/trips/{id}.
type TripResponse struct {
	Data                   []Trip         `json:"data"`
	OriginCoordinates      *Coordinates   `json:"origin_coordinates"`
	DestinationCoordinates *Coordinates   `json:"destination_coordinates"`
	BookingLinks           []BookingLink  `json:"booking_links"`
}

// Trip represents a specific trip option.
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

// Segment represents a flight segment within a trip.
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

// Coordinates represents a geographic point.
type Coordinates struct {
	Lat float64 `json:"Lat"`
	Lon float64 `json:"Lon"`
}

// BookingLink represents a link to book a trip.
type BookingLink struct {
	Label   string `json:"label"`
	Link    string `json:"link"`
	Primary bool   `json:"primary"`
}

// LiveResponse is the response from POST /partnerapi/live.
type LiveResponse struct {
	Results []LiveResult `json:"results"`
}

// LiveResult represents a single live search result (trip-level data).
type LiveResult struct {
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
	FlightNumbers        string    `json:"FlightNumbers"`
	DepartsAt            string    `json:"DepartsAt"`
	ArrivesAt            string    `json:"ArrivesAt"`
	Cabin                string    `json:"Cabin"`
	Source               string    `json:"Source"`
	Filtered             bool      `json:"Filtered"`
}

// FlatRow is a flattened representation of one cabin from an Availability record.
// Used for search output where each cabin becomes its own row.
type FlatRow struct {
	ID       string
	Date     string
	Program  string
	Cabin    string // economy, premium, business, first
	Miles    int
	Taxes    int    // in cents, 0 for cached search (only populated from live results)
	Airline  string
	Direct   bool
	Stops    int    // 0 if Direct, -1 if unknown (cached search only knows direct/not)
	Seats    int
	Route    Route
}

// Flatten converts an Availability record into FlatRows, one per available cabin.
func (a *Availability) Flatten() []FlatRow {
	var rows []FlatRow
	type cabinInfo struct {
		code      string
		name      string
		available bool
		miles     string
		airline   string
		direct    bool
		seats     int
	}
	cabins := []cabinInfo{
		{"Y", "economy", a.YAvailable, a.YMileageCost, a.YAirlines, a.YDirect, a.YRemainingSeats},
		{"W", "premium", a.WAvailable, a.WMileageCost, a.WAirlines, a.WDirect, a.WRemainingSeats},
		{"J", "business", a.JAvailable, a.JMileageCost, a.JAirlines, a.JDirect, a.JRemainingSeats},
		{"F", "first", a.FAvailable, a.FMileageCost, a.FAirlines, a.FDirect, a.FRemainingSeats},
	}
	for _, c := range cabins {
		if !c.available {
			continue
		}
		miles := ParseMiles(c.miles)
		stops := -1 // unknown for cached search
		if c.direct {
			stops = 0
		}
		rows = append(rows, FlatRow{
			ID:      a.ID,
			Date:    a.Date,
			Program: a.Source,
			Cabin:   c.name,
			Miles:   miles,
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

// ParseMiles extracts an integer from a mileage cost string like "12500".
func ParseMiles(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
```

- [ ] **Step 4: Create programs list**

```go
// internal/model/programs.go
package model

// Program represents a mileage program supported by Seats.aero.
type Program struct {
	ID   string
	Name string
}

// Programs is the list of known mileage programs.
var Programs = []Program{
	{ID: "aeroplan", Name: "Air Canada Aeroplan"},
	{ID: "alaska", Name: "Alaska Mileage Plan"},
	{ID: "american", Name: "American AAdvantage"},
	{ID: "delta", Name: "Delta SkyMiles"},
	{ID: "etihad", Name: "Etihad Guest"},
	{ID: "eurobonus", Name: "SAS EuroBonus"},
	{ID: "flyingblue", Name: "Air France/KLM Flying Blue"},
	{ID: "lifemiles", Name: "Avianca LifeMiles"},
	{ID: "qantas", Name: "Qantas Frequent Flyer"},
	{ID: "united", Name: "United MileagePlus"},
	{ID: "velocity", Name: "Virgin Australia Velocity"},
	{ID: "virginatlantic", Name: "Virgin Atlantic Flying Club"},
}
```

- [ ] **Step 5: Run tests**

```bash
cd ~/repo/seats-cli && go test ./internal/model/ -v
```
Expected: all 3 tests PASS.

- [ ] **Step 6: Write test for Flatten**

Add to `internal/model/types_test.go`:

```go
func TestAvailabilityFlatten(t *testing.T) {
	a := Availability{
		ID:              "abc123",
		Date:            "2026-04-03",
		Source:          "united",
		YAvailable:      true,
		JAvailable:      true,
		FAvailable:      false,
		WAvailable:      false,
		YMileageCost:    "12500",
		JMileageCost:    "35000",
		YRemainingSeats: 3,
		JRemainingSeats: 2,
		YAirlines:       "UA",
		JAirlines:       "UA",
		YDirect:         true,
		JDirect:         true,
		Route: Route{
			OriginAirport:      "SFO",
			DestinationAirport: "BOS",
		},
	}

	rows := a.Flatten()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].Cabin != "economy" || rows[0].Miles != 12500 {
		t.Errorf("row 0: expected economy/12500, got %s/%d", rows[0].Cabin, rows[0].Miles)
	}
	if rows[0].Stops != 0 {
		t.Errorf("row 0: expected stops 0 (direct), got %d", rows[0].Stops)
	}
	if rows[1].Cabin != "business" || rows[1].Miles != 35000 {
		t.Errorf("row 1: expected business/35000, got %s/%d", rows[1].Cabin, rows[1].Miles)
	}
}

func TestAvailabilityFlattenNoneAvailable(t *testing.T) {
	a := Availability{}
	rows := a.Flatten()
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}
```

- [ ] **Step 7: Run tests**

```bash
cd ~/repo/seats-cli && go test ./internal/model/ -v
```
Expected: all 5 tests PASS.

- [ ] **Step 8: Commit**

```bash
cd ~/repo/seats-cli
git add internal/model/
git commit -m "feat: add API response models with JSON parsing and cabin flattening"
```

---

### Task 3: API Client

**Files:**
- Create: `internal/api/client.go`
- Create: `internal/api/client_test.go`

- [ ] **Step 1: Write test for API client using httptest**

```go
// internal/api/client_test.go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Partner-Authorization") != "test-key" {
			t.Errorf("expected auth header test-key, got %s", r.Header.Get("Partner-Authorization"))
		}
		if r.URL.Path != "/partnerapi/search" {
			t.Errorf("expected path /partnerapi/search, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("origin_airport") != "SFO" {
			t.Errorf("expected origin SFO, got %s", r.URL.Query().Get("origin_airport"))
		}
		if r.URL.Query().Get("destination_airport") != "BOS" {
			t.Errorf("expected dest BOS, got %s", r.URL.Query().Get("destination_airport"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"count":0,"hasMore":false,"cursor":0}`))
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL
	resp, err := c.Search(SearchParams{
		OriginAirport:      "SFO",
		DestinationAirport: "BOS",
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("expected count 0, got %d", resp.Count)
	}
}

func TestClientLive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/partnerapi/live" {
			t.Errorf("expected path /partnerapi/live, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL
	resp, err := c.Live(LiveParams{
		OriginAirport:      "SFO",
		DestinationAirport: "BOS",
		DepartureDate:      "2026-04-03",
		Source:             "united",
	})
	if err != nil {
		t.Fatalf("live search failed: %v", err)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}

func TestClientErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantExit   int
	}{
		{"bad request", 400, 1},
		{"unauthorized", 401, 2},
		{"rate limited", 429, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{}`))
			}))
			defer server.Close()

			c := NewClient("test-key")
			c.BaseURL = server.URL
			_, err := c.Search(SearchParams{OriginAirport: "SFO", DestinationAirport: "BOS"})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("expected *APIError, got %T", err)
			}
			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, apiErr.StatusCode)
			}
			if apiErr.ExitCode() != tt.wantExit {
				t.Errorf("expected exit code %d, got %d", tt.wantExit, apiErr.ExitCode())
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/repo/seats-cli && go test ./internal/api/ -v
```
Expected: FAIL — client not defined.

- [ ] **Step 3: Implement API client**

```go
// internal/api/client.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/derek/seats-cli/internal/model"
)

const defaultBaseURL = "https://seats.aero"

// APIError represents an error response from the API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

// ExitCode returns the CLI exit code for this error.
func (e *APIError) ExitCode() int {
	switch {
	case e.StatusCode == 401 || e.StatusCode == 403:
		return 2
	case e.StatusCode == 429:
		return 3
	case e.StatusCode >= 400 && e.StatusCode < 500:
		return 1
	default:
		return 4
	}
}

// Client is the Seats.aero API client.
type Client struct {
	BaseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL:    defaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// SearchParams are the parameters for the cached search endpoint.
type SearchParams struct {
	OriginAirport      string
	DestinationAirport string
	StartDate          string
	EndDate            string
	Cabins             string
	Sources            string
	Carriers           string
	OnlyDirectFlights  bool
	Take               int
	Skip               int
}

// Search calls GET /partnerapi/search.
func (c *Client) Search(p SearchParams) (*model.SearchResponse, error) {
	params := url.Values{}
	params.Set("origin_airport", p.OriginAirport)
	params.Set("destination_airport", p.DestinationAirport)
	if p.StartDate != "" {
		params.Set("start_date", p.StartDate)
	}
	if p.EndDate != "" {
		params.Set("end_date", p.EndDate)
	}
	if p.Cabins != "" {
		params.Set("cabins", p.Cabins)
	}
	if p.Sources != "" {
		params.Set("sources", p.Sources)
	}
	if p.Carriers != "" {
		params.Set("carriers", p.Carriers)
	}
	if p.OnlyDirectFlights {
		params.Set("only_direct_flights", "true")
	}
	if p.Take > 0 {
		params.Set("take", strconv.Itoa(p.Take))
	}
	if p.Skip > 0 {
		params.Set("skip", strconv.Itoa(p.Skip))
	}

	var resp model.SearchResponse
	if err := c.get("/partnerapi/search", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// LiveParams are the parameters for the live search endpoint.
type LiveParams struct {
	OriginAirport      string `json:"origin_airport"`
	DestinationAirport string `json:"destination_airport"`
	DepartureDate      string `json:"departure_date"`
	Source             string `json:"source"`
	SeatCount          int    `json:"seat_count,omitempty"`
}

// Live calls POST /partnerapi/live.
func (c *Client) Live(p LiveParams) (*model.LiveResponse, error) {
	var resp model.LiveResponse
	if err := c.post("/partnerapi/live", p, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AvailabilityParams are the parameters for the bulk availability endpoint.
type AvailabilityParams struct {
	Source            string
	Cabin             string
	StartDate         string
	EndDate           string
	OriginRegion      string
	DestinationRegion string
	Take              int
	Skip              int
}

// Availability calls GET /partnerapi/availability.
func (c *Client) Availability(p AvailabilityParams) (*model.SearchResponse, error) {
	params := url.Values{}
	params.Set("source", p.Source)
	if p.Cabin != "" {
		params.Set("cabin", p.Cabin)
	}
	if p.StartDate != "" {
		params.Set("start_date", p.StartDate)
	}
	if p.EndDate != "" {
		params.Set("end_date", p.EndDate)
	}
	if p.OriginRegion != "" {
		params.Set("origin_region", p.OriginRegion)
	}
	if p.DestinationRegion != "" {
		params.Set("destination_region", p.DestinationRegion)
	}
	if p.Take > 0 {
		params.Set("take", strconv.Itoa(p.Take))
	}
	if p.Skip > 0 {
		params.Set("skip", strconv.Itoa(p.Skip))
	}

	var resp model.SearchResponse
	if err := c.get("/partnerapi/availability", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Routes calls GET /partnerapi/routes.
// Note: the API returns a raw JSON array (not wrapped in {"data": [...]}).
// If this fails at runtime, wrap in a RoutesResponse struct.
func (c *Client) Routes(source string) ([]model.Route, error) {
	params := url.Values{}
	params.Set("source", source)

	var routes []model.Route
	if err := c.get("/partnerapi/routes", params, &routes); err != nil {
		return nil, err
	}
	return routes, nil
}

// Trip calls GET /partnerapi/trips/{id}.
func (c *Client) Trip(id string) (*model.TripResponse, error) {
	var resp model.TripResponse
	if err := c.get("/partnerapi/trips/"+id, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) get(path string, params url.Values, result interface{}) error {
	u := c.BaseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return &APIError{StatusCode: 0, Message: fmt.Sprintf("failed to create request: %v", err)}
	}
	return c.do(req, result)
}

func (c *Client) post(path string, body interface{}, result interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return &APIError{StatusCode: 0, Message: fmt.Sprintf("failed to marshal request: %v", err)}
	}
	req, err := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return &APIError{StatusCode: 0, Message: fmt.Sprintf("failed to create request: %v", err)}
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, result)
}

func (c *Client) do(req *http.Request, result interface{}) error {
	req.Header.Set("Partner-Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIError{StatusCode: 0, Message: "Request failed — could not reach seats.aero."}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{StatusCode: resp.StatusCode, Message: "failed to read response body"}
	}

	switch {
	case resp.StatusCode == 401 || resp.StatusCode == 403:
		return &APIError{StatusCode: resp.StatusCode, Message: "Invalid or missing API key."}
	case resp.StatusCode == 429:
		return &APIError{StatusCode: resp.StatusCode, Message: "Rate limit exceeded (1,000 calls/day). Try again tomorrow."}
	case resp.StatusCode >= 400:
		return &APIError{StatusCode: resp.StatusCode, Message: "API returned 400 — check airport codes."}
	}

	if err := json.Unmarshal(body, result); err != nil {
		return &APIError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("failed to parse response: %v", err)}
	}
	return nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd ~/repo/seats-cli && go test ./internal/api/ -v
```
Expected: all 3 tests PASS (Search, Live, ErrorHandling subtests).

- [ ] **Step 5: Commit**

```bash
cd ~/repo/seats-cli
git add internal/api/
git commit -m "feat: add API client with all endpoint methods and error handling"
```

---

### Task 4: Sort Engine

**Files:**
- Create: `internal/sort/sort.go`
- Create: `internal/sort/sort_test.go`

- [ ] **Step 1: Write tests for sort parsing and execution**

```go
// internal/sort/sort_test.go
package sort

import (
	"testing"

	"github.com/derek/seats-cli/internal/model"
)

func TestParseSortKeys(t *testing.T) {
	keys, err := ParseSortKeys("cabin,miles:desc,date")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0].Field != "cabin" || keys[0].Desc != true {
		t.Errorf("key 0: expected cabin desc, got %s %v", keys[0].Field, keys[0].Desc)
	}
	if keys[1].Field != "miles" || keys[1].Desc != true {
		t.Errorf("key 1: expected miles desc, got %s %v", keys[1].Field, keys[1].Desc)
	}
	if keys[2].Field != "date" || keys[2].Desc != false {
		t.Errorf("key 2: expected date asc, got %s %v", keys[2].Field, keys[2].Desc)
	}
}

func TestParseSortKeysInvalid(t *testing.T) {
	_, err := ParseSortKeys("cabin,invalid_field")
	if err == nil {
		t.Fatal("expected error for invalid field")
	}
}

func TestParseSortKeysEmpty(t *testing.T) {
	keys, err := ParseSortKeys("")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	// Empty string returns default sort (miles asc)
	if len(keys) != 1 || keys[0].Field != "miles" {
		t.Errorf("expected default [miles], got %v", keys)
	}
}

func TestSortFlatRows(t *testing.T) {
	rows := []model.FlatRow{
		{Cabin: "economy", Miles: 12500, Date: "2026-04-03"},
		{Cabin: "business", Miles: 35000, Date: "2026-04-03"},
		{Cabin: "economy", Miles: 10000, Date: "2026-04-04"},
		{Cabin: "business", Miles: 32000, Date: "2026-04-03"},
	}

	keys, _ := ParseSortKeys("cabin,miles")
	SortFlatRows(rows, keys)

	// first > business > economy; within cabin, lowest miles first
	if rows[0].Cabin != "business" || rows[0].Miles != 32000 {
		t.Errorf("row 0: expected business/32000, got %s/%d", rows[0].Cabin, rows[0].Miles)
	}
	if rows[1].Cabin != "business" || rows[1].Miles != 35000 {
		t.Errorf("row 1: expected business/35000, got %s/%d", rows[1].Cabin, rows[1].Miles)
	}
	if rows[2].Cabin != "economy" || rows[2].Miles != 10000 {
		t.Errorf("row 2: expected economy/10000, got %s/%d", rows[2].Cabin, rows[2].Miles)
	}
	if rows[3].Cabin != "economy" || rows[3].Miles != 12500 {
		t.Errorf("row 3: expected economy/12500, got %s/%d", rows[3].Cabin, rows[3].Miles)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/repo/seats-cli && go test ./internal/sort/ -v
```
Expected: FAIL — sort package not defined.

- [ ] **Step 3: Implement sort engine**

```go
// internal/sort/sort.go
package sort

import (
	"fmt"
	gosort "sort"
	"strings"

	"github.com/derek/seats-cli/internal/model"
)

// SortKey represents a single sort key with direction.
type SortKey struct {
	Field string
	Desc  bool
}

// Valid sort fields and their default directions.
var fieldDefaults = map[string]bool{
	"cabin":   true,  // desc by default (first > economy)
	"miles":   false, // asc by default (cheapest first)
	"stops":   false, // asc (fewest first)
	"date":    false, // asc (earliest first)
	"taxes":   false, // asc (lowest first)
	"seats":   true,  // desc (most first)
	"airline": false, // asc (alphabetical)
}

// cabinRank maps cabin names to sort rank (higher = better).
var cabinRank = map[string]int{
	"economy":  0,
	"premium":  1,
	"business": 2,
	"first":    3,
}

// ParseSortKeys parses a comma-separated sort string like "cabin,miles:desc,date".
func ParseSortKeys(s string) ([]SortKey, error) {
	if s == "" {
		return []SortKey{{Field: "miles", Desc: false}}, nil
	}

	parts := strings.Split(s, ",")
	keys := make([]SortKey, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		field := part
		dirOverride := ""
		if idx := strings.Index(part, ":"); idx != -1 {
			field = part[:idx]
			dirOverride = part[idx+1:]
		}

		defaultDesc, valid := fieldDefaults[field]
		if !valid {
			return nil, fmt.Errorf("unknown sort field: %q (valid: cabin, miles, stops, date, taxes, seats, airline)", field)
		}

		desc := defaultDesc
		if dirOverride == "asc" {
			desc = false
		} else if dirOverride == "desc" {
			desc = true
		} else if dirOverride != "" {
			return nil, fmt.Errorf("invalid sort direction: %q (use :asc or :desc)", dirOverride)
		}

		keys = append(keys, SortKey{Field: field, Desc: desc})
	}
	return keys, nil
}

// SortFlatRows sorts flattened search rows by the given keys.
func SortFlatRows(rows []model.FlatRow, keys []SortKey) {
	gosort.SliceStable(rows, func(i, j int) bool {
		for _, key := range keys {
			cmp := compareFlatRows(rows[i], rows[j], key.Field)
			if cmp == 0 {
				continue
			}
			if key.Desc {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})
}

func compareFlatRows(a, b model.FlatRow, field string) int {
	switch field {
	case "cabin":
		return cabinRank[a.Cabin] - cabinRank[b.Cabin]
	case "miles":
		return a.Miles - b.Miles
	case "stops":
		return a.Stops - b.Stops
	case "taxes":
		return a.Taxes - b.Taxes
	case "date":
		return strings.Compare(a.Date, b.Date)
	case "seats":
		return a.Seats - b.Seats
	case "airline":
		return strings.Compare(a.Airline, b.Airline)
	default:
		return 0
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd ~/repo/seats-cli && go test ./internal/sort/ -v
```
Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
cd ~/repo/seats-cli
git add internal/sort/
git commit -m "feat: add multi-key sort engine with direction overrides"
```

---

### Task 5: Markdown Formatter

**Files:**
- Create: `internal/format/markdown.go`
- Create: `internal/format/markdown_test.go`

- [ ] **Step 1: Write test for search markdown output**

```go
// internal/format/markdown_test.go
package format

import (
	"strings"
	"testing"

	"github.com/derek/seats-cli/internal/model"
)

func TestFormatSearchMarkdown(t *testing.T) {
	rows := []model.FlatRow{
		{
			ID: "abc123", Date: "2026-04-03", Program: "united",
			Cabin: "economy", Miles: 12500, Airline: "UA", Direct: true, Seats: 3,
		},
	}
	result := FormatSearchMarkdown("SFO", "BOS", "2026-04-03", rows, false)

	if !strings.Contains(result, "## Award Search: SFO -> BOS") {
		t.Error("missing header")
	}
	if !strings.Contains(result, "abc123") {
		t.Error("missing ID")
	}
	if !strings.Contains(result, "12,500") {
		t.Error("missing formatted miles")
	}
	if !strings.Contains(result, "united") {
		t.Error("missing program")
	}
	if !strings.Contains(result, "| yes |") {
		t.Error("missing direct flag")
	}
	if !strings.Contains(result, "1 results") {
		t.Error("missing result count")
	}
}

func TestFormatSearchMarkdownNoResults(t *testing.T) {
	result := FormatSearchMarkdown("SFO", "BOS", "2026-04-03", nil, false)
	if !strings.Contains(result, "No results found") {
		t.Error("expected no results message")
	}
}

func TestFormatProgramsMarkdown(t *testing.T) {
	result := FormatProgramsMarkdown()
	if !strings.Contains(result, "## Available Programs") {
		t.Error("missing header")
	}
	if !strings.Contains(result, "united") {
		t.Error("missing united program")
	}
	if !strings.Contains(result, "United MileagePlus") {
		t.Error("missing program name")
	}
}

func TestFormatLiveMarkdown(t *testing.T) {
	results := []model.LiveResult{
		{
			FlightNumbers: "UA123",
			Cabin:         "economy",
			MileageCost:   12500,
			TotalTaxes:    560,
			TaxesCurrencySymbol: "$",
			Stops:         0,
			TotalDuration: 330,
			DepartsAt:     "2026-04-03T08:00:00Z",
			ArrivesAt:     "2026-04-03T16:30:00Z",
			RemainingSeats: 3,
		},
	}
	result := FormatLiveMarkdown("SFO", "BOS", "2026-04-03", "united", results)
	if !strings.Contains(result, "## Live Search: SFO -> BOS") {
		t.Error("missing header")
	}
	if !strings.Contains(result, "UA123") {
		t.Error("missing flight number")
	}
	if !strings.Contains(result, "12,500") {
		t.Error("missing miles")
	}
	if !strings.Contains(result, "5h 30m") {
		t.Error("missing duration")
	}
}

func TestFormatTripMarkdown(t *testing.T) {
	resp := &model.TripResponse{
		Data: []model.Trip{
			{
				Cabin:       "economy",
				MileageCost: 12500,
				TotalTaxes:  560,
				TaxesCurrency: "USD",
				TaxesCurrencySymbol: "$",
				TotalDuration: 330,
				Stops:       0,
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
			{Label: "United MileagePlus", Link: "https://united.com", Primary: true},
		},
	}
	result := FormatTripMarkdown("abc123", resp)
	if !strings.Contains(result, "## Trip Details: abc123") {
		t.Error("missing header")
	}
	if !strings.Contains(result, "economy") {
		t.Error("missing cabin")
	}
	if !strings.Contains(result, "12,500") {
		t.Error("missing miles")
	}
	if !strings.Contains(result, "Booking Links") {
		t.Error("missing booking links section")
	}
	if !strings.Contains(result, "United MileagePlus") {
		t.Error("missing booking link label")
	}
}

func TestFormatRoutesMarkdown(t *testing.T) {
	routes := []model.Route{
		{
			OriginAirport: "SFO", DestinationAirport: "NRT",
			OriginRegion: "North America", DestinationRegion: "Asia",
			Distance: 5130, NumDaysOut: 60,
		},
	}
	result := FormatRoutesMarkdown("aeroplan", routes)
	if !strings.Contains(result, "## Routes: aeroplan") {
		t.Error("missing header")
	}
	if !strings.Contains(result, "SFO") {
		t.Error("missing origin")
	}
	if !strings.Contains(result, "5,130") {
		t.Error("missing distance")
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
	result := FormatAvailabilityMarkdown("american", "", "", "", data, false)
	if !strings.Contains(result, "## Bulk Availability: american") {
		t.Error("missing header")
	}
	if !strings.Contains(result, "DFW -> NRT") {
		t.Error("missing route")
	}
	if !strings.Contains(result, "35,000 (3)") {
		t.Error("missing economy cost with seats")
	}
	if !strings.Contains(result, "70,000 (2)") {
		t.Error("missing business cost with seats")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/repo/seats-cli && go test ./internal/format/ -v
```
Expected: FAIL — format functions not defined.

- [ ] **Step 3: Implement markdown formatter**

```go
// internal/format/markdown.go
package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/derek/seats-cli/internal/model"
)

// FormatSearchMarkdown renders flattened search rows as a markdown table.
func FormatSearchMarkdown(from, to, date string, rows []model.FlatRow, hasMore bool) string {
	if len(rows) == 0 {
		return fmt.Sprintf("## Award Search: %s -> %s (%s)\n\nNo results found.\n", from, to, date)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Award Search: %s -> %s (%s)\n\n", from, to, date)
	b.WriteString("| ID | Date | Program | Cabin | Miles | Airline | Direct | Seats |\n")
	b.WriteString("|----|------|---------|-------|-------|---------|--------|-------|\n")

	for _, r := range rows {
		direct := "no"
		if r.Direct {
			direct = "yes"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | %d |\n",
			truncateID(r.ID), r.Date, r.Program, r.Cabin,
			formatNumber(r.Miles), r.Airline, direct, r.Seats)
	}

	b.WriteString(fmt.Sprintf("\n%d results", len(rows)))
	if hasMore {
		b.WriteString(". More available — use --skip to paginate")
	}
	b.WriteString(".\n")
	return b.String()
}

// FormatLiveMarkdown renders live search results as a markdown table.
func FormatLiveMarkdown(from, to, date, program string, results []model.LiveResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("## Live Search: %s -> %s (%s, %s)\n\nNo results found.\n", from, to, date, program)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Live Search: %s -> %s (%s, %s)\n\n", from, to, date, program)
	b.WriteString("| Flight | Cabin | Miles | Taxes | Stops | Duration | Departs | Arrives | Seats |\n")
	b.WriteString("|--------|-------|-------|-------|-------|----------|---------|---------|-------|\n")

	for _, r := range results {
		taxes := formatTaxes(r.TotalTaxes, r.TaxesCurrencySymbol)
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %d | %s | %s | %s | %d |\n",
			r.FlightNumbers, r.Cabin, formatNumber(r.MileageCost),
			taxes, r.Stops, formatDuration(r.TotalDuration),
			formatTime(r.DepartsAt), formatTime(r.ArrivesAt), r.RemainingSeats)
	}

	fmt.Fprintf(&b, "\n%d results found.\n", len(results))
	return b.String()
}

// FormatAvailabilityMarkdown renders bulk availability as a markdown table with per-cabin columns.
func FormatAvailabilityMarkdown(program, startDate, endDate, cabin string, data []model.Availability, hasMore bool) string {
	if len(data) == 0 {
		return fmt.Sprintf("## Bulk Availability: %s\n\nNo results found.\n", program)
	}

	var b strings.Builder
	header := program
	if startDate != "" || endDate != "" {
		header += " (" + startDate
		if endDate != "" {
			header += " -> " + endDate
		}
		if cabin != "" {
			header += ", " + cabin
		}
		header += ")"
	}
	fmt.Fprintf(&b, "## Bulk Availability: %s\n\n", header)
	b.WriteString("| ID | Date | Route | Economy | Business | First | Direct |\n")
	b.WriteString("|----|------|-------|---------|----------|-------|--------|\n")

	for _, a := range data {
		route := fmt.Sprintf("%s -> %s", a.Route.OriginAirport, a.Route.DestinationAirport)
		econ := formatCabinCell(a.YAvailable, a.YMileageCost, a.YRemainingSeats)
		biz := formatCabinCell(a.JAvailable, a.JMileageCost, a.JRemainingSeats)
		first := formatCabinCell(a.FAvailable, a.FMileageCost, a.FRemainingSeats)
		direct := "no"
		if a.YDirect || a.JDirect || a.FDirect {
			direct = "yes"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n",
			truncateID(a.ID), a.Date, route, econ, biz, first, direct)
	}

	fmt.Fprintf(&b, "\n%d results returned", len(data))
	if hasMore {
		b.WriteString(". More available — use --skip to paginate")
	}
	b.WriteString(".\n")
	return b.String()
}

// FormatTripMarkdown renders trip details as markdown.
func FormatTripMarkdown(id string, resp *model.TripResponse) string {
	if len(resp.Data) == 0 {
		return fmt.Sprintf("## Trip Details: %s\n\nNo trip data found.\n", id)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Trip Details: %s\n", id)

	for i, trip := range resp.Data {
		fmt.Fprintf(&b, "\n### Option %d: %s — %s miles\n\n",
			i+1, trip.Cabin, formatNumber(trip.MileageCost))
		b.WriteString("| Segment | Flight | Route | Departs | Arrives | Aircraft |\n")
		b.WriteString("|---------|--------|-------|---------|---------|----------|\n")

		for _, seg := range trip.AvailabilitySegments {
			route := fmt.Sprintf("%s -> %s", seg.OriginAirport, seg.DestinationAirport)
			fmt.Fprintf(&b, "| %d | %s | %s | %s | %s | %s |\n",
				seg.Order, seg.FlightNumber, route,
				formatTime(seg.DepartsAt), formatTime(seg.ArrivesAt), seg.AircraftName)
		}

		taxes := formatTaxes(trip.TotalTaxes, trip.TaxesCurrencySymbol)
		fmt.Fprintf(&b, "\n- **Duration**: %s | **Stops**: %d | **Taxes**: %s %s\n",
			formatDuration(trip.TotalDuration), trip.Stops, taxes, trip.TaxesCurrency)
	}

	if len(resp.BookingLinks) > 0 {
		b.WriteString("\n### Booking Links\n")
		for _, link := range resp.BookingLinks {
			suffix := ""
			if link.Primary {
				suffix = " (primary)"
			}
			fmt.Fprintf(&b, "- [%s](%s)%s\n", link.Label, link.Link, suffix)
		}
	}

	return b.String()
}

// FormatRoutesMarkdown renders route list as markdown.
func FormatRoutesMarkdown(program string, routes []model.Route) string {
	if len(routes) == 0 {
		return fmt.Sprintf("## Routes: %s\n\nNo routes found.\n", program)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Routes: %s\n\n", program)
	b.WriteString("| Origin | Destination | Origin Region | Dest Region | Distance | Days Out |\n")
	b.WriteString("|--------|-------------|---------------|-------------|----------|----------|\n")

	for _, r := range routes {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %d |\n",
			r.OriginAirport, r.DestinationAirport,
			r.OriginRegion, r.DestinationRegion,
			formatNumber(r.Distance), r.NumDaysOut)
	}

	fmt.Fprintf(&b, "\n%d routes found.\n", len(routes))
	return b.String()
}

// FormatProgramsMarkdown renders the programs list as markdown.
func FormatProgramsMarkdown() string {
	var b strings.Builder
	b.WriteString("## Available Programs\n\n")
	b.WriteString("| Program | Name |\n")
	b.WriteString("|---------|------|\n")
	for _, p := range model.Programs {
		fmt.Fprintf(&b, "| %s | %s |\n", p.ID, p.Name)
	}
	return b.String()
}

// --- helpers ---

func truncateID(id string) string {
	if len(id) > 10 {
		return id[:10]
	}
	return id
}

func formatNumber(n int) string {
	if n == 0 {
		return "0"
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		result.WriteString(s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		if result.Len() > 0 {
			result.WriteByte(',')
		}
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

func formatDuration(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}

func formatTime(isoTime string) string {
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		return isoTime
	}
	return t.Format("15:04")
}

func formatTaxes(cents int, symbol string) string {
	if symbol == "" {
		symbol = "$"
	}
	dollars := float64(cents) / 100.0
	return fmt.Sprintf("%s%.2f", symbol, dollars)
}

func formatCabinCell(available bool, milesCost string, seats int) string {
	if !available {
		return "—"
	}
	miles := model.ParseMiles(milesCost)
	return fmt.Sprintf("%s (%d)", formatNumber(miles), seats)
}
```

- [ ] **Step 4: Run tests**

```bash
cd ~/repo/seats-cli && go test ./internal/format/ -v
```
Expected: all 7 tests PASS.

- [ ] **Step 5: Commit**

```bash
cd ~/repo/seats-cli
git add internal/format/markdown.go internal/format/markdown_test.go
git commit -m "feat: add markdown formatter for all output types"
```

---

### Task 6: Programs Command

**Files:**
- Create: `cmd/programs.go`

- [ ] **Step 1: Implement programs command**

```go
// cmd/programs.go
package cmd

import (
	"fmt"

	"github.com/derek/seats-cli/internal/format"
	"github.com/spf13/cobra"
)

var programsCmd = &cobra.Command{
	Use:   "programs",
	Short: "List available mileage programs",
	Long:  "Lists all mileage programs supported by Seats.aero and their identifiers for use with --program.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: pretty mode will be added in Task 11
		fmt.Print(format.FormatProgramsMarkdown())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(programsCmd)
}
```

- [ ] **Step 2: Verify it compiles and runs**

```bash
cd ~/repo/seats-cli && go build -o seats . && ./seats programs
```
Expected: prints markdown table of programs.

- [ ] **Step 3: Commit**

```bash
cd ~/repo/seats-cli
git add cmd/programs.go
git commit -m "feat: add programs subcommand"
```

---

### Task 7: Search Command

**Files:**
- Create: `cmd/search.go`

- [ ] **Step 1: Implement search command**

```go
// cmd/search.go
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
	Long:  "Search cached award availability for specific airports and dates across all mileage programs.",
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().String("from", "", "Origin airport(s), comma-separated (required)")
	searchCmd.Flags().String("to", "", "Destination airport(s), comma-separated (required)")
	searchCmd.Flags().String("date", "", "Start date (YYYY-MM-DD)")
	searchCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	searchCmd.Flags().String("cabin", "", "Cabin class: economy, premium, business, first (comma-separated)")
	searchCmd.Flags().String("program", "", "Mileage program filter (comma-separated)")
	searchCmd.Flags().String("carrier", "", "Airline filter (comma-separated)")
	searchCmd.Flags().Bool("direct", false, "Nonstop flights only")
	searchCmd.Flags().String("sort", "", "Sort keys: cabin, miles, stops, date, taxes, seats, airline (comma-separated)")
	searchCmd.Flags().Int("limit", 50, "Max results (10-1000)")
	searchCmd.Flags().Int("skip", 0, "Skip N results for pagination")

	searchCmd.MarkFlagRequired("from")
	searchCmd.MarkFlagRequired("to")

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

	// Validate limit range
	if limit < 10 || limit > 1000 {
		fmt.Fprintln(os.Stderr, "Error: --limit must be between 10 and 1000.")
		os.Exit(1)
	}

	// Parse sort keys
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

	// Flatten availability records into per-cabin rows
	var rows []model.FlatRow
	for _, a := range resp.Data {
		rows = append(rows, a.Flatten()...)
	}

	// Sort
	seatsSort.SortFlatRows(rows, sortKeys)

	fmt.Print(format.FormatSearchMarkdown(from, to, date, rows, resp.HasMore))
	return nil
}
```

- [ ] **Step 2: Add handleAPIError helper to root.go**

Add to `cmd/root.go`:

```go
func handleAPIError(err error) error {
	apiErr, ok := err.(*api.APIError)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(4)
	}
	fmt.Fprintf(os.Stderr, "Error: %s\n", apiErr.Message)
	os.Exit(apiErr.ExitCode())
	return nil // unreachable
}
```

Add import for `"github.com/derek/seats-cli/internal/api"` to root.go.

- [ ] **Step 3: Verify it compiles**

```bash
cd ~/repo/seats-cli && go build -o seats .
```
Expected: clean build.

- [ ] **Step 4: Test help output**

```bash
cd ~/repo/seats-cli && ./seats search --help
```
Expected: shows all flags with descriptions.

- [ ] **Step 5: Commit**

```bash
cd ~/repo/seats-cli
git add cmd/search.go cmd/root.go
git commit -m "feat: add search subcommand with sorting and pagination"
```

---

### Task 8: Live Command

**Files:**
- Create: `cmd/live.go`

- [ ] **Step 1: Implement live command**

```go
// cmd/live.go
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
	Long:  "Perform a real-time search for a specific route, date, and mileage program.",
	RunE:  runLive,
}

func init() {
	liveCmd.Flags().String("from", "", "Origin airport (required)")
	liveCmd.Flags().String("to", "", "Destination airport (required)")
	liveCmd.Flags().String("date", "", "Departure date YYYY-MM-DD (required)")
	liveCmd.Flags().String("program", "", "Mileage program (required, single value)")
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
		fmt.Fprintln(os.Stderr, "Error: --seats must be between 1 and 9.")
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

	fmt.Print(format.FormatLiveMarkdown(from, to, date, program, resp.Results))
	return nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd ~/repo/seats-cli && go build -o seats .
```

- [ ] **Step 3: Commit**

```bash
cd ~/repo/seats-cli
git add cmd/live.go
git commit -m "feat: add live subcommand for real-time search"
```

---

### Task 9: Availability Command

**Files:**
- Create: `cmd/availability.go`

- [ ] **Step 1: Implement availability command**

```go
// cmd/availability.go
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
	availabilityCmd.Flags().String("program", "", "Mileage program (required)")
	availabilityCmd.Flags().String("cabin", "", "Cabin class: economy, premium, business, first")
	availabilityCmd.Flags().String("date", "", "Start date (YYYY-MM-DD)")
	availabilityCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	availabilityCmd.Flags().String("origin-region", "", "Origin region: North America, South America, Africa, Asia, Europe, Oceania")
	availabilityCmd.Flags().String("dest-region", "", "Destination region")
	availabilityCmd.Flags().String("sort", "", "Sort keys (comma-separated)")
	availabilityCmd.Flags().Int("limit", 50, "Max results (10-1000)")
	availabilityCmd.Flags().Int("skip", 0, "Skip N results for pagination")

	availabilityCmd.MarkFlagRequired("program")

	rootCmd.AddCommand(availabilityCmd)
}

func runAvailability(cmd *cobra.Command, args []string) error {
	program, _ := cmd.Flags().GetString("program")
	cabin, _ := cmd.Flags().GetString("cabin")
	date, _ := cmd.Flags().GetString("date")
	endDate, _ := cmd.Flags().GetString("end-date")
	originRegion, _ := cmd.Flags().GetString("origin-region")
	destRegion, _ := cmd.Flags().GetString("dest-region")
	limit, _ := cmd.Flags().GetInt("limit")
	skip, _ := cmd.Flags().GetInt("skip")

	sortStr, _ := cmd.Flags().GetString("sort")

	if limit < 10 || limit > 1000 {
		fmt.Fprintln(os.Stderr, "Error: --limit must be between 10 and 1000.")
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

	// Flatten and sort when --sort is provided
	if sortStr != "" {
		var rows []model.FlatRow
		for _, a := range resp.Data {
			rows = append(rows, a.Flatten()...)
		}
		seatsSort.SortFlatRows(rows, sortKeys)
		// When sorted, use search-style flattened output
		fmt.Print(format.FormatSearchMarkdown(program, "(bulk)", date, rows, resp.HasMore))
		return nil
	}

	fmt.Print(format.FormatAvailabilityMarkdown(program, date, endDate, cabin, resp.Data, resp.HasMore))
	return nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd ~/repo/seats-cli && go build -o seats .
```

- [ ] **Step 3: Commit**

```bash
cd ~/repo/seats-cli
git add cmd/availability.go
git commit -m "feat: add availability subcommand for bulk program search"
```

---

### Task 10: Routes and Trip Commands

**Files:**
- Create: `cmd/routes.go`
- Create: `cmd/trip.go`

- [ ] **Step 1: Implement routes command**

```go
// cmd/routes.go
package cmd

import (
	"fmt"
	"os"

	"github.com/derek/seats-cli/internal/api"
	"github.com/derek/seats-cli/internal/format"
	"github.com/spf13/cobra"
)

var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "List routes for a program",
	Long:  "List all routes covered by a mileage program.",
	RunE:  runRoutes,
}

func init() {
	routesCmd.Flags().String("program", "", "Mileage program (required)")
	routesCmd.MarkFlagRequired("program")
	rootCmd.AddCommand(routesCmd)
}

func runRoutes(cmd *cobra.Command, args []string) error {
	program, _ := cmd.Flags().GetString("program")

	client := api.NewClient(os.Getenv("SEATS_AERO_API_KEY"))
	routes, err := client.Routes(program)
	if err != nil {
		return handleAPIError(err)
	}

	fmt.Print(format.FormatRoutesMarkdown(program, routes))
	return nil
}
```

- [ ] **Step 2: Implement trip command**

```go
// cmd/trip.go
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
	Long:  "Get detailed segments, taxes, and booking links for a specific availability result.",
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
```

- [ ] **Step 3: Verify it compiles**

```bash
cd ~/repo/seats-cli && go build -o seats .
```

- [ ] **Step 4: Test all help outputs**

```bash
cd ~/repo/seats-cli && ./seats --help && ./seats search --help && ./seats live --help && ./seats availability --help && ./seats routes --help && ./seats trip --help && ./seats programs --help
```
Expected: all commands show proper help text.

- [ ] **Step 5: Commit**

```bash
cd ~/repo/seats-cli
git add cmd/routes.go cmd/trip.go
git commit -m "feat: add routes and trip subcommands"
```

---

### Task 11: Pretty Formatter

**Files:**
- Create: `internal/format/pretty.go`
- Modify: `cmd/programs.go` — add pretty mode
- Modify: `cmd/search.go` — add pretty mode
- Modify: `cmd/live.go` — add pretty mode
- Modify: `cmd/availability.go` — add pretty mode
- Modify: `cmd/routes.go` — add pretty mode
- Modify: `cmd/trip.go` — add pretty mode

- [ ] **Step 1: Implement pretty formatter using Lip Gloss table**

```go
// internal/format/pretty.go
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

func newTable(headers []string, rows [][]string) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(headers...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	for _, row := range rows {
		t.Row(row...)
	}
	return t.Render()
}

// PrettySearchMarkdown renders search results as a styled terminal table.
func PrettySearch(from, to, date string, rows []model.FlatRow, hasMore bool) string {
	title := titleStyle.Render(fmt.Sprintf("Award Search: %s → %s (%s)", from, to, date))

	if len(rows) == 0 {
		return title + "\n\nNo results found.\n"
	}

	headers := []string{"ID", "Date", "Program", "Cabin", "Miles", "Airline", "Direct", "Seats"}
	var tableRows [][]string
	for _, r := range rows {
		direct := "✗"
		if r.Direct {
			direct = "✓"
		}
		tableRows = append(tableRows, []string{
			truncateID(r.ID), r.Date, r.Program, r.Cabin,
			formatNumber(r.Miles), r.Airline, direct, fmt.Sprintf("%d", r.Seats),
		})
	}

	result := title + "\n" + newTable(headers, tableRows)
	footer := fmt.Sprintf("\n%d results", len(rows))
	if hasMore {
		footer += ". More available — use --skip to paginate"
	}
	return result + footer + ".\n"
}

// PrettyLive renders live search results as a styled terminal table.
func PrettyLive(from, to, date, program string, results []model.LiveResult) string {
	title := titleStyle.Render(fmt.Sprintf("Live Search: %s → %s (%s, %s)", from, to, date, program))

	if len(results) == 0 {
		return title + "\n\nNo results found.\n"
	}

	headers := []string{"Flight", "Cabin", "Miles", "Taxes", "Stops", "Duration", "Departs", "Arrives", "Seats"}
	var tableRows [][]string
	for _, r := range results {
		tableRows = append(tableRows, []string{
			r.FlightNumbers, r.Cabin, formatNumber(r.MileageCost),
			formatTaxes(r.TotalTaxes, r.TaxesCurrencySymbol),
			fmt.Sprintf("%d", r.Stops), formatDuration(r.TotalDuration),
			formatTime(r.DepartsAt), formatTime(r.ArrivesAt),
			fmt.Sprintf("%d", r.RemainingSeats),
		})
	}

	return title + "\n" + newTable(headers, tableRows) + fmt.Sprintf("\n%d results found.\n", len(results))
}

// PrettyAvailability renders bulk availability as a styled terminal table.
func PrettyAvailability(program, startDate, endDate, cabin string, data []model.Availability, hasMore bool) string {
	header := program
	if startDate != "" {
		header += " (" + startDate
		if endDate != "" {
			header += " → " + endDate
		}
		if cabin != "" {
			header += ", " + cabin
		}
		header += ")"
	}
	title := titleStyle.Render("Bulk Availability: " + header)

	if len(data) == 0 {
		return title + "\n\nNo results found.\n"
	}

	headers := []string{"ID", "Date", "Route", "Economy", "Business", "First", "Direct"}
	var tableRows [][]string
	for _, a := range data {
		route := fmt.Sprintf("%s → %s", a.Route.OriginAirport, a.Route.DestinationAirport)
		direct := "✗"
		if a.YDirect || a.JDirect || a.FDirect {
			direct = "✓"
		}
		tableRows = append(tableRows, []string{
			truncateID(a.ID), a.Date, route,
			formatCabinCell(a.YAvailable, a.YMileageCost, a.YRemainingSeats),
			formatCabinCell(a.JAvailable, a.JMileageCost, a.JRemainingSeats),
			formatCabinCell(a.FAvailable, a.FMileageCost, a.FRemainingSeats),
			direct,
		})
	}

	result := title + "\n" + newTable(headers, tableRows)
	footer := fmt.Sprintf("\n%d results returned", len(data))
	if hasMore {
		footer += ". More available — use --skip to paginate"
	}
	return result + footer + ".\n"
}

// PrettyTrip renders trip details as styled terminal output.
func PrettyTrip(id string, resp *model.TripResponse) string {
	title := titleStyle.Render("Trip Details: " + id)

	if len(resp.Data) == 0 {
		return title + "\n\nNo trip data found.\n"
	}

	var b strings.Builder
	b.WriteString(title + "\n")

	for i, trip := range resp.Data {
		optTitle := lipgloss.NewStyle().Bold(true).Render(
			fmt.Sprintf("Option %d: %s — %s miles", i+1, trip.Cabin, formatNumber(trip.MileageCost)))
		b.WriteString("\n" + optTitle + "\n")

		headers := []string{"Seg", "Flight", "Route", "Departs", "Arrives", "Aircraft"}
		var tableRows [][]string
		for _, seg := range trip.AvailabilitySegments {
			tableRows = append(tableRows, []string{
				fmt.Sprintf("%d", seg.Order), seg.FlightNumber,
				fmt.Sprintf("%s → %s", seg.OriginAirport, seg.DestinationAirport),
				formatTime(seg.DepartsAt), formatTime(seg.ArrivesAt), seg.AircraftName,
			})
		}
		b.WriteString(newTable(headers, tableRows) + "\n")

		taxes := formatTaxes(trip.TotalTaxes, trip.TaxesCurrencySymbol)
		b.WriteString(fmt.Sprintf("Duration: %s | Stops: %d | Taxes: %s %s\n",
			formatDuration(trip.TotalDuration), trip.Stops, taxes, trip.TaxesCurrency))
	}

	if len(resp.BookingLinks) > 0 {
		linkTitle := lipgloss.NewStyle().Bold(true).Render("Booking Links")
		b.WriteString("\n" + linkTitle + "\n")
		for _, link := range resp.BookingLinks {
			suffix := ""
			if link.Primary {
				suffix = " (primary)"
			}
			b.WriteString(fmt.Sprintf("  %s: %s%s\n", link.Label, link.Link, suffix))
		}
	}

	return b.String()
}

// PrettyRoutes renders routes as a styled terminal table.
func PrettyRoutes(program string, routes []model.Route) string {
	title := titleStyle.Render("Routes: " + program)

	if len(routes) == 0 {
		return title + "\n\nNo routes found.\n"
	}

	headers := []string{"Origin", "Destination", "Origin Region", "Dest Region", "Distance", "Days Out"}
	var tableRows [][]string
	for _, r := range routes {
		tableRows = append(tableRows, []string{
			r.OriginAirport, r.DestinationAirport,
			r.OriginRegion, r.DestinationRegion,
			formatNumber(r.Distance), fmt.Sprintf("%d", r.NumDaysOut),
		})
	}

	return title + "\n" + newTable(headers, tableRows) + fmt.Sprintf("\n%d routes found.\n", len(routes))
}

// PrettyPrograms renders programs as a styled terminal table.
func PrettyPrograms() string {
	title := titleStyle.Render("Available Programs")

	headers := []string{"Program", "Name"}
	var tableRows [][]string
	for _, p := range model.Programs {
		tableRows = append(tableRows, []string{p.ID, p.Name})
	}

	return title + "\n" + newTable(headers, tableRows) + "\n"
}
```

- [ ] **Step 2: Update all commands to support --pretty flag**

In each command's `RunE` function, add the pretty mode check. Example pattern for `cmd/search.go`:

```go
// Replace the final fmt.Print line with:
if pretty {
    fmt.Print(format.PrettySearch(from, to, date, rows, resp.HasMore))
} else {
    fmt.Print(format.FormatSearchMarkdown(from, to, date, rows, resp.HasMore))
}
```

Apply the same pattern to all commands:
- `cmd/programs.go`: `PrettyPrograms()` vs `FormatProgramsMarkdown()`
- `cmd/live.go`: `PrettyLive(...)` vs `FormatLiveMarkdown(...)`
- `cmd/availability.go`: `PrettyAvailability(...)` vs `FormatAvailabilityMarkdown(...)`
- `cmd/routes.go`: `PrettyRoutes(...)` vs `FormatRoutesMarkdown(...)`
- `cmd/trip.go`: `PrettyTrip(...)` vs `FormatTripMarkdown(...)`

- [ ] **Step 3: Verify it compiles**

```bash
cd ~/repo/seats-cli && go build -o seats .
```

- [ ] **Step 4: Test pretty output**

```bash
cd ~/repo/seats-cli && ./seats programs --pretty
```
Expected: styled table with borders and colored headers.

- [ ] **Step 5: Commit**

```bash
cd ~/repo/seats-cli
git add internal/format/pretty.go cmd/
git commit -m "feat: add pretty terminal output mode with Lip Gloss styling"
```

---

### Task 12: Final Wiring and Smoke Test

**Files:**
- Modify: `cmd/root.go` — final cleanup
- No new files

- [ ] **Step 1: Run all tests**

```bash
cd ~/repo/seats-cli && go test ./... -v
```
Expected: all tests PASS across model, api, format, sort packages.

- [ ] **Step 2: Run go vet and build**

```bash
cd ~/repo/seats-cli && go vet ./... && go build -o seats .
```
Expected: no warnings, clean build.

- [ ] **Step 3: Smoke test all commands (help only, no API key needed)**

```bash
cd ~/repo/seats-cli
./seats --help
./seats search --help
./seats live --help
./seats availability --help
./seats routes --help
./seats trip --help
./seats programs
./seats programs --pretty
```
Expected: all help text displays correctly, programs outputs both markdown and pretty.

- [ ] **Step 4: Test error for missing API key**

```bash
cd ~/repo/seats-cli && unset SEATS_AERO_API_KEY && ./seats search --from SFO --to BOS; echo "Exit code: $?"
```
Expected: `Error: SEATS_AERO_API_KEY environment variable not set.` and exit code 2.

- [ ] **Step 5: Install binary**

```bash
cd ~/repo/seats-cli && go install .
```
Expected: `seats` binary available in `$GOPATH/bin`.

- [ ] **Step 6: Commit**

```bash
cd ~/repo/seats-cli
git add -A
git commit -m "chore: final wiring and cleanup"
```

---

## Summary

| Task | Description | Dependencies |
|------|-------------|-------------|
| 1 | Project scaffolding | none |
| 2 | API response models | Task 1 |
| 3 | API client | Task 2 |
| 4 | Sort engine | Task 2 |
| 5 | Markdown formatter | Task 2 |
| 6 | Programs command | Task 5 |
| 7 | Search command | Tasks 3, 4, 5 |
| 8 | Live command | Tasks 3, 5 |
| 9 | Availability command | Tasks 3, 5 |
| 10 | Routes + Trip commands | Tasks 3, 5 |
| 11 | Pretty formatter | Tasks 6-10 |
| 12 | Final wiring + smoke test | Task 11 |

**Parallelizable:** Tasks 3, 4, 5 can run in parallel after Task 2. Tasks 6-10 can run in parallel after Tasks 3+5.
