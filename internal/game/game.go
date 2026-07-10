package game

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
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
	gs                       *state.GameState
	renderer                 *render.Renderer
	evts                     []*events.Event
	pendingHistoricalEvt     *events.Event
	pendingSortie            *pendingSortieState
	pendingConquestDecisions []pendingConquestDecision
	loading                  *loadingJob
	aiTurn                   *aiTurnState
}

type pendingSortieState struct {
	step           ai.TurnStep
	aiArmy         *army.Army
	siegeArmy      *army.Army
	target         *world.Region
	homeRegion     world.RegionID
	waitFrames     int
	showBattlePlan bool
	focus          int
}

type aiTurnState struct {
	order        []faction.FactionID
	index        int
	stepper      *ai.TurnStepper
	waitFrames   int
	camera       render.CameraState
	cameraLocked bool
}

type eventCodexEntry struct {
	evt          *events.Event
	status       string
	timingReason string
	reasons      []string
	monthsUntil  int
}

const scenarioBaseDir = "assets/scenarios"

const (
	aiTurnVisibleStepFrames  = 34
	aiTurnHiddenStepFrames   = 12
	aiTurnFactionIntroFrames = 10
)

const maxDiplomaticOfferHistoryEntries = 10

type loadingKind int

const (
	loadingScenario loadingKind = iota + 1
	loadingSave
	loadingWorldMap
)

type loadingJob struct {
	kind     loadingKind
	done     chan loadingResult
	run      func() loadingResult
	started  bool
	progress atomic.Int32
}

type loadingResult struct {
	gs           *state.GameState
	evts         []*events.Event
	scenarioPath string
	successMsg   string
	fallback     state.Phase
	worldMap     *render.WorldMap
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
			g.gs.AvailableVictories = scenario.FilterVictoryOptionsForFaction(g.gs.ScenarioVictories, string(action.TargetFaction))
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
			g.startPreparePlayerTurn()
		case render.ActionBack:
			g.gs.Phase = state.PhaseFactionSelect
			g.renderer.SetCursor(0)
		}

	case state.PhasePlayerTurn:
		switch action.Kind {
		case render.ActionOpenEventCodex:
			// renderer event detail popup'ını kendi açıyor; burada state değişikliği gerekmiyor.
		case render.ActionChooseHistoricalEvent:
			g.resolveHistoricalChoice(action.ChoiceIndex)
		case render.ActionScheduleCapitalMove:
			g.queueCapitalMove(g.gs.PlayerFactionID, action.BuildingID, state.DefaultCapitalMoveTurns, "yerel karar")
		case render.ActionEndTurn:
			// Araştırma seçimi artık turn resolution içinde otomatik tamamlanıyor.
			if !g.saveToSlot("autosave", false, "") {
				break
			}
			g.startAITurnSequence()
		case render.ActionConfirmEndTurn:
			if !g.saveToSlot("autosave", false, "") {
				break
			}
			g.startAITurnSequence()
		case render.ActionMoveArmy:
			g.moveArmyWithStance(action.ArmyID, action.TargetRegion, action.BattleStance)
		case render.ActionEmbarkArmy:
			g.embarkArmyOntoFleet(action.ArmyID, action.TargetArmyID)
		case render.ActionDisembarkArmy:
			g.forceDisembarkFleetWithStance(action.ArmyID, action.TargetRegion, action.BattleStance)
		case render.ActionStartSiege:
			g.startSiege(action.ArmyID, action.TargetRegion)
		case render.ActionAssaultSiege:
			g.assaultSiegeWithStance(action.ArmyID, action.TargetRegion, action.BattleStance)
		case render.ActionLiftSiege:
			g.liftSiege(action.ArmyID, action.TargetRegion)
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
		case render.ActionCancelBuilding:
			g.cancelBuilding(action.TargetRegion, action.BuildingID)
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
			g.moveArmyWithStance(action.ArmyID, action.TargetRegion, action.BattleStance)
		case render.ActionProposePeace:
			g.proposePeace(action.TargetFaction)
		case render.ActionImproveRelations:
			g.improveRelations(action.TargetFaction)
		case render.ActionSendGift:
			g.sendGift(action.TargetFaction)
		case render.ActionProposeAlliance:
			g.proposeAlliance(action.TargetFaction)
		case render.ActionProposeTrade:
			g.proposeTrade(action.TargetFaction)
		case render.ActionCancelAlliance:
			g.cancelAlliance(action.TargetFaction)
		case render.ActionCancelTrade:
			g.cancelTrade(action.TargetFaction)
		case render.ActionOfferVassalization:
			g.offerVassalization(action.TargetFaction)
		case render.ActionReleaseVassal:
			g.releaseVassal(action.TargetFaction)
		case render.ActionAnnexVassal:
			g.annexVassal(action.TargetFaction)
		case render.ActionAnnexDefeatedFaction:
			g.resolvePendingConquestDecision(false)
		case render.ActionVassalizeDefeatedFaction:
			g.resolvePendingConquestDecision(true)
		case render.ActionRespondDiplomacyOffer:
			g.respondDiplomacyOffer(action.OfferIndex, action.OfferAccepted)
		case render.ActionCreateTradeRoute:
			g.proposeTrade(action.TargetFaction)
		case render.ActionOneTimeTrade:
			g.oneTimeTrade(action.TargetFaction, action.BuildingID, action.Delta)
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
		if action.Kind == render.ActionRespondDiplomacyOffer {
			g.handleAITurnOfferResponse(action.OfferIndex, action.OfferAccepted)
		}
		g.updateAITurnSequence()

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

	if action.Kind != render.ActionNone {
		g.refreshEventCodex()
	}

	return nil
}

func (g *Game) startLoading(kind loadingKind, message string, fn func(func(int)) loadingResult) {
	g.gs.Phase = state.PhaseLoading
	g.renderer.SetLoadingMessage(message)
	g.renderer.SetLoadingProgress(0)
	done := make(chan loadingResult, 1)
	g.loading = &loadingJob{
		kind: kind,
		done: done,
		run: func() loadingResult {
			return fn(func(progress int) {
				if progress < 0 {
					progress = 0
				}
				if progress > 100 {
					progress = 100
				}
				g.loading.progress.Store(int32(progress))
			})
		},
	}
}

