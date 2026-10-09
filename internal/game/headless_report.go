package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/events"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const defaultHeadlessScenario = "assets/scenarios/1300_ottoman_rise"
const defaultHeadlessSeed uint64 = 13100301

// HeadlessSimulationOptions, pencere açmadan AI kampanyası çalıştırmak için
// kullanılan ayarları taşır.
type HeadlessSimulationOptions struct {
	ScenarioPath    string
	PlayerFactionID faction.FactionID
	Turns           int
	Seed            uint64
	Difficulty      int
	Progress        func(completedTurns, totalTurns int)
}

// HeadlessSimulationReport, headless kampanyanın ölçülebilir sonuçlarını taşır.
type HeadlessSimulationReport struct {
	ScenarioID      string                   `json:"scenario_id"`
	ScenarioPath    string                   `json:"scenario_path"`
	PlayerFactionID faction.FactionID        `json:"player_faction_id"`
	RequestedTurns  int                      `json:"requested_turns"`
	CompletedTurns  int                      `json:"completed_turns"`
	StartTurn       int                      `json:"start_turn"`
	EndTurn         int                      `json:"end_turn"`
	StartYear       int                      `json:"start_year"`
	StartMonth      int                      `json:"start_month"`
	EndYear         int                      `json:"end_year"`
	EndMonth        int                      `json:"end_month"`
	Seed            uint64                   `json:"seed"`
	Difficulty      int                      `json:"difficulty"`
	DecisionPolicy  string                   `json:"decision_policy"`
	StopReason      string                   `json:"stop_reason,omitempty"`
	Factions        []HeadlessFactionSummary `json:"factions"`
	Wars            []HeadlessWarSummary     `json:"wars"`
	TradeChanges    []HeadlessDiplomacyEvent `json:"trade_changes"`
	AllianceChanges []HeadlessDiplomacyEvent `json:"alliance_changes"`
	Eliminations    []HeadlessElimination    `json:"eliminations"`
	Events          []HeadlessEvent          `json:"events"`
	Checkpoints     []HeadlessCheckpoint     `json:"checkpoints"`
}

type HeadlessFactionSummary struct {
	ID                    faction.FactionID `json:"id"`
	NameTR                string            `json:"name_tr"`
	InitialRegions        int               `json:"initial_regions"`
	FinalRegions          int               `json:"final_regions"`
	RegionsGained         []string          `json:"regions_gained"`
	RegionsLost           []string          `json:"regions_lost"`
	InitialArmies         int               `json:"initial_armies"`
	FinalArmies           int               `json:"final_armies"`
	InitialLandUnits      int               `json:"initial_land_units"`
	FinalLandUnits        int               `json:"final_land_units"`
	InitialNavalUnits     int               `json:"initial_naval_units"`
	FinalNavalUnits       int               `json:"final_naval_units"`
	ProductionOrders      int               `json:"production_orders_queued"`
	BuildingOrders        int               `json:"building_orders_queued"`
	InitialBuildings      int               `json:"initial_buildings"`
	FinalBuildings        int               `json:"final_buildings"`
	TechnologiesCompleted int               `json:"technologies_completed"`
	InitialGold           int               `json:"initial_gold"`
	FinalGold             int               `json:"final_gold"`
	InitialGrain          int               `json:"initial_grain"`
	FinalGrain            int               `json:"final_grain"`
	Eliminated            bool              `json:"eliminated"`
}

type HeadlessWarSummary struct {
	FactionA         faction.FactionID `json:"faction_a"`
	FactionB         faction.FactionID `json:"faction_b"`
	StartedTurn      int               `json:"started_turn"`
	EndedTurn        int               `json:"ended_turn,omitempty"`
	Active           bool              `json:"active"`
	RegionsCapturedA int               `json:"regions_captured_a"`
	RegionsCapturedB int               `json:"regions_captured_b"`
	CasualtiesA      int               `json:"casualties_a"`
	CasualtiesB      int               `json:"casualties_b"`
}

