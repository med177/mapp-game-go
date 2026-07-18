package save

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/klauspost/compress/zstd"
)

const saveStateEncodingZstdBase64 = "zstd+base64"

type stringList []string

type settlementPatch struct {
	Replace []world.Settlement `json:"r,omitempty"`
	Order   []string           `json:"o,omitempty"`
	Added   []world.Settlement `json:"a,omitempty"`
	Updated []world.Settlement `json:"u,omitempty"`
	Removed []string           `json:"d,omitempty"`
}

type regionSaveState struct {
	OwnerID         *string          `json:"o,omitempty"`
	Settlements     *settlementPatch `json:"sp,omitempty"`
	IsLocked        *bool            `json:"l,omitempty"`
	Satisfaction    *int             `json:"sat,omitempty"`
	TaxRate         *int             `json:"tx,omitempty"`
	Population      *int             `json:"pop,omitempty"`
	Religion        *string          `json:"rel,omitempty"`
	ConversionTurns *int             `json:"conv,omitempty"`
	ActiveEventID   *string          `json:"evt,omitempty"`
	Buildings       *stringList      `json:"b,omitempty"`
}

type factionSaveState struct {
	IsEliminated               *bool                  `json:"el,omitempty"`
	OverlordID                 *faction.FactionID     `json:"ov,omitempty"`
	CapitalSettlementID        *string                `json:"cap,omitempty"`
	PendingCapitalSettlementID *string                `json:"pcap,omitempty"`
	PendingCapitalTurns        *int                   `json:"pct,omitempty"`
	Gold                       *int                   `json:"g,omitempty"`
	Grain                      *int                   `json:"gr,omitempty"`
	Iron                       *int                   `json:"ir,omitempty"`
	Timber                     *int                   `json:"tm,omitempty"`
	Stone                      *int                   `json:"st,omitempty"`
	Spice                      *int                   `json:"sp,omitempty"`
	Cloth                      *int                   `json:"cl,omitempty"`
	Research                   *faction.ResearchState `json:"rs,omitempty"`
}

type relationSaveState struct {
	Score   *int   `json:"s,omitempty"`
	Stance  *uint8 `json:"t,omitempty"`
	Deleted bool   `json:"x,omitempty"`
}

type stackedUnitSaveState struct {
	TypeID     string `json:"t"`
	Count      int    `json:"c"`
	CurrentHP  int    `json:"h,omitempty"`
	Experience int    `json:"x,omitempty"`
}

type armySaveState struct {
	OwnerID            string                 `json:"o"`
	RegionID           world.RegionID         `json:"r"`
	DockedRegionID     world.RegionID         `json:"dr,omitempty"`
	DockedSettlementID string                 `json:"ds,omitempty"`
	Units              []stackedUnitSaveState `json:"u"`
	EmbarkedUnits      []stackedUnitSaveState `json:"eu,omitempty"`
	MovePoints         int                    `json:"mp"`
	MaxMovePoints      int                    `json:"mm"`
	IsNaval            bool                   `json:"n,omitempty"`
	IsGarrison         bool                   `json:"g,omitempty"`
	Commander          *army.Commander        `json:"c,omitempty"`
	EmbarkedCommander  *army.Commander        `json:"ec,omitempty"`
	InAmbush           bool                   `json:"a,omitempty"`
	OverCapacityTurns  int                    `json:"oc,omitempty"`
	TurnsWithoutPort   int                    `json:"tp,omitempty"`
}

type campaignSaveState struct {
	Turn                    int                                      `json:"t"`
	Year                    int                                      `json:"y"`
	Month                   int                                      `json:"m"`
	StartYear               int                                      `json:"sy,omitempty"`
	ScenarioID              string                                   `json:"sc"`
	ScenarioPath            string                                   `json:"scp,omitempty"`
	PlayerFactionID         faction.FactionID                        `json:"pf"`
	Difficulty              int                                      `json:"d,omitempty"`
	DevelopmentMode         bool                                     `json:"dev,omitempty"`
	EditMode                bool                                     `json:"em,omitempty"`
	Victory                 state.VictoryCondition                   `json:"v"`
	SelectedVictoryOptionID string                                   `json:"sv,omitempty"`
	Regions                 map[world.RegionID]regionSaveState       `json:"rg,omitempty"`
	Factions                map[faction.FactionID]factionSaveState   `json:"fx,omitempty"`
	Armies                  map[army.ArmyID]armySaveState            `json:"ar,omitempty"`
	Commanders              map[string]*army.Commander               `json:"cmd,omitempty"`
	AIPlans                 map[faction.FactionID]*state.AIPlanState `json:"ap,omitempty"`
	EconomicVictoryTurns    int                                      `json:"evt,omitempty"`
	FactionsEliminated      int                                      `json:"fel,omitempty"`
	ReligiousVictoryTurns   int                                      `json:"rvt,omitempty"`
	VictoryAchieved         bool                                     `json:"va,omitempty"`
	VictoryAchievedTurn     int                                      `json:"vat,omitempty"`
	FiredEventIDs           []string                                 `json:"fe,omitempty"`
	Relations               map[string]relationSaveState             `json:"rl,omitempty"`
	DiplomaticOffers        []state.DiplomaticOffer                  `json:"do,omitempty"`
	DiplomaticOfferHistory  []state.DiplomaticOfferHistoryEntry      `json:"dh,omitempty"`
	DiplomacyOfferCounts    map[faction.FactionID]int                `json:"dq,omitempty"`
	TradeRoutes             []*economy.TradeRoute                    `json:"tr,omitempty"`
	Sieges                  map[world.RegionID]*state.SiegeState     `json:"sg,omitempty"`
	ProductionQueue         []state.ProductionOrder                  `json:"pq,omitempty"`
	NextProductionSeq       int                                      `json:"np,omitempty"`
	NextArmySeq             int                                      `json:"na,omitempty"`
	NextCommanderSeq        int                                      `json:"nc,omitempty"`
	Phase                   state.Phase                              `json:"ph,omitempty"`
	WinnerID                faction.FactionID                        `json:"w,omitempty"`
	ActiveRegionEvents      []state.RegionEventStatus                `json:"ae,omitempty"`
}

type legacyRegionSaveState struct {
	OwnerID         string             `json:"owner_id"`
	Settlements     []world.Settlement `json:"settlements"`
	IsLocked        bool               `json:"is_locked"`
	Satisfaction    int                `json:"satisfaction"`
	TaxRate         int                `json:"tax_rate"`
	Population      int                `json:"population"`
	Religion        string             `json:"religion"`
	ConversionTurns int                `json:"conversion_turns,omitempty"`
	ActiveEventID   string             `json:"active_event_id"`
	Buildings       []string           `json:"buildings"`
}

