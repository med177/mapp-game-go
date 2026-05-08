package city

import (
	"encoding/json"
	"fmt"
	"os"
)

// Building bir bina tipini tanımlar (JSON'dan yüklenir).
type Building struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	NameTR          string  `json:"name_tr"`
	GoldCost        int     `json:"gold_cost"`
	TurnsRequired   int     `json:"turns_required"`
	GoldMod         float64 `json:"gold_mod"`         // altın gelir çarpanı (1.0 = değişmez)
	GrainMod        float64 `json:"grain_mod"`        // tahıl üretim çarpanı
	SatBonus        int     `json:"sat_bonus"`        // tur başına memnuniyet bonusu
	DefBonus        int     `json:"def_bonus"`        // savunma bonusu
	MaxPerRegion    int     `json:"max_per_region"`   // bölgede max adet (genelde 1)
	RequiredTerrain string  `json:"required_terrain"` // "" = her arazi
}

// LoadBuildings bina tiplerini JSON'dan yükler.
func LoadBuildings(path string) (map[string]*Building, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("binalar okunamadı: %w", err)
	}
	var list []*Building
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("binalar parse edilemedi: %w", err)
	}
	m := make(map[string]*Building, len(list))
	for _, b := range list {
		if b.TurnsRequired <= 0 {
			b.TurnsRequired = 2
		}
		m[b.ID] = b
	}
	return m, nil
}