type HeadlessDiplomacyEvent struct {
	Turn     int               `json:"turn"`
	Action   string            `json:"action"`
	FactionA faction.FactionID `json:"faction_a"`
	FactionB faction.FactionID `json:"faction_b"`
	Good     string            `json:"good,omitempty"`
}

type HeadlessElimination struct {
	Turn    int               `json:"turn"`
	Faction faction.FactionID `json:"faction_id"`
	NameTR  string            `json:"name_tr"`
}

type HeadlessEvent struct {
	Turn   int    `json:"turn"`
	ID     string `json:"id"`
	NameTR string `json:"name_tr"`
}

type HeadlessCheckpoint struct {
	Turn     int                      `json:"turn"`
	Year     int                      `json:"year"`
	Month    int                      `json:"month"`
	Rankings []HeadlessFactionRanking `json:"rankings"`
}

type HeadlessFactionRanking struct {
	FactionID     faction.FactionID `json:"faction_id"`
	NameTR        string            `json:"name_tr"`
	Regions       int               `json:"regions"`
	Armies        int               `json:"armies"`
	LandUnits     int               `json:"land_units"`
	NavalUnits    int               `json:"naval_units"`
	MilitaryPower int               `json:"military_power"`
}

type headlessFactionSnapshot struct {
	Regions    map[world.RegionID]string
	Armies     int
	LandUnits  int
	NavalUnits int
	Buildings  int
	Techs      int
	Gold       int
	Grain      int
	Eliminated bool
}

type headlessWarSnapshot struct {
	Active      bool
	StartedTurn int
}

type headlessTradeSnapshot struct {
	From faction.FactionID
	To   faction.FactionID
	Good string
}

