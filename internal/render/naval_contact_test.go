package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestNavalContactWithdrawButtonCanBeDisabled(t *testing.T) {
	r := &Renderer{}
	r.ShowThreeChoiceDialogWithThirdEnabled(
		"Düşman Filo Tespit Edildi",
		"Temas",
		"Çatış",
		"Geri Çekil",
		"Pozisyonu Koru",
		InputAction{Kind: ActionResolveNavalContact, ChoiceIndex: 0},
		InputAction{Kind: ActionResolveNavalContact, ChoiceIndex: 1},
		InputAction{Kind: ActionResolveNavalContact, ChoiceIndex: 2},
		false,
	)
	_, third, _, hasThird := buildConfirmDialogButtons(r.confirmDialog)
	if !hasThird || third.Enabled {
		t.Fatal("hareket puanı olmayan filo için geri çekil seçeneği pasif olmalı")
	}
}

func TestNavalContactDialogShowsFleetComparisonAndUpperPlacement(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() { ScreenWidth, ScreenHeight = oldW, oldH }()
	ScreenWidth, ScreenHeight = 1280, 720
	r := &Renderer{gs: &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"sea": {ID: "sea", NameTR: "Ege Denizi", IsSea: true},
		},
		UnitTypes: map[string]*army.UnitType{},
	}}
	r.ShowNavalContactDialog("player-fleet", "enemy-fleet", "sea", state.NavalContactClash, false)

	if !r.confirmDialog.show || r.confirmDialog.navalContact == nil {
		t.Fatal("temas modalı özel filo karşılaştırma state'iyle açılmalı")
	}
	if !strings.Contains(r.confirmDialog.message, "Ege Denizi") || !strings.Contains(r.confirmDialog.message, "varsayılan tutumu") {
		t.Fatalf("temas bağlamı modal mesajında görünmeli: %q", r.confirmDialog.message)
	}
	modal := buildConfirmDialogModalFor(r.confirmDialog)
	if modal.Panel.Rect.Y >= ScreenHeight/2 {
		t.Fatalf("temas modalı ekranın üst kısmında olmalı: y=%v", modal.Panel.Rect.Y)
	}
	_, retreat, _, hasThird := buildConfirmDialogButtons(r.confirmDialog)
	if !hasThird || retreat.Enabled {
		t.Fatal("hareket hakkı olmayan filo için geri çekil düğmesi pasif olmalı")
	}
}

func TestNavalContactCameraTargetLeavesMapBelowModal(t *testing.T) {
	targetY := navalContactCameraTargetY(1338, 945)
	if targetY <= 945 {
		t.Fatalf("temas anchor'ı modalın altında kalmalı: y=%v", targetY)
	}
	if targetY > 1338-40 {
		t.Fatalf("temas anchor'ı ekranın altına taşmamalı: y=%v", targetY)
	}
}
