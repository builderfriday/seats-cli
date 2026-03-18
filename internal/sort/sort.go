package sort

import (
	"fmt"
	gosort "sort"
	"strings"

	"github.com/derek/seats-cli/internal/model"
)

// SortKey represents a single sort field with direction.
type SortKey struct {
	Field string
	Desc  bool
}

// fieldDefaults maps valid field names to their default sort direction (true = descending).
var fieldDefaults = map[string]bool{
	"cabin":   true,
	"miles":   false,
	"stops":   false,
	"date":    false,
	"taxes":   false,
	"seats":   true,
	"airline": false,
}

// cabinRank maps cabin names to a numeric rank for ordering.
var cabinRank = map[string]int{
	"economy":  0,
	"premium":  1,
	"business": 2,
	"first":    3,
}

// ParseSortKeys parses a comma-separated sort key string such as "cabin,miles:desc,date".
// Each token may have an optional ":asc" or ":desc" suffix to override the default direction.
// If s is empty, the default [miles asc] key is returned.
// Returns an error if any field name is not recognised.
func ParseSortKeys(s string) ([]SortKey, error) {
	if strings.TrimSpace(s) == "" {
		return []SortKey{{Field: "miles", Desc: false}}, nil
	}

	tokens := strings.Split(s, ",")
	keys := make([]SortKey, 0, len(tokens))

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		var field string
		var desc bool

		if idx := strings.LastIndex(token, ":"); idx != -1 {
			field = token[:idx]
			suffix := strings.ToLower(token[idx+1:])
			switch suffix {
			case "asc":
				desc = false
			case "desc":
				desc = true
			default:
				return nil, fmt.Errorf("invalid sort direction %q in token %q: use :asc or :desc", suffix, token)
			}
		} else {
			field = token
			defaultDesc, ok := fieldDefaults[field]
			if !ok {
				return nil, fmt.Errorf("invalid sort field %q: valid fields are cabin, miles, stops, date, taxes, seats, airline", field)
			}
			desc = defaultDesc
		}

		if _, ok := fieldDefaults[field]; !ok {
			return nil, fmt.Errorf("invalid sort field %q: valid fields are cabin, miles, stops, date, taxes, seats, airline", field)
		}

		keys = append(keys, SortKey{Field: field, Desc: desc})
	}

	if len(keys) == 0 {
		return []SortKey{{Field: "miles", Desc: false}}, nil
	}

	return keys, nil
}

// SortFlatRows sorts rows in-place using the provided keys.
// Keys are applied left-to-right; ties on the first key are broken by the second, and so on.
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

// compareFlatRows returns a negative, zero, or positive integer comparing a and b by field.
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