func (g *Game) pollLoading() {
	if g.loading != nil && !g.loading.started {
		job := g.loading
		job.started = true
		go func() {
			job.done <- job.run()
		}()
		return
	}
	g.renderer.SetLoadingProgress(int(g.loading.progress.Load()))
	select {
	case res := <-g.loading.done:
		kind := g.loading.kind
		g.renderer.SetLoadingProgress(100)
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
		g.pendingConquestDecisions = nil
		g.sanitizeOccupiedNeutralRegions()
		g.sanitizeDockedFleets()
		g.evts = res.evts
		g.renderer.ReloadGameState(res.gs)
		g.startScenarioMusic(res.gs.ScenarioPath)
		g.renderer.SetCursor(0)
		g.refreshEventCodex()
	case loadingSave:
		res.gs.Phase = state.PhasePlayerTurn
		g.gs = res.gs
		g.pendingConquestDecisions = nil
		g.sanitizeOccupiedNeutralRegions()
		g.sanitizeDockedFleets()
		g.evts = res.evts
		g.renderer.ReloadGameStateWithPreparedMap(res.gs, res.worldMap)
		g.startScenarioMusic(res.gs.ScenarioPath)
		g.renderer.HasSave = save.AnySlotExists()
		g.renderer.HasAutoSave = save.SaveExists()
		g.renderer.ShowCombatResult(res.successMsg)
		g.refreshEventCodex()
	case loadingWorldMap:
		g.gs.Phase = state.PhasePlayerTurn
		g.renderer.ReloadGameStateWithPreparedMap(g.gs, res.worldMap)
		g.refreshEventCodex()
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

func (g *Game) startAITurnSequence() {
	if g == nil || g.gs == nil || g.renderer == nil {
		return
	}
	camera := g.renderer.CameraSnapshot()
	g.renderer.PrepareForTurnAdvance()
	g.aiTurn = &aiTurnState{
		order:        g.orderedAIFactions(),
		camera:       camera,
		cameraLocked: true,
	}
	g.gs.Phase = state.PhaseAITurn
	g.renderer.ClearAITurnStatus()
}

func (g *Game) orderedAIFactions() []faction.FactionID {
	if g == nil || g.gs == nil {
		return nil
	}
	order := make([]faction.FactionID, 0, len(g.gs.Factions))
	seen := make(map[faction.FactionID]struct{}, len(g.gs.Factions))
	for _, fid := range g.gs.FactionOrder {
		if fid == g.gs.PlayerFactionID || g.gs.Factions[fid] == nil {
			continue
		}
		order = append(order, fid)
		seen[fid] = struct{}{}
	}
	extra := make([]faction.FactionID, 0, len(g.gs.Factions))
	for fid := range g.gs.Factions {
		if fid == g.gs.PlayerFactionID {
			continue
		}
		if _, ok := seen[fid]; ok {
			continue
		}
		extra = append(extra, fid)
	}
	sort.Slice(extra, func(i, j int) bool { return extra[i] < extra[j] })
	return append(order, extra...)
}

func (g *Game) updateAITurnSequence() {
	if g == nil || g.gs == nil || g.renderer == nil {
		return
	}
	if g.aiTurn == nil {
		g.startAITurnSequence()
	}
	if g.aiTurn == nil {
		return
	}
	if offer, waiting := g.pendingPlayerDiplomacyOffer(); waiting {
		g.renderer.SetAITurnStatus(turnActorName(g.gs, offer.FromFactionID), "Diplomasi cevabınız bekleniyor.")
		return
	}
	if g.aiTurn.waitFrames > 0 {
		g.aiTurn.waitFrames--
		return
	}

	for {
		if g.aiTurn.index >= len(g.aiTurn.order) {
			g.finishAITurnSequence()
			g.renderer.MarkMapDirty()
			g.gs.Phase = state.PhaseTurnResolution
			return
		}

		if g.aiTurn.stepper == nil {
			fid := g.aiTurn.order[g.aiTurn.index]
			f := g.gs.Factions[fid]
			if f == nil || f.IsEliminated {
				g.aiTurn.index++
				continue
			}
			g.aiTurn.stepper = ai.NewTurnStepper(g.gs, fid)
			g.renderer.SetAITurnStatus(g.aiTurn.stepper.FactionNameTR(), "Hamle sırası bu devlette.")
			g.aiTurn.waitFrames = aiTurnFactionIntroFrames
			return
		}

		step, done := g.aiTurn.stepper.Step()
		if done {
			g.aiTurn.index++
			g.aiTurn.stepper = nil
			continue
		}
		g.handleAITurnStep(step)
		return
	}
}

func (g *Game) handleAITurnStep(step ai.TurnStep) {
	if g == nil || g.aiTurn == nil {
		return
	}
	// Araştırma ve bina inşaatı adımlarını göstermeden, beklemeden geç.
	if step.Kind == ai.TurnStepResearch || step.Kind == ai.TurnStepBuild {
		g.aiTurn.waitFrames = 0
		return
	}
	actor := turnActorName(g.gs, step.FactionID)
	detail := step.Message
	if detail == "" {
		detail = "Hamle işleniyor."
	}
	nearPlayer := g.aiStepVisible(step)
	if nearPlayer && step.FocusRegion != "" {
		g.renderer.CenterCameraOnRegion(step.FocusRegion)
	}
	g.renderer.SetAITurnStatus(actor, aiTurnOverlayDetail(step, nearPlayer))
	if shouldLogAITurnStep(step) {
		g.renderer.AddEvent("[AI] " + detail)
	}
	if nearPlayer || step.Kind == ai.TurnStepDiplomacy {
		g.renderer.ShowCombatResult(detail)
		g.aiTurn.waitFrames = aiTurnVisibleStepFrames
		return
	}
	g.aiTurn.waitFrames = aiTurnHiddenStepFrames
}

func (g *Game) finishAITurnSequence() {
	if g == nil || g.renderer == nil {
		return
	}
	if g.aiTurn != nil && g.aiTurn.cameraLocked {
		g.renderer.RestoreCamera(g.aiTurn.camera)
	}
	g.renderer.ClearAITurnStatus()
	g.aiTurn = nil
}

func (g *Game) aiStepVisible(step ai.TurnStep) bool {
	if g == nil || g.gs == nil {
		return false
	}
	if step.Kind == ai.TurnStepDiplomacy {
		return step.TargetFaction == g.gs.PlayerFactionID
	}
	if step.FocusRegion == "" {
		return false
	}
	return g.regionNearPlayer(step.FocusRegion, 3)
}

func (g *Game) regionNearPlayer(target world.RegionID, maxDepth int) bool {
	if g == nil || g.gs == nil || target == "" || maxDepth < 0 {
		return false
	}
	type queueItem struct {
		id    world.RegionID
		depth int
	}
	seen := make(map[world.RegionID]struct{}, len(g.gs.Regions))
	queue := make([]queueItem, 0, len(g.gs.Regions))
	for _, region := range g.gs.Regions {
		if region == nil || region.OwnerID != string(g.gs.PlayerFactionID) {
			continue
		}
		seen[region.ID] = struct{}{}
		queue = append(queue, queueItem{id: region.ID, depth: 0})
	}
	for _, playerArmy := range g.gs.Armies {
		if playerArmy == nil || playerArmy.OwnerID != string(g.gs.PlayerFactionID) {
			continue
		}
		if _, ok := seen[playerArmy.RegionID]; ok {
			continue
		}
		seen[playerArmy.RegionID] = struct{}{}
		queue = append(queue, queueItem{id: playerArmy.RegionID, depth: 0})
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.id == target {
			return true
		}
		if cur.depth >= maxDepth {
			continue
		}
		region := g.gs.Regions[cur.id]
		if region == nil {
			continue
		}
		for _, nxt := range region.Neighbors {
			if _, ok := seen[nxt]; ok || g.gs.Regions[nxt] == nil {
				continue
			}
			seen[nxt] = struct{}{}
			queue = append(queue, queueItem{id: nxt, depth: cur.depth + 1})
		}
	}
	return false
}

func turnActorName(gs *state.GameState, fid faction.FactionID) string {
	if gs == nil {
		return string(fid)
	}
	if f := gs.Factions[fid]; f != nil && f.NameTR != "" {
		return f.NameTR
	}
	return string(fid)
}

func aiTurnOverlayDetail(step ai.TurnStep, nearPlayer bool) string {
	if nearPlayer {
		return step.Message
	}
	switch step.Kind {
	case ai.TurnStepMove, ai.TurnStepBattle, ai.TurnStepEmbark, ai.TurnStepDisembark, ai.TurnStepConquest:
		return "Uzak cephede hamleler çözülüyor."
	case ai.TurnStepRecruit, ai.TurnStepBuild, ai.TurnStepResearch:
		return "Arka plan üretim ve hazırlıkları tamamlanıyor."
	case ai.TurnStepDiplomacy:
		return step.Message
	default:
		return "Hamle işleniyor."
	}
}

func shouldLogAITurnStep(step ai.TurnStep) bool {
	switch step.Kind {
	case ai.TurnStepDiplomacy, ai.TurnStepBattle, ai.TurnStepConquest:
		return true
	default:
		return false
	}
}

func (g *Game) currentAITurnFactionID() (faction.FactionID, bool) {
	if g == nil || g.aiTurn == nil {
		return "", false
	}
	if g.aiTurn.stepper != nil {
		return g.aiTurn.stepper.FactionID(), true
	}
	if g.aiTurn.index >= 0 && g.aiTurn.index < len(g.aiTurn.order) {
		return g.aiTurn.order[g.aiTurn.index], true
	}
	return "", false
}

func (g *Game) pendingPlayerDiplomacyOffer() (state.DiplomaticOffer, bool) {
	if g == nil || g.gs == nil || len(g.gs.DiplomaticOffers) == 0 {
		return state.DiplomaticOffer{}, false
	}
	if offerIdx, ok := diplomacy.BestOfferIndex(g.gs, g.gs.PlayerFactionID); ok {
		return g.gs.DiplomaticOffers[offerIdx], true
	}
	return state.DiplomaticOffer{}, false
}

func (g *Game) resolveTurn() {
	g.sanitizeOccupiedNeutralRegions()
	g.sanitizeDockedFleets()
	applySeasonEffects(g.gs)
	economyReport := applyEconomyTick(g.gs)
	navalVoyageAlerts := applyEmbarkedVoyageAttrition(g.gs)
	completedTechs := applyTechTicks(g.gs)
	productionResults := g.applyProductionTicks()
	applyReligionConversion(g.gs)
	siegeUpdates := g.resolveSieges()
	checkRebellions(g.gs)
	checkEliminations(g.gs)
	capitalMoveUpdates := g.gs.AdvanceCapitalMoves()
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

	for _, su := range siegeUpdates {
		if su.Message == "" {
			continue
		}
		if su.Popup {
			g.renderer.ShowCombatResult(su.Message)
		}
		if su.Detail != "" {
			g.renderer.AddEventDetail("[KUSATMA] "+su.Message, su.Detail)
		} else {
			g.renderer.AddEvent("[KUSATMA] " + su.Message)
		}
	}

	g.showRegionalLogisticsAlerts(economyReport.PlayerLogisticsAlerts)
	g.showEmbarkedVoyageAlerts(navalVoyageAlerts)
	g.handleCapitalMoveProgress(capitalMoveUpdates)

	// Olaylar
	if evt := events.Tick(g.gs, g.evts); evt != nil {
		events.Apply(g.gs, evt)
		g.handleTriggeredEvent(evt)
	}

	// Bölge event ikon sürelerini güncelle
	events.TickActiveRegionEvents(g.gs)

	g.gs.AdvanceTurn()
	unlocked := checkRegionUnlocks(g.gs)
	g.showRegionUnlockNotifications(unlocked)
	if g.gs.Phase != state.PhaseGameOver {
		// Aktif araştırma yoksa oyuncu için uygun sonraki teknoloji otomatik başlatılır.
		g.autoStartResearchIfIdle()
		g.gs.Phase = state.PhasePlayerTurn
		g.renderer.MarkMapDirty()
	}
	g.refreshEventCodex()
}

func (g *Game) autoStartResearchIfIdle() bool {
	if g == nil || g.gs == nil || g.gs.TechTypes == nil {
		return false
	}
	f := g.gs.Factions[g.gs.PlayerFactionID]
	if f == nil || f.Research.ActiveID != "" || len(f.Research.PausedTurns) > 0 {
		return false
	}
	techID, ok := tech.NextResearchableTechID(&f.Research, g.gs.TechTypes, f.Gold)
	if !ok {
		return false
	}
	g.startResearch(techID)
	return true
}

func (g *Game) showRegionalLogisticsAlerts(alerts []state.RegionLogisticsStatus) {
	for _, alert := range alerts {
		regionName := string(alert.RegionID)
		if region, ok := g.gs.Regions[alert.RegionID]; ok && region != nil {
			regionName = region.NameTR
		}
		msg := fmt.Sprintf(
			"%s: ikmal kapasitesi %d/%d, %d ordu zayiat verdi",
			regionName, alert.Capacity, alert.Demand, alert.ArmyCount,
		)
		detail := fmt.Sprintf(
			"%s bölgesinde yerel tahıl %d, depo/yerleşim tamponu %d, stok desteği %d kaldı. Aşım: %d. Etkilenen birlik: %d, kayıp birlik: %d, toplam HP kaybı: %d.",
			regionName,
			alert.LocalProduction,
			alert.SettlementBuffer,
			alert.ReserveSupport,
			alert.Overload,
			alert.UnitsAffected,
			alert.UnitsLost,
			alert.TotalHPDamage,
		)
		g.renderer.AddEventDetail("[LOJISTIK] "+msg, detail)
		if alert.PeakOverloadTurns >= 2 {
			g.renderer.ShowCombatResult(regionName + ": ordular ikmal baskısı altında zayiat veriyor")
		}
	}
}

func (g *Game) showEmbarkedVoyageAlerts(alerts []navalVoyageAlert) {
	for _, alert := range alerts {
		fleet := g.gs.Armies[alert.FleetID]
		if fleet == nil || fleet.OwnerID != string(g.gs.PlayerFactionID) {
			continue
		}
		regionName := string(alert.RegionID)
		if region, ok := g.gs.Regions[alert.RegionID]; ok && region != nil {
			regionName = region.NameTR
		}
		msg := fmt.Sprintf("%s: limansız uzun sefer gemideki birlikleri yıpratıyor", regionName)
		detail := fmt.Sprintf(
			"%s bölgesindeki %s filosu %d turdur limana uğramadı. Gemideki birlikler birim başına %d HP kaybetti. Etkilenen birlik: %d, kayıp birlik: %d, toplam HP kaybı: %d.",
			regionName,
			string(alert.FleetID),
			alert.TurnsWithoutPort,
			alert.DamagePerUnit,
			alert.UnitsAffected,
			alert.UnitsLost,
			alert.TotalHPDamage,
		)
		g.renderer.AddEventDetail("[DENIZ LOJISTIK] "+msg, detail)
		g.renderer.ShowCombatResult(msg)
	}
}

func (g *Game) handleTriggeredEvent(evt *events.Event) {
	if evt == nil {
		return
	}
	baseMsg := "OLAY: " + evt.NameTR + ": " + evt.DescTR
	g.renderer.ShowCombatResult(baseMsg)
	g.renderer.AddEventDetail("[OLAY] "+evt.NameTR, g.historicalEventDetail(evt))
	if !events.RequiresPlayerChoice(g.gs, evt) {
		if idx := events.AutoChoose(evt); idx >= 0 {
			g.applyHistoricalChoice(evt, idx)
		}
		if evt.HistoricalYear != 0 {
			g.renderer.ShowHistoricalEvent(evt.NameTR, evt.DescTR, "", nil)
		}
		return
	}
	g.pendingHistoricalEvt = evt
	g.renderer.ShowHistoricalEvent(evt.NameTR, evt.DescTR, evt.ChoicePromptTR, g.historicalChoiceViews(evt))
}

func (g *Game) resolveHistoricalChoice(idx int) {
	if g.pendingHistoricalEvt == nil {
		g.renderer.HideHistoricalEvent()
		return
	}
	g.applyHistoricalChoice(g.pendingHistoricalEvt, idx)
	g.pendingHistoricalEvt = nil
	g.renderer.HideHistoricalEvent()
}

func (g *Game) applyHistoricalChoice(evt *events.Event, idx int) {
	choice, ok := events.ApplyChoice(g.gs, evt, idx)
	if !ok {
		return
	}
	msg := fmt.Sprintf("Karar: %s -> %s", evt.NameTR, choice.LabelTR)
	g.renderer.ShowCombatResult(msg)
	g.renderer.AddEventDetail("[KARAR] "+evt.NameTR+": "+choice.LabelTR, g.historicalChoiceDetail(evt, choice))
}

func (g *Game) historicalChoiceViews(evt *events.Event) []render.HistoricalEventChoice {
	if evt == nil || len(evt.Choices) == 0 {
		return nil
	}
	views := make([]render.HistoricalEventChoice, 0, len(evt.Choices))
	for _, choice := range evt.Choices {
		followUp, conditions := g.historicalChoiceFollowUpSummary(evt, choice)
		views = append(views, render.HistoricalEventChoice{
			Label:      choice.LabelTR,
			Desc:       choice.DescTR,
			Effect:     historicalChoiceEffectSummary(g.gs, choice.Effect),
			FollowUp:   followUp,
			Conditions: conditions,
		})
	}
	return views
}

func (g *Game) refreshEventCodex() {
	if g == nil || g.renderer == nil {
		return
	}
	g.renderer.SetEventCodexEntries(g.buildEventCodexPages())
}

func (g *Game) buildEventCodexPages() [4][]render.EventCodexEntry {
	return [4][]render.EventCodexEntry{
		g.buildEventCodexFor("all"),
		g.buildEventCodexFor("ready"),
		g.buildEventCodexFor("calendar"),
		g.buildEventCodexFor("locked"),
	}
}

func (g *Game) buildEventCodexFor(filter string) []render.EventCodexEntry {
	if g == nil || g.gs == nil || len(g.evts) == 0 {
		return nil
	}
	entries := g.collectEventCodexEntries(filter)
	views := make([]render.EventCodexEntry, 0, len(entries))
	for _, entry := range entries {
		evt := entry.evt
		dateLabel := fmt.Sprintf("%d", evt.HistoricalYear)
		if evt.HistoricalMonth > 0 {
			dateLabel = fmt.Sprintf("%d/%02d", evt.HistoricalYear, evt.HistoricalMonth)
		}
		detail := make([]string, 0, 8)
		if evt.DescTR != "" {
			detail = append(detail, evt.DescTR)
		}
		if entry.monthsUntil > 0 {
			detail = append(detail, fmt.Sprintf("Kalan süre: %d ay", entry.monthsUntil))
		}
		if len(entry.reasons) > 0 {
			detail = append(detail, "Kritik eksik: "+g.codexReasonLabel(entry.reasons[0]))
			for _, reason := range entry.reasons {
				detail = append(detail, "Neden: "+g.codexReasonLabel(reason))
			}
		} else if entry.timingReason != "" {
			detail = append(detail, "Neden: "+entry.timingReason)
		} else {
			detail = append(detail, "Koşullar sağlanıyor.")
		}
		views = append(views, render.EventCodexEntry{
			Title:       evt.NameTR,
			Status:      entry.status,
			DateLabel:   dateLabel,
			Summary:     evt.DescTR,
			Detail:      strings.Join(detail, "\n"),
			MonthsUntil: entry.monthsUntil,
		})
		if len(views) >= 12 {
			break
		}
	}
	return views
}

func (g *Game) collectEventCodexEntries(filter string) []eventCodexEntry {
	entries := make([]eventCodexEntry, 0, len(g.evts))
	for _, evt := range g.evts {
		if evt == nil || evt.HistoricalYear == 0 || g.gs.FiredEventIDs[evt.ID] {
			continue
		}
		if !g.eventRelevantToPlayer(evt) {
			continue
		}
		if evt.HistoricalYear < g.gs.Year || (evt.HistoricalYear == g.gs.Year && evt.HistoricalMonth != 0 && evt.HistoricalMonth < g.gs.Month) {
			continue
		}
		entry := eventCodexEntry{
			evt:         evt,
			status:      "Hazir",
			monthsUntil: monthsUntilHistoricalEvent(g.gs, evt),
			reasons:     events.ConditionFailureReasons(g.gs, evt),
		}
		if entry.monthsUntil > 0 {
			entry.status = "Takvim"
			entry.timingReason = "takvim bekleniyor"
		}
		if len(entry.reasons) > 0 {
			entry.status = "Kilitli"
		}
		if filter == "ready" && entry.status != "Hazir" {
			continue
		}
		if filter == "calendar" && entry.status != "Takvim" {
			continue
		}
		if filter == "locked" && entry.status != "Kilitli" {
			continue
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if ar, br := eventCodexStatusRank(a.status), eventCodexStatusRank(b.status); ar != br {
			return ar < br
		}
		if a.monthsUntil != b.monthsUntil {
			return a.monthsUntil < b.monthsUntil
		}
		if len(a.reasons) != len(b.reasons) {
			return len(a.reasons) < len(b.reasons)
		}
		if a.evt.HistoricalYear != b.evt.HistoricalYear {
			return a.evt.HistoricalYear < b.evt.HistoricalYear
		}
		if a.evt.HistoricalMonth != b.evt.HistoricalMonth {
			return a.evt.HistoricalMonth < b.evt.HistoricalMonth
		}
		return a.evt.NameTR < b.evt.NameTR
	})
	return entries
}

func eventCodexStatusRank(status string) int {
	switch status {
	case "Hazir":
		return 0
	case "Takvim":
		return 1
	case "Kilitli":
		return 2
	default:
		return 3
	}
}

func monthsUntilHistoricalEvent(gs *state.GameState, evt *events.Event) int {
	if gs == nil || evt == nil || evt.HistoricalYear == 0 {
		return 0
	}
	targetMonth := evt.HistoricalMonth
	if targetMonth <= 0 {
		targetMonth = 1
	}
	currentMonth := gs.Month
	if currentMonth <= 0 {
		currentMonth = 1
	}
	currentAbs := gs.Year*12 + (currentMonth - 1)
	targetAbs := evt.HistoricalYear*12 + (targetMonth - 1)
	if targetAbs <= currentAbs {
		return 0
	}
	return targetAbs - currentAbs
}

func (g *Game) eventRelevantToPlayer(evt *events.Event) bool {
	if g == nil || g.gs == nil || evt == nil {
		return false
	}
	switch evt.Target {
	case "player_faction", "all_factions", "all_armies":
		return true
	case "specific_faction":
		return evt.AffectedFaction == string(g.gs.PlayerFactionID)
	default:
		return false
	}
}

func (g *Game) codexReasonLabel(reason string) string {
	if reason == "" {
		return ""
	}
	parts := strings.SplitN(reason, ": ", 2)
	if len(parts) != 2 {
		return reason
	}
	key, value := parts[0], parts[1]
	switch key {
	case "gerekli tech", "zaten acik tech":
		value = strings.Join(techLabels(g.gs, []string{value}), ", ")
	case "bolge gerekli":
		value = strings.Join(regionLabels(g.gs, []world.RegionID{world.RegionID(value)}), ", ")
	}
	return key + ": " + value
}

func (g *Game) historicalEventDetail(evt *events.Event) string {
	if evt == nil {
		return ""
	}
	lines := []string{evt.NameTR, "", "Kaynak: Olay kaydı", "", evt.DescTR}
	if evt.ChoicePromptTR != "" {
		lines = append(lines, "", "Seçim:", evt.ChoicePromptTR)
	}
	if len(evt.Choices) > 0 {
		for _, choice := range evt.Choices {
			followUp, conditions := g.historicalChoiceFollowUpSummary(evt, choice)
			lines = append(lines, "", "- "+choice.LabelTR)
			if choice.DescTR != "" {
				lines = append(lines, "  "+choice.DescTR)
			}
			if eff := historicalChoiceEffectSummary(g.gs, choice.Effect); eff != "" {
				lines = append(lines, "  Etki: "+eff)
			}
			if followUp != "" {
				lines = append(lines, "  "+followUp)
			}
			if conditions != "" {
				lines = append(lines, "  Kosul: "+conditions)
			}
		}
	}
	return strings.Join(lines, "\n")
}

func (g *Game) historicalChoiceDetail(evt *events.Event, choice events.Choice) string {
	if evt == nil {
		return ""
	}
	followUp, conditions := g.historicalChoiceFollowUpSummary(evt, choice)
	lines := []string{
		evt.NameTR + " -> " + choice.LabelTR,
		"",
		"Kaynak: Karar kaydı",
		"",
		choice.DescTR,
	}
	if eff := historicalChoiceEffectSummary(g.gs, choice.Effect); eff != "" {
		lines = append(lines, "", "Etki:", eff)
	}
	if followUp != "" {
		lines = append(lines, "", followUp)
	}
	if conditions != "" {
		lines = append(lines, "Kosul: "+conditions)
	}
	return strings.Join(lines, "\n")
}

func (g *Game) historicalChoiceFollowUpSummary(evt *events.Event, choice events.Choice) (string, string) {
	if g == nil || evt == nil {
		return "", ""
	}
	eff := choice.Effect
	if eff.Target == "" {
		eff.Target = evt.Target
	}
	if eff.AffectedFaction == "" {
		eff.AffectedFaction = evt.AffectedFaction
	}
	flagSet := make(map[string]bool, len(eff.SetFlags))
	for _, flag := range eff.SetFlags {
		if flag != "" {
			flagSet[flag] = true
		}
	}
	if len(flagSet) == 0 {
		return "", ""
	}
	candidates := make([]*events.Event, 0, 2)
	for _, candidate := range g.evts {
		if candidate == nil || candidate.ID == evt.ID || len(candidate.RequiresFlags) == 0 {
			continue
		}
		if !sameEventBranch(evt, candidate) {
			continue
		}
		for _, req := range candidate.RequiresFlags {
			if flagSet[req] {
				candidates = append(candidates, candidate)
				break
			}
		}
	}
	if len(candidates) == 0 {
		return "", ""
	}
	labels := make([]string, 0, len(candidates))
	conds := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		labels = append(labels, historicalFollowUpLabel(candidate))
		if cond := historicalFollowUpConditions(g.gs, candidate); cond != "" {
			conds = append(conds, cond)
		}
	}
	return "Zincir: " + strings.Join(labels, " | "), strings.Join(conds, " | ")
}

func sameEventBranch(source, candidate *events.Event) bool {
	if source == nil || candidate == nil {
		return false
	}
	if source.Target != candidate.Target {
		return false
	}
	switch source.Target {
	case "specific_faction":
		return source.AffectedFaction == candidate.AffectedFaction
	default:
		return true
	}
}

func historicalFollowUpLabel(evt *events.Event) string {
	if evt == nil {
		return ""
	}
	if evt.HistoricalMonth > 0 {
		return fmt.Sprintf("%s (%d/%02d)", evt.NameTR, evt.HistoricalYear, evt.HistoricalMonth)
	}
	if evt.HistoricalYear > 0 {
		return fmt.Sprintf("%s (%d)", evt.NameTR, evt.HistoricalYear)
	}
	return evt.NameTR
}

func historicalFollowUpConditions(gs *state.GameState, evt *events.Event) string {
	if evt == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	if len(evt.RequiresOwnedRegions) > 0 {
		parts = append(parts, "Bolge: "+strings.Join(regionLabels(gs, evt.RequiresOwnedRegions), ", "))
	}
	if len(evt.RequiresTechs) > 0 {
		parts = append(parts, "Gerekli tech: "+strings.Join(techLabels(gs, evt.RequiresTechs), ", "))
	}
	if len(evt.BlocksTechs) > 0 {
		parts = append(parts, "Kapali tech: "+strings.Join(techLabels(gs, evt.BlocksTechs), ", "))
	}
	if len(evt.RelationRequirements) > 0 {
		parts = append(parts, "Diplomasi: "+strings.Join(relationRequirementLabels(gs, evt.RelationRequirements), "; "))
	}
	return strings.Join(parts, "  ")
}

func regionLabels(gs *state.GameState, ids []world.RegionID) []string {
	labels := make([]string, 0, len(ids))
	for _, rid := range ids {
		label := string(rid)
		if gs != nil && gs.Regions != nil {
			if r := gs.Regions[rid]; r != nil && r.NameTR != "" {
				label = r.NameTR
			}
		}
		labels = append(labels, label)
	}
	return labels
}

func techLabels(gs *state.GameState, ids []string) []string {
	labels := make([]string, 0, len(ids))
	for _, id := range ids {
		label := id
		if gs != nil && gs.TechTypes != nil {
			if t := gs.TechTypes[id]; t != nil && t.NameTR != "" {
				label = t.NameTR
			}
		}
		labels = append(labels, label)
	}
	return labels
}

func relationRequirementLabels(gs *state.GameState, reqs []events.RelationRequirement) []string {
	labels := make([]string, 0, len(reqs))
	for _, req := range reqs {
		name := req.FactionID
		if gs != nil && gs.Factions != nil {
			if f := gs.Factions[faction.FactionID(req.FactionID)]; f != nil && f.NameTR != "" {
				name = f.NameTR
			}
		}
		parts := make([]string, 0, 4)
		if req.Stance != "" {
			parts = append(parts, "durus="+stanceLabel(faction.DiplomaticStance(req.Stance)))
		}
		if len(req.AnyOfStances) > 0 {
			stanceNames := make([]string, 0, len(req.AnyOfStances))
			for _, stance := range req.AnyOfStances {
				stanceNames = append(stanceNames, stanceLabel(faction.DiplomaticStance(stance)))
			}
			parts = append(parts, "durus "+strings.Join(stanceNames, "/"))
		}
		if len(req.BlocksStances) > 0 {
			stanceNames := make([]string, 0, len(req.BlocksStances))
			for _, stance := range req.BlocksStances {
				stanceNames = append(stanceNames, stanceLabel(faction.DiplomaticStance(stance)))
			}
			parts = append(parts, "savasma: "+strings.Join(stanceNames, "/"))
		}
		if req.MinScore != 0 {
			parts = append(parts, fmt.Sprintf("skor>=%d", req.MinScore))
		}
		if req.MaxScore != 0 {
			parts = append(parts, fmt.Sprintf("skor<=%d", req.MaxScore))
		}
		if len(parts) == 0 {
			continue
		}
		labels = append(labels, name+" "+strings.Join(parts, ", "))
	}
	return labels
}

func stanceLabel(stance faction.DiplomaticStance) string {
	switch stance {
	case faction.StanceWar:
		return "savas"
	case faction.StanceTrade:
		return "ticaret"
	case faction.StanceAllied:
		return "ittifak"
	case faction.StancePeace:
		return "baris"
	default:
		return string(stance)
	}
}

func historicalChoiceEffectSummary(gs *state.GameState, eff events.Effect) string {
	parts := make([]string, 0, 8)
	if eff.GoldDelta != 0 {
		parts = append(parts, fmt.Sprintf("Altın %+d", eff.GoldDelta))
	}
	if eff.GrainDelta != 0 {
		parts = append(parts, fmt.Sprintf("Tahıl %+d", eff.GrainDelta))
	}
	if eff.SatDelta != 0 {
		parts = append(parts, fmt.Sprintf("Memnuniyet %+d", eff.SatDelta))
	}
	if eff.RelationDeltaAll != 0 {
		parts = append(parts, fmt.Sprintf("Diplomasi %+d", eff.RelationDeltaAll))
	}
	if eff.ArmyHPMod > 0 && eff.ArmyHPMod != 1 {
		parts = append(parts, fmt.Sprintf("Ordu HP x%.2f", eff.ArmyHPMod))
	}
	if eff.StartResearchTech != "" {
		parts = append(parts, "Arastirma: "+strings.Join(techLabels(gs, []string{eff.StartResearchTech}), ", "))
	}
	if len(eff.CompleteTechs) > 0 {
		parts = append(parts, "Teknoloji: "+strings.Join(techLabels(gs, eff.CompleteTechs), ", "))
	}
	if eff.CapitalSettlementID != "" {
		turns := eff.CapitalMoveTurns
		if turns <= 0 {
			turns = state.DefaultCapitalMoveTurns
		}
		name := eff.CapitalSettlementID
		if gs != nil {
			if region, settlement, _, ok := gs.FindSettlementByID(eff.CapitalSettlementID); ok {
				name = capitalSettlementName(region, settlement, eff.CapitalSettlementID)
			}
		}
		parts = append(parts, fmt.Sprintf("Başkent taşıma: %s (%d tur)", name, turns))
	}
	if len(eff.Relations) > 0 {
		relParts := make([]string, 0, len(eff.Relations))
		for _, rel := range eff.Relations {
			name := rel.FactionID
			if gs != nil && gs.Factions != nil {
				if f := gs.Factions[faction.FactionID(rel.FactionID)]; f != nil && f.NameTR != "" {
					name = f.NameTR
				}
			}
			chunks := make([]string, 0, 2)
			if rel.ScoreDelta != 0 {
				chunks = append(chunks, fmt.Sprintf("%+d", rel.ScoreDelta))
			}
			if rel.Stance != "" {
				chunks = append(chunks, stanceLabel(faction.DiplomaticStance(rel.Stance)))
			}
			if len(chunks) > 0 {
				relParts = append(relParts, name+" "+strings.Join(chunks, " "))
			}
		}
		if len(relParts) > 0 {
			parts = append(parts, "Iliski: "+strings.Join(relParts, ", "))
		}
	}
	return strings.Join(parts, "  |  ")
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
	case state.VictorySurviveTurns:
		return "Hayatta Kalma"
	default:
		return "Zafer"
	}
}

