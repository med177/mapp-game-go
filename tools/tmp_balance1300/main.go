package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/victory"
	"mapp-game-go/internal/world"
)

type regionJSON struct {
	ID string `json:"id"`
	BaseGoldIncome int `json:"base_gold_income"`
}

func main() {
	base := "assets/scenarios/1300_ottoman_rise/data/"
	regions, order, err := world.LoadRegionsWithOrder(base + "regions.json"); must(err)
	if err := world.LoadRegionSettlements(base+"settlements.json", regions); err != nil { panic(err) }
	factions, factionOrder, err := faction.LoadFactionsWithOrder(base + "factions.json"); must(err)
	units, err := army.LoadUnitTypes(base + "units.json"); must(err)
	armies, err := army.LoadArmies(base+"armies.json", units); must(err)
	buildings, err := city.LoadBuildings(base + "buildings.json"); must(err)
	techs, err := tech.LoadTechnologies(base + "technologies.json"); must(err)
	centers, err := world.LoadTradeCenters(base+"trade_centers.json", regions); must(err)
	relations, err := faction.LoadRelations(base+"relations.json", factions); must(err)
	gs := &state.GameState{Turn: 1, Year: 1300, Month: 3, Regions: regions, RegionOrder: order, Factions: factions, FactionOrder: factionOrder, Armies: armies, UnitTypes: units, BuildingTypes: buildings, TechTypes: techs, TradeCenters: centers, Relations: relations}
	gs.ApplyHistoricalFactionChanges()
	gs.NormalizeFactionCapitals()
	diplomacy.NormalizeVassalage(gs)
	diplomacy.EnsureTradeRoutesForActiveRelations(gs)

	original := make(map[string]int, len(regions))
	for id, r := range regions { original[string(id)] = r.BaseGoldIncome }
	for _, fid := range factionOrder {
		before := victory.GoldEconomyPreview(gs, fid).NetChange
		if before >= 50 { continue }
		lo, hi := 1.0, 2.0
		for { applyFactionScale(gs, fid, original, hi); if victory.GoldEconomyPreview(gs, fid).NetChange >= 50 { break }; hi *= 2 }
		for i := 0; i < 40; i++ { mid := (lo + hi) / 2; applyFactionScale(gs, fid, original, mid); if victory.GoldEconomyPreview(gs, fid).NetChange >= 50 { hi = mid } else { lo = mid } }
		applyFactionScale(gs, fid, original, hi)
		after := victory.GoldEconomyPreview(gs, fid).NetChange
		fmt.Printf("# %s %d -> %d scale=%.4f\n", fid, before, after, hi)
	}

	// Emit an apply_patch patch containing only changed base_gold_income lines.
	data, err := osRead(base + "regions.json"); must(err)
	var raw []map[string]any
	must(json.Unmarshal(data, &raw))
	for _, obj := range raw {
		id, _ := obj["id"].(string)
		if r := regions[world.RegionID(id)]; r != nil && r.BaseGoldIncome != original[id] {
			old := fmt.Sprintf(`"base_gold_income": %d`, original[id])
			new := fmt.Sprintf(`"base_gold_income": %d`, r.BaseGoldIncome)
			fmt.Printf("@@ %s @@\n-    %s\n+    %s\n", id, old, new)
		}
	}
}

func applyFactionScale(gs *state.GameState, fid faction.FactionID, original map[string]int, scale float64) {
	for _, r := range gs.Regions {
		if r != nil && r.OwnerID == string(fid) { r.BaseGoldIncome = int(math.Ceil(float64(original[string(r.ID)]) * scale)) }
	}
}

func must(err error) { if err != nil { panic(err) } }
func osRead(path string) ([]byte, error) { return os.ReadFile(path) }
