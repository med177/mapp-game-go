package render

import (
	"math"
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

func TestEventIconScreenAnchorDoesNotApplyShapeTransformTwice(t *testing.T) {
	oldShapeOffX, oldShapeOffY := shapeOffX, shapeOffY
	oldShapeScaleX, oldShapeScaleY := shapeScaleX, shapeScaleY
	defer func() {
		shapeOffX, shapeOffY = oldShapeOffX, oldShapeOffY
		shapeScaleX, shapeScaleY = oldShapeScaleX, oldShapeScaleY
	}()
	shapeOffX, shapeOffY = -100, -50
	shapeScaleX, shapeScaleY = 2, 3

	r := &Renderer{
		camScale: 1,
		worldMap: &WorldMap{
			regionAnchor: map[world.RegionID][2]int{
				"land_region": {400, 300},
			},
		},
	}

	sx, sy, ok := r.eventIconScreenAnchor("land_region")
	if !ok {
		t.Fatal("event anchor bulunmalıydı")
	}
	wantX, wantY := r.worldToScreen(400, 300)
	if math.Abs(sx-wantX) > 0.001 || math.Abs(sy-wantY) > 0.001 {
		t.Fatalf("world-pixel anchor doğrudan ekrana çevrilmeli: got=(%.2f,%.2f) want=(%.2f,%.2f)", sx, sy, wantX, wantY)
	}
	doubleX, doubleY := r.worldToScreen(wcX(400), wcY(300))
	if math.Abs(sx-doubleX) < 0.001 && math.Abs(sy-doubleY) < 0.001 {
		t.Fatalf("anchor shape dönüşümünü iki kez uyguladı: got=(%.2f,%.2f)", sx, sy)
	}
}

func TestEventIconAnchorFallbackConvertsRegionCenterToWorldPixels(t *testing.T) {
	oldShapeOffX, oldShapeOffY := shapeOffX, shapeOffY
	oldShapeScaleX, oldShapeScaleY := shapeScaleX, shapeScaleY
	defer func() {
		shapeOffX, shapeOffY = oldShapeOffX, oldShapeOffY
		shapeScaleX, shapeScaleY = oldShapeScaleX, oldShapeScaleY
	}()
	shapeOffX, shapeOffY = -100, -50
	shapeScaleX, shapeScaleY = 2, 3

	r := &Renderer{gs: &state.GameState{Regions: map[world.RegionID]*world.Region{
		"land_region": {ID: "land_region", WorldX: 250, WorldY: 120},
	}}}
	ax, ay, ok := r.eventIconAnchor("land_region")
	if !ok {
		t.Fatal("bölge merkezi fallback anchor üretmeliydi")
	}
	if ax != 400 || ay != 310 {
		t.Fatalf("fallback world-pixel uzayında olmalı: got=(%d,%d) want=(400,310)", ax, ay)
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

func TestClampActiveRegionEventScreenPointKeepsMarkerInsideRegion(t *testing.T) {
	oldShapeOffX, oldShapeOffY := shapeOffX, shapeOffY
	oldShapeScaleX, oldShapeScaleY := shapeScaleX, shapeScaleY
	defer func() {
		shapeOffX, shapeOffY = oldShapeOffX, oldShapeOffY
		shapeScaleX, shapeScaleY = oldShapeScaleX, oldShapeScaleY
	}()
	shapeOffX, shapeOffY = 0, 0
	shapeScaleX, shapeScaleY = 2, 2

	region := &world.Region{
		ID: "land_region",
		Shape: [][][2]float32{
			{
				{90, 90},
				{110, 90},
				{110, 110},
				{90, 110},
			},
		},
	}
	r := &Renderer{
		camScale: 1,
		camX:     100,
		camY:     100,
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"land_region": region,
			},
		},
		worldMap: &WorldMap{
			regionPx: map[world.RegionID][]int{
				"land_region": {200 + 200*WorldW, 202 + 200*WorldW, 204 + 202*WorldW},
			},
		},
	}

	srcSX, srcSY := r.worldToScreen(wcX(130), wcY(130))
	dstSX, dstSY := r.clampActiveRegionEventScreenPoint("land_region", srcSX, srcSY)
	if dstSX == srcSX && dstSY == srcSY {
		t.Fatal("bolge disi event noktasi clamp edilmedi")
	}
	wx, wy := r.screenToWorld(dstSX, dstSY)
	pureX := (wx - shapeOffX) / shapeScaleX
	pureY := (wy - shapeOffY) / shapeScaleY
	if !regionContainsPoint(region, pureX, pureY) {
		t.Fatalf("clamp edilen nokta hala bolge icinde degil: pure=(%.2f,%.2f)", pureX, pureY)
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
				"land_region": {int(math.Round(wcX(400))), int(math.Round(wcY(100)))},
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
