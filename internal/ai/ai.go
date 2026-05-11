package ai

import (
	"fmt"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

const (
	aiMilitiaID      = "militia"
	aiMilitiaCost    = 60  // units.json'daki milis maliyeti
	aiMinGoldReserve = 80  // AI bu miktarın altına düşmemeli
	aiTechReserve    = 100 // Teknoloji için ayırılacak minimum altın
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

	// Teknoloji araştırma (önce yap, altın biterse diğerlerini etkilemesin)
	aiResearch(gs, fid)

	// Ekonomi optimizasyonu (pazar, çiftlik)
	aiEconomyBuild(gs, fid)

	// Deniz stratejisi (liman + gemi)
	aiNavalStrategy(gs, fid)

	// Birim alımı ve kışla inşası (elite birimler dahil)
	aiRecruitAndBuild(gs, fid)

	// Aynı bölgede olan orduları konsolide et (önceki turlardan veya yeni alımlardan kalan)
	aiConsolidateArmies(gs, fid)

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

// aiRecruitAndBuild AI fraksiyonu için kışla inşa eder ve manpower sınırına kadar birim alır.
func aiRecruitAndBuild(gs *state.GameState, fid faction.FactionID) {
	f, ok := gs.Factions[fid]
	if !ok || f.IsEliminated {
		return
	}

	// Manpower dar ve altın yeterliyse kışla inşa et
	cap := gs.ManpowerCap(fid)
	deployed := gs.DeployedLandUnits(fid)
	barracksCost := 150
	if b, ok2 := gs.BuildingTypes["barracks"]; ok2 {
		barracksCost = b.GoldCost
	}
	if cap-deployed <= state.ManpowerPerRegion && f.Gold >= barracksCost+aiMinGoldReserve {
		aiBuildBarracks(gs, fid, barracksCost)
	}

	// Kapasite dolana veya altın bitene kadar birim al
	for {
		if gs.DeployedLandUnits(fid) >= gs.ManpowerCap(fid) {
			break
		}
		if f.Gold < aiMilitiaCost+aiMinGoldReserve {
			break
		}
		if !aiRecruitOne(gs, fid) {
			break
		}
	}
}

// aiBuildBarracks kışlası olmayan ilk uygun bölgeye kışla inşa eder.
func aiBuildBarracks(gs *state.GameState, fid faction.FactionID, cost int) {
	f := gs.Factions[fid]
	for _, r := range gs.Regions {
		if r.OwnerID != string(fid) || r.IsSea {
			continue
		}
		hasBarracks := false
		for _, bid := range r.Buildings {
			if bid == "barracks" {
				hasBarracks = true
				break
			}
		}
		if hasBarracks {
			continue
		}
		r.Buildings = append(r.Buildings, "barracks")
		f.Gold -= cost
		return
	}
}

// aiRecruitOne kışlası olan bir bölgede en iyi uygun birimi alır.
// Askeri teknoloji ve altın durumuna göre milis, piyade, süvari veya topçu seçer.
// Başarılıysa true, koşul sağlanamadıysa false döner.
func aiRecruitOne(gs *state.GameState, fid faction.FactionID) bool {
	f := gs.Factions[fid]
	if gs.UnitTypes == nil {
		return false
	}

	// Kışlası olan bir bölge bul
	var recruitRegion world.RegionID
	for _, r := range gs.Regions {
		if r.OwnerID != string(fid) || r.IsSea {
			continue
		}
		for _, bid := range r.Buildings {
			if bid == "barracks" {
				recruitRegion = r.ID
				break
			}
		}
		if recruitRegion != "" {
			break
		}
	}
	if recruitRegion == "" {
		return false
	}

	// Bölgedeki mevcut kara ordusu
	var targetArmy *army.Army
	for _, a := range gs.Armies {
		if a.RegionID == recruitRegion && a.OwnerID == string(fid) && !a.IsNaval {
			targetArmy = a
			break
		}
	}

	if targetArmy != nil {
		if len(targetArmy.Units) >= army.MaxArmySize {
			return false
		}
	} else {
		// Yeni ordu limiti kontrolü
		if gs.CurrentLandArmies(fid) >= gs.MaxLandArmies(fid) {
			return false
		}
		gs.NextArmySeq++
		newID := army.ArmyID(fmt.Sprintf("army_%s_%d", string(fid), gs.NextArmySeq))
		targetArmy = &army.Army{
			ID: newID, OwnerID: string(fid),
			RegionID:      recruitRegion,
			MovePoints:    2,
			MaxMovePoints: 2,
		}
		gs.Armies[newID] = targetArmy
	}

	// En iyi birimi seç (stratejik karar)
	unitTypeID := aiSelectBestUnit(gs, f)
	if unitTypeID == "" {
		return false
	}

	utype, ok := gs.UnitTypes[unitTypeID]
	if !ok {
		return false
	}

	targetArmy.Units = append(targetArmy.Units, army.Unit{TypeID: unitTypeID, CurrentHP: 100})
	f.Gold -= utype.GoldCost
	return true
}

// aiSelectBestUnit altın ve teknoloji durumuna göre en uygun birim tipini seçer.
// Öncelik: piyade > süvari > milis. Topçu sadece zengin AI'ler için.
func aiSelectBestUnit(gs *state.GameState, f *faction.Faction) string {
	// Askeri güç istatistiği
	armyCount := 0
	cavalryCount := 0
	for _, a := range gs.Armies {
		if a.OwnerID == string(f.ID) && !a.IsNaval {
			armyCount++
			for _, u := range a.Units {
				if ut, ok := gs.UnitTypes[u.TypeID]; ok && ut.Category == "cavalry" {
					cavalryCount++
				}
			}
		}
	}

	// Tier 3 elite piyade (seçkin piyade) - çok zenginse ve teknolojisi varsa
	if f.Gold >= 350+aiMinGoldReserve {
		if ut, ok := gs.UnitTypes["elite_infantry"]; ok {
			if ut.RequiredTech == "" || f.Research.Completed[ut.RequiredTech] {
				return "elite_infantry"
			}
		}
	}

	// Ağır süvari - zengin ve teknolojisi varsa
	if f.Gold >= 450+aiMinGoldReserve && cavalryCount < armyCount*2 {
		if ut, ok := gs.UnitTypes["heavy_cavalry"]; ok {
			if ut.RequiredTech == "" || f.Research.Completed[ut.RequiredTech] {
				return "heavy_cavalry"
			}
		}
	}

	// Tier 2 piyade (normal piyade) - orta düzey altın ve teknoloji
	if f.Gold >= 180+aiMinGoldReserve {
		if ut, ok := gs.UnitTypes["infantry"]; ok {
			if ut.RequiredTech == "" || f.Research.Completed[ut.RequiredTech] {
				return "infantry"
			}
		}
	}

	// Süvari - teknolojisi varsa ve altın yeterliyse
	if f.Gold >= 300+aiMinGoldReserve && cavalryCount < armyCount*3 {
		if ut, ok := gs.UnitTypes["cavalry"]; ok {
			if ut.RequiredTech == "" || f.Research.Completed[ut.RequiredTech] {
				return "cavalry"
			}
		}
	}

	// Hafif süvari - her zaman uygun
	if f.Gold >= 200+aiMinGoldReserve && cavalryCount < armyCount*4 {
		if _, ok := gs.UnitTypes["light_cavalry"]; ok {
			return "light_cavalry"
		}
	}

	// Topçu - çok zenginse ve savaşta ise
	if f.Gold >= 650+aiMinGoldReserve {
		// Savaş halinde mi kontrol et
		atWar := false
		for _, rel := range gs.Relations {
			if (rel.FactionA == f.ID || rel.FactionB == f.ID) && rel.Stance == faction.StanceWar {
				atWar = true
				break
			}
		}
		if atWar {
			if ut, ok := gs.UnitTypes["cannon"]; ok {
				if ut.RequiredTech == "" || f.Research.Completed[ut.RequiredTech] {
					return "cannon"
				}
			}
			if ut, ok := gs.UnitTypes["bombard"]; ok {
				if ut.RequiredTech == "" || f.Research.Completed[ut.RequiredTech] {
					return "bombard"
				}
			}
		}
	}

	// Varsayılan: milis
	return "militia"
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

	// Eğer komşularda mantıklı bir hedef yoksa, uzun menzilli planlama yap (BFS)
	if bestScore == 0 {
		bestTarget = findLongRangeMove(gs, a, src)
	}

	return bestTarget
}

// findLongRangeMove BFS kullanarak en yakın değerli (score > 0) bölgeye giden ilk adımı bulur.
func findLongRangeMove(gs *state.GameState, a *army.Army, start *world.Region) world.RegionID {
	type queueItem struct {
		id   world.RegionID
		path []world.RegionID
	}

	visited := make(map[world.RegionID]bool)
	queue := []queueItem{{id: start.ID, path: nil}}
	visited[start.ID] = true

	maxDepth := 8 // En fazla 8 bölge uzağa bak

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if len(curr.path) > maxDepth {
			continue
		}

		r, ok := gs.Regions[curr.id]
		if !ok {
			continue
		}

		// Kendi bölgesi değilse ve score > 0 ise hedef bulduk demektir
		if curr.id != start.ID {
			score := scoreMove(gs, a, r)
			if score > 0 {
				return curr.path[0] // Hedefe giden ilk adımı dön
			}
			// Düşman toprağıysa daha ileri gitme
			if r.OwnerID != a.OwnerID && r.OwnerID != "" {
				continue
			}
		}

		for _, nid := range r.Neighbors {
			n, ok := gs.Regions[nid]
			if !ok || n.IsSea || visited[nid] {
				continue
			}
			visited[nid] = true

			newPath := make([]world.RegionID, len(curr.path))
			copy(newPath, curr.path)
			newPath = append(newPath, nid)

			queue = append(queue, queueItem{id: nid, path: newPath})
		}
	}
	return ""
}

// scoreMove bir hedefe yapılacak hareketin değerini puanlar.
func scoreMove(gs *state.GameState, a *army.Army, target *world.Region) int {
	fid := faction.FactionID(a.OwnerID)
	if target.OwnerID == a.OwnerID {
		// Dost bölgede birleşebileceğimiz ordu var mı? (Konsolidasyon)
		for _, ea := range gs.Armies {
			if ea.RegionID == target.ID && ea.OwnerID == a.OwnerID && ea.ID != a.ID && ea.IsNaval == a.IsNaval {
				if len(a.Units)+len(ea.Units) <= army.MaxArmySize {
					return 60 // Birleşmek için iyi bir hedef
				}
			}
		}
		return 0
	}

	// Barış halindeki fraksiyona saldırma
	if target.OwnerID != "" {
		_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
		if stance == faction.StancePeace || stance == faction.StanceAllied {
			return -1
		}
	}

	// Kapasite doluysa fetih yaparak manpower artırmak öncelikli
	atCapacity := gs.DeployedLandUnits(fid) >= gs.ManpowerCap(fid)

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

	// Kapasite doluysa sahipsiz bölge almak çok değerli (manpower genişler)
	if target.OwnerID == "" {
		if atCapacity {
			return 70
		}
		return 50
	}
	// Düşman bölgesi, ordu yok — ilişkiye göre puanla
	score, stance := relationScore(gs, a.OwnerID, target.OwnerID)
	if stance == faction.StanceWar || score < -40 {
		if atCapacity {
			return 100
		}
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
				a.DockedRegionID = ""
				a.DockedSettlementID = ""
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
	a.DockedRegionID = ""
	a.DockedSettlementID = ""
	a.MovePoints--
	targetRegion.OwnerID = a.OwnerID

	// Konsolidasyon (Dost orduyla birleşme)
	if tryMergeAIArmies(gs, a) {
		return false // Ordu başka bir orduya katıldı ve silindi
	}

	return true
}

// aiResearch aktif araştırma yoksa stratejik teknoloji seçer ve başlatır.
// Öncelik: askeri > ekonomi > diplomasi > diğer
func aiResearch(gs *state.GameState, fid faction.FactionID) {
	f := gs.Factions[fid]
	if f.IsEliminated || gs.TechTypes == nil {
		return
	}
	// Zaten araştırma var mı?
	if f.Research.ActiveID != "" {
		return
	}
	// Yeterli altın var mı?
	if f.Gold < aiTechReserve {
		return
	}

	// Uygun teknolojileri puanla
	type scoredTech struct {
		t     *tech.Technology
		score int
	}
	var candidates []scoredTech

	for _, t := range gs.TechTypes {
		// Zaten tamamlandı mı?
		if f.Research.Completed[t.ID] {
			continue
		}
		// Gereksinimler sağlanıyor mu?
		if !tech.IsUnlocked(&f.Research, t) {
			continue
		}
		// Yeterli altın var mı?
		if f.Gold < t.GoldCost+aiMinGoldReserve {
			continue
		}

		score := 0
		switch t.Category {
		case tech.CategoryMilitary:
			score = 100 // Askeri teknolojiler en yüksek öncelik
			if t.Effects.InfantryAttackMod > 0 || t.Effects.CavalryAttackMod > 0 {
				score += 20
			}
		case tech.CategoryEconomy:
			score = 70 // Ekonomi ikinci öncelik
			if t.Effects.GoldPerRegion > 0 {
				score += 15
			}
		case tech.CategoryNaval:
			score = 50 // Deniz teknolojisi (kıyı fraksiyonları için daha yüksek olabilir)
		case tech.CategoryDiplomacy:
			score = 40
		case tech.CategoryReligion:
			score = 30
		}

		// Daha kısa süren teknolojilere hafif bonus
		score -= t.TurnsRequired / 2

		candidates = append(candidates, scoredTech{t, score})
	}

	if len(candidates) == 0 {
		return
	}

	// En yüksek puanlı teknolojiyi seç
	var best *tech.Technology
	bestScore := -1
	for _, c := range candidates {
		if c.score > bestScore {
			bestScore = c.score
			best = c.t
		}
	}

	if best != nil {
		if tech.StartResearch(&f.Research, best, &f.Gold) {
			// Araştırma başladı
		}
	}
}

// aiEconomyBuild ekonomik binalar inşa eder (pazar, çiftlik).
// Her tur sadece bir bina inşa eder (limitli).
func aiEconomyBuild(gs *state.GameState, fid faction.FactionID) {
	f := gs.Factions[fid]
	if f.IsEliminated || gs.BuildingTypes == nil {
		return
	}
	if f.Gold < 200+aiMinGoldReserve {
		return
	}

	// Bina öncelikleri ve maliyet kontrolü
	type buildingPlan struct {
		id     string
		needFn func(*world.Region) bool
		prio   int
	}

	plans := []buildingPlan{
		{"farm", func(r *world.Region) bool {
			// Tahıl üretimi düşük bölgelere çiftlik
			return r.BaseGrainOutput < 20
		}, 60},
		{"market", func(r *world.Region) bool {
			// Geliri artırmak için pazar
			return true
		}, 80},
		{"walls", func(r *world.Region) bool {
			// Sınır bölgelerine sur
			for _, nid := range r.Neighbors {
				if n, ok := gs.Regions[nid]; ok && n.OwnerID != "" && n.OwnerID != string(fid) {
					return true
				}
			}
			return false
		}, 50},
	}

	for _, plan := range plans {
		btype, ok := gs.BuildingTypes[plan.id]
		if !ok {
			continue
		}
		if f.Gold < btype.GoldCost+aiMinGoldReserve {
			continue
		}

		// Uygun bölge bul
		for _, r := range gs.Regions {
			if r.OwnerID != string(fid) || r.IsSea {
				continue
			}
			// Zaten var mı?
			hasIt := false
			for _, bid := range r.Buildings {
				if bid == plan.id {
					hasIt = true
					break
				}
			}
			if hasIt {
				continue
			}
			// Max per region kontrolü
			count := 0
			for _, bid := range r.Buildings {
				if bid == plan.id {
					count++
				}
			}
			if count >= btype.MaxPerRegion {
				continue
			}
			// İhtiyaç var mı?
			if plan.needFn(r) {
				r.Buildings = append(r.Buildings, plan.id)
				f.Gold -= btype.GoldCost
				return // Bir bina inşa ettik, turu bitir
			}
		}
	}
}

// aiNavalStrategy kıyı fraksiyonları için liman ve gemi inşası yapar.
func aiNavalStrategy(gs *state.GameState, fid faction.FactionID) {
	f := gs.Factions[fid]
	if f.IsEliminated || gs.BuildingTypes == nil || gs.UnitTypes == nil {
		return
	}

	// Kıyı bölgesi var mı?
	var coastalRegions []*world.Region
	for _, r := range gs.Regions {
		if r.OwnerID == string(fid) && !r.IsSea && r.IsCoastal(gs.Regions) {
			coastalRegions = append(coastalRegions, r)
		}
	}
	if len(coastalRegions) == 0 {
		return
	}

	// Liman tipi var mı?
	portType, hasPort := gs.BuildingTypes["port"]
	if !hasPort {
		return
	}
	transportType, hasTransport := gs.UnitTypes["transport"]
	if !hasTransport {
		return
	}

	// Liman inşası (en az bir liman olsun)
	for _, r := range coastalRegions {
		hasPortBldg := false
		for _, bid := range r.Buildings {
			if bid == "port" {
				hasPortBldg = true
				break
			}
		}
		if !hasPortBldg && f.Gold >= portType.GoldCost+aiMinGoldReserve {
			r.Buildings = append(r.Buildings, "port")
			f.Gold -= portType.GoldCost
			break // Bir liman yeter bu tur
		}
	}

	// Gemi alımı (liman olan bölgelerden)
	const fleetLimit = 2 // AI en fazla 2 filo
	fleetCount := 0
	for _, a := range gs.Armies {
		if a.OwnerID == string(fid) && a.IsNaval {
			fleetCount++
		}
	}
	if fleetCount >= fleetLimit {
		return
	}

	for _, r := range coastalRegions {
		// Liman var mı?
		hasPortBldg := false
		for _, bid := range r.Buildings {
			if bid == "port" {
				hasPortBldg = true
				break
			}
		}
		if !hasPortBldg {
			continue
		}

		// Komşu deniz bölgesi bul
		var seaRegion world.RegionID
		for _, nid := range r.Neighbors {
			if n, ok := gs.Regions[nid]; ok && n.IsSea {
				seaRegion = nid
				break
			}
		}
		if seaRegion == "" {
			continue
		}

		// Altın kontrolü
		if f.Gold < transportType.GoldCost+aiMinGoldReserve {
			return
		}

		// Yeni filo oluştur
		gs.NextArmySeq++
		newID := army.ArmyID(fmt.Sprintf("fleet_%s_%d", string(fid), gs.NextArmySeq))
		gs.Armies[newID] = &army.Army{
			ID:                 newID,
			OwnerID:            string(fid),
			RegionID:           seaRegion,
			DockedRegionID:     r.ID,
			DockedSettlementID: aiPreferredDockSettlementID(r),
			Units:              []army.Unit{{TypeID: "transport", CurrentHP: 100}},
			MovePoints:         3,
			MaxMovePoints:      3,
			IsNaval:            true,
		}
		f.Gold -= transportType.GoldCost
		return // Bir gemi aldık, turu bitir
	}
}

// aiConsolidateArmies aynı bölgedeki aynı tipteki (kara/deniz) kendi ordularını birleştirir.
func aiConsolidateArmies(gs *state.GameState, fid faction.FactionID) {
	var armies []*army.Army
	for _, a := range gs.Armies {
		if a.OwnerID == string(fid) {
			armies = append(armies, a)
		}
	}

	for i := 0; i < len(armies); i++ {
		a1 := armies[i]
		if _, ok := gs.Armies[a1.ID]; !ok {
			continue
		}
		for j := i + 1; j < len(armies); j++ {
			a2 := armies[j]
			if _, ok := gs.Armies[a2.ID]; !ok {
				continue
			}
			if a1.RegionID == a2.RegionID && a1.IsNaval == a2.IsNaval {
				if len(a1.Units)+len(a2.Units) <= army.MaxArmySize {
					a1.Units = append(a1.Units, a2.Units...)
					delete(gs.Armies, a2.ID)
				} else {
					transfer := army.MaxArmySize - len(a1.Units)
					if transfer > 0 {
						a1.Units = append(a1.Units, a2.Units[:transfer]...)
						a2.Units = a2.Units[transfer:]
					}
				}
			}
		}
	}
}

// tryMergeAIArmies hareket sonrası dost bölgede başka dost ordu varsa kapasite dahilinde birleşir.
// Birleşme sonucu ordu tamamen silinirse true döner.
func tryMergeAIArmies(gs *state.GameState, a *army.Army) bool {
	for otherID, other := range gs.Armies {
		if otherID == a.ID || other.RegionID != a.RegionID || other.OwnerID != a.OwnerID || other.IsNaval != a.IsNaval {
			continue
		}
		if len(a.Units)+len(other.Units) <= army.MaxArmySize {
			other.Units = append(other.Units, a.Units...)
			delete(gs.Armies, a.ID)
			return true
		} else {
			// Kapasite kadar aktar
			transfer := army.MaxArmySize - len(other.Units)
			if transfer > 0 {
				other.Units = append(other.Units, a.Units[:transfer]...)
				a.Units = a.Units[transfer:]
			}
		}
	}
	return false
}

func aiPreferredDockSettlementID(region *world.Region) string {
	if region == nil {
		return ""
	}
	for _, settlement := range region.Settlements {
		if settlement.Type == world.SettlementPort {
			return settlement.ID
		}
	}
	if len(region.Settlements) > 0 {
		return region.Settlements[0].ID
	}
	return ""
}