func (g *Game) playerHasResearchableTechs() bool {
	if g == nil || g.gs == nil {
		return false
	}
	f := g.gs.Factions[g.gs.PlayerFactionID]
	if f == nil || g.gs.TechTypes == nil {
		return false
	}
	completed := f.Research.Completed
	for techID, t := range g.gs.TechTypes {
		if t == nil {
			continue
		}
		if completed != nil && completed[techID] {
			continue
		}
		if tech.IsUnlocked(&f.Research, t) {
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
	if g.gs.SiegeAt(rid) != nil {
		g.renderer.ShowCombatResult("Kuşatma altındaki bölgede inşa yapılamaz!")
		return
	}
	b, ok := g.gs.BuildingTypes[buildingID]
	if !ok {
		return
	}
	if buildingID == "port" && !region.IsCoastal(g.gs.Regions) {
		g.renderer.ShowCombatResult("Liman sadece kıyı bölgelerinde inşa edilebilir!")
		return
	}
	// Gerekli arazi kontrolü
	if buildingID != "port" && b.RequiredTerrain != "" && string(region.Terrain) != b.RequiredTerrain {
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
	if order, exists := g.hasProduction(productionKindBuilding, rid, buildingID, g.gs.PlayerFactionID); exists {
		msg := fmt.Sprintf("%s inşaatı devam ediyor (%d tur kaldı). İptal etmek istediğinize emin misiniz?", b.NameTR, order.TurnsLeft)
		g.renderer.ShowConfirmDialog("İnşaatı İptal Et", msg, "İptal Et", "Vazgeç",
			render.InputAction{
				Kind:         render.ActionCancelBuilding,
				TargetRegion: rid,
				BuildingID:   buildingID,
			}, nil)
		return
	}
	cost := economy.ResourceCost{
		Gold:   b.GoldCost,
		Grain:  b.GrainCost,
		Iron:   b.IronCost,
		Timber: b.TimberCost,
		Stone:  b.StoneCost,
	}
	if !cost.CanAfford(f) {
		g.renderer.ShowCombatResult("Yetersiz kaynak! Gerekli: " + cost.ShortTR())
		return
	}
	queuedLevels := g.queuedBuildingCount(rid, buildingID)
	if count+queuedLevels >= b.MaxPerRegion {
		g.renderer.ShowCombatResult(fmt.Sprintf("%s için seviye kuyruğu dolu! (Lv%d)", b.NameTR, b.MaxPerRegion))
		return
	}

	// Seviye arttıkça inşa süresini uzat:
	// Lv1 için base, Lv2 için base+1, Lv3 için base+2 ...
	targetLevel := count + queuedLevels + 1
	turnsRequired := b.TurnsRequired + (targetLevel - 1)
	if turnsRequired < 1 {
		turnsRequired = 1
	}

	cost.Apply(f)
	g.enqueueProduction(productionKindBuilding, rid, buildingID, turnsRequired)
	g.renderer.ShowCombatResult(fmt.Sprintf("%s seviye inşaatı başladı! Lv%d→Lv%d (%d tur)", b.NameTR, count+1, count+2, turnsRequired))
}

// declareWar hedef fraksiyona savaş ilan eder.
func (g *Game) declareWar(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionDeclareWar)
	g.renderer.ShowCombatResult(result.Message)
}

// proposeAlliance hedefe ittifak teklif eder; kabul aynı diplomacy assessment helper'ı ile belirlenir.
func (g *Game) proposeAlliance(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionProposeAlliance)
	g.renderer.ShowCombatResult(result.Message)
}

// proposeTrade hedefe ticaret anlaşması teklif eder.
func (g *Game) proposeTrade(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionProposeTrade)
	g.renderer.ShowCombatResult(result.Message)
}

