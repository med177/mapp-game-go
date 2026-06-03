package events

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

type Effect struct {
	Target           string  `json:"target,omitempty"` // boşsa event target'ı kullanılır
	SatDelta         int     `json:"sat_delta,omitempty"`
	GoldDelta        int     `json:"gold_delta,omitempty"`
	GrainDelta       int     `json:"grain_delta,omitempty"`
	ArmyHPMod        float64 `json:"army_hp_mod,omitempty"`        // 1.0 = değişmez
	AffectedFaction  string  `json:"affected_faction,omitempty"`   // specific_faction için
	RelationDeltaAll int     `json:"relation_delta_all,omitempty"` // etkilenen fraksiyonun tüm ilişkilerine uygulanır
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
	ID          string  `json:"id"`
	NameTR      string  `json:"name_tr"`
	DescTR      string  `json:"desc_tr"`
	Probability float64 `json:"probability"` // 0 = sadece tarihsel tetiklenme
	MinTurn     int     `json:"min_turn"`    // en erken tur (rastgele olaylar için)
	Target      string  `json:"target"`      // "player_faction"|"random_region"|"all_armies"|"all_factions"
	SatDelta    int     `json:"sat_delta"`
	GoldDelta   int     `json:"gold_delta"`
	GrainDelta  int     `json:"grain_delta"`
	ArmyHPMod   float64 `json:"army_hp_mod"` // 1.0 = değişmez

	// Tarihsel tetiklenme alanları
	HistoricalYear  int    `json:"historical_year,omitempty"`  // 0 = tarihsel değil
	HistoricalMonth int    `json:"historical_month,omitempty"` // 0 = yılın herhangi bir ayı
	OneShot         bool   `json:"one_shot,omitempty"`         // true = yalnızca bir kez tetiklenir
	AffectedFaction string `json:"affected_faction,omitempty"` // belirli fraksiyonu hedefle

	ChoicePromptTR string   `json:"choice_prompt_tr,omitempty"`
	Choices        []Choice `json:"choices,omitempty"`
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

	// Önce tarihsel olayları kontrol et (kesinlikle tetiklenir)
	for _, e := range evts {
		if e.HistoricalYear == 0 {
			continue
		}
		if gs.FiredEventIDs[e.ID] {
			continue
		}
		if gs.Year != e.HistoricalYear {
			continue
		}
		if e.HistoricalMonth != 0 && gs.Month != e.HistoricalMonth {
			continue
		}
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

func (e *Event) BaseEffect() Effect {
	return Effect{
		Target:           e.Target,
		SatDelta:         e.SatDelta,
		GoldDelta:        e.GoldDelta,
		GrainDelta:       e.GrainDelta,
		ArmyHPMod:        e.ArmyHPMod,
		AffectedFaction:  e.AffectedFaction,
		RelationDeltaAll: 0,
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
		return e.AffectedFaction == string(gs.PlayerFactionID)
	default:
		return false
	}
}

func Apply(gs *state.GameState, e *Event) {
	if gs == nil || e == nil {
		return
	}
	applyEffect(gs, e.BaseEffect())
}

func ApplyChoice(gs *state.GameState, e *Event, idx int) (Choice, bool) {
	if e == nil || idx < 0 || idx >= len(e.Choices) {
		return Choice{}, false
	}
	choice := e.Choices[idx]
	applyEffect(gs, choiceEffect(e, choice))
	return choice, true
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

func applyEffect(gs *state.GameState, eff Effect) {
	target := eff.Target
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
			return
		}
		rid := candidates[rand.Intn(len(candidates))]
		r := gs.Regions[rid]
		r.Satisfaction = clamp(r.Satisfaction+eff.SatDelta, 0, 100)
		if eff.GrainDelta != 0 {
			if f, ok := gs.Factions[faction.FactionID(r.OwnerID)]; ok {
				f.Grain = max0(f.Grain + eff.GrainDelta)
			}
		}

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
}

// applyToFaction bir fraksiyonun tüm bölgelerine ve hazinesine olay etkilerini uygular.
func applyToFaction(gs *state.GameState, fid string, eff Effect) {
	for _, r := range gs.Regions {
		if r.IsSea || r.OwnerID != fid {
			continue
		}
		r.Satisfaction = clamp(r.Satisfaction+eff.SatDelta, 0, 100)
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
