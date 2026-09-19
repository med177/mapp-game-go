package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestCommanderPanelRecruitButtonHitUsesDrawnButtonGeometry(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "venice",
		Factions: map[faction.FactionID]*faction.Faction{
			"venice": {ID: "venice", Gold: state.CommanderRecruitCost.Gold, Grain: state.CommanderRecruitCost.Grain},
		},
	}

	button := commanderPanelRecruitButton(gs)
	if !commanderPanelRecruitButtonHit(gs, button.X+button.W/2, button.Y+button.H/2) {
		t.Fatal("Yeni Komutan düğmesinin çizilen alanı tıklanabilir değil")
	}
	if commanderPanelRecruitButtonHit(gs, button.X-1, button.Y+button.H/2) {
		t.Fatal("Yeni Komutan düğmesinin dışı tıklanabilir kabul edildi")
	}
}

func TestCommanderPanelRecruitButtonHitRejectsUnaffordableRecruitment(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "venice",
		Factions: map[faction.FactionID]*faction.Faction{
			"venice": {ID: "venice"},
		},
	}

	button := commanderPanelRecruitButton(gs)
	if button.Enabled {
		t.Fatal("kaynak yetersizken Yeni Komutan düğmesi etkin kaldı")
	}
	if commanderPanelRecruitButtonHit(gs, button.X+button.W/2, button.Y+button.H/2) {
		t.Fatal("kaynak yetersizken Yeni Komutan düğmesi tıklanabilir kabul edildi")
	}
}
