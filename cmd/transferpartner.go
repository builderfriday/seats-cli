package cmd

import (
	"fmt"
	"strings"
)

// transferPartnerPrograms maps canonical partner names to their transferable mileage programs.
var transferPartnerPrograms = map[string][]string{
	"Amex":          {"aeromexico", "aeroplan", "delta", "emirates", "etihad", "flyingblue", "jetblue", "qantas", "qatar", "singapore", "virginatlantic"},
	"AmexAustralia": {"emirates", "etihad", "qantas", "qatar", "singapore", "velocity", "virginatlantic"},
	"AmexCanada":    {"aeroplan", "delta", "etihad", "flyingblue"},
	"Bilt":          {"aeroplan", "alaska", "emirates", "etihad", "flyingblue", "qatar", "spirit", "turkish", "united", "virginatlantic"},
	"CapitalOne":    {"aeromexico", "aeroplan", "emirates", "etihad", "finnair", "flyingblue", "jetblue", "qantas", "qatar", "singapore", "turkish", "virginatlantic"},
	"Chase":         {"aeroplan", "flyingblue", "jetblue", "singapore", "united", "virginatlantic"},
	"Citi":          {"aeromexico", "american", "emirates", "etihad", "flyingblue", "jetblue", "qatar", "singapore", "turkish", "virginatlantic"},
	"Rove":          {"aeromexico", "etihad", "eurobonus", "finnair", "flyingblue", "lufthansa", "qatar", "turkish"},
	"WellsFargo":    {"flyingblue", "jetblue", "virginatlantic"},
}

// resolvePartnerAlias maps user-supplied aliases (lowercased) to canonical partner names.
func resolvePartnerAlias(alias string) (string, bool) {
	switch strings.ToLower(alias) {
	case "amex":
		return "Amex", true
	case "amex-australia", "amexaustralia":
		return "AmexAustralia", true
	case "amex-canada", "amexcanada":
		return "AmexCanada", true
	case "bilt":
		return "Bilt", true
	case "capitalone", "capital-one":
		return "CapitalOne", true
	case "chase":
		return "Chase", true
	case "citi":
		return "Citi", true
	case "rove":
		return "Rove", true
	case "wellsfargo", "wells-fargo":
		return "WellsFargo", true
	}
	return "", false
}

// expandTransferPartners takes a comma-separated list of partner aliases and returns
// the union of all transferable programs for those partners.
// If explicitPrograms is non-empty, the result is intersected with those programs.
func expandTransferPartners(transferPartner string, explicitPrograms string) (string, error) {
	partners := strings.Split(transferPartner, ",")
	seen := map[string]bool{}
	var unionPrograms []string

	for _, p := range partners {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		canonical, ok := resolvePartnerAlias(p)
		if !ok {
			return "", fmt.Errorf("unknown transfer partner %q; valid options: amex, amex-australia, amex-canada, bilt, capital-one, chase, citi, rove, wells-fargo", p)
		}
		for _, prog := range transferPartnerPrograms[canonical] {
			if !seen[prog] {
				seen[prog] = true
				unionPrograms = append(unionPrograms, prog)
			}
		}
	}

	if len(unionPrograms) == 0 {
		return "", fmt.Errorf("no programs found for transfer partner(s): %s", transferPartner)
	}

	// If --program was also specified, intersect.
	if explicitPrograms != "" {
		explicit := strings.Split(explicitPrograms, ",")
		explicitSet := map[string]bool{}
		for _, e := range explicit {
			explicitSet[strings.TrimSpace(e)] = true
		}
		var intersected []string
		for _, prog := range unionPrograms {
			if explicitSet[prog] {
				intersected = append(intersected, prog)
			}
		}
		if len(intersected) == 0 {
			return "", fmt.Errorf("intersection of --transfer-partner programs and --program filter is empty")
		}
		return strings.Join(intersected, ","), nil
	}

	return strings.Join(unionPrograms, ","), nil
}