// RunHeadlessSimulation, oyunun AI ve tur çözümleme akışını pencere açmadan
// çalıştırır. Oyuncu devleti de AI'ye devredilir; bekleyen modal kararları
// deterministik bir politika ile otomatik sonuçlandırılır.
func RunHeadlessSimulation(options HeadlessSimulationOptions) (*HeadlessSimulationReport, error) {
	if options.ScenarioPath == "" {
		options.ScenarioPath = defaultHeadlessScenario
	}
	if options.Turns <= 0 {
		options.Turns = 100
	}
	if options.Seed == 0 {
		options.Seed = defaultHeadlessSeed
	}
	if options.Difficulty < 1 || options.Difficulty > 3 {
		options.Difficulty = 2
	}
	rand.Seed(int64(options.Seed))

	gs, eventDefs, err := loadScenarioDataForMode(options.ScenarioPath, options.Difficulty, false, nil)
	if err != nil {
		return nil, err
	}
	if options.PlayerFactionID == "" {
		options.PlayerFactionID = firstPlayableFaction(gs)
	}
	if gs.Factions[options.PlayerFactionID] == nil {
		return nil, fmt.Errorf("oyuncu devleti bulunamadı: %s", options.PlayerFactionID)
	}

	gs.PlayerFactionID = options.PlayerFactionID
	gs.DecisionSeed = options.Seed
	gs.AIControlsPlayerFaction = true
	gs.AIControlsPlayerEconomy = true
	gs.Phase = state.PhasePlayerTurn
	gs.RefreshArmyMovePoints(true)
	gs.SyncWarLedgers()

	// Renderer yalnızca çözümleme sırasında kullanılan olay/uyarı state'ini
	// taşır; burada Draw veya Ebitengine penceresi çalıştırılmaz.
	g := &Game{
		gs:       gs,
		evts:     eventDefs,
		renderer: render.New(gs),
	}
	audio.SetSoundEnabled(false)
	refreshMarketOrdersAndPricesWithEvents(gs, eventDefs)

	report := &HeadlessSimulationReport{
		ScenarioID:      gs.ScenarioID,
		ScenarioPath:    gs.ScenarioPath,
		PlayerFactionID: gs.PlayerFactionID,
		RequestedTurns:  options.Turns,
		StartTurn:       gs.Turn,
		StartYear:       gs.Year,
		StartMonth:      gs.Month,
		Seed:            options.Seed,
		Difficulty:      options.Difficulty,
		DecisionPolicy:  "Tüm devletler AI; AI, muharebe ve rastgele olay zarları aynı seed ile başlatılır; oyuncuya gelen barış/ittifak/ticaret/savaş çağrıları kabul edilir, teslimiyet ve kuşatma vassallığı teklifleri reddedilir; olay seçeneklerinde senaryo otomatik seçimi kullanılır.",
	}

	initial := captureHeadlessFactionSnapshots(gs)
	for _, fid := range sortedFactionIDs(gs) {
		f := gs.Factions[fid]
		if f == nil {
			continue
		}
		report.Factions = append(report.Factions, HeadlessFactionSummary{
			ID:     fid,
			NameTR: f.NameTR,
		})
	}
	previousRegions := captureRegionOwners(gs)
	previousWars := captureWarSnapshots(gs)
	previousTrades := captureTradeSnapshots(gs)
	previousAlliances := captureAllianceSnapshots(gs)
	seenEvents := make(map[string]bool, len(gs.FiredEventIDs))
	for id := range gs.FiredEventIDs {
		seenEvents[id] = true
	}

	for completed := 0; completed < options.Turns && gs.Phase != state.PhaseGameOver; completed++ {
		turn := gs.Turn
		preparations := refreshMarketOrdersAndPricesWithEvents(gs, eventDefs)
		for _, fid := range headlessFactionOrder(gs) {
			f := gs.Factions[fid]
			if f == nil || f.IsEliminated {
				continue
			}
			queueBefore := append([]state.ProductionOrder(nil), gs.ProductionQueue...)
			stepper := ai.NewTurnStepperWithPreparedContextAndEvents(gs, fid, eventDefs, preparations[fid])
			for {
				step, done := stepper.Step()
				if done {
					break
				}
				if step.Kind == ai.TurnStepSortie {
					g.presentAISortieDecision(step)
					if g.pendingSortie != nil {
						g.resolvePendingSortie(true)
					}
				}
				resolveHeadlessContacts(gs)
				resolveHeadlessOffers(g)
			}
			countHeadlessProductionOrders(report, queueBefore, gs.ProductionQueue, fid)
		}
		resolveHeadlessContacts(gs)
		resolveHeadlessOffers(g)
		observeHeadlessWarChanges(report, gs, previousWars, turn)
		observeHeadlessDiplomacyChanges(report, gs, previousTrades, previousAlliances, turn)

		g.resolveTurn()
		if g.pendingHistoricalEvt != nil {
			originalChoice := events.AutoChooseForFaction(g.pendingHistoricalEvt, string(gs.PlayerFactionID))
			visibleChoices := events.ChoiceIndicesForFaction(g.pendingHistoricalEvt, string(gs.PlayerFactionID))
			visibleChoice := -1
			for i, choiceIndex := range visibleChoices {
				if choiceIndex == originalChoice {
					visibleChoice = i
					break
				}
			}
			if visibleChoice >= 0 {
				g.resolveHistoricalChoice(visibleChoice)
			}
		}
		resolveHeadlessContacts(gs)
		resolveHeadlessOffers(g)
		closeHeadlessWarChanges(report, gs, turn)

		for id, evt := range eventDefinitionsByID(eventDefs) {
			if gs.FiredEventIDs[id] && !seenEvents[id] {
				report.Events = append(report.Events, HeadlessEvent{Turn: turn, ID: id, NameTR: evt.NameTR})
				seenEvents[id] = true
			}
		}
		recordHeadlessRegionChanges(report, gs, previousRegions, turn)
		recordHeadlessEliminations(report, gs, initial, turn)
		previousRegions = captureRegionOwners(gs)
		previousWars = captureWarSnapshots(gs)
		previousTrades = captureTradeSnapshots(gs)
		previousAlliances = captureAllianceSnapshots(gs)
		if gs.Turn%10 == 0 || completed == options.Turns-1 || gs.Phase == state.PhaseGameOver {
			report.Checkpoints = append(report.Checkpoints, captureHeadlessCheckpoint(gs, completed+1))
		}
		if options.Progress != nil {
			options.Progress(completed+1, options.Turns)
		}
	}

	report.CompletedTurns = gs.Turn - report.StartTurn
	report.EndTurn = gs.Turn
	report.EndYear = gs.Year
	report.EndMonth = gs.Month
	if gs.Phase == state.PhaseGameOver {
		report.StopReason = "oyun_bitti"
	} else if report.CompletedTurns < options.Turns {
		report.StopReason = "simülasyon_durduruldu"
	}
	report.Factions = finalizeHeadlessFactionSummaries(gs, initial, report)
	report.Wars = finalizeHeadlessWars(gs, report)
	sort.Slice(report.Events, func(i, j int) bool {
		if report.Events[i].Turn != report.Events[j].Turn {
			return report.Events[i].Turn < report.Events[j].Turn
		}
		return report.Events[i].ID < report.Events[j].ID
	})
	sort.Slice(report.TradeChanges, func(i, j int) bool {
		return headlessDiplomacyEventLess(report.TradeChanges[i], report.TradeChanges[j])
	})
	sort.Slice(report.AllianceChanges, func(i, j int) bool {
		return headlessDiplomacyEventLess(report.AllianceChanges[i], report.AllianceChanges[j])
	})
	return report, nil
}

