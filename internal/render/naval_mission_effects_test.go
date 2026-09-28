package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/state"
)

func TestNavalSupplyCargoTooltipTextShowsForeignFleet(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Armies:          map[army.ArmyID]*army.Army{},
	}
	fleet := &army.Army{
		ID:          "foreign-fleet",
		OwnerID:     "other-faction",
		IsNaval:     true,
		SupplyCargo: economy.ResourceCost{Grain: 24},
	}

	title, detail, ok := navalSupplyCargoTooltipText(gs, fleet)
	if !ok {
		t.Fatal("başka devletin ikmal filosu için bilgi popup'ı gösterilmedi")
	}
	if title != "İkmal Yükü" {
		t.Fatalf("popup başlığı = %q, want %q", title, "İkmal Yükü")
	}
	if detail == "" || !strings.Contains(detail, "Tahıl: 24") {
		t.Fatalf("popup ikmal miktarını göstermedi: %q", detail)
	}
}
