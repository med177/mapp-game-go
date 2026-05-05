package game

import (
	"fmt"
	"log"
	"os"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/events"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/save"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/victory"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game Ebitengine'in Game interface'ini uygular.
type Game struct {
	gs       *state.GameState
	renderer *render.Renderer
	evts     []*events.Event
}

// New oyunu başlatır, veri dosyalarını yükler.
func New() *Game {
	gs, err := loadGameState()
	if err != nil {
		log.Fatalf("Oyun verisi yüklenemedi: %v", err)
	}
	evts, err := events.LoadEvents("assets/data/events.json")
	if err != nil {
		log.Printf("Olaylar yüklenemedi: %v", err)
	}
	r := render.New(gs)
	r.HasSave = saveExists()
	r.CurrentSettings = render.DefaultSettings()
	return &Game{
		gs:       gs,
		renderer: r,
		evts:     evts,
	}
}

// Update oyun mantığını günceller — 60 TPS.
func (g *Game) Update() error {
	action := g.renderer.HandleInput()

	switch g.gs.Phase {
	case state.PhaseMainMenu:
		switch action.Kind {
		case render.ActionNewGame:
			g.resetToNewGame()
		case render.ActionContinue:
			g.loadGame()
		case render.ActionOpenSettings:
			g.gs.Phase = state.PhaseSettings
			g.renderer.SetCursor(0)
		case render.ActionQuit:
			os.Exit(0)
		}

	case state.PhaseSettings:
		if action.Kind == render.ActionSaveSettings {
			g.gs.Difficulty = g.renderer.CurrentSettings.Difficulty
			g.gs.Phase = state.PhaseMainMenu
			g.renderer.SetCursor(0)
		}

	case state.PhaseFactionSelect:
		switch action.Kind {
		case render.ActionSelectFaction:
			g.gs.PlayerFactionID = action.TargetFaction
			g.gs.Phase = state.PhaseVictorySelect
		case render.ActionBack:
			g.gs.Phase = state.PhaseMainMenu
			g.renderer.SetCursor(0)
		}

	case state.PhaseVictorySelect:
		switch action.Kind {
		case render.ActionSelectVictory:
			g.applyVictoryChoice(state.VictoryType(action.BuildingID))
			g.gs.Phase = state.PhasePlayerTurn
		case render.ActionBack:
			g.gs.Phase = state.PhaseFactionSelect
			g.renderer.SetCursor(0)
		}

	case state.PhasePlayerTurn:
		switch action.Kind {
		case render.ActionEndTurn:
			g.gs.Phase = state.PhaseAITurn
		case render.ActionMoveArmy:
			g.moveArmy(action.ArmyID, action.TargetRegion)
		case render.ActionRecruitUnit:
			g.recruitUnit(action.TargetRegion)
		case render.ActionRecruitNaval:
			g.recruitNaval(action.TargetRegion)
		case render.ActionBuild:
			g.buildBuilding(action.TargetRegion, action.BuildingID)
		case render.ActionResearch:
			g.startResearch(action.BuildingID) // BuildingID alanını tech ID için yeniden kullanıyoruz
		case render.ActionDeclareWar:
			g.declareWar(action.TargetFaction)
		case render.ActionProposePeace:
			g.proposePeace(action.TargetFaction)
		case render.ActionProposeAlliance:
			g.proposeAlliance(action.TargetFaction)
		case render.ActionProposeTrade:
			g.proposeTrade(action.TargetFaction)
		case render.ActionSave:
			g.saveGame()
		case render.ActionLoad:
			g.loadGame()
		case render.ActionAdjustTax:
			g.adjustTax(action.TargetRegion, action.Delta)
		}

	case state.PhaseAITurn:
		for fid := range g.gs.Factions {
			if fid == g.gs.PlayerFactionID {
				continue
			}
			ai.TakeTurn(g.gs, fid)
		}
		g.renderer.MarkMapDirty()
		g.gs.Phase = state.PhaseTurnResolution

	case state.PhaseTurnResolution:
		g.resolveTurn()

	case state.PhaseGameOver:
		if action.Kind == render.ActionBack || action.Kind == render.ActionQuit {
			g.resetToNewGame()
		}
	}

	return nil
}

// Draw ekranı çizer.
func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen)
}

