package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

// coalitionThreshold oyuncunun bu kadar bölgeyi geçmesi koalisyon tetikler.
const coalitionThreshold = 8

// TakeTurn belirtilen fraksiyon için tüm AI kararlarını verir ve uygular.
func aiTechMods(gs *state.GameState, ownerID string) combat.TechMods {
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

// relationScore iki fraksiyon arasındaki ilişki puanını döner; yoksa 0.
func relationScore(gs *state.GameState, a, b string) (int, faction.DiplomaticStance) {
	key := faction.RelationKey(faction.FactionID(a), faction.FactionID(b))
	if rel, ok := gs.Relations[key]; ok {
		return rel.Score, rel.Stance
	}
	return 0, faction.StancePeace
}

// TakeTurn belirtilen fraksiyon için tüm AI kararlarını verir ve uygular.
func TakeTurn(gs *state.GameState, fid faction.FactionID) {
	// Difficulty 3: koalisyon mantığını çalıştır
	if gs.Difficulty >= 3 {
		FormCoalitionAgainstPlayer(gs, fid)
	}

	// Ordu listesinin anlık kopyasını al — iterasyon sırasında map değişebilir
	var ownArmies []*army.Army
	for _, a := range gs.Armies {
		if a.OwnerID == string(fid) {
			ownArmies = append(ownArmies, a)
		}
	}

	for _, a := range ownArmies {
		// Ordu hâlâ haritada mı?
		if _, alive := gs.Armies[a.ID]; !alive {
			continue
		}
		moveArmy(gs, a)
	}
}

// FormCoalitionAgainstPlayer oyuncu tehdit eşiğini geçmişse diğer AI fraksiyonlarla ittifak kurar.
func FormCoalitionAgainstPlayer(gs *state.GameState, fid faction.FactionID) {
	playerRegions := len(gs.RegionsOwnedBy(gs.PlayerFactionID))
	if playerRegions < coalitionThreshold {
		return
	}

	// Bu fraksiyon oyuncuyla savaş halinde değilse; savaş ilan et
	playerKey := faction.RelationKey(fid, gs.PlayerFactionID)
	if rel, ok := gs.Relations[playerKey]; ok {
		if rel.Stance == faction.StancePeace || rel.Stance == faction.StanceAllied {
			rel.Stance = faction.StanceWar
			rel.Score = -80
		}
	}

	// Diğer AI fraksiyonlarla ittifak kur (düşman değillerse)
	for otherFID := range gs.Factions {
		if otherFID == fid || otherFID == gs.PlayerFactionID {
			continue
		}
		if gs.Factions[otherFID].IsEliminated {
			continue
		}
		key := faction.RelationKey(fid, otherFID)
		rel, ok := gs.Relations[key]
		if !ok {
			continue
		}
		if rel.Stance == faction.StanceWar {
			continue
		}
		// Skor yeterince iyiyse ittifak kur
		if rel.Score >= -20 {
			rel.Stance = faction.StanceAllied
			if rel.Score < 30 {
				rel.Score = 30
			}
		}
	}
}

// moveArmy tek bir orduyu hareket puanı tükenene kadar hareket ettirir.
func moveArmy(gs *state.GameState, a *army.Army) {
	for a.MovePoints > 0 {
		target := chooseBestMove(gs, a)
		if target == "" {
			break
		}
		survived := executeMove(gs, a, target)
		if !survived {
			break
		}
	}
}

// chooseBestMove ordunun komşuları arasında en iyi hedefi seçer.
// Negatif skor dönen hedefler atlanır; hiç geçerli hedef yoksa "" döner.
func chooseBestMove(gs *state.GameState, a *army.Army) world.RegionID {
	src, ok := gs.Regions[a.RegionID]
	if !ok {
		return ""
	}

	bestScore := 0
	var bestTarget world.RegionID

	for _, nid := range src.Neighbors {
		n, ok := gs.Regions[nid]
		if !ok || n.IsSea {
			continue
		}
		score := scoreMove(gs, a, n)
		if score > bestScore {
			bestScore = score
			bestTarget = nid
		}
	}
	return bestTarget
}

// scoreMove bir hedefe yapılacak hareketin değerini puanlar.
func scoreMove(gs *state.GameState, a *army.Army, target *world.Region) int {
	if target.OwnerID == a.OwnerID {
		return 0
	}

	// Barış halindeki fraksiyona saldırma
	if target.OwnerID != "" {
		_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
		if stance == faction.StancePeace || stance == faction.StanceAllied {
			return -1
		}
	}

	// Düşman ordusu var mı?
	for _, ea := range gs.Armies {
		if ea.RegionID != target.ID || ea.OwnerID == a.OwnerID {
			continue
		}
		atkStr := a.TotalStrength(gs.UnitTypes)
		defStr := ea.TotalStrength(gs.UnitTypes)
		if atkStr > defStr {
			// Savaş halindeyse öncelikli hedef
			_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
			if stance == faction.StanceWar {
				return 95
			}
			return 75
		}
		return -1
	}

	if target.OwnerID == "" {
		return 50
	}
	// Düşman bölgesi, ordu yok — ilişkiye göre puanla
	score, stance := relationScore(gs, a.OwnerID, target.OwnerID)
	if stance == faction.StanceWar || score < -40 {
		return 90
	}
	return 30
}

// executeMove hareketi ve varsa savaşı uygular.
// Ordu hayatta kaldıysa true, yok edildiyse false döner.
func executeMove(gs *state.GameState, a *army.Army, target world.RegionID) (survived bool) {
	targetRegion, ok := gs.Regions[target]
	if !ok {
		return true
	}

	// Hedefte düşman ordusu var mı?
	var enemyArmy *army.Army
	for _, ea := range gs.Armies {
		if ea.RegionID == target && ea.OwnerID != a.OwnerID {
			enemyArmy = ea
			break
		}
	}

	if enemyArmy != nil {
		atkMods := aiTechMods(gs, a.OwnerID)
		defMods := aiTechMods(gs, enemyArmy.OwnerID)
		result := combat.ResolveBattleWithMods(a, enemyArmy, targetRegion.Terrain, gs.UnitTypes, atkMods, defMods)
		if result.AttackerWins {
			if len(enemyArmy.Units) == 0 {
				delete(gs.Armies, enemyArmy.ID)
			}
			if len(a.Units) > 0 {
				a.RegionID = target
				targetRegion.OwnerID = a.OwnerID
				a.MovePoints--
				return true
			}
			delete(gs.Armies, a.ID)
			return false
		}
		// Saldıran yenildi
		if len(a.Units) == 0 {
			delete(gs.Armies, a.ID)
		}
		return false
	}

	// Savaşsız hareket
	a.RegionID = target
	a.MovePoints--
	targetRegion.OwnerID = a.OwnerID
	return true
}