func headlessDiplomacyEventLess(left, right HeadlessDiplomacyEvent) bool {
	if left.Turn != right.Turn {
		return left.Turn < right.Turn
	}
	if left.Action != right.Action {
		return left.Action < right.Action
	}
	if left.FactionA != right.FactionA {
		return left.FactionA < right.FactionA
	}
	if left.FactionB != right.FactionB {
		return left.FactionB < right.FactionB
	}
	return left.Good < right.Good
}

func firstPlayableFaction(gs *state.GameState) faction.FactionID {
	for _, fid := range gs.FactionOrder {
		if f := gs.Factions[fid]; f != nil && f.IsPlayable && !f.IsVirtual {
			return fid
		}
	}
	ids := make([]faction.FactionID, 0, len(gs.Factions))
	for fid := range gs.Factions {
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	if len(ids) > 0 {
		return ids[0]
	}
	return ""
}

func headlessFactionOrder(gs *state.GameState) []faction.FactionID {
	order := make([]faction.FactionID, 0, len(gs.Factions))
	seen := make(map[faction.FactionID]bool, len(gs.Factions))
	for _, fid := range gs.FactionOrder {
		if f := gs.Factions[fid]; f != nil && !f.IsVirtual {
			order = append(order, fid)
			seen[fid] = true
		}
	}
	extra := make([]faction.FactionID, 0)
	for fid, f := range gs.Factions {
		if f != nil && !f.IsVirtual && !seen[fid] {
			extra = append(extra, fid)
		}
	}
	sort.Slice(extra, func(i, j int) bool { return extra[i] < extra[j] })
	return append(order, extra...)
}

func resolveHeadlessOffers(g *Game) {
	if g == nil || g.gs == nil {
		return
	}
	for i := 0; i < len(g.gs.DiplomaticOffers); {
		offer := g.gs.DiplomaticOffers[i]
		if offer.ToFactionID != g.gs.PlayerFactionID {
			i++
			continue
		}
		accepted := true
		switch diplomacy.Action(offer.Action) {
		case diplomacy.ActionProposeSurrender, diplomacy.ActionProposeSiegeVassalization:
			accepted = false
		}
		if isSiegeSettlementOfferAction(offer.Action) {
			g.resolveDiplomacyOffer(i, accepted)
		} else {
			diplomacy.ResolveOffer(g.gs, i, accepted)
		}
	}
}

func resolveHeadlessContacts(gs *state.GameState) {
	if gs == nil {
		return
	}
	if contact := gs.PendingNavalContact; contact != nil {
		ai.ResolveAIOnlyNavalContact(gs, contact)
	}
	if contact := gs.PendingLandContact; contact != nil {
		attacker := gs.Armies[contact.AttackerArmyID]
		if attacker == nil {
			gs.ClearLandContact()
			return
		}
		if contact.AttackerDecision == state.LandContactUndecided || contact.DefenderDecision == state.LandContactUndecided {
			// Tüm taraflar AI olduğundan normal karar helper'ı temasın iki
			// tarafını da doldurur; eski state'te kalan temas da kilitlenmez.
			ai.ResolveLandContactDecision(gs, contact)
		}
		if gs.LandContactWillClash(contact) {
			ai.ResolveLandContactBattle(gs, attacker.ID, contact.LandRegionID, contact.MovementConsumed, contact.AttackerDecision == state.LandContactHold, contact.DefenderDecision == state.LandContactHold)
		} else {
			ai.ResolveLandContactWithoutBattle(gs, contact, gs.Armies[contact.DefenderArmyID])
			gs.ClearLandContact()
		}
	}
}

func captureHeadlessFactionSnapshots(gs *state.GameState) map[faction.FactionID]headlessFactionSnapshot {
	snapshots := make(map[faction.FactionID]headlessFactionSnapshot, len(gs.Factions))
	for fid, f := range gs.Factions {
		if f == nil {
			continue
		}
		regions := make(map[world.RegionID]string)
		for _, r := range gs.LandRegionsOwnedBy(fid) {
			regions[r.ID] = r.NameTR
		}
		buildings := 0
		for _, r := range gs.LandRegionsOwnedBy(fid) {
			buildings += len(r.Buildings)
		}
		armies := 0
		landUnits := 0
		navalUnits := 0
		for _, a := range gs.Armies {
			if a == nil || faction.FactionID(a.OwnerID) != fid {
				continue
			}
			armies++
			if a.IsNaval {
				navalUnits += len(a.Units)
			} else {
				landUnits += len(a.Units)
			}
		}
		snapshots[fid] = headlessFactionSnapshot{
			Regions: regions, Armies: armies, LandUnits: landUnits, NavalUnits: navalUnits,
			Buildings: buildings, Techs: len(f.Research.Completed), Gold: f.Gold,
			Grain: f.Grain, Eliminated: f.IsEliminated,
		}
	}
	return snapshots
}

func captureRegionOwners(gs *state.GameState) map[world.RegionID]string {
	owners := make(map[world.RegionID]string)
	for id, r := range gs.Regions {
		if r != nil && !r.IsSea && !r.IsTerrainArea {
			owners[id] = r.OwnerID
		}
	}
	return owners
}

func captureWarSnapshots(gs *state.GameState) map[string]headlessWarSnapshot {
	result := make(map[string]headlessWarSnapshot)
	ids := sortedFactionIDs(gs)
	for i, a := range ids {
		for _, b := range ids[i+1:] {
			rel := diplomacy.Relation(gs, a, b)
			if rel == nil || rel.Stance != faction.StanceWar {
				continue
			}
			key := faction.RelationKey(a, b)
			started := gs.Turn
			if ledger := gs.WarLedgers[key]; ledger != nil && ledger.StartedTurn > 0 {
				started = ledger.StartedTurn
			}
			result[key] = headlessWarSnapshot{Active: true, StartedTurn: started}
		}
	}
	return result
}

func captureTradeSnapshots(gs *state.GameState) map[string]headlessTradeSnapshot {
	result := make(map[string]headlessTradeSnapshot)
	for _, route := range gs.TradeRoutes {
		if route == nil {
			continue
		}
		key := fmt.Sprintf("%s|%s|%s", route.FromFactionID, route.ToFactionID, route.Good)
		result[key] = headlessTradeSnapshot{From: faction.FactionID(route.FromFactionID), To: faction.FactionID(route.ToFactionID), Good: string(route.Good)}
	}
	return result
}

func captureAllianceSnapshots(gs *state.GameState) map[string]bool {
	result := make(map[string]bool)
	ids := sortedFactionIDs(gs)
	for i, a := range ids {
		for _, b := range ids[i+1:] {
			rel := diplomacy.Relation(gs, a, b)
			if rel != nil && rel.Stance == faction.StanceAllied {
				result[faction.RelationKey(a, b)] = true
			}
		}
	}
	return result
}

func sortedFactionIDs(gs *state.GameState) []faction.FactionID {
	ids := make([]faction.FactionID, 0, len(gs.Factions))
	for fid, f := range gs.Factions {
		if f != nil && !f.IsVirtual {
			ids = append(ids, fid)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func observeHeadlessWarChanges(report *HeadlessSimulationReport, gs *state.GameState, previous map[string]headlessWarSnapshot, turn int) {
	current := captureWarSnapshots(gs)
	for key, now := range current {
		if previous[key].Active {
			continue
		}
		alreadyOpen := false
		for _, war := range report.Wars {
			if faction.RelationKey(war.FactionA, war.FactionB) == key && war.EndedTurn == 0 {
				alreadyOpen = true
				break
			}
		}
		if alreadyOpen {
			continue
		}
		a, b := splitRelationKey(key)
		report.Wars = append(report.Wars, HeadlessWarSummary{FactionA: a, FactionB: b, StartedTurn: now.StartedTurn})
	}
	for key, before := range previous {
		if !before.Active || current[key].Active {
			continue
		}
		for i := range report.Wars {
			if faction.RelationKey(report.Wars[i].FactionA, report.Wars[i].FactionB) == key && report.Wars[i].EndedTurn == 0 {
				report.Wars[i].EndedTurn = turn
				break
			}
		}
	}
}

func closeHeadlessWarChanges(report *HeadlessSimulationReport, gs *state.GameState, turn int) {
	current := captureWarSnapshots(gs)
	for i := range report.Wars {
		war := &report.Wars[i]
		key := faction.RelationKey(war.FactionA, war.FactionB)
		if war.EndedTurn == 0 && !current[key].Active {
			war.EndedTurn = turn
		}
	}
}

func observeHeadlessDiplomacyChanges(report *HeadlessSimulationReport, gs *state.GameState, previousTrades map[string]headlessTradeSnapshot, previousAlliances map[string]bool, turn int) {
	currentTrades := captureTradeSnapshots(gs)
	for key, route := range currentTrades {
		if _, ok := previousTrades[key]; !ok {
			report.TradeChanges = append(report.TradeChanges, HeadlessDiplomacyEvent{Turn: turn, Action: "başladı", FactionA: route.From, FactionB: route.To, Good: route.Good})
		}
	}
	for key, route := range previousTrades {
		if _, ok := currentTrades[key]; !ok {
			report.TradeChanges = append(report.TradeChanges, HeadlessDiplomacyEvent{Turn: turn, Action: "bitti", FactionA: route.From, FactionB: route.To, Good: route.Good})
		}
	}
	currentAlliances := captureAllianceSnapshots(gs)
	for key := range currentAlliances {
		if !previousAlliances[key] {
			a, b := splitRelationKey(key)
			report.AllianceChanges = append(report.AllianceChanges, HeadlessDiplomacyEvent{Turn: turn, Action: "ittifak kuruldu", FactionA: a, FactionB: b})
		}
	}
	for key := range previousAlliances {
		if !currentAlliances[key] {
			a, b := splitRelationKey(key)
			report.AllianceChanges = append(report.AllianceChanges, HeadlessDiplomacyEvent{Turn: turn, Action: "ittifak bitti", FactionA: a, FactionB: b})
		}
	}
}

func splitRelationKey(key string) (faction.FactionID, faction.FactionID) {
	parts := strings.SplitN(key, "|", 2)
	if len(parts) != 2 {
		return faction.FactionID(key), ""
	}
	return faction.FactionID(parts[0]), faction.FactionID(parts[1])
}

func recordHeadlessRegionChanges(report *HeadlessSimulationReport, gs *state.GameState, previous map[world.RegionID]string, turn int) {
	current := captureRegionOwners(gs)
	for regionID, before := range previous {
		after := current[regionID]
		if before == after || after == "" {
			continue
		}
		for i := range report.Factions {
			if report.Factions[i].ID == faction.FactionID(after) {
				report.Factions[i].RegionsGained = appendUniqueString(report.Factions[i].RegionsGained, regionLabel(gs, regionID))
			}
			if report.Factions[i].ID == faction.FactionID(before) {
				report.Factions[i].RegionsLost = appendUniqueString(report.Factions[i].RegionsLost, regionLabel(gs, regionID))
			}
		}
		_ = turn
	}
}

func recordHeadlessEliminations(report *HeadlessSimulationReport, gs *state.GameState, initial map[faction.FactionID]headlessFactionSnapshot, turn int) {
	seen := make(map[faction.FactionID]bool, len(report.Eliminations))
	for _, elimination := range report.Eliminations {
		seen[elimination.Faction] = true
	}
	for fid, f := range gs.Factions {
		if f == nil || !f.IsEliminated || initial[fid].Eliminated || seen[fid] {
			continue
		}
		report.Eliminations = append(report.Eliminations, HeadlessElimination{Turn: turn, Faction: fid, NameTR: f.NameTR})
	}
}

func regionLabel(gs *state.GameState, id world.RegionID) string {
	if r := gs.Regions[id]; r != nil && r.NameTR != "" {
		return r.NameTR
	}
	return string(id)
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func countHeadlessProductionOrders(report *HeadlessSimulationReport, before, after []state.ProductionOrder, fid faction.FactionID) {
	beforeOrders := make(map[string]int, len(before))
	for _, order := range before {
		beforeOrders[headlessProductionOrderKey(order)]++
	}
	for _, order := range after {
		key := headlessProductionOrderKey(order)
		if beforeOrders[key] > 0 {
			beforeOrders[key]--
			continue
		}
		if faction.FactionID(order.FactionID) != fid {
			continue
		}
		switch order.Kind {
		case "building":
			for i := range report.Factions {
				if report.Factions[i].ID == fid {
					report.Factions[i].BuildingOrders++
				}
			}
		case "unit":
			for i := range report.Factions {
				if report.Factions[i].ID == fid {
					report.Factions[i].ProductionOrders++
				}
			}
		}
	}
}

func headlessProductionOrderKey(order state.ProductionOrder) string {
	return fmt.Sprintf("%s|%s|%s|%s|%d|%d", order.ID, order.Kind, order.FactionID, order.RegionID, order.TurnsLeft, order.BuildingLevel)
}

func finalizeHeadlessFactionSummaries(gs *state.GameState, initial map[faction.FactionID]headlessFactionSnapshot, report *HeadlessSimulationReport) []HeadlessFactionSummary {
	final := captureHeadlessFactionSnapshots(gs)
	ids := sortedFactionIDs(gs)
	result := make([]HeadlessFactionSummary, 0, len(ids))
	for _, fid := range ids {
		before := initial[fid]
		after := final[fid]
		f := gs.Factions[fid]
		queuedProduction, queuedBuildings := 0, 0
		for _, tracked := range report.Factions {
			if tracked.ID == fid {
				queuedProduction = tracked.ProductionOrders
				queuedBuildings = tracked.BuildingOrders
				break
			}
		}
		gained := append([]string(nil), reportFactionRegionChanges(report, fid, true)...)
		lost := append([]string(nil), reportFactionRegionChanges(report, fid, false)...)
		result = append(result, HeadlessFactionSummary{
			ID: fid, NameTR: f.NameTR, InitialRegions: len(before.Regions), FinalRegions: len(after.Regions),
			RegionsGained: gained, RegionsLost: lost, InitialArmies: before.Armies, FinalArmies: after.Armies,
			InitialLandUnits: before.LandUnits, FinalLandUnits: after.LandUnits, InitialNavalUnits: before.NavalUnits,
			FinalNavalUnits: after.NavalUnits, ProductionOrders: queuedProduction, BuildingOrders: queuedBuildings,
			InitialBuildings: before.Buildings, FinalBuildings: after.Buildings,
			TechnologiesCompleted: maxHeadlessInt(0, after.Techs-before.Techs), InitialGold: before.Gold, FinalGold: after.Gold,
			InitialGrain: before.Grain, FinalGrain: after.Grain, Eliminated: after.Eliminated,
		})
	}
	return result
}

func maxHeadlessInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func reportFactionRegionChanges(report *HeadlessSimulationReport, fid faction.FactionID, gained bool) []string {
	values := make([]string, 0)
	for _, summary := range report.Factions {
		if summary.ID != fid {
			continue
		}
		if gained {
			return summary.RegionsGained
		}
		return summary.RegionsLost
	}
	return values
}

func finalizeHeadlessWars(gs *state.GameState, report *HeadlessSimulationReport) []HeadlessWarSummary {
	result := append([]HeadlessWarSummary(nil), report.Wars...)
	for key, ledger := range gs.WarLedgers {
		if ledger == nil {
			continue
		}
		found := -1
		for i := range result {
			if faction.RelationKey(result[i].FactionA, result[i].FactionB) == key {
				found = i
				break
			}
		}
		if found < 0 {
			result = append(result, HeadlessWarSummary{FactionA: ledger.FactionA, FactionB: ledger.FactionB, StartedTurn: ledger.StartedTurn})
			found = len(result) - 1
		}
		result[found].RegionsCapturedA = ledger.RegionsCapturedA
		result[found].RegionsCapturedB = ledger.RegionsCapturedB
		result[found].CasualtiesA = ledger.CasualtiesA
		result[found].CasualtiesB = ledger.CasualtiesB
		result[found].Active = diplomacy.IsWar(gs, ledger.FactionA, ledger.FactionB)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].StartedTurn != result[j].StartedTurn {
			return result[i].StartedTurn < result[j].StartedTurn
		}
		return result[i].FactionA+result[i].FactionB < result[j].FactionA+result[j].FactionB
	})
	return result
}

func captureHeadlessCheckpoint(gs *state.GameState, elapsedTurn int) HeadlessCheckpoint {
	ids := sortedFactionIDs(gs)
	rankings := make([]HeadlessFactionRanking, 0, len(ids))
	for _, fid := range ids {
		f := gs.Factions[fid]
		if f == nil || f.IsEliminated {
			continue
		}
		armies, land, naval, power := 0, 0, 0, 0
		for _, a := range gs.Armies {
			if a == nil || faction.FactionID(a.OwnerID) != fid {
				continue
			}
			armies++
			power += gs.EffectiveArmyStrength(a)
			if a.IsNaval {
				naval += len(a.Units)
			} else {
				land += len(a.Units)
			}
		}
		rankings = append(rankings, HeadlessFactionRanking{FactionID: fid, NameTR: f.NameTR, Regions: len(gs.LandRegionsOwnedBy(fid)), Armies: armies, LandUnits: land, NavalUnits: naval, MilitaryPower: power})
	}
	sort.Slice(rankings, func(i, j int) bool {
		if rankings[i].Regions != rankings[j].Regions {
			return rankings[i].Regions > rankings[j].Regions
		}
		if rankings[i].MilitaryPower != rankings[j].MilitaryPower {
			return rankings[i].MilitaryPower > rankings[j].MilitaryPower
		}
		return rankings[i].FactionID < rankings[j].FactionID
	})
	return HeadlessCheckpoint{Turn: elapsedTurn, Year: gs.Year, Month: gs.Month, Rankings: rankings}
}

func eventDefinitionsByID(eventsList []*events.Event) map[string]*events.Event {
	result := make(map[string]*events.Event, len(eventsList))
	for _, evt := range eventsList {
		if evt != nil && evt.ID != "" {
			result[evt.ID] = evt
		}
	}
	return result
}

// HeadlessReportFactionNames, rapor formatlayıcıların ortak isim çözümlemesi
// için küçük bir yardımcıdır.
func HeadlessReportFactionNames(report *HeadlessSimulationReport) map[faction.FactionID]string {
	result := make(map[faction.FactionID]string, len(report.Factions))
	for _, f := range report.Factions {
		result[f.ID] = f.NameTR
	}
	return result
}
