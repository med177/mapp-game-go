package world

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"mapp-game-go/internal/economy"
)

type TradeCenterTier string

const (
	TradeCenterPrimary   TradeCenterTier = "primary"
	TradeCenterSecondary TradeCenterTier = "secondary"

	defaultPrimaryTradeCapacityBonus      = 2
	defaultSecondaryTradeCapacityBonus    = 1
	defaultPrimaryTradeIncomeBonus        = 4
	defaultSecondaryTradeIncomeBonus      = 2
	defaultPrimaryMerchantCapacityBonus   = 2
	defaultSecondaryMerchantCapacityBonus = 1
	defaultPrimaryMerchantIncomeBonus     = 2
	defaultSecondaryMerchantIncomeBonus   = 1
)

type TradeCenterDef struct {
	ID                    RegionID                 `json:"region_id"`
	NameTR                string                   `json:"name_tr,omitempty"`
	Tier                  TradeCenterTier          `json:"tier,omitempty"`
	TradeCapacityBonus    int                      `json:"trade_capacity_bonus,omitempty"`
	TradeIncomeBonus      int                      `json:"trade_income_bonus,omitempty"`
	MerchantCapacityBonus int                      `json:"merchant_capacity_bonus,omitempty"`
	MerchantIncomeBonus   int                      `json:"merchant_income_bonus,omitempty"`
	Links                 []RegionID               `json:"links,omitempty"`
	WorldX                int                      `json:"world_x,omitempty"`
	WorldY                int                      `json:"world_y,omitempty"`
	OffMap                bool                     `json:"off_map,omitempty"`
	UnlockYear            int                      `json:"unlock_year,omitempty"`
	CompetitionImpacts    []TradeCompetitionImpact `json:"competition_impacts,omitempty"`
}

// TradeCompetitionImpact, yeni bir merkezin açılmasıyla eski bir merkezin
// tarihsel akışında oluşan gelir/hacim değişimini taşır. Etkinlik tarihi
// impact üzerinde tekrar edilmez; sahibi olan TradeCenterDef.UnlockYear'dan
// gelir.
type TradeCompetitionImpact struct {
	CenterID      RegionID `json:"center_id"`
	IncomePercent int      `json:"income_percent,omitempty"`
	AmountPercent int      `json:"amount_percent,omitempty"`
}

// HistoricalTradeFlow, devletler arası diplomatik rotalardan bağımsız olarak
// tarihin belirli bir döneminde merkezler arasında akan ticari hattı taşır.
// Off-map merkezler tarihsel akışın bir ucu olabilir; gelir yalnızca gerçek
// bölgesi bulunan endpoint sahiplerine yazılır.
type HistoricalTradeFlow struct {
	FromRegionID      RegionID         `json:"from"`
	ToRegionID        RegionID         `json:"to"`
	Good              economy.GoodType `json:"good"`
	AmountPerTurn     int              `json:"amount_per_turn"`
	GoldIncomePerTurn int              `json:"gold_income_per_turn"`
	StartYear         int              `json:"start_year,omitempty"`
	EndYear           int              `json:"end_year,omitempty"`
}

type TradeCenterConfig struct {
	PrimaryTradeCapacityBonus      int                   `json:"primary_trade_capacity_bonus,omitempty"`
	SecondaryTradeCapacityBonus    int                   `json:"secondary_trade_capacity_bonus,omitempty"`
	PrimaryTradeIncomeBonus        int                   `json:"primary_trade_income_bonus,omitempty"`
	SecondaryTradeIncomeBonus      int                   `json:"secondary_trade_income_bonus,omitempty"`
	PrimaryMerchantCapacityBonus   int                   `json:"primary_merchant_capacity_bonus,omitempty"`
	SecondaryMerchantCapacityBonus int                   `json:"secondary_merchant_capacity_bonus,omitempty"`
	PrimaryMerchantIncomeBonus     int                   `json:"primary_merchant_income_bonus,omitempty"`
	SecondaryMerchantIncomeBonus   int                   `json:"secondary_merchant_income_bonus,omitempty"`
	Centers                        []TradeCenterDef      `json:"centers"`
	HistoricalFlows                []HistoricalTradeFlow `json:"historical_flows,omitempty"`
}

