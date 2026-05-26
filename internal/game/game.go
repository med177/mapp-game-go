package game

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
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
	audio.LoadGlobalSounds("assets/sounds")

	// Senaryo listesini yükle — render paketinin global değişkenine yaz
	scenarios, err := scenario.LoadAll(scenarioBaseDir)
	if err != nil {
		log.Printf("Senaryolar yüklenemedi: %v", err)
	}
	render.ScenarioList = scenarios

	r := render.New(gs)
	r.HasSave = save.AnySlotExists()
	r.HasAutoSave = save.SaveExists()
	r.CurrentSettings = render.LoadSettings()
	audio.SetMusicEnabled(r.CurrentSettings.MusicOn)
	audio.SetMusicVolume(r.CurrentSettings.MusicVolume)
	audio.SetSoundEnabled(r.CurrentSettings.SoundOn)
	audio.SetSoundVolume(r.CurrentSettings.SoundVolume)
	return &Game{
		gs:       gs,
		renderer: r,
	}
}

// Update oyun mantığını günceller — 60 TPS.
func (g *Game) Update() error {
	if g.loading != nil {
		audio.UpdateMusic()
		g.pollLoading()
		return nil
	}
	audio.UpdateMusic()

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
			audio.SetMusicEnabled(g.renderer.CurrentSettings.MusicOn)
			audio.SetMusicVolume(g.renderer.CurrentSettings.MusicVolume)
			audio.SetSoundEnabled(g.renderer.CurrentSettings.SoundOn)
			audio.SetSoundVolume(g.renderer.CurrentSettings.SoundVolume)
			render.SaveSettingsToFile(g.renderer.CurrentSettings)
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
			g.applyAIDifficultyStartBonus()
			g.gs.Phase = state.PhasePlayerTurn
		case render.ActionBack:
			g.gs.Phase = state.PhaseFactionSelect
			g.renderer.SetCursor(0)
		}

	case state.PhasePlayerTurn:
		switch action.Kind {
		case render.ActionEndTurn:
			if f, ok := g.gs.Factions[g.gs.PlayerFactionID]; ok &&
				f.Research.ActiveID == "" &&
				g.playerHasRemainingTechs() {
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
			if !g.saveToSlot("autosave", false, "") {
				break
			}
			g.gs.Phase = state.PhaseAITurn
		case render.ActionConfirmEndTurn:
			if !g.saveToSlot("autosave", false, "") {
				break
			}
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
			g.recruitSpecific(action.TargetRegion, action.BuildingID, action.Quantity)
		case render.ActionCancelRecruitOrder:
			g.cancelRecruitOrder(action.TargetRegion, action.BuildingID)
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
		case render.ActionRespondDiplomacyOffer:
			g.respondDiplomacyOffer(action.OfferIndex, action.OfferAccepted)
		case render.ActionSave:
			g.saveToSlot("quicksave", true, "Hızlı kayıt alındı.")
		case render.ActionLoad:
			g.startLoadSlot("autosave", state.PhasePlayerTurn)
		case render.ActionAdjustTax:
			g.adjustTax(action.TargetRegion, action.Delta)
		case render.ActionOpenPauseMenu:
			g.gs.Phase = state.PhasePauseMenu
		case render.ActionToggleMusic:
			g.renderer.CurrentSettings.MusicOn = audio.ToggleMusic()
			render.SaveSettingsToFile(g.renderer.CurrentSettings)
		case render.ActionNextMusic:
			audio.NextMusic()
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
		case render.ActionToggleMusic:
			g.renderer.CurrentSettings.MusicOn = audio.ToggleMusic()
			render.SaveSettingsToFile(g.renderer.CurrentSettings)
		case render.ActionAdjustMusic:
			g.renderer.CurrentSettings.MusicVolume = audio.AdjustMusicVolume(action.Delta)
			render.SaveSettingsToFile(g.renderer.CurrentSettings)
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

	case state.PhaseEditMode:
		switch action.Kind {
		case render.ActionSaveScenario:
			g.saveScenarioRegions()
		case render.ActionSaveScenarioAndGoMainMenu:
			if g.saveScenarioRegions() {
				g.resetToNewGame()
			}
		case render.ActionGoMainMenu:
			g.resetToNewGame()
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
		g.sanitizeDockedFleets()
		g.evts = res.evts
		g.renderer.ReloadGameState(res.gs)
		g.startScenarioMusic(res.gs.ScenarioPath)
		g.renderer.SetCursor(0)
	case loadingSave:
		res.gs.Phase = state.PhasePlayerTurn
		g.gs = res.gs
		g.sanitizeDockedFleets()
		g.evts = res.evts
		g.renderer.ReloadGameState(res.gs)
		g.startScenarioMusic(res.gs.ScenarioPath)
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
	g.sanitizeDockedFleets()
	applySeasonEffects(g.gs)
	applyEconomyTick(g.gs)
	completedTechs := applyTechTicks(g.gs)
	productionResults := g.applyProductionTicks()
	applyReligionConversion(g.gs)
	checkRebellions(g.gs)
	checkEliminations(g.gs)
	applyRelationDecay(g.gs)
	prevVictoryAchieved := g.gs.VictoryAchieved
	victory.Check(g.gs)
	if !prevVictoryAchieved && g.gs.VictoryAchieved && g.gs.WinnerID == g.gs.PlayerFactionID {
		msg := "Zafer hedefi tamamlandı: " + victoryLabel(g.gs.Victory.Type) + ". Oyun devam ediyor."
		g.renderer.ShowCombatResult(msg)
		g.renderer.AddEvent("[ZAFER] " + msg)
	}

	// Tamamlanan teknolojiler için mesaj göster
	for _, ct := range completedTechs {
		if t, ok := g.gs.TechTypes[ct.techID]; ok {
			if f, ok := g.gs.Factions[faction.FactionID(ct.factionID)]; ok {
				msg := f.NameTR + ": " + t.NameTR + " teknolojisi tamamlandı!"
				g.renderer.ShowCombatResult(msg)
				g.renderer.AddEvent("[TEKNOLOJI] " + msg)
			}
		}
	}

	for _, pr := range productionResults {
		if pr.factionID != g.gs.PlayerFactionID {
			continue
		}
		name := g.productionName(pr)
		regionName := string(pr.regionID)
		if r, ok := g.gs.Regions[pr.regionID]; ok {
			regionName = r.NameTR
		}
		switch {
		case pr.delayed:
			g.renderer.ShowCombatResult(fmt.Sprintf("%s hazır, ancak %s nedeniyle bekliyor.", name, pr.reason))
		case pr.canceled:
			g.renderer.ShowCombatResult(fmt.Sprintf("%s üretimi iptal oldu: %s.", name, pr.reason))
		case pr.kind == productionKindBuilding:
			msg := fmt.Sprintf("%s bölgesinde %s tamamlandı!", regionName, name)
			g.renderer.ShowCombatResult(msg)
			g.renderer.AddEvent("[INSA] " + msg)
		case pr.kind == productionKindUnit:
			msg := fmt.Sprintf("%s bölgesinde %s hazır!", regionName, name)
			g.renderer.ShowCombatResult(msg)
			g.renderer.AddEvent("[BIRIM] " + msg)
		}
	}

	// Olaylar
	if name, desc := events.Tick(g.gs, g.evts); name != "" {
		g.renderer.ShowCombatResult("OLAY: " + name + ": " + desc)
	}

	g.gs.AdvanceTurn()
	unlocked := checkRegionUnlocks(g.gs)
	g.showRegionUnlockNotifications(unlocked)
	if g.gs.Phase != state.PhaseGameOver {
		g.gs.Phase = state.PhasePlayerTurn
		g.renderer.MarkMapDirty()
	}
}

func victoryLabel(vtype state.VictoryType) string {
	switch vtype {
	case state.VictoryDomination:
		return "Toprak Hakimiyeti"
	case state.VictoryEconomic:
		return "Ekonomik Üstünlük"
	case state.VictoryMilitary:
		return "Askeri Üstünlük"
	case state.VictoryReligious:
		return "Dinî Zafer"
	case state.VictoryConquerCity:
		return "Fetih"
	default:
		return "Zafer"
	}
}

func (g *Game) playerHasRemainingTechs() bool {
	if g == nil || g.gs == nil {
		return false
	}
	f := g.gs.Factions[g.gs.PlayerFactionID]
	if f == nil {
		return false
	}
	completed := f.Research.Completed
	for techID := range g.gs.TechTypes {
		if completed == nil || !completed[techID] {
			return true
		}
	}
	return false
}

func (g *Game) showRegionUnlockNotifications(ids []world.RegionID) {
	if len(ids) == 0 {
		return
	}
	names := make([]string, 0, len(ids))
	seen := make(map[world.RegionID]bool, len(ids))
	for _, rid := range ids {
		if seen[rid] {
			continue
		}
		seen[rid] = true
		name := string(rid)
		if region := g.gs.Regions[rid]; region != nil {
			if region.NameTR != "" {
				name = region.NameTR
			} else if region.Name != "" {
				name = region.Name
			}
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return
	}
	sort.Strings(names)
	msg := "Yeni bölge açıldı: " + names[0]
	if len(names) > 1 {
		msg = "Yeni bölgeler açıldı: " + names[0]
		limit := len(names)
		if limit > 3 {
			limit = 3
		}
		for i := 1; i < limit; i++ {
			msg += ", " + names[i]
		}
		if len(names) > limit {
			msg += fmt.Sprintf(" +%d", len(names)-limit)
		}
	}
	g.renderer.ShowCombatResult(msg)
	g.renderer.AddEvent("[UNLOCK] " + msg)
}

// buildBuilding oyuncunun kendi bölgesine bina inşa eder.
func (g *Game) buildBuilding(rid world.RegionID, buildingID string) {
	region, ok := g.gs.Regions[rid]
	if !ok || region.IsSea || region.OwnerID != string(g.gs.PlayerFactionID) {
		g.renderer.ShowCombatResult("Sadece kendi bölgene bina yapabilirsin!")
		return
	}
	if region.IsLocked {
		g.renderer.ShowCombatResult("Bu bölge kilitli; inşa açılamaz.")
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
	// Maks seviye kontrolü
	count := 0
	for _, bid := range region.Buildings {
		if bid == buildingID {
			count++
		}
	}
	if count >= b.MaxPerRegion {
		g.renderer.ShowCombatResult(fmt.Sprintf("%s maksimum seviyede! (Lv%d)", b.NameTR, b.MaxPerRegion))
		return
	}
	f := g.gs.Factions[g.gs.PlayerFactionID]
	if g.cancelProduction(productionKindBuilding, rid, buildingID, g.gs.PlayerFactionID) {
		f.Gold += b.GoldCost
		g.renderer.ShowCombatResult(fmt.Sprintf("%s inşaatı iptal edildi. %d altın iade edildi.", b.NameTR, b.GoldCost))
		return
	}
	if f.Gold < b.GoldCost {
		g.renderer.ShowCombatResult(fmt.Sprintf("Yeterli altın yok! Gerekli: %d", b.GoldCost))
		return
	}
	if count+g.queuedBuildingCount(rid, buildingID) >= b.MaxPerRegion {
		g.renderer.ShowCombatResult(fmt.Sprintf("%s için seviye kuyruğu dolu! (Lv%d)", b.NameTR, b.MaxPerRegion))
		return
	}
	f.Gold -= b.GoldCost
	g.enqueueProduction(productionKindBuilding, rid, buildingID, b.TurnsRequired)
	g.renderer.ShowCombatResult(fmt.Sprintf("%s seviye inşaatı başladı! Lv%d→Lv%d (%d tur)", b.NameTR, count+1, count+2, b.TurnsRequired))
}

// declareWar hedef fraksiyona savaş ilan eder.
func (g *Game) declareWar(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionDeclareWar)
	g.renderer.ShowCombatResult(result.Message)
}

// proposeAlliance hedefe ittifak teklif eder (savaş halinde değilse kabul edilir).
func (g *Game) proposeAlliance(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionProposeAlliance)
	g.renderer.ShowCombatResult(result.Message)
}

// proposeTrade hedefe ticaret anlaşması teklif eder.
func (g *Game) proposeTrade(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionProposeTrade)
	g.renderer.ShowCombatResult(result.Message)
}

// proposePeace hedefe barış teklif eder (her zaman kabul edilir — basit versiyon).
func (g *Game) proposePeace(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionProposePeace)
	g.renderer.ShowCombatResult(result.Message)
}