type legacyFactionSaveState struct {
	IsEliminated               bool                  `json:"is_eliminated"`
	OverlordID                 faction.FactionID     `json:"overlord_id,omitempty"`
	CapitalSettlementID        string                `json:"capital_settlement_id,omitempty"`
	PendingCapitalSettlementID string                `json:"pending_capital_settlement_id,omitempty"`
	PendingCapitalTurns        int                   `json:"pending_capital_turns,omitempty"`
	Gold                       int                   `json:"gold"`
	Grain                      int                   `json:"grain"`
	Iron                       int                   `json:"iron"`
	Timber                     int                   `json:"timber"`
	Stone                      int                   `json:"stone"`
	Spice                      int                   `json:"spice"`
	Cloth                      int                   `json:"cloth"`
	Research                   faction.ResearchState `json:"research"`
}

type legacyCampaignSaveState struct {
	Turn                    int                                          `json:"turn"`
	Year                    int                                          `json:"year"`
	Month                   int                                          `json:"month"`
	StartYear               int                                          `json:"start_year"`
	ScenarioID              string                                       `json:"scenario_id"`
	ScenarioPath            string                                       `json:"scenario_path,omitempty"`
	PlayerFactionID         faction.FactionID                            `json:"player_faction_id"`
	Difficulty              int                                          `json:"difficulty"`
	DevelopmentMode         bool                                         `json:"development_mode"`
	EditMode                bool                                         `json:"edit_mode"`
	Victory                 state.VictoryCondition                       `json:"victory"`
	SelectedVictoryOptionID string                                       `json:"selected_victory_option_id"`
	Regions                 map[world.RegionID]legacyRegionSaveState     `json:"regions"`
	Factions                map[faction.FactionID]legacyFactionSaveState `json:"factions"`
	Armies                  map[army.ArmyID]*army.Army                   `json:"armies"`
	Commanders              map[string]*army.Commander                   `json:"commanders,omitempty"`
	AIPlans                 map[faction.FactionID]*state.AIPlanState     `json:"ai_plans,omitempty"`
	EconomicVictoryTurns    int                                          `json:"economic_victory_turns"`
	FactionsEliminated      int                                          `json:"factions_eliminated"`
	ReligiousVictoryTurns   int                                          `json:"religious_victory_turns"`
	VictoryAchieved         bool                                         `json:"victory_achieved"`
	VictoryAchievedTurn     int                                          `json:"victory_achieved_turn"`
	FiredEventIDs           map[string]bool                              `json:"fired_event_ids"`
	Relations               map[string]*faction.Relation                 `json:"relations"`
	DiplomaticOffers        []state.DiplomaticOffer                      `json:"diplomatic_offers,omitempty"`
	DiplomaticOfferHistory  []state.DiplomaticOfferHistoryEntry          `json:"diplomatic_offer_history,omitempty"`
	DiplomacyOfferCounts    map[faction.FactionID]int                    `json:"diplomacy_offer_counts,omitempty"`
	TradeRoutes             []*economy.TradeRoute                        `json:"trade_routes"`
	Sieges                  map[world.RegionID]*state.SiegeState         `json:"sieges,omitempty"`
	ProductionQueue         []state.ProductionOrder                      `json:"production_queue"`
	NextProductionSeq       int                                          `json:"next_production_seq"`
	NextArmySeq             int                                          `json:"next_army_seq"`
	NextCommanderSeq        int                                          `json:"next_commander_seq,omitempty"`
	Phase                   state.Phase                                  `json:"phase"`
	WinnerID                faction.FactionID                            `json:"winner_id"`
	ActiveRegionEvents      []state.RegionEventStatus                    `json:"active_region_events,omitempty"`
}

func encodeCompressedStatePayload(saved campaignSaveState) (string, string, error) {
	raw, err := json.Marshal(saved)
	if err != nil {
		return "", "", err
	}
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return "", "", err
	}
	compressed := encoder.EncodeAll(raw, nil)
	if err := encoder.Close(); err != nil {
		return "", "", err
	}
	return saveStateEncodingZstdBase64, base64.StdEncoding.EncodeToString(compressed), nil
}

func decodeCompressedStatePayload(encoding, payload string) ([]byte, error) {
	if payload == "" {
		return nil, fmt.Errorf("boş sıkıştırılmış save payload")
	}
	if encoding != "" && encoding != saveStateEncodingZstdBase64 {
		return nil, fmt.Errorf("desteklenmeyen save state encoding: %s", encoding)
	}
	compressed, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	defer decoder.Close()
	return decoder.DecodeAll(compressed, nil)
}

func decodeCampaignSaveState(payload []byte) (campaignSaveState, error) {
	var saved campaignSaveState
	if err := json.Unmarshal(payload, &saved); err == nil && looksLikeCompactSaveState(saved) {
		return saved, nil
	}
	var legacy legacyCampaignSaveState
	if err := json.Unmarshal(payload, &legacy); err != nil {
		return campaignSaveState{}, err
	}
	if !looksLikeLegacySaveState(legacy) {
		return campaignSaveState{}, fmt.Errorf("desteklenmeyen kayıt gövdesi")
	}
	return convertLegacyCampaignSaveState(legacy), nil
}

func looksLikeCompactSaveState(saved campaignSaveState) bool {
	return saved.ScenarioID != "" || saved.ScenarioPath != "" || saved.PlayerFactionID != "" || len(saved.Armies) > 0 || len(saved.Regions) > 0
}

func looksLikeLegacySaveState(saved legacyCampaignSaveState) bool {
	return saved.ScenarioID != "" || saved.ScenarioPath != "" || saved.PlayerFactionID != "" || len(saved.Armies) > 0 || len(saved.Regions) > 0
}

