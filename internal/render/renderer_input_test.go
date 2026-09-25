package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestInfoPopupClickDismissesOnlyInsidePopup(t *testing.T) {
	popup := gameui.Rect{X: 100, Y: 100, W: 200, H: 80}

	tests := []struct {
		name  string
		timer int
		mx    float64
		my    float64
		want  bool
	}{
		{name: "popup ici", timer: 10, mx: 150, my: 120, want: true},
		{name: "harita disi", timer: 10, mx: 50, my: 120, want: false},
		{name: "sure bitmis", timer: 0, mx: 150, my: 120, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := infoPopupClickDismisses(tt.timer, popup, tt.mx, tt.my, true, false)
			if got != tt.want {
				t.Fatalf("popup kapatma = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfoPopupClickDismissesRequiresNewClick(t *testing.T) {
	popup := gameui.Rect{X: 100, Y: 100, W: 200, H: 80}
	if infoPopupClickDismisses(10, popup, 150, 120, true, true) {
		t.Fatal("basili tutulmus eski tiklama popup'i kapatti")
	}
}

func TestMapRegionDoubleClickUsesTerrainParentIdentity(t *testing.T) {
	parentID := world.RegionID("parent")
	fragmentID := world.RegionID("area::forest::parent")
	r := &Renderer{gs: &state.GameState{Regions: map[world.RegionID]*world.Region{
		parentID:   {ID: parentID, OwnerID: "target"},
		fragmentID: {ID: fragmentID, IsTerrainArea: true, ParentRegionID: parentID},
	}}}

	if r.mapRegionDoubleClicked(fragmentID) {
		t.Fatal("first terrain click unexpectedly counted as double click")
	}
	if !r.mapRegionDoubleClicked(parentID) {
		t.Fatal("terrain fragment and its parent were not treated as the same map click target")
	}
}

func TestToggleArmyDetailPanel(t *testing.T) {
	r := &Renderer{}

	r.toggleArmyDetailPanel()
	if !r.showArmyDetailPanel || !r.armyDetailPanelPreference {
		t.Fatal("marker sağ tıklaması ordu detay panelini açmadı")
	}

	r.toggleArmyDetailPanel()
	if r.showArmyDetailPanel || r.armyDetailPanelPreference {
		t.Fatal("aynı marker'a ikinci sağ tıklama açık ordu detay panelini kapatmadı")
	}
}

func TestDisbandButtonUsesTheSameGeometryForDrawAndHitTest(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", Units: []army.Unit{{TypeID: "infantry"}}},
		},
	}
	selected := map[int]bool{0: true}

	button, ok := buildDisbandArmyButton(gs, "army", selected)
	if !ok {
		t.Fatal("seçili birim varken Sil düğmesi oluşturulamadı")
	}
	mx := button.X + button.W/2
	my := button.Y + button.H/2
	if !DisbandButtonHitTest(mx, my, gs, "army", selected) {
		t.Fatal("Sil düğmesinin merkezi hit-test tarafından yakalanmadı")
	}
	if !ArmyPanelInteractiveHit(mx, my, gs, "army", selected) {
		t.Fatal("Sil düğmesinin merkezi panel cursor/input hit-test'i tarafından yakalanmadı")
	}
}

func TestBottomArmyActionUsesContextualSelection(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Armies: map[army.ArmyID]*army.Army{
			"land":  {ID: "land", OwnerID: "player"},
			"naval": {ID: "naval", OwnerID: "player", IsNaval: true},
		},
	}

	tests := []struct {
		name                               string
		selectedArmy                       army.ArmyID
		showArmyDetail                     bool
		showRecruit                        bool
		recruitEnabled                     bool
		wantLabel                          string
		wantEnabled, wantActive, wantNaval bool
	}{
		{name: "bölge", showRecruit: true, recruitEnabled: true, wantLabel: "Kışla", wantEnabled: true, wantActive: true},
		{name: "kara ordusu", selectedArmy: "land", wantLabel: "Ordu", wantEnabled: true},
		{name: "donanma", selectedArmy: "naval", showArmyDetail: true, wantLabel: "Donanma", wantEnabled: true, wantActive: true, wantNaval: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, enabled, active, naval, _ := bottomArmyAction(gs, tt.selectedArmy, tt.showArmyDetail, tt.showRecruit, tt.recruitEnabled)
			if label != tt.wantLabel || enabled != tt.wantEnabled || active != tt.wantActive || naval != tt.wantNaval {
				t.Fatalf("alt HUD aksiyonu = (%q, %v, %v, %v), want (%q, %v, %v, %v)", label, enabled, active, naval, tt.wantLabel, tt.wantEnabled, tt.wantActive, tt.wantNaval)
			}
			buttons := buildBottomActionButtons(label, enabled)
			if buttons[0].Label != tt.wantLabel || buttons[0].Enabled != tt.wantEnabled {
				t.Fatalf("alt HUD düğmesi = (%q, %v), want (%q, %v)", buttons[0].Label, buttons[0].Enabled, tt.wantLabel, tt.wantEnabled)
			}
			wantLabels := []string{tt.wantLabel, "Teknoloji", "Pazar", "Diplomasi", "Tur Bitir ►"}
			for i, want := range wantLabels {
				if buttons[i].Label != want {
					t.Fatalf("alt HUD sıra[%d] = %q, want %q", i, buttons[i].Label, want)
				}
			}
		})
	}
}

func TestRecruitPanelStateForRegionFollowsPreference(t *testing.T) {
	ownedRegion := world.RegionID("owned")
	enemyRegion := world.RegionID("enemy")
	r := &Renderer{
		gs: &state.GameState{
			PlayerFactionID: "player",
			Regions: map[world.RegionID]*world.Region{
				ownedRegion: {ID: ownedRegion, OwnerID: "player"},
				enemyRegion: {ID: enemyRegion, OwnerID: "enemy"},
			},
		},
		recruitPanelPreference: true,
	}

	if !r.recruitPanelStateForRegion(ownedRegion) {
		t.Fatal("açık Kışla tercihi oyuncunun uygun bölgesine taşınmadı")
	}
	if r.recruitPanelStateForRegion(enemyRegion) {
		t.Fatal("Kışla paneli düşman bölgesinde açık kabul edildi")
	}

	r.recruitPanelPreference = false
	if r.recruitPanelStateForRegion(ownedRegion) {
		t.Fatal("Kışla tercihi kapatıldıktan sonra panel açık kaldı")
	}
}