// Layout pencere boyutlarını bildirir — mantıksal ekran = fiziksel pencere (letterbox yok).
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	render.ScreenWidth = float64(outsideWidth)
	render.ScreenHeight = float64(outsideHeight)
	return outsideWidth, outsideHeight
}

func (g *Game) resolveTurn() {
	applySeasonEffects(g.gs)
	applyEconomyTick(g.gs)
	applyTechTicks(g.gs)
	applyReligionConversion(g.gs)
	checkRegionUnlocks(g.gs)
	checkRebellions(g.gs)
	checkEliminations(g.gs)
	applyRelationDecay(g.gs)
	victory.Check(g.gs)

	// Olaylar
	if name, desc := events.Tick(g.gs, g.evts); name != "" {
		g.renderer.ShowCombatResult("📜 " + name + ": " + desc)
	}

	g.gs.AdvanceTurn()
	if g.gs.Phase != state.PhaseGameOver {
		g.gs.Phase = state.PhasePlayerTurn
		g.renderer.MarkMapDirty()
	}
}

// buildBuilding oyuncunun kendi bölgesine bina inşa eder.
func (g *Game) buildBuilding(rid world.RegionID, buildingID string) {
	region, ok := g.gs.Regions[rid]
	if !ok || region.IsSea || region.OwnerID != string(g.gs.PlayerFactionID) {
		g.renderer.ShowCombatResult("Sadece kendi bölgene bina yapabilirsin!")
		return
	}
	b, ok := g.gs.BuildingTypes[buildingID]
	if !ok {
		return
	}
	// Gerekli arazi kontrolü
	if b.RequiredTerrain != "" && string(region.Terrain) != b.RequiredTerrain {
		g.renderer.ShowCombatResult(b.NameTR + " sadece " + b.RequiredTerrain + " arazisine yapılır!")
		return
	}
	// Zaten inşa edilmiş mi?
	count := 0
	for _, bid := range region.Buildings {
		if bid == buildingID {
			count++
		}
	}
	if count >= b.MaxPerRegion {
		g.renderer.ShowCombatResult(b.NameTR + " bu bölgede zaten var!")
		return
	}
	f := g.gs.Factions[g.gs.PlayerFactionID]
	if f.Gold < b.GoldCost {
		g.renderer.ShowCombatResult(fmt.Sprintf("Yeterli altın yok! Gerekli: %d", b.GoldCost))
		return
	}
	f.Gold -= b.GoldCost
	region.Buildings = append(region.Buildings, buildingID)
	g.renderer.ShowCombatResult(b.NameTR + " inşa edildi!")
}

// declareWar hedef fraksiyona savaş ilan eder.
func (g *Game) declareWar(targetID faction.FactionID) {
	key := faction.RelationKey(g.gs.PlayerFactionID, targetID)
	rel, ok := g.gs.Relations[key]
	if !ok {
		rel = &faction.Relation{FactionA: g.gs.PlayerFactionID, FactionB: targetID}
		g.gs.Relations[key] = rel
	}
	if rel.Stance == faction.StanceWar {
		g.renderer.ShowCombatResult("Zaten savaş halindeyiz!")
		return
	}
	rel.Stance = faction.StanceWar
	rel.Score = -80
	if f, ok := g.gs.Factions[targetID]; ok {
		g.renderer.ShowCombatResult(f.NameTR + "'a savaş ilan edildi!")
	}
}

// proposeAlliance hedefe ittifak teklif eder (savaş halinde değilse kabul edilir).
func (g *Game) proposeAlliance(targetID faction.FactionID) {
	key := faction.RelationKey(g.gs.PlayerFactionID, targetID)
	rel, ok := g.gs.Relations[key]
	if !ok {
		rel = &faction.Relation{FactionA: g.gs.PlayerFactionID, FactionB: targetID}
		g.gs.Relations[key] = rel
	}
	if rel.Stance == faction.StanceWar {
		g.renderer.ShowCombatResult("Savaş halindeyken ittifak kurulamaz!")
		return
	}
	if rel.Stance == faction.StanceAllied {
		g.renderer.ShowCombatResult("Zaten müttefiksiniz.")
		return
	}
	if rel.Score < -20 {
		g.renderer.ShowCombatResult("İlişki çok düşük — önce ilişkileri iyileştir.")
		return
	}
	rel.Stance = faction.StanceAllied
	rel.Score = clamp(rel.Score+30, -100, 100)
	if f, ok := g.gs.Factions[targetID]; ok {
		g.renderer.ShowCombatResult(f.NameTR + " ile ittifak kuruldu!")
	}
}