func convertLegacyCampaignSaveState(legacy legacyCampaignSaveState) campaignSaveState {
	savedRegions := make(map[world.RegionID]regionSaveState, len(legacy.Regions))
	for rid, region := range legacy.Regions {
		regionCopy := region
		converted := regionSaveState{
			OwnerID:         cloneStringPtr(regionCopy.OwnerID),
			IsLocked:        cloneBoolPtr(regionCopy.IsLocked),
			Satisfaction:    cloneIntPtr(regionCopy.Satisfaction),
			TaxRate:         cloneIntPtr(regionCopy.TaxRate),
			Population:      cloneIntPtr(regionCopy.Population),
			Religion:        cloneStringPtr(regionCopy.Religion),
			ConversionTurns: cloneIntPtr(regionCopy.ConversionTurns),
			ActiveEventID:   cloneStringPtr(regionCopy.ActiveEventID),
		}
		if regionCopy.Buildings != nil {
			buildings := stringList(append([]string(nil), regionCopy.Buildings...))
			converted.Buildings = &buildings
		}
		if regionCopy.Settlements != nil {
			converted.Settlements = &settlementPatch{
				Replace: append([]world.Settlement(nil), regionCopy.Settlements...),
			}
		}
		savedRegions[rid] = converted
	}

	savedFactions := make(map[faction.FactionID]factionSaveState, len(legacy.Factions))
	for fid, fx := range legacy.Factions {
		factionCopy := fx
		savedFactions[fid] = factionSaveState{
			IsEliminated:               cloneBoolPtr(factionCopy.IsEliminated),
			OverlordID:                 cloneFactionIDPtr(factionCopy.OverlordID),
			CapitalSettlementID:        cloneStringPtr(factionCopy.CapitalSettlementID),
			PendingCapitalSettlementID: cloneStringPtr(factionCopy.PendingCapitalSettlementID),
			PendingCapitalTurns:        cloneIntPtr(factionCopy.PendingCapitalTurns),
			Gold:                       cloneIntPtr(factionCopy.Gold),
			Grain:                      cloneIntPtr(factionCopy.Grain),
			Iron:                       cloneIntPtr(factionCopy.Iron),
			Timber:                     cloneIntPtr(factionCopy.Timber),
			Stone:                      cloneIntPtr(factionCopy.Stone),
			Spice:                      cloneIntPtr(factionCopy.Spice),
			Cloth:                      cloneIntPtr(factionCopy.Cloth),
			Research:                   cloneResearchStatePtr(factionCopy.Research),
		}
	}

	return campaignSaveState{
		Turn:                    legacy.Turn,
		Year:                    legacy.Year,
		Month:                   legacy.Month,
		StartYear:               legacy.StartYear,
		ScenarioID:              legacy.ScenarioID,
		ScenarioPath:            legacy.ScenarioPath,
		PlayerFactionID:         legacy.PlayerFactionID,
		Difficulty:              legacy.Difficulty,
		DevelopmentMode:         legacy.DevelopmentMode,
		EditMode:                legacy.EditMode,
		Victory:                 legacy.Victory,
		SelectedVictoryOptionID: legacy.SelectedVictoryOptionID,
		Regions:                 savedRegions,
		Factions:                savedFactions,
		Armies:                  convertArmiesToSaveState(legacy.Armies),
		Commanders:              cloneCommanders(legacy.Commanders),
		AIPlans:                 cloneAIPlans(legacy.AIPlans),
		EconomicVictoryTurns:    legacy.EconomicVictoryTurns,
		FactionsEliminated:      legacy.FactionsEliminated,
		ReligiousVictoryTurns:   legacy.ReligiousVictoryTurns,
		VictoryAchieved:         legacy.VictoryAchieved,
		VictoryAchievedTurn:     legacy.VictoryAchievedTurn,
		FiredEventIDs:           firedEventIDsToSlice(legacy.FiredEventIDs),
		Relations:               makeLegacyRelationState(legacy.Relations),
		DiplomaticOffers:        append([]state.DiplomaticOffer(nil), legacy.DiplomaticOffers...),
		DiplomaticOfferHistory:  append([]state.DiplomaticOfferHistoryEntry(nil), legacy.DiplomaticOfferHistory...),
		DiplomacyOfferCounts:    cloneFactionIntMap(legacy.DiplomacyOfferCounts),
		TradeRoutes:             cloneTradeRoutes(legacy.TradeRoutes),
		Sieges:                  cloneSieges(legacy.Sieges),
		ProductionQueue:         append([]state.ProductionOrder(nil), legacy.ProductionQueue...),
		NextProductionSeq:       legacy.NextProductionSeq,
		NextArmySeq:             legacy.NextArmySeq,
		NextCommanderSeq:        legacy.NextCommanderSeq,
		Phase:                   legacy.Phase,
		WinnerID:                legacy.WinnerID,
		ActiveRegionEvents:      append([]state.RegionEventStatus(nil), legacy.ActiveRegionEvents...),
	}
}

func makeLegacyRelationState(relations map[string]*faction.Relation) map[string]relationSaveState {
	if len(relations) == 0 {
		return nil
	}
	out := make(map[string]relationSaveState, len(relations))
	for key, rel := range relations {
		if rel == nil {
			continue
		}
		score := rel.Score
		stance := encodeStance(rel.Stance)
		out[key] = relationSaveState{
			Score:  &score,
			Stance: &stance,
		}
	}
	return out
}

