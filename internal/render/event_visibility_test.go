package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestEventIconAnchorUsesSeaRegionAnchor(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"aegean_sea": {
					ID:     "aegean_sea",
					WorldX: 400,
					WorldY: 300,
					IsSea:  true,
				},
			},
		},
		worldMap: &WorldMap{
			regionAnchor: map[world.RegionID][2]int{
				"aegean_sea": {400, 300},
			},
		},
	}

	ax, ay, ok := r.eventIconAnchor("aegean_sea")
	if !ok {
		t.Fatal("deniz bölgesi için anchor bulunmalıydı")
	}
	if ax != 400 || ay != 300 {
		t.Fatalf("deniz bölgesi anchor'u yanlış: got=(%d,%d) want=(400,300)", ax, ay)
	}
}

func TestActiveRegionEventVisibleSkipsSeaRegions(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"land_region": {ID: "land_region"},
			"sea_region":  {ID: "sea_region", IsSea: true},
		},
	}

	if !activeRegionEventVisible(gs, state.RegionEventStatus{RegionID: "land_region"}) {
		t.Fatal("kara bolgesi event ikonu gorunur olmali")
	}
	if activeRegionEventVisible(gs, state.RegionEventStatus{RegionID: "sea_region"}) {
		t.Fatal("deniz bolgesi event ikonu gorunmemeli")
	}
	if activeRegionEventVisible(gs, state.RegionEventStatus{RegionID: "missing"}) {
		t.Fatal("bilinmeyen bolge event ikonu gorunmemeli")
	}
}

func TestMinimapEventMarkerPositionAndStackOffset(t *testing.T) {
	oldWorldW, oldWorldH := WorldW, WorldH
	defer func() {
		WorldW = oldWorldW
		WorldH = oldWorldH
	}()
	WorldW, WorldH = 2892, 1440

	region := &world.Region{
		ID:     "land_region",
		WorldX: 700,
		WorldY: 420,
	}

	scaleX := float32(minimapW) / float32(WorldW)
	scaleY := float32(minimapH) / float32(WorldH)
	px, py := minimapEventMarkerPosition(region, scaleX, scaleY, minimapX(), minimapY())
	if px <= minimapX() || py <= minimapY() {
		t.Fatalf("marker pozisyonu minimap içine düşmeli: got=(%.2f,%.2f)", px, py)
	}

	events := []state.RegionEventStatus{
		{RegionID: "land_region"},
		{RegionID: "sea_region"},
		{RegionID: "land_region"},
		{RegionID: "land_region"},
	}
	if got := minimapEventMarkerStackOffset(events, 0, "land_region"); got != 0 {
		t.Fatalf("ilk event stack ofseti 0 olmalı, got=%d", got)
	}
	if got := minimapEventMarkerStackOffset(events, 2, "land_region"); got != 1 {
		t.Fatalf("ikinci land event stack ofseti 1 olmalı, got=%d", got)
	}
	if got := minimapEventMarkerStackOffset(events, 3, "land_region"); got != 2 {
		t.Fatalf("üçüncü land event stack ofseti 2 olmalı, got=%d", got)
	}
}

func TestActiveRegionEventHitAndDetail(t *testing.T) {
	r := &Renderer{
		camScale: 1,
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"land_region": {
					ID:     "land_region",
					NameTR: "Test Bölgesi",
					WorldX: 400,
					WorldY: 100,
				},
			},
			ActiveRegionEvents: []state.RegionEventStatus{
				{EventID: "evt_1", RegionID: "land_region", TurnsLeft: 4, Type: "plague", LabelTR: "Veba"},
				{EventID: "evt_2", RegionID: "land_region", TurnsLeft: 2, Type: "revolt", LabelTR: "İsyan"},
			},
		},
		worldMap: &WorldMap{
			regionAnchor: map[world.RegionID][2]int{
				"land_region": {400, 100},
			},
		},
	}

	x, y := r.worldToScreen(wcX(400), wcY(100))
	if idx, ok := r.activeRegionEventHitAt(x, y-12); !ok || idx != 0 {
		t.Fatalf("ilk stacked event hit edilmeliydi, got=(%d,%t)", idx, ok)
	}
	if idx, ok := r.activeRegionEventHitAt(x, y+12); !ok || idx != 1 {
		t.Fatalf("ikinci stacked event hit edilmeliydi, got=(%d,%t)", idx, ok)
	}

	detail := r.activeRegionEventDetailAt(1)
	if !strings.Contains(detail, "Test Bölgesi") {
		t.Fatalf("detay bölge adını içermeli: %s", detail)
	}
	if !strings.Contains(detail, "Kalan tur: 2") {
		t.Fatalf("detay kalan tur bilgisini içermeli: %s", detail)
	}
	if !strings.Contains(detail, "Event ID: evt_2") {
		t.Fatalf("detay event id içermeli: %s", detail)
	}
}
