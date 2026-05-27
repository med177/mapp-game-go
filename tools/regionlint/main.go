package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type settlement struct {
	ID     string `json:"id"`
	NameTR string `json:"name_tr"`
}

type region struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	NameTR      string       `json:"name_tr"`
	IsSea       bool         `json:"is_sea"`
	Settlements []settlement `json:"settlements"`
}

var (
	snakeCaseNameRe  = regexp.MustCompile(`^[a-z0-9_]+$`)
	placeholderNumRe = regexp.MustCompile(`\b\d+\b`)
)

func main() {
	defaultPath := "assets/scenarios/1300_ottoman_rise/data/regions.json"
	path := flag.String("file", defaultPath, "Kontrol edilecek regions.json yolu")
	flag.Parse()

	data, err := os.ReadFile(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dosya okunamadı: %v\n", err)
		os.Exit(2)
	}

	var regions []region
	if err := json.Unmarshal(data, &regions); err != nil {
		fmt.Fprintf(os.Stderr, "JSON parse hatası: %v\n", err)
		os.Exit(2)
	}

	var issues []string
	for _, r := range regions {
		issues = append(issues, checkRegion(r)...)
	}

	if len(issues) == 0 {
		fmt.Printf("OK: %d bölge kontrol edildi, sorun bulunmadı.\n", len(regions))
		return
	}

	fmt.Printf("FAIL: %d sorun bulundu.\n", len(issues))
	for _, issue := range issues {
		fmt.Println(issue)
	}
	os.Exit(1)
}

func checkRegion(r region) []string {
	var issues []string

	if strings.TrimSpace(r.Name) == "" {
		issues = append(issues, fmtIssue(r.ID, "name boş"))
	}
	if snakeCaseNameRe.MatchString(r.Name) {
		issues = append(issues, fmtIssue(r.ID, fmt.Sprintf("name snake_case görünüyor: %q", r.Name)))
	}
	if strings.Contains(r.Name, "_") {
		issues = append(issues, fmtIssue(r.ID, fmt.Sprintf("name içinde '_' var: %q", r.Name)))
	}

	if strings.TrimSpace(r.NameTR) == "" {
		issues = append(issues, fmtIssue(r.ID, "name_tr boş"))
	}
	issues = append(issues, lintNameTR(r.ID, "region.name_tr", r.NameTR, r.IsSea)...)

	for _, s := range r.Settlements {
		if strings.TrimSpace(s.NameTR) == "" {
			issues = append(issues, fmtIssue(r.ID, fmt.Sprintf("settlement(%s).name_tr boş", s.ID)))
			continue
		}
		issues = append(issues, lintNameTR(r.ID, "settlement("+s.ID+").name_tr", s.NameTR, false)...)
	}

	return issues
}

func lintNameTR(regionID, field, value string, allowNumbered bool) []string {
	var issues []string
	lower := strings.ToLower(value)

	asciiArtifacts := []string{
		"kralligi", "kiyisi", "korpezi", "norvec", "dag", "sinir", "yakasi",
		"acik", "ilhanli", "altin", "memluk", "sirp", "ogullari", "despotlugu",
	}
	for _, token := range asciiArtifacts {
		if strings.Contains(lower, token) {
			issues = append(issues, fmtIssue(regionID, fmt.Sprintf("%s ASCII kalıntı içeriyor (%s): %q", field, token, value)))
			break
		}
	}

	if !allowNumbered && placeholderNumRe.MatchString(value) {
		issues = append(issues, fmtIssue(regionID, fmt.Sprintf("%s sayı/placeholder içeriyor: %q", field, value)))
	}
	if strings.Contains(value, "  ") {
		issues = append(issues, fmtIssue(regionID, fmt.Sprintf("%s çift boşluk içeriyor: %q", field, value)))
	}

	return issues
}

func fmtIssue(regionID, msg string) string {
	return fmt.Sprintf("- region=%s: %s", regionID, msg)
}
