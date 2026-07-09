package render

import (
	"image/color"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestOwnerLabelOutlineColorUsesOppositeContrast(t *testing.T) {
	darkOutline := ownerLabelOutlineColor(color.RGBA{210, 210, 210, 255})
	if darkOutline != (color.RGBA{18, 16, 12, 220}) {
		t.Fatalf("acik renk icin koyu outline bekleniyordu, alinan=%v", darkOutline)
	}

	lightOutline := ownerLabelOutlineColor(color.RGBA{30, 70, 145, 255})
	if lightOutline != (color.RGBA{245, 240, 230, 210}) {
		t.Fatalf("koyu renk icin acik outline bekleniyordu, alinan=%v", lightOutline)
	}
}

func TestOwnerDisplayUsesSingleColor(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
			"ven": {ID: "ven", NameTR: "Venedik"},
		},
	}

	if _, col := ownerDisplay(gs, "osm"); col != ColorWhite {
		t.Fatalf("oyuncu devleti tek renk cikmali, alinan=%v", col)
	}
	if _, col := ownerDisplay(gs, "ven"); col != ColorWhite {
		t.Fatalf("ai devleti tek renk cikmali, alinan=%v", col)
	}
}

func TestVassalOverlordDisplayUsesOverlordInfo(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı", Color: [3]uint8{180, 60, 40}},
			"kar": {ID: "kar", NameTR: "Karaman", OverlordID: "osm"},
		},
	}

	label, col, ok := vassalOverlordDisplay(gs, "kar")
	if !ok {
		t.Fatal("vassal overlord satırı görünmeliydi")
	}
	if label != "Osmanlı (Siz)" {
		t.Fatalf("overlord etiketi oyuncu ekiyle gelmeliydi, got=%q", label)
	}
	if got := col.(color.RGBA); got != (color.RGBA{180, 60, 40, 235}) {
		t.Fatalf("overlord rengi kullanılmalıydı, got=%v", got)
	}
	if got := regionOwnerBlockHeight(gs, "kar"); got != float64(regionOwnerNameH)+regionVassalInfoH+regionVassalInfoH {
		t.Fatalf("vassal owner blok yüksekliği artmalıydı, got=%v", got)
	}
}

func TestVassalTributeDisplayUsesProjectedTribute(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
			"kar": {ID: "kar", NameTR: "Karaman", OverlordID: "osm"},
		},
		Regions: map[world.RegionID]*world.Region{
			"kon": {ID: "kon", OwnerID: "kar", BaseGoldIncome: 100, TaxRate: 50, Satisfaction: 50},
		},
	}

	label, col, ok := vassalTributeDisplay(gs, "kar")
	if !ok {
		t.Fatal("oyuncu vassalinin harac satiri görünmeliydi")
	}
	if label != "Haraç: +10 altın/tur" {
		t.Fatalf("beklenmeyen harac etiketi: %q", label)
	}
	if col != ColorGold {
		t.Fatalf("harac rengi altin olmaliydi, got=%v", col)
	}
}
