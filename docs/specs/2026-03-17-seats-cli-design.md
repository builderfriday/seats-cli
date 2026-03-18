# seats CLI — Design Spec

## Overview

A Go-based CLI tool for searching award flight availability via the [Seats.aero Partner API](https://developers.seats.aero/reference/getting-started-p). Designed primarily for AI agent consumption, with markdown output by default and optional pretty terminal rendering for human use.

## Goals

- Provide programmatic access to Seats.aero's award search across 20+ mileage programs
- Agent-first: markdown output, clean stdout/stderr separation, predictable flag names
- Thin API wrapper: each subcommand maps 1:1 to an API endpoint
- Good terminal aesthetics when used by humans (`--pretty` mode via Charm stack)

## Non-Goals

- No config files, caching layer, or database
- No interactive TUI mode
- Not a replacement for cash fare searches (that's `fli`)

---

## Authentication

- **Environment variable only**: `SEATS_AERO_API_KEY`
- Validated on startup; clear error to stderr if missing
- Pro users: 1,000 API calls/day

---

## Subcommands

### `seats search`

Searches cached award availability for specific airports and dates.

**Use case**: "What award flights exist SFO -> BOS on Apr 3?"

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--from` | string | yes | Origin airport(s), comma-separated |
| `--to` | string | yes | Destination airport(s), comma-separated |
| `--date` | string | no | Start date (YYYY-MM-DD) |
| `--end-date` | string | no | End date (YYYY-MM-DD) |
| `--cabin` | string | no | Comma-separated: economy, premium, business, first |
| `--program` | string | no | Mileage program filter, comma-separated for multiple (e.g. `--program united,aeroplan`) |
| `--carrier` | string | no | Airline filter, comma-separated |
| `--direct` | bool | no | Nonstop only |
| `--sort` | string | no | Multi-key sort (see Sorting) |
| `--limit` | int | no | Max results, range 10-1000 (default: 50) |
| `--skip` | int | no | Skip N results for pagination (default: 0) |
| `--pretty` | bool | no | Styled terminal output |

**API mapping**: `GET /partnerapi/search`
- `--from` -> `origin_airport`
- `--to` -> `destination_airport`
- `--date` -> `start_date`
- `--end-date` -> `end_date`
- `--cabin` -> `cabins`
- `--program` -> `sources`
- `--carrier` -> `carriers`
- `--direct` -> `only_direct_flights`
- `--limit` -> `take`
- `--skip` -> `skip`

**Data transformation**: The API returns one object per route/date with all cabin classes (Y/W/J/F) as separate fields. The CLI flattens this into one row per cabin class that has availability, so a single API result may produce up to 4 output rows. This makes sorting and filtering by cabin intuitive.

### `seats live`

Real-time search for a specific route, date, and mileage program.

**Use case**: "What does United actually have right now for SFO -> BOS on Apr 3?"

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--from` | string | yes | Origin airport |
| `--to` | string | yes | Destination airport |
| `--date` | string | yes | Departure date (YYYY-MM-DD) |
| `--program` | string | yes | Mileage program (single value only) |
| `--seats` | int | no | Passenger count, 1-9 (default: 1) |
| `--pretty` | bool | no | Styled terminal output |

**API mapping**: `POST /partnerapi/live`
- `--from` -> `origin_airport`
- `--to` -> `destination_airport`
- `--date` -> `departure_date`
- `--program` -> `source`
- `--seats` -> `seat_count`

**Note**: Unlike `search`, `--program` accepts only a single program since the live endpoint queries one mileage program in real-time. The response returns trip-level data (segments, duration, stops) rather than per-cabin availability summaries — see output format below.

### `seats availability`

Bulk dump of all cached availability for a mileage program.

**Use case**: "I have American miles — where can I fly business class next week?"

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--program` | string | yes | Mileage program |
| `--cabin` | string | no | economy, premium, business, first |
| `--date` | string | no | Start date (YYYY-MM-DD) |
| `--end-date` | string | no | End date (YYYY-MM-DD) |
| `--origin-region` | string | no | North America, South America, Africa, Asia, Europe, Oceania |
| `--dest-region` | string | no | Same options as origin-region |
| `--sort` | string | no | Multi-key sort (see Sorting) |
| `--limit` | int | no | Max results, range 10-1000 (default: 50) |
| `--skip` | int | no | Skip N results for pagination (default: 0) |
| `--pretty` | bool | no | Styled terminal output |

**API mapping**: `GET /partnerapi/availability`
- `--program` -> `source`
- `--cabin` -> `cabin`
- `--date` -> `start_date`
- `--end-date` -> `end_date`
- `--origin-region` -> `origin_region`
- `--dest-region` -> `destination_region`
- `--limit` -> `take`
- `--skip` -> `skip`

**Note**: This endpoint does not support airport-level filtering (`--from`/`--to`). Use `--origin-region`/`--dest-region` for geographic filtering, or use `seats search` for airport-specific queries.

### `seats routes`

Lists all routes a mileage program covers.

**Use case**: "What routes does Aeroplan have in their system?"

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--program` | string | yes | Mileage program |
| `--pretty` | bool | no | Styled terminal output |

**API mapping**: `GET /partnerapi/routes`
- `--program` -> `source`

### `seats trip <availability-id>`

Gets detailed segments, taxes, and booking links for a specific result.

**Use case**: "Show me flight details and how to book this option."

| Arg/Flag | Type | Required | Description |
|----------|------|----------|-------------|
| `<id>` | string | yes | Availability ID (from search/availability results) |
| `--pretty` | bool | no | Styled terminal output |

**API mapping**: `GET /partnerapi/trips/{id}`

### `seats programs`

Lists all available mileage programs and their identifiers.

**Use case**: "What are the valid `--program` values?"

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--pretty` | bool | no | Styled terminal output |

**Implementation**: Hardcoded list of known programs with human-readable names. Updated when Seats.aero adds new programs. Source of truth: [API docs](https://developers.seats.aero/reference/getting-started-p).

---

## Output Format

### Default: Markdown

All output is markdown by default (agent-first design). Results go to stdout, errors to stderr.

#### `seats search`

The API returns one object per route/date with all cabins. The CLI flattens to one row per cabin with availability, adding an `ID` column for piping to `seats trip`.

```markdown
## Award Search: SFO -> BOS (2026-04-03)

| ID | Date | Program | Cabin | Miles | Airline | Direct | Seats |
|----|------|---------|-------|-------|---------|--------|-------|
| 2PPrELk9Wc | 2026-04-03 | united | economy | 12,500 | UA | yes | 3 |
| 2PPrELk9Wc | 2026-04-03 | united | business | 35,000 | UA | yes | 2 |
| 2QghXwB1Qj | 2026-04-03 | aeroplan | economy | 10,000 | UA | yes | 3 |
| 2QghXwB1Qj | 2026-04-03 | aeroplan | business | 32,000 | UA | yes | 1 |

4 results (from 2 availability records). More available — use --skip 50 to paginate.
```

#### `seats live`

Returns trip-level data with segments, duration, and stops (different shape from `search`).

```markdown
## Live Search: SFO -> BOS (2026-04-03, united)

| Flight | Cabin | Miles | Taxes | Stops | Duration | Departs | Arrives | Seats |
|--------|-------|-------|-------|-------|----------|---------|---------|-------|
| UA 123 | economy | 12,500 | $5.60 | 0 | 5h 30m | 08:00 | 16:30 | 3 |
| UA 456 | economy | 12,500 | $5.60 | 0 | 5h 45m | 14:00 | 22:45 | 5 |
| UA 123 | business | 35,000 | $5.60 | 0 | 5h 30m | 08:00 | 16:30 | 2 |

3 results found.
```

#### `seats availability`

Shows all cabins as columns since this is a multi-route overview. Miles shown with remaining seats in parentheses.

```markdown
## Bulk Availability: american (2026-04-01 -> 2026-04-07, business)

| ID | Date | Route | Economy | Business | First | Direct |
|----|------|-------|---------|----------|-------|--------|
| 2PPrELk9Wc | 2026-04-03 | DFW -> NRT | 35,000 (3) | 70,000 (2) | 120,000 (1) | yes |
| 2QghXwB1Qj | 2026-04-04 | JFK -> LHR | 30,000 (5) | 57,500 (2) | — | yes |
| 2S8Cm9dHOR | 2026-04-05 | LAX -> SYD | 40,000 (1) | 80,000 (1) | — | no |

3 results returned. More available — use --skip 50 to paginate.
```

#### `seats trip`

Returns all trip options for an availability ID. May include multiple routing/cabin options.

```markdown
## Trip Details: 2PPrELk9Wc

### Option 1: economy — 12,500 miles

| Segment | Flight | Route | Departs | Arrives | Aircraft |
|---------|--------|-------|---------|---------|----------|
| 1 | UA 123 | SFO -> BOS | 08:00 | 16:30 | 737-900 |

- **Duration**: 5h 30m | **Stops**: 0 | **Taxes**: $5.60 USD

### Option 2: business — 35,000 miles

| Segment | Flight | Route | Departs | Arrives | Aircraft |
|---------|--------|-------|---------|---------|----------|
| 1 | UA 123 | SFO -> BOS | 08:00 | 16:30 | 737-900 |

- **Duration**: 5h 30m | **Stops**: 0 | **Taxes**: $5.60 USD

### Booking Links
- [United MileagePlus](https://...) (primary)
```

#### `seats routes`

```markdown
## Routes: aeroplan

| Origin | Destination | Origin Region | Dest Region | Distance | Days Out |
|--------|-------------|---------------|-------------|----------|----------|
| SFO | NRT | North America | Asia | 5,130 | 60 |
| YYZ | LHR | North America | Europe | 3,544 | 90 |
| TPE | PNH | Asia | Asia | 1,423 | 60 |

312 routes found.
```

#### `seats programs`

```markdown
## Available Programs

| Program | Name |
|---------|------|
| aeroplan | Air Canada Aeroplan |
| alaska | Alaska Mileage Plan |
| american | American AAdvantage |
| delta | Delta SkyMiles |
| united | United MileagePlus |
```

### Pretty Mode (`--pretty`)

Same data rendered with Lip Gloss styled tables, colored headers, and box-drawing characters.

---

## Sorting

Available on `search` and `availability` subcommands via `--sort`.

### Multi-key sort

Comma-separated keys applied in priority order:

```
--sort cabin,miles,stops,date
```

"Group by class, then cheapest miles, then fewest stops, then earliest date."

### Available sort keys

| Key | Default direction | Description |
|-----|-------------------|-------------|
| `cabin` | desc (first > business > premium > economy) | Cabin class tier |
| `miles` | asc (cheapest first) | Mileage cost |
| `stops` | asc (fewest first) | Number of stops |
| `date` | asc (earliest first) | Departure date |
| `taxes` | asc (lowest first) | Tax/fee amount |
| `seats` | desc (most first) | Remaining seats |
| `airline` | asc (alphabetical) | Operating airline |

### Direction override

Append `:asc` or `:desc` to any key:

```
--sort cabin,miles:desc,date
```

### Default sort

When `--sort` is omitted: `miles` (lowest mileage cost first).

---

## Error Handling

- Errors to stderr (prefixed with `Error: `), results to stdout
- No automatic retries — caller (agent) decides retry strategy

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (including zero results) |
| 1 | Bad request — invalid flags or API returned 400 |
| 2 | Auth error — missing or invalid API key |
| 3 | Rate limited — 1,000 calls/day exceeded |
| 4 | Network error — could not reach seats.aero |

### Error Messages

```
# Missing API key (exit 2)
Error: SEATS_AERO_API_KEY environment variable not set.

# Invalid parameters (exit 1)
Error: API returned 400 — check airport codes.

# No results (exit 0 — not an error)
No results found.

# Rate limited (exit 3)
Error: Rate limit exceeded (1,000 calls/day). Try again tomorrow.

# Network failure (exit 4)
Error: Request failed — could not reach seats.aero.
```

---

## Project Structure

```
seats-cli/
├── cmd/
│   ├── root.go           # root command, global flags, env var validation
│   ├── search.go         # seats search
│   ├── live.go           # seats live
│   ├── availability.go   # seats availability
│   ├── routes.go         # seats routes
│   ├── trip.go           # seats trip <id>
│   └── programs.go       # seats programs
├── internal/
│   ├── api/
│   │   └── client.go     # HTTP client, auth, request/response handling
│   ├── format/
│   │   ├── markdown.go   # markdown table rendering
│   │   └── pretty.go     # Lip Gloss styled output
│   └── model/
│       └── types.go      # API response structs
├── go.mod
├── go.sum
└── main.go               # entry point
```

### Dependencies

- `github.com/spf13/cobra` — subcommands and flag parsing
- `github.com/charmbracelet/lipgloss` — styled terminal output
- `github.com/charmbracelet/lipgloss/table` — table rendering

---

## Typical Agent Workflow

1. `seats search --from SFO --to BOS --date 2026-04-03 --sort cabin,miles` — find options
2. Pick an availability ID from results
3. `seats trip <id>` — get segments, taxes, booking links
4. Present recommendation to user with booking link

## API Reference

- Base URL: `https://seats.aero/partnerapi`
- Auth header: `Partner-Authorization: <api-key>`
- Rate limit: 1,000 calls/day (Pro tier)
- All responses: JSON
- [Full API docs](https://developers.seats.aero/reference/getting-started-p)
