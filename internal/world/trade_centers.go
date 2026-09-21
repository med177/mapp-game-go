package world

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"mapp-game-go/internal/economy"
)

type TradeCenterTier string

type TradeRouteType string

// TradeCenterLink, haritada çizilecek görsel bağlantının hedefini ve fiziksel
// yol türünü taşır. Normal merkezlerde tek link ekonomik olarak iki yönlüdür;
// sources altındaki off-map bağlantılar yalnızca listelenen yönde akar.
type TradeCenterLink struct {
	RegionID RegionID       `json:"region_id"`
	Type     TradeRouteType `json:"type,omitempty"`
}

const (
	TradeCenterPrimary   TradeCenterTier = "primary"
	TradeCenterSecondary TradeCenterTier = "secondary"
	TradeRouteLand       TradeRouteType  = "land"
	TradeRouteSea        TradeRouteType  = "sea"

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
	Links                 []TradeCenterLink        `json:"links,omitempty"`
	WorldX                int                      `json:"world_x,omitempty"`
	WorldY                int                      `json:"world_y,omitempty"`
	OffMap                bool                     `json:"off_map,omitempty"`
	UnlockYear            int                      `json:"unlock_year,omitempty"`
	MainRoute             bool                     `json:"main_route,omitempty"`
	CompetitionImpacts    []TradeCompetitionImpact `json:"competition_impacts,omitempty"`
	SourceGoods           []HistoricalTradeGood    `json:"source_goods,omitempty"`
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

// HistoricalTradeGood, bir merkezden ticaret ağına yayılan tek malın kaynak
// tanımıdır. Bağlantıların her biri için tekrar yazılmaz; runtime aynı malı
// merkez grafiğinde bağlı düğümlere taşır.
type HistoricalTradeGood struct {
	Good              economy.GoodType `json:"good"`
	AmountPerTurn     int              `json:"amount_per_turn"`
	GoldIncomePerTurn int              `json:"gold_income_per_turn,omitempty"`
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
	SourceCenterID    RegionID         `json:"-"`
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
	Sources                        []TradeCenterDef      `json:"sources,omitempty"`
	Centers                        []TradeCenterDef      `json:"centers"`
	HistoricalFlows                []HistoricalTradeFlow `json:"historical_flows,omitempty"`
}

// TradeAdjacency returns the economic connectivity of the center graph.
// Normal center links are bidirectional for trade even when only one visual
// link is listed; off-map source links remain one-way from the source.
func (c TradeCenterConfig) TradeAdjacency() map[RegionID][]RegionID {
	centersByID := make(map[RegionID]TradeCenterDef, len(c.Centers))
	for _, center := range c.Centers {
		centersByID[center.ID] = center
	}

	adjacency := make(map[RegionID][]RegionID, len(centersByID))
	for _, center := range c.Centers {
		centerIsSource := center.OffMap || len(center.SourceGoods) > 0
		for _, link := range center.Links {
			linked, ok := centersByID[link.RegionID]
			if !ok || linked.ID == center.ID {
				continue
			}
			adjacency[center.ID] = appendUniqueTradeCenterRegionID(adjacency[center.ID], linked.ID)
			linkedIsSource := linked.OffMap || len(linked.SourceGoods) > 0
			if !centerIsSource && !linkedIsSource {
				adjacency[linked.ID] = appendUniqueTradeCenterRegionID(adjacency[linked.ID], center.ID)
			}
		}
	}
	for id := range adjacency {
		sort.Slice(adjacency[id], func(i, j int) bool { return adjacency[id][i] < adjacency[id][j] })
	}
	return adjacency
}

func appendUniqueTradeCenterRegionID(ids []RegionID, id RegionID) []RegionID {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
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
	allDefs := make([]TradeCenterDef, 0, len(payload.Sources)+len(payload.Centers))
	for i := range payload.Sources {
		payload.Sources[i].OffMap = true
	}
	allDefs = append(allDefs, payload.Sources...)
	allDefs = append(allDefs, payload.Centers...)
	if len(allDefs) == 0 {
		return out, nil
	}

	seen := make(map[RegionID]bool, len(allDefs))
	validCenter := make(map[RegionID]bool, len(allDefs))
	filtered := make([]TradeCenterDef, 0, len(allDefs))
	for _, c := range allDefs {
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
		if c.MainRoute {
			c.Tier = ""
		} else if c.Tier != TradeCenterPrimary && c.Tier != TradeCenterSecondary {
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
		links := make([]TradeCenterLink, 0, len(filtered[i].Links))
		for _, link := range filtered[i].Links {
			lid := link.RegionID
			if lid == "" || lid == filtered[i].ID || linkSeen[lid] || !validCenter[lid] {
				continue
			}
			if link.Type != "" && link.Type != TradeRouteLand && link.Type != TradeRouteSea {
				link.Type = ""
			}
			linkSeen[lid] = true
			links = append(links, link)
		}
		sort.Slice(links, func(a, b int) bool { return links[a].RegionID < links[b].RegionID })
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
		goods := filtered[i].SourceGoods[:0]
		goodSeen := make(map[economy.GoodType]bool, len(filtered[i].SourceGoods))
		for _, good := range filtered[i].SourceGoods {
			if !isHistoricalGood(good.Good) || good.AmountPerTurn <= 0 || goodSeen[good.Good] {
				continue
			}
			goodSeen[good.Good] = true
			goods = append(goods, good)
		}
		filtered[i].SourceGoods = goods
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
		if !containsTradeCenterLink(fromDef.Links, flow.ToRegionID) && !containsTradeCenterLink(toDef.Links, flow.FromRegionID) {
			continue
		}
		validFlows = append(validFlows, flow)
	}

	out = payload.ApplyDefaultBonuses()
	out.Centers = filtered
	out.HistoricalFlows = validFlows
	return out, nil
}

func containsTradeCenterLink(links []TradeCenterLink, target RegionID) bool {
	for _, link := range links {
		if link.RegionID == target {
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

func isHistoricalGood(good economy.GoodType) bool {
	switch good {
	case economy.GoodGold, economy.GoodGrain, economy.GoodIron, economy.GoodTimber,
		economy.GoodStone, economy.GoodSpice, economy.GoodCloth:
		return true
	default:
		return false
	}
}
