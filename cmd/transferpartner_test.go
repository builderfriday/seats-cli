package cmd

import (
	"sort"
	"strings"
	"testing"
)

// sortedPrograms splits comma-separated programs and sorts them for stable comparison.
func sortedPrograms(s string) []string {
	parts := strings.Split(s, ",")
	sort.Strings(parts)
	return parts
}

// TestTransferPartnerSingleChase verifies Chase expands to the correct programs.
func TestTransferPartnerSingleChase(t *testing.T) {
	result, err := expandTransferPartners("chase", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := sortedPrograms(result)
	want := []string{"aeroplan", "flyingblue", "jetblue", "singapore", "united", "virginatlantic"}
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("chase: got %v, want %v", got, want)
	}
}

// TestTransferPartnerUnionChaseAndBilt verifies the union of Chase and Bilt is deduplicated.
func TestTransferPartnerUnionChaseAndBilt(t *testing.T) {
	result, err := expandTransferPartners("chase,bilt", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := sortedPrograms(result)

	// Union of Chase + Bilt (deduped, sorted)
	chasePrograms := []string{"aeroplan", "flyingblue", "jetblue", "singapore", "united", "virginatlantic"}
	biltPrograms := []string{"aeroplan", "alaska", "emirates", "etihad", "flyingblue", "qatar", "spirit", "turkish", "united", "virginatlantic"}
	unionSet := map[string]bool{}
	for _, p := range chasePrograms {
		unionSet[p] = true
	}
	for _, p := range biltPrograms {
		unionSet[p] = true
	}
	var want []string
	for p := range unionSet {
		want = append(want, p)
	}
	sort.Strings(want)

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("chase,bilt union: got %v, want %v", got, want)
	}
}

// TestTransferPartnerIntersectionWithProgram verifies --program intersection logic.
func TestTransferPartnerIntersectionWithProgram(t *testing.T) {
	// Chase programs: aeroplan, flyingblue, jetblue, singapore, united, virginatlantic
	// --program filter: united,delta
	// Intersection: united (delta is not in Chase)
	result, err := expandTransferPartners("chase", "united,delta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "united" {
		t.Errorf("intersection chase + united,delta: got %q, want %q", result, "united")
	}
}

// TestTransferPartnerCaseInsensitive verifies Chase, CHASE, chase all resolve identically.
func TestTransferPartnerCaseInsensitive(t *testing.T) {
	variants := []string{"chase", "Chase", "CHASE"}
	var results []string
	for _, v := range variants {
		r, err := expandTransferPartners(v, "")
		if err != nil {
			t.Fatalf("variant %q: unexpected error: %v", v, err)
		}
		results = append(results, r)
	}
	for i := 1; i < len(results); i++ {
		got := sortedPrograms(results[i])
		base := sortedPrograms(results[0])
		if strings.Join(got, ",") != strings.Join(base, ",") {
			t.Errorf("variant %q: got %v, want %v", variants[i], got, base)
		}
	}
}

// TestTransferPartnerCapitalOneAliases verifies capital-one, capitalone, CapitalOne all work.
func TestTransferPartnerCapitalOneAliases(t *testing.T) {
	variants := []string{"capital-one", "capitalone", "CapitalOne"}
	var results []string
	for _, v := range variants {
		r, err := expandTransferPartners(v, "")
		if err != nil {
			t.Fatalf("variant %q: unexpected error: %v", v, err)
		}
		results = append(results, r)
	}
	for i := 1; i < len(results); i++ {
		got := sortedPrograms(results[i])
		base := sortedPrograms(results[0])
		if strings.Join(got, ",") != strings.Join(base, ",") {
			t.Errorf("CapitalOne alias %q: got %v, want %v", variants[i], got, base)
		}
	}
}

// TestTransferPartnerUnknownReturnsError verifies that an unknown partner returns an error.
func TestTransferPartnerUnknownReturnsError(t *testing.T) {
	_, err := expandTransferPartners("notabank", "")
	if err == nil {
		t.Fatal("expected error for unknown partner, got nil")
	}
	if !strings.Contains(err.Error(), "unknown transfer partner") {
		t.Errorf("error message should mention 'unknown transfer partner', got: %v", err)
	}
}

// TestTransferPartnerEmptyIntersection verifies that an all-empty intersection returns an error.
func TestTransferPartnerEmptyIntersection(t *testing.T) {
	// Chase has no delta; intersection should be empty → error
	_, err := expandTransferPartners("chase", "delta")
	if err == nil {
		t.Fatal("expected error for empty intersection, got nil")
	}
	if !strings.Contains(err.Error(), "intersection") {
		t.Errorf("error message should mention 'intersection', got: %v", err)
	}
}

// TestTransferPartnerAllMappings validates every partner against the canonical mapping.
func TestTransferPartnerAllMappings(t *testing.T) {
	canonicalWant := map[string][]string{
		"amex":           {"aeromexico", "aeroplan", "delta", "emirates", "etihad", "flyingblue", "jetblue", "qantas", "qatar", "singapore", "virginatlantic"},
		"amex-australia": {"emirates", "etihad", "qantas", "qatar", "singapore", "velocity", "virginatlantic"},
		"amex-canada":    {"aeroplan", "delta", "etihad", "flyingblue"},
		"bilt":           {"aeroplan", "alaska", "emirates", "etihad", "flyingblue", "qatar", "spirit", "turkish", "united", "virginatlantic"},
		"capital-one":    {"aeromexico", "aeroplan", "emirates", "etihad", "finnair", "flyingblue", "jetblue", "qantas", "qatar", "singapore", "turkish", "virginatlantic"},
		"chase":          {"aeroplan", "flyingblue", "jetblue", "singapore", "united", "virginatlantic"},
		"citi":           {"aeromexico", "american", "emirates", "etihad", "flyingblue", "jetblue", "qatar", "singapore", "turkish", "virginatlantic"},
		"rove":           {"aeromexico", "etihad", "eurobonus", "finnair", "flyingblue", "lufthansa", "qatar", "turkish"},
		"wells-fargo":    {"flyingblue", "jetblue", "virginatlantic"},
	}

	for alias, want := range canonicalWant {
		t.Run(alias, func(t *testing.T) {
			result, err := expandTransferPartners(alias, "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := sortedPrograms(result)
			wantSorted := make([]string, len(want))
			copy(wantSorted, want)
			sort.Strings(wantSorted)
			if strings.Join(got, ",") != strings.Join(wantSorted, ",") {
				t.Errorf("%s: got %v, want %v", alias, got, wantSorted)
			}
		})
	}
}
