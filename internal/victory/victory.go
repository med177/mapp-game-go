package victory

import (
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	defaultDominationTarget = 15
	defaultGoldThreshold    = 5000
	defaultGoldHoldTurns    = 3
	defaultArmyStrength     = 200
	defaultDefeated         = 3
	aiDominationLimit       = 20
)

// Check her tur sonunda zafer/mağlubiyet koşullarını kontrol eder.
func Check(gs *state.GameState) {
	if gs.Phase == state.PhaseGameOver {
		return
	}

	playerRegions := gs.RegionsOwnedBy(gs.PlayerFactionID)

	// Oyuncu elendi mi?
	if len(playerRegions) == 0 && gs.Turn > 1 {
		gs.Phase = state.PhaseGameOver
		gs.WinnerID = ""
		return
	}

	// Seçilen zafer tipine göre kontrol
	switch gs.Victory.Type {
	case state.VictoryDomination, "":
		checkDomination(gs, playerRegions)
	case state.VictoryEconomic:
		checkEconomic(gs)
	case state.VictoryMilitary:
		checkMilitary(gs)
	case state.VictoryReligious:
		checkReligious(gs, playerRegions)
	}

	if gs.Phase == state.PhaseGameOver {
		return
	}

	// Herhangi bir AI fraksiyonu çok büyürse oyuncu kaybeder
	for fid := range gs.Factions {
		if fid == gs.PlayerFactionID {
			continue
		}
		if len(gs.RegionsOwnedBy(fid)) >= aiDominationLimit {
			gs.Phase = state.PhaseGameOver
			gs.WinnerID = fid
			return
		}
	}
}

// checkDomination bölge sayısına ve kritik bölgelere göre zafer kontrol eder.
func checkDomination(gs *state.GameState, playerRegions []*world.Region) {
	target := gs.Victory.TargetRegionCount
	if target == 0 {
		target = defaultDominationTarget
	}
	if len(playerRegions) < target {
		return
	}
	// Zorunlu bölgeler var mı?
	for _, rid := range gs.Victory.RequiredRegions {
		region, ok := gs.Regions[rid]
		if !ok || region.OwnerID != string(gs.PlayerFactionID) {
			return
		}
	}
	gs.Phase = state.PhaseGameOver
	gs.WinnerID = gs.PlayerFactionID
}

// checkEconomic altın miktarını belirli tur süre boyunca koruma zaferini kontrol eder.
func checkEconomic(gs *state.GameState) {
	threshold := gs.Victory.TargetGoldIncome
	if threshold == 0 {
		threshold = defaultGoldThreshold
	}
	holdTurns := gs.Victory.GoldHoldTurns
	if holdTurns == 0 {
		holdTurns = defaultGoldHoldTurns
	}

	f, ok := gs.Factions[gs.PlayerFactionID]
	if !ok {
		return
	}
	if f.Gold >= threshold {
		gs.EconomicVictoryTurns++
		if gs.EconomicVictoryTurns >= holdTurns {
			gs.Phase = state.PhaseGameOver
			gs.WinnerID = gs.PlayerFactionID
		}
	} else {
		gs.EconomicVictoryTurns = 0
	}
}

// checkMilitary ordu gücü ve fraksiyon yenilgisi sayısına göre zafer kontrol eder.
func checkMilitary(gs *state.GameState) {
	// Elenmiş fraksiyon sayısını güncelle
	eliminated := 0
	for fid, f := range gs.Factions {
		if fid != gs.PlayerFactionID && f.IsEliminated {
			eliminated++
		}
	}
	gs.FactionsEliminated = eliminated

	targetStr := gs.Victory.TargetArmyStrength
	if targetStr == 0 {
		targetStr = defaultArmyStrength
	}
	targetDefeated := gs.Victory.TargetDefeated
	if targetDefeated == 0 {
		targetDefeated = defaultDefeated
	}

	// Oyuncunun toplam ordu gücünü hesapla
	totalStr := 0
	for _, a := range gs.Armies {
		if a.OwnerID == string(gs.PlayerFactionID) {
			totalStr += a.TotalStrength(gs.UnitTypes)
		}
	}

	if totalStr >= targetStr && eliminated >= targetDefeated {
		gs.Phase = state.PhaseGameOver
		gs.WinnerID = gs.PlayerFactionID
	}
}

// checkReligious kutsal şehirlerin oyuncu tarafından belirli tur süre tutulması zaferini kontrol eder.
func checkReligious(gs *state.GameState, _ []*world.Region) {
	if len(gs.Victory.RequiredRegions) == 0 {
		return
	}

	allHeld := true
	for _, rid := range gs.Victory.RequiredRegions {
		region, ok := gs.Regions[rid]
		if !ok || region.OwnerID != string(gs.PlayerFactionID) {
			allHeld = false
			break
		}
	}

	if allHeld {
		gs.ReligiousVictoryTurns++
		// 12 tur (~1 yıl) kutsal şehirleri tutmak = zafer
		if gs.ReligiousVictoryTurns >= 12 {
			gs.Phase = state.PhaseGameOver
			gs.WinnerID = gs.PlayerFactionID
		}
	} else {
		gs.ReligiousVictoryTurns = 0
	}
}
