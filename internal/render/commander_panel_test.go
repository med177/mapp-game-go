package render

import (
	"path/filepath"
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	gameui "mapp-game-go/internal/ui"
)

func TestCommanderPortraitPathUsesCommandersDirForBasename(t *testing.T) {
	ActiveScenarioPath = filepath.Join("assets", "scenarios", "1300_ottoman_rise")
	t.Cleanup(func() {
		ActiveScenarioPath = ""
		resetCommanderPortraitCache()
	})

	got := commanderPortraitPath("ottoman_osman.png")
	want := filepath.Join(ActiveScenarioPath, "sprites", "commanders", "ottoman_osman.png")
	if got != want {
		t.Fatalf("basename portrait yanlis cozuldu: got=%q want=%q", got, want)
	}
}

func TestCommanderPortraitPathKeepsRelativeSubdir(t *testing.T) {
	ActiveScenarioPath = filepath.Join("assets", "scenarios", "1300_ottoman_rise")
	t.Cleanup(func() {
		ActiveScenarioPath = ""
		resetCommanderPortraitCache()
	})

	got := commanderPortraitPath("commanders/ottoman_osman.png")
	want := filepath.Join(ActiveScenarioPath, "sprites", "commanders", "ottoman_osman.png")
	if got != want {
		t.Fatalf("relative portrait path korunmadi: got=%q want=%q", got, want)
	}
}

func TestCommanderPortraitPathRejectsTraversal(t *testing.T) {
	ActiveScenarioPath = filepath.Join("assets", "scenarios", "1300_ottoman_rise")
	t.Cleanup(func() {
		ActiveScenarioPath = ""
		resetCommanderPortraitCache()
	})

	if got := commanderPortraitPath("../secret.png"); got != "" {
		t.Fatalf("traversal path reddedilmeliydi: %q", got)
	}
}

func TestCommanderTraitBadgeUsesCompactLabels(t *testing.T) {
	tests := map[army.CommanderTrait]string{
		army.CommanderTraitVeteran:   "Tecrübe",
		army.CommanderTraitTactician: "Taktik",
		army.CommanderTraitDefender:  "Savunma",
		army.CommanderTraitAggressor: "Saldırı",
	}

	for trait, want := range tests {
		if got := commanderTraitBadge(trait).Label; got != want {
			t.Fatalf("trait %q için badge etiketi yanlış: got=%q want=%q", trait, got, want)
		}
	}
}

func TestCommanderOverflowBadgeUsesCountLabel(t *testing.T) {
	if got := commanderOverflowBadge(3).Label; got != "+3" {
		t.Fatalf("overflow badge etiketi yanlış: got=%q want=%q", got, "+3")
	}
}

func TestCommanderCardEffectLinesSplitModifiersPerLine(t *testing.T) {
	commander := &army.Commander{
		Traits: []army.CommanderTrait{
			army.CommanderTraitVeteran,
			army.CommanderTraitTactician,
			army.CommanderTraitDefender,
			army.CommanderTraitAggressor,
		},
	}
	lines := commanderCardEffectLines(commander)
	if len(lines) != 5 {
		t.Fatalf("beklenen 5 effect line, got=%d", len(lines))
	}
	for _, want := range []string{"Saldırı +", "Savunma +", "Moral +", "Hareket +1", "Kuşatma +1/+1"} {
		found := false
		for _, line := range lines {
			if strings.Contains(line.Text, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("effect line listesinde %q yok: %+v", want, lines)
		}
	}
}

func TestCommanderSummaryDividerYMovesBelowEffectLines(t *testing.T) {
	portraitY := 28.0
	got := commanderSummaryDividerY(portraitY, 5)
	lastEffectY := portraitY + 64 + float64(5-1)*12
	if got <= lastEffectY {
		t.Fatalf("divider effect line altına inmeli: got=%.1f last=%.1f", got, lastEffectY)
	}
}

func TestCommanderSummaryHeaderTextsSwapNameAndRole(t *testing.T) {
	commander := army.NewCommander("cmd_1", "Amadeus V of Savoy")
	top, right := commanderSummaryHeaderTexts("Komutan", commander)
	if top != "Komutan" {
		t.Fatalf("ust etiket rol olmaliydi: %q", top)
	}
	if right != "Amadeus V of Savoy" {
		t.Fatalf("sag baslik isim olmaliydi: %q", right)
	}
}

func TestArmyPanelCommanderCardFitsProfileAndTraitOverflow(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	layout := armyPanelGeometry()
	textW := commanderSummaryTextWidth(float64(layout.commanderW))
	for _, line := range []string{
		"Seviye 5  |  1500 XP",
		"Savaş 15  |  Zafer 15",
	} {
		if MeasureText(line, FaceSmall) > textW {
			t.Fatalf("komutan profil satırı karta sığmıyor: line=%q width=%.1f textW=%.1f", line, MeasureText(line, FaceSmall), textW)
		}
	}

	dividerY := commanderSummaryDividerY(float64(layout.commanderY)+28, 5)
	badgesBottomY := dividerY + 24 + 18 + 6 + 18
	cardBottomY := float64(layout.commanderY + layout.commanderH)
	if badgesBottomY > cardBottomY {
		t.Fatalf("komutan uzmanlık rozetleri karttan taşıyor: badgesBottom=%.1f cardBottom=%.1f", badgesBottomY, cardBottomY)
	}
}

func TestCommanderPanelListViewportAndScrollClamp(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	viewport := commanderPanelListViewport(nil, "army_1")
	if viewport.Y <= commanderPanelRect().Y || viewport.Y+viewport.H > commanderPanelRect().Y+commanderPanelRect().H-24 {
		t.Fatalf("komutan listesi panel sınırları içinde değil: viewport=%+v panel=%+v", viewport, commanderPanelRect())
	}
	if visible := commanderPanelVisibleRows(viewport); visible != 4 {
		t.Fatalf("beklenen görünür komutan satırı 4, got=%d viewport=%+v", visible, viewport)
	}
	if got := clampCommanderPanelScroll(99, 7, viewport); got != 3 {
		t.Fatalf("scroll üst sınırı yanlış: got=%d want=3", got)
	}
	if got := clampCommanderPanelScroll(-1, 7, viewport); got != 0 {
		t.Fatalf("scroll alt sınırı yanlış: got=%d want=0", got)
	}
}

func TestCommanderPanelRowHitOnlyUsesVisibleRows(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	viewport := commanderPanelListViewport(nil, "army_1")
	lastVisible := commanderPanelRow(3, 0)
	if got := commanderPanelRowAt(lastVisible.X+4, lastVisible.Y+4, 7, 0, viewport); got != 3 {
		t.Fatalf("görünür satır yanlış eşlendi: got=%d want=3", got)
	}
	if got := commanderPanelRowAt(lastVisible.X+4, lastVisible.Y+4, 7, 3, viewport); got != 6 {
		t.Fatalf("scroll sonrası satır yanlış eşlendi: got=%d want=6", got)
	}
	outside := gameui.Rect{X: viewport.X, Y: viewport.Y + viewport.H + 2, W: viewport.W, H: 20}
	if got := commanderPanelRowAt(outside.X+4, outside.Y+4, 7, 0, viewport); got != -1 {
		t.Fatalf("viewport dışı satır tıklanabilir olmamalı: got=%d", got)
	}
}
