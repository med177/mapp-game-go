package events

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

type RelationEffect struct {
	FactionID  string `json:"faction_id"`
	Stance     string `json:"stance,omitempty"`
	ScoreDelta int    `json:"score_delta,omitempty"`
}

// CoalitionEffect bir grup faction arasında müttefiklik kurar ve grubun
// belirlenen rakiplere karşı ortak savaş ilişkisi oluşturmasını sağlar.
type CoalitionEffect struct {
	Members            []string `json:"members"`
	Opponents          []string `json:"opponents"`
	MemberStance       string   `json:"member_stance,omitempty"`
	OpponentStance     string   `json:"opponent_stance,omitempty"`
	MemberScoreDelta   int      `json:"member_score_delta,omitempty"`
	OpponentScoreDelta int      `json:"opponent_score_delta,omitempty"`
}

type RelationRequirement struct {
	FactionID     string   `json:"faction_id"`
	Stance        string   `json:"stance,omitempty"`
	AnyOfStances  []string `json:"any_of_stances,omitempty"`
	BlocksStances []string `json:"blocks_stances,omitempty"`
	MinScore      int      `json:"min_score,omitempty"`
	MaxScore      int      `json:"max_score,omitempty"`
}

// DiplomaticOfferEffect event seçiminin doğrudan ilişki değiştirmek yerine
// mevcut diplomasi teklif kuyruğuna bırakacağı teklifi tanımlar.
type DiplomaticOfferEffect struct {
	FactionID string `json:"faction_id"`
	Action    string `json:"action"`
	Priority  int    `json:"priority,omitempty"`
	ReasonTR  string `json:"reason_tr,omitempty"`
}

// SuccessorRevivalEffect tarihsel bir event'in elenmiş ardıl faction'ı
// belirli bir bölgede yeniden kurmasını tanımlar.
type SuccessorRevivalEffect struct {
	FactionID    string `json:"faction_id"`
	RegionID     string `json:"region_id"`
	Mode         string `json:"mode,omitempty"` // independent | vassal
	OverlordID   string `json:"overlord_id,omitempty"`
	MilitiaCount int    `json:"militia_count,omitempty"`
}

// TradeNetworkModifierEffect, bir event'in ticaret ağı üzerindeki kalıcı
// gelir ve kaynak etkisini veri üzerinden tanımlar.
type TradeNetworkModifierEffect struct {
	ID                 string   `json:"id"`
	CenterIDs          []string `json:"center_ids,omitempty"`
	RegionIDs          []string `json:"region_ids,omitempty"`
	TradeIncomePercent int      `json:"trade_income_percent,omitempty"`
	SpicePercent       int      `json:"spice_percent,omitempty"`
}

// FactionSubjugationTrigger, bir event'in oyuncunun seçtiği faction başka
// bir üyeyi elediğinde veya vassal yaptığında açılmasını sağlar.
type FactionSubjugationTrigger struct {
	FactionIDs         []string `json:"faction_ids"`
	RequirePlayerActor bool     `json:"require_player_actor,omitempty"`
}

type Effect struct {
	Target                    string                       `json:"target,omitempty"` // boşsa event target'ı kullanılır
	SatDelta                  int                          `json:"sat_delta,omitempty"`
	GoldDelta                 int                          `json:"gold_delta,omitempty"`
	GrainDelta                int                          `json:"grain_delta,omitempty"`
	ArmyHPMod                 float64                      `json:"army_hp_mod,omitempty"`                // 1.0 = değişmez
	GrainProductionPercent    int                          `json:"grain_production_percent,omitempty"`   // aktif olay süresince üretim etkisi
	GrainDemandPercent        int                          `json:"grain_demand_percent,omitempty"`       // aktif olay süresince sivil tüketim etkisi
	TradeIncomePercent        int                          `json:"trade_income_percent,omitempty"`       // aktif olay süresince ticaret geliri etkisi
	RegionGoldIncomePercent   int                          `json:"region_gold_income_percent,omitempty"` // aktif olay süresince vergi geliri etkisi
	ArmyUpkeepPercent         int                          `json:"army_upkeep_percent,omitempty"`        // aktif bölgede ordu ikmal etkisi
	PopulationDelta           int                          `json:"population_delta,omitempty"`           // anlık bölge nüfusu etkisi
	BuildingEfficiencyPercent int                          `json:"building_efficiency_percent,omitempty"`
	CombatAttackPercent       int                          `json:"combat_attack_percent,omitempty"`
	CombatDefensePercent      int                          `json:"combat_defense_percent,omitempty"`
	AffectedFaction           string                       `json:"affected_faction,omitempty"`   // specific_faction için
	RelationDeltaAll          int                          `json:"relation_delta_all,omitempty"` // etkilenen fraksiyonun tüm ilişkilerine uygulanır
	CompleteTechs             []string                     `json:"complete_techs,omitempty"`
	StartResearchTech         string                       `json:"start_research_tech,omitempty"`
	Relations                 []RelationEffect             `json:"relations,omitempty"`
	Coalition                 *CoalitionEffect             `json:"coalition,omitempty"`
	DiplomaticOffers          []DiplomaticOfferEffect      `json:"diplomatic_offers,omitempty"`
	PoliticalTransformationID string                       `json:"political_transformation_id,omitempty"`
	PoliticalWinnerFactionID  string                       `json:"political_winner_faction_id,omitempty"`
	PoliticalWinnerFromPlayer bool                         `json:"political_winner_from_player,omitempty"`
	PlayerFactionID           string                       `json:"player_faction_id,omitempty"`
	SuccessorRevival          *SuccessorRevivalEffect      `json:"successor_revival,omitempty"`
	SuccessorRevivals         []SuccessorRevivalEffect     `json:"successor_revivals,omitempty"`
	TradeNetworkModifiers     []TradeNetworkModifierEffect `json:"trade_network_modifiers,omitempty"`
	SetFlags                  []string                     `json:"set_flags,omitempty"`
	ClearFlags                []string                     `json:"clear_flags,omitempty"`
	CapitalSettlementID       string                       `json:"capital_settlement_id,omitempty"`
	CapitalMoveTurns          int                          `json:"capital_move_turns,omitempty"`
}

