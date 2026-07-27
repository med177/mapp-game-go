package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestImperialPanelAvailableOnlyForHREPlayer(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "hre",
		Factions:        map[faction.FactionID]*faction.Faction{"hre": {ID: "hre"}},
		Imperial:        &state.ImperialState{EmpireID: "hre"},
	}
	if !imperialPanelAvailable(gs) {
		t.Fatal("HRE oyuncusu için imparatorluk paneli görünür olmalı")
	}
	gs.PlayerFactionID = "milan"
	if imperialPanelAvailable(gs) {
		t.Fatal("HRE üyesi veya normal devlet için oyuncu paneli görünmemeli")
	}
}

func TestImperialPanelUsesSharedBackIconButton(t *testing.T) {
	button := imperialPanelCloseButton()
	if button.Label != "" || button.Icon != gameui.IconBack {
		t.Fatalf("HRE paneli ortak geri ikonlu düğmeyi kullanmalı: %+v", button)
	}
	if button.IconSize != 14 {
		t.Fatalf("geri ikonu panel başlığı için doğru boyutta olmalı: got=%v", button.IconSize)
	}
}

func TestImperialPanelLayoutDrawsAtCampaignResolutions(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "hre",
		Factions: map[faction.FactionID]*faction.Faction{
			"hre":   {ID: "hre", NameTR: "Kutsal Roma İmparatorluğu"},
			"milan": {ID: "milan", NameTR: "Milano"},
		},
		Imperial: &state.ImperialState{
			EmpireID:  "hre",
			EmperorID: "hre",
			Members: map[faction.FactionID]*state.ImperialMember{
				"milan": {FactionID: "milan", Status: state.ImperialMemberPrince, Loyalty: 60, Autonomy: 40, MilitaryCommitment: 50},
			},
		},
	}
	r := &Renderer{gs: gs}
	for _, size := range [][2]float64{{1280, 720}, {1024, 768}, {800, 600}} {
		ScreenWidth, ScreenHeight = size[0], size[1]
		screen := ebiten.NewImage(int(size[0]), int(size[1]))
		r.DrawImperialPanel(screen)
		panel := imperialPanelRect()
		if panel.X < 0 || panel.Y < 0 || panel.X+panel.W > ScreenWidth+0.1 || panel.Y+panel.H > ScreenHeight+0.1 {
			t.Fatalf("imparatorluk paneli ekrana sığmıyor: size=%v panel=%+v", size, panel)
		}
		_, viewport, footer, visible := imperialMemberListLayout(panel)
		if viewport.Y+float64(visible)*imperialPanelMemberRowH > footer.Y {
			t.Fatalf("üye satırları alt bilgi alanına taşıyor: size=%v viewport=%+v footer=%+v visible=%d", size, viewport, footer, visible)
		}
	}

	for _, width := range []float64{1280, 800, 640} {
		ScreenWidth, ScreenHeight = width, 720
		button := imperialHUDButtonRect()
		if button[0] < 0 || button[0]+button[2] > float32(ScreenWidth) {
			t.Fatalf("HRE HUD düğmesi ekrana sığmıyor: width=%v rect=%v", width, button)
		}
	}
}

func TestImperialDietOptionDescriptionsDoNotOverlapButtons(t *testing.T) {
	box := gameui.Rect{X: 600, Y: 100, W: 360, H: 420}
	buttons := imperialDecisionButtonRects(box)
	for i := 1; i < len(buttons); i++ {
		previousDescriptionBottom := buttons[i-1].Y + imperialDecisionDescY + 16
		if previousDescriptionBottom > buttons[i].Y {
			t.Fatalf("Diyet açıklaması sonraki düğmeye taşıyor: previous=%+v next=%+v", buttons[i-1], buttons[i])
		}
	}
	last := buttons[len(buttons)-1]
	if last.Y+imperialDecisionDescY+16 > box.Y+box.H {
		t.Fatalf("son Diyet açıklaması karar alanından taşıyor: button=%+v box=%+v", last, box)
	}
}
