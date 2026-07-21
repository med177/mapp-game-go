package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestBuildFactionDiplomacySummaryGroupsRelations(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Sunni, OverlordID: "lord"},
			"lord":   {ID: "lord", NameTR: "Üst Devlet", Religion: religion.Sunni},
			"ally":   {ID: "ally", NameTR: "Müttefik", Religion: religion.Catholic},
			"trade":  {ID: "trade", NameTR: "Tüccar", Religion: religion.Catholic},
			"enemy":  {ID: "enemy", NameTR: "Düşman", Religion: religion.Catholic},
			"vassal": {ID: "vassal", NameTR: "Vassal", Religion: religion.Sunni, OverlordID: "player"},
			"sib":    {ID: "sib", NameTR: "Aynı Realm", Religion: religion.Sunni, OverlordID: "lord"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "lord"):   {FactionA: "player", FactionB: "lord", Stance: faction.StanceAllied, Score: 80},
			faction.RelationKey("player", "ally"):   {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied, Score: 55},
			faction.RelationKey("player", "trade"):  {FactionA: "player", FactionB: "trade", Stance: faction.StanceTrade, Score: 32},
			faction.RelationKey("player", "enemy"):  {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar, Score: -68},
			faction.RelationKey("player", "sib"):    {FactionA: "player", FactionB: "sib", Stance: faction.StanceAllied, Score: 45},
			faction.RelationKey("player", "vassal"): {FactionA: "player", FactionB: "vassal", Stance: faction.StanceAllied, Score: 65},
		},
		PlayerFactionID: "player",
	}

	summary := buildFactionDiplomacySummary(gs, "player")

	if !summary.HasOverlord {
		t.Fatalf("üst devlet görünmeliydi")
	}
	if summary.Overlord.ID != "lord" {
		t.Fatalf("üst devlet yanlış: got=%q want=%q", summary.Overlord.ID, "lord")
	}
	if summary.VassalCount != 1 || len(summary.Vassals) != 1 || summary.Vassals[0].ID != "vassal" {
		t.Fatalf("vassal listesi yanlış: %+v", summary.Vassals)
	}
	if summary.AllianceCount != 1 || len(summary.Allies) != 1 || summary.Allies[0].ID != "ally" {
		t.Fatalf("ittifak listesi yanlış: %+v", summary.Allies)
	}
	if summary.TradeCount != 1 || len(summary.Trade) != 1 || summary.Trade[0].ID != "trade" {
		t.Fatalf("ticaret listesi yanlış: %+v", summary.Trade)
	}
	if summary.EnemyCount != 1 || len(summary.Enemies) != 1 || summary.Enemies[0].ID != "enemy" {
		t.Fatalf("düşman listesi yanlış: %+v", summary.Enemies)
	}
	for _, entry := range append(append(append(summary.Allies, summary.Trade...), summary.Enemies...), summary.Vassals...) {
		if entry.ID == "lord" || entry.ID == "sib" {
			t.Fatalf("same-realm entry dış listelere girmemeliydi: %+v", entry)
		}
	}
}

func TestClampFactionPanelScroll(t *testing.T) {
	if got := clampFactionPanelScroll(100, 200, 50); got != 0 {
		t.Fatalf("viewport content'tan büyükken scroll sifir olmali, got=%v", got)
	}
	if got := clampFactionPanelScroll(1000, 592, 999); got != 408 {
		t.Fatalf("scroll üst sınırına clamp edilmeliydi, got=%v want=%v", got, 408)
	}
	if got := clampFactionPanelScroll(1000, 592, -10); got != 0 {
		t.Fatalf("negatif scroll sifirlanmali, got=%v", got)
	}
}

func TestSyncFactionPanelToSelectedRegionFollowsOwner(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"region_b": {ID: "region_b", OwnerID: "owner_b"},
			},
		},
		SelectedRegion:       "region_b",
		selectedFactionPanel: "owner_a",
		factionPanelScroll:   140,
	}

	r.syncFactionPanelToSelectedRegion()

	if r.selectedFactionPanel != "owner_b" {
		t.Fatalf("panel yeni bölgenin sahibine geçmeliydi: got=%q want=%q", r.selectedFactionPanel, "owner_b")
	}
	if r.factionPanelScroll != 0 {
		t.Fatalf("farklı devlet seçilince panel scroll'u sıfırlanmalıydı: got=%v", r.factionPanelScroll)
	}
}

func TestSyncFactionPanelToSelectedRegionKeepsPanelForSameOwner(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"region_b": {ID: "region_b", OwnerID: "owner_a"},
			},
		},
		SelectedRegion:       "region_b",
		selectedFactionPanel: "owner_a",
		factionPanelScroll:   140,
	}

	r.syncFactionPanelToSelectedRegion()

	if r.selectedFactionPanel != "owner_a" || r.factionPanelScroll != 140 {
		t.Fatalf("aynı devlet seçilince panel korunmalıydı: faction=%q scroll=%v", r.selectedFactionPanel, r.factionPanelScroll)
	}
}
