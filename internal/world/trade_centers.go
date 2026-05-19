package world

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type TradeCenterTier string

const (
	TradeCenterPrimary   TradeCenterTier = "primary"
	TradeCenterSecondary TradeCenterTier = "secondary"
)

type TradeCenterDef struct {
	ID    RegionID         `json:"id"`
	Tier  TradeCenterTier  `json:"tier,omitempty"`
	Links []RegionID       `json:"links,omitempty"`
}

type TradeCenterConfig struct {
	Centers []TradeCenterDef `json:"centers"`
}

// LoadTradeCenters tarihsel ticaret merkezlerini scenario data dosyasından okur.
// Dosya yoksa boş config ve nil hata döner (opsiyonel veri).
func LoadTradeCenters(path string, regions map[RegionID]*Region) (TradeCenterConfig, error) {
	var out TradeCenterConfig
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return out, fmt.Errorf("trade_centers dosyası okunamadı: %w", err)
	}
	var payload TradeCenterConfig
	if err := json.Unmarshal(data, &payload); err != nil {
		return out, fmt.Errorf("trade_centers JSON parse hatası: %w", err)
	}
	if len(payload.Centers) == 0 {
		return out, nil
	}

	seen := make(map[RegionID]bool, len(payload.Centers))
	validCenter := make(map[RegionID]bool, len(payload.Centers))
	filtered := make([]TradeCenterDef, 0, len(payload.Centers))
	for _, c := range payload.Centers {
		if c.ID == "" || seen[c.ID] {
			continue
		}
		region, ok := regions[c.ID]
		if !ok || region.IsSea || region.TradeCapacity <= 0 {
			continue
		}
		if c.Tier != TradeCenterPrimary && c.Tier != TradeCenterSecondary {
			c.Tier = TradeCenterSecondary
		}
		seen[c.ID] = true
		validCenter[c.ID] = true
		filtered = append(filtered, c)
	}

	if len(filtered) == 0 {
		return out, nil
	}

	// Link temizliği: sadece geçerli center ID'leri tut.
	for i := range filtered {
		if len(filtered[i].Links) == 0 {
			continue
		}
		linkSeen := make(map[RegionID]bool, len(filtered[i].Links))
		links := make([]RegionID, 0, len(filtered[i].Links))
		for _, lid := range filtered[i].Links {
			if lid == "" || lid == filtered[i].ID || linkSeen[lid] || !validCenter[lid] {
				continue
			}
			linkSeen[lid] = true
			links = append(links, lid)
		}
		sort.Slice(links, func(a, b int) bool { return links[a] < links[b] })
		filtered[i].Links = links
	}

	out.Centers = filtered
	return out, nil
}
