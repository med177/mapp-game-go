package render

import (
	"path/filepath"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
)

func TestFactionSelectBackgroundLoadsFromScenarioDirectory(t *testing.T) {
	oldPath := factionSelectBackgroundPath
	oldImage := factionSelectBackground
	t.Cleanup(func() {
		factionSelectBackgroundPath = oldPath
		factionSelectBackground = oldImage
	})

	scenarioPath, err := filepath.Abs(filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("senaryo yolu çözülemedi: %v", err)
	}

	background := factionSelectBackgroundImage(&state.GameState{ScenarioPath: scenarioPath})
	if background == nil {
		t.Fatal("senaryo arka planı yüklenemedi")
	}
	if gotW, gotH := background.Bounds().Dx(), background.Bounds().Dy(); gotW != 1408 || gotH != 768 {
		t.Fatalf("senaryo arka planı boyutu yanlış: got=%dx%d want=1408x768", gotW, gotH)
	}
}

func TestSelectableFactionsPrioritizeHistoricalVictoryGroups(t *testing.T) {
	gs := &state.GameState{
		FactionOrder: []faction.FactionID{"venice", "ottoman", "england", "east_rome"},
		Factions: map[faction.FactionID]*faction.Faction{
			"venice":    {ID: "venice", NameTR: "Venedik", IsPlayable: true},
			"ottoman":   {ID: "ottoman", NameTR: "Osmanlı", IsPlayable: true},
			"england":   {ID: "england", NameTR: "İngiltere", IsPlayable: true},
			"east_rome": {ID: "east_rome", NameTR: "Doğu Roma", IsPlayable: true},
		},
		ScenarioVictories: []scenario.VictoryOptionDef{
			{ID: "general_domination", Type: "domination"},
			{ID: "ottoman_rise", Type: "conquer_city", AllowedFactions: []string{"ottoman"}},
			{ID: "restore_rome", Type: "domination", AllowedFactions: []string{"east_rome"}},
		},
	}

	ordered, historicalCount := selectableFactions(gs)

	if historicalCount != 2 {
		t.Fatalf("2 tarihsel fraksiyon bekleniyordu, got=%d", historicalCount)
	}
	if len(ordered) != 4 {
		t.Fatalf("4 oynanabilir fraksiyon bekleniyordu, got=%d", len(ordered))
	}
	if ordered[0] != "ottoman" || ordered[1] != "east_rome" {
		t.Fatalf("tarihsel blok beklenen sirada degil: %+v", ordered)
	}
	if ordered[2] != "venice" || ordered[3] != "england" {
		t.Fatalf("genel blok beklenen sirada degil: %+v", ordered)
	}
}