// proposeTrade hedefe ticaret anlaşması teklif eder.
func (g *Game) proposeTrade(targetID faction.FactionID) {
	key := faction.RelationKey(g.gs.PlayerFactionID, targetID)
	rel, ok := g.gs.Relations[key]
	if !ok {
		rel = &faction.Relation{FactionA: g.gs.PlayerFactionID, FactionB: targetID}
		g.gs.Relations[key] = rel
	}
	if rel.Stance == faction.StanceWar {
		g.renderer.ShowCombatResult("Savaş halindeyken ticaret yapılamaz!")
		return
	}
	if rel.Stance == faction.StanceTrade || rel.Stance == faction.StanceAllied {
		g.renderer.ShowCombatResult("Zaten ticaret anlaşması var.")
		return
	}
	rel.Stance = faction.StanceTrade
	rel.Score = clamp(rel.Score+15, -100, 100)
	if f, ok := g.gs.Factions[targetID]; ok {
		g.renderer.ShowCombatResult(f.NameTR + " ile ticaret anlaşması imzalandı!")
	}
}

// proposePeace hedefe barış teklif eder (her zaman kabul edilir — basit versiyon).
func (g *Game) proposePeace(targetID faction.FactionID) {
	key := faction.RelationKey(g.gs.PlayerFactionID, targetID)
	rel, ok := g.gs.Relations[key]
	if !ok || rel.Stance != faction.StanceWar {
		g.renderer.ShowCombatResult("Savaş halinde olmadığınız bir fraksiyona barış teklifiniz geçersiz.")
		return
	}
	rel.Stance = faction.StancePeace
	rel.Score = -20
	if f, ok := g.gs.Factions[targetID]; ok {
		g.renderer.ShowCombatResult(f.NameTR + " barışı kabul etti.")
	}
}

// saveGame oyunu kaydeder.
func (g *Game) saveGame() {
	if err := save.Save(g.gs); err != nil {
		g.renderer.ShowCombatResult("Kayıt hatası: " + err.Error())
		return
	}
	g.renderer.HasSave = true
	g.renderer.ShowCombatResult("Oyun kaydedildi!")
}

// loadGame kaydedilmiş oyunu yükler.
func (g *Game) loadGame() {
	gs, err := save.Load("assets/data/units.json", "assets/data/buildings.json")
	if err != nil {
		g.renderer.ShowCombatResult("Yükleme hatası: " + err.Error())
		return
	}
	g.gs = gs
	g.renderer.ReloadGameState(gs)
	g.renderer.ShowCombatResult("Oyun yüklendi!")
}

// resetToNewGame tüm state'i sıfırlayıp fraksiyon seçimine geçer.
func (g *Game) resetToNewGame() {
	gs, err := loadGameState()
	if err != nil {
		log.Printf("Yeni oyun başlatılamadı: %v", err)
		return
	}
	gs.Phase = state.PhaseFactionSelect
	gs.Difficulty = g.gs.Difficulty // ayarlardan zorluk koru

	// Difficulty 3: AI fraksiyonlara başlangıç kaynak bonusu ver
	if gs.Difficulty >= 3 {
		for fid, f := range gs.Factions {
			if fid != gs.PlayerFactionID {
				f.Gold += 300
				f.Grain += 100
			}
		}
	}

	g.gs = gs
	g.renderer.ReloadGameState(gs)
	g.renderer.SetCursor(0)
}

// saveExists autosave dosyasının var olup olmadığını kontrol eder.
func saveExists() bool {
	_, err := os.Stat("saves/autosave.json")
	return err == nil
}

