package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestVictoryDetailContentBecomesScrollableForLongChecklist(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	required := make([]world.RegionID, 0, 18)
	regions := make(map[world.RegionID]*world.Region, 18)
	for i := 0; i < 18; i++ {
		id := world.RegionID("region_" + itoa(i))
		required = append(required, id)
		regions[id] = &world.Region{
			ID:      id,
			NameTR:  "Hedef Bölge " + itoa(i),
			OwnerID: "player",
		}
	}

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Osmanli"},
		},
		Regions: regions,
		Victory: state.VictoryCondition{
			Type:              state.VictoryDomination,
			TargetRegionCount: 32,
			RequiredRegions:   required,
			DeadlineYear:      1481,
			DeadlineMonth:     5,
		},
		SelectedVictoryOptionID: "long_goal",
		ScenarioVictories: []scenario.VictoryOptionDef{
			{
				ID:                "long_goal",
				Type:              "domination",
				Title:             "Fatih'in Yolu",
				Description:       "Konstantiniyye'den Balkanlara uzanan uzun hedef zinciri.",
				Detail:            strings.Repeat("Bu hedef paneli kaydirilabilir olmali. ", 8),
				TargetRegionCount: 32,
				RequiredRegions: func() []string {
					out := make([]string, 0, len(required))
					for _, rid := range required {
						out = append(out, string(rid))
					}
					return out
				}(),
				DeadlineYear:  1481,
				DeadlineMonth: 5,
			},
		},
	}

	content := buildVictoryDetailContentLines(gs)
	layout := buildVictoryDetailLayout()
	if content.contentHeight <= layout.scrollRect.H {
		t.Fatalf("uzun victory detayinda scroll gerekirdi, content=%.1f viewport=%.1f", content.contentHeight, layout.scrollRect.H)
	}
	if max := victoryDetailMaxScroll(gs); max <= 0 {
		t.Fatalf("max scroll pozitif olmali, got=%.1f", max)
	}
	if got := clampVictoryDetailScroll(gs, 9999); got != victoryDetailMaxScroll(gs) {
		t.Fatalf("scroll clamp max'a sabitlenmeli, got=%.1f max=%.1f", got, victoryDetailMaxScroll(gs))
	}
}
