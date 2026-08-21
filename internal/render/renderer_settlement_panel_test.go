package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSettlementPanelHitRequiresVisiblePanel(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"morea": {
					ID: "morea",
					Settlements: []world.Settlement{
						{Name: "Mora"},
					},
				},
			},
		},
		SelectedRegion: "morea",
	}

	mx := float64(settlementPanelX() + 20)
	my := float64(settlementPanelY() + 20)

	if r.settlementPanelHit(mx, my) {
		t.Fatal("yerlesim paneli kapaliyken hit-test map alani tuketmemeli")
	}
	if r.settlementPanelCloseHit(mx, my) {
		t.Fatal("yerlesim paneli kapaliyken kapatma butonu aktif olmamali")
	}

	r.selectSettlement("morea", 0)

	if !r.settlementPanelHit(mx, my) {
		t.Fatal("yerlesim paneli acikken hit-test aktif olmali")
	}
}

func TestSettlementCapitalActionButtonVisibility(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_city"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:      "home",
				OwnerID: "player",
				Settlements: []world.Settlement{
					{ID: "capital_city", NameTR: "Başkent"},
					{ID: "other_city", NameTR: "Bursa"},
				},
			},
		},
	}

	if _, ok := settlementCapitalActionButton(gs, gs.Regions["home"], &gs.Regions["home"].Settlements[0]); ok {
		t.Fatal("mevcut başkent için taşıma butonu görünmemeli")
	}
	if _, ok := settlementCapitalActionButton(gs, gs.Regions["home"], &gs.Regions["home"].Settlements[1]); !ok {
		t.Fatal("oyuncunun diğer settlement'ı için taşıma butonu görünmeli")
	}
}

func TestSettlementCapitalStatusTextIncludesCapitalBonuses(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_city"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:      "home",
				OwnerID: "player",
				Settlements: []world.Settlement{
					{ID: "capital_city", NameTR: "Başkent"},
					{ID: "other_city", NameTR: "Bursa"},
				},
			},
		},
	}

	current := settlementCapitalStatusText(gs, gs.Regions["home"], &gs.Regions["home"].Settlements[0])
	if !strings.Contains(current, "+35 altın") || !strings.Contains(current, "+6 lojistik") {
		t.Fatalf("başkent statüsünde bonus özeti görünmeli: %q", current)
	}
	if !strings.Contains(current, "depoların yarısı") {
		t.Fatalf("başkent düşüş riski görünmeli: %q", current)
	}

	other := settlementCapitalStatusText(gs, gs.Regions["home"], &gs.Regions["home"].Settlements[1])
	if !strings.Contains(other, "5 tur") || !strings.Contains(other, "+35 altın") {
		t.Fatalf("taşıma öncesi başkent özeti görünmeli: %q", other)
	}
}

func TestSettlementSelectionOverlayTargetsSelectedRegionCenter(t *testing.T) {
	region := &world.Region{
		ID: "home",
		Settlements: []world.Settlement{
			{ID: "home_center", IsCenter: true},
			{ID: "home_town"},
		},
	}
	r := &Renderer{
		gs:             &state.GameState{Regions: map[world.RegionID]*world.Region{"home": region}},
		SelectedRegion: "home",
	}

	if !r.shouldDrawSettlementSelectionOverlay(region, 0, region.Settlements[0]) {
		t.Fatal("seçili bölgenin merkez settlement marker'ı vurgulanmalı")
	}
	if r.shouldDrawSettlementSelectionOverlay(region, 1, region.Settlements[1]) {
		t.Fatal("seçili bölgenin merkez olmayan settlement marker'ı vurgulanmamalı")
	}

	other := &world.Region{
		ID:          "other",
		Settlements: []world.Settlement{{ID: "other_center", IsCenter: true}},
	}
	if r.shouldDrawSettlementSelectionOverlay(other, 0, other.Settlements[0]) {
		t.Fatal("başka bölgenin merkez settlement marker'ı vurgulanmamalı")
	}
}

