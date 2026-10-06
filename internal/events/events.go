package events

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"

	"mapp-game-go/internal/army"
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

// SuccessorRevivalEffect tarihsel bir event'in elenmiş ardıl faction'ını
// bir veya daha fazla bölgede yeniden kurmasını tanımlar.
type SuccessorRevivalEffect struct {
	FactionID string   `json:"faction_id"`
	Regions   []string `json:"regions,omitempty"`
	// RegionID eski event kayıtlarıyla uyumluluk için korunur. Yeni kayıtlar
	// Regions kullanmalıdır.
	RegionID         string                    `json:"region_id,omitempty"`
	Mode             string                    `json:"mode,omitempty"` // independent | vassal
	OverlordID       string                    `json:"overlord_id,omitempty"`
	UnitType         string                    `json:"unit_type,omitempty"` // boşsa legacy milis
	UnitCount        int                       `json:"unit_count,omitempty"`
	MilitiaCount     int                       `json:"militia_count,omitempty"`
	Units            []UnitReinforcementEffect `json:"units,omitempty"`
	SuppressRelation bool                      `json:"suppress_relation,omitempty"`
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

// UnitReinforcementEffect, event sonucu bir faction'a verilecek yeni birlikleri
// tanımlar. Her kayıt ayrı bir kara ordusu veya filo olarak oluşturulur.
type UnitReinforcementEffect struct {
	UnitType  string `json:"unit_type"`
	UnitCount int    `json:"unit_count"`
}

// ArmyDefectionEffect, bir event sırasında mevcut bir ordu grubunun başka bir
// faction'a katılmasını tanımlar. Seçim deterministik olsun diye uygun ordular
// ArmyID sırasına göre değerlendirilir; hedef bölge verilirse ordu olayın
// siyasi merkezine intikal etmiş kabul edilir.
type ArmyDefectionEffect struct {
	SourceFactionID     string           `json:"source_faction_id"`
	RecipientFactionID  string           `json:"recipient_faction_id"`
	SourceRegionIDs     []world.RegionID `json:"source_region_ids,omitempty"`
	DestinationRegionID world.RegionID   `json:"destination_region_id,omitempty"`
	ArmyCount           int              `json:"army_count,omitempty"`
	ArmyPercent         int              `json:"army_percent,omitempty"`
	IncludeNaval        bool             `json:"include_naval,omitempty"`
}

// DynasticSettlementEffect, bir hanedan anlaşmasıyla belirli bölgelerin
// savaş dışı aktarılmasını tanımlar. Tam siyasi birleşmeden farklı olarak
// kaynak faction aktarım sonrasında yaşamaya devam edebilir.
type DynasticSettlementEffect struct {
	SourceFactionID          string           `json:"source_faction_id"`
	RecipientFactionID       string           `json:"recipient_faction_id"`
	Mode                     string           `json:"mode,omitempty"`
	RegionIDs                []world.RegionID `json:"region_ids"`
	ArmyTransferPercent      int              `json:"army_transfer_percent,omitempty"`
	ResourceTransferPercent  int              `json:"resource_transfer_percent,omitempty"`
	RelationStance           string           `json:"relation_stance,omitempty"`
	RelationScoreDelta       int              `json:"relation_score_delta,omitempty"`
	AutoUnionWhenSourceEmpty bool             `json:"auto_union_when_source_empty,omitempty"`
	UnionResultFactionID     string           `json:"union_result_faction_id,omitempty"`
}

// ImperialSuccessionEffect tarihsel bir event'in HRE imparatorunu belirleyip
// sonraki elektör seçimini kapatmasını tanımlar. Elektör üyeleri state'te
// korunur; yalnızca seçim takvimi bu tarihsel hanedan sonucuna bağlanır.
type ImperialSuccessionEffect struct {
	EmperorID      string `json:"emperor_id"`
	ElectionLocked bool   `json:"election_locked,omitempty"`
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
	OtherIncomeDelta          int                          `json:"other_income_delta,omitempty"` // kalıcı devlet düzeyi gelir değişimi
	GrainDelta                int                          `json:"grain_delta,omitempty"`
	ArmyHPMod                 float64                      `json:"army_hp_mod,omitempty"`                // 1.0 = değişmez
	GrainProductionPercent    int                          `json:"grain_production_percent,omitempty"`   // aktif olay süresince üretim etkisi
	GrainDemandPercent        int                          `json:"grain_demand_percent,omitempty"`       // aktif olay süresince sivil tüketim etkisi
	TradeIncomePercent        int                          `json:"trade_income_percent,omitempty"`       // aktif olay süresince ticaret geliri etkisi
	RegionGoldIncomePercent   int                          `json:"region_gold_income_percent,omitempty"` // aktif olay süresince vergi geliri etkisi
	ArmyUpkeepPercent         int                          `json:"army_upkeep_percent,omitempty"`        // aktif bölgede ordu ikmal etkisi
	EffectDurationTurns       int                          `json:"effect_duration_turns,omitempty"`      // bölgesel etkinin süresi
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
	UnitReinforcements        []UnitReinforcementEffect    `json:"unit_reinforcements,omitempty"`
	ArmyDefections            []ArmyDefectionEffect        `json:"army_defections,omitempty"`
	DynasticSettlement        *DynasticSettlementEffect    `json:"dynastic_settlement,omitempty"`
	ImperialSuccession        *ImperialSuccessionEffect    `json:"imperial_succession,omitempty"`
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
	OtherIncomeDelta          int                          `json:"other_income_delta,omitempty"`
	GrainDelta                int                          `json:"grain_delta"`
	ArmyHPMod                 float64                      `json:"army_hp_mod"` // 1.0 = değişmez
	GrainProductionPercent    int                          `json:"grain_production_percent,omitempty"`
	GrainDemandPercent        int                          `json:"grain_demand_percent,omitempty"`
	TradeIncomePercent        int                          `json:"trade_income_percent,omitempty"`
	RegionGoldIncomePercent   int                          `json:"region_gold_income_percent,omitempty"`
	ArmyUpkeepPercent         int                          `json:"army_upkeep_percent,omitempty"`
	EffectDurationTurns       int                          `json:"effect_duration_turns,omitempty"`
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
	UnitReinforcements        []UnitReinforcementEffect    `json:"unit_reinforcements,omitempty"`
	ArmyDefections            []ArmyDefectionEffect        `json:"army_defections,omitempty"`
	DynasticSettlement        *DynasticSettlementEffect    `json:"dynastic_settlement,omitempty"`

	// Tarihsel tetiklenme alanları
	HistoricalYear       int    `json:"historical_year,omitempty"`        // 0 = tarihsel değil
	HistoricalMonth      int    `json:"historical_month,omitempty"`       // 0 = yılın herhangi bir ayı
	HistoricalDateStrict bool   `json:"historical_date_strict,omitempty"` // tarihsel zincirin state görünürlük işareti
	OneShot              bool   `json:"one_shot,omitempty"`               // true = yalnızca bir kez tetiklenir
	AffectedFaction      string `json:"affected_faction,omitempty"`       // belirli fraksiyonu hedefle

	ChoicePromptTR            string                     `json:"choice_prompt_tr,omitempty"`
	Choices                   []Choice                   `json:"choices,omitempty"`
	PlayerChoiceFactions      []string                   `json:"player_choice_factions,omitempty"`
	RequiresFlags             []string                   `json:"requires_flags,omitempty"`
	BlocksFlags               []string                   `json:"blocks_flags,omitempty"`
	BlocksCapitalSettlementID string                     `json:"blocks_capital_settlement_id,omitempty"`
	RequiresTechs             []string                   `json:"requires_techs,omitempty"`
	BlocksTechs               []string                   `json:"blocks_techs,omitempty"`
	RequiresOwnedRegions      []world.RegionID           `json:"requires_owned_regions,omitempty"`
	RequiresOwnedRegionsAny   []world.RegionID           `json:"requires_owned_regions_any,omitempty"`
	RequiresUnownedRegions    []world.RegionID           `json:"requires_unowned_regions,omitempty"`
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

// TickOpeningHistoricalEvent yeni oyun açılırken başlangıç tarihine denk gelen
// tarihsel event'i işler. Rastgele veya siyasi üstünlük event'lerini çalıştırmaz;
// böylece senaryo başlangıç olayları ilk oyuncu turundan önce uygulanabilir.
func TickOpeningHistoricalEvent(gs *state.GameState, evts []*Event) *Event {
	if gs == nil {
		return nil
	}
	if gs.FiredEventIDs == nil {
		gs.FiredEventIDs = make(map[string]bool)
	}

	for _, e := range evts {
		if e == nil || e.HistoricalYear == 0 || gs.FiredEventIDs[e.ID] ||
			!eventConditionsSatisfied(gs, e) || !historicalEventDueThisTurn(gs, e) {
			continue
		}
		if e.OneShot {
			gs.FiredEventIDs[e.ID] = true
		}
		queueHistoricalEventsSharingDate(gs, evts, e)
		return e
	}
	return nil
}

// historicalEventDueThisTurn tarihsel olayın minimum tarihine ulaşılıp
// ulaşılmadığını bildirir. Tarih geçtikten sonra koşulları daha sonraki bir
// turda sağlayan event o turda tetiklenebilir.
func historicalEventDueThisTurn(gs *state.GameState, e *Event) bool {
	if gs == nil || e == nil {
		return false
	}
	if e.HistoricalYear <= 0 || gs.Year <= 0 {
		return false
	}
	endYear, endMonth := gs.CurrentTurnEndDate()
	if e.HistoricalMonth <= 0 {
		return endYear >= e.HistoricalYear
	}
	if e.HistoricalMonth > 12 {
		return false
	}
	endAbs := endYear*12 + endMonth - 1
	targetAbs := e.HistoricalYear*12 + e.HistoricalMonth - 1
	return endAbs >= targetAbs
}

func historicalEventHasStateTrigger(e *Event) bool {
	if e == nil {
		return false
	}
	if e.HistoricalDateStrict {
		return false
	}
	return len(e.RequiresFlags) > 0 ||
		len(e.RequiresTechs) > 0 ||
		len(e.RequiresOwnedRegions) > 0 ||
		len(e.RequiresUnownedRegions) > 0 ||
		len(e.RelationRequirements) > 0 ||
		e.FactionSubjugationTrigger != nil
}

// HasStateTrigger, Kodex gibi dış tüketicilerin tarihi geçmiş olsa bile state
// koşullarıyla anlamlı bir kilitli giriş olarak gösterilecek event'i ayırt
// etmesini sağlar. Bu yardımcı, Tick'in tarih penceresini bypass etmez.
func HasStateTrigger(e *Event) bool {
	return historicalEventHasStateTrigger(e)
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
		OtherIncomeDelta:          e.OtherIncomeDelta,
		GrainDelta:                e.GrainDelta,
		ArmyHPMod:                 e.ArmyHPMod,
		GrainProductionPercent:    e.GrainProductionPercent,
		GrainDemandPercent:        e.GrainDemandPercent,
		TradeIncomePercent:        e.TradeIncomePercent,
		RegionGoldIncomePercent:   e.RegionGoldIncomePercent,
		ArmyUpkeepPercent:         e.ArmyUpkeepPercent,
		EffectDurationTurns:       e.EffectDurationTurns,
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
		UnitReinforcements:        e.UnitReinforcements,
		ArmyDefections:            e.ArmyDefections,
		DynasticSettlement:        e.DynasticSettlement,
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
		if e.DynasticSettlement != nil && e.DynasticSettlement.RecipientFactionID == string(gs.PlayerFactionID) {
			return true
		}
		for _, defection := range e.ArmyDefections {
			if defection.SourceFactionID == string(gs.PlayerFactionID) || defection.RecipientFactionID == string(gs.PlayerFactionID) {
				return true
			}
		}
		for _, choice := range e.Choices {
			for _, defection := range choice.Effect.ArmyDefections {
				if defection.SourceFactionID == string(gs.PlayerFactionID) || defection.RecipientFactionID == string(gs.PlayerFactionID) {
					return true
				}
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
	eff := e.BaseEffect()
	targetRegionID := applyEffect(gs, eff)
	applyDynasticSettlement(gs, eff.DynasticSettlement)
	applySuccessorRevival(gs, eff)
	applyArmyDefections(gs, eff.ArmyDefections)
	addRegionEventStatus(gs, e, nil, targetRegionID)
}

func ApplyChoice(gs *state.GameState, e *Event, idx int) (Choice, bool) {
	if e == nil || idx < 0 || idx >= len(e.Choices) {
		return Choice{}, false
	}
	choice := e.Choices[idx]
	eff := choiceEffect(e, choice)
	targetRegionID := applyEffect(gs, eff)
	applyDynasticSettlement(gs, eff.DynasticSettlement)
	applySuccessorRevival(gs, eff)
	applyArmyDefections(gs, eff.ArmyDefections)
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
	regionIDs := revivalRegionIDs(revival)
	if successorID == "" || len(regionIDs) == 0 {
		return
	}
	var armyID army.ArmyID
	var ok bool
	if len(revival.Units) > 0 {
		units := make([]army.Unit, 0)
		for _, composition := range revival.Units {
			if composition.UnitType == "" || composition.UnitCount <= 0 {
				continue
			}
			units = append(units, army.MakeUnits(composition.UnitType, composition.UnitCount)...)
		}
		if len(units) == 0 {
			return
		}
		armyID, ok = gs.ReviveSuccessorAtRegionsWithUnits(regionIDs, successorID, units)
	} else {
		unitType := revival.UnitType
		unitCount := revival.UnitCount
		if unitType == "" {
			unitType = "militia"
			if unitCount <= 0 {
				unitCount = revival.MilitiaCount
			}
		}
		armyID, ok = gs.ReviveSuccessorAtRegionsWithUnit(regionIDs, successorID, unitType, unitCount)
	}
	if !ok {
		return
	}
	for _, regionID := range regionIDs {
		world.EnsureSuccessorFoundingBuildings(gs.Regions[regionID])
	}
	gs.AssignStrongestCommanderToArmy(armyID)
	overlordID := faction.FactionID(revival.OverlordID)
	if overlordID == "" && revival.Mode == "vassal" {
		overlordID = faction.FactionID(eff.AffectedFaction)
	}
	if revival.Mode == "vassal" && overlordID != "" && gs.Factions[overlordID] != nil {
		successor := gs.Factions[successorID]
		successor.OverlordID = overlordID
		successor.TributeRate = 20
		successor.TributeRateConfigured = true
		if !revival.SuppressRelation {
			diplomacy.ForceRelation(gs, overlordID, successorID, faction.StanceAllied, 50)
		}
	} else if eff.AffectedFaction != "" && !revival.SuppressRelation {
		diplomacy.ForceRelation(gs, faction.FactionID(eff.AffectedFaction), successorID, faction.StanceAllied, 50)
	}
}

func revivalRegionIDs(revival SuccessorRevivalEffect) []world.RegionID {
	if len(revival.Regions) > 0 {
		regionIDs := make([]world.RegionID, 0, len(revival.Regions))
		for _, regionID := range revival.Regions {
			if regionID != "" {
				regionIDs = append(regionIDs, world.RegionID(regionID))
			}
		}
		return regionIDs
	}
	if revival.RegionID != "" {
		return []world.RegionID{world.RegionID(revival.RegionID)}
	}
	return nil
}

func applyArmyDefections(gs *state.GameState, defections []ArmyDefectionEffect) {
	if gs == nil || len(defections) == 0 || gs.Armies == nil || gs.Factions == nil {
		return
	}
	for _, defection := range defections {
		applyOneArmyDefection(gs, defection)
	}
}

func applyOneArmyDefection(gs *state.GameState, defection ArmyDefectionEffect) {
	sourceID := faction.FactionID(defection.SourceFactionID)
	recipientID := faction.FactionID(defection.RecipientFactionID)
	if sourceID == "" || recipientID == "" || sourceID == recipientID {
		return
	}
	if gs.Factions[sourceID] == nil || gs.Factions[recipientID] == nil || gs.Factions[recipientID].IsEliminated {
		return
	}

	allowedRegions := make(map[world.RegionID]struct{}, len(defection.SourceRegionIDs))
	for _, regionID := range defection.SourceRegionIDs {
		if regionID != "" {
			allowedRegions[regionID] = struct{}{}
		}
	}

	candidates := make([]army.ArmyID, 0)
	for armyID, current := range gs.Armies {
		if current == nil || current.OwnerID != string(sourceID) {
			continue
		}
		if current.IsNaval && !defection.IncludeNaval {
			continue
		}
		if len(allowedRegions) > 0 {
			locationID := current.RegionID
			if current.IsNaval && current.DockedRegionID != "" {
				locationID = current.DockedRegionID
			}
			if _, ok := allowedRegions[locationID]; !ok {
				continue
			}
		}
		candidates = append(candidates, armyID)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i] < candidates[j] })

	transferCount := 0
	if defection.ArmyCount > 0 {
		transferCount = defection.ArmyCount
	} else if defection.ArmyPercent > 0 {
		transferCount = (len(candidates)*clamp(defection.ArmyPercent, 0, 100) + 99) / 100
	}
	if transferCount > len(candidates) {
		transferCount = len(candidates)
	}
	if transferCount <= 0 {
		return
	}

	for _, armyID := range candidates[:transferCount] {
		current := gs.Armies[armyID]
		if current == nil {
			continue
		}
		if destination := gs.Regions[defection.DestinationRegionID]; destination != nil && current.IsNaval == destination.IsSea {
			current.PreviousRegionID = current.RegionID
			current.RegionID = destination.ID
			current.MovePoints = 0
			if current.IsNaval {
				current.DockedRegionID = ""
				current.DockedSettlementID = ""
			}
		}
		gs.TransferArmyOwnership(current, string(recipientID))
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
	if len(e.RequiresOwnedRegionsAny) > 0 {
		fid := eventConditionFactionID(gs, e)
		if fid == "" {
			reasons = append(reasons, "herhangi bir sahip olunan bölge koşulu için fraksiyon yok")
		} else if !anyOwnedRegionSatisfied(gs, fid, e.RequiresOwnedRegionsAny) {
			reasons = append(reasons, "gerekli bölgelerden en az biri sahiplenilmeli")
		}
	}
	if len(e.RequiresUnownedRegions) > 0 {
		fid := eventConditionFactionID(gs, e)
		if fid == "" {
			reasons = append(reasons, "sahipsiz olmasi gereken bolge kosulu icin fraksiyon yok")
		} else {
			for _, rid := range e.RequiresUnownedRegions {
				r := gs.Regions[rid]
				if r != nil && r.OwnerID == string(fid) {
					reasons = append(reasons, "bolge zaten hedef faction'da: "+string(rid))
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
	if len(e.RequiresOwnedRegionsAny) > 0 {
		fid := eventConditionFactionID(gs, e)
		if fid == "" || !anyOwnedRegionSatisfied(gs, fid, e.RequiresOwnedRegionsAny) {
			return false
		}
	}
	if len(e.RequiresUnownedRegions) > 0 {
		fid := eventConditionFactionID(gs, e)
		if fid == "" {
			return false
		}
		for _, rid := range e.RequiresUnownedRegions {
			r := gs.Regions[rid]
			if r != nil && r.OwnerID == string(fid) {
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

func anyOwnedRegionSatisfied(gs *state.GameState, fid faction.FactionID, regionIDs []world.RegionID) bool {
	if gs == nil || fid == "" {
		return false
	}
	for _, rid := range regionIDs {
		r := gs.Regions[rid]
		if r != nil && !r.IsSea && !r.IsTerrainArea && r.OwnerID == string(fid) {
			return true
		}
	}
	return false
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
		score = diplomacy.RelationScore(gs, source, target)
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
		score = diplomacy.RelationScore(gs, source, target)
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

func applyDynasticSettlement(gs *state.GameState, settlement *DynasticSettlementEffect) {
	if gs == nil || settlement == nil {
		return
	}
	sourceID := faction.FactionID(settlement.SourceFactionID)
	recipientID := faction.FactionID(settlement.RecipientFactionID)
	if sourceID == "" || recipientID == "" || sourceID == recipientID {
		return
	}
	source := gs.Factions[sourceID]
	recipient := gs.Factions[recipientID]
	if source == nil || recipient == nil {
		return
	}

	for _, regionID := range settlement.RegionIDs {
		region := gs.Regions[regionID]
		if region == nil || region.IsSea || region.OwnerID != string(sourceID) {
			continue
		}
		region.OwnerID = string(recipientID)
	}

	resourcePercent := clamp(settlement.ResourceTransferPercent, 0, 100)
	transferFactionResources(source, recipient, resourcePercent)

	armyPercent := clamp(settlement.ArmyTransferPercent, 0, 100)
	if armyPercent > 0 {
		armyIDs := make([]army.ArmyID, 0)
		for armyID, current := range gs.Armies {
			if current == nil || current.OwnerID != string(sourceID) || current.IsNaval {
				continue
			}
			armyIDs = append(armyIDs, armyID)
		}
		sort.Slice(armyIDs, func(i, j int) bool { return armyIDs[i] < armyIDs[j] })
		transferCount := (len(armyIDs)*armyPercent + 99) / 100
		if transferCount > len(armyIDs) {
			transferCount = len(armyIDs)
		}
		for _, armyID := range armyIDs[:transferCount] {
			gs.Armies[armyID].OwnerID = string(recipientID)
		}
	}

	if stance := dynasticSettlementStance(settlement.RelationStance); stance != "" {
		diplomacy.ForceRelation(gs, sourceID, recipientID, stance, settlement.RelationScoreDelta)
	}

	if !settlement.AutoUnionWhenSourceEmpty || len(gs.LandRegionsOwnedBy(sourceID)) != 0 {
		gs.NormalizeFactionCapitals()
		return
	}
	resultID := recipientID
	if settlement.UnionResultFactionID != "" {
		resultID = faction.FactionID(settlement.UnionResultFactionID)
	}
	if gs.Factions[resultID] == nil {
		return
	}
	// Bu birleşme savaş sonucu değildir: WarLedger ve fetih istatistikleri
	// değiştirilmeden mevcut siyasi birleşme state yardımcısı kullanılır.
	gs.UnitePoliticalFactions([]faction.FactionID{resultID, sourceID}, resultID)
}

func transferFactionResources(source, recipient *faction.Faction, percent int) {
	if source == nil || recipient == nil || percent <= 0 {
		return
	}
	transfer := func(value *int, target *int) {
		amount := *value * percent / 100
		*value -= amount
		*target += amount
	}
	transfer(&source.Gold, &recipient.Gold)
	transfer(&source.Grain, &recipient.Grain)
	transfer(&source.Iron, &recipient.Iron)
	transfer(&source.Timber, &recipient.Timber)
	transfer(&source.Stone, &recipient.Stone)
	transfer(&source.Spice, &recipient.Spice)
	transfer(&source.Cloth, &recipient.Cloth)
}

func dynasticSettlementStance(value string) faction.DiplomaticStance {
	switch faction.DiplomaticStance(value) {
	case faction.StancePeace, faction.StanceAllied, faction.StanceTrade:
		return faction.DiplomaticStance(value)
	default:
		return ""
	}
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
		f.OtherIncomeDelta += eff.OtherIncomeDelta
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
	applyUnitReinforcements(gs, fid, eff.UnitReinforcements)
}

func applyUnitReinforcements(gs *state.GameState, ownerID string, reinforcements []UnitReinforcementEffect) {
	if gs == nil || ownerID == "" || len(reinforcements) == 0 || gs.Factions[faction.FactionID(ownerID)] == nil {
		return
	}
	if gs.Armies == nil {
		gs.Armies = make(map[army.ArmyID]*army.Army)
	}

	capitalRegion, capitalSettlement, _, ok := gs.FactionCapital(faction.FactionID(ownerID))
	if !ok || capitalRegion == nil || capitalSettlement == nil {
		return
	}
	for _, reinforcement := range reinforcements {
		if reinforcement.UnitType == "" || reinforcement.UnitCount <= 0 {
			continue
		}
		unitType := gs.UnitTypes[reinforcement.UnitType]
		if unitType == nil {
			continue
		}
		count := reinforcement.UnitCount
		if count > army.MaxArmySize {
			count = army.MaxArmySize
		}
		gs.NextArmySeq++
		id := army.ArmyID(fmt.Sprintf("army_%s_event_%d", ownerID, gs.NextArmySeq))
		for gs.Armies[id] != nil {
			gs.NextArmySeq++
			id = army.ArmyID(fmt.Sprintf("army_%s_event_%d", ownerID, gs.NextArmySeq))
		}

		newArmy := &army.Army{
			ID:            id,
			OwnerID:       ownerID,
			RegionID:      capitalRegion.ID,
			Units:         army.MakeUnits(reinforcement.UnitType, count),
			MovePoints:    unitType.BaseMovementPoints(),
			MaxMovePoints: unitType.BaseMovementPoints(),
		}
		if unitType.Category == army.CategoryNavalWar || unitType.Category == army.CategoryNavalTrans || unitType.Category == army.CategoryNavalTrade {
			seaRegionID := firstAdjacentSeaRegion(gs, capitalRegion)
			if seaRegionID == "" {
				continue
			}
			newArmy.IsNaval = true
			newArmy.RegionID = seaRegionID
			newArmy.DockedRegionID = capitalRegion.ID
			newArmy.DockedSettlementID = capitalSettlement.ID
		}
		gs.Armies[id] = newArmy
	}
}

func firstAdjacentSeaRegion(gs *state.GameState, region *world.Region) world.RegionID {
	if gs == nil || region == nil {
		return ""
	}
	for _, neighborID := range region.Neighbors {
		neighbor := gs.Regions[neighborID]
		if neighbor != nil && neighbor.IsSea {
			return neighbor.ID
		}
	}
	return ""
}

func applyRelationDeltaAll(gs *state.GameState, fid faction.FactionID, delta int) {
	if gs == nil || fid == "" || delta == 0 {
		return
	}
	for otherID, other := range gs.Factions {
		if otherID == fid || other == nil || other.IsEliminated {
			continue
		}
		diplomacy.EnsureRelation(gs, fid, otherID)
		diplomacy.AddRelationScoreBoth(gs, fid, otherID, delta)
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
	if effect.EffectDurationTurns > 0 {
		turnsVisible = effect.EffectDurationTurns
	}
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