type Choice struct {
	ID       string `json:"id,omitempty"`
	LabelTR  string `json:"label_tr"`
	DescTR   string `json:"desc_tr"`
	AIWeight int    `json:"ai_weight,omitempty"`
	Effect   Effect `json:"effect"`
}

// Event bir tarihsel olayı tanımlar.
type Event struct {
	ID                        string                       `json:"id"`
	NameTR                    string                       `json:"name_tr"`
	DescTR                    string                       `json:"desc_tr"`
	Probability               float64                      `json:"probability"` // 0 = sadece tarihsel tetiklenme
	MinTurn                   int                          `json:"min_turn"`    // en erken tur (rastgele olaylar için)
	Target                    string                       `json:"target"`      // "player_faction"|"random_region"|"all_armies"|"all_factions"
	SatDelta                  int                          `json:"sat_delta"`
	GoldDelta                 int                          `json:"gold_delta"`
	GrainDelta                int                          `json:"grain_delta"`
	ArmyHPMod                 float64                      `json:"army_hp_mod"` // 1.0 = değişmez
	GrainProductionPercent    int                          `json:"grain_production_percent,omitempty"`
	GrainDemandPercent        int                          `json:"grain_demand_percent,omitempty"`
	TradeIncomePercent        int                          `json:"trade_income_percent,omitempty"`
	RegionGoldIncomePercent   int                          `json:"region_gold_income_percent,omitempty"`
	ArmyUpkeepPercent         int                          `json:"army_upkeep_percent,omitempty"`
	PopulationDelta           int                          `json:"population_delta,omitempty"`
	BuildingEfficiencyPercent int                          `json:"building_efficiency_percent,omitempty"`
	CombatAttackPercent       int                          `json:"combat_attack_percent,omitempty"`
	CombatDefensePercent      int                          `json:"combat_defense_percent,omitempty"`
	RelationDeltaAll          int                          `json:"relation_delta_all,omitempty"`
	CompleteTechs             []string                     `json:"complete_techs,omitempty"`
	StartResearchTech         string                       `json:"start_research_tech,omitempty"`
	Relations                 []RelationEffect             `json:"relations,omitempty"`
	Coalition                 *CoalitionEffect             `json:"coalition,omitempty"`
	SetFlags                  []string                     `json:"set_flags,omitempty"`
	SuccessorRevival          *SuccessorRevivalEffect      `json:"successor_revival,omitempty"`
	SuccessorRevivals         []SuccessorRevivalEffect     `json:"successor_revivals,omitempty"`
	ClearFlags                []string                     `json:"clear_flags,omitempty"`
	TradeNetworkModifiers     []TradeNetworkModifierEffect `json:"trade_network_modifiers,omitempty"`

	// Tarihsel tetiklenme alanları
	HistoricalYear  int    `json:"historical_year,omitempty"`  // 0 = tarihsel değil
	HistoricalMonth int    `json:"historical_month,omitempty"` // 0 = yılın herhangi bir ayı
	OneShot         bool   `json:"one_shot,omitempty"`         // true = yalnızca bir kez tetiklenir
	AffectedFaction string `json:"affected_faction,omitempty"` // belirli fraksiyonu hedefle

	ChoicePromptTR            string                     `json:"choice_prompt_tr,omitempty"`
	Choices                   []Choice                   `json:"choices,omitempty"`
	PlayerChoiceFactions      []string                   `json:"player_choice_factions,omitempty"`
	RequiresFlags             []string                   `json:"requires_flags,omitempty"`
	BlocksFlags               []string                   `json:"blocks_flags,omitempty"`
	BlocksCapitalSettlementID string                     `json:"blocks_capital_settlement_id,omitempty"`
	RequiresTechs             []string                   `json:"requires_techs,omitempty"`
	BlocksTechs               []string                   `json:"blocks_techs,omitempty"`
	RequiresOwnedRegions      []world.RegionID           `json:"requires_owned_regions,omitempty"`
	RequiresActiveFactions    []string                   `json:"requires_active_factions,omitempty"`
	RelationRequirements      []RelationRequirement      `json:"relation_requirements,omitempty"`
	FactionSubjugationTrigger *FactionSubjugationTrigger `json:"faction_subjugation_trigger,omitempty"`
}

// LoadEvents olayları JSON'dan yükler.
func LoadEvents(path string) ([]*Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("olaylar okunamadı: %w", err)
	}
	var list []*Event
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("olaylar parse edilemedi: %w", err)
	}
	return list, nil
}

// Tick her tur sonunda olayları kontrol eder ve tetiklenen ilk olayı döner.
func Tick(gs *state.GameState, evts []*Event) *Event {
	if gs.FiredEventIDs == nil {
		gs.FiredEventIDs = make(map[string]bool)
	}

	// Önce aynı turda oluşan siyasi üstünlüğe bağlı event'leri kontrol et.
	// Böylece oyuncu rakip hanedanı elediğinde veya vassal yaptığında 1485'i
	// beklemek zorunda kalmaz.
	actor, target := gs.LastSubjugationActorID, gs.LastSubjugatedFactionID
	if actor != "" || target != "" {
		for _, e := range evts {
			if e == nil || e.FactionSubjugationTrigger == nil || gs.FiredEventIDs[e.ID] ||
				!eventConditionsSatisfied(gs, e) || !factionSubjugationTriggerSatisfied(gs, e, actor, target) {
				continue
			}
			if e.OneShot {
				gs.FiredEventIDs[e.ID] = true
			}
			gs.ConsumeFactionSubjugation()
			return e
		}
		gs.ConsumeFactionSubjugation()
	}

	// Aynı takvim penceresinde birden fazla tarihsel olay varsa, önceki turda
	// ertelenen olayları normal tarih taramasından önce sıraya al.
	for _, e := range evts {
		if e == nil || e.HistoricalYear == 0 || gs.FiredEventIDs[e.ID] ||
			!gs.FiredEventIDs[pendingHistoricalEventKey(e.ID)] {
			continue
		}
		if !eventConditionsSatisfied(gs, e) {
			continue
		}
		if e.OneShot {
			gs.FiredEventIDs[e.ID] = true
		}
		delete(gs.FiredEventIDs, pendingHistoricalEventKey(e.ID))
		return e
	}

	// Önce tarihsel olayları kontrol et (kesinlikle tetiklenir)
	for _, e := range evts {
		if e.HistoricalYear == 0 {
			continue
		}
		if gs.FiredEventIDs[e.ID] {
			continue
		}
		if !eventConditionsSatisfied(gs, e) {
			continue
		}
		if !historicalEventDueThisTurn(gs, e) {
			continue
		}
		queueHistoricalEventsSharingDate(gs, evts, e)
		if e.OneShot {
			gs.FiredEventIDs[e.ID] = true
		}
		return e
	}

	// Rastgele olaylar
	for _, e := range evts {
		if e.Probability <= 0 {
			continue
		}
		if e.OneShot && gs.FiredEventIDs[e.ID] {
			continue
		}
		if !eventConditionsSatisfied(gs, e) {
			continue
		}
		if gs.Turn < e.MinTurn {
			continue
		}
		if rand.Float64() > e.Probability {
			continue
		}
		if e.OneShot {
			gs.FiredEventIDs[e.ID] = true
		}
		return e
	}
	return nil
}