func TestAppendSettlementDrawsKeepsFactionCapitalLabelVisible(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	region := &world.Region{
		ID:      "home",
		OwnerID: "player",
		Settlements: []world.Settlement{
			{ID: "capital_city", NameTR: "Yeni Başkent", Type: world.SettlementTown},
		},
	}
	r := &Renderer{
		camScale: 0.7,
		gs: &state.GameState{
			PlayerFactionID: "player",
			Factions: map[faction.FactionID]*faction.Faction{
				"player": {ID: "player", CapitalSettlementID: "capital_city"},
			},
			Regions: map[world.RegionID]*world.Region{
				"home": region,
			},
		},
		worldMap: &WorldMap{
			settlementAnchor: map[settlementAnchorKey][2]int{
				{Region: "home", Index: 0}: {120, 120},
			},
		},
	}

	r.appendSettlementDraws(region)
	if len(r.regionLabelBuf) != 1 {
		t.Fatalf("1 settlement draw bekleniyordu, got=%d", len(r.regionLabelBuf))
	}
	item := r.regionLabelBuf[0]
	if !item.DrawLabel {
		t.Fatalf("faction capital düşük zoomda da label çizdirmeli: %+v", item)
	}
	if !item.CapitalIcon {
		t.Fatalf("faction capital label başında star ikonu taşımalı: %+v", item)
	}
	if item.TextX <= item.X {
		t.Fatalf("capital icon için text başlangıcı kaydırılmalı: %+v", item)
	}
}

func TestSettlementVisibilityUsesZoomTiers(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	region := &world.Region{
		ID: "home",
		Settlements: []world.Settlement{
			{ID: "capital", NameTR: "Başkent", Type: world.SettlementCity, IsCenter: true},
			{ID: "castle", NameTR: "Kale", Type: world.SettlementFortress},
			{ID: "port", NameTR: "Liman", Type: world.SettlementPort},
			{ID: "city", NameTR: "Şehir", Type: world.SettlementCity},
			{ID: "town", NameTR: "Kasaba", Type: world.SettlementTown},
		},
	}
	anchors := make(map[settlementAnchorKey][2]int, len(region.Settlements))
	for i := range region.Settlements {
		anchors[settlementAnchorKey{Region: region.ID, Index: i}] = [2]int{120 + i*40, 120}
	}

	r := &Renderer{
		worldMap: &WorldMap{settlementAnchor: anchors},
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{region.ID: region},
		},
	}

	tests := []struct {
		name  string
		scale float64
		want  map[string]bool
	}{
		{
			name:  "uzak",
			scale: 0.45,
			want:  map[string]bool{"capital": true, "castle": true},
		},
		{
			name:  "orta",
			scale: 1.4,
			want:  map[string]bool{"capital": true, "castle": true, "port": true, "city": true},
		},
		{
			name:  "yakin",
			scale: 2.0,
			want:  map[string]bool{"capital": true, "castle": true, "port": true, "city": true, "town": true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r.camScale = tc.scale
			r.regionLabelBuf = r.regionLabelBuf[:0]
			r.appendSettlementDraws(region)

			got := make(map[string]bool, len(r.regionLabelBuf))
			for _, item := range r.regionLabelBuf {
				got[item.Region.Settlements[item.Index].ID] = true
			}
			if len(got) != len(tc.want) {
				t.Fatalf("zoom %.2f için settlement sayısı: got=%v want=%v", tc.scale, got, tc.want)
			}
			for id, wantVisible := range tc.want {
				if got[id] != wantVisible {
					t.Fatalf("zoom %.2f için %q görünürlüğü: got=%v want=%v", tc.scale, id, got[id], wantVisible)
				}
			}
		})
	}
}

