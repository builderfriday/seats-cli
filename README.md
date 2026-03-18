# seats

A CLI tool for searching award flight availability across 20+ mileage programs via the [Seats.aero](https://seats.aero) Partner API.

Designed primarily for AI agent consumption (markdown output by default), with optional pretty terminal rendering for humans.

## Install

```bash
go install github.com/derek/seats-cli@latest
```

Or build from source:

```bash
git clone https://github.com/derek/seats-cli.git
cd seats-cli
go install .
```

## Setup

Set your Seats.aero API key as an environment variable:

```bash
export SEATS_AERO_API_KEY=your_key_here
```

Pro users can generate a key from the [API tab in settings](https://seats.aero/settings) (1,000 calls/day).

## Commands

### `seats search` — Cached award search

Search cached availability for specific airports and dates.

```bash
# All programs, SFO to BOS
seats search --from SFO --to BOS

# Date range with sorting
seats search --from SFO --to BOS --date 2026-04-01 --end-date 2026-04-03 --sort cabin,miles

# Direct flights only, specific program
seats search --from SFO --to BOS --direct --program united --sort miles

# Multiple airports and programs
seats search --from SFO,OAK --to BOS,JFK --program united,aeroplan
```

| Flag | Description |
|------|-------------|
| `--from` | Origin airport(s), comma-separated (required) |
| `--to` | Destination airport(s), comma-separated (required) |
| `--date` | Start date (YYYY-MM-DD) |
| `--end-date` | End date (YYYY-MM-DD) |
| `--cabin` | economy, premium, business, first (comma-separated) |
| `--program` | Mileage program filter (comma-separated) |
| `--carrier` | Airline filter (comma-separated) |
| `--direct` | Nonstop flights only |
| `--sort` | Sort keys (see Sorting) |
| `--limit` | Max results, 10-1000 (default: 50) |
| `--skip` | Skip N results for pagination |

### `seats live` — Real-time search

Query a mileage program in real-time for a specific route and date. Requires commercial API key.

```bash
seats live --from SFO --to BOS --date 2026-04-01 --program united --seats 2
```

### `seats availability` — Bulk program availability

Browse all cached availability for a mileage program.

```bash
# All American business class next week
seats availability --program american --cabin business --date 2026-04-01 --end-date 2026-04-07

# Filter by region
seats availability --program united --origin-region "North America" --dest-region "Asia"
```

### `seats trip <id>` — Trip details

Get flight segments, taxes, and booking links for a specific availability result.

```bash
seats trip 2woPm8LP2uJ33okIQzMJoikCMC8
```

### `seats routes` — Program routes

List all routes a mileage program covers.

```bash
seats routes --program aeroplan
```

### `seats programs` — List programs

List all available mileage programs and their identifiers.

```bash
seats programs
```

## Sorting

Available on `search` and `availability` via `--sort`. Comma-separated, multi-key:

```bash
--sort cabin,miles,date
```

| Key | Default | Description |
|-----|---------|-------------|
| `cabin` | desc | first > business > premium > economy |
| `miles` | asc | Mileage cost (cheapest first) |
| `stops` | asc | Fewest stops first |
| `date` | asc | Earliest first |
| `taxes` | asc | Lowest fees first |
| `seats` | desc | Most remaining seats first |
| `airline` | asc | Alphabetical |

Override direction: `--sort miles:desc,cabin:asc`

## Output

**Markdown** (default) — agent-friendly, results to stdout, errors to stderr.

**Pretty** (`--pretty`) — styled terminal tables with borders and colors.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (including zero results) |
| 1 | Bad request (invalid flags or API 400) |
| 2 | Auth error (missing or invalid API key) |
| 3 | Rate limited (1,000 calls/day) |
| 4 | Network error |
