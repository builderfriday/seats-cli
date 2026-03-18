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

// APIError represents an HTTP-level error from the seats.aero API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

// ExitCode maps HTTP status codes to CLI exit codes.
//
//	401/403 → 2 (auth failure)
//	429     → 3 (rate limit)
//	400-499 → 1 (client error)
//	else    → 4 (unexpected)
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

// Client is an HTTP client for the seats.aero partner API.
type Client struct {
	BaseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient returns a Client configured for the seats.aero production API.
func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL:    "https://seats.aero",
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// SearchParams holds query parameters for the /partnerapi/search endpoint.
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

// LiveParams holds the request body for the /partnerapi/live endpoint.
type LiveParams struct {
	OriginAirport      string `json:"origin_airport"`
	DestinationAirport string `json:"destination_airport"`
	DepartureDate      string `json:"departure_date"`
	Source             string `json:"source"`
	SeatCount          int    `json:"seat_count,omitempty"`
}

// AvailabilityParams holds query parameters for the /partnerapi/availability endpoint.
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

// Search calls GET /partnerapi/search and returns cached availability results.
func (c *Client) Search(p SearchParams) (*model.SearchResponse, error) {
	q := url.Values{}
	if p.OriginAirport != "" {
		q.Set("origin_airport", p.OriginAirport)
	}
	if p.DestinationAirport != "" {
		q.Set("destination_airport", p.DestinationAirport)
	}
	if p.StartDate != "" {
		q.Set("start_date", p.StartDate)
	}
	if p.EndDate != "" {
		q.Set("end_date", p.EndDate)
	}
	if p.Cabins != "" {
		q.Set("cabins", p.Cabins)
	}
	if p.Sources != "" {
		q.Set("sources", p.Sources)
	}
	if p.Carriers != "" {
		q.Set("carriers", p.Carriers)
	}
	if p.OnlyDirectFlights {
		q.Set("only_direct_flights", "true")
	}
	if p.Take != 0 {
		q.Set("take", strconv.Itoa(p.Take))
	}
	if p.Skip != 0 {
		q.Set("skip", strconv.Itoa(p.Skip))
	}

	var resp model.SearchResponse
	if err := c.get("/partnerapi/search", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Live calls POST /partnerapi/live and returns real-time availability results.
func (c *Client) Live(p LiveParams) (*model.LiveResponse, error) {
	var resp model.LiveResponse
	if err := c.post("/partnerapi/live", p, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Availability calls GET /partnerapi/availability and returns broad availability data.
func (c *Client) Availability(p AvailabilityParams) (*model.SearchResponse, error) {
	q := url.Values{}
	if p.Source != "" {
		q.Set("source", p.Source)
	}
	if p.Cabin != "" {
		q.Set("cabin", p.Cabin)
	}
	if p.StartDate != "" {
		q.Set("start_date", p.StartDate)
	}
	if p.EndDate != "" {
		q.Set("end_date", p.EndDate)
	}
	if p.OriginRegion != "" {
		q.Set("origin_region", p.OriginRegion)
	}
	if p.DestinationRegion != "" {
		q.Set("destination_region", p.DestinationRegion)
	}
	if p.Take != 0 {
		q.Set("take", strconv.Itoa(p.Take))
	}
	if p.Skip != 0 {
		q.Set("skip", strconv.Itoa(p.Skip))
	}

	var resp model.SearchResponse
	if err := c.get("/partnerapi/availability", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Routes calls GET /partnerapi/routes and returns routes for the given source program.
func (c *Client) Routes(source string) ([]model.Route, error) {
	q := url.Values{}
	if source != "" {
		q.Set("source", source)
	}

	var resp []model.Route
	if err := c.get("/partnerapi/routes", q, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Trip calls GET /partnerapi/trips/{id} and returns detailed trip information.
func (c *Client) Trip(id string) (*model.TripResponse, error) {
	var resp model.TripResponse
	if err := c.get("/partnerapi/trips/"+id, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// get executes a GET request to the given path with the provided query parameters.
func (c *Client) get(path string, params url.Values, result interface{}) error {
	u := c.BaseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return &APIError{StatusCode: 0, Message: "Request failed — could not reach seats.aero."}
	}

	return c.do(req, result)
}

// post executes a POST request to the given path with a JSON-encoded body.
func (c *Client) post(path string, body interface{}, result interface{}) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(encoded))
	if err != nil {
		return &APIError{StatusCode: 0, Message: "Request failed — could not reach seats.aero."}
	}
	req.Header.Set("Content-Type", "application/json")

	return c.do(req, result)
}

// do sets the auth header, executes the request, and decodes the response.
// Non-2xx responses are converted to *APIError values.
func (c *Client) do(req *http.Request, result interface{}) error {
	req.Header.Set("Partner-Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIError{StatusCode: 0, Message: "Request failed — could not reach seats.aero."}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return c.apiError(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{StatusCode: 0, Message: "Request failed — could not reach seats.aero."}
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// apiError constructs an *APIError with the canonical message for a given status code.
func (c *Client) apiError(code int) *APIError {
	var msg string
	switch {
	case code == 401 || code == 403:
		msg = "Invalid or missing API key."
	case code == 429:
		msg = "Rate limit exceeded (1,000 calls/day). Try again tomorrow."
	default:
		msg = fmt.Sprintf("API returned %d — check airport codes.", code)
	}
	return &APIError{StatusCode: code, Message: msg}
}
