package render

import (
	"testing"

	"mapp-game-go/internal/events"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestNormalizeEditID(t *testing.T) {
	if got := normalizeEditID("  NeW_Region-ID  "); got != "new_region-id" {
		t.Fatalf("normalizeEditID() = %q, want new_region-id", got)
	}
}

func TestAppendFactionFormRuneNormalizesUppercaseID(t *testing.T) {
	r := &Renderer{editFactionForm: editFactionFormState{active: editFactionFieldID}}
	r.appendFactionFormRune('A')

	if got := r.editFactionForm.id; got != "a" {
		t.Fatalf("faction ID after uppercase input = %q, want a", got)
	}
}

func TestRenameRegionIDUpdatesFactionTerritorialClaims(t *testing.T) {
	const fid faction.FactionID = "faction"
	oldID := world.RegionID("old_region")
	newID := world.RegionID("new_region")
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				oldID: {ID: oldID},
			},
			Factions: map[faction.FactionID]*faction.Faction{
				fid: {
					ID: fid,
					TerritorialClaims: []faction.TerritorialClaim{
						{RegionID: string(oldID), Value: 80, Core: true},
						{RegionID: "other_region", Value: 40},
					},
				},
			},
		},
		scenarioEvents: []*events.Event{{
			SuccessorRevival:     &events.SuccessorRevivalEffect{RegionID: string(oldID)},
			RequiresOwnedRegions: []world.RegionID{oldID},
			Choices: []events.Choice{{Effect: events.Effect{
				ArmyDefections: []events.ArmyDefectionEffect{{
					SourceRegionIDs:     []world.RegionID{oldID},
					DestinationRegionID: oldID,
				}},
			}}},
		}},
	}

	r.renameRegionID(oldID, newID)

	claims := r.gs.Factions[fid].TerritorialClaims
	if claims[0].RegionID != string(newID) || claims[0].Value != 80 || !claims[0].Core {
		t.Fatalf("territorial claim was not updated without losing metadata: %+v", claims[0])
	}
	if claims[1].RegionID != "other_region" {
		t.Fatalf("unrelated territorial claim changed: %+v", claims[1])
	}
	event := r.scenarioEvents[0]
	if event.SuccessorRevival.RegionID != string(newID) || event.RequiresOwnedRegions[0] != newID ||
		event.Choices[0].Effect.ArmyDefections[0].SourceRegionIDs[0] != newID ||
		event.Choices[0].Effect.ArmyDefections[0].DestinationRegionID != newID {
		t.Fatalf("event region references were not updated: %+v", event)
	}
}

func TestSaveFactionFormPreservesUneditedFactionData(t *testing.T) {
	const fid faction.FactionID = "papal_states_f"
	periods := []faction.OtherIncomePeriod{{StartYear: 1310, Amount: 153, Description: "Kilise gelirleri"}}
	historicalChanges := []faction.HistoricalChange{{Year: 1500, NameTR: "Papalık"}}
	original := &faction.Faction{
		ID:                  fid,
		Name:                "Papal States",
		NameTR:              "Papalık Devletleri",
		Flag:                "papal.png",
		Religion:            religion.Catholic,
		Color:               [3]uint8{230, 230, 230},
		CapitalSettlementID: "rome",
		HistoricalChanges:   historicalChanges,
		OtherIncomePeriods:  periods,
		Research: faction.ResearchState{Completed: map[string]bool{
			"canon_law": true,
		}},
	}
	r := &Renderer{
		gs:       &state.GameState{Factions: map[faction.FactionID]*faction.Faction{fid: original}},
		worldMap: &WorldMap{},
		editFactionForm: editFactionFormState{
			create:     false,
			originalID: fid,
			id:         string(fid),
			name:       original.Name,
			nameTR:     "Yeni Papalık Adı",
			religion:   original.Religion,
			color:      original.Color,
			gold:       "500",
			grain:      "100",
			iron:       "50",
			timber:     "50",
			spice:      "0",
			cloth:      "0",
			ai:         "50",
		},
	}

	if !r.saveFactionForm() {
		t.Fatal("saveFactionForm() failed")
	}
	saved := r.gs.Factions[fid]
	if saved == nil {
		t.Fatal("faction disappeared after edit")
	}
	if saved.Flag != original.Flag || saved.CapitalSettlementID != original.CapitalSettlementID {
		t.Fatalf("unmodified faction metadata changed: flag=%q capital=%q", saved.Flag, saved.CapitalSettlementID)
	}
	if len(saved.OtherIncomePeriods) != 1 || saved.OtherIncomePeriods[0].Description != "Kilise gelirleri" {
		t.Fatalf("other income periods were not preserved: %+v", saved)
	}
	if len(saved.HistoricalChanges) != 1 || !saved.Research.Completed["canon_law"] {
		t.Fatalf("historical or research data was not preserved: %+v", saved)
	}
}
