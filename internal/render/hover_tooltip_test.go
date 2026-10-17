package render

import (
	"math"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestUnitFishingIncomeTooltipLineShowsConfiguredYield(t *testing.T) {
	if got, want := unitFishingIncomeTooltipLine(&army.UnitType{FishingGrainIncome: 40}), "Balıkçılık: +40 tahıl/tur"; got != want {
		t.Fatalf("balıkçılık tooltip satırı = %q, want %q", got, want)
	}
	if got := unitFishingIncomeTooltipLine(&army.UnitType{}); got != "" {
		t.Fatalf("getirisi olmayan birim için balıkçılık satırı = %q, want empty", got)
	}
	if got := unitFishingIncomeTooltipLine(nil); got != "" {
		t.Fatalf("nil birim için balıkçılık satırı = %q, want empty", got)
	}
}

func TestUnitTooltipLayoutPlacesRequirementsBelowImage(t *testing.T) {
	const iconH = 201.0

	layout := newUnitTooltipLayout(iconH, 6, 12)

	if layout.requirementY < 14+iconH+10 {
		t.Fatalf("gereksinim başlığı ikonun altına taşmıyor: got %v", layout.requirementY)
	}
	if layout.requirementLinesY != layout.requirementY+14 {
		t.Fatalf("gereksinim satır başlangıcı = %v, want %v", layout.requirementLinesY, layout.requirementY+14)
	}
	if layout.upkeepY >= layout.requirementY {
		t.Fatalf("bakım üst bölüme taşınmamış: bakım=%v gereksinim=%v", layout.upkeepY, layout.requirementY)
	}
	if layout.attributesY >= layout.requirementY {
		t.Fatalf("nitelikler üst bölüme taşınmamış: nitelik=%v gereksinim=%v", layout.attributesY, layout.requirementY)
	}
	if layout.height <= layout.requirementLinesY+12*14 {
		t.Fatalf("tooltip yüksekliği gereksinim bloğunu kapsamıyor: height=%v", layout.height)
	}
}

func TestUnitTooltipImageMetricsPreservesSpriteAspect(t *testing.T) {
	sprite := ebiten.NewImage(300, 500)
	width, height := unitTooltipImageMetrics(sprite)

	if got, want := width/height, 300.0/500.0; math.Abs(got-want) > 1e-9 {
		t.Fatalf("tooltip görsel oranı = %v, want %v", got, want)
	}
}

func TestArmyMarkerSummaryTooltipReservesCommanderPortraitSpace(t *testing.T) {
	withoutCommanderWidth, _ := armyMarkerSummaryTooltipSize(0, false, 0)
	withCommanderWidth, withCommanderHeight := armyMarkerSummaryTooltipSize(0, true, 0)
	wantPortraitSize := float64(armyPanelCommanderPortrait)

	if got, want := withCommanderWidth-withoutCommanderWidth, wantPortraitSize+armySummaryCommanderGap; got != want {
		t.Fatalf("komutanlı popup genişlik farkı = %.0f, want %.0f", got, want)
	}
	if withCommanderHeight < wantPortraitSize+2*armySummaryTooltipPad {
		t.Fatalf("komutanlı popup yüksekliği %.0f, portreyi iç boşlukla sığdırmıyor", withCommanderHeight)
	}
	_, withUnitRowHeight := armyMarkerSummaryTooltipSize(1, true, 0)
	if withUnitRowHeight < withCommanderHeight {
		t.Fatalf("birim satırlı popup yüksekliği %.0f, portreli boş popup'tan küçük", withUnitRowHeight)
	}
}

func TestArmyMarkerSummaryTooltipReservesCommanderEffectLines(t *testing.T) {
	_, portraitOnlyHeight := armyMarkerSummaryTooltipSize(0, true, 0)
	_, effectHeight := armyMarkerSummaryTooltipSize(0, true, 5)
	wantHeight := armySummaryTooltipPad*2 + armySummaryCommanderSize + armySummaryEffectsGap + 5*armySummaryEffectLineH

	if effectHeight != wantHeight {
		t.Fatalf("beş bonus satırlı popup yüksekliği = %.0f, want %.0f", effectHeight, wantHeight)
	}
	if effectHeight <= portraitOnlyHeight {
		t.Fatalf("bonuslu popup yüksekliği %.0f, portreli popup'tan uzun olmalı (%.0f)", effectHeight, portraitOnlyHeight)
	}
}

func TestArmyMarkerCasualtyHPUsesArmyOrRegionLogistics(t *testing.T) {
	const (
		landArmy army.ArmyID    = "land-army"
		fleet    army.ArmyID    = "fleet"
		region   world.RegionID = "front"
	)
	gs := &state.GameState{
		RegionLogistics: map[world.RegionID]state.RegionLogisticsStatus{
			region: {TotalHPDamage: 275},
		},
		ArmyLogistics: map[army.ArmyID]state.ArmyLogisticsStatus{
			fleet: {TotalHPDamage: 120},
		},
	}

	if got := armyMarkerCasualtyHP(gs, &army.Army{ID: landArmy, RegionID: region}); got != 275 {
		t.Fatalf("kara ordusu zayiatı = %d HP, want 275 HP", got)
	}
	if got := armyMarkerCasualtyHP(gs, &army.Army{ID: fleet, IsNaval: true, RegionID: region}); got != 120 {
		t.Fatalf("filo zayiatı = %d HP, want 120 HP", got)
	}
}

func TestArmyMarkerUnitImageScalePreservesAspectAndFillsTileWidth(t *testing.T) {
	scale, ok := armyMarkerUnitImageScale(300, 500, float32(armySummaryTileWidth))
	if !ok {
		t.Fatal("geçerli kaynak görsel ölçüleri reddedildi")
	}

	imageWidth := float64(300) * scale
	imageHeight := float64(500) * scale
	if imageWidth != armySummaryTileWidth {
		t.Errorf("görsel genişliği = %.0f, kart genişliği %.0f olmalı", imageWidth, armySummaryTileWidth)
	}
	if got, want := imageWidth/imageHeight, 300.0/500.0; math.Abs(got-want) > 1e-9 {
		t.Errorf("görsel oranı = %v, kaynak oranı %v olmalı", got, want)
	}
	if imageHeight >= armySummaryTileHeight {
		t.Errorf("görsel yüksekliği = %.0f, kart altındaki adet alanına yer kalmalı", imageHeight)
	}
}
