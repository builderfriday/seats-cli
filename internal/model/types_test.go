package model

import (
	"encoding/json"
	"testing"
)

func TestSearchResponseUnmarshal(t *testing.T) {
	raw := `{
		"data": [
			{
				"ID": "avail-1",
				"RouteID": "route-1",
				"Route": {
					"ID": "route-1",
					"OriginAirport": "SFO",
					"OriginRegion": "NA",
					"DestinationAirport": "NRT",
					"DestinationRegion": "JP",
					"NumDaysOut": 30,
					"Distance": 5150,
					"Source": "united"
				},
				"Date": "2025-06-01",
				"YAvailable": true,
				"JAvailable": true,
				"FAvailable": false,
				"WAvailable": false,
				"YMileageCost": "12500",
				"JMileageCost": "35000",
				"YRemainingSeats": 4,
				"JRemainingSeats": 2,
				"Source": "united"
			}
		],
		"count": 1,
		"hasMore": false,
		"cursor": 0
	}`

	var resp SearchResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 availability record, got %d", len(resp.Data))
	}

	a := resp.Data[0]
	if a.ID != "avail-1" {
		t.Errorf("expected ID 'avail-1', got %q", a.ID)
	}
	if a.Route.OriginAirport != "SFO" {
		t.Errorf("expected OriginAirport 'SFO', got %q", a.Route.OriginAirport)
	}
	if !a.YAvailable {
		t.Error("expected YAvailable true")
	}
	if a.YMileageCost != "12500" {
		t.Errorf("expected YMileageCost '12500', got %q", a.YMileageCost)
	}
	if a.YRemainingSeats != 4 {
		t.Errorf("expected YRemainingSeats 4, got %d", a.YRemainingSeats)
	}
}

func TestTripResponseUnmarshal(t *testing.T) {
	raw := `{
		"data": [
			{
				"ID": "trip-1",
				"RouteID": "route-1",
				"AvailabilityID": "avail-1",
				"AvailabilitySegments": [
					{
						"ID": "seg-1",
						"FlightNumber": "UA123",
						"Distance": 5150,
						"FareClass": "X",
						"AircraftName": "Boeing 787",
						"AircraftCode": "787",
						"OriginAirport": "SFO",
						"DestinationAirport": "NRT",
						"DepartsAt": "2025-06-01T10:00:00Z",
						"ArrivesAt": "2025-06-02T14:30:00Z",
						"Source": "united",
						"Cabin": "economy",
						"Order": 0
					}
				],
				"TotalDuration": 330,
				"Stops": 0,
				"MileageCost": 12500,
				"TotalTaxes": 56,
				"Cabin": "economy",
				"Source": "united"
			}
		],
		"bookingLinks": [
			{
				"label": "United",
				"link": "https://united.com",
				"primary": true
			}
		]
	}`

	var resp TripResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 trip, got %d", len(resp.Data))
	}

	trip := resp.Data[0]
	if trip.TotalDuration != 330 {
		t.Errorf("expected TotalDuration 330, got %d", trip.TotalDuration)
	}
	if trip.MileageCost != 12500 {
		t.Errorf("expected MileageCost 12500, got %d", trip.MileageCost)
	}

	if len(trip.AvailabilitySegments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(trip.AvailabilitySegments))
	}
	seg := trip.AvailabilitySegments[0]
	if seg.FlightNumber != "UA123" {
		t.Errorf("expected FlightNumber 'UA123', got %q", seg.FlightNumber)
	}

	if len(resp.BookingLinks) != 1 {
		t.Fatalf("expected 1 booking link, got %d", len(resp.BookingLinks))
	}
	if !resp.BookingLinks[0].Primary {
		t.Error("expected BookingLinks[0].Primary true")
	}
}

func TestLiveResponseUnmarshal(t *testing.T) {
	raw := `{
		"results": [
			{
				"ID": "trip-1",
				"RouteID": "route-1",
				"AvailabilityID": "avail-1",
				"AvailabilitySegments": [],
				"TotalDuration": 330,
				"Stops": 0,
				"MileageCost": 12500,
				"TotalTaxes": 56,
				"Cabin": "economy",
				"Source": "united",
				"Filtered": false
			}
		]
	}`

	var resp LiveResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].MileageCost != 12500 {
		t.Errorf("expected MileageCost 12500, got %d", resp.Results[0].MileageCost)
	}
}

func TestAvailabilityFlatten(t *testing.T) {
	a := Availability{
		ID:              "avail-1",
		Date:            "2025-06-01",
		Source:          "united",
		YAvailable:      true,
		JAvailable:      true,
		FAvailable:      false,
		WAvailable:      false,
		YMileageCost:    "12500",
		JMileageCost:    "35000",
		YDirect:         true,
		JDirect:         true,
		YRemainingSeats: 4,
		JRemainingSeats: 2,
	}

	rows := a.Flatten()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// Row 0: economy
	if rows[0].Cabin != "economy" {
		t.Errorf("expected cabin 'economy', got %q", rows[0].Cabin)
	}
	if rows[0].Miles != 12500 {
		t.Errorf("expected Miles 12500, got %d", rows[0].Miles)
	}
	if rows[0].Stops != 0 {
		t.Errorf("expected Stops 0 (direct), got %d", rows[0].Stops)
	}

	// Row 1: business
	if rows[1].Cabin != "business" {
		t.Errorf("expected cabin 'business', got %q", rows[1].Cabin)
	}
	if rows[1].Miles != 35000 {
		t.Errorf("expected Miles 35000, got %d", rows[1].Miles)
	}
}

func TestAvailabilityFlattenNoneAvailable(t *testing.T) {
	a := Availability{}
	rows := a.Flatten()
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}
