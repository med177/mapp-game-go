package game

import (
	"fmt"
	"log"
	"os"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/events"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/save"
	"mapp-game-go/internal/scenario"
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
	loading  *loadingJob
}

const scenarioBaseDir = "assets/scenarios"

type loadingKind int

const (
	loadingScenario loadingKind = iota + 1
	loadingSave
)

type loadingJob struct {
	kind loadingKind
	done chan loadingResult
}

type loadingResult struct {
	gs           *state.GameState
	evts         []*events.Event
	scenarioPath string
	successMsg   string
	fallback     state.Phase
	err          error
}

// New oyunu başlatır, senaryo listesini yükler, ana menüde bekler.
func New() *Game {
	gs := &state.GameState{Phase: state.PhaseMainMenu}

	// Senaryo listesini yükle — render paketinin global değişkenine yaz
	scenarios, err := scenario.LoadAll(scenarioBaseDir)
	if err != nil {
		log.Printf("Senaryolar yüklenemedi: %v", err)
	}
	render.ScenarioList = scenarios

	r := render.New(gs)
	r.HasSave = save.AnySlotExists()
	r.HasAutoSave = save.SaveExists()
	r.CurrentSettings = render.DefaultSettings()
	return &Game{
		gs:       gs,
		renderer: r,
	}
}