func TestSettlementMarkerSpriteSwitchesToSiegeAssetForPrimarySiegedSettlement(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"home": {ID: "home", OwnerID: "player"},
			},
			Sieges: map[world.RegionID]*state.SiegeState{
				"home": {RegionID: "home", AttackerArmyID: "atk"},
			},
		},
	}
	region := r.gs.Regions["home"]
	settlement := world.Settlement{ID: "capital_city", Type: world.SettlementCity}
	if img := r.settlementMarkerSprite(region, settlement, true); img == nil {
		t.Fatal("kuşatma altındaki ana yerleşim için siege sprite seçilmeli")
	}
	if img := r.settlementMarkerSprite(region, settlement, true); img != settlementMarkerSiegeImage() {
		t.Fatal("kuşatma altındaki ana yerleşim siege.png ile çizilmeli")
	}
}

func TestSettlementMarkerSizeShrinksPortsOnly(t *testing.T) {
	r := &Renderer{}
	region := &world.Region{ID: "coast"}

	for _, settlementType := range []world.SettlementType{
		world.SettlementCity,
		world.SettlementTown,
		world.SettlementFortress,
	} {
		if got := r.settlementMarkerSize(region, world.Settlement{Type: settlementType}, true); got != settlementMarkerSpriteSize {
			t.Fatalf("%s marker boyutu değişmemeli: got=%.1f want=%.1f", settlementType, got, settlementMarkerSpriteSize)
		}
	}
	if got := r.settlementMarkerSize(region, world.Settlement{Type: world.SettlementPort}, true); got != settlementPortMarkerSize {
		t.Fatalf("liman marker'ı küçültülmeli: got=%.1f want=%.1f", got, settlementPortMarkerSize)
	}
}

func TestSettlementBreachBorderRequiresOpenedBreachAndIgnoresSelection(t *testing.T) {
	region := &world.Region{ID: "home", Settlements: []world.Settlement{{ID: "home_center", IsCenter: true}}}
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{"home": region},
			Sieges: map[world.RegionID]*state.SiegeState{
				"home": {RegionID: "home", BreachProgress: 10, BreachLevel: 0},
			},
		},
	}

	if r.isSettlementBreachOpen(region) {
		t.Fatal("gedik eşiğine ulaşmayan kuşatma yeşil border almamalı")
	}
	r.gs.Sieges["home"].BreachLevel = 1
	if !r.isSettlementBreachOpen(region) {
		t.Fatal("gedik oluşan kuşatma yeşil border almalı")
	}
	r.SelectedRegion = "home"
	if !r.isSettlementBreachOpen(region) {
		t.Fatal("seçim durumu gedik border'ını kapatmamalı")
	}
}

func TestArmySiegeBadgeUsesSwordAsset(t *testing.T) {
	if armySiegeBadgeImage() != settlementMarkerSwordImage() {
		t.Fatal("kuşatan ordu badge'i sword.png sprite'ını kullanmalı")
	}
}