func makeCampaignSaveState(gs *state.GameState) (campaignSaveState, error) {
	base, err := loadScenarioBaseState(gs.ScenarioID, gs.ScenarioPath)
	if err != nil {
		return campaignSaveState{}, err
	}

	savedRegions := make(map[world.RegionID]regionSaveState)
	for _, rid := range gs.RegionOrder {
		region := gs.Regions[rid]
		if region == nil {
			continue
		}
		baseRegion := base.Regions[rid]
		delta, ok := makeRegionSaveState(region, baseRegion)
		if ok {
			savedRegions[rid] = delta
		}
	}
	for rid, region := range gs.Regions {
		if region == nil {
			continue
		}
		if _, seen := savedRegions[rid]; seen {
			continue
		}
		baseRegion := base.Regions[rid]
		delta, ok := makeRegionSaveState(region, baseRegion)
		if ok {
			savedRegions[rid] = delta
		}
	}

	savedFactions := make(map[faction.FactionID]factionSaveState)
	for _, fid := range gs.FactionOrder {
		current := gs.Factions[fid]
		if current == nil {
			continue
		}
		baseFaction := base.Factions[fid]
		delta, ok := makeFactionSaveState(current, baseFaction)
		if ok {
			savedFactions[fid] = delta
		}
	}
	for fid, current := range gs.Factions {
		if current == nil {
			continue
		}
		if _, seen := savedFactions[fid]; seen {
			continue
		}
		baseFaction := base.Factions[fid]
		delta, ok := makeFactionSaveState(current, baseFaction)
		if ok {
			savedFactions[fid] = delta
		}
	}

	return campaignSaveState{
		Turn:                    gs.Turn,
		Year:                    gs.Year,
		Month:                   gs.Month,
		StartYear:               gs.StartYear,
		ScenarioID:              gs.ScenarioID,
		ScenarioPath:            saveScenarioPath(gs.ScenarioID, gs.ScenarioPath),
		PlayerFactionID:         gs.PlayerFactionID,
		Difficulty:              gs.Difficulty,
		DevelopmentMode:         gs.DevelopmentMode,
		EditMode:                gs.EditMode,
		Victory:                 gs.Victory,
		SelectedVictoryOptionID: gs.SelectedVictoryOptionID,
		Regions:                 emptyMapAsNil(savedRegions),
		Factions:                emptyMapAsNil(savedFactions),
		Armies:                  convertArmiesToSaveState(gs.Armies),
		Commanders:              cloneCommanders(gs.Commanders),
		AIPlans:                 cloneAIPlans(gs.AIPlans),
		EconomicVictoryTurns:    gs.EconomicVictoryTurns,
		FactionsEliminated:      gs.FactionsEliminated,
		ReligiousVictoryTurns:   gs.ReligiousVictoryTurns,
		VictoryAchieved:         gs.VictoryAchieved,
		VictoryAchievedTurn:     gs.VictoryAchievedTurn,
		FiredEventIDs:           firedEventIDsToSlice(gs.FiredEventIDs),
		Relations:               makeRelationDelta(gs.Relations, base.Relations),
		DiplomaticOffers:        append([]state.DiplomaticOffer(nil), gs.DiplomaticOffers...),
		DiplomaticOfferHistory:  append([]state.DiplomaticOfferHistoryEntry(nil), gs.DiplomaticOfferHistory...),
		DiplomacyOfferCounts:    cloneFactionIntMap(gs.DiplomacyOfferCounts),
		TradeRoutes:             cloneTradeRoutes(gs.TradeRoutes),
		Sieges:                  cloneSieges(gs.Sieges),
		ProductionQueue:         append([]state.ProductionOrder(nil), gs.ProductionQueue...),
		NextProductionSeq:       gs.NextProductionSeq,
		NextArmySeq:             gs.NextArmySeq,
		NextCommanderSeq:        gs.NextCommanderSeq,
		Phase:                   gs.Phase,
		WinnerID:                gs.WinnerID,
		ActiveRegionEvents:      append([]state.RegionEventStatus(nil), gs.ActiveRegionEvents...),
	}, nil
}

func makeDebugCampaignSaveState(gs *state.GameState) legacyCampaignSaveState {
	if gs == nil {
		return legacyCampaignSaveState{}
	}

	regions := make(map[world.RegionID]legacyRegionSaveState, len(gs.Regions))
	for rid, region := range gs.Regions {
		if region == nil {
			continue
		}
		regions[rid] = legacyRegionSaveState{
			OwnerID:         region.OwnerID,
			Settlements:     append([]world.Settlement(nil), region.Settlements...),
			IsLocked:        region.IsLocked,
			Satisfaction:    region.Satisfaction,
			TaxRate:         region.TaxRate,
			Population:      region.Population,
			Religion:        region.Religion,
			ConversionTurns: region.ConversionTurns,
			ActiveEventID:   region.ActiveEventID,
			Buildings:       append([]string(nil), region.Buildings...),
		}
	}

	factions := make(map[faction.FactionID]legacyFactionSaveState, len(gs.Factions))
	for fid, fx := range gs.Factions {
		if fx == nil {
			continue
		}
		factions[fid] = legacyFactionSaveState{
			IsEliminated:               fx.IsEliminated,
			OverlordID:                 fx.OverlordID,
			CapitalSettlementID:        fx.CapitalSettlementID,
			PendingCapitalSettlementID: fx.PendingCapitalSettlementID,
			PendingCapitalTurns:        fx.PendingCapitalTurns,
			Gold:                       fx.Gold,
			Grain:                      fx.Grain,
			Iron:                       fx.Iron,
			Timber:                     fx.Timber,
			Stone:                      fx.Stone,
			Spice:                      fx.Spice,
			Cloth:                      fx.Cloth,
			Research:                   cloneResearchStateValue(fx.Research),
		}
	}

	return legacyCampaignSaveState{
		Turn:                    gs.Turn,
		Year:                    gs.Year,
		Month:                   gs.Month,
		StartYear:               gs.StartYear,
		ScenarioID:              gs.ScenarioID,
		ScenarioPath:            saveScenarioPath(gs.ScenarioID, gs.ScenarioPath),
		PlayerFactionID:         gs.PlayerFactionID,
		Difficulty:              gs.Difficulty,
		DevelopmentMode:         gs.DevelopmentMode,
		EditMode:                gs.EditMode,
		Victory:                 gs.Victory,
		SelectedVictoryOptionID: gs.SelectedVictoryOptionID,
		Regions:                 regions,
		Factions:                factions,
		Armies:                  cloneArmies(gs.Armies),
		Commanders:              cloneCommanders(gs.Commanders),
		AIPlans:                 cloneAIPlans(gs.AIPlans),
		EconomicVictoryTurns:    gs.EconomicVictoryTurns,
		FactionsEliminated:      gs.FactionsEliminated,
		ReligiousVictoryTurns:   gs.ReligiousVictoryTurns,
		VictoryAchieved:         gs.VictoryAchieved,
		VictoryAchievedTurn:     gs.VictoryAchievedTurn,
		FiredEventIDs:           firedEventIDsFromSlice(firedEventIDsToSlice(gs.FiredEventIDs)),
		Relations:               cloneRelations(gs.Relations),
		DiplomaticOffers:        append([]state.DiplomaticOffer(nil), gs.DiplomaticOffers...),
		DiplomaticOfferHistory:  append([]state.DiplomaticOfferHistoryEntry(nil), gs.DiplomaticOfferHistory...),
		DiplomacyOfferCounts:    cloneFactionIntMap(gs.DiplomacyOfferCounts),
		TradeRoutes:             cloneTradeRoutes(gs.TradeRoutes),
		Sieges:                  cloneSieges(gs.Sieges),
		ProductionQueue:         append([]state.ProductionOrder(nil), gs.ProductionQueue...),
		NextProductionSeq:       gs.NextProductionSeq,
		NextArmySeq:             gs.NextArmySeq,
		NextCommanderSeq:        gs.NextCommanderSeq,
		Phase:                   gs.Phase,
		WinnerID:                gs.WinnerID,
		ActiveRegionEvents:      append([]state.RegionEventStatus(nil), gs.ActiveRegionEvents...),
	}
}

