package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

type region struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	IsSea  bool     `json:"is_sea"`
	Neighs []string `json:"neighbors"`
}

func main() {
	baseDir := flag.String("dir", "assets/scenarios/1300_ottoman_rise/data", "Senaryo data klasörü")
	flag.Parse()

	regionsPath := filepath.Join(*baseDir, "regions.json")
	raw, err := os.ReadFile(regionsPath)
	if err != nil {
		fail("regions.json okunamadı: %v", err)
	}

	var regions []region
	if err := json.Unmarshal(raw, &regions); err != nil {
		fail("regions.json parse hatası: %v", err)
	}

	mapping, changed := buildMapping(regions)
	if len(changed) == 0 {
		fmt.Println("ID değişikliği gerekmiyor.")
		return
	}

	files, err := filepath.Glob(filepath.Join(*baseDir, "*.json"))
	if err != nil {
		fail("json dosyaları listelenemedi: %v", err)
	}
	slices.Sort(files)

	for _, p := range files {
		if err := rewriteJSONFile(p, mapping); err != nil {
			fail("dosya güncellenemedi (%s): %v", p, err)
		}
	}

	fmt.Printf("%d region id güncellendi.\n", len(changed))
	for _, c := range changed {
		fmt.Println(c)
	}
	fmt.Printf("%d JSON dosyası eşzamanlandı.\n", len(files))
}

func buildMapping(regions []region) (map[string]string, []string) {
	mapping := make(map[string]string, len(regions))
	targetOwner := make(map[string]string, len(regions))
	var changed []string

	for _, r := range regions {
		if r.ID == "" {
			continue
		}
		newID := r.ID
		slug := slugify(r.Name)
		if slug != "" {
			newID = slug
		}

		if owner, exists := targetOwner[newID]; exists && owner != r.ID {
			// Çakışmada ilk sahibini koru, ikinciyi eski id ile bırak.
			newID = r.ID
		} else {
			targetOwner[newID] = r.ID
		}

		mapping[r.ID] = newID
		if newID != r.ID {
			changed = append(changed, fmt.Sprintf("- %s -> %s", r.ID, newID))
		}
	}

	slices.Sort(changed)
	return mapping, changed
}

func rewriteJSONFile(path string, mapping map[string]string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}

	v = rewriteAny(v, mapping)
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0o644)
}

func rewriteAny(v any, mapping map[string]string) any {
	switch t := v.(type) {
	case map[string]any:
		for k, vv := range t {
			t[k] = rewriteAny(vv, mapping)
		}
		return t
	case []any:
		for i := range t {
			t[i] = rewriteAny(t[i], mapping)
		}
		return t
	case string:
		if nv, ok := mapping[t]; ok {
			return nv
		}
		return t
	default:
		return v
	}
}

func slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return ""
	}

	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		if rr, ok := translitRune(r); ok {
			r = rr
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// ASCII dışı harf/rakamlar normalize edilemediyse atla.
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}

	out := strings.Trim(b.String(), "_")
	out = strings.ReplaceAll(out, "__", "_")
	return out
}

func translitRune(r rune) (rune, bool) {
	switch r {
	case 'ç':
		return 'c', true
	case 'ğ':
		return 'g', true
	case 'ı':
		return 'i', true
	case 'ö':
		return 'o', true
	case 'ş':
		return 's', true
	case 'ü':
		return 'u', true
	case 'â', 'á', 'à', 'ä':
		return 'a', true
	case 'é', 'è', 'ê', 'ë':
		return 'e', true
	case 'í', 'ì', 'î', 'ï':
		return 'i', true
	case 'ó', 'ò', 'ô', 'õ', 'ø':
		return 'o', true
	case 'ú', 'ù', 'û':
		return 'u', true
	case 'ñ':
		return 'n', true
	case 'ß':
		return 's', true
	default:
		return 0, false
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
