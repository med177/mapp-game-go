package world

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// LandPassageType karasal iki bölge arasındaki özel geçişin türünü tanımlar.
// Şimdilik yalnızca boğaz görselleştirmesi kullanılıyor; diğer türler ileride
// hareket ve savaş kurallarına bağlanabilir.
type LandPassageType string

const (
	LandPassageStrait LandPassageType = "strait"
)

// LandPassage iki kara bölgesi arasındaki, haritada doğrudan sınır paylaşmasa
// bile kullanılabilen özel karasal bağlantıyı temsil eder.
type LandPassage struct {
	From         RegionID        `json:"from"`
	To           RegionID        `json:"to"`
	Type         LandPassageType `json:"type"`
	MoveCost     int             `json:"move_cost"`
	DefenseBonus int             `json:"defense_bonus"`
	// Senaryo koordinatlarında [x,y] biçimindeki çizgi uçları. Eski kayıtlarla
	// uyumluluk için isteğe bağlıdır; yoksa bölge anchor'ı kullanılır.
	Start *[2]int `json:"start,omitempty"`
	End   *[2]int `json:"end,omitempty"`
}

// HasCustomEndpoints geçişin haritada açıkça tanımlanmış uçları olup olmadığını
// döner.
func (p LandPassage) HasCustomEndpoints() bool {
	return p.Start != nil && p.End != nil
}

// LoadLandPassages opsiyonel land_passages.json dosyasını okur.
// Geçersiz, deniz bölgesine bağlanan, kendine bağlanan veya yinelenen kayıtlar
// sessizce atılır; dosyanın biçimindeki gerçek hatalar hata olarak döner.
func LoadLandPassages(path string, regions map[RegionID]*Region) ([]LandPassage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("land_passages dosyası okunamadı: %w", err)
	}

	var passages []LandPassage
	if err := json.Unmarshal(data, &passages); err != nil {
		return nil, fmt.Errorf("land_passages JSON parse hatası: %w", err)
	}

	seen := make(map[string]bool, len(passages))
	valid := make([]LandPassage, 0, len(passages))
	for _, passage := range passages {
		from, fromOK := regions[passage.From]
		to, toOK := regions[passage.To]
		if !fromOK || !toOK || from == nil || to == nil || from.IsSea || to.IsSea ||
			passage.From == "" || passage.To == "" || passage.From == passage.To {
			continue
		}
		if passage.Type == "" {
			passage.Type = LandPassageStrait
		}
		if passage.MoveCost <= 0 {
			passage.MoveCost = 1
		}
		key := undirectedLandPassageKey(passage.From, passage.To)
		if seen[key] {
			continue
		}
		seen[key] = true
		valid = append(valid, passage)
	}

	sort.SliceStable(valid, func(i, j int) bool {
		if valid[i].From != valid[j].From {
			return valid[i].From < valid[j].From
		}
		return valid[i].To < valid[j].To
	})
	return valid, nil
}

// HasLandPassage iki bölge arasında yön bağımsız özel geçiş olup olmadığını döner.
func HasLandPassage(passages []LandPassage, from, to RegionID) bool {
	return LandPassageBetween(passages, from, to) != nil
}

// LandPassageBetween iki bölge arasındaki yön bağımsız özel geçişi döner.
func LandPassageBetween(passages []LandPassage, from, to RegionID) *LandPassage {
	for i := range passages {
		passage := &passages[i]
		if (passage.From == from && passage.To == to) ||
			(passage.From == to && passage.To == from) {
			return passage
		}
	}
	return nil
}

func undirectedLandPassageKey(from, to RegionID) string {
	if from > to {
		from, to = to, from
	}
	return string(from) + "\x00" + string(to)
}
