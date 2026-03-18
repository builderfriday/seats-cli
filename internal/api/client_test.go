package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header.
		if got := r.Header.Get("Partner-Authorization"); got != "test-key" {
			t.Errorf("Partner-Authorization = %q, want %q", got, "test-key")
		}

		// Verify path.
		if r.URL.Path != "/partnerapi/search" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/partnerapi/search")
		}

		// Verify query parameters.
		q := r.URL.Query()
		if got := q.Get("origin_airport"); got != "SFO" {
			t.Errorf("origin_airport = %q, want %q", got, "SFO")
		}
		if got := q.Get("destination_airport"); got != "BOS" {
			t.Errorf("destination_airport = %q, want %q", got, "BOS")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[],"count":0,"hasMore":false,"cursor":0}`))
	}))
	defer srv.Close()

	client := NewClient("test-key")
	client.BaseURL = srv.URL

	resp, err := client.Search(SearchParams{
		OriginAirport:      "SFO",
		DestinationAirport: "BOS",
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("Count = %d, want 0", resp.Count)
	}
}

func TestClientLive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method.
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		// Verify path.
		if r.URL.Path != "/partnerapi/live" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/partnerapi/live")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()

	client := NewClient("test-key")
	client.BaseURL = srv.URL

	resp, err := client.Live(LiveParams{
		OriginAirport:      "SFO",
		DestinationAirport: "BOS",
		DepartureDate:      "2026-04-01",
		Source:             "united",
	})
	if err != nil {
		t.Fatalf("Live() error = %v", err)
	}
	if len(resp.Results) != 0 {
		t.Errorf("len(Results) = %d, want 0", len(resp.Results))
	}
}

func TestClientErrorHandling(t *testing.T) {
	cases := []struct {
		name         string
		statusCode   int
		wantExitCode int
	}{
		{"bad request", 400, 1},
		{"unauthorized", 401, 2},
		{"rate limited", 429, 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				json.NewEncoder(w).Encode(map[string]string{"error": "test error"})
			}))
			defer srv.Close()

			client := NewClient("test-key")
			client.BaseURL = srv.URL

			_, err := client.Search(SearchParams{
				OriginAirport:      "SFO",
				DestinationAirport: "BOS",
			})
			if err == nil {
				t.Fatalf("expected error for status %d, got nil", tc.statusCode)
			}

			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("expected *APIError, got %T: %v", err, err)
			}
			if apiErr.ExitCode() != tc.wantExitCode {
				t.Errorf("ExitCode() = %d, want %d (status %d)", apiErr.ExitCode(), tc.wantExitCode, tc.statusCode)
			}
		})
	}
}