func (g *Game) cancelAlliance(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionCancelAlliance)
	g.renderer.ShowCombatResult(result.Message)
	if result.Applied {
		g.renderer.AddEventDetail("[DİPLOMASİ] "+result.Message, g.factionNameTR(string(targetID))+" ile kurulu askeri ittifak sona erdirildi.")
	}
}

func (g *Game) cancelTrade(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionCancelTrade)
	g.renderer.ShowCombatResult(result.Message)
	if result.Applied {
		g.renderer.AddEventDetail("[DİPLOMASİ] "+result.Message, g.factionNameTR(string(targetID))+" ile kurulu ticaret rotaları kapatıldı.")
	}
}

func (g *Game) improveRelations(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionImproveRelations)
	g.renderer.ShowCombatResult(result.Message)
}

func (g *Game) sendGift(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionSendGift)
	g.renderer.ShowCombatResult(result.Message)
}

func (g *Game) offerVassalization(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionOfferVassalization)
	g.renderer.ShowCombatResult(result.Message)
}

func (g *Game) releaseVassal(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionReleaseVassal)
	g.renderer.ShowCombatResult(result.Message)
	if result.Applied {
		g.renderer.AddEventDetail("[DİPLOMASİ] "+result.Message, g.factionNameTR(string(targetID))+" üzerindeki vassallık bağı sona erdirildi; devlet bağımsızlığını geri kazandı.")
	}
}

func (g *Game) annexVassal(targetID faction.FactionID) {
	playerID := g.gs.PlayerFactionID
	if reason := diplomacy.ActionBlockReason(g.gs, playerID, targetID, diplomacy.ActionAnnexVassal); reason != "" {
		g.renderer.ShowCombatResult(reason)
		return
	}
	target := g.gs.Factions[targetID]
	player := g.gs.Factions[playerID]
	if target == nil || player == nil {
		g.renderer.ShowCombatResult("Fraksiyon bulunamadı.")
		return
	}

	annexedRegions := 0
	for _, region := range g.gs.Regions {
		if region == nil || region.OwnerID != string(targetID) {
			continue
		}
		g.clearSiege(region.ID)
		region.OwnerID = string(playerID)
		annexedRegions++
	}
	for i := range g.gs.ProductionQueue {
		order := &g.gs.ProductionQueue[i]
		if order.FactionID == string(targetID) {
			order.FactionID = string(playerID)
		}
	}
	player.Gold += target.Gold
	player.Grain += target.Grain
	player.Iron += target.Iron
	player.Timber += target.Timber
	player.Stone += target.Stone
	player.Spice += target.Spice
	player.Cloth += target.Cloth
	target.Gold, target.Grain, target.Iron, target.Timber = 0, 0, 0, 0
	target.Stone, target.Spice, target.Cloth = 0, 0, 0

	collapse := eliminateFaction(g.gs, targetID, playerID)
	diplomacy.NormalizeVassalage(g.gs)
	g.sanitizeDockedFleets()
	g.renderer.CloseDiplomacyPanel()
	g.renderer.MarkMapDirty()
	name := g.factionNameTR(string(targetID))
	message := fmt.Sprintf("%s ilhak edildi; %d bölge doğrudan yönetimine geçti.", name, annexedRegions)
	g.renderer.ShowCombatResult(message)
	g.renderer.AddEventDetail("[İLHAK] "+message, fmt.Sprintf("%s devletinin bölgeleri, kaynakları ve kalan kuvvetleri %s tarafından devralındı. Kara ordusu: %d, donanma: %d.", name, g.factionNameTR(string(playerID)), collapse.TransferredArmies, collapse.TransferredFleets))
}

func (g *Game) oneTimeTrade(targetID faction.FactionID, goodID string, delta int) {
	if delta == 0 || targetID == "" {
		return
	}
	if _, ok := g.gs.Factions[targetID]; !ok {
		g.renderer.ShowCombatResult("Hedef fraksiyon bulunamadı.")
		return
	}
	if !canPlayerOneTimeTradeWith(g.gs, targetID) {
		g.renderer.ShowCombatResult("Bu fraksiyonla pazar işlemi için barış/ticaret ilişkisi ve aktif ticaret ağı bağlantısı gerekiyor.")
		return
	}
	if rel := diplomacy.Relation(g.gs, g.gs.PlayerFactionID, targetID); rel != nil && rel.Stance == faction.StanceWar {
		g.renderer.ShowCombatResult("Savaşta olduğun fraksiyonla pazar işlemi yapamazsın.")
		return
	}
	good := economy.GoodType(goodID)
	amount := delta
	if amount < 0 {
		amount = -amount
	}
	if amount <= 0 {
		return
	}
	if g.gs.MarketPrices == nil {
		g.gs.MarketPrices = economy.ComputeMarketPrices(g.gs.Factions)
	}
	price := g.gs.MarketPrices[good]
	if price <= 0 {
		g.renderer.ShowCombatResult("Geçerli piyasa fiyatı yok.")
		return
	}
	actualAmount := amount
	if delta > 0 {
		target := g.gs.Factions[targetID]
		player := g.gs.Factions[g.gs.PlayerFactionID]
		maxByGold := player.Gold / price
		maxByStock := tradeGoodAmount(target, good)
		actualAmount = minTradeInt(amount, minTradeInt(maxByGold, maxByStock))
		if actualAmount <= 0 {
			g.renderer.ShowCombatResult("Satın alma başarısız: altın yetersiz veya satıcıda stok yok.")
			return
		}
		if economy.TransferGoods(g.gs.Factions, targetID, g.gs.PlayerFactionID, good, actualAmount, g.gs.MarketPrices) {
			g.renderer.ShowCombatResult(fmt.Sprintf("%s fraksiyonundan %d %s satın alındı. (%d altın)", targetID, actualAmount, tradeGoodLabelTR(good), actualAmount*price))
			return
		}
	} else {
		target := g.gs.Factions[targetID]
		player := g.gs.Factions[g.gs.PlayerFactionID]
		maxByBuyerGold := target.Gold / price
		maxByStock := tradeGoodAmount(player, good)
		actualAmount = minTradeInt(amount, minTradeInt(maxByBuyerGold, maxByStock))
		if actualAmount <= 0 {
			g.renderer.ShowCombatResult("Satış başarısız: sende stok yok veya alıcıda altın yok.")
			return
		}
		if economy.TransferGoods(g.gs.Factions, g.gs.PlayerFactionID, targetID, good, actualAmount, g.gs.MarketPrices) {
			g.renderer.ShowCombatResult(fmt.Sprintf("%s fraksiyonuna %d %s satıldı. (%d altın)", targetID, actualAmount, tradeGoodLabelTR(good), actualAmount*price))
			return
		}
	}
	g.renderer.ShowCombatResult("Pazar işlemi başarısız.")
}

func canPlayerOneTimeTradeWith(gs *state.GameState, targetID faction.FactionID) bool {
	if gs == nil || targetID == "" || targetID == gs.PlayerFactionID {
		return false
	}
	target := gs.Factions[targetID]
	if target == nil || target.IsEliminated {
		return false
	}
	rel := diplomacy.Relation(gs, gs.PlayerFactionID, targetID)
	if rel == nil {
		return false
	}
	if rel.Stance != faction.StancePeace && rel.Stance != faction.StanceTrade && rel.Stance != faction.StanceAllied {
		return false
	}
	if len(gs.TradeRoutes) == 0 {
		return false
	}
	seen := map[faction.FactionID]bool{gs.PlayerFactionID: true}
	q := []faction.FactionID{gs.PlayerFactionID}
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		if cur == targetID {
			return true
		}
		for _, tr := range gs.TradeRoutes {
			if tr == nil || tr.SuspendedTurns > 0 {
				continue
			}
			var nxt faction.FactionID
			switch faction.FactionID(tr.FromFactionID) {
			case cur:
				nxt = faction.FactionID(tr.ToFactionID)
			default:
				if faction.FactionID(tr.ToFactionID) == cur {
					nxt = faction.FactionID(tr.FromFactionID)
				}
			}
			if nxt == "" || seen[nxt] {
				continue
			}
			seen[nxt] = true
			q = append(q, nxt)
		}
	}
	return false
}

func tradeGoodLabelTR(good economy.GoodType) string {
	return economy.GoodLowerNameTR(good)
}

func tradeGoodAmount(f *faction.Faction, good economy.GoodType) int {
	if kind, ok := economy.GoodToResourceKind(good); ok {
		return economy.FactionResourceAmount(f, kind)
	}
	return 0
}

func minTradeInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// proposePeace hedefe barış teklif eder (her zaman kabul edilir — basit versiyon).
func (g *Game) proposePeace(targetID faction.FactionID) {
	result := diplomacy.Execute(g.gs, g.gs.PlayerFactionID, targetID, diplomacy.ActionProposePeace)
	g.renderer.ShowCombatResult(result.Message)
}

func (g *Game) respondDiplomacyOffer(index int, accepted bool) {
	_, result, _ := g.resolveDiplomacyOffer(index, accepted)
	g.renderer.ShowCombatResult(result.Message)
	if accepted && result.Applied {
		g.renderer.AddEvent("[DIPLOMASI] " + result.Message)
	}
}

func (g *Game) handleAITurnOfferResponse(index int, accepted bool) {
	offer, result, _ := g.resolveDiplomacyOffer(index, accepted)
	g.renderer.ShowCombatResult(result.Message)
	if accepted && result.Applied {
		g.renderer.AddEvent("[DIPLOMASI] " + result.Message)
		if currentFID, ok := g.currentAITurnFactionID(); ok && offer.FromFactionID == currentFID {
			g.aiTurn.index++
			g.aiTurn.stepper = nil
			g.aiTurn.waitFrames = 0
		}
	}
}

func (g *Game) resolveDiplomacyOffer(index int, accepted bool) (state.DiplomaticOffer, diplomacy.Result, bool) {
	if g == nil || g.gs == nil || index < 0 || index >= len(g.gs.DiplomaticOffers) {
		return state.DiplomaticOffer{}, diplomacy.Result{Message: "Geçersiz diplomasi teklifi."}, false
	}
	offer := g.gs.DiplomaticOffers[index]
	result := diplomacy.ResolveOffer(g.gs, index, accepted)
	g.appendDiplomacyOfferHistory(offer, accepted, result)
	return offer, result, true
}