func applyCampaignSaveState(gs *state.GameState, saved campaignSaveState) {
	if saved.Turn > 0 {
		gs.Turn = saved.Turn
	}
	if saved.Year > 0 {
		gs.Year = saved.Year
	}
	if saved.Month > 0 {
		gs.Month = saved.Month
	}
	if saved.StartYear > 0 {
		gs.StartYear = saved.StartYear
	}
	if saved.ScenarioID != "" {
		gs.ScenarioID = saved.ScenarioID
	}
	if saved.ScenarioPath != "" {
		gs.ScenarioPath = saved.ScenarioPath
	}
	if saved.PlayerFactionID != "" {
		gs.PlayerFactionID = saved.PlayerFactionID
	}
	if saved.Difficulty > 0 {
		gs.Difficulty = saved.Difficulty
	}
	gs.DevelopmentMode = saved.DevelopmentMode
	gs.EditMode = saved.EditMode
	gs.Victory = saved.Victory
	gs.SelectedVictoryOptionID = saved.SelectedVictoryOptionID
	gs.EconomicVictoryTurns = saved.EconomicVictoryTurns
	gs.FactionsEliminated = saved.FactionsEliminated
	gs.ReligiousVictoryTurns = saved.ReligiousVictoryTurns
	gs.VictoryAchieved = saved.VictoryAchieved
	gs.VictoryAchievedTurn = saved.VictoryAchievedTurn
	gs.FiredEventIDs = firedEventIDsFromSlice(saved.FiredEventIDs)
	gs.DiplomaticOffers = append([]state.DiplomaticOffer(nil), saved.DiplomaticOffers...)
	gs.DiplomaticOfferHistory = append([]state.DiplomaticOfferHistoryEntry(nil), saved.DiplomaticOfferHistory...)
	gs.DiplomacyOfferCounts = cloneFactionIntMap(saved.DiplomacyOfferCounts)
	gs.TradeRoutes = cloneTradeRoutes(saved.TradeRoutes)
	gs.Sieges = cloneSieges(saved.Sieges)
	gs.ProductionQueue = append([]state.ProductionOrder(nil), saved.ProductionQueue...)
	gs.NextProductionSeq = saved.NextProductionSeq
	gs.NextArmySeq = saved.NextArmySeq
	if saved.Phase != "" {
		gs.Phase = saved.Phase
	}
	gs.WinnerID = saved.WinnerID
	gs.ActiveRegionEvents = append([]state.RegionEventStatus(nil), saved.ActiveRegionEvents...)

	for rid, regionState := range saved.Regions {
		region := gs.Regions[rid]
		if region == nil {
			continue
		}
		applyRegionSaveState(region, regionState)
	}

	for fid, factionState := range saved.Factions {
		fx := gs.Factions[fid]
		if fx == nil {
			continue
		}
		applyFactionSaveState(fx, factionState)
	}

	if saved.Armies != nil {
		gs.Armies = restoreArmiesFromSaveState(saved.Armies)
	} else {
		gs.Armies = map[army.ArmyID]*army.Army{}
	}
	gs.Commanders = cloneCommanders(saved.Commanders)
	gs.AIPlans = cloneAIPlans(saved.AIPlans)
	gs.NextCommanderSeq = saved.NextCommanderSeq
	gs.SyncCommanderLinks()

	applyRelationDelta(gs, saved.Relations)
}

func makeRegionSaveState(current, base *world.Region) (regionSaveState, bool) {
	var out regionSaveState
	if current == nil {
		return out, false
	}
	if base == nil || current.OwnerID != base.OwnerID {
		out.OwnerID = cloneStringPtr(current.OwnerID)
	}
	if patch, ok := diffSettlements(base, current); ok {
		out.Settlements = patch
	}
	if base == nil || current.IsLocked != base.IsLocked {
		out.IsLocked = cloneBoolPtr(current.IsLocked)
	}
	if base == nil || current.Satisfaction != base.Satisfaction {
		out.Satisfaction = cloneIntPtr(current.Satisfaction)
	}
	if base == nil || current.TaxRate != base.TaxRate {
		out.TaxRate = cloneIntPtr(current.TaxRate)
	}
	if base == nil || current.Population != base.Population {
		out.Population = cloneIntPtr(current.Population)
	}
	if base == nil || current.Religion != base.Religion {
		out.Religion = cloneStringPtr(current.Religion)
	}
	if base == nil || current.ConversionTurns != base.ConversionTurns {
		out.ConversionTurns = cloneIntPtr(current.ConversionTurns)
	}
	if base == nil || current.ActiveEventID != base.ActiveEventID {
		out.ActiveEventID = cloneStringPtr(current.ActiveEventID)
	}
	if base == nil || !slices.Equal(current.Buildings, base.Buildings) {
		buildings := stringList(append([]string(nil), current.Buildings...))
		out.Buildings = &buildings
	}
	return out, !isZeroRegionSaveState(out)
}

func applyRegionSaveState(region *world.Region, saved regionSaveState) {
	if region == nil {
		return
	}
	if saved.OwnerID != nil {
		region.OwnerID = *saved.OwnerID
	}
	if saved.IsLocked != nil {
		region.IsLocked = *saved.IsLocked
	}
	if saved.Satisfaction != nil {
		region.Satisfaction = *saved.Satisfaction
	}
	if saved.TaxRate != nil {
		region.TaxRate = *saved.TaxRate
	}
	if saved.Population != nil {
		region.Population = *saved.Population
	}
	if saved.Religion != nil {
		region.Religion = *saved.Religion
	}
	if saved.ConversionTurns != nil {
		region.ConversionTurns = *saved.ConversionTurns
	}
	if saved.ActiveEventID != nil {
		region.ActiveEventID = *saved.ActiveEventID
	}
	if saved.Buildings != nil {
		region.Buildings = append([]string(nil), []string(*saved.Buildings)...)
	}
	if saved.Settlements != nil {
		region.Settlements = applySettlementPatch(region.Settlements, *saved.Settlements)
	}
}

