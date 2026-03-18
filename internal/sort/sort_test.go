package sort

import (
	"testing"

	"github.com/derek/seats-cli/internal/model"
)

// TestParseSortKeys verifies that "cabin,miles:desc,date" produces 3 keys with correct directions.
func TestParseSortKeys(t *testing.T) {
	keys, err := ParseSortKeys("cabin,miles:desc,date")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}

	// cabin: default direction is desc=true
	if keys[0].Field != "cabin" || !keys[0].Desc {
		t.Errorf("key[0]: want {cabin true}, got {%s %v}", keys[0].Field, keys[0].Desc)
	}
	// miles: default is asc (desc=false), but overridden to :desc
	if keys[1].Field != "miles" || !keys[1].Desc {
		t.Errorf("key[1]: want {miles true}, got {%s %v}", keys[1].Field, keys[1].Desc)
	}
	// date: default direction is desc=false
	if keys[2].Field != "date" || keys[2].Desc {
		t.Errorf("key[2]: want {date false}, got {%s %v}", keys[2].Field, keys[2].Desc)
	}
}

// TestParseSortKeysInvalid verifies that an unrecognised field name returns an error.
func TestParseSortKeysInvalid(t *testing.T) {
	_, err := ParseSortKeys("cabin,invalid_field")
	if err == nil {
		t.Fatal("expected error for invalid field, got nil")
	}
}

// TestParseSortKeysEmpty verifies that an empty string returns the default [miles asc] key.
func TestParseSortKeysEmpty(t *testing.T) {
	keys, err := ParseSortKeys("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 default key, got %d", len(keys))
	}
	if keys[0].Field != "miles" || keys[0].Desc {
		t.Errorf("default key: want {miles false}, got {%s %v}", keys[0].Field, keys[0].Desc)
	}
}

// TestSortFlatRows verifies multi-key sorting by "cabin,miles".
// Input: economy/12500, business/35000, economy/10000, business/32000
// Expected (cabin desc=true groups business before economy, then miles asc within each):
//
//	business/32000, business/35000, economy/10000, economy/12500
func TestSortFlatRows(t *testing.T) {
	rows := []model.FlatRow{
		{Cabin: "economy", Miles: 12500},
		{Cabin: "business", Miles: 35000},
		{Cabin: "economy", Miles: 10000},
		{Cabin: "business", Miles: 32000},
	}

	keys, err := ParseSortKeys("cabin,miles")
	if err != nil {
		t.Fatalf("unexpected error parsing keys: %v", err)
	}

	SortFlatRows(rows, keys)

	want := []struct {
		cabin string
		miles int
	}{
		{"business", 32000},
		{"business", 35000},
		{"economy", 10000},
		{"economy", 12500},
	}

	if len(rows) != len(want) {
		t.Fatalf("expected %d rows, got %d", len(want), len(rows))
	}

	for i, w := range want {
		if rows[i].Cabin != w.cabin || rows[i].Miles != w.miles {
			t.Errorf("row[%d]: want {%s %d}, got {%s %d}", i, w.cabin, w.miles, rows[i].Cabin, rows[i].Miles)
		}
	}
}