// historicalEventDueThisTurn tarihsel olayın aktif turun takvim aralığına
// denk gelip gelmediğini bildirir.
func historicalEventDueThisTurn(gs *state.GameState, e *Event) bool {
	if gs == nil || e == nil {
		return false
	}
	if gs.HistoricalDateOccursThisTurn(e.HistoricalYear, e.HistoricalMonth) {
		return true
	}
	if e.HistoricalMonth <= 0 || e.HistoricalMonth > 12 || gs.Year <= 0 {
		return false
	}
	startMonth := gs.Month
	if startMonth < 1 || startMonth > 12 {
		startMonth = 1
	}
	startAbs := gs.Year*12 + startMonth - 1
	targetAbs := e.HistoricalYear*12 + e.HistoricalMonth - 1
	graceMonths := gs.CalendarMonthsPerTurn() - 1
	return targetAbs < startAbs && targetAbs >= startAbs-graceMonths
}

func pendingHistoricalEventKey(id string) string {
	return "pending:event:" + id
}

func queueHistoricalEventsSharingDate(gs *state.GameState, evts []*Event, selected *Event) {
	if gs == nil || selected == nil || gs.FiredEventIDs == nil {
		return
	}
	for _, candidate := range evts {
		if candidate == nil || candidate == selected || candidate.HistoricalYear == 0 ||
			candidate.HistoricalYear != selected.HistoricalYear ||
			candidate.HistoricalMonth != selected.HistoricalMonth ||
			gs.FiredEventIDs[candidate.ID] ||
			!eventConditionsSatisfied(gs, candidate) ||
			!historicalEventDueThisTurn(gs, candidate) {
			continue
		}
		gs.FiredEventIDs[pendingHistoricalEventKey(candidate.ID)] = true
	}
}

func (e *Event) BaseEffect() Effect {
	return Effect{
		Target:                    e.Target,
		SatDelta:                  e.SatDelta,
		GoldDelta:                 e.GoldDelta,
		GrainDelta:                e.GrainDelta,
		ArmyHPMod:                 e.ArmyHPMod,
		GrainProductionPercent:    e.GrainProductionPercent,
		GrainDemandPercent:        e.GrainDemandPercent,
		TradeIncomePercent:        e.TradeIncomePercent,
		RegionGoldIncomePercent:   e.RegionGoldIncomePercent,
		ArmyUpkeepPercent:         e.ArmyUpkeepPercent,
		PopulationDelta:           e.PopulationDelta,
		BuildingEfficiencyPercent: e.BuildingEfficiencyPercent,
		CombatAttackPercent:       e.CombatAttackPercent,
		CombatDefensePercent:      e.CombatDefensePercent,
		AffectedFaction:           e.AffectedFaction,
		RelationDeltaAll:          e.RelationDeltaAll,
		CompleteTechs:             e.CompleteTechs,
		StartResearchTech:         e.StartResearchTech,
		Relations:                 e.Relations,
		Coalition:                 e.Coalition,
		SetFlags:                  e.SetFlags,
		ClearFlags:                e.ClearFlags,
		SuccessorRevival:          e.SuccessorRevival,
		SuccessorRevivals:         e.SuccessorRevivals,
		TradeNetworkModifiers:     e.TradeNetworkModifiers,
		CapitalSettlementID:       "",
		CapitalMoveTurns:          0,
	}
}