func makeFactionSaveState(current, base *faction.Faction) (factionSaveState, bool) {
	var out factionSaveState
	if current == nil {
		return out, false
	}
	if base == nil || current.IsEliminated != base.IsEliminated {
		out.IsEliminated = cloneBoolPtr(current.IsEliminated)
	}
	if base == nil || current.OverlordID != base.OverlordID {
		out.OverlordID = cloneFactionIDPtr(current.OverlordID)
	}
	if base == nil || current.CapitalSettlementID != base.CapitalSettlementID {
		out.CapitalSettlementID = cloneStringPtr(current.CapitalSettlementID)
	}
	if base == nil || current.PendingCapitalSettlementID != base.PendingCapitalSettlementID {
		out.PendingCapitalSettlementID = cloneStringPtr(current.PendingCapitalSettlementID)
	}
	if base == nil || current.PendingCapitalTurns != base.PendingCapitalTurns {
		out.PendingCapitalTurns = cloneIntPtr(current.PendingCapitalTurns)
	}
	if base == nil || current.Gold != base.Gold {
		out.Gold = cloneIntPtr(current.Gold)
	}
	if base == nil || current.Grain != base.Grain {
		out.Grain = cloneIntPtr(current.Grain)
	}
	if base == nil || current.Iron != base.Iron {
		out.Iron = cloneIntPtr(current.Iron)
	}
	if base == nil || current.Timber != base.Timber {
		out.Timber = cloneIntPtr(current.Timber)
	}
	if base == nil || current.Stone != base.Stone {
		out.Stone = cloneIntPtr(current.Stone)
	}
	if base == nil || current.Spice != base.Spice {
		out.Spice = cloneIntPtr(current.Spice)
	}
	if base == nil || current.Cloth != base.Cloth {
		out.Cloth = cloneIntPtr(current.Cloth)
	}
	if base == nil || !reflect.DeepEqual(current.Research, base.Research) {
		out.Research = cloneResearchStatePtr(current.Research)
	}
	return out, !isZeroFactionSaveState(out)
}

func applyFactionSaveState(fx *faction.Faction, saved factionSaveState) {
	if fx == nil {
		return
	}
	if saved.IsEliminated != nil {
		fx.IsEliminated = *saved.IsEliminated
	}
	if saved.OverlordID != nil {
		fx.OverlordID = *saved.OverlordID
	}
	if saved.CapitalSettlementID != nil {
		fx.CapitalSettlementID = *saved.CapitalSettlementID
	}
	if saved.PendingCapitalSettlementID != nil {
		fx.PendingCapitalSettlementID = *saved.PendingCapitalSettlementID
	}
	if saved.PendingCapitalTurns != nil {
		fx.PendingCapitalTurns = *saved.PendingCapitalTurns
	}
	if saved.Gold != nil {
		fx.Gold = *saved.Gold
	}
	if saved.Grain != nil {
		fx.Grain = *saved.Grain
	}
	if saved.Iron != nil {
		fx.Iron = *saved.Iron
	}
	if saved.Timber != nil {
		fx.Timber = *saved.Timber
	}
	if saved.Stone != nil {
		fx.Stone = *saved.Stone
	}
	if saved.Spice != nil {
		fx.Spice = *saved.Spice
	}
	if saved.Cloth != nil {
		fx.Cloth = *saved.Cloth
	}
	if saved.Research != nil {
		fx.Research = *saved.Research
	}
}

func diffSettlements(baseRegion, currentRegion *world.Region) (*settlementPatch, bool) {
	currentSettlements := []world.Settlement(nil)
	if currentRegion != nil {
		currentSettlements = currentRegion.Settlements
	}
	baseSettlements := []world.Settlement(nil)
	if baseRegion != nil {
		baseSettlements = baseRegion.Settlements
	}
	if reflect.DeepEqual(currentSettlements, baseSettlements) {
		return nil, false
	}
	if !canPatchSettlements(baseSettlements) || !canPatchSettlements(currentSettlements) {
		return &settlementPatch{Replace: append([]world.Settlement(nil), currentSettlements...)}, true
	}

	baseByID := make(map[string]world.Settlement, len(baseSettlements))
	baseOrder := make([]string, 0, len(baseSettlements))
	for _, settlement := range baseSettlements {
		baseByID[settlement.ID] = settlement
		baseOrder = append(baseOrder, settlement.ID)
	}
	currentByID := make(map[string]world.Settlement, len(currentSettlements))
	currentOrder := make([]string, 0, len(currentSettlements))
	added := make([]world.Settlement, 0)
	updated := make([]world.Settlement, 0)
	for _, settlement := range currentSettlements {
		currentByID[settlement.ID] = settlement
		currentOrder = append(currentOrder, settlement.ID)
		baseSettlement, ok := baseByID[settlement.ID]
		if !ok {
			added = append(added, settlement)
			continue
		}
		if !reflect.DeepEqual(settlement, baseSettlement) {
			updated = append(updated, settlement)
		}
	}
	removed := make([]string, 0)
	for _, settlement := range baseSettlements {
		if _, ok := currentByID[settlement.ID]; !ok {
			removed = append(removed, settlement.ID)
		}
	}

	patch := settlementPatch{
		Added:   added,
		Updated: updated,
		Removed: removed,
	}
	if len(added) > 0 || len(removed) > 0 || !slices.Equal(baseOrder, currentOrder) {
		patch.Order = currentOrder
	}
	return &patch, true
}

func canPatchSettlements(list []world.Settlement) bool {
	seen := make(map[string]bool, len(list))
	for _, settlement := range list {
		if settlement.ID == "" || seen[settlement.ID] {
			return false
		}
		seen[settlement.ID] = true
	}
	return true
}

func applySettlementPatch(base []world.Settlement, patch settlementPatch) []world.Settlement {
	if patch.Replace != nil {
		return append([]world.Settlement(nil), patch.Replace...)
	}
	current := make([]world.Settlement, 0, len(base)+len(patch.Added))
	byID := make(map[string]world.Settlement, len(base)+len(patch.Added))
	for _, settlement := range base {
		current = append(current, settlement)
		byID[settlement.ID] = settlement
	}
	if len(patch.Removed) > 0 {
		removed := make(map[string]bool, len(patch.Removed))
		for _, id := range patch.Removed {
			removed[id] = true
			delete(byID, id)
		}
		filtered := current[:0]
		for _, settlement := range current {
			if removed[settlement.ID] {
				continue
			}
			filtered = append(filtered, settlement)
		}
		current = filtered
	}
	if len(patch.Updated) > 0 {
		for _, settlement := range patch.Updated {
			byID[settlement.ID] = settlement
		}
		for i := range current {
			if replacement, ok := byID[current[i].ID]; ok {
				current[i] = replacement
			}
		}
	}
	if len(patch.Added) > 0 {
		for _, settlement := range patch.Added {
			byID[settlement.ID] = settlement
		}
	}
	if len(patch.Order) > 0 {
		out := make([]world.Settlement, 0, len(byID))
		seen := make(map[string]bool, len(patch.Order))
		for _, id := range patch.Order {
			settlement, ok := byID[id]
			if !ok || seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, settlement)
		}
		if len(out) < len(byID) {
			leftoverIDs := make([]string, 0, len(byID)-len(out))
			for id := range byID {
				if seen[id] {
					continue
				}
				leftoverIDs = append(leftoverIDs, id)
			}
			sort.Strings(leftoverIDs)
			for _, id := range leftoverIDs {
				out = append(out, byID[id])
			}
		}
		return out
	}
	if len(patch.Added) > 0 {
		for _, settlement := range patch.Added {
			current = append(current, settlement)
		}
	}
	return current
}