// recruitNaval kıyı bölgesinde nakliye gemisi oluşturur.
func (g *Game) recruitNaval(rid world.RegionID) {
	region, ok := g.gs.Regions[rid]
	if !ok || region.IsSea || region.OwnerID != string(g.gs.PlayerFactionID) {
		g.renderer.ShowCombatResult("Sadece kendi kıyı bölgene gemi yapabilirsin!")
		return
	}
	// Liman gereksinimi
	hasPort := false
	for _, bid := range region.Buildings {
		if bid == "port" {
			hasPort = true
			break
		}
	}
	if !hasPort {
		g.renderer.ShowCombatResult("Gemi için liman gerekli!")
		return
	}
	// Kıyı bölgesi mi?
	if !region.IsCoastal(g.gs.Regions) {
		g.renderer.ShowCombatResult("Bu bölge kıyıda değil!")
		return
	}
	f := g.gs.Factions[g.gs.PlayerFactionID]
	const cost = 200
	if f.Gold < cost {
		g.renderer.ShowCombatResult(fmt.Sprintf("Nakliye gemisi için %d altın gerekli!", cost))
		return
	}

	// Komşu deniz bölgesini bul
	var seaRegion world.RegionID
	for _, nid := range region.Neighbors {
		if n, ok := g.gs.Regions[nid]; ok && n.IsSea {
			seaRegion = nid
			break
		}
	}
	if seaRegion == "" {
		g.renderer.ShowCombatResult("Komşu deniz bölgesi bulunamadı!")
		return
	}

	f.Gold -= cost
	g.gs.NextArmySeq++
	newID := army.ArmyID(fmt.Sprintf("fleet_%s_%d", string(g.gs.PlayerFactionID), g.gs.NextArmySeq))
	g.gs.Armies[newID] = &army.Army{
		ID:            newID,
		OwnerID:       string(g.gs.PlayerFactionID),
		RegionID:      seaRegion,
		Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
		MovePoints:    3,
		MaxMovePoints: 3,
		IsNaval:       true,
	}
	g.renderer.ShowCombatResult(fmt.Sprintf("Nakliye gemisi denize indi! Kalan altın: %d", f.Gold))
}

// recruitUnit seçili bölgede oyuncu adına bir milis birimi alır.
func (g *Game) recruitUnit(rid world.RegionID) {
	region, ok := g.gs.Regions[rid]
	if !ok || region.IsSea || region.OwnerID != string(g.gs.PlayerFactionID) {
		g.renderer.ShowCombatResult("Sadece kendi bölgene asker alabilirsin!")
		return
	}
	f, ok := g.gs.Factions[g.gs.PlayerFactionID]
	if !ok {
		return
	}
	const cost = 60
	if f.Gold < cost {
		g.renderer.ShowCombatResult(fmt.Sprintf("Yeterli altın yok! Gerekli: %d, Mevcut: %d", cost, f.Gold))
		return
	}

	// Bölgede mevcut ordu var mı?
	var targetArmy *army.Army
	for _, a := range g.gs.Armies {
		if a.RegionID == rid && a.OwnerID == string(g.gs.PlayerFactionID) {
			targetArmy = a
			break
		}
	}

	if targetArmy != nil {
		if len(targetArmy.Units) >= army.MaxArmySize {
			g.renderer.ShowCombatResult("Ordu dolu! (max 20 birim)")
			return
		}
		targetArmy.Units = append(targetArmy.Units, army.Unit{TypeID: "militia", CurrentHP: 100})
	} else {
		// Yeni ordu oluştur
		g.gs.NextArmySeq++
		newID := army.ArmyID(fmt.Sprintf("army_%s_%d", string(g.gs.PlayerFactionID), g.gs.NextArmySeq))
		g.gs.Armies[newID] = &army.Army{
			ID:            newID,
			OwnerID:       string(g.gs.PlayerFactionID),
			RegionID:      rid,
			Units:         []army.Unit{{TypeID: "militia", CurrentHP: 100}},
			MovePoints:    2,
			MaxMovePoints: 2,
		}
	}

	f.Gold -= cost
	g.renderer.ShowCombatResult(fmt.Sprintf("Milis alındı! Kalan altın: %d", f.Gold))
}

