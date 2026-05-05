package events

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

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

// Tick her tur sonunda olayları kontrol eder.
// Tetiklenen olayların adı ve açıklaması döner; hiç tetiklenmezse ("","").
func Tick(gs *state.GameState, evts []*Event) (name, desc string) {
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
		apply(gs, e)
		if e.OneShot {
			gs.FiredEventIDs[e.ID] = true
		}
		return e.NameTR, e.DescTR
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
		apply(gs, e)
		if e.OneShot {
			gs.FiredEventIDs[e.ID] = true
		}
		return e.NameTR, e.DescTR
	}
	return "", ""
}

func apply(gs *state.GameState, e *Event) {
	switch e.Target {

	case "player_faction":
		applyToFaction(gs, string(gs.PlayerFactionID), e)

	case "all_factions":
		for fid := range gs.Factions {
			applyToFaction(gs, string(fid), e)
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
		r.Satisfaction = clamp(r.Satisfaction+e.SatDelta, 0, 100)
		if e.GrainDelta != 0 {
			if f, ok := gs.Factions[faction.FactionID(r.OwnerID)]; ok {
				f.Grain = max0(f.Grain + e.GrainDelta)
			}
		}

	case "all_armies":
		if e.ArmyHPMod > 0 && e.ArmyHPMod < 1.0 {
			for _, a := range gs.Armies {
				for i := range a.Units {
					a.Units[i].CurrentHP = max0(int(float64(a.Units[i].CurrentHP) * e.ArmyHPMod))
				}
			}
		}
		for _, f := range gs.Factions {
			f.Grain = max0(f.Grain + e.GrainDelta)
		}

	case "specific_faction":
		if e.AffectedFaction != "" {
			applyToFaction(gs, e.AffectedFaction, e)
		}
	}
}

// applyToFaction bir fraksiyonun tüm bölgelerine ve hazinesine olay etkilerini uygular.
func applyToFaction(gs *state.GameState, fid string, e *Event) {
	for _, r := range gs.Regions {
		if r.IsSea || r.OwnerID != fid {
			continue
		}
		r.Satisfaction = clamp(r.Satisfaction+e.SatDelta, 0, 100)
	}
	if f, ok := gs.Factions[faction.FactionID(fid)]; ok {
		f.Gold = max0(f.Gold + e.GoldDelta)
		f.Grain = max0(f.Grain + e.GrainDelta)
	}
	if e.ArmyHPMod > 0 && e.ArmyHPMod < 1.0 {
		for _, a := range gs.Armies {
			if a.OwnerID != fid {
				continue
			}
			for i := range a.Units {
				a.Units[i].CurrentHP = max0(int(float64(a.Units[i].CurrentHP) * e.ArmyHPMod))
			}
		}
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
