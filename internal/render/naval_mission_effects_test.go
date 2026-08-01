package render

import (
	"image/color"
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func navalMissionEffectsFixture() *state.GameState {
	return &state.GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"warship":   {ID: "warship", Category: army.CategoryNavalWar},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10},
			"infantry":  {ID: "infantry", Category: army.CategoryInfantry},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
			"sea-2": {ID: "sea-2", IsSea: true},
			"coast": {ID: "coast", Neighbors: []world.RegionID{"sea"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"patrol":    {ID: "patrol", OwnerID: "player", IsNaval: true, RegionID: "sea", Units: []army.Unit{{TypeID: "warship"}}, NavalMission: &army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}},
			"escort":    {ID: "escort", OwnerID: "player", IsNaval: true, RegionID: "sea", Units: []army.Unit{{TypeID: "warship"}}, NavalMission: &army.NavalMission{Kind: army.NavalMissionEscort, TargetFleetID: "transport"}},
			"transport": {ID: "transport", OwnerID: "player", IsNaval: true, RegionID: "sea", Units: []army.Unit{{TypeID: "transport"}}, EmbarkedUnits: []army.Unit{{TypeID: "infantry"}}},
			"loaded":    {ID: "loaded", OwnerID: "player", IsNaval: true, RegionID: "sea", Units: []army.Unit{{TypeID: "transport"}}, EmbarkedUnits: []army.Unit{{TypeID: "infantry"}}, NavalMission: &army.NavalMission{Kind: army.NavalMissionTransport, TargetRegionID: "coast"}},
		},
	}
}

func TestNavalMissionReachedRegionOnlyShowsActiveTarget(t *testing.T) {
	gs := navalMissionEffectsFixture()
	if got := navalMissionReachedRegion(gs, gs.Armies["patrol"]); got == nil || got.ID != "sea" {
		t.Fatalf("hedef denizdeki devriye aktif görünmeli: %+v", got)
	}
	if got := navalMissionReachedRegion(gs, gs.Armies["escort"]); got == nil || got.ID != "sea" {
		t.Fatalf("aynı denizdeki escort aktif görünmeli: %+v", got)
	}
	if got := navalMissionReachedRegion(gs, gs.Armies["loaded"]); got == nil || got.ID != "coast" {
		t.Fatalf("hedef kıyının deniz komşusundaki nakliye aktif görünmeli: %+v", got)
	}

	gs.Armies["patrol"].RegionID = "sea-2"
	if got := navalMissionReachedRegion(gs, gs.Armies["patrol"]); got != nil {
		t.Fatalf("hedefe ulaşmayan devriye aktif bonus etiketi göstermemeli: %+v", got)
	}
}

func TestNavalMissionReachedLabelsExposeDifferentEffects(t *testing.T) {
	want := map[army.NavalMissionKind]string{
		army.NavalMissionPatrol:    "DEVRİYE",
		army.NavalMissionBlockade:  "ABLUKA",
		army.NavalMissionEscort:    "ESCORT",
		army.NavalMissionTransport: "NAKLİYE",
	}
	for kind, marker := range want {
		if navalMissionReachedLabel(kind) == "" {
			t.Fatalf("%s için harita bonus etiketi boş", kind)
		}
		if !strings.Contains(navalMissionReachedLabel(kind), marker) {
			t.Fatalf("%s için marker etiketi görev türünü taşımıyor: %q", kind, navalMissionReachedLabel(kind))
		}
	}
}

func TestNavalMissionBonusBadgeUsesCompactActiveValues(t *testing.T) {
	gs := navalMissionEffectsFixture()
	if badge, _, ok := navalMissionBonusBadge(gs, gs.Armies["patrol"]); !ok || badge != "+1" {
		t.Fatalf("devriye rozeti aktif gemi bonusunu göstermeli: badge=%q ok=%t", badge, ok)
	}

	blockade := gs.Armies["patrol"]
	blockade.NavalMission = &army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "sea"}
	if badge, _, ok := navalMissionBonusBadge(gs, blockade); !ok || badge != "50" {
		t.Fatalf("tek savaş gemili abluka rozeti 50 göstermeli: badge=%q ok=%t", badge, ok)
	}
	blockade.Units = append(blockade.Units, army.Unit{TypeID: "warship"})
	if badge, _, ok := navalMissionBonusBadge(gs, blockade); !ok || badge != "100" {
		t.Fatalf("iki savaş gemili abluka rozeti 100 göstermeli: badge=%q ok=%t", badge, ok)
	}

	if badge, _, ok := navalMissionBonusBadge(gs, gs.Armies["escort"]); !ok || badge != "15" {
		t.Fatalf("escort rozeti 15 göstermeli: badge=%q ok=%t", badge, ok)
	}
	if badge, _, ok := navalMissionBonusBadge(gs, gs.Armies["loaded"]); ok || badge != "" {
		t.Fatalf("nakliye çıkarma görevi sayısal bonus rozeti üretmemeli: badge=%q ok=%t", badge, ok)
	}
}

func TestNavalMissionBonusTooltipIncludesTargetAndEffect(t *testing.T) {
	gs := navalMissionEffectsFixture()
	gs.Regions["sea"].NameTR = "Marmara"
	fleet := gs.Armies["patrol"]
	fleet.NavalMission = &army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "sea"}
	title, detail, ok := navalMissionBonusTooltipText(gs, fleet)
	if !ok || title != "Abluka bonusu" || !strings.Contains(detail, "Marmara") || !strings.Contains(detail, "-%50") {
		t.Fatalf("abluka tooltip hedef ve yüzde etkisini göstermeli: title=%q detail=%q ok=%t", title, detail, ok)
	}
}

func TestNavalEmbarkedArmyTooltipText(t *testing.T) {
	fleet := &army.Army{IsNaval: true, EmbarkedUnits: make([]army.Unit, 4)}
	title, detail, ok := navalEmbarkedArmyTooltipText(fleet)
	if !ok || title != "Nakliye Görevi" || detail != "Taşınan ordu 4 birim" {
		t.Fatalf("nakliye tooltip metni yanlış: title=%q detail=%q ok=%t", title, detail, ok)
	}
	if _, _, ok := navalEmbarkedArmyTooltipText(&army.Army{IsNaval: true}); ok {
		t.Fatal("boş filoda taşınan ordu tooltip'i gösterilmemeli")
	}
}

func TestNavalMissionBonusBadgeTextContrast(t *testing.T) {
	if got := navalMissionBonusBadgeTextColor(army.NavalMissionPatrol); got != (color.RGBA{35, 25, 15, 255}) {
		t.Fatalf("devriye bonus rozeti metin rengi yanlış: %+v", got)
	}
	if got := navalMissionBonusBadgeTextColor(army.NavalMissionBlockade); got != (color.RGBA{255, 255, 255, 255}) {
		t.Fatalf("abluka bonus rozeti metin rengi yanlış: %+v", got)
	}
}
