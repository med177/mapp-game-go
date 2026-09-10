package city

import (
	"encoding/json"
	"fmt"
	"os"
)

// Building bir bina tipini tanımlar (JSON'dan yüklenir).
type Building struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	NameTR           string  `json:"name_tr"`
	GoldCost         int     `json:"gold_cost"`
	GoldMaintenance  int     `json:"gold_maintenance"` // seviye başına turda altın bakımı
	GrainCost        int     `json:"grain_cost"`
	IronCost         int     `json:"iron_cost"`
	TimberCost       int     `json:"timber_cost"`
	StoneCost        int     `json:"stone_cost"`
	SpiceCost        int     `json:"spice_cost"`
	ClothCost        int     `json:"cloth_cost"`
	TurnsRequired    int     `json:"turns_required"`
	GoldMod          float64 `json:"gold_mod"`           // altın gelir çarpanı (1.0 = değişmez)
	GrainMod         float64 `json:"grain_mod"`          // tahıl üretim çarpanı
	GrainBonus       int     `json:"grain_bonus"`        // bina seviyesi başına sabit tahıl katkısı
	IronMod          float64 `json:"iron_mod"`           // demir üretim çarpanı (1.0 = değişmez)
	IronBonus        int     `json:"iron_bonus"`         // bina seviyesi başına sabit demir katkısı
	TimberMod        float64 `json:"timber_mod"`         // kereste üretim çarpanı (1.0 = değişmez)
	TimberBonus      int     `json:"timber_bonus"`       // bina seviyesi başına sabit kereste katkısı
	StoneMod         float64 `json:"stone_mod"`          // taş üretim çarpanı (1.0 = değişmez)
	StoneBonus       int     `json:"stone_bonus"`        // bina seviyesi başına sabit taş katkısı
	SpiceMod         float64 `json:"spice_mod"`          // baharat üretim çarpanı (1.0 = değişmez)
	SpiceBonus       int     `json:"spice_bonus"`        // bina seviyesi başına sabit baharat katkısı
	ClothMod         float64 `json:"cloth_mod"`          // kumaş üretim çarpanı (1.0 = değişmez)
	ClothBonus       int     `json:"cloth_bonus"`        // bina seviyesi başına sabit kumaş katkısı
	TradeCapacityMod float64 `json:"trade_capacity_mod"` // ticaret kapasitesi çarpanı (1.0 = değişmez)
	SatBonus         int     `json:"sat_bonus"`          // tur başına memnuniyet bonusu
	DefBonus         int     `json:"def_bonus"`          // savunma bonusu
	StorageCapacity  int     `json:"storage_capacity"`   // tahıl depolama kapasitesi bonusu
	MaxPerRegion     int     `json:"max_per_region"`     // bölgede max adet (genelde 1)
	RequiredTerrain  string  `json:"required_terrain"`   // "" = her arazi
}

// LoadBuildings bina tiplerini JSON'dan yükler.
func LoadBuildings(path string) (map[string]*Building, error) {
	buildings, _, err := LoadBuildingsWithOrder(path)
	return buildings, err
}

// LoadBuildingsWithOrder bina tiplerini JSON'dan yükler ve dosyadaki bina ID
// sırasını ayrıca döner. Map bina tanımlarını hızlı erişim için, slice ise
// JSON'daki gösterim sırasını korumak için kullanılır.
func LoadBuildingsWithOrder(path string) (map[string]*Building, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("binalar okunamadı: %w", err)
	}
	var list []*Building
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, nil, fmt.Errorf("binalar parse edilemedi: %w", err)
	}
	m := make(map[string]*Building, len(list))
	order := make([]string, 0, len(list))
	for _, b := range list {
		if b == nil {
			continue
		}
		if b.TurnsRequired <= 0 {
			b.TurnsRequired = 2
		}
		if b.TradeCapacityMod <= 0 {
			b.TradeCapacityMod = 1.0
		}
		if b.GrainMod <= 0 {
			b.GrainMod = 1.0
		}
		if b.IronMod <= 0 {
			b.IronMod = 1.0
		}
		if b.TimberMod <= 0 {
			b.TimberMod = 1.0
		}
		if b.StoneMod <= 0 {
			b.StoneMod = 1.0
		}
		if b.SpiceMod <= 0 {
			b.SpiceMod = 1.0
		}
		if b.ClothMod <= 0 {
			b.ClothMod = 1.0
		}
		m[b.ID] = b
		order = append(order, b.ID)
	}
	return m, order, nil
}

// GetGoldMod Building.GoldMod değerini döner (interface uyumluluğu için).
func (b *Building) GetGoldMod() float64 {
	return b.GoldMod
}