// moveArmy oyuncu ordusunu hedef bölgeye taşır; gerekirse savaş başlatır.
func (g *Game) moveArmy(aid army.ArmyID, target world.RegionID) {
	a, ok := g.gs.Armies[aid]
	if !ok || a.OwnerID != string(g.gs.PlayerFactionID) {
		return
	}
	if a.MovePoints <= 0 {
		g.renderer.ShowCombatResult("Hareket puanı kalmadı!")
		return
	}

	// Komşu mu kontrol et
	src, ok := g.gs.Regions[a.RegionID]
	if !ok {
		return
	}
	isNeighbor := false
	for _, n := range src.Neighbors {
		if n == target {
			isNeighbor = true
			break
		}
	}
	if !isNeighbor {
		return
	}

	targetRegion, ok := g.gs.Regions[target]
	if !ok {
		return
	}

	// Naval/kara uyumluluk kontrolü
	if a.IsNaval && !targetRegion.CanNavalEnter() {
		g.renderer.ShowCombatResult("Deniz ordusu sadece deniz bölgelerine gidebilir!")
		return
	}
	if !a.IsNaval && !targetRegion.CanLandEnter() {
		g.renderer.ShowCombatResult("Kara ordusu denize giremez! (Nakliye gemisi gerekir)")
		return
	}
	var enemyArmy *army.Army
	for _, ea := range g.gs.Armies {
		if ea.RegionID == target && ea.OwnerID != a.OwnerID {
			enemyArmy = ea
			break
		}
	}

	if enemyArmy != nil {
		// --- Savaş ---
		atkMods := techModsFor(g.gs, a.OwnerID)
		defMods := techModsFor(g.gs, enemyArmy.OwnerID)
		result := combat.ResolveBattleWithMods(a, enemyArmy, targetRegion.Terrain, g.gs.UnitTypes, atkMods, defMods)

		if result.AttackerWins {
			if len(enemyArmy.Units) == 0 {
				delete(g.gs.Armies, enemyArmy.ID)
			}
			if len(a.Units) > 0 {
				a.RegionID = target
				attackerReligion := ownerReligion(g.gs, a.OwnerID)
				targetRegion.ApplyConquest(a.OwnerID, attackerReligion)
				a.MovePoints--
				g.renderer.MarkMapDirty()
			} else {
				delete(g.gs.Armies, aid)
			}
		} else {
			// Saldıran yenildi — yerinde kalır
			if len(a.Units) == 0 {
				delete(g.gs.Armies, aid)
			}
		}

		msg := fmt.Sprintf("%s: +%d / -%d birim", result.Description, result.DefenderLost, result.AttackerLost)
		g.renderer.ShowCombatResult(msg)

	} else {
		// --- Savaşsız hareket ve bölge ele geçirme ---
		a.RegionID = target
		a.MovePoints--
		if targetRegion.OwnerID != a.OwnerID {
			attackerReligion := ownerReligion(g.gs, a.OwnerID)
			targetRegion.ApplyConquest(a.OwnerID, attackerReligion)
			g.renderer.MarkMapDirty()
		}
	}
}

// adjustTax oyuncunun bölgesinde vergi oranını ayarlar.
func (g *Game) adjustTax(rid world.RegionID, delta int) {
	r, ok := g.gs.Regions[rid]
	if !ok || r.OwnerID != string(g.gs.PlayerFactionID) {
		return
	}
	r.TaxRate = clamp(r.TaxRate+delta, 0, 100)
}

// applyVictoryChoice seçilen zafer koşulunu GameState'e yazar.
func (g *Game) applyVictoryChoice(vtype state.VictoryType) {
	vc := state.VictoryCondition{Type: vtype}
	switch vtype {
	case state.VictoryDomination:
		vc.TargetRegionCount = 20
		vc.RequiredRegions = []world.RegionID{"constantinople", "rome", "paris", "cairo", "jerusalem"}
	case state.VictoryEconomic:
		vc.TargetGoldIncome = 500
		vc.GoldHoldTurns = 5
	case state.VictoryMilitary:
		vc.TargetDefeated = 3
		vc.TargetArmyStrength = 200
	case state.VictoryReligious:
		vc.RequiredRegions = []world.RegionID{"jerusalem", "rome", "mecca"}
	}
	g.gs.Victory = vc
}