func (g *Game) appendDiplomacyOfferHistory(offer state.DiplomaticOffer, accepted bool, result diplomacy.Result) {
	if g == nil || g.gs == nil {
		return
	}
	history := state.DiplomaticOfferHistoryEntry{
		FromFactionID:  offer.FromFactionID,
		ToFactionID:    offer.ToFactionID,
		Action:         offer.Action,
		CreatedTurn:    offer.CreatedTurn,
		ResolvedTurn:   g.gs.Turn,
		Accepted:       accepted,
		Applied:        result.Applied,
		Priority:       offer.Priority,
		PriorityReason: offer.PriorityReason,
		ResultMessage:  result.Message,
	}
	g.gs.DiplomaticOfferHistory = append(g.gs.DiplomaticOfferHistory, history)
	if overflow := len(g.gs.DiplomaticOfferHistory) - maxDiplomaticOfferHistoryEntries; overflow > 0 {
		copy(g.gs.DiplomaticOfferHistory, g.gs.DiplomaticOfferHistory[overflow:])
		g.gs.DiplomaticOfferHistory = g.gs.DiplomaticOfferHistory[:maxDiplomaticOfferHistoryEntries]
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
	if err := writeScenarioSettlements(gs); err != nil {
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
	type regionExport struct {
		ID               world.RegionID    `json:"id"`
		Name             string            `json:"name"`
		NameTR           string            `json:"name_tr"`
		Terrain          world.TerrainType `json:"terrain"`
		OwnerID          string            `json:"owner_id"`
		Neighbors        []world.RegionID  `json:"neighbors"`
		WorldX           int               `json:"world_x"`
		WorldY           int               `json:"world_y"`
		ShapeID          string            `json:"shape_id,omitempty"`
		IsSea            bool              `json:"is_sea"`
		IsLocked         bool              `json:"is_locked"`
		UnlockTurn       int               `json:"unlock_turn"`
		BaseGoldIncome   int               `json:"base_gold_income"`
		BaseGrainOutput  int               `json:"base_grain_output"`
		BaseIronOutput   int               `json:"base_iron_output"`
		BaseTimberOutput int               `json:"base_timber_output"`
		BaseStoneOutput  int               `json:"base_stone_output"`
		BaseSpiceOutput  int               `json:"base_spice_output"`
		BaseClothOutput  int               `json:"base_cloth_output"`
		TradeCapacity    int               `json:"trade_capacity"`
		Satisfaction     int               `json:"satisfaction"`
		TaxRate          int               `json:"tax_rate"`
		Population       int               `json:"population"`
		Religion         string            `json:"religion"`
		ConversionTurns  int               `json:"conversion_turns,omitempty"`
		ActiveEventID    string            `json:"active_event_id"`
		Buildings        []string          `json:"buildings"`
	}
	cloneRegion := func(region *world.Region) *regionExport {
		if region == nil {
			return nil
		}
		out := &regionExport{
			ID:               region.ID,
			Name:             region.Name,
			NameTR:           region.NameTR,
			Terrain:          region.Terrain,
			OwnerID:          region.OwnerID,
			Neighbors:        append([]world.RegionID(nil), region.Neighbors...),
			WorldX:           region.WorldX,
			WorldY:           region.WorldY,
			ShapeID:          region.ShapeID,
			IsSea:            region.IsSea,
			IsLocked:         region.IsLocked,
			UnlockTurn:       region.UnlockTurn,
			BaseGoldIncome:   region.BaseGoldIncome,
			BaseGrainOutput:  region.BaseGrainOutput,
			BaseIronOutput:   region.BaseIronOutput,
			BaseTimberOutput: region.BaseTimberOutput,
			BaseStoneOutput:  region.BaseStoneOutput,
			BaseSpiceOutput:  region.BaseSpiceOutput,
			BaseClothOutput:  region.BaseClothOutput,
			TradeCapacity:    region.TradeCapacity,
			Satisfaction:     region.Satisfaction,
			TaxRate:          region.TaxRate,
			Population:       region.Population,
			Religion:         region.Religion,
			ConversionTurns:  region.ConversionTurns,
			ActiveEventID:    region.ActiveEventID,
			Buildings:        append([]string(nil), region.Buildings...),
		}
		return out
	}

	regions := make([]*regionExport, 0, len(gs.Regions))
	if len(gs.RegionOrder) > 0 {
		seen := make(map[world.RegionID]bool, len(gs.RegionOrder))
		for _, rid := range gs.RegionOrder {
			if region, ok := gs.Regions[rid]; ok {
				regions = append(regions, cloneRegion(region))
				seen[rid] = true
			}
		}
		for rid, region := range gs.Regions {
			if !seen[rid] {
				regions = append(regions, cloneRegion(region))
			}
		}
	} else {
		for _, region := range gs.Regions {
			regions = append(regions, cloneRegion(region))
		}
	}

	data, err := json.MarshalIndent(regions, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func writeScenarioSettlements(gs *state.GameState) error {
	path := filepath.Join(gs.ScenarioPath, "data", "settlements.json")
	entries := make([]world.SettlementListEntry, 0, len(gs.Regions))
	appendEntry := func(rid world.RegionID, region *world.Region) {
		if region == nil {
			return
		}
		entry := world.SettlementListEntry{
			RegionID:    rid,
			Settlements: append([]world.Settlement(nil), region.Settlements...),
		}
		entries = append(entries, entry)
	}
	if len(gs.RegionOrder) > 0 {
		seen := make(map[world.RegionID]bool, len(gs.RegionOrder))
		for _, rid := range gs.RegionOrder {
			if region, ok := gs.Regions[rid]; ok {
				appendEntry(rid, region)
				seen[rid] = true
			}
		}
		for rid, region := range gs.Regions {
			if !seen[rid] {
				appendEntry(rid, region)
			}
		}
	} else {
		for rid, region := range gs.Regions {
			appendEntry(rid, region)
		}
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func writeScenarioFactions(gs *state.GameState) error {
	path := filepath.Join(gs.ScenarioPath, "data", "factions.json")
	ids := make([]faction.FactionID, 0, len(gs.Factions))
	seen := make(map[faction.FactionID]struct{}, len(gs.Factions))
	for _, fid := range gs.FactionOrder {
		if gs.Factions[fid] == nil {
			continue
		}
		ids = append(ids, fid)
		seen[fid] = struct{}{}
	}
	orderedCount := len(ids)
	for fid := range gs.Factions {
		if _, ok := seen[fid]; ok {
			continue
		}
		ids = append(ids, fid)
	}
	if len(ids) > orderedCount {
		sort.Slice(ids[orderedCount:], func(i, j int) bool {
			return ids[orderedCount+i] < ids[orderedCount+j]
		})
	}

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
	render.SyncLandShapesFromRegionPaint(gs)
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
		IsGarrison         bool            `json:"is_garrison,omitempty"`
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
			IsGarrison:         a.IsGarrison,
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
	g.startLoading(loadingSave, "Kayıt yükleniyor...", func(setProgress func(int)) loadingResult {
		setProgress(10)
		gs, err := save.LoadSlot(slotName)
		if err != nil {
			return loadingResult{err: err, fallback: fallback}
		}
		setProgress(65)
		evts, err := loadScenarioEvents(gs.ScenarioPath)
		if err != nil {
			return loadingResult{err: err, fallback: fallback}
		}
		worldMap := render.PrepareWorldMap(gs, "", render.MapModeNormal, func(progress int) {
			setProgress(65 + progress*35/100)
		})
		return loadingResult{
			gs:         gs,
			evts:       evts,
			successMsg: "Oyun yüklendi!",
			worldMap:   worldMap,
		}
	})
}

func (g *Game) startPreparePlayerTurn() {
	g.finishAITurnSequence()
	g.startLoading(loadingWorldMap, "Harita hazırlanıyor...", func(setProgress func(int)) loadingResult {
		worldMap := render.PrepareWorldMap(g.gs, "", render.MapModeNormal, setProgress)
		return loadingResult{worldMap: worldMap}
	})
}

// resetToNewGame state'i temizler ve senaryo seçimine geçer.
func (g *Game) resetToNewGame() {
	difficulty := g.gs.Difficulty
	audio.StopMusic()
	g.finishAITurnSequence()
	gs := &state.GameState{
		Phase:      state.PhaseScenarioSelect,
		Difficulty: difficulty,
	}
	g.gs = gs
	g.pendingConquestDecisions = nil
	g.renderer.ReloadGameState(gs)
	g.renderer.SetEventCodexEntries([4][]render.EventCodexEntry{})
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
	g.startLoading(loadingScenario, "Senaryo yükleniyor...", func(setProgress func(int)) loadingResult {
		gs, evts, err := loadScenarioData(scenarioPath, difficulty, setProgress)
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

func loadScenarioData(scenarioPath string, difficulty int, setProgress func(int)) (*state.GameState, []*events.Event, error) {
	sc := scenarioByPath(scenarioPath)
	yield := func() { runtime.Gosched() }
	progressTotal := 12
	progressStep := 0
	advance := func() {
		progressStep++
		if setProgress != nil {
			setProgress(progressStep * 100 / progressTotal)
		}
	}
	if setProgress != nil {
		setProgress(0)
	}

	dp := func(f string) string { return scenarioPath + "/data/" + f }

	regions, regionOrder, err := world.LoadRegionsWithOrder(dp("regions.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("bölgeler yüklenemedi: %w", err)
	}
	advance()
	yield()
	if err := world.LoadRegionSettlements(dp("settlements.json"), regions); err != nil {
		return nil, nil, fmt.Errorf("yerleşimler yüklenemedi: %w", err)
	}
	advance()
	yield()
	shapeData, err := world.LoadCountryShapes(dp("country_shapes.json"), regions)
	if err != nil {
		log.Printf("Ülke sınırları yüklenemedi: %v", err)
	}
	advance()
	yield()
	factions, factionOrder, err := faction.LoadFactionsWithOrder(dp("factions.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("fraksiyonlar yüklenemedi: %w", err)
	}
	advance()
	yield()
	relations, err := faction.LoadRelations(dp("relations.json"), factions)
	if err != nil {
		return nil, nil, fmt.Errorf("ilişkiler yüklenemedi: %w", err)
	}
	advance()
	yield()
	unitTypes, err := army.LoadUnitTypes(dp("units.json"))
	if err != nil {
		log.Printf("Birim tipleri yüklenemedi: %v", err)
	}
	advance()
	yield()
	buildingTypes, err := city.LoadBuildings(dp("buildings.json"))
	if err != nil {
		log.Printf("Binalar yüklenemedi: %v", err)
	}
	advance()
	yield()
	techTypes, err := tech.LoadTechnologies(dp("technologies.json"))
	if err != nil {
		log.Printf("Teknolojiler yüklenemedi: %v", err)
	}
	advance()
	yield()
	evts, err := events.LoadEvents(dp("events.json"))
	if err != nil {
		log.Printf("Olaylar yüklenemedi: %v", err)
	}
	advance()
	yield()
	armies, err := army.LoadArmies(dp("armies.json"))
	if err != nil {
		log.Printf("Ordular yüklenemedi: %v", err)
		armies = map[army.ArmyID]*army.Army{}
	}
	army.NormalizeLegacyGarrisons(armies)
	advance()
	yield()
	tradeCenters, err := world.LoadTradeCenters(dp("trade_centers.json"), regions)
	if err != nil {
		log.Printf("Ticaret merkezleri yüklenemedi: %v", err)
	}
	advance()
	yield()

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
		FactionOrder:       factionOrder,
		Armies:             armies,
		ShapeData:          shapeData,
		UnitTypes:          unitTypes,
		BuildingTypes:      buildingTypes,
		TechTypes:          techTypes,
		ScenarioVictories:  victoryOpts,
		AvailableVictories: scenario.FilterVictoryOptionsForFaction(victoryOpts, ""),
		Relations:          relations,
		TradeCenters:       tradeCenters,
		NextArmySeq:        len(armies),
		FiredEventIDs:      map[string]bool{},
	}
	army.InitializeLegacyFleetDocking(gs.Armies, gs.Regions)
	diplomacy.NormalizeVassalage(gs)
	diplomacy.EnsureTradeRoutesForActiveRelations(gs)
	gs.SyncTimedRegionUnlocks()
	gs.NormalizeFactionCapitals()
	if editMode {
		gs.Phase = state.PhaseEditMode
	}
	advance()
	yield()

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
	if g.gs.SiegeAt(rid) != nil {
		g.renderer.ShowCombatResult("Kuşatma altındaki bölgede asker alımı yapılamaz!")
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

	// Bina seviyesi kontrolü (aynı ID'nin tekrar sayısı = seviye)
	requiredLevel := utype.RequiredBldgLevel
	if utype.RequiredBldg != "" && requiredLevel <= 0 {
		requiredLevel = 1
	}
	bldgLevel := 0
	for _, bid := range region.Buildings {
		if bid == utype.RequiredBldg {
			bldgLevel++
		}
	}
	if utype.RequiredBldg != "" && bldgLevel < requiredLevel {
		bldgName := utype.RequiredBldg
		if b, ok2 := g.gs.BuildingTypes[bldgName]; ok2 {
			bldgName = b.NameTR
		}
		g.renderer.ShowCombatResult(fmt.Sprintf("Bu birlik için %s Lv%d gerekli! (mevcut: Lv%d)", bldgName, requiredLevel, bldgLevel))
		return
	}

	// Teknoloji kontrolü
	if utype.RequiredTech != "" && !f.Research.Completed[utype.RequiredTech] {
		g.renderer.ShowCombatResult("Araştırma gerekli: " + utype.RequiredTech)
		return
	}

	cost := economy.ResourceCost{
		Gold:   utype.GoldCost,
		Grain:  utype.GrainCost,
		Iron:   utype.IronCost,
		Timber: utype.TimberCost,
		Stone:  utype.StoneCost,
	}
	if !cost.CanAfford(f) {
		g.renderer.ShowCombatResult("Yetersiz kaynak! Gerekli: " + cost.ShortTR())
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
		maxByCost := maxAffordableByCost(f, cost)
		if maxByCost <= 0 || quantity <= 0 {
			g.renderer.ShowCombatResult("Yetersiz kaynak! Gerekli: " + cost.ShortTR())
			return
		}
		if quantity > maxByCost {
			quantity = maxByCost
		}
		for i := 0; i < quantity; i++ {
			cost.Apply(f)
			g.enqueueProduction(productionKindUnit, rid, unitTypeID, utype.TurnsRequired)
		}
		g.renderer.ShowCombatResult(fmt.Sprintf("%s üretimi başladı! x%d (%d tur)", utype.NameTR, quantity, utype.TurnsRequired))
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

	targetArmy, canCreateNew := g.findRecruitableLandArmy(rid, pid)
	if targetArmy == nil && !canCreateNew {
		g.renderer.ShowCombatResult(fmt.Sprintf("Maksimum ordu sayısına ulaşıldı! (%d/%d)", g.gs.CurrentLandArmies(pid), g.gs.MaxLandArmies(pid)))
		return
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
	maxByCost := maxAffordableByCost(f, cost)
	if maxByCost <= 0 || quantity <= 0 {
		g.renderer.ShowCombatResult("Yetersiz kaynak! Gerekli: " + cost.ShortTR())
		return
	}
	if quantity > maxByCost {
		quantity = maxByCost
	}
	for i := 0; i < quantity; i++ {
		cost.Apply(f)
		g.enqueueProduction(productionKindUnit, rid, unitTypeID, utype.TurnsRequired)
	}
	g.renderer.ShowCombatResult(fmt.Sprintf("%s eğitimi başladı! x%d (%d tur)", utype.NameTR, quantity, utype.TurnsRequired))
}

func maxAffordableByCost(f *faction.Faction, cost economy.ResourceCost) int {
	if f == nil {
		return 0
	}
	maxVal := 1 << 30
	for _, kind := range economy.CostResourceKinds() {
		amount := cost.Amount(kind)
		if amount <= 0 {
			continue
		}
		maxVal = minInt(maxVal, economy.FactionResourceAmount(f, kind)/amount)
	}
	if maxVal == 1<<30 {
		return 9999
	}
	if maxVal < 0 {
		return 0
	}
	return maxVal
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// cancelBuilding onay diyaloğundan sonra bina inşaatını iptal eder.
func (g *Game) cancelBuilding(rid world.RegionID, buildingID string) {
	b, ok := g.gs.BuildingTypes[buildingID]
	if !ok {
		return
	}
	if !g.cancelProduction(productionKindBuilding, rid, buildingID, g.gs.PlayerFactionID) {
		g.renderer.ShowCombatResult("İptal edilecek inşaat bulunamadı.")
		return
	}
	f := g.gs.Factions[g.gs.PlayerFactionID]
	cost := economy.ResourceCost{
		Gold:   b.GoldCost,
		Grain:  b.GrainCost,
		Iron:   b.IronCost,
		Timber: b.TimberCost,
		Stone:  b.StoneCost,
	}
	cost.Refund(f)
	g.renderer.ShowCombatResult(fmt.Sprintf("%s inşaatı iptal edildi. İade: %s", b.NameTR, cost.ShortTR()))
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
	refund := economy.ResourceCost{
		Gold:   utype.GoldCost,
		Grain:  utype.GrainCost,
		Iron:   utype.IronCost,
		Timber: utype.TimberCost,
		Stone:  utype.StoneCost,
	}
	refund.Refund(f)
	g.renderer.ShowCombatResult(fmt.Sprintf("%s emri iptal edildi. İade: %s", utype.NameTR, refund.ShortTR()))
}

func (g *Game) regionUnitProductionCapacity(region *world.Region, unitTypeID string) int {
	if g == nil || g.gs == nil {
		return 0
	}
	return state.UnitProductionLimit(region, g.gs.UnitTypes[unitTypeID])
}

func (g *Game) productionCapacityLane(unitTypeID string) string {
	if g != nil && g.gs != nil {
		if utype := g.gs.UnitTypes[unitTypeID]; utype != nil && utype.RequiredBldg == "port" {
			return "port"
		}
	}
	return "barracks"
}

func (g *Game) productionCapacityReason(unitTypeID string) string {
	if g.productionCapacityLane(unitTypeID) == "port" {
		return "liman tur limiti"
	}
	return "kışla tur limiti"
}

func (g *Game) fleetHasTransportCapacity(fleet *army.Army, unitCount int) bool {
	if g == nil || g.gs == nil || fleet == nil || !fleet.IsNaval {
		return false
	}
	return fleet.CanEmbarkUnits(g.gs.UnitTypes, unitCount)
}

func (g *Game) findFriendlyTransportFleet(ownerID string, seaRegionID world.RegionID) *army.Army {
	for _, candidate := range g.gs.Armies {
		if candidate.OwnerID != ownerID || !candidate.IsNaval || candidate.RegionID != seaRegionID {
			continue
		}
		if candidate.TransportCapacity(g.gs.UnitTypes) > 0 {
			return candidate
		}
	}
	return nil
}

func (g *Game) canEmbarkLandArmy(a *army.Army) bool {
	return a.CanEmbark(g.gs.UnitTypes)
}

func (g *Game) embarkBlockedMessage(a *army.Army) string {
	if g == nil || g.gs == nil || a == nil {
		return "Bu ordudaki bazı birimler denizden taşınamaz."
	}
	blockers := a.EmbarkBlockerNames(g.gs.UnitTypes)
	if len(blockers) == 0 {
		return "Bu ordudaki bazı birimler denizden taşınamaz."
	}
	return fmt.Sprintf("Bu ordu denizden taşınamaz. Uygun olmayan birlikler: %s.", strings.Join(blockers, ", "))
}

func (g *Game) fleetCanEmbarkFromRegion(fleet *army.Army, sourceRegionID world.RegionID) bool {
	if g == nil || g.gs == nil || fleet == nil || !fleet.IsNaval || sourceRegionID == "" {
		return false
	}
	if fleet.DockedRegionID == sourceRegionID {
		return true
	}
	src := g.gs.Regions[sourceRegionID]
	if src == nil {
		return false
	}
	for _, nid := range src.Neighbors {
		if nid == fleet.RegionID {
			return true
		}
	}
	return false
}

func (g *Game) findFriendlyEmbarkFleet(ownerID string, seaRegionID world.RegionID, unitCount int) *army.Army {
	return g.findFriendlyEmbarkFleetFromRegion(ownerID, "", seaRegionID, unitCount)
}

func (g *Game) findFriendlyEmbarkFleetFromRegion(ownerID string, sourceRegionID, seaRegionID world.RegionID, unitCount int) *army.Army {
	var fallback *army.Army
	for _, candidate := range g.gs.Armies {
		if candidate.OwnerID != ownerID || !candidate.IsNaval || candidate.RegionID != seaRegionID {
			continue
		}
		if !g.fleetHasTransportCapacity(candidate, unitCount) {
			continue
		}
		if sourceRegionID != "" && candidate.DockedRegionID == sourceRegionID {
			return candidate
		}
		if fallback == nil {
			fallback = candidate
		}
	}
	return fallback
}

func (g *Game) embarkArmyOntoFleet(armyID, fleetID army.ArmyID) {
	a, ok := g.gs.Armies[armyID]
	if !ok || a.OwnerID != string(g.gs.PlayerFactionID) || a.IsNaval || a.MovePoints <= 0 {
		return
	}
	fleet := g.gs.Armies[fleetID]
	if fleet == nil || !fleet.IsNaval || fleet.OwnerID != a.OwnerID {
		return
	}
	if !g.canEmbarkLandArmy(a) {
		g.renderer.ShowCombatResult(g.embarkBlockedMessage(a))
		return
	}
	if !g.fleetCanEmbarkFromRegion(fleet, a.RegionID) {
		g.renderer.ShowCombatResult("Bu filo bu bölgeden yükleme menzilinde değil.")
		return
	}
	if !g.fleetHasTransportCapacity(fleet, len(a.Units)) {
		g.renderer.ShowCombatResult("Seçilen filoda yeterli nakliye kapasitesi yok.")
		return
	}
	fleet.EmbarkedUnits = append(fleet.EmbarkedUnits, a.Units...)
	fleet.MovePoints = max(0, fleet.MovePoints-1)
	delete(g.gs.Armies, armyID)
	g.renderer.SelectedArmy = fleet.ID
	g.renderer.ShowCombatResult(fmt.Sprintf("Ordu nakliye filosuna bindi. Kalan kapasite: %d.", fleet.AvailableTransportCapacity(g.gs.UnitTypes)))
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
	if diplomacy.SameRealm(g.gs, faction.FactionID(fleet.OwnerID), faction.FactionID(targetRegion.OwnerID)) {
		return true
	}
	key := faction.RelationKey(faction.FactionID(fleet.OwnerID), faction.FactionID(targetRegion.OwnerID))
	rel, ok := g.gs.Relations[key]
	return ok && (rel.Stance == faction.StanceWar || rel.Stance == faction.StanceAllied)
}

func (g *Game) resolveFleetDisembark(fleet *army.Army, target world.RegionID, targetRegion *world.Region) bool {
	return g.resolveFleetDisembarkWithStance(fleet, target, targetRegion, combat.BattleStanceBalanced)
}

func (g *Game) resolveFleetDisembarkWithStance(fleet *army.Army, target world.RegionID, targetRegion *world.Region, battleStance combat.BattleStance) bool {
	if fleet == nil || targetRegion == nil || !fleet.IsNaval {
		return false
	}
	battleStance = combat.NormalizeBattleStance(battleStance)
	if !g.canDisembarkToLand(fleet, targetRegion) {
		if len(fleet.EmbarkedUnits) == 0 {
			g.renderer.ShowCombatResult("Çıkarma emri reddedildi: filoda taşınan kara birimi yok.")
		} else {
			g.renderer.ShowCombatResult("Çıkarma emri reddedildi: düşman kıyıya çıkarmak için savaş halinde olmalısın.")
		}
		return true
	}
	var enemyArmy *army.Army
	isAlliedDisembark := false
	if targetRegion.OwnerID != fleet.OwnerID && targetRegion.OwnerID != "" {
		if diplomacy.SameRealm(g.gs, faction.FactionID(fleet.OwnerID), faction.FactionID(targetRegion.OwnerID)) {
			isAlliedDisembark = true
		}
		key := faction.RelationKey(faction.FactionID(fleet.OwnerID), faction.FactionID(targetRegion.OwnerID))
		if rel, exists := g.gs.Relations[key]; exists && rel.Stance == faction.StanceAllied {
			isAlliedDisembark = true
		}
	}
	if !isAlliedDisembark {
		for _, ea := range g.gs.Armies {
			if ea.RegionID == target && ea.OwnerID != fleet.OwnerID {
				enemyArmy = ea
				break
			}
		}
	}
	if enemyArmy != nil {
		landing := &army.Army{
			OwnerID: fleet.OwnerID,
			Units:   append([]army.Unit(nil), fleet.EmbarkedUnits...),
		}
		attackerBefore := snapshotBattleArmy(landing, g.gs.UnitTypes)
		defenderBefore := snapshotBattleArmy(enemyArmy, g.gs.UnitTypes)
		atkMods := techModsFor(g.gs, fleet.OwnerID)
		defMods := techModsFor(g.gs, enemyArmy.OwnerID)
		result := combat.ResolveBattleWithContextPlan(landing, enemyArmy, targetRegion.Terrain, g.gs.UnitTypes, atkMods, defMods, combat.BattleContextAmphibious, battleStance)
		fleet.EmbarkedUnits = fleet.EmbarkedUnits[:0]
		fleet.MovePoints--

		if result.AttackerWins {
			if len(enemyArmy.Units) == 0 {
				delete(g.gs.Armies, enemyArmy.ID)
			}
			g.spawnDisembarkedArmy(fleet.OwnerID, target, landing.Units)
			prompted := g.queueConquestDecision(faction.FactionID(fleet.OwnerID), targetRegion, true)
			collapse := eliminationResult{}
			outcomeDetail := "Kıyı başı kuruldu, savunma kırıldı ve bölge ele geçirildi."
			if !prompted {
				collapse = g.applyConquestWithNavalEviction(targetRegion, fleet.OwnerID)
				g.renderer.MarkMapDirty()
			} else {
				outcomeDetail = "Kıyı başı kuruldu; ilhak ya da vassallık için savaş sonrası karar bekleniyor."
			}
			g.presentBattleReport(g.makeBattleReport(
				render.BattleSceneAmphibious,
				targetRegion.NameTR,
				battleStance,
				result.Description,
				outcomeDetail,
				"Çıkarma Gücü",
				"Kıyı Savunması",
				g.factionNameTR(fleet.OwnerID),
				g.factionNameTR(enemyArmy.OwnerID),
				attackerBefore,
				landing,
				defenderBefore,
				enemyArmy,
			))
			g.announceElimination(collapse)
			return true
		}

		g.presentBattleReport(g.makeBattleReport(
			render.BattleSceneAmphibious,
			targetRegion.NameTR,
			battleStance,
			result.Description,
			"Çıkarma sahilde kırıldı; birlikler tutunamadı.",
			"Çıkarma Gücü",
			"Kıyı Savunması",
			g.factionNameTR(fleet.OwnerID),
			g.factionNameTR(enemyArmy.OwnerID),
			attackerBefore,
			landing,
			defenderBefore,
			enemyArmy,
		))
		return true
	}

	landingBefore := snapshotBattleArmy(&army.Army{
		OwnerID: fleet.OwnerID,
		Units:   append([]army.Unit(nil), fleet.EmbarkedUnits...),
	}, g.gs.UnitTypes)
	g.disembarkFleet(fleet, target)
	fleet.MovePoints--
	isAlliedDisembark = false
	if targetRegion.OwnerID != fleet.OwnerID && targetRegion.OwnerID != "" {
		if diplomacy.SameRealm(g.gs, faction.FactionID(fleet.OwnerID), faction.FactionID(targetRegion.OwnerID)) {
			isAlliedDisembark = true
		}
		key := faction.RelationKey(faction.FactionID(fleet.OwnerID), faction.FactionID(targetRegion.OwnerID))
		if rel, exists := g.gs.Relations[key]; exists && rel.Stance == faction.StanceAllied {
			isAlliedDisembark = true
		}
	}
	if targetRegion.OwnerID != fleet.OwnerID && !isAlliedDisembark {
		defenderFaction := g.factionNameTR(targetRegion.OwnerID)
		prompted := g.queueConquestDecision(faction.FactionID(fleet.OwnerID), targetRegion, true)
		collapse := eliminationResult{}
		detail := "Kıyıda savunan ordu yoktu; çıkarma savaşsız tamamlandı ve bölge ele geçirildi."
		if !prompted {
			collapse = g.applyConquestWithNavalEviction(targetRegion, fleet.OwnerID)
			g.renderer.MarkMapDirty()
		} else {
			detail = "Kıyıda savunan ordu yoktu; devletin geleceği için ilhak veya vassallık kararı bekleniyor."
		}
		g.presentBattleReport(g.makeBattleReportFromSnapshots(
			render.BattleSceneAmphibious,
			targetRegion.NameTR,
			"",
			"Direniş Görülmedi",
			detail,
			"Çıkarma Gücü",
			"Direniş Yok",
			g.factionNameTR(fleet.OwnerID),
			defenderFaction,
			landingBefore,
			landingBefore,
			battleArmySnapshot{},
			battleArmySnapshot{},
		))
		g.announceElimination(collapse)
		return true
	}
	if isAlliedDisembark {
		g.renderer.ShowCombatResult("Çıkarma tamamlandı: birlikler müttefik toprağına indi.")
		g.renderer.AddEvent(fmt.Sprintf("Birlikler müttefik kıyısına çıktı: %s", targetRegion.NameTR))
	} else {
		g.renderer.ShowCombatResult("Çıkarma tamamlandı: birlikler karaya indi.")
		g.renderer.AddEvent(fmt.Sprintf("Birlikler karaya çıktı: %s", targetRegion.NameTR))
	}
	return true
}

func (g *Game) forceDisembarkFleet(aid army.ArmyID, target world.RegionID) {
	g.forceDisembarkFleetWithStance(aid, target, combat.BattleStanceBalanced)
}

func (g *Game) forceDisembarkFleetWithStance(aid army.ArmyID, target world.RegionID, battleStance combat.BattleStance) {
	fleet, ok := g.gs.Armies[aid]
	if !ok || fleet.OwnerID != string(g.gs.PlayerFactionID) || fleet.MovePoints <= 0 {
		return
	}
	targetRegion, ok := g.gs.Regions[target]
	if !ok {
		return
	}
	src, ok := g.gs.Regions[fleet.RegionID]
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
	if !isNeighbor || !targetRegion.CanLandEnter() {
		return
	}
	g.resolveFleetDisembarkWithStance(fleet, target, targetRegion, battleStance)
}

func (g *Game) fleetsCanSharePort(fleetOwnerID, regionOwnerID string) bool {
	if fleetOwnerID == "" || regionOwnerID == "" {
		return false
	}
	if fleetOwnerID == regionOwnerID {
		return true
	}
	if diplomacy.SameRealm(g.gs, faction.FactionID(fleetOwnerID), faction.FactionID(regionOwnerID)) {
		return true
	}
	key := faction.RelationKey(faction.FactionID(fleetOwnerID), faction.FactionID(regionOwnerID))
	rel, ok := g.gs.Relations[key]
	return ok && rel.Stance == faction.StanceAllied
}

func (g *Game) dockSettlementIDForRegion(region *world.Region) string {
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

func (g *Game) canDockFleetAtRegion(fleet *army.Army, targetRegion *world.Region) bool {
	if fleet == nil || targetRegion == nil || !fleet.IsNaval || targetRegion.IsSea {
		return false
	}
	if !g.fleetsCanSharePort(fleet.OwnerID, targetRegion.OwnerID) {
		return false
	}
	return targetRegion.HasPortBuilding()
}

// applyConquestWithNavalEviction bölge sahipliği değiştiğinde limanda bekleyen
// eski sahip filolarını en yakın deniz bölgesine çıkarır. Eğer bu fetih eski
// sahibi tamamen yıkarsa kalan ordular galibe devrolur.
func (g *Game) applyConquestWithNavalEviction(targetRegion *world.Region, newOwnerID string) eliminationResult {
	if targetRegion == nil {
		return eliminationResult{}
	}
	g.clearSiege(targetRegion.ID)
	prevOwnerID := targetRegion.OwnerID
	attackerReligion := ownerReligion(g.gs, newOwnerID)
	targetRegion.ApplyConquest(newOwnerID, attackerReligion)
	if prevOwnerID == "" || prevOwnerID == newOwnerID {
		return eliminationResult{}
	}
	prevFactionID := faction.FactionID(prevOwnerID)
	g.handleCapitalCapture(prevFactionID, newOwnerID, targetRegion)
	if len(g.gs.LandRegionsOwnedBy(prevFactionID)) == 0 {
		result := eliminateFaction(g.gs, prevFactionID, faction.FactionID(newOwnerID))
		g.retreatArmiesFromCapturedRegion(targetRegion.ID, newOwnerID)
		g.evictDockedFleetsFromCapturedPort(targetRegion.ID, newOwnerID)
		g.sanitizeDockedFleets()
		return result
	}
	g.retreatArmiesFromCapturedRegion(targetRegion.ID, newOwnerID)
	g.evictDockedFleetsFromCapturedPort(targetRegion.ID, newOwnerID)
	return eliminationResult{}
}

func (g *Game) announceElimination(result eliminationResult) {
	if g == nil || g.renderer == nil || result.FactionID == "" {
		return
	}
	defeatedName := string(result.FactionID)
	if f := g.gs.Factions[result.FactionID]; f != nil && f.NameTR != "" {
		defeatedName = f.NameTR
	}
	msg := defeatedName + " devleti yıkıldı."
	if result.SuccessorID != "" {
		successorName := string(result.SuccessorID)
		if f := g.gs.Factions[result.SuccessorID]; f != nil && f.NameTR != "" {
			successorName = f.NameTR
		}
		msg = fmt.Sprintf("%s devleti yıkıldı. %s kalan kuvvetleri devraldı.", defeatedName, successorName)
	}
	g.renderer.ShowCombatResult(msg)
	detail := msg
	if result.TransferredArmies > 0 || result.TransferredFleets > 0 {
		detail = fmt.Sprintf("%s Devralınan kuvvetler: %d kara ordusu, %d donanma.", msg, result.TransferredArmies, result.TransferredFleets)
	}
	g.renderer.AddEventDetail("[YIKILIS] "+msg, detail)
}

func (g *Game) retreatArmiesFromCapturedRegion(capturedRegionID world.RegionID, protectedOwnerID string) {
	if g == nil || g.gs == nil {
		return
	}
	capturedRegion := g.gs.Regions[capturedRegionID]
	if capturedRegion == nil {
		return
	}
	for _, a := range g.gs.Armies {
		if a == nil || a.RegionID != capturedRegionID || a.OwnerID == "" || a.OwnerID == protectedOwnerID || a.IsNaval {
			continue
		}
		if retreatRegion := g.nearestOwnedRegionForArmy(a, capturedRegion); retreatRegion != "" {
			a.RegionID = retreatRegion
			a.DockedRegionID = ""
			a.DockedSettlementID = ""
		}
	}
}

func (g *Game) nearestOwnedRegionForArmy(a *army.Army, reference *world.Region) world.RegionID {
	if g == nil || g.gs == nil || a == nil || reference == nil || a.OwnerID == "" {
		return ""
	}
	bestRegion := world.RegionID("")
	bestDist := 0.0
	found := false
	for _, region := range g.gs.Regions {
		if region == nil || region.OwnerID != a.OwnerID {
			continue
		}
		if a.IsNaval != region.IsSea {
			continue
		}
		if region.ID == reference.ID {
			continue
		}
		dx := float64(region.WorldX - reference.WorldX)
		dy := float64(region.WorldY - reference.WorldY)
		dist := dx*dx + dy*dy
		if !found || dist < bestDist || (dist == bestDist && region.ID < bestRegion) {
			bestRegion = region.ID
			bestDist = dist
			found = true
		}
	}
	return bestRegion
}

func (g *Game) evictDockedFleetsFromCapturedPort(capturedRegionID world.RegionID, protectedOwnerID string) {
	for _, fleet := range g.gs.Armies {
		if fleet == nil || !fleet.IsNaval || fleet.OwnerID == "" || fleet.OwnerID == protectedOwnerID || fleet.DockedRegionID != capturedRegionID {
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
		invalidDock := dockedRegion == nil || dockedRegion.IsSea || !g.canDockFleetAtRegion(fleet, dockedRegion)
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

// sanitizeOccupiedNeutralRegions eski bug'lı save'lerde veya edge-case akışlarda
// sahipsiz kalmış ama tek taraflı kara ordusunca tutulduğu açık olan kara
// bölgelerini fiilen işgal eden fraksiyona yazar.
func (g *Game) sanitizeOccupiedNeutralRegions() {
	if g == nil || g.gs == nil {
		return
	}
	claimants := make(map[world.RegionID]string)
	contested := make(map[world.RegionID]bool)
	for _, landArmy := range g.gs.Armies {
		if landArmy == nil || landArmy.IsNaval || landArmy.RegionID == "" {
			continue
		}
		region := g.gs.Regions[landArmy.RegionID]
		if region == nil || region.IsSea || region.OwnerID != "" {
			continue
		}
		if ownerID, ok := claimants[landArmy.RegionID]; ok && ownerID != landArmy.OwnerID {
			contested[landArmy.RegionID] = true
			continue
		}
		claimants[landArmy.RegionID] = landArmy.OwnerID
	}
	for regionID, ownerID := range claimants {
		if ownerID == "" || contested[regionID] {
			continue
		}
		region := g.gs.Regions[regionID]
		if region == nil || region.OwnerID != "" || region.IsSea {
			continue
		}
		region.ApplyConquest(ownerID, ownerReligion(g.gs, ownerID))
	}
}

// moveArmy oyuncu ordusunu hedef bölgeye taşır; gerekirse savaş başlatır.
func (g *Game) moveArmy(aid army.ArmyID, target world.RegionID) {
	g.moveArmyWithStance(aid, target, combat.BattleStanceBalanced)
}

// moveArmyWithStance oyuncu ordusunu hedef bölgeye taşır; savaş çıkarsa seçilen saldırı duruşunu uygular.
func (g *Game) moveArmyWithStance(aid army.ArmyID, target world.RegionID, battleStance combat.BattleStance) {
	battleStance = combat.NormalizeBattleStance(battleStance)
	a, ok := g.gs.Armies[aid]
	if !ok || a.OwnerID != string(g.gs.PlayerFactionID) {
		return
	}
	aid = g.deployGarrisonArmy(aid)
	a = g.gs.Armies[aid]
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
			shouldDisembark := len(a.EmbarkedUnits) > 0 &&
				(a.DockedRegionID == targetRegion.ID || !g.canDockFleetAtRegion(a, targetRegion))
			if shouldDisembark {
				g.resolveFleetDisembarkWithStance(a, target, targetRegion, battleStance)
				return
			}
		}
		if g.canDockFleetAtRegion(a, targetRegion) {
			a.DockedRegionID = targetRegion.ID
			a.DockedSettlementID = g.dockSettlementIDForRegion(targetRegion)
			a.MovePoints--
			g.renderer.MarkMapDirty()
			g.renderer.ShowCombatResult("Donanma limana konuşlandı.")
			return
		}
		if targetRegion.CanLandEnter() {
			if len(a.EmbarkedUnits) == 0 {
				g.renderer.ShowCombatResult("Çıkarma emri reddedildi: filoda taşınan kara birimi yok.")
			} else {
				g.renderer.ShowCombatResult("Çıkarma emri reddedildi: düşman kıyıya çıkarmak için savaş halinde olmalısın.")
			}
			return
		}
		if !targetRegion.CanNavalEnter() {
			g.renderer.ShowCombatResult("Deniz ordusu sadece deniz bölgelerine gidebilir!")
			return
		}
	} else {
		if targetRegion.CanNavalEnter() {
			if !g.canEmbarkLandArmy(a) {
				g.renderer.ShowCombatResult(g.embarkBlockedMessage(a))
				return
			}
			fleet := g.findFriendlyEmbarkFleetFromRegion(a.OwnerID, a.RegionID, target, len(a.Units))
			if fleet == nil {
				if transportFleet := g.findFriendlyTransportFleet(a.OwnerID, target); transportFleet != nil {
					g.renderer.ShowCombatResult(fmt.Sprintf("Nakliye kapasitesi yetersiz: %d/%d boş slot.", transportFleet.AvailableTransportCapacity(g.gs.UnitTypes), len(a.Units)))
				} else {
					g.renderer.ShowCombatResult("Embark için komşu denizde uygun nakliye filosu yok!")
				}
				return
			}
			fleet.EmbarkedUnits = append(fleet.EmbarkedUnits, a.Units...)
			fleet.MovePoints = max(0, fleet.MovePoints-1)
			delete(g.gs.Armies, aid)
			g.renderer.SelectedArmy = fleet.ID
			g.renderer.ShowCombatResult(fmt.Sprintf("Ordu nakliye filosuna bindi. Kalan kapasite: %d.", fleet.AvailableTransportCapacity(g.gs.UnitTypes)))
			return
		}
		if !targetRegion.CanLandEnter() {
			g.renderer.ShowCombatResult("Kara ordusu denize giremez! (Nakliye gemisi gerekir)")
			return
		}
	}
	// Sahipli yabancı kara bölgesine girmek için savaş, ittifak veya aynı vassal
	// zincirinde askeri geçiş hakkı gerekir.
	// Donanma-deniz hareketinde bu kural uygulanmaz; denizde serbest dolaşım var.
	isAlliedRegion := false
	if !navalSeaMove && targetRegion.OwnerID != "" && targetRegion.OwnerID != a.OwnerID {
		if diplomacy.SameRealm(g.gs, faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID)) {
			isAlliedRegion = true
		}
		key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID))
		rel, exists := g.gs.Relations[key]
		if !isAlliedRegion && (!exists || (rel.Stance != faction.StanceWar && rel.Stance != faction.StanceAllied)) {
			g.renderer.ShowCombatResult("Savaş ilanı, ittifak veya bağlı devlet askeri geçişi olmadan yabancı toprağa girilemez!")
			return
		}
		if exists && rel.Stance == faction.StanceAllied {
			isAlliedRegion = true
		}
	}
	liftedSiegeRegion := world.RegionID("")
	if activeSiege := g.gs.SiegeByArmy(aid); activeSiege != nil && activeSiege.RegionID != target {
		liftedSiegeRegion = activeSiege.RegionID
	}
	allyJoiningSiege := false
	if !a.IsNaval && targetRegion.IsFortified() && targetRegion.OwnerID != "" && targetRegion.OwnerID != a.OwnerID {
		activeSiege := g.gs.SiegeAt(target)
		if activeSiege != nil && activeSiege.AttackerArmyID == aid {
			g.renderer.ShowCombatResult("Bu tahkimata girmek için kuşatma üzerinden genel hücum seçmelisin.")
			return
		}
		if activeSiege != nil {
			if g.gs.CanJoinActiveSiege(a, target) {
				allyJoiningSiege = true
			} else {
				g.renderer.ShowCombatResult("Bu bölge zaten başka bir ordu tarafından kuşatılıyor.")
				return
			}
		} else {
			if ok, reason := canArmyStartSiege(g.gs, a, targetRegion); !ok {
				if reason != "" {
					g.renderer.ShowCombatResult(reason)
				}
				return
			}
			g.renderer.ShowCombatResult("Bu bölge tahkimli. Önce kuşatma başlatmalı veya genel hücum seçmelisin.")
			return
		}
	}

	enemyArmy := g.gs.SelectBattleDefender(a, target, navalSeaMove)

	if enemyArmy != nil && !allyJoiningSiege {
		// --- Savaş ---
		// Bölgedeki TÜM düşman orduları (müttefikler dahil) birleşik savunur
		combinedDef, defSourceIDs := g.gs.CollectDefenders(a, target, navalSeaMove)
		if combinedDef == nil {
			combinedDef = enemyArmy
		}
		attackerBefore := snapshotBattleArmy(a, g.gs.UnitTypes)
		defenderBefore := snapshotBattleArmy(combinedDef, g.gs.UnitTypes)
		if liftedSiegeRegion != "" {
			g.clearSiege(liftedSiegeRegion)
		}
		atkMods := techModsFor(g.gs, a.OwnerID)
		// Savunma modlarını bölge sahibinden al (birleşik orduda ilk ordu sahibi referans)
		defOwnerID := enemyArmy.OwnerID
		defMods := techModsFor(g.gs, defOwnerID)
		battleContext := combat.BattleContextLand
		if navalSeaMove {
			battleContext = combat.BattleContextNaval
		}
		result := combat.ResolveBattleWithContextPlan(a, combinedDef, targetRegion.Terrain, g.gs.UnitTypes, atkMods, defMods, battleContext, battleStance)
		var collapse eliminationResult
		scene := render.BattleSceneLand
		attackerLabel := "Saldıran Ordu"
		defenderLabel := "Savunma Hattı"
		if navalSeaMove {
			scene = render.BattleSceneNaval
			attackerLabel = "Taarruz Filosu"
			defenderLabel = "Savunma Filosu"
		}
		defenderFaction := g.factionNameTR(enemyArmy.OwnerID)
		if len(defSourceIDs) > 1 {
			defenderFaction = fmt.Sprintf("Birleşik Savunma (%d ordu)", len(defSourceIDs))
		}
		outcomeDetail := "Saldırı püskürtüldü."

		if result.AttackerWins {
			if len(defSourceIDs) > 0 {
				g.gs.DistributeDefenderLosses(defSourceIDs, result.DefenderLost)
			} else if len(enemyArmy.Units) == 0 {
				delete(g.gs.Armies, enemyArmy.ID)
			}
			if len(a.Units) > 0 {
				a.RegionID = target
				a.DockedRegionID = ""
				a.DockedSettlementID = ""
				a.MovePoints--
				prompted := g.queueConquestDecision(faction.FactionID(a.OwnerID), targetRegion, true)
				if !prompted {
					collapse = g.applyConquestWithNavalEviction(targetRegion, a.OwnerID)
					g.renderer.MarkMapDirty()
				}
				if navalSeaMove {
					if prompted {
						outcomeDetail = "Düşman filo dağıtıldı; teslim şartları için savaş sonrası karar bekleniyor."
					} else {
						outcomeDetail = "Düşman filo dağıtıldı ve deniz hattı açıldı."
					}
				} else {
					if prompted {
						outcomeDetail = "Savunma yarıldı; ilhak ya da vassallık için savaş sonrası karar bekleniyor."
					} else {
						outcomeDetail = "Savunma yarıldı, bölge ele geçirildi."
					}
				}
			} else {
				delete(g.gs.Armies, aid)
				if navalSeaMove {
					outcomeDetail = "Düşman filo dağıldı fakat taarruz filosu da dağıldı."
				} else {
					outcomeDetail = "Savunma çöktü fakat saldırı gücü tükendi; bölge elde tutulamadı."
				}
			}
		} else {
			// Saldıran yenildi — yerinde kalır
			if len(a.Units) == 0 {
				delete(g.gs.Armies, aid)
			}
			if navalSeaMove {
				outcomeDetail = "Taarruz filosu geri çekildi."
			}
		}

		g.presentBattleReport(g.makeBattleReport(
			scene,
			targetRegion.NameTR,
			battleStance,
			result.Description,
			outcomeDetail,
			attackerLabel,
			defenderLabel,
			g.factionNameTR(a.OwnerID),
			defenderFaction,
			attackerBefore,
			a,
			defenderBefore,
			combinedDef,
		))
		g.announceElimination(collapse)

	} else {
		// --- Savaşsız hareket ve bölge ele geçirme ---
		attackerBefore := snapshotBattleArmy(a, g.gs.UnitTypes)
		if liftedSiegeRegion != "" {
			g.clearSiege(liftedSiegeRegion)
		}
		a.RegionID = target
		a.DockedRegionID = ""
		a.DockedSettlementID = ""
		a.MovePoints--
		if allyJoiningSiege {
			// Kuşatmaya katılım: bölge fethedilmez, sadece birleşme yapılır.
			if merged := g.tryMergeArmies(aid, target); merged != "" {
				g.renderer.SelectedArmy = merged
			}
			g.renderer.ShowCombatResult("Ordu kuşatmaya katıldı.")
			return
		}
		// Müttefik bölgesi fethedilemez, sadece içinden geçilir.
		if !targetRegion.IsSea && targetRegion.OwnerID != a.OwnerID && !isAlliedRegion {
			defenderFaction := g.factionNameTR(targetRegion.OwnerID)
			prompted := g.queueConquestDecision(faction.FactionID(a.OwnerID), targetRegion, true)
			collapse := eliminationResult{}
			detail := "Bölgede savunan ordu yoktu; ilerleme savaşsız şekilde tamamlandı ve bölge ele geçirildi."
			if !prompted {
				collapse = g.applyConquestWithNavalEviction(targetRegion, a.OwnerID)
				g.renderer.MarkMapDirty()
			} else {
				detail = "Bölgede savunan ordu yoktu; devletin geleceği için ilhak veya vassallık kararı bekleniyor."
			}
			g.presentBattleReport(g.makeBattleReportFromSnapshots(
				render.BattleSceneLand,
				targetRegion.NameTR,
				"",
				"Direniş Görülmedi",
				detail,
				"İlerleyen Ordu",
				"Direniş Yok",
				g.factionNameTR(a.OwnerID),
				defenderFaction,
				attackerBefore,
				snapshotBattleArmy(a, g.gs.UnitTypes),
				battleArmySnapshot{},
				battleArmySnapshot{},
			))
			g.announceElimination(collapse)
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
	movingID = g.deployGarrisonArmy(movingID)
	moving, ok := g.gs.Armies[movingID]
	if !ok {
		return ""
	}
	var targetID army.ArmyID
	for otherID, other := range g.gs.Armies {
		if otherID == movingID || other.RegionID != regionID ||
			other.OwnerID != moving.OwnerID || other.IsNaval != moving.IsNaval {
			continue
		}
		targetID = otherID
		break
	}
	if targetID == "" {
		return ""
	}
	targetID = g.deployGarrisonArmy(targetID)
	target := g.gs.Armies[targetID]
	if target == nil || len(moving.Units)+len(target.Units) > army.MaxArmySize {
		return ""
	}
	target.Units = append(target.Units, moving.Units...)
	delete(g.gs.Armies, movingID)
	g.renderer.AddEvent("Ordular birleşti: " + fmt.Sprintf("%d", len(target.Units)) + " birim")
	return targetID
}

// splitArmy seçili orduyu birim sayısına göre ikiye böler.
func (g *Game) splitArmy(aid army.ArmyID) {
	aid = g.deployGarrisonArmy(aid)
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
		TurnsWithoutPort:   a.TurnsWithoutPort,
	}
	g.renderer.AddEvent(fmt.Sprintf("Ordu bölündü: %d + %d birim", len(a.Units), len(newUnits)))
}

// mergeArmiesManual seçili orduyu aynı bölgedeki dost orduya elle birleştirir (20 kapasitesine kadar).
func (g *Game) mergeArmiesManual(aid army.ArmyID) {
	aid = g.deployGarrisonArmy(aid)
	a, ok := g.gs.Armies[aid]
	if !ok {
		return
	}
	// Aynı bölgede dost ordu bul
	var targetID army.ArmyID
	for oid, other := range g.gs.Armies {
		if oid == aid || other.RegionID != a.RegionID ||
			other.OwnerID != a.OwnerID || other.IsNaval != a.IsNaval {
			continue
		}
		targetID = oid
		break
	}
	if targetID == "" {
		return
	}
	targetID = g.deployGarrisonArmy(targetID)
	target := g.gs.Armies[targetID]
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

func (g *Game) deployGarrisonArmy(aid army.ArmyID) army.ArmyID {
	if g == nil || g.gs == nil {
		return aid
	}
	a := g.gs.Armies[aid]
	if a == nil || a.IsNaval || !a.IsGarrison {
		return aid
	}
	a.IsGarrison = false
	if !army.LooksLikeGarrisonID(aid) {
		return aid
	}
	g.gs.NextArmySeq++
	newID := army.ArmyID(fmt.Sprintf("army_%s_%d", a.OwnerID, g.gs.NextArmySeq))
	delete(g.gs.Armies, aid)
	a.ID = newID
	g.gs.Armies[newID] = a
	if g.gs.ArmyLogistics != nil {
		if status, ok := g.gs.ArmyLogistics[aid]; ok {
			delete(g.gs.ArmyLogistics, aid)
			status.ArmyID = newID
			g.gs.ArmyLogistics[newID] = status
		}
	}
	for _, siege := range g.gs.Sieges {
		if siege == nil {
			continue
		}
		if siege.AttackerArmyID == aid {
			siege.AttackerArmyID = newID
		}
		if siege.DefenderArmyID == aid {
			siege.DefenderArmyID = newID
		}
	}
	if g.renderer != nil && g.renderer.SelectedArmy == aid {
		g.renderer.SelectedArmy = newID
	}
	return newID
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

	regionTargets := opt.RegionTargets()
	requiredRegions := make([]world.RegionID, len(regionTargets))
	for i, r := range regionTargets {
		requiredRegions[i] = world.RegionID(r)
	}

	g.gs.Victory = state.VictoryCondition{
		Type:               state.VictoryType(opt.Type),
		TargetRegionCount:  opt.TargetRegionCount,
		RequiredRegions:    requiredRegions,
		TargetGoldIncome:   opt.TargetGoldIncome,
		GoldHoldTurns:      opt.GoldHoldTurns,
		TargetArmyStrength: opt.TargetArmyStrength,
		TargetDefeated:     opt.TargetDefeated,
		TargetTurns:        opt.Turns,
		DeadlineYear:       opt.DeadlineYear,
		DeadlineMonth:      opt.DeadlineMonth,
	}
	g.gs.SelectedVictoryOptionID = opt.ID
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
	if f.Research.ActiveID == techID {
		g.renderer.ShowCombatResult(t.NameTR + " zaten araştırılıyor.")
		return
	}
	canResume := f.Research.PausedTurns != nil && f.Research.PausedTurns[techID] > 0
	if !tech.IsUnlocked(&f.Research, t) || (f.Research.Completed != nil && f.Research.Completed[techID]) {
		g.renderer.ShowCombatResult("Araştırma başlatılamadı. Altın veya gereksinim eksik.")
		return
	}
	if !canResume && f.Gold < t.GoldCost {
		g.renderer.ShowCombatResult("Araştırma başlatılamadı. Altın veya gereksinim eksik.")
		return
	}
	if f.Research.ActiveID != "" {
		tech.PauseResearch(&f.Research)
	}
	if tech.StartResearch(&f.Research, t, &f.Gold) {
		if canResume {
			g.renderer.ShowCombatResult(t.NameTR + " araştırması kaldığı yerden devam ediyor! (" + fmt.Sprintf("%d tur kaldı", f.Research.TurnsLeft) + ")")
		} else {
			g.renderer.ShowCombatResult(t.NameTR + " araştırması başladı! (" + fmt.Sprintf("%d tur", t.TurnsRequired) + ")")
		}
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
	activeID := f.Research.ActiveID
	tech.PauseResearch(&f.Research)
	if t, ok := g.gs.TechTypes[activeID]; ok {
		g.renderer.ShowCombatResult(fmt.Sprintf("%s araştırması duraklatıldı. Kaldığı yerden devam edebilirsin.", t.NameTR))
		return
	}
	f.Research.ActiveID = ""
	f.Research.TurnsLeft = 0
	g.renderer.ShowCombatResult("Araştırma duraklatıldı.")
}