func (g *Game) respondDiplomacyOffer(index int, accepted bool) {
	result := diplomacy.ResolveOffer(g.gs, index, accepted)
	g.renderer.ShowCombatResult(result.Message)
	if accepted && result.Applied {
		g.renderer.AddEvent("[DIPLOMASI] " + result.Message)
	}
}

func (g *Game) saveToSlot(slotName string, showSuccess bool, successMsg string) bool {
	if err := save.SaveToSlot(g.gs, slotName); err != nil {
		g.renderer.ShowCombatResult("Kayıt hatası: " + err.Error())
		return false
	}
	g.renderer.HasSave = true
	if slotName == "autosave" {
		g.renderer.HasAutoSave = true
	}
	if showSuccess {
		msg := successMsg
		if msg == "" {
			msg = "Oyun kaydedildi!"
		}
		g.renderer.ShowCombatResult(msg)
	}
	return true
}

// saveGame oyunu otomatik kayıt slotuna kaydeder (geriye dönük çağrılar için).
func (g *Game) saveGame() {
	g.saveToSlot("autosave", true, "Oyun kaydedildi!")
}

func (g *Game) saveScenarioRegions() bool {
	if g.gs.ScenarioPath == "" {
		g.renderer.ShowCombatResult("Senaryo yolu yok; kaydedilemedi.")
		return false
	}
	if err := writeScenarioEditData(g.gs); err != nil {
		g.renderer.ShowCombatResult("Senaryo kayıt hatası: " + err.Error())
		return false
	}
	g.renderer.MarkEditSaved()
	g.renderer.ShowCombatResult("Senaryo verileri kaydedildi.")
	return true
}