func makeRelationDelta(current, base map[string]*faction.Relation) map[string]relationSaveState {
	if len(current) == 0 && len(base) == 0 {
		return nil
	}
	keys := make(map[string]struct{}, len(current)+len(base))
	for key := range current {
		keys[key] = struct{}{}
	}
	for key := range base {
		keys[key] = struct{}{}
	}
	out := make(map[string]relationSaveState)
	for key := range keys {
		currentRel := current[key]
		baseRel := base[key]
		switch {
		case currentRel == nil && baseRel == nil:
			continue
		case currentRel == nil:
			out[key] = relationSaveState{Deleted: true}
		case baseRel == nil:
			score := currentRel.Score
			stance := encodeStance(currentRel.Stance)
			out[key] = relationSaveState{Score: &score, Stance: &stance}
		default:
			var delta relationSaveState
			if currentRel.Score != baseRel.Score {
				score := currentRel.Score
				delta.Score = &score
			}
			if currentRel.Stance != baseRel.Stance {
				stance := encodeStance(currentRel.Stance)
				delta.Stance = &stance
			}
			if delta.Score != nil || delta.Stance != nil {
				out[key] = delta
			}
		}
	}
	return emptyMapAsNil(out)
}

func applyRelationDelta(gs *state.GameState, deltas map[string]relationSaveState) {
	if gs == nil || len(deltas) == 0 {
		return
	}
	if gs.Relations == nil {
		gs.Relations = make(map[string]*faction.Relation)
	}
	for key, delta := range deltas {
		if delta.Deleted {
			delete(gs.Relations, key)
			continue
		}
		rel := gs.Relations[key]
		if rel == nil {
			a, b := splitRelationKey(key)
			rel = &faction.Relation{FactionA: a, FactionB: b}
			gs.Relations[key] = rel
		}
		if delta.Score != nil {
			rel.Score = *delta.Score
		}
		if delta.Stance != nil {
			rel.Stance = decodeStance(*delta.Stance)
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

func encodeStance(stance faction.DiplomaticStance) uint8 {
	switch stance {
	case faction.StanceWar:
		return 1
	case faction.StanceAllied:
		return 2
	case faction.StanceTrade:
		return 3
	default:
		return 0
	}
}

func decodeStance(code uint8) faction.DiplomaticStance {
	switch code {
	case 1:
		return faction.StanceWar
	case 2:
		return faction.StanceAllied
	case 3:
		return faction.StanceTrade
	default:
		return faction.StancePeace
	}
}

func convertArmiesToSaveState(armies map[army.ArmyID]*army.Army) map[army.ArmyID]armySaveState {
	if len(armies) == 0 {
		return nil
	}
	out := make(map[army.ArmyID]armySaveState, len(armies))
	for id, current := range armies {
		if current == nil {
			continue
		}
		out[id] = armySaveState{
			OwnerID:            current.OwnerID,
			RegionID:           current.RegionID,
			DockedRegionID:     current.DockedRegionID,
			DockedSettlementID: current.DockedSettlementID,
			Units:              stackUnits(current.Units),
			EmbarkedUnits:      stackUnits(current.EmbarkedUnits),
			MovePoints:         current.MovePoints,
			MaxMovePoints:      current.MaxMovePoints,
			IsNaval:            current.IsNaval,
			IsGarrison:         current.IsGarrison,
			Commander:          cloneCommander(current.Commander),
			EmbarkedCommander:  cloneCommander(current.EmbarkedCommander),
			InAmbush:           current.InAmbush,
			OverCapacityTurns:  current.OverCapacityTurns,
			TurnsWithoutPort:   current.TurnsWithoutPort,
		}
	}
	return out
}

func restoreArmiesFromSaveState(saved map[army.ArmyID]armySaveState) map[army.ArmyID]*army.Army {
	if len(saved) == 0 {
		return map[army.ArmyID]*army.Army{}
	}
	out := make(map[army.ArmyID]*army.Army, len(saved))
	for id, current := range saved {
		out[id] = &army.Army{
			ID:                 id,
			OwnerID:            current.OwnerID,
			RegionID:           current.RegionID,
			DockedRegionID:     current.DockedRegionID,
			DockedSettlementID: current.DockedSettlementID,
			Units:              restoreStackedUnits(current.Units),
			EmbarkedUnits:      restoreStackedUnits(current.EmbarkedUnits),
			MovePoints:         current.MovePoints,
			MaxMovePoints:      current.MaxMovePoints,
			IsNaval:            current.IsNaval,
			IsGarrison:         current.IsGarrison,
			Commander:          cloneCommander(current.Commander),
			EmbarkedCommander:  cloneCommander(current.EmbarkedCommander),
			InAmbush:           current.InAmbush,
			OverCapacityTurns:  current.OverCapacityTurns,
			TurnsWithoutPort:   current.TurnsWithoutPort,
		}
	}
	return out
}

func stackUnits(units []army.Unit) []stackedUnitSaveState {
	if len(units) == 0 {
		return nil
	}
	type unitKey struct {
		TypeID     string
		CurrentHP  int
		Experience int
	}
	indexByKey := make(map[unitKey]int, len(units))
	stacked := make([]stackedUnitSaveState, 0, len(units))
	for _, unit := range units {
		key := unitKey{
			TypeID:     unit.TypeID,
			CurrentHP:  unit.CurrentHP,
			Experience: unit.Experience,
		}
		if idx, ok := indexByKey[key]; ok {
			stacked[idx].Count++
			continue
		}
		entry := stackedUnitSaveState{
			TypeID: unit.TypeID,
			Count:  1,
		}
		if unit.CurrentHP != army.MaxUnitHP {
			entry.CurrentHP = unit.CurrentHP
		}
		if unit.Experience != 0 {
			entry.Experience = unit.Experience
		}
		indexByKey[key] = len(stacked)
		stacked = append(stacked, entry)
	}
	return stacked
}

func restoreStackedUnits(stacked []stackedUnitSaveState) []army.Unit {
	if len(stacked) == 0 {
		return nil
	}
	total := 0
	for _, entry := range stacked {
		if entry.Count <= 0 {
			continue
		}
		total += entry.Count
	}
	out := make([]army.Unit, 0, total)
	for _, entry := range stacked {
		if entry.Count <= 0 {
			continue
		}
		hp := entry.CurrentHP
		if hp == 0 {
			hp = army.MaxUnitHP
		}
		for i := 0; i < entry.Count; i++ {
			out = append(out, army.Unit{
				TypeID:     entry.TypeID,
				CurrentHP:  hp,
				Experience: entry.Experience,
			})
		}
	}
	return out
}

func firedEventIDsToSlice(m map[string]bool) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for id, fired := range m {
		if !fired {
			continue
		}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func firedEventIDsFromSlice(list []string) map[string]bool {
	if len(list) == 0 {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(list))
	for _, id := range list {
		if id == "" {
			continue
		}
		out[id] = true
	}
	return out
}

func saveScenarioPath(scenarioID, scenarioPath string) string {
	if scenarioPath == "" {
		return ""
	}
	resolved := resolveScenarioPath(scenarioID, "")
	if resolved != "" && filepathEqual(resolved, scenarioPath) {
		return ""
	}
	return scenarioPath
}

func filepathEqual(a, b string) bool {
	cleanA := strings.TrimRight(a, `/\`)
	cleanB := strings.TrimRight(b, `/\`)
	return cleanA == cleanB
}

func cloneTradeRoutes(routes []*economy.TradeRoute) []*economy.TradeRoute {
	if len(routes) == 0 {
		return nil
	}
	out := make([]*economy.TradeRoute, 0, len(routes))
	for _, route := range routes {
		if route == nil {
			continue
		}
		copyRoute := *route
		out = append(out, &copyRoute)
	}
	return out
}

func cloneSieges(sieges map[world.RegionID]*state.SiegeState) map[world.RegionID]*state.SiegeState {
	if len(sieges) == 0 {
		return nil
	}
	out := make(map[world.RegionID]*state.SiegeState, len(sieges))
	for rid, siege := range sieges {
		if siege == nil {
			continue
		}
		copySiege := *siege
		out[rid] = &copySiege
	}
	return out
}

func cloneStringPtr(v string) *string {
	return &v
}

func cloneBoolPtr(v bool) *bool {
	return &v
}

func cloneIntPtr(v int) *int {
	return &v
}

func cloneFactionIDPtr(v faction.FactionID) *faction.FactionID {
	return &v
}

func cloneResearchStatePtr(v faction.ResearchState) *faction.ResearchState {
	copyState := cloneResearchStateValue(v)
	return &copyState
}

func cloneResearchStateValue(v faction.ResearchState) faction.ResearchState {
	return faction.ResearchState{
		Completed:   cloneStringBoolMap(v.Completed),
		PausedTurns: cloneStringIntMap(v.PausedTurns),
		ActiveID:    v.ActiveID,
		TurnsLeft:   v.TurnsLeft,
	}
}

func cloneStringBoolMap(src map[string]bool) map[string]bool {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]bool, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func cloneStringIntMap(src map[string]int) map[string]int {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]int, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func cloneFactionIntMap(src map[faction.FactionID]int) map[faction.FactionID]int {
	if len(src) == 0 {
		return nil
	}
	out := make(map[faction.FactionID]int, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func cloneRelations(src map[string]*faction.Relation) map[string]*faction.Relation {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]*faction.Relation, len(src))
	for key, rel := range src {
		if rel == nil {
			continue
		}
		copyRel := *rel
		out[key] = &copyRel
	}
	return out
}

func cloneArmies(src map[army.ArmyID]*army.Army) map[army.ArmyID]*army.Army {
	if len(src) == 0 {
		return nil
	}
	out := make(map[army.ArmyID]*army.Army, len(src))
	for id, current := range src {
		if current == nil {
			continue
		}
		copyArmy := *current
		copyArmy.Units = append([]army.Unit(nil), current.Units...)
		copyArmy.EmbarkedUnits = append([]army.Unit(nil), current.EmbarkedUnits...)
		copyArmy.Commander = cloneCommander(current.Commander)
		copyArmy.EmbarkedCommander = cloneCommander(current.EmbarkedCommander)
		out[id] = &copyArmy
	}
	return out
}

func cloneCommander(src *army.Commander) *army.Commander {
	if src == nil {
		return nil
	}
	copyCommander := *src
	copyCommander.Traits = append([]army.CommanderTrait(nil), src.Traits...)
	return &copyCommander
}

func cloneCommanders(src map[string]*army.Commander) map[string]*army.Commander {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]*army.Commander, len(src))
	for id, commander := range src {
		if commander == nil {
			continue
		}
		out[id] = cloneCommander(commander)
	}
	return out
}

func cloneAIPlans(src map[faction.FactionID]*state.AIPlanState) map[faction.FactionID]*state.AIPlanState {
	if len(src) == 0 {
		return nil
	}
	out := make(map[faction.FactionID]*state.AIPlanState, len(src))
	for fid, plan := range src {
		if plan == nil {
			continue
		}
		copyPlan := *plan
		copyPlan.TargetRegionIDs = append([]world.RegionID(nil), plan.TargetRegionIDs...)
		copyPlan.AnnexRegionIDs = append([]world.RegionID(nil), plan.AnnexRegionIDs...)
		out[fid] = &copyPlan
	}
	return out
}

func isZeroRegionSaveState(saved regionSaveState) bool {
	return saved.OwnerID == nil &&
		saved.Settlements == nil &&
		saved.IsLocked == nil &&
		saved.Satisfaction == nil &&
		saved.TaxRate == nil &&
		saved.Population == nil &&
		saved.Religion == nil &&
		saved.ConversionTurns == nil &&
		saved.ActiveEventID == nil &&
		saved.Buildings == nil
}

func isZeroFactionSaveState(saved factionSaveState) bool {
	return saved.IsEliminated == nil &&
		saved.OverlordID == nil &&
		saved.CapitalSettlementID == nil &&
		saved.PendingCapitalSettlementID == nil &&
		saved.PendingCapitalTurns == nil &&
		saved.Gold == nil &&
		saved.Grain == nil &&
		saved.Iron == nil &&
		saved.Timber == nil &&
		saved.Stone == nil &&
		saved.Spice == nil &&
		saved.Cloth == nil &&
		saved.Research == nil
}

func emptyMapAsNil[K comparable, V any](m map[K]V) map[K]V {
	if len(m) == 0 {
		return nil
	}
	return m
}