// startResearch oyuncu fraksiyonu için teknolojiyi araştırmaya başlar.
func (g *Game) startResearch(techID string) {
	t, ok := g.gs.TechTypes[techID]
	if !ok {
		return
	}
	f, ok := g.gs.Factions[g.gs.PlayerFactionID]
	if !ok {
		return
	}
	if tech.StartResearch(&f.Research, t, &f.Gold) {
		g.renderer.ShowCombatResult(t.NameTR + " araştırması başladı! (" + fmt.Sprintf("%d tur", t.TurnsRequired) + ")")
	} else if f.Research.ActiveID != "" {
		g.renderer.ShowCombatResult("Zaten bir araştırma sürüyor!")
	} else {
		g.renderer.ShowCombatResult("Araştırma başlatılamadı. Altın veya gereksinim eksik.")
	}
}

// ownerReligion bir fraksiyonun dinini string olarak döner.
func ownerReligion(gs *state.GameState, ownerID string) string {
	for fid, f := range gs.Factions {
		if string(fid) == ownerID {
			return string(f.Religion)
		}
	}
	return ""
}

// loadGameState JSON dosyalarından tam bir oyun state'i yükler.
func loadGameState() (*state.GameState, error) {
	regions, err := world.LoadRegions("assets/data/regions.json")
	if err != nil {
		return nil, fmt.Errorf("bölgeler: %w", err)
	}
	shapeData, err := world.LoadCountryShapes("assets/data/generated/country_shapes.json", regions)
	if err != nil {
		return nil, fmt.Errorf("ülke sınırları: %w", err)
	}

	factions, err := faction.LoadFactions("assets/data/factions.json")
	if err != nil {
		return nil, fmt.Errorf("fraksiyonlar: %w", err)
	}

	unitTypes, err := army.LoadUnitTypes("assets/data/units.json")
	if err != nil {
		return nil, fmt.Errorf("birim tipleri: %w", err)
	}
	buildingTypes, err := city.LoadBuildings("assets/data/buildings.json")
	if err != nil {
		return nil, fmt.Errorf("binalar: %w", err)
	}

	techTypes, err := tech.LoadTechnologies("assets/data/technologies.json")
	if err != nil {
		return nil, fmt.Errorf("teknolojiler: %w", err)
	}

	relations := faction.BuildInitialRelations(factions)

	armies := buildStartingArmies()

	return &state.GameState{
		Turn:            1,
		Year:            1300,
		Month:           3,
		StartYear:       1300,
		Phase:           state.PhaseMainMenu,
		PlayerFactionID: "",
		Difficulty:      2,
		Regions:         regions,
		Factions:        factions,
		Armies:          armies,
		ShapeData:       shapeData,
		UnitTypes:       unitTypes,
		BuildingTypes:   buildingTypes,
		TechTypes:       techTypes,
		Relations:       relations,
	}, nil
}

// buildStartingArmies her fraksiyon için başlangıç ordularını oluşturur.
func buildStartingArmies() map[army.ArmyID]*army.Army {
	armies := make(map[army.ArmyID]*army.Army)

	type armySpec struct {
		id      army.ArmyID
		owner   string
		region  world.RegionID
		militia int
		cavalry int
	}

	specs := []armySpec{
		{"army_ottoman_1", "ottoman", "anatolia", 5, 2},
		{"army_france_1", "france", "france", 4, 1},
		{"army_england_1", "england", "england", 4, 0},
		{"army_venice_1", "venice", "venice", 3, 1},
		{"army_mamluk_1", "mamluk", "cairo", 4, 1},
		{"army_safavid_1", "safavid", "tabriz", 4, 1},
		{"army_russia_1", "russia", "moscow", 4, 0},
		{"army_aragon_1", "aragon", "aragon", 3, 1},
		{"army_portugal_1", "portugal", "portugal", 3, 0},
	}

	for _, s := range specs {
		units := army.MakeUnits("militia", s.militia)
		if s.cavalry > 0 {
			units = append(units, army.MakeUnits("light_cavalry", s.cavalry)...)
		}
		armies[s.id] = &army.Army{
			ID:            s.id,
			OwnerID:       s.owner,
			RegionID:      s.region,
			Units:         units,
			MovePoints:    2,
			MaxMovePoints: 2,
		}
	}

	return armies
}
