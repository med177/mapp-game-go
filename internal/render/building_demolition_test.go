package render

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestBuildingDemolishButtonIsAtCardTopRightAndHitTestable(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Factions:        map[faction.FactionID]*faction.Faction{"p1": {ID: "p1"}},
		Regions:         map[world.RegionID]*world.Region{"r1": {ID: "r1", OwnerID: "p1", Buildings: []string{"market"}}},
		BuildingTypes:   map[string]*city.Building{"market": {ID: "market", NameTR: "Pazar", MaxPerRegion: 2}},
	}
	cards := buildBuildingCardComponents(gs, gs.Regions["r1"], 100, 200, 300)
	if len(cards) != 1 || !cards[0].CanDemolish {
		t.Fatalf("tamamlanmış bina için yıkım düğmesi görünür olmalıydı: %+v", cards)
	}
	btn := cards[0].DemolishBtn
	if btn.Icon != gameui.IconX {
		t.Fatalf("yıkım düğmesi X ikonu kullanmalıydı, got=%q", btn.Icon)
	}
	if btn.X+btn.W > cards[0].Rect.X+cards[0].Rect.W || btn.Y < cards[0].Rect.Y {
		t.Fatalf("yıkım düğmesi kartın üst sağında olmalıydı: card=%+v button=%+v", cards[0].Rect, btn)
	}
	if !btn.HitTest(btn.X+btn.W/2, btn.Y+btn.H/2) {
		t.Fatal("yıkım düğmesi kendi merkezinde hit-test edilmeliydi")
	}
}

func TestBuildingDemolitionOpensConfirmationDialog(t *testing.T) {
	r := &Renderer{gs: &state.GameState{
		PlayerFactionID: "p1",
		Regions:         map[world.RegionID]*world.Region{"r1": {ID: "r1", OwnerID: "p1", Buildings: []string{"market"}}},
		BuildingTypes:   map[string]*city.Building{"market": {ID: "market", NameTR: "Pazar", MaxPerRegion: 1}},
	}}

	r.openBuildingDemolitionConfirm("r1", "market")

	if !r.confirmDialog.show {
		t.Fatal("yıkım tıklaması onay modalı açmalıydı")
	}
	if r.confirmDialog.pendingAction.Kind != ActionDemolishBuilding {
		t.Fatalf("modal yanlış aksiyon taşıyor: %+v", r.confirmDialog.pendingAction)
	}
	if r.confirmDialog.pendingAction.TargetRegion != "r1" || r.confirmDialog.pendingAction.BuildingID != "market" {
		t.Fatalf("modal hedefi yanlış: %+v", r.confirmDialog.pendingAction)
	}
}