func RequiresPlayerChoice(gs *state.GameState, e *Event) bool {
	if gs == nil || e == nil || len(e.Choices) == 0 {
		return false
	}
	switch e.Target {
	case "player_faction", "all_factions", "all_armies":
		return true
	case "specific_faction":
		if e.AffectedFaction == string(gs.PlayerFactionID) {
			return true
		}
		for _, playerFactionID := range e.PlayerChoiceFactions {
			if playerFactionID == string(gs.PlayerFactionID) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// IsPlayerRelevant, bir event'in oyuncuya görünür bir bildirim olarak
// sunulması gerekip gerekmediğini bildirir. Oyuncuya ait olmayan specific
// faction event'leri oyun durumuna uygulanmaya devam eder; yalnızca oyuncunun
// ekranını gereksiz yere meşgul etmez.
func IsPlayerRelevant(gs *state.GameState, e *Event) bool {
	if gs == nil || e == nil {
		return false
	}

	switch e.Target {
	case "specific_faction":
		if e.AffectedFaction == string(gs.PlayerFactionID) {
			return true
		}
		for _, playerFactionID := range e.PlayerChoiceFactions {
			if playerFactionID == string(gs.PlayerFactionID) {
				return true
			}
		}
		return false
	case "player_faction", "all_factions", "all_armies", "random_region":
		return true
	default:
		// Tanımsız hedeflerde mevcut görünür davranışı koru; yeni event
		// hedefleri eklenirken sessizce kaybolmasın.
		return true
	}
}

func Apply(gs *state.GameState, e *Event) {
	if gs == nil || e == nil {
		return
	}
	targetRegionID := applyEffect(gs, e.BaseEffect())
	applySuccessorRevival(gs, e.BaseEffect())
	addRegionEventStatus(gs, e, nil, targetRegionID)
}

func ApplyChoice(gs *state.GameState, e *Event, idx int) (Choice, bool) {
	if e == nil || idx < 0 || idx >= len(e.Choices) {
		return Choice{}, false
	}
	choice := e.Choices[idx]
	eff := choiceEffect(e, choice)
	targetRegionID := applyEffect(gs, eff)
	applySuccessorRevival(gs, eff)
	addRegionEventStatus(gs, e, &choice, targetRegionID)
	return choice, true
}

func applySuccessorRevival(gs *state.GameState, eff Effect) {
	if gs == nil {
		return
	}
	if eff.SuccessorRevival != nil {
		applyOneSuccessorRevival(gs, eff, *eff.SuccessorRevival)
	}
	for _, revival := range eff.SuccessorRevivals {
		applyOneSuccessorRevival(gs, eff, revival)
	}
}

func applyOneSuccessorRevival(gs *state.GameState, eff Effect, revival SuccessorRevivalEffect) {
	successorID := faction.FactionID(revival.FactionID)
	regionID := world.RegionID(revival.RegionID)
	if successorID == "" || regionID == "" {
		return
	}
	if _, ok := gs.ReviveSuccessorAtRegion(regionID, successorID, revival.MilitiaCount); !ok {
		return
	}
	overlordID := faction.FactionID(revival.OverlordID)
	if overlordID == "" && revival.Mode == "vassal" {
		overlordID = faction.FactionID(eff.AffectedFaction)
	}
	if revival.Mode == "vassal" && overlordID != "" && gs.Factions[overlordID] != nil {
		successor := gs.Factions[successorID]
		successor.OverlordID = overlordID
		successor.TributeRate = 20
		successor.TributeRateConfigured = true
		diplomacy.ForceRelation(gs, overlordID, successorID, faction.StanceAllied, 50)
	} else if eff.AffectedFaction != "" {
		diplomacy.ForceRelation(gs, faction.FactionID(eff.AffectedFaction), successorID, faction.StanceAllied, 50)
	}
}

func ConditionsMet(gs *state.GameState, e *Event) bool {
	return eventConditionsSatisfied(gs, e)
}

func ConditionFailureReasons(gs *state.GameState, e *Event) []string {
	if gs == nil || e == nil {
		return []string{"gecersiz event"}
	}
	if !eventTargetFactionActive(gs, e) {
		reasons := []string{"hedef faction aktif değil"}
		return reasons
	}
	reasons := make([]string, 0, 6)
	for _, flag := range e.RequiresFlags {
		if flag == "" || gs.FiredEventIDs[eventFlagKey(flag)] {
			continue
		}
		reasons = append(reasons, "flag bekleniyor: "+flag)
	}
	for _, flag := range e.BlocksFlags {
		if flag != "" && gs.FiredEventIDs[eventFlagKey(flag)] {
			reasons = append(reasons, "bloklayan flag: "+flag)
		}
	}
	f := eventConditionFaction(gs, e)
	if len(e.RequiresTechs) > 0 {
		if f == nil {
			reasons = append(reasons, "faction tech durumu okunamadi")
		} else {
			for _, techID := range e.RequiresTechs {
				if techID == "" || f.Research.Completed[techID] {
					continue
				}
				reasons = append(reasons, "gerekli tech: "+techID)
			}
		}
	}
	if len(e.BlocksTechs) > 0 && f != nil {
		for _, techID := range e.BlocksTechs {
			if techID != "" && f.Research.Completed[techID] {
				reasons = append(reasons, "zaten acik tech: "+techID)
			}
		}
	}
	if len(e.RequiresOwnedRegions) > 0 {
		fid := eventConditionFactionID(gs, e)
		if fid == "" {
			reasons = append(reasons, "kosul fraksiyonu yok")
		} else {
			for _, rid := range e.RequiresOwnedRegions {
				r := gs.Regions[rid]
				if r == nil || r.OwnerID != string(fid) {
					reasons = append(reasons, "bolge gerekli: "+string(rid))
				}
			}
		}
	}
	if failedFaction := firstInactiveRequiredFaction(gs, e.RequiresActiveFactions); failedFaction != "" {
		reasons = append(reasons, "aktif faction gerekli: "+failedFaction)
	}
	if blockedSettlement := blockedCapitalSettlementID(gs, e); blockedSettlement != "" {
		reasons = append(reasons, "zaten hedef başkent: "+blockedSettlement)
	}
	if len(e.RelationRequirements) > 0 {
		fid := eventConditionFactionID(gs, e)
		if fid == "" {
			reasons = append(reasons, "diplomasi kosulu icin fraksiyon yok")
		} else {
			for _, req := range e.RelationRequirements {
				if relationRequirementSatisfied(gs, fid, req) {
					continue
				}
				reasons = append(reasons, relationRequirementReason(gs, fid, req))
			}
		}
	}
	if !factionSubjugationTriggerSatisfied(gs, e, gs.LastSubjugationActorID, gs.LastSubjugatedFactionID) {
		reasons = append(reasons, "siyasi üstünlük koşulu bekleniyor")
	}
	return reasons
}

func eventConditionsSatisfied(gs *state.GameState, e *Event) bool {
	if gs == nil || e == nil {
		return false
	}
	if !eventTargetFactionActive(gs, e) {
		return false
	}
	for _, flag := range e.RequiresFlags {
		if flag == "" || !gs.FiredEventIDs[eventFlagKey(flag)] {
			return false
		}
	}
	for _, flag := range e.BlocksFlags {
		if flag != "" && gs.FiredEventIDs[eventFlagKey(flag)] {
			return false
		}
	}
	if !eventTechsSatisfied(gs, e) {
		return false
	}
	if len(e.RequiresOwnedRegions) > 0 {
		fid := eventConditionFactionID(gs, e)
		if fid == "" {
			return false
		}
		for _, rid := range e.RequiresOwnedRegions {
			r := gs.Regions[rid]
			if r == nil || r.OwnerID != string(fid) {
				return false
			}
		}
	}
	if firstInactiveRequiredFaction(gs, e.RequiresActiveFactions) != "" {
		return false
	}
	if blockedCapitalSettlementID(gs, e) != "" {
		return false
	}
	if !eventRelationsSatisfied(gs, e) {
		return false
	}
	if !factionSubjugationTriggerSatisfied(gs, e, gs.LastSubjugationActorID, gs.LastSubjugatedFactionID) {
		return false
	}
	return true
}

func firstInactiveRequiredFaction(gs *state.GameState, factionIDs []string) string {
	if len(factionIDs) == 0 {
		return ""
	}
	for _, id := range factionIDs {
		if id == "" {
			continue
		}
		f := gs.Factions[faction.FactionID(id)]
		if f == nil || f.IsEliminated {
			return id
		}
	}
	return ""
}

func factionSubjugationTriggerSatisfied(gs *state.GameState, e *Event, actor, target faction.FactionID) bool {
	if e == nil || e.FactionSubjugationTrigger == nil {
		return true
	}
	if gs == nil || actor == "" || target == "" || actor == target {
		return false
	}
	trigger := e.FactionSubjugationTrigger
	actorListed, targetListed := false, false
	for _, id := range trigger.FactionIDs {
		if id == string(actor) {
			actorListed = true
		}
		if id == string(target) {
			targetListed = true
		}
	}
	if !actorListed || !targetListed {
		return false
	}
	return !trigger.RequirePlayerActor || actor == gs.PlayerFactionID
}

// eventTargetFactionActive tarihsel veya rastgele olayın doğrudan hedeflediği
// faction'ın oyunda kalmasını zorunlu kılar. Böylece elenmiş bir İngiltere
// Güller Savaşı'nı, elenmiş bir Osmanlı da Mohaç/Çaldıran zincirini başlatamaz.
// all_factions ve all_armies gibi toplu hedeflerde özel faction aranmaz.
func eventTargetFactionActive(gs *state.GameState, e *Event) bool {
	if gs == nil || e == nil {
		return false
	}
	var fid faction.FactionID
	switch e.Target {
	case "specific_faction":
		fid = faction.FactionID(e.AffectedFaction)
	case "player_faction":
		fid = gs.PlayerFactionID
	default:
		return true
	}
	if fid == "" {
		return false
	}
	target := gs.Factions[fid]
	return target != nil && !target.IsEliminated
}

func blockedCapitalSettlementID(gs *state.GameState, e *Event) string {
	if gs == nil || e == nil || e.BlocksCapitalSettlementID == "" {
		return ""
	}
	f := eventConditionFaction(gs, e)
	if f == nil {
		return e.BlocksCapitalSettlementID
	}
	if f.CapitalSettlementID == e.BlocksCapitalSettlementID {
		return e.BlocksCapitalSettlementID
	}
	if f.PendingCapitalSettlementID == e.BlocksCapitalSettlementID && f.PendingCapitalTurns > 0 {
		return e.BlocksCapitalSettlementID
	}
	return ""
}

func eventTechsSatisfied(gs *state.GameState, e *Event) bool {
	f := eventConditionFaction(gs, e)
	if f == nil {
		return len(e.RequiresTechs) == 0 && len(e.BlocksTechs) == 0
	}
	for _, techID := range e.RequiresTechs {
		if techID == "" || !f.Research.Completed[techID] {
			return false
		}
	}
	for _, techID := range e.BlocksTechs {
		if techID != "" && f.Research.Completed[techID] {
			return false
		}
	}
	return true
}

func eventRelationsSatisfied(gs *state.GameState, e *Event) bool {
	if len(e.RelationRequirements) == 0 {
		return true
	}
	fid := eventConditionFactionID(gs, e)
	if fid == "" {
		return false
	}
	for _, req := range e.RelationRequirements {
		if !relationRequirementSatisfied(gs, fid, req) {
			return false
		}
	}
	return true
}

func relationRequirementSatisfied(gs *state.GameState, source faction.FactionID, req RelationRequirement) bool {
	if gs == nil || source == "" || req.FactionID == "" {
		return false
	}
	target := faction.FactionID(req.FactionID)
	rel := diplomacy.Relation(gs, source, target)
	score := 0
	stance := faction.StancePeace
	if rel != nil {
		score = rel.Score
		if rel.Stance != "" {
			stance = rel.Stance
		}
	}
	if req.MinScore != 0 && score < req.MinScore {
		return false
	}
	if req.MaxScore != 0 && score > req.MaxScore {
		return false
	}
	if req.Stance != "" && stance != faction.DiplomaticStance(req.Stance) {
		return false
	}
	if len(req.AnyOfStances) > 0 && !stanceInList(stance, req.AnyOfStances) {
		return false
	}
	if len(req.BlocksStances) > 0 && stanceInList(stance, req.BlocksStances) {
		return false
	}
	return true
}

func relationRequirementReason(gs *state.GameState, source faction.FactionID, req RelationRequirement) string {
	targetName := req.FactionID
	if gs != nil && gs.Factions != nil {
		if f := gs.Factions[faction.FactionID(req.FactionID)]; f != nil && f.NameTR != "" {
			targetName = f.NameTR
		}
	}
	target := faction.FactionID(req.FactionID)
	rel := diplomacy.Relation(gs, source, target)
	score := 0
	stance := faction.StancePeace
	if rel != nil {
		score = rel.Score
		if rel.Stance != "" {
			stance = rel.Stance
		}
	}
	parts := make([]string, 0, 4)
	if req.Stance != "" && stance != faction.DiplomaticStance(req.Stance) {
		parts = append(parts, "durus="+req.Stance)
	}
	if len(req.AnyOfStances) > 0 && !stanceInList(stance, req.AnyOfStances) {
		parts = append(parts, "durus="+strings.Join(req.AnyOfStances, "/"))
	}
	if len(req.BlocksStances) > 0 && stanceInList(stance, req.BlocksStances) {
		parts = append(parts, "yasak durus="+string(stance))
	}
	if req.MinScore != 0 && score < req.MinScore {
		parts = append(parts, fmt.Sprintf("skor>=%d", req.MinScore))
	}
	if req.MaxScore != 0 && score > req.MaxScore {
		parts = append(parts, fmt.Sprintf("skor<=%d", req.MaxScore))
	}
	if len(parts) == 0 {
		return "diplomasi kosulu: " + targetName
	}
	return "diplomasi kosulu (" + targetName + "): " + strings.Join(parts, ", ")
}

func stanceInList(stance faction.DiplomaticStance, list []string) bool {
	for _, candidate := range list {
		if candidate != "" && stance == faction.DiplomaticStance(candidate) {
			return true
		}
	}
	return false
}

func eventConditionFactionID(gs *state.GameState, e *Event) faction.FactionID {
	if gs == nil || e == nil {
		return ""
	}
	switch e.Target {
	case "player_faction":
		return gs.PlayerFactionID
	case "specific_faction":
		return faction.FactionID(e.AffectedFaction)
	default:
		return ""
	}
}

func eventConditionFaction(gs *state.GameState, e *Event) *faction.Faction {
	fid := eventConditionFactionID(gs, e)
	if fid == "" {
		return nil
	}
	return gs.Factions[fid]
}

func AutoChoose(e *Event) int {
	if e == nil || len(e.Choices) == 0 {
		return -1
	}
	bestIdx := 0
	bestWeight := e.Choices[0].AIWeight
	for i := 1; i < len(e.Choices); i++ {
		if e.Choices[i].AIWeight > bestWeight {
			bestIdx = i
			bestWeight = e.Choices[i].AIWeight
		}
	}
	return bestIdx
}

func choiceEffect(e *Event, c Choice) Effect {
	eff := c.Effect
	if eff.Target == "" {
		eff.Target = e.Target
	}
	if eff.AffectedFaction == "" {
		eff.AffectedFaction = e.AffectedFaction
	}
	return eff
}

func effectiveEventEffect(e *Event, choice *Choice) Effect {
	if e == nil {
		return Effect{}
	}
	if choice == nil {
		return e.BaseEffect()
	}
	return choiceEffect(e, *choice)
}

func eventFlagKey(flag string) string {
	return "flag:" + flag
}

func applyEffect(gs *state.GameState, eff Effect) world.RegionID {
	target := eff.Target
	var targetRegionID world.RegionID
	switch target {

	case "player_faction":
		applyToFaction(gs, string(gs.PlayerFactionID), eff)

	case "all_factions":
		for fid := range gs.Factions {
			applyToFaction(gs, string(fid), eff)
		}

	case "random_region":
		var candidates []world.RegionID
		for rid, r := range gs.Regions {
			if !r.IsSea && r.OwnerID != "" {
				candidates = append(candidates, rid)
			}
		}
		if len(candidates) == 0 {
			return ""
		}
		sort.Slice(candidates, func(i, j int) bool { return candidates[i] < candidates[j] })
		rid := candidates[rand.Intn(len(candidates))]
		r := gs.Regions[rid]
		r.Satisfaction = clamp(r.Satisfaction+eff.SatDelta, 0, 100)
		applyPopulationDelta(r, eff.PopulationDelta)
		if eff.GrainDelta != 0 {
			if f, ok := gs.Factions[faction.FactionID(r.OwnerID)]; ok {
				f.Grain = max0(f.Grain + eff.GrainDelta)
			}
		}
		targetRegionID = rid

	case "all_armies":
		if eff.ArmyHPMod > 0 && eff.ArmyHPMod < 1.0 {
			for _, a := range gs.Armies {
				for i := range a.Units {
					a.Units[i].CurrentHP = max0(int(float64(a.Units[i].CurrentHP) * eff.ArmyHPMod))
				}
			}
		}
		for _, f := range gs.Factions {
			f.Grain = max0(f.Grain + eff.GrainDelta)
		}

	case "specific_faction":
		if eff.AffectedFaction != "" {
			applyToFaction(gs, eff.AffectedFaction, eff)
		}
	}
	applyCoalition(gs, eff.Coalition)
	applyTradeNetworkModifiers(gs, eff.TradeNetworkModifiers)
	applyFlags(gs, eff)
	return targetRegionID
}

func applyTradeNetworkModifiers(gs *state.GameState, modifiers []TradeNetworkModifierEffect) {
	if gs == nil || len(modifiers) == 0 {
		return
	}
	for _, modifier := range modifiers {
		if modifier.ID == "" {
			continue
		}
		converted := state.TradeNetworkModifier{
			ID:                 modifier.ID,
			CenterIDs:          make([]world.RegionID, 0, len(modifier.CenterIDs)),
			RegionIDs:          make([]world.RegionID, 0, len(modifier.RegionIDs)),
			TradeIncomePercent: modifier.TradeIncomePercent,
			SpicePercent:       modifier.SpicePercent,
		}
		for _, id := range modifier.CenterIDs {
			if id != "" {
				converted.CenterIDs = append(converted.CenterIDs, world.RegionID(id))
			}
		}
		for _, id := range modifier.RegionIDs {
			if id != "" {
				converted.RegionIDs = append(converted.RegionIDs, world.RegionID(id))
			}
		}
		found := false
		for i := range gs.TradeNetworkModifiers {
			if gs.TradeNetworkModifiers[i].ID == converted.ID {
				gs.TradeNetworkModifiers[i] = converted
				found = true
				break
			}
		}
		if !found {
			gs.TradeNetworkModifiers = append(gs.TradeNetworkModifiers, converted)
		}
	}
}

// applyToFaction bir fraksiyonun tüm bölgelerine ve hazinesine olay etkilerini uygular.
func applyToFaction(gs *state.GameState, fid string, eff Effect) {
	for _, r := range gs.Regions {
		if r.IsSea || r.OwnerID != fid {
			continue
		}
		r.Satisfaction = clamp(r.Satisfaction+eff.SatDelta, 0, 100)
		applyPopulationDelta(r, eff.PopulationDelta)
	}
	if f, ok := gs.Factions[faction.FactionID(fid)]; ok {
		f.Gold = max0(f.Gold + eff.GoldDelta)
		f.Grain = max0(f.Grain + eff.GrainDelta)
	}
	if eff.ArmyHPMod > 0 && eff.ArmyHPMod < 1.0 {
		for _, a := range gs.Armies {
			if a.OwnerID != fid {
				continue
			}
			for i := range a.Units {
				a.Units[i].CurrentHP = max0(int(float64(a.Units[i].CurrentHP) * eff.ArmyHPMod))
			}
		}
	}
	if eff.RelationDeltaAll != 0 {
		applyRelationDeltaAll(gs, faction.FactionID(fid), eff.RelationDeltaAll)
	}
	applyCompletedTechs(gs, faction.FactionID(fid), eff.CompleteTechs)
	applyStartedResearch(gs, faction.FactionID(fid), eff.StartResearchTech)
	applyRelationEffects(gs, faction.FactionID(fid), eff.Relations)
	applyCapitalMove(gs, faction.FactionID(fid), eff.CapitalSettlementID, eff.CapitalMoveTurns)
}

func applyRelationDeltaAll(gs *state.GameState, fid faction.FactionID, delta int) {
	if gs == nil || fid == "" || delta == 0 {
		return
	}
	for otherID, other := range gs.Factions {
		if otherID == fid || other == nil || other.IsEliminated {
			continue
		}
		rel := diplomacy.EnsureRelation(gs, fid, otherID)
		rel.Score = clamp(rel.Score+delta, -100, 100)
	}
}

func applyFlags(gs *state.GameState, eff Effect) {
	if gs == nil {
		return
	}
	if gs.FiredEventIDs == nil {
		gs.FiredEventIDs = make(map[string]bool)
	}
	for _, flag := range eff.SetFlags {
		if flag == "" {
			continue
		}
		gs.FiredEventIDs[eventFlagKey(flag)] = true
	}
	for _, flag := range eff.ClearFlags {
		if flag == "" {
			continue
		}
		delete(gs.FiredEventIDs, eventFlagKey(flag))
	}
}

func applyCompletedTechs(gs *state.GameState, fid faction.FactionID, techIDs []string) {
	if gs == nil || fid == "" || len(techIDs) == 0 {
		return
	}
	f := gs.Factions[fid]
	if f == nil {
		return
	}
	if f.Research.Completed == nil {
		f.Research.Completed = make(map[string]bool)
	}
	for _, techID := range techIDs {
		if techID == "" {
			continue
		}
		if gs.TechTypes != nil {
			if _, ok := gs.TechTypes[techID]; !ok {
				continue
			}
		}
		f.Research.Completed[techID] = true
		if f.Research.ActiveID == techID {
			f.Research.ActiveID = ""
			f.Research.TurnsLeft = 0
		}
	}
}

func applyStartedResearch(gs *state.GameState, fid faction.FactionID, techID string) {
	if gs == nil || fid == "" || techID == "" || gs.TechTypes == nil {
		return
	}
	f := gs.Factions[fid]
	if f == nil {
		return
	}
	if f.Research.Completed != nil && f.Research.Completed[techID] {
		return
	}
	if f.Research.ActiveID != "" {
		return
	}
	t, ok := gs.TechTypes[techID]
	if !ok || t == nil {
		return
	}
	ownedRegions := make(map[string]bool)
	for _, region := range gs.LandRegionsOwnedBy(fid) {
		ownedRegions[string(region.ID)] = true
	}
	if !tech.IsUnlockedForContext(&f.Research, t, gs.Year, ownedRegions) {
		return
	}
	if f.Research.Completed == nil {
		f.Research.Completed = make(map[string]bool)
	}
	f.Research.ActiveID = techID
	f.Research.TurnsLeft = t.TurnsRequired
}

func applyRelationEffects(gs *state.GameState, fid faction.FactionID, rels []RelationEffect) {
	if gs == nil || fid == "" || len(rels) == 0 {
		return
	}
	for _, rel := range rels {
		if rel.FactionID == "" {
			continue
		}
		var stance faction.DiplomaticStance
		switch rel.Stance {
		case string(faction.StanceWar):
			stance = faction.StanceWar
		case string(faction.StancePeace):
			stance = faction.StancePeace
		case string(faction.StanceAllied):
			stance = faction.StanceAllied
		case string(faction.StanceTrade):
			stance = faction.StanceTrade
		}
		diplomacy.ForceRelation(gs, fid, faction.FactionID(rel.FactionID), stance, rel.ScoreDelta)
	}
}

func applyCoalition(gs *state.GameState, coalition *CoalitionEffect) {
	if gs == nil || coalition == nil || len(coalition.Members) == 0 {
		return
	}
	memberStance := parseCoalitionStance(coalition.MemberStance, faction.StanceAllied)
	opponentStance := parseCoalitionStance(coalition.OpponentStance, faction.StanceWar)
	for i, leftID := range coalition.Members {
		left := faction.FactionID(leftID)
		if gs.Factions[left] == nil {
			continue
		}
		for _, rightID := range coalition.Members[i+1:] {
			right := faction.FactionID(rightID)
			if gs.Factions[right] == nil {
				continue
			}
			diplomacy.ForceRelation(gs, left, right, memberStance, coalition.MemberScoreDelta)
		}
		for _, opponentID := range coalition.Opponents {
			opponent := faction.FactionID(opponentID)
			if gs.Factions[opponent] == nil {
				continue
			}
			diplomacy.ForceRelation(gs, left, opponent, opponentStance, coalition.OpponentScoreDelta)
		}
	}
}

func parseCoalitionStance(value string, fallback faction.DiplomaticStance) faction.DiplomaticStance {
	switch faction.DiplomaticStance(value) {
	case faction.StanceWar, faction.StancePeace, faction.StanceAllied, faction.StanceTrade:
		return faction.DiplomaticStance(value)
	default:
		return fallback
	}
}

func applyCapitalMove(gs *state.GameState, fid faction.FactionID, settlementID string, turns int) {
	if gs == nil || fid == "" || settlementID == "" {
		return
	}
	gs.StartCapitalMove(fid, settlementID, turns)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

// addRegionEventStatus event uygulandıktan sonra etkilenen bölgelere
// harita üzerinde ikon gösterimi için RegionEventStatus kaydı ekler.
func addRegionEventStatus(gs *state.GameState, e *Event, choice *Choice, targetRegionID world.RegionID) {
	if gs == nil || e == nil {
		return
	}

	eventType := eventIconType(e)
	effect := effectiveEventEffect(e, choice)
	labelTR := e.NameTR
	if choice != nil {
		labelTR = e.NameTR + " — " + choice.LabelTR
	}

	// Hangi bölgelerin etkilendiğini belirle
	affectedRegions := affectedRegionIDs(gs, e, choice, targetRegionID)

	// Her etkilenen bölge için status kaydı ekle (3-6 tur görünür)
	turnsVisible := 4
	if eventType == "blessing" {
		turnsVisible = 3 // pozitif olaylar daha kısa görünür
	} else if eventType == "plague" || eventType == "revolt" {
		turnsVisible = 6 // negatif olaylar daha uzun görünür
	}

	for _, rid := range affectedRegions {
		// Aynı bölgede aynı event zaten varsa atla
		exists := false
		for _, existing := range gs.ActiveRegionEvents {
			if existing.RegionID == rid && existing.EventID == e.ID {
				exists = true
				break
			}
		}
		if exists {
			continue
		}

		gs.ActiveRegionEvents = append(gs.ActiveRegionEvents, state.RegionEventStatus{
			EventID:                   e.ID,
			RegionID:                  rid,
			TurnsLeft:                 turnsVisible,
			Type:                      eventType,
			LabelTR:                   labelTR,
			GrainProductionPercent:    effect.GrainProductionPercent,
			GrainDemandPercent:        effect.GrainDemandPercent,
			TradeIncomePercent:        effect.TradeIncomePercent,
			RegionGoldIncomePercent:   effect.RegionGoldIncomePercent,
			ArmyUpkeepPercent:         effect.ArmyUpkeepPercent,
			BuildingEfficiencyPercent: effect.BuildingEfficiencyPercent,
			CombatAttackPercent:       effect.CombatAttackPercent,
			CombatDefensePercent:      effect.CombatDefensePercent,
		})
	}
}

func applyPopulationDelta(region *world.Region, delta int) {
	if region == nil || delta == 0 {
		return
	}
	region.RuralPopulation = max0(region.RuralPopulation + delta)
	region.Population = max0(region.Population + delta)
}

// eventIconType bir event'in haritada hangi ikon tipiyle gösterileceğini belirler.
func eventIconType(e *Event) string {
	id := strings.ToLower(e.ID)
	name := strings.ToLower(e.NameTR)
	desc := strings.ToLower(e.DescTR)

	if strings.Contains(id, "plague") || strings.Contains(name, "veba") || strings.Contains(name, "salgın") || strings.Contains(desc, "veba") || strings.Contains(desc, "salgın") {
		return "plague"
	}
	if strings.Contains(id, "famine") || strings.Contains(id, "drought") || strings.Contains(id, "bad_harvest") || strings.Contains(name, "kıtlık") || strings.Contains(name, "kurak") || strings.Contains(name, "kötü hasat") || strings.Contains(desc, "kıtlık") || strings.Contains(desc, "kurak") {
		return "famine"
	}
	if strings.Contains(id, "harvest") || strings.Contains(name, "hasat") {
		return "blessing"
	}
	if strings.Contains(id, "revolt") || strings.Contains(name, "isyan") || strings.Contains(desc, "isyan") || strings.Contains(name, "taht") || strings.Contains(name, "krizi") {
		return "revolt"
	}
	if strings.Contains(id, "golden") || strings.Contains(id, "trade_boom") || strings.Contains(name, "altın") || strings.Contains(name, "ticaret") || strings.Contains(name, "patlama") {
		return "blessing"
	}
	return "notification"
}

// affectedRegionIDs event ve choice'tan etkilenen bölge ID'lerini döner.
func affectedRegionIDs(gs *state.GameState, e *Event, choice *Choice, targetRegionID world.RegionID) []world.RegionID {
	if gs == nil || e == nil {
		return nil
	}

	target := e.Target
	affectedFaction := e.AffectedFaction
	if choice != nil {
		if choice.Effect.Target != "" {
			target = choice.Effect.Target
		}
		if choice.Effect.AffectedFaction != "" {
			affectedFaction = choice.Effect.AffectedFaction
		}
	}

	switch target {
	case "random_region":
		if targetRegionID != "" {
			return []world.RegionID{targetRegionID}
		}
	case "player_faction":
		return factionOwnerRegions(gs, string(gs.PlayerFactionID))
	case "specific_faction":
		if affectedFaction != "" {
			return factionOwnerRegions(gs, affectedFaction)
		}
	case "all_factions":
		var all []world.RegionID
		for _, r := range gs.Regions {
			if !r.IsSea && r.OwnerID != "" {
				all = append(all, r.ID)
			}
		}
		sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
		return all
	case "all_armies":
		var all []world.RegionID
		seen := make(map[world.RegionID]struct{}, len(gs.Armies))
		for _, a := range gs.Armies {
			if a == nil {
				continue
			}
			rid := a.RegionID
			if a.IsNaval && a.DockedRegionID != "" {
				rid = a.DockedRegionID
			}
			region := gs.Regions[rid]
			if region == nil || region.IsSea {
				continue
			}
			if _, ok := seen[rid]; ok {
				continue
			}
			seen[rid] = struct{}{}
			all = append(all, rid)
		}
		sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
		return all
	}
	return nil
}

// factionOwnerRegions bir fraksiyonun sahip olduğu kara bölgelerini döner.
func factionOwnerRegions(gs *state.GameState, fid string) []world.RegionID {
	var regions []world.RegionID
	for _, r := range gs.Regions {
		if !r.IsSea && r.OwnerID == fid {
			regions = append(regions, r.ID)
		}
	}
	sort.Slice(regions, func(i, j int) bool { return regions[i] < regions[j] })
	return regions
}

// TickActiveRegionEvents her tur çözümlemesinde çağrılır.
// ActiveRegionEvents listesindeki TurnsLeft değerlerini azaltır,
// süresi dolanları temizler.
func TickActiveRegionEvents(gs *state.GameState) {
	if gs == nil || len(gs.ActiveRegionEvents) == 0 {
		return
	}

	kept := gs.ActiveRegionEvents[:0]
	for i := range gs.ActiveRegionEvents {
		gs.ActiveRegionEvents[i].TurnsLeft--
		if gs.ActiveRegionEvents[i].TurnsLeft > 0 {
			kept = append(kept, gs.ActiveRegionEvents[i])
		}
	}
	gs.ActiveRegionEvents = kept
}