// ApplyDefaultBonuses eski senaryo verilerinde henüz bulunmayan merkez bonus
// alanlarını oyun dengesiyle uyumlu varsayılanlara taşır. Sıfır, mevcut modelde
// "tier varsayılanını kullan" anlamına geldiği için bu migration yalnız
// tanımlanmamış alanları tamamlar; per-center override'lar korunur.
func (c TradeCenterConfig) ApplyDefaultBonuses() TradeCenterConfig {
	if c.PrimaryTradeCapacityBonus == 0 {
		c.PrimaryTradeCapacityBonus = defaultPrimaryTradeCapacityBonus
	}
	if c.SecondaryTradeCapacityBonus == 0 {
		c.SecondaryTradeCapacityBonus = defaultSecondaryTradeCapacityBonus
	}
	if c.PrimaryTradeIncomeBonus == 0 {
		c.PrimaryTradeIncomeBonus = defaultPrimaryTradeIncomeBonus
	}
	if c.SecondaryTradeIncomeBonus == 0 {
		c.SecondaryTradeIncomeBonus = defaultSecondaryTradeIncomeBonus
	}
	if c.PrimaryMerchantCapacityBonus == 0 {
		c.PrimaryMerchantCapacityBonus = defaultPrimaryMerchantCapacityBonus
	}
	if c.SecondaryMerchantCapacityBonus == 0 {
		c.SecondaryMerchantCapacityBonus = defaultSecondaryMerchantCapacityBonus
	}
	if c.PrimaryMerchantIncomeBonus == 0 {
		c.PrimaryMerchantIncomeBonus = defaultPrimaryMerchantIncomeBonus
	}
	if c.SecondaryMerchantIncomeBonus == 0 {
		c.SecondaryMerchantIncomeBonus = defaultSecondaryMerchantIncomeBonus
	}
	return c
}

func (c TradeCenterDef) ActiveInYear(year int) bool {
	return c.UnlockYear <= 0 || year >= c.UnlockYear
}

func (f HistoricalTradeFlow) ActiveInYear(year int) bool {
	return (f.StartYear <= 0 || year >= f.StartYear) && (f.EndYear <= 0 || year <= f.EndYear)
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
		if ok {
			if region.IsSea || region.TradeCapacity <= 0 {
				continue
			}
		} else {
			if !c.OffMap || c.NameTR == "" {
				continue
			}
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
		impacts := filtered[i].CompetitionImpacts[:0]
		impactSeen := make(map[RegionID]bool, len(filtered[i].CompetitionImpacts))
		for _, impact := range filtered[i].CompetitionImpacts {
			if !validCenter[impact.CenterID] || impact.CenterID == filtered[i].ID || impactSeen[impact.CenterID] {
				continue
			}
			impactSeen[impact.CenterID] = true
			impacts = append(impacts, impact)
		}
		filtered[i].CompetitionImpacts = impacts
	}

	validFlows := make([]HistoricalTradeFlow, 0, len(payload.HistoricalFlows))
	validGoods := map[economy.GoodType]bool{
		economy.GoodGold: true, economy.GoodGrain: true, economy.GoodIron: true,
		economy.GoodTimber: true, economy.GoodStone: true, economy.GoodSpice: true,
		economy.GoodCloth: true,
	}
	for _, flow := range payload.HistoricalFlows {
		if !validCenter[flow.FromRegionID] || !validCenter[flow.ToRegionID] ||
			flow.FromRegionID == flow.ToRegionID || !validGoods[flow.Good] ||
			flow.AmountPerTurn <= 0 || flow.GoldIncomePerTurn < 0 {
			continue
		}
		fromDef := filtered[indexOfTradeCenter(filtered, flow.FromRegionID)]
		toDef := filtered[indexOfTradeCenter(filtered, flow.ToRegionID)]
		if !containsRegionID(fromDef.Links, flow.ToRegionID) && !containsRegionID(toDef.Links, flow.FromRegionID) {
			continue
		}
		validFlows = append(validFlows, flow)
	}

	out = payload.ApplyDefaultBonuses()
	out.Centers = filtered
	out.HistoricalFlows = validFlows
	return out, nil
}

func containsRegionID(ids []RegionID, target RegionID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func indexOfTradeCenter(centers []TradeCenterDef, id RegionID) int {
	for i := range centers {
		if centers[i].ID == id {
			return i
		}
	}
	return -1
}