func writeScenarioEditData(gs *state.GameState) error {
	if err := writeScenarioRegions(gs); err != nil {
		return err
	}
	if err := writeScenarioShapes(gs); err != nil {
		return err
	}
	if err := writeScenarioFactions(gs); err != nil {
		return err
	}
	if err := writeScenarioRelations(gs); err != nil {
		return err
	}
	if err := writeScenarioArmies(gs); err != nil {
		return err
	}
	// Region paint overrides'larını region_shapes.json'a kaydet
	if gs.RegionPaintOverrides != nil {
		path := filepath.Join(gs.ScenarioPath, "data", "region_shapes.json")
		if err := render.SaveRegionPaintOverrides(path, gs.RegionPaintOverrides); err != nil {
			return err
		}
	}
	return nil
}

func writeScenarioRegions(gs *state.GameState) error {
	path := filepath.Join(gs.ScenarioPath, "data", "regions.json")
	regions := make([]*world.Region, 0, len(gs.Regions))
	if len(gs.RegionOrder) > 0 {
		seen := make(map[world.RegionID]bool, len(gs.RegionOrder))
		for _, rid := range gs.RegionOrder {
			if region, ok := gs.Regions[rid]; ok {
				regions = append(regions, region)
				seen[rid] = true
			}
		}
		for rid, region := range gs.Regions {
			if !seen[rid] {
				regions = append(regions, region)
			}
		}
	} else {
		for _, region := range gs.Regions {
			regions = append(regions, region)
		}
	}

	data, err := json.MarshalIndent(regions, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func writeScenarioFactions(gs *state.GameState) error {
	path := filepath.Join(gs.ScenarioPath, "data", "factions.json")
	ids := make([]faction.FactionID, 0, len(gs.Factions))
	for fid := range gs.Factions {
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	factions := make([]*faction.Faction, 0, len(ids))
	for _, fid := range ids {
		if f := gs.Factions[fid]; f != nil {
			factions = append(factions, f)
		}
	}
	data, err := json.MarshalIndent(factions, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func writeScenarioShapes(gs *state.GameState) error {
	path := filepath.Join(gs.ScenarioPath, "data", "country_shapes.json")
	type shapeEntryJSON struct {
		ID    string     `json:"id"`
		Name  string     `json:"name,omitempty"`
		Rings [][][2]int `json:"rings"`
	}
	type shapeFileJSON struct {
		Shapes []shapeEntryJSON `json:"shapes"`
	}

	ids := make([]string, 0, len(gs.ShapeData.Shapes))
	for id := range gs.ShapeData.Shapes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	entries := make([]shapeEntryJSON, 0, len(ids))
	for _, id := range ids {
		rings := gs.ShapeData.Shapes[id]
		intRings := make([][][2]int, 0, len(rings))
		for _, ring := range rings {
			if len(ring) < 3 {
				continue
			}
			intRing := make([][2]int, 0, len(ring))
			for _, pt := range ring {
				intRing = append(intRing, [2]int{int(pt[0] + 0.5), int(pt[1] + 0.5)})
			}
			if len(intRing) >= 3 {
				intRings = append(intRings, intRing)
			}
		}
		if len(intRings) == 0 {
			continue
		}
		entry := shapeEntryJSON{ID: id, Rings: intRings}
		if name := gs.ShapeData.Names[id]; name != "" {
			entry.Name = name
		}
		entries = append(entries, entry)
	}

	data, err := json.MarshalIndent(shapeFileJSON{Shapes: entries}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func writeScenarioRelations(gs *state.GameState) error {
	path := filepath.Join(gs.ScenarioPath, "data", "relations.json")
	keys := make([]string, 0, len(gs.Relations))
	for key := range gs.Relations {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	relations := make([]*faction.Relation, 0, len(keys))
	for _, key := range keys {
		rel := gs.Relations[key]
		if rel == nil || gs.Factions[rel.FactionA] == nil || gs.Factions[rel.FactionB] == nil {
			continue
		}
		relations = append(relations, rel)
	}
	data, err := json.MarshalIndent(relations, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func writeScenarioArmies(gs *state.GameState) error {
	path := filepath.Join(gs.ScenarioPath, "data", "armies.json")
	ids := make([]army.ArmyID, 0, len(gs.Armies))
	for aid := range gs.Armies {
		ids = append(ids, aid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	type unitCountJSON struct {
		TypeID string `json:"type_id"`
		Count  int    `json:"count"`
	}
	type armySpecJSON struct {
		ID                 string          `json:"id"`
		OwnerID            string          `json:"owner_id"`
		Region             world.RegionID  `json:"region_id"`
		DockedRegion       world.RegionID  `json:"docked_region_id,omitempty"`
		DockedSettlementID string          `json:"docked_settlement_id,omitempty"`
		IsNaval            bool            `json:"is_naval,omitempty"`
		Units              []unitCountJSON `json:"units"`
	}
	specs := make([]armySpecJSON, 0, len(ids))
	for _, aid := range ids {
		a := gs.Armies[aid]
		if a == nil {
			continue
		}
		counts := make(map[string]int, len(a.Units))
		for _, u := range a.Units {
			counts[u.TypeID]++
		}
		unitIDs := make([]string, 0, len(counts))
		for typeID := range counts {
			unitIDs = append(unitIDs, typeID)
		}
		sort.Strings(unitIDs)
		units := make([]unitCountJSON, 0, len(unitIDs))
		for _, typeID := range unitIDs {
			units = append(units, unitCountJSON{TypeID: typeID, Count: counts[typeID]})
		}
		specs = append(specs, armySpecJSON{
			ID:                 string(a.ID),
			OwnerID:            a.OwnerID,
			Region:             a.RegionID,
			DockedRegion:       a.DockedRegionID,
			DockedSettlementID: a.DockedSettlementID,
			IsNaval:            a.IsNaval,
			Units:              units,
		})
	}
	data, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
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
		evts, err := loadScenarioEvents(gs.ScenarioPath)
		if err != nil {
			return loadingResult{err: err, fallback: fallback}
		}
		return loadingResult{
			gs:         gs,
			evts:       evts,
			successMsg: "Oyun yüklendi!",
		}
	})
}

// resetToNewGame state'i temizler ve senaryo seçimine geçer.
func (g *Game) resetToNewGame() {
	difficulty := g.gs.Difficulty
	audio.StopMusic()
	gs := &state.GameState{
		Phase:      state.PhaseScenarioSelect,
		Difficulty: difficulty,
	}
	g.gs = gs
	g.renderer.ReloadGameState(gs)
	g.renderer.SetCursor(0)
}

func (g *Game) startScenarioMusic(scenarioPath string) {
	sc := scenarioByPath(scenarioPath)
	if sc == nil {
		audio.StopMusic()
		return
	}
	playlistName := sc.Music.DefaultPlaylist
	if playlistName == "" {
		playlistName = "campaign"
	}
	defs := sc.Music.Playlists[playlistName]
	if len(defs) == 0 {
		audio.StopMusic()
		return
	}
	tracks := make([]audio.MusicTrack, 0, len(defs))
	for _, def := range defs {
		if def.File == "" {
			continue
		}
		tracks = append(tracks, audio.MusicTrack{
			File:   def.File,
			Weight: def.Weight,
		})
	}
	audio.StartMusicPlaylist(filepath.Join(scenarioPath, "musics"), tracks)
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

	regions, regionOrder, err := world.LoadRegionsWithOrder(dp("regions.json"))
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
	relations, err := faction.LoadRelations(dp("relations.json"), factions)
	if err != nil {
		return nil, nil, fmt.Errorf("ilişkiler yüklenemedi: %w", err)
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
	tradeCenters, err := world.LoadTradeCenters(dp("trade_centers.json"), regions)
	if err != nil {
		log.Printf("Ticaret merkezleri yüklenemedi: %v", err)
	}

	devMode := os.Getenv("DEV_MODE") == "true"
	editMode := os.Getenv("EDIT_MODE") == "true"

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
		EditMode:           editMode,
		ScenarioID:         scenarioIDFromPath(scenarioPath),
		ScenarioPath:       scenarioPath,
		MapConfig:          mapConfig,
		Regions:            regions,
		RegionOrder:        regionOrder,
		Factions:           factions,
		Armies:             armies,
		ShapeData:          shapeData,
		UnitTypes:          unitTypes,
		BuildingTypes:      buildingTypes,
		TechTypes:          techTypes,
		AvailableVictories: victoryOpts,
		Relations:          relations,
		TradeCenters:       tradeCenters,
		NextArmySeq:        len(armies),
		FiredEventIDs:      map[string]bool{},
	}
	army.InitializeLegacyFleetDocking(gs.Armies, gs.Regions)
	gs.SyncTimedRegionUnlocks()
	if editMode {
		gs.Phase = state.PhaseEditMode
	}

	return gs, evts, nil
}

func loadScenarioEvents(scenarioPath string) ([]*events.Event, error) {
	if scenarioPath == "" {
		return nil, fmt.Errorf("senaryo yolu yok")
	}
	evts, err := events.LoadEvents(filepath.Join(scenarioPath, "data", "events.json"))
	if err != nil {
		return nil, fmt.Errorf("olaylar yüklenemedi: %w", err)
	}
	return evts, nil
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
	g.recruitSpecific(rid, "transport", 1)
}

// recruitUnit seçili bölgede oyuncu adına bir milis birimi alır.
func (g *Game) recruitUnit(rid world.RegionID) {
	g.recruitSpecific(rid, "militia", 1)
}

// recruitSpecific seçili bölgede belirli türde bir birim alır.
func (g *Game) recruitSpecific(rid world.RegionID, unitTypeID string, quantity int) {
	if quantity < 1 {
		quantity = 1
	}
	region, ok := g.gs.Regions[rid]
	if !ok || region.IsSea || region.OwnerID != string(g.gs.PlayerFactionID) {
		g.renderer.ShowCombatResult("Sadece kendi bölgende asker alabilirsin!")
		return
	}
	if region.IsLocked {
		g.renderer.ShowCombatResult("Bu bölge kilitli; asker alımı yapılamaz.")
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

	// Deniz birimi — tamamlandığında komşu deniz bölgesine yerleşir.
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
		queued := 0
		for _, a := range g.gs.Armies {
			if a.RegionID == seaRegion && a.OwnerID == string(g.gs.PlayerFactionID) && a.IsNaval {
				queued = len(a.Units)
				break
			}
		}
		queued += g.pendingNavalUnitCount(seaRegion, g.gs.PlayerFactionID)
		if queued >= army.MaxArmySize {
			g.renderer.ShowCombatResult("Filo dolu veya üretim kuyruğuyla dolacak! (max 20 birim)")
			return
		}
			seaFree := army.MaxArmySize - queued
			if quantity > seaFree {
				quantity = seaFree
			}
			pendingInRegion := g.pendingUnitCountByRegion(rid, g.gs.PlayerFactionID)
			if pendingInRegion >= 20 {
				g.renderer.ShowCombatResult("Egitim sirasi dolu! (max 20 emir)")
				return
			}
			queueFree := 20 - pendingInRegion
			if quantity > queueFree {
				quantity = queueFree
			}
			maxByGold := f.Gold / utype.GoldCost
			if maxByGold <= 0 || quantity <= 0 {
				g.renderer.ShowCombatResult(fmt.Sprintf("Yetersiz altın! Gerekli: %d, Mevcut: %d", utype.GoldCost, f.Gold))
				return
			}
		if quantity > maxByGold {
			quantity = maxByGold
		}
		for i := 0; i < quantity; i++ {
			f.Gold -= utype.GoldCost
			g.enqueueProduction(productionKindUnit, rid, unitTypeID, utype.TurnsRequired)
		}
		g.renderer.ShowCombatResult(fmt.Sprintf("%s üretimi başladı! x%d (%d tur) Kalan altın: %d", utype.NameTR, quantity, utype.TurnsRequired, f.Gold))
		return
	}

	// Kara birimi — manpower ve ordu sayısı kontrolü
	pid := g.gs.PlayerFactionID
	deployed := g.gs.DeployedLandUnits(pid) + g.pendingLandUnitCount(pid)
	cap := g.gs.ManpowerCap(pid)
	if deployed >= cap {
		g.renderer.ShowCombatResult(fmt.Sprintf("Savaşçı kapasitesi dolu! (%d/%d) — Bölge fethet veya kışla yap.", deployed, cap))
		return
	}
	availableManpower := cap - deployed
	if quantity > availableManpower {
		quantity = availableManpower
	}

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
		} else {
			if g.gs.CurrentLandArmies(pid) >= g.gs.MaxLandArmies(pid) {
				g.renderer.ShowCombatResult(fmt.Sprintf("Maksimum ordu sayısına ulaşıldı! (%d/%d)", g.gs.CurrentLandArmies(pid), g.gs.MaxLandArmies(pid)))
				return
			}
		}
		pendingInRegion := g.pendingUnitCountByRegion(rid, g.gs.PlayerFactionID)
		if pendingInRegion >= 20 {
			g.renderer.ShowCombatResult("Egitim sirasi dolu! (max 20 emir)")
			return
		}
		queueFree := 20 - pendingInRegion
		if quantity > queueFree {
			quantity = queueFree
		}
		maxByGold := f.Gold / utype.GoldCost
		if maxByGold <= 0 || quantity <= 0 {
			g.renderer.ShowCombatResult(fmt.Sprintf("Yetersiz altın! Gerekli: %d, Mevcut: %d", utype.GoldCost, f.Gold))
			return
		}
	if quantity > maxByGold {
		quantity = maxByGold
	}
	for i := 0; i < quantity; i++ {
		f.Gold -= utype.GoldCost
		g.enqueueProduction(productionKindUnit, rid, unitTypeID, utype.TurnsRequired)
	}
	g.renderer.ShowCombatResult(fmt.Sprintf("%s eğitimi başladı! x%d (%d tur) Kalan altın: %d", utype.NameTR, quantity, utype.TurnsRequired, f.Gold))
}

func (g *Game) cancelRecruitOrder(rid world.RegionID, orderID string) {
	region, ok := g.gs.Regions[rid]
	if !ok || region.IsSea || region.OwnerID != string(g.gs.PlayerFactionID) {
		return
	}
	order, ok := g.cancelProductionByID(orderID, productionKindUnit, rid, g.gs.PlayerFactionID)
	if !ok {
		g.renderer.ShowCombatResult("Iptal edilecek emir bulunamadi.")
		return
	}
	utype, ok := g.gs.UnitTypes[order.TypeID]
	if !ok {
		return
	}
	f, ok := g.gs.Factions[g.gs.PlayerFactionID]
	if !ok {
		return
	}
	refund := utype.GoldCost
	f.Gold += refund
	g.renderer.ShowCombatResult(fmt.Sprintf("%s emri iptal edildi. %d altin iade.", utype.NameTR, refund))
}

func (g *Game) regionUnitProductionCapacity(region *world.Region) int {
	if region == nil || region.IsSea {
		return 0
	}
	capacity := region.Population / 100
	if capacity < 1 {
		capacity = 1
	}
	for _, bid := range region.Buildings {
		if bid == "barracks" {
			capacity++
		}
	}
	return capacity
}

func (g *Game) fleetHasTransportCapacity(fleet *army.Army) bool {
	if fleet == nil || !fleet.IsNaval || len(fleet.EmbarkedUnits) > 0 {
		return false
	}
	for _, u := range fleet.Units {
		if ut, ok := g.gs.UnitTypes[u.TypeID]; ok && ut.Category == army.CategoryNavalTrans {
			return true
		}
	}
	return false
}

func (g *Game) canEmbarkLandArmy(a *army.Army) bool {
	if a == nil || a.IsNaval || len(a.Units) == 0 {
		return false
	}
	for _, u := range a.Units {
		ut, ok := g.gs.UnitTypes[u.TypeID]
		if !ok || !ut.Embarkable {
			return false
		}
	}
	return true
}

func (g *Game) findFriendlyEmbarkFleet(ownerID string, seaRegionID world.RegionID) *army.Army {
	for _, candidate := range g.gs.Armies {
		if candidate.OwnerID != ownerID || !candidate.IsNaval || candidate.RegionID != seaRegionID {
			continue
		}
		if g.fleetHasTransportCapacity(candidate) {
			return candidate
		}
	}
	return nil
}

func (g *Game) disembarkFleet(fleet *army.Army, target world.RegionID) {
	if fleet == nil || !fleet.IsNaval || len(fleet.EmbarkedUnits) == 0 {
		return
	}
	units := make([]army.Unit, len(fleet.EmbarkedUnits))
	copy(units, fleet.EmbarkedUnits)
	fleet.EmbarkedUnits = fleet.EmbarkedUnits[:0]
	g.spawnDisembarkedArmy(fleet.OwnerID, target, units)
}

func (g *Game) spawnDisembarkedArmy(ownerID string, target world.RegionID, units []army.Unit) {
	if len(units) == 0 {
		return
	}
	g.gs.NextArmySeq++
	newID := army.ArmyID(fmt.Sprintf("army_%s_%d", ownerID, g.gs.NextArmySeq))
	g.gs.Armies[newID] = &army.Army{
		ID:            newID,
		OwnerID:       ownerID,
		RegionID:      target,
		Units:         units,
		MovePoints:    0,
		MaxMovePoints: 2,
		IsNaval:       false,
	}
}

func (g *Game) canDisembarkToLand(fleet *army.Army, targetRegion *world.Region) bool {
	if fleet == nil || targetRegion == nil || !fleet.IsNaval || len(fleet.EmbarkedUnits) == 0 {
		return false
	}
	if targetRegion.OwnerID == "" || targetRegion.OwnerID == fleet.OwnerID {
		return true
	}
	key := faction.RelationKey(faction.FactionID(fleet.OwnerID), faction.FactionID(targetRegion.OwnerID))
	rel, ok := g.gs.Relations[key]
	return ok && rel.Stance == faction.StanceWar
}

// applyConquestWithNavalEviction bölge sahipliği değiştiğinde limanda bekleyen
// eski sahip filolarını en yakın deniz bölgesine çıkarır.
func (g *Game) applyConquestWithNavalEviction(targetRegion *world.Region, newOwnerID string) {
	if targetRegion == nil {
		return
	}
	prevOwnerID := targetRegion.OwnerID
	attackerReligion := ownerReligion(g.gs, newOwnerID)
	targetRegion.ApplyConquest(newOwnerID, attackerReligion)
	if prevOwnerID == "" || prevOwnerID == newOwnerID {
		return
	}
	g.evictDockedFleetsFromCapturedPort(targetRegion.ID, prevOwnerID)
}

func (g *Game) evictDockedFleetsFromCapturedPort(capturedRegionID world.RegionID, prevOwnerID string) {
	for _, fleet := range g.gs.Armies {
		if fleet == nil || !fleet.IsNaval || fleet.OwnerID != prevOwnerID || fleet.DockedRegionID != capturedRegionID {
			continue
		}
		if nearestSea := g.nearestSeaRegionForFleet(fleet, capturedRegionID); nearestSea != "" {
			fleet.RegionID = nearestSea
		}
		// Liman artık ele geçirildi: filo burada bağlı kalamaz.
		fleet.DockedRegionID = ""
		fleet.DockedSettlementID = ""
	}
}

func (g *Game) nearestSeaRegionForFleet(fleet *army.Army, capturedRegionID world.RegionID) world.RegionID {
	if fleet != nil {
		if r, ok := g.gs.Regions[fleet.RegionID]; ok && r != nil && r.IsSea {
			return fleet.RegionID
		}
	}
	if r, ok := g.gs.Regions[capturedRegionID]; ok && r != nil {
		for _, nid := range r.Neighbors {
			if n, ok := g.gs.Regions[nid]; ok && n != nil && n.IsSea {
				return n.ID
			}
		}
	}
	return ""
}

// sanitizeDockedFleets limana bağlı donanmaların geçerli sahiplikte olmasını zorunlu tutar.
// Donanma kendi sahip olmadığı limana bağlıysa limandan ayrılır ve en yakın deniz bölgesine çıkar.
func (g *Game) sanitizeDockedFleets() {
	if g == nil || g.gs == nil {
		return
	}
	for _, fleet := range g.gs.Armies {
		if fleet == nil || !fleet.IsNaval || fleet.DockedRegionID == "" {
			continue
		}
		dockedRegion := g.gs.Regions[fleet.DockedRegionID]
		invalidDock := dockedRegion == nil || dockedRegion.IsSea || dockedRegion.OwnerID != fleet.OwnerID
		if !invalidDock {
			continue
		}
		if nearestSea := g.nearestSeaRegionForFleet(fleet, fleet.DockedRegionID); nearestSea != "" {
			fleet.RegionID = nearestSea
		}
		fleet.DockedRegionID = ""
		fleet.DockedSettlementID = ""
	}
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

	// Limana bağlı donanma, bulunduğu deniz bölgesinin merkezine çıkabilir (undock).
	if a.IsNaval && a.DockedRegionID != "" && target == a.RegionID && src.IsSea {
		a.DockedRegionID = ""
		a.DockedSettlementID = ""
		a.MovePoints--
		g.renderer.MarkMapDirty()
		g.renderer.ShowCombatResult("Donanma limandan ayrılıp açık denize çıktı.")
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
	navalSeaMove := a.IsNaval && targetRegion.CanNavalEnter()

	// Naval/kara uyumluluk kontrolü
	if a.IsNaval {
		if targetRegion.CanLandEnter() {
			if !g.canDisembarkToLand(a, targetRegion) {
				if len(a.EmbarkedUnits) == 0 {
					g.renderer.ShowCombatResult("Çıkarma emri reddedildi: filoda taşınan kara birimi yok.")
				} else {
					g.renderer.ShowCombatResult("Çıkarma emri reddedildi: düşman kıyıya çıkarmak için savaş halinde olmalısın.")
				}
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
				landing := &army.Army{
					OwnerID: a.OwnerID,
					Units:   append([]army.Unit(nil), a.EmbarkedUnits...),
				}
				atkMods := techModsFor(g.gs, a.OwnerID)
				defMods := techModsFor(g.gs, enemyArmy.OwnerID)
				result := combat.ResolveBattleWithMods(landing, enemyArmy, targetRegion.Terrain, g.gs.UnitTypes, atkMods, defMods)
				a.EmbarkedUnits = a.EmbarkedUnits[:0]
				a.MovePoints--

				if result.AttackerWins {
					if len(enemyArmy.Units) == 0 {
						delete(g.gs.Armies, enemyArmy.ID)
					}
					g.spawnDisembarkedArmy(a.OwnerID, target, landing.Units)
					g.applyConquestWithNavalEviction(targetRegion, a.OwnerID)
					g.renderer.MarkMapDirty()
					g.renderer.ShowCombatResult(fmt.Sprintf("Çıkarma savaşı kazanıldı (%s): düşman kaybı %d, çıkarma kaybı %d.", result.Description, result.DefenderLost, result.AttackerLost))
					g.renderer.AddEvent(fmt.Sprintf("Amfibi zafer: %s (%d/%d kayıp)", targetRegion.NameTR, result.AttackerLost, result.DefenderLost))
					return
				}

				g.renderer.ShowCombatResult(fmt.Sprintf("Çıkarma savaşı kaybedildi (%s): düşman kaybı %d, çıkarma kaybı %d.", result.Description, result.DefenderLost, result.AttackerLost))
				g.renderer.AddEvent(fmt.Sprintf("Amfibi yenilgi: %s (%d/%d kayıp)", targetRegion.NameTR, result.AttackerLost, result.DefenderLost))
				return
			}
			g.disembarkFleet(a, target)
			a.MovePoints--
			if targetRegion.OwnerID != "" && targetRegion.OwnerID != a.OwnerID {
				g.applyConquestWithNavalEviction(targetRegion, a.OwnerID)
				g.renderer.MarkMapDirty()
				g.renderer.ShowCombatResult("Çıkarma tamamlandı: kıyı bölgesi savaşsız ele geçirildi.")
				g.renderer.AddEvent(fmt.Sprintf("Amfibi fetih: %s", targetRegion.NameTR))
				return
			}
			g.renderer.ShowCombatResult("Çıkarma tamamlandı: birlikler karaya indi.")
			g.renderer.AddEvent(fmt.Sprintf("Birlikler karaya çıktı: %s", targetRegion.NameTR))
			return
		}
		if !targetRegion.CanNavalEnter() {
			g.renderer.ShowCombatResult("Deniz ordusu sadece deniz bölgelerine gidebilir!")
			return
		}
	} else {
		if targetRegion.CanNavalEnter() {
			if !g.canEmbarkLandArmy(a) {
				g.renderer.ShowCombatResult("Bu ordudaki bazı birimler denizden taşınamaz.")
				return
			}
			fleet := g.findFriendlyEmbarkFleet(a.OwnerID, target)
			if fleet == nil {
				g.renderer.ShowCombatResult("Embark için komşu denizde uygun nakliye filosu yok!")
				return
			}
			fleet.EmbarkedUnits = append(fleet.EmbarkedUnits[:0], a.Units...)
			fleet.MovePoints = max(0, fleet.MovePoints-1)
			delete(g.gs.Armies, aid)
			g.renderer.SelectedArmy = fleet.ID
			g.renderer.ShowCombatResult("Ordu nakliye filosuna bindi.")
			return
		}
		if !targetRegion.CanLandEnter() {
			g.renderer.ShowCombatResult("Kara ordusu denize giremez! (Nakliye gemisi gerekir)")
			return
		}
	}
	// Sahipli düşman kara bölgesine girmek için savaş hali zorunlu.
	// Donanma-deniz hareketinde bu kural uygulanmaz; denizde serbest dolaşım var.
	if !navalSeaMove && targetRegion.OwnerID != "" && targetRegion.OwnerID != a.OwnerID {
		key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID))
		rel, exists := g.gs.Relations[key]
		if !exists || rel.Stance != faction.StanceWar {
			g.renderer.ShowCombatResult("Savaş ilan edilmeden düşman topraklarına girilemez!")
			return
		}
	}

	var enemyArmy *army.Army
	for _, ea := range g.gs.Armies {
		if ea.RegionID != target || ea.OwnerID == a.OwnerID {
			continue
		}
		if navalSeaMove {
			key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(ea.OwnerID))
			rel, exists := g.gs.Relations[key]
			if !exists || rel.Stance != faction.StanceWar {
				// Denizde savaş halinde değilsek düşman donanmayla çarpışma tetiklenmez.
				continue
			}
		}
		enemyArmy = ea
		break
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
				a.DockedRegionID = ""
				a.DockedSettlementID = ""
				g.applyConquestWithNavalEviction(targetRegion, a.OwnerID)
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
		a.DockedRegionID = ""
		a.DockedSettlementID = ""
		a.MovePoints--
		if targetRegion.OwnerID != a.OwnerID {
			g.applyConquestWithNavalEviction(targetRegion, a.OwnerID)
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
		ID:                 newID,
		OwnerID:            a.OwnerID,
		RegionID:           a.RegionID,
		DockedRegionID:     a.DockedRegionID,
		DockedSettlementID: a.DockedSettlementID,
		Units:              newUnits,
		MovePoints:         a.MovePoints,
		MaxMovePoints:      a.MaxMovePoints,
		IsNaval:            a.IsNaval,
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
	if !ok || r.OwnerID != string(g.gs.PlayerFactionID) || r.IsLocked {
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

func (g *Game) applyAIDifficultyStartBonus() {
	if g.gs == nil || g.gs.Difficulty < 3 || g.gs.PlayerFactionID == "" {
		return
	}
	for fid, f := range g.gs.Factions {
		if fid == g.gs.PlayerFactionID || f == nil || f.IsEliminated {
			continue
		}
		f.Gold += 300
		f.Grain += 100
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
