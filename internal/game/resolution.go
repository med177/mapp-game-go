package game

import (
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/season"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

// techModsFor bir fraksiyonun araştırdığı teknolojilerden savaş modlarını hesaplar.
func techModsFor(gs *state.GameState, ownerID string) combat.TechMods {
	f, ok := gs.Factions[faction.FactionID(ownerID)]
	if !ok || gs.TechTypes == nil {
		return combat.TechMods{}
	}
	fx := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
	return combat.TechMods{
		AttackMod:  fx.InfantryAttackMod + fx.CavalryAttackMod + fx.SiegeAttackMod,
		DefenseMod: fx.LandDefenseMod,
	}
}

// checkRegionUnlocks kilidi kalkan bölgeleri açar.
// UnlockTurn atanmış bölgeler sadece zamanla açılır; UnlockTurn=0 ise keşif tipi
// kilit sayılır ve komşuya ulaşan ordu ile açılabilir.
func checkRegionUnlocks(gs *state.GameState) []world.RegionID {
	unlocked := gs.SyncTimedRegionUnlocks()
	for _, r := range gs.Regions {
		if !r.IsLocked || r.UnlockTurn > 0 {
			continue
		}
		// Komşuya ulaşan ordu kilidi açar
		for _, a := range gs.Armies {
			src, ok := gs.Regions[a.RegionID]
			if !ok {
				continue
			}
			for _, nid := range src.Neighbors {
				if nid == r.ID {
					r.IsLocked = false
					unlocked = append(unlocked, r.ID)
					break
				}
			}
			if !r.IsLocked {
				break
			}
		}
	}
	return unlocked
}

// applyTechTicks tüm fraksiyonların aktif araştırmalarını bir tur ilerletir.
// Tamamlanan teknolojileri (fraksiyonID, techID) çiftleri olarak döner.
func applyTechTicks(gs *state.GameState) []struct {
	factionID string
	techID    string
} {
	var completed []struct {
		factionID string
		techID    string
	}
	for fid, f := range gs.Factions {
		if f.IsEliminated {
			continue
		}
		if completedID := tech.Tick(&f.Research); completedID != "" {
			completed = append(completed, struct {
				factionID string
				techID    string
			}{string(fid), completedID})
		}
	}
	return completed
}

// applySeasonEffects mevsim etkilerini tüm ordulara uygular.
func applySeasonEffects(gs *state.GameState) {
	s := gs.CurrentSeason()

	if s.IsWinter() {
		for _, a := range gs.Armies {
			a.ApplyWinterAttrition()
		}
	}

	movMod := s.MovementMod()
	for _, a := range gs.Armies {
		mp := 2 * movMod / 100
		if mp < 1 {
			mp = 1
		}
		// Kartografya tech harekete +1 ekler
		if f, ok := gs.Factions[faction.FactionID(a.OwnerID)]; ok && gs.TechTypes != nil {
			fx := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
			mp += fx.MoveBonus
		}
		// Difficulty 3: AI fraksiyonlar +1 hareket puanı bonusu alır
		if gs.Difficulty >= 3 && a.OwnerID != string(gs.PlayerFactionID) {
			mp++
		}
		a.MaxMovePoints = mp
		a.ResetMovePoints()
	}
}

// applyEconomyTick tur başında her fraksiyonun ekonomisini günceller.
// Artık ticaret rotalarını işletir, mal transferi yapar ve piyasa fiyatlarını günceller.
func applyEconomyTick(gs *state.GameState) {
	s := gs.CurrentSeason()
	harvestMod := s.HarvestMod()

	incomeByFaction := make(map[string]int)
	grainByFaction := make(map[string]int)
	ironByFaction := make(map[string]int)
	timberByFaction := make(map[string]int)
	spiceByFaction := make(map[string]int)
	clothByFaction := make(map[string]int)

	for _, r := range gs.Regions {
		if r.IsSea || r.OwnerID == "" {
			continue
		}

		// Bina çarpanları
		goldMod := 1.0
		grainMod := 1.0
		tradeCapMod := 1.0
		satBonus := 0
		for _, bid := range r.Buildings {
			if b, ok := gs.BuildingTypes[bid]; ok {
				goldMod *= b.GoldMod
				grainMod *= b.GrainMod
				tradeCapMod *= b.TradeCapacityMod
				satBonus += b.SatBonus
			}
		}

		income := int(float64(r.GoldIncome()) * goldMod * float64(harvestMod) / 100)
		grain := int(float64(r.BaseGrainOutput) * grainMod)

		// Pasif ticaret geliri (TradeCapacity bazlı)
		// TradeCapacityMod: pazar ve liman gibi binalar ticaret kapasitesini artırır
		tradeIncome := economy.RegionTradeIncome(r.TradeCapacity, tradeCapMod)
		// Mevsimsel ticaret modu uygula
		tradeIncome = tradeIncome * s.TradeMod() / 100

		incomeByFaction[r.OwnerID] += income + tradeIncome
		grainByFaction[r.OwnerID] += grain
		ironByFaction[r.OwnerID] += r.BaseIronOutput
		timberByFaction[r.OwnerID] += r.BaseTimberOutput
		spiceByFaction[r.OwnerID] += r.BaseSpiceOutput
		clothByFaction[r.OwnerID] += r.BaseClothOutput

		// Vergi memnuniyet etkisi + bina bonusu
		delta := economy.TaxSatisfactionDelta(r.TaxRate) + satBonus
		r.Satisfaction = clamp(r.Satisfaction+delta, 0, 100)
	}

	// --- Ticaret rotalarını işlet (mal + altın transferi) ---
	tradeLogs := economy.ApplyTradeRoutes(gs.Factions, gs.TradeRoutes)
	for _, log := range tradeLogs {
		// Ticaret logları oyuncuya aitse göster
		if gs.PlayerFactionID != "" {
			_ = log // ileride oyuncuya bildirim gösterilebilir
		}
	}

	// Gerçek ordu bakım maliyetleri (UnitType.GrainUpkeep)
	upkeepByFaction := make(map[string]int)
	for _, a := range gs.Armies {
		for _, u := range a.Units {
			if t, ok := gs.UnitTypes[u.TypeID]; ok {
				upkeepByFaction[a.OwnerID] += t.GrainUpkeep
			}
		}
	}

	for fid, f := range gs.Factions {
		fidStr := string(fid)

		// Teknoloji bonusları
		var fx tech.Effects
		if gs.TechTypes != nil {
			fx = tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
		}

		// GoldPerRegion tech bonusu
		ownedCount := len(gs.RegionsOwnedBy(fid))
		techGold := fx.GoldPerRegion * ownedCount

		f.Gold += incomeByFaction[fidStr] + techGold
		netGrain := int(float64(grainByFaction[fidStr]) * (1.0 + fx.GrainMod))
		f.Grain += netGrain - upkeepByFaction[fidStr]
		f.Iron += ironByFaction[fidStr]
		f.Timber += timberByFaction[fidStr]
		f.Spice += spiceByFaction[fidStr]
		f.Cloth += clothByFaction[fidStr]

		// Memnuniyet tech bonusu tüm bölgelere
		if fx.SatisfactionBonus > 0 {
			for _, r := range gs.Regions {
				if r.OwnerID == fidStr {
					r.Satisfaction = clamp(r.Satisfaction+fx.SatisfactionBonus, 0, 100)
				}
			}
		}

		if f.Gold < 0 {
			f.Gold = 0
		}
		if f.Grain < 0 {
			f.Grain = 0
		}
	}

	// --- Dinamik piyasa fiyatlarını güncelle ---
	gs.MarketPrices = economy.ComputeMarketPrices(gs.Factions)
}

// checkRebellions isyan riski olan bölgeleri kontrol eder.
func checkRebellions(gs *state.GameState) {
	for _, r := range gs.Regions {
		if !r.IsRebellionRisk() {
			continue
		}
		hasGarrison := false
		for _, a := range gs.Armies {
			if a.RegionID == r.ID && !a.IsNaval {
				hasGarrison = true
				break
			}
		}
		// Surlar isyanı bastırır
		for _, bid := range r.Buildings {
			if bid == "walls" {
				hasGarrison = true
				break
			}
		}
		if !hasGarrison {
			r.OwnerID = ""
			r.Satisfaction = 50
		}
	}
}

// checkEliminations kara toprağı kalmayan fraksiyonları elendi olarak işaretler ve ordularını kaldırır.
func checkEliminations(gs *state.GameState) {
	for fid, f := range gs.Factions {
		if f.IsEliminated {
			continue
		}
		if len(gs.LandRegionsOwnedBy(fid)) == 0 {
			f.IsEliminated = true
			// Tüm ordularını haritadan kaldır
			for aid, a := range gs.Armies {
				if a.OwnerID == string(fid) {
					delete(gs.Armies, aid)
				}
			}
			// Elenen fraksiyonun tüm diplomasi kayıtlarını kaldır.
			for key, rel := range gs.Relations {
				if rel == nil || rel.FactionA == fid || rel.FactionB == fid {
					delete(gs.Relations, key)
				}
			}
		}
	}
}

// applyRelationDecay savaş halindeki ilişkileri kötüleştirir, barış halindekini iyileştirir.
func applyRelationDecay(gs *state.GameState) {
	diplomacy.ApplyRelationDecay(gs)
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// applyReligionConversion her tur sahip olunan bölgelerin din dönüşümünü ilerletir.
// Sahibin dini bölgenin dininden farklıysa ConversionTurns artar.
// 24 turda din değişir ve memnuniyet cezası uygulanır.
func applyReligionConversion(gs *state.GameState) {
	const conversionThreshold = 24

	for _, r := range gs.Regions {
		if r.IsSea || r.OwnerID == "" {
			r.ConversionTurns = 0
			continue
		}
		ownerRel := ownerReligionStr(gs, r.OwnerID)
		if ownerRel == "" || ownerRel == r.Religion {
			r.ConversionTurns = 0
			continue
		}

		r.ConversionTurns++
		if r.ConversionTurns >= conversionThreshold {
			r.Religion = ownerRel
			r.ConversionTurns = 0
			// Din değişimi halk memnuniyetini düşürür
			r.Satisfaction -= 20
			if r.Satisfaction < 0 {
				r.Satisfaction = 0
			}
		}
	}
}

// ownerReligionStr bir fraksiyonun dinini döner; bulunamazsa "".
func ownerReligionStr(gs *state.GameState, ownerID string) string {
	for fid, f := range gs.Factions {
		if string(fid) == ownerID {
			return string(f.Religion)
		}
	}
	return ""
}

var _ = season.SeasonWinter
