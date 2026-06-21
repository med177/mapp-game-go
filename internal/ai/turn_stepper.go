package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

type TurnStepKind string

const (
	TurnStepInfo      TurnStepKind = "info"
	TurnStepDiplomacy TurnStepKind = "diplomacy"
	TurnStepResearch  TurnStepKind = "research"
	TurnStepBuild     TurnStepKind = "build"
	TurnStepRecruit   TurnStepKind = "recruit"
	TurnStepMove      TurnStepKind = "move"
	TurnStepBattle    TurnStepKind = "battle"
	TurnStepEmbark    TurnStepKind = "embark"
	TurnStepDisembark TurnStepKind = "disembark"
	TurnStepConquest  TurnStepKind = "conquest"
	TurnStepSortie    TurnStepKind = "sortie"
	TurnStepComplete  TurnStepKind = "complete"
)

type TurnStep struct {
	FactionID     faction.FactionID
	Kind          TurnStepKind
	ArmyID        army.ArmyID
	FromRegion    world.RegionID
	TargetRegion  world.RegionID
	FocusRegion   world.RegionID
	TargetFaction faction.FactionID
	Message       string
}

type TurnStepper struct {
	gs          *state.GameState
	fid         faction.FactionID
	prelude     []TurnStep
	preludeIdx  int
	preludeDone bool
	armyOrder   []army.ArmyID
	armyIdx     int
}

func NewTurnStepper(gs *state.GameState, fid faction.FactionID) *TurnStepper {
	return &TurnStepper{
		gs:      gs,
		fid:     fid,
		prelude: make([]TurnStep, 0, 8),
	}
}

func (s *TurnStepper) FactionID() faction.FactionID {
	if s == nil {
		return ""
	}
	return s.fid
}

func (s *TurnStepper) FactionNameTR() string {
	if s == nil {
		return ""
	}
	return turnFactionName(s.gs, s.fid)
}

func (s *TurnStepper) Step() (TurnStep, bool) {
	if s == nil || s.gs == nil || s.fid == "" {
		return TurnStep{}, true
	}
	for {
		if !s.preludeDone {
			runTurnPrelude(s.gs, s.fid, &s.prelude)
			s.preludeDone = true
			if len(s.prelude) > 0 {
				step := s.prelude[0]
				s.preludeIdx = 1
				return step, false
			}
		}
		if s.preludeIdx < len(s.prelude) {
			step := s.prelude[s.preludeIdx]
			s.preludeIdx++
			return step, false
		}
		if s.armyOrder == nil {
			s.initArmyOrder()
		}
		for s.armyIdx < len(s.armyOrder) {
			aid := s.armyOrder[s.armyIdx]
			a := s.gs.Armies[aid]
			if a == nil || a.OwnerID != string(s.fid) {
				s.armyIdx++
				continue
			}
			if a.MovePoints <= 0 {
				s.armyIdx++
				continue
			}
			target := chooseBestMove(s.gs, a)
			if target == "" {
				s.armyIdx++
				continue
			}
			outcome := executeMove(s.gs, a, target, s.fid)
			updated := s.gs.Armies[aid]
			if !outcome.survived || updated == nil || updated.MovePoints <= 0 {
				s.armyIdx++
			}
			if outcome.step.Message != "" {
				return outcome.step, false
			}
		}
		return TurnStep{
			FactionID: s.fid,
			Kind:      TurnStepComplete,
		}, true
	}
}

func (s *TurnStepper) initArmyOrder() {
	ids := make([]army.ArmyID, 0, len(s.gs.Armies))
	for id, candidate := range s.gs.Armies {
		if candidate == nil || candidate.OwnerID != string(s.fid) {
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	s.armyOrder = ids
}