func TestBuildingGridHitTestUsesDrawnSpriteBoundsOnly(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	oldCachedRegion := lastBuildingGridRegionID
	oldCachedCards := lastBuildingGridCards
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
		lastBuildingGridRegionID = oldCachedRegion
		lastBuildingGridCards = oldCachedCards
	}()

	regionID := world.RegionID("bitinya")
	region := &world.Region{
		ID:        regionID,
		NameTR:    "Bitinya",
		OwnerID:   "osm",
		Terrain:   world.TerrainPlain,
		Buildings: []string{"market"},
	}
	gs := &state.GameState{
		PlayerFactionID: "osm",
		Regions: map[world.RegionID]*world.Region{
			regionID: region,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", NameTR: "Pazar", MaxPerRegion: 1},
		},
	}

	startY := buildingGridStartY(gs, region, false)
	cards := buildBuildingCardComponents(gs, region, infoPanelX(), startY, infoPanelW)
	if len(cards) != 1 {
		t.Fatalf("1 bina karti bekleniyordu, got=%d", len(cards))
	}
	card := cards[0]

	if got := BuildingGridHitTest(card.ImageRect.X+card.ImageRect.W/2, card.ImageRect.Y+card.ImageRect.H/2, gs, regionID, false); got != "market" {
		t.Fatalf("resim merkezinde market bekleniyordu, got=%q", got)
	}
	if got := BuildingGridHitTest(card.Rect.X+card.Rect.W/2, card.LabelY+1, gs, regionID, false); got != "" {
		t.Fatalf("label alani hit olmamali, got=%q", got)
	}
	if got := BuildingGridHitTest(card.Rect.X+card.Rect.W/2, card.Rect.Y+card.Rect.H+3, gs, regionID, false); got != "" {
		t.Fatalf("kart disi bosluk hit olmamali, got=%q", got)
	}

	card.SpriteRect = card.ImageRect
	card.SpriteRect.Y += 20
	card.SpriteRect.H -= 20
	cacheBuildingGridComponents(regionID, []buildingCardComponent{card})
	if got := BuildingGridHitTest(card.ImageRect.X+card.ImageRect.W/2, card.ImageRect.Y+4, gs, regionID, false); got != "" {
		t.Fatalf("son cizilen sprite disindaki ust alan hit olmamali, got=%q", got)
	}
	if got := BuildingGridHitTest(card.SpriteRect.X+card.SpriteRect.W/2, card.SpriteRect.Y+card.SpriteRect.H/2, gs, regionID, false); got != "market" {
		t.Fatalf("son cizilen sprite merkezinde market bekleniyordu, got=%q", got)
	}
}

func TestNonOwnedBuildingCardsAreNotActionableOrHoverable(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	oldCachedRegion := lastBuildingGridRegionID
	oldCachedCards := lastBuildingGridCards
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
		lastBuildingGridRegionID = oldCachedRegion
		lastBuildingGridCards = oldCachedCards
	}()

	regionID := world.RegionID("enemy_region")
	region := &world.Region{
		ID:        regionID,
		OwnerID:   "enemy",
		Buildings: []string{"market"},
	}
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions:         map[world.RegionID]*world.Region{regionID: region},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 1000, Grain: 1000, Iron: 1000, Timber: 1000, Stone: 1000},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", NameTR: "Pazar", GoldCost: 10, MaxPerRegion: 2},
		},
	}

	startY := buildingGridStartY(gs, region, false)
	cards := buildBuildingCardComponents(gs, region, infoPanelX(), startY, infoPanelW)
	if len(cards) != 1 {
		t.Fatalf("1 bina kartı bekleniyordu, got=%d", len(cards))
	}
	card := cards[0]
	cacheBuildingGridComponents(regionID, cards)

	if card.CanAfford {
		t.Fatal("oyuncuya ait olmayan bölgedeki bina kartı yeşile dönebilecek şekilde CanAfford olmamalı")
	}
	if got := BuildingGridHoverID(card.SpriteRect.X+card.SpriteRect.W/2, card.SpriteRect.Y+card.SpriteRect.H/2, gs, regionID); got != "" {
		t.Fatalf("oyuncuya ait olmayan bölgedeki bina hover tooltip üretmemeliydi, got=%q", got)
	}
	if got := BuildingGridHoverIDForTab(card.SpriteRect.X+card.SpriteRect.W/2, card.SpriteRect.Y+card.SpriteRect.H/2, gs, regionID, regionPanelTabEvents); got != "" {
		t.Fatalf("olaylar sekmesinde bina hover tooltip üretmemeliydi, got=%q", got)
	}
	if got := BuildingGridHitTest(card.SpriteRect.X+card.SpriteRect.W/2, card.SpriteRect.Y+card.SpriteRect.H/2, gs, regionID, false); got != "" {
		t.Fatalf("oyuncuya ait olmayan bölgedeki bina tıklanabilir olmamalıydı, got=%q", got)
	}
}