// Update oyun mantığını günceller — 60 TPS.
func (g *Game) Update() error {
	if g.loading != nil {
		g.pollLoading()
		return nil
	}

	action := g.renderer.HandleInput()

	if action.Kind != render.ActionNone {
		audio.PlaySound("click")
	}

	switch g.gs.Phase {
	case state.PhaseMainMenu:
		switch action.Kind {
		case render.ActionNewGame:
			g.resetToNewGame()
		case render.ActionContinue:
			g.startLoadSlot("autosave", state.PhaseMainMenu)
		case render.ActionOpenLoadSelect:
			render.SaveSlots = save.ListSlots()
			g.gs.Phase = state.PhaseLoadSelect
		case render.ActionOpenSettings:
			g.gs.Phase = state.PhaseSettings
			g.renderer.SetCursor(0)
		case render.ActionQuit:
			os.Exit(0)
		}

	case state.PhaseLoadSelect:
		switch action.Kind {
		case render.ActionSelectSave:
			g.startLoadSlot(action.BuildingID, state.PhaseMainMenu)
		case render.ActionDeleteSave:
			if err := save.DeleteSlot(action.BuildingID); err != nil {
				g.renderer.ShowCombatResult("Silme hatası: " + err.Error())
			}
			render.SaveSlots = save.ListSlots()
			g.renderer.HasSave = save.AnySlotExists()
			g.renderer.HasAutoSave = save.SaveExists()
		case render.ActionBack:
			g.gs.Phase = state.PhaseMainMenu
		}

	case state.PhaseSettings:
		if action.Kind == render.ActionSaveSettings {
			g.gs.Difficulty = g.renderer.CurrentSettings.Difficulty
			g.gs.Phase = state.PhaseMainMenu
			g.renderer.SetCursor(0)
		}

	case state.PhaseScenarioSelect:
		switch action.Kind {
		case render.ActionSelectScenario:
			g.startLoadScenario(action.BuildingID)
		case render.ActionBack:
			g.gs.Phase = state.PhaseMainMenu
			g.renderer.SetCursor(0)
		}

	case state.PhaseFactionSelect:
		switch action.Kind {
		case render.ActionSelectFaction:
			g.gs.PlayerFactionID = action.TargetFaction
			g.gs.Phase = state.PhaseVictorySelect
		case render.ActionBack:
			g.gs.Phase = state.PhaseScenarioSelect
			g.renderer.SetCursor(0)
		}

	case state.PhaseVictorySelect:
		switch action.Kind {
		case render.ActionSelectVictory:
			g.applyVictoryChoice(action.BuildingID)
			g.gs.Phase = state.PhasePlayerTurn
		case render.ActionBack:
			g.gs.Phase = state.PhaseFactionSelect
			g.renderer.SetCursor(0)
		}

	case state.PhasePlayerTurn:
		switch action.Kind {
		case render.ActionEndTurn:
			if f, ok := g.gs.Factions[g.gs.PlayerFactionID]; ok && f.Research.ActiveID == "" {
				g.renderer.ShowConfirmDialog(
					"Araştırma Yok",
					"Teknoloji araştırması seçilmedi. Turu yine de bitirmek istiyor musunuz?",
					"Evet",
					"Hayır",
					render.InputAction{Kind: render.ActionConfirmEndTurn},
					func() {
						g.renderer.ShowTechPanel()
					},
				)
				break
			}
			g.gs.Phase = state.PhaseAITurn
		case render.ActionConfirmEndTurn:
			g.gs.Phase = state.PhaseAITurn
		case render.ActionMoveArmy:
			g.moveArmy(action.ArmyID, action.TargetRegion)
		case render.ActionSplitArmy:
			g.splitArmy(action.ArmyID)
		case render.ActionMergeArmies:
			g.mergeArmiesManual(action.ArmyID)
		case render.ActionRecruitUnit:
			g.recruitUnit(action.TargetRegion)
		case render.ActionRecruitNaval:
			g.recruitNaval(action.TargetRegion)
		case render.ActionRecruitSpecific:
			g.recruitSpecific(action.TargetRegion, action.BuildingID)
		case render.ActionBuild:
			g.buildBuilding(action.TargetRegion, action.BuildingID)
		case render.ActionResearch:
			g.startResearch(action.BuildingID) // BuildingID alanını tech ID için yeniden kullanıyoruz
		case render.ActionCancelResearch:
			g.cancelResearch()
		case render.ActionDeclareWar:
			g.declareWar(action.TargetFaction)
		case render.ActionDeclareWarAndMove:
			g.declareWar(action.TargetFaction)
			// Savaş ilan edildikten sonra relation map güncelleniyor,
			// moveArmy içinde bu güncel durum kontrol edilecek.
			g.moveArmy(action.ArmyID, action.TargetRegion)
		case render.ActionProposePeace:
			g.proposePeace(action.TargetFaction)
		case render.ActionProposeAlliance:
			g.proposeAlliance(action.TargetFaction)
		case render.ActionProposeTrade:
			g.proposeTrade(action.TargetFaction)
		case render.ActionSave:
			g.saveGame()
		case render.ActionLoad:
			g.startLoadSlot("autosave", state.PhasePlayerTurn)
		case render.ActionAdjustTax:
			g.adjustTax(action.TargetRegion, action.Delta)
		case render.ActionOpenPauseMenu:
			g.gs.Phase = state.PhasePauseMenu
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

	case state.PhasePauseMenu:
		switch action.Kind {
		case render.ActionResume:
			g.gs.Phase = state.PhasePlayerTurn
		case render.ActionOpenSaveSelect:
			render.SaveSlots = save.ListSlots()
			g.gs.Phase = state.PhaseSaveSelect
		case render.ActionLoadFromPause:
			render.SaveSlots = save.ListSlots()
			g.gs.Phase = state.PhaseLoadSelect
		case render.ActionGoMainMenu:
			g.resetToNewGame()
		case render.ActionQuit:
			os.Exit(0)
		}

	case state.PhaseSaveSelect:
		switch action.Kind {
		case render.ActionSelectSave:
			if err := save.SaveToSlot(g.gs, action.BuildingID); err != nil {
				g.renderer.ShowCombatResult("Kayıt hatası: " + err.Error())
			} else {
				g.renderer.HasSave = true
				g.renderer.HasAutoSave = save.SaveExists()
				g.renderer.ShowCombatResult("Kaydedildi!")
			}
			g.gs.Phase = state.PhasePlayerTurn
		case render.ActionDeleteSave:
			if err := save.DeleteSlot(action.BuildingID); err != nil {
				g.renderer.ShowCombatResult("Silme hatası: " + err.Error())
			}
			render.SaveSlots = save.ListSlots()
			g.renderer.HasSave = save.AnySlotExists()
			g.renderer.HasAutoSave = save.SaveExists()
		case render.ActionBack:
			g.gs.Phase = state.PhasePauseMenu
		}

	case state.PhaseGameOver:
		if action.Kind == render.ActionBack || action.Kind == render.ActionQuit {
			g.resetToNewGame()
		}
	}

	return nil
}

func (g *Game) startLoading(kind loadingKind, message string, fn func() loadingResult) {
	g.gs.Phase = state.PhaseLoading
	g.renderer.SetLoadingMessage(message)
	done := make(chan loadingResult, 1)
	g.loading = &loadingJob{kind: kind, done: done}
	go func() {
		done <- fn()
	}()
}

func (g *Game) pollLoading() {
	select {
	case res := <-g.loading.done:
		kind := g.loading.kind
		g.loading = nil
		g.finishLoading(kind, res)
	default:
	}
}

func (g *Game) finishLoading(kind loadingKind, res loadingResult) {
	if res.err != nil {
		g.renderer.ShowCombatResult("Yükleme hatası: " + res.err.Error())
		if res.fallback == "" {
			res.fallback = state.PhaseMainMenu
		}
		g.gs.Phase = res.fallback
		return
	}
	switch kind {
	case loadingScenario:
		g.gs = res.gs
		g.evts = res.evts
		audio.LoadScenarioSounds(res.scenarioPath)
		g.renderer.ReloadGameState(res.gs)
		g.renderer.SetCursor(0)
	case loadingSave:
		res.gs.Phase = state.PhasePlayerTurn
		g.gs = res.gs
		g.renderer.ReloadGameState(res.gs)
		g.renderer.HasSave = save.AnySlotExists()
		g.renderer.HasAutoSave = save.SaveExists()
		g.renderer.ShowCombatResult(res.successMsg)
	}
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
	completedTechs := applyTechTicks(g.gs)
	applyReligionConversion(g.gs)
	checkRegionUnlocks(g.gs)
	checkRebellions(g.gs)
	checkEliminations(g.gs)
	applyRelationDecay(g.gs)
	victory.Check(g.gs)

	// Tamamlanan teknolojiler için mesaj göster
	for _, ct := range completedTechs {
		if t, ok := g.gs.TechTypes[ct.techID]; ok {
			if f, ok := g.gs.Factions[faction.FactionID(ct.factionID)]; ok {
				msg := f.NameTR + ": " + t.NameTR + " teknolojisi tamamlandı!"
				g.renderer.ShowCombatResult(msg)
				g.renderer.AddEvent("🔬 " + msg)
			}
		}
	}

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
	g.renderer.HasAutoSave = true
	g.renderer.ShowCombatResult("Oyun kaydedildi!")
}

// loadGame otomatik kayıt slotundan yükler.
func (g *Game) loadGame() {
	g.startLoadSlot("autosave", state.PhasePlayerTurn)
}

// loadSlot belirtilen slottan oyunu yükler ve oyuncu turuna geçer.
func (g *Game) loadSlot(slotName string) {
	g.startLoadSlot(slotName, state.PhaseMainMenu)
}

func (g *Game) startLoadSlot(slotName string, fallback state.Phase) {
	g.startLoading(loadingSave, "Kayıt yükleniyor...", func() loadingResult {
		gs, err := save.LoadSlot(slotName)
		if err != nil {
			return loadingResult{err: err, fallback: fallback}
		}
		return loadingResult{
			gs:         gs,
			successMsg: "Oyun yüklendi!",
		}
	})
}

// resetToNewGame state'i temizler ve senaryo seçimine geçer.
func (g *Game) resetToNewGame() {
	difficulty := g.gs.Difficulty
	gs := &state.GameState{
		Phase:      state.PhaseScenarioSelect,
		Difficulty: difficulty,
	}
	g.gs = gs
	g.renderer.ReloadGameState(gs)
	g.renderer.SetCursor(0)
}

// loadScenario seçilen senaryo klasöründen tüm oyun verilerini yükler.
func (g *Game) loadScenario(scenarioPath string) {
	g.startLoadScenario(scenarioPath)
}

func (g *Game) startLoadScenario(scenarioPath string) {
	difficulty := g.gs.Difficulty
	g.startLoading(loadingScenario, "Senaryo yükleniyor...", func() loadingResult {
		gs, evts, err := loadScenarioData(scenarioPath, difficulty)
		if err != nil {
			return loadingResult{err: err, fallback: state.PhaseScenarioSelect}
		}
		return loadingResult{
			gs:           gs,
			evts:         evts,
			scenarioPath: scenarioPath,
		}
	})
}

func loadScenarioData(scenarioPath string, difficulty int) (*state.GameState, []*events.Event, error) {
	sc := scenarioByPath(scenarioPath)

	dp := func(f string) string { return scenarioPath + "/data/" + f }

	regions, err := world.LoadRegions(dp("regions.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("bölgeler yüklenemedi: %w", err)
	}
	shapeData, err := world.LoadCountryShapes(dp("country_shapes.json"), regions)
	if err != nil {
		log.Printf("Ülke sınırları yüklenemedi: %v", err)
	}
	factions, err := faction.LoadFactions(dp("factions.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("fraksiyonlar yüklenemedi: %w", err)
	}
	unitTypes, err := army.LoadUnitTypes(dp("units.json"))
	if err != nil {
		log.Printf("Birim tipleri yüklenemedi: %v", err)
	}
	buildingTypes, err := city.LoadBuildings(dp("buildings.json"))
	if err != nil {
		log.Printf("Binalar yüklenemedi: %v", err)
	}
	techTypes, err := tech.LoadTechnologies(dp("technologies.json"))
	if err != nil {
		log.Printf("Teknolojiler yüklenemedi: %v", err)
	}
	evts, err := events.LoadEvents(dp("events.json"))
	if err != nil {
		log.Printf("Olaylar yüklenemedi: %v", err)
	}
	armies, err := army.LoadArmies(dp("armies.json"))
	if err != nil {
		log.Printf("Ordular yüklenemedi: %v", err)
		armies = map[army.ArmyID]*army.Army{}
	}

	devMode := os.Getenv("DEV_MODE") == "true"

	year := 1300
	month := 3
	var mapConfig scenario.MapConfig
	var victoryOpts []scenario.VictoryOptionDef
	if sc != nil {
		year = sc.Year
		month = sc.Month
		mapConfig = sc.MapConfig
		victoryOpts = sc.VictoryConditions
	}

	gs := &state.GameState{
		Turn:               1,
		Year:               year,
		Month:              month,
		StartYear:          year,
		Phase:              state.PhaseFactionSelect,
		Difficulty:         difficulty,
		DevelopmentMode:    devMode,
		ScenarioID:         scenarioIDFromPath(scenarioPath),
		ScenarioPath:       scenarioPath,
		MapConfig:          mapConfig,
		Regions:            regions,
		Factions:           factions,
		Armies:             armies,
		ShapeData:          shapeData,
		UnitTypes:          unitTypes,
		BuildingTypes:      buildingTypes,
		TechTypes:          techTypes,
		AvailableVictories: victoryOpts,
		Relations:          faction.BuildInitialRelations(factions),
		NextArmySeq:        len(armies),
		FiredEventIDs:      map[string]bool{},
	}

	if gs.Difficulty >= 3 {
		for fid, f := range gs.Factions {
			if fid != gs.PlayerFactionID {
				f.Gold += 300
				f.Grain += 100
			}
		}
	}

	return gs, evts, nil
}

// scenarioByPath ScenarioList içinde verilen path'e sahip senaryoyu bulur.
func scenarioByPath(path string) *scenario.Scenario {
	for _, s := range render.ScenarioList {
		if s.Path == path {
			return s
		}
	}
	return nil
}

// scenarioIDFromPath klasör yolundan senaryo ID'sini çıkarır.
func scenarioIDFromPath(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
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
	hasBarracks := false
	for _, bid := range region.Buildings {
		if bid == "barracks" {
			hasBarracks = true
			break
		}
	}
	if !hasBarracks {
		g.renderer.ShowCombatResult("Asker almak için kışla gerekli!")
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

	// Manpower kontrolü
	pid := g.gs.PlayerFactionID
	deployed := g.gs.DeployedLandUnits(pid)
	cap := g.gs.ManpowerCap(pid)
	if deployed >= cap {
		g.renderer.ShowCombatResult(fmt.Sprintf("Savaşçı kapasitesi dolu! (%d/%d) — Bölge fethet veya kışla yap.", deployed, cap))
		return
	}

	// Bölgede mevcut ordu var mı?
	var targetArmy *army.Army
	for _, a := range g.gs.Armies {
		if a.RegionID == rid && a.OwnerID == string(pid) && !a.IsNaval {
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
		// Ordu sayısı limiti
		if g.gs.CurrentLandArmies(pid) >= g.gs.MaxLandArmies(pid) {
			g.renderer.ShowCombatResult(fmt.Sprintf("Maksimum ordu sayısına ulaşıldı! (%d/%d) — Daha fazla bölge gerekli.", g.gs.CurrentLandArmies(pid), g.gs.MaxLandArmies(pid)))
			return
		}
		// Yeni ordu oluştur
		g.gs.NextArmySeq++
		newID := army.ArmyID(fmt.Sprintf("army_%s_%d", string(pid), g.gs.NextArmySeq))
		g.gs.Armies[newID] = &army.Army{
			ID:            newID,
			OwnerID:       string(pid),
			RegionID:      rid,
			Units:         []army.Unit{{TypeID: "militia", CurrentHP: 100}},
			MovePoints:    2,
			MaxMovePoints: 2,
		}
	}

	f.Gold -= cost
	g.renderer.ShowCombatResult(fmt.Sprintf("Milis alındı! Kalan altın: %d", f.Gold))
}

// recruitSpecific seçili bölgede belirli türde bir birim alır.
func (g *Game) recruitSpecific(rid world.RegionID, unitTypeID string) {
	region, ok := g.gs.Regions[rid]
	if !ok || region.IsSea || region.OwnerID != string(g.gs.PlayerFactionID) {
		g.renderer.ShowCombatResult("Sadece kendi bölgende asker alabilirsin!")
		return
	}
	utype, ok := g.gs.UnitTypes[unitTypeID]
	if !ok {
		return
	}
	f, ok := g.gs.Factions[g.gs.PlayerFactionID]
	if !ok {
		return
	}

	// Bina kontrolü
	hasBldg := false
	for _, bid := range region.Buildings {
		if bid == utype.RequiredBldg {
			hasBldg = true
			break
		}
	}
	if !hasBldg {
		bldgName := utype.RequiredBldg
		if b, ok2 := g.gs.BuildingTypes[bldgName]; ok2 {
			bldgName = b.NameTR
		}
		g.renderer.ShowCombatResult("Bu birlik için " + bldgName + " gerekli!")
		return
	}

	// Teknoloji kontrolü
	if utype.RequiredTech != "" && !f.Research.Completed[utype.RequiredTech] {
		g.renderer.ShowCombatResult("Araştırma gerekli: " + utype.RequiredTech)
		return
	}

	// Altın kontrolü
	if f.Gold < utype.GoldCost {
		g.renderer.ShowCombatResult(fmt.Sprintf("Yetersiz altın! Gerekli: %d, Mevcut: %d", utype.GoldCost, f.Gold))
		return
	}

	// Deniz birimi — komşu deniz bölgesine yerleş
	if utype.RequiredBldg == "port" {
		if !region.IsCoastal(g.gs.Regions) {
			g.renderer.ShowCombatResult("Bu bölge kıyıda değil!")
			return
		}
		var seaRegion world.RegionID
		for _, nid := range region.Neighbors {
			if n, ok2 := g.gs.Regions[nid]; ok2 && n.IsSea {
				seaRegion = nid
				break
			}
		}
		if seaRegion == "" {
			g.renderer.ShowCombatResult("Komşu deniz bölgesi bulunamadı!")
			return
		}
		// Mevcut filo var mı?
		var fleet *army.Army
		for _, a := range g.gs.Armies {
			if a.RegionID == seaRegion && a.OwnerID == string(g.gs.PlayerFactionID) && a.IsNaval {
				fleet = a
				break
			}
		}
		f.Gold -= utype.GoldCost
		if fleet != nil {
			if len(fleet.Units) >= army.MaxArmySize {
				g.renderer.ShowCombatResult("Filo dolu! (max 20 birim)")
				return
			}
			fleet.Units = append(fleet.Units, army.Unit{TypeID: unitTypeID, CurrentHP: 100})
		} else {
			g.gs.NextArmySeq++
			newID := army.ArmyID(fmt.Sprintf("fleet_%s_%d", string(g.gs.PlayerFactionID), g.gs.NextArmySeq))
			g.gs.Armies[newID] = &army.Army{
				ID: newID, OwnerID: string(g.gs.PlayerFactionID),
				RegionID: seaRegion, IsNaval: true,
				Units:      []army.Unit{{TypeID: unitTypeID, CurrentHP: 100}},
				MovePoints: 3, MaxMovePoints: 3,
			}
		}
		g.renderer.ShowCombatResult(fmt.Sprintf("%s denize indi! Kalan altın: %d", utype.NameTR, f.Gold))
		return
	}

	// Kara birimi — manpower ve ordu sayısı kontrolü
	pid := g.gs.PlayerFactionID
	deployed := g.gs.DeployedLandUnits(pid)
	cap := g.gs.ManpowerCap(pid)
	if deployed >= cap {
		g.renderer.ShowCombatResult(fmt.Sprintf("Savaşçı kapasitesi dolu! (%d/%d) — Bölge fethet veya kışla yap.", deployed, cap))
		return
	}

	var targetArmy *army.Army
	for _, a := range g.gs.Armies {
		if a.RegionID == rid && a.OwnerID == string(pid) && !a.IsNaval {
			targetArmy = a
			break
		}
	}
	f.Gold -= utype.GoldCost
	if targetArmy != nil {
		if len(targetArmy.Units) >= army.MaxArmySize {
			g.renderer.ShowCombatResult("Ordu dolu! (max 20 birim)")
			f.Gold += utype.GoldCost
			return
		}
		targetArmy.Units = append(targetArmy.Units, army.Unit{TypeID: unitTypeID, CurrentHP: 100})
	} else {
		if g.gs.CurrentLandArmies(pid) >= g.gs.MaxLandArmies(pid) {
			g.renderer.ShowCombatResult(fmt.Sprintf("Maksimum ordu sayısına ulaşıldı! (%d/%d)", g.gs.CurrentLandArmies(pid), g.gs.MaxLandArmies(pid)))
			f.Gold += utype.GoldCost
			return
		}
		g.gs.NextArmySeq++
		newID := army.ArmyID(fmt.Sprintf("army_%s_%d", string(pid), g.gs.NextArmySeq))
		g.gs.Armies[newID] = &army.Army{
			ID: newID, OwnerID: string(pid),
			RegionID:   rid,
			Units:      []army.Unit{{TypeID: unitTypeID, CurrentHP: 100}},
			MovePoints: 2, MaxMovePoints: 2,
		}
	}
	g.renderer.ShowCombatResult(fmt.Sprintf("%s alındı! Kalan altın: %d", utype.NameTR, f.Gold))
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
	// Sahipli düşman bölgeye girmek için savaş hali zorunlu
	if targetRegion.OwnerID != "" && targetRegion.OwnerID != a.OwnerID {
		key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID))
		rel, exists := g.gs.Relations[key]
		if !exists || rel.Stance != faction.StanceWar {
			g.renderer.ShowCombatResult("Savaş ilan edilmeden düşman topraklarına girilemez!")
			return
		}
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
		// Dost bölgede başka ordu varsa birleştir
		if merged := g.tryMergeArmies(aid, target); merged != "" {
			g.renderer.SelectedArmy = merged
		}
	}
}

// tryMergeArmies taşınan orduyu hedefteki dost orduyla birleştirir.
// Birleşme olursa hayatta kalan ordu ID'sini döner; yoksa "".
func (g *Game) tryMergeArmies(movingID army.ArmyID, regionID world.RegionID) army.ArmyID {
	moving, ok := g.gs.Armies[movingID]
	if !ok {
		return ""
	}
	for otherID, other := range g.gs.Armies {
		if otherID == movingID || other.RegionID != regionID ||
			other.OwnerID != moving.OwnerID || other.IsNaval != moving.IsNaval {
			continue
		}
		if len(moving.Units)+len(other.Units) <= army.MaxArmySize {
			// Taşınanı hedefe ekle, taşınanı sil
			other.Units = append(other.Units, moving.Units...)
			delete(g.gs.Armies, movingID)
			g.renderer.AddEvent("Ordular birleşti: " + fmt.Sprintf("%d", len(other.Units)) + " birim")
			return otherID
		}
		// 20'yi aşıyor — iki ayrı ordu olarak bırak
		return ""
	}
	return ""
}

// splitArmy seçili orduyu birim sayısına göre ikiye böler.
func (g *Game) splitArmy(aid army.ArmyID) {
	a, ok := g.gs.Armies[aid]
	if !ok || len(a.Units) < 2 {
		return
	}
	half := len(a.Units) / 2
	newUnits := make([]army.Unit, half)
	copy(newUnits, a.Units[len(a.Units)-half:])
	a.Units = a.Units[:len(a.Units)-half]

	g.gs.NextArmySeq++
	newID := army.ArmyID(fmt.Sprintf("army_%s_%d", string(g.gs.PlayerFactionID), g.gs.NextArmySeq))
	g.gs.Armies[newID] = &army.Army{
		ID:            newID,
		OwnerID:       a.OwnerID,
		RegionID:      a.RegionID,
		Units:         newUnits,
		MovePoints:    a.MovePoints,
		MaxMovePoints: a.MaxMovePoints,
		IsNaval:       a.IsNaval,
	}
	g.renderer.AddEvent(fmt.Sprintf("Ordu bölündü: %d + %d birim", len(a.Units), len(newUnits)))
}

// mergeArmiesManual seçili orduyu aynı bölgedeki dost orduya elle birleştirir (20 kapasitesine kadar).
func (g *Game) mergeArmiesManual(aid army.ArmyID) {
	a, ok := g.gs.Armies[aid]
	if !ok {
		return
	}
	// Aynı bölgede dost ordu bul
	var target *army.Army
	var targetID army.ArmyID
	for oid, other := range g.gs.Armies {
		if oid == aid || other.RegionID != a.RegionID ||
			other.OwnerID != a.OwnerID || other.IsNaval != a.IsNaval {
			continue
		}
		target = other
		targetID = oid
		break
	}
	if target == nil {
		return
	}
	capacity := army.MaxArmySize - len(target.Units)
	if capacity <= 0 {
		g.renderer.ShowCombatResult("Hedef ordu dolu!")
		return
	}
	transfer := a.Units
	if len(transfer) > capacity {
		transfer = transfer[:capacity]
	}
	target.Units = append(target.Units, transfer...)
	a.Units = a.Units[len(transfer):]

	if len(a.Units) == 0 {
		delete(g.gs.Armies, aid)
		g.renderer.SelectedArmy = targetID
	}
	g.renderer.AddEvent(fmt.Sprintf("Ordular birleşti: %d birim", len(target.Units)))
}

// adjustTax oyuncunun bölgesinde vergi oranını ayarlar.
func (g *Game) adjustTax(rid world.RegionID, delta int) {
	r, ok := g.gs.Regions[rid]
	if !ok || r.OwnerID != string(g.gs.PlayerFactionID) {
		return
	}
	r.TaxRate = clamp(r.TaxRate+delta, 0, 100)
}

// applyVictoryChoice seçilen zafer koşulunu senaryodan okuyarak GameState'e yazar.
func (g *Game) applyVictoryChoice(optionID string) {
	opt, ok := render.VictoryOptionByID(g.gs, optionID)
	if !ok {
		return
	}

	requiredRegions := make([]world.RegionID, len(opt.RequiredRegions))
	for i, r := range opt.RequiredRegions {
		requiredRegions[i] = world.RegionID(r)
	}
	// conquer_city: tek hedef bölgeyi required listesine çevir
	if opt.Type == "conquer_city" && opt.Target != "" {
		requiredRegions = []world.RegionID{world.RegionID(opt.Target)}
	}

	g.gs.Victory = state.VictoryCondition{
		Type:               state.VictoryType(opt.Type),
		TargetRegionCount:  opt.TargetRegionCount,
		RequiredRegions:    requiredRegions,
		TargetGoldIncome:   opt.TargetGoldIncome,
		GoldHoldTurns:      opt.GoldHoldTurns,
		TargetArmyStrength: opt.TargetArmyStrength,
		TargetDefeated:     opt.TargetDefeated,
		DeadlineTurn:       opt.DeadlineTurn,
	}
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

// cancelResearch aktif teknoloji araştırmasını iptal eder.
func (g *Game) cancelResearch() {
	f := g.gs.Factions[g.gs.PlayerFactionID]
	if f.Research.ActiveID == "" {
		g.renderer.ShowCombatResult("Aktif araştırma yok!")
		return
	}
	refundedGold := 0
	if tech, ok := g.gs.TechTypes[f.Research.ActiveID]; ok {
		refundedGold = tech.GoldCost
	}
	f.Gold += refundedGold
	f.Research.ActiveID = ""
	f.Research.TurnsLeft = 0
	if refundedGold > 0 {
		g.renderer.ShowCombatResult(fmt.Sprintf("Araştırma iptal edildi! %d altın iade edildi.", refundedGold))
	} else {
		g.renderer.ShowCombatResult("Araştırma iptal edildi!")
	}
}
