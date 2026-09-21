package render

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSortedFactionsForMarketSortsSellersBySupply(t *testing.T) {
	gs := marketSortingTestState()
	gs.SetMarketSellOffer("seller_low", economy.GoodIron, 10)
	gs.SetMarketSellOffer("seller_high", economy.GoodIron, 30)

	got := sortedFactionsForMarket(gs, 1, TradeListSellers, TradeSortDistance)
	want := []faction.FactionID{"seller_high", "seller_low"}
	assertFactionIDs(t, got, want)
}

func TestSortedFactionsForMarketSortsBuyersByDemand(t *testing.T) {
	gs := marketSortingTestState()
	gs.SetMarketBuyOrder("buyer_low", economy.GoodIron, 10)
	gs.SetMarketBuyOrder("buyer_high", economy.GoodIron, 30)

	got := sortedFactionsForMarket(gs, 1, TradeListBuyers, TradeSortDistance)
	want := []faction.FactionID{"buyer_high", "buyer_low"}
	assertFactionIDs(t, got, want)
}

func marketSortingTestState() *state.GameState {
	return &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":      {ID: "player", Gold: 1000, Iron: 100},
			"seller_low":  {ID: "seller_low", Gold: 100, Iron: 10},
			"seller_high": {ID: "seller_high", Gold: 100, Iron: 30},
			"buyer_low":   {ID: "buyer_low", Gold: 1000, Iron: 0},
			"buyer_high":  {ID: "buyer_high", Gold: 1000, Iron: 0},
		},
		MarketOrders: state.MarketOrderBook{
			SellOffers: make(map[faction.FactionID]map[economy.GoodType]int),
			BuyOrders:  make(map[faction.FactionID]map[economy.GoodType]int),
		},
	}
}

func assertFactionIDs(t *testing.T, got, want []faction.FactionID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("devlet sayısı = %d, want %d (%v)", len(got), len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("devlet sırası = %v, want %v", got, want)
		}
	}
}

func TestShortestTradeRegionPathHonorsLandAndSeaType(t *testing.T) {
	regions := map[world.RegionID]*world.Region{
		"land_a":   {ID: "land_a", Neighbors: []world.RegionID{"land_mid", "sea_a"}},
		"land_mid": {ID: "land_mid", Neighbors: []world.RegionID{"land_a", "land_b"}},
		"land_b":   {ID: "land_b", Neighbors: []world.RegionID{"land_mid", "sea_b"}},
		"sea_a":    {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"land_a", "sea_mid"}},
		"sea_mid":  {ID: "sea_mid", IsSea: true, Neighbors: []world.RegionID{"sea_a", "sea_b"}},
		"sea_b":    {ID: "sea_b", IsSea: true, Neighbors: []world.RegionID{"sea_mid", "land_b"}},
	}

	landPath := shortestTradeRegionPath(regions, "land_a", "land_b", false)
	if got, want := landPath, []world.RegionID{"land_a", "land_mid", "land_b"}; !sameRegionIDs(got, want) {
		t.Fatalf("kara ticaret yolu = %v, want %v", got, want)
	}
	seaPath := shortestTradeRegionPath(regions, "sea_a", "sea_b", true)
	if got, want := seaPath, []world.RegionID{"sea_a", "sea_mid", "sea_b"}; !sameRegionIDs(got, want) {
		t.Fatalf("deniz ticaret yolu = %v, want %v", got, want)
	}
}

func TestCurvedTradeSegmentPointsCreateStableArc(t *testing.T) {
	start := tradeOverlayPoint{x: 10, y: 20}
	end := tradeOverlayPoint{x: 110, y: 20}
	points := curvedTradeSegmentPoints(start, end, "sea_segment:sea_a|sea_b")

	if len(points) < 3 {
		t.Fatalf("kıvrımlı deniz segmenti örnek sayısı = %d, want en az 3", len(points))
	}
	if points[0] != start {
		t.Fatalf("segment başlangıcı = %+v, want %+v", points[0], start)
	}
	if got := points[len(points)-1]; got != end {
		t.Fatalf("segment sonu = %+v, want %+v", got, end)
	}
	if points[len(points)/2].y == start.y {
		t.Fatalf("segment orta noktası düz kaldı: %+v", points[len(points)/2])
	}
}

func TestSmoothTradePathSegmentPassesThroughSeaFocusAndRoundsTurn(t *testing.T) {
	points := []tradeOverlayPoint{
		{x: 0, y: 0},
		{x: 50, y: 0},
		{x: 50, y: 50},
		{x: 100, y: 50},
	}
	segment := smoothTradePathSegmentPoints(points, 1)
	if len(segment) < 3 {
		t.Fatalf("yumuşatılmış deniz segmenti örnek sayısı = %d, want en az 3", len(segment))
	}
	if segment[0] != points[1] || segment[len(segment)-1] != points[2] {
		t.Fatalf("deniz odakları korunmadı: start=%+v end=%+v", segment[0], segment[len(segment)-1])
	}
	curved := false
	for _, point := range segment[1 : len(segment)-1] {
		if point.x != points[1].x && point.y != points[1].y && point.x != points[2].x && point.y != points[2].y {
			curved = true
			break
		}
	}
	if !curved {
		t.Fatalf("deniz dönüşü köşeli kaldı: %+v", segment)
	}
}

func TestTradeRouteTypeLabel(t *testing.T) {
	tests := []struct {
		routeType world.TradeRouteType
		want      string
	}{
		{routeType: world.TradeRouteLand, want: "Kara yolu"},
		{routeType: world.TradeRouteSea, want: "Deniz yolu"},
	}
	for _, tt := range tests {
		if got := tradeRouteTypeLabel(tt.routeType); got != tt.want {
			t.Errorf("tradeRouteTypeLabel(%q) = %q, want %q", tt.routeType, got, tt.want)
		}
	}
}

func TestTradeCorridorTooltipTitle(t *testing.T) {
	if got, want := tradeCorridorTooltipTitle(tradeCorridorInfo{dashed: true}), "Ticaret Anlaşması"; got != want {
		t.Fatalf("kesikli rota tooltip başlığı = %q, want %q", got, want)
	}
	if got, want := tradeCorridorTooltipTitle(tradeCorridorInfo{}), "Ticaret Koridoru"; got != want {
		t.Fatalf("normal rota tooltip başlığı = %q, want %q", got, want)
	}
}

func TestTradeRoutePaletteSeparatesLandAndSea(t *testing.T) {
	landGlow, landCore, landArrow := tradeRoutePalette(world.TradeRouteLand)
	seaGlow, seaCore, seaArrow := tradeRoutePalette(world.TradeRouteSea)
	if landGlow == seaGlow || landCore == seaCore || landArrow == seaArrow {
		t.Fatalf("kara ve deniz rota paletleri ayrışmıyor: kara=(%v,%v,%v), deniz=(%v,%v,%v)", landGlow, landCore, landArrow, seaGlow, seaCore, seaArrow)
	}
}

func TestSourceTradeLinkUsesDirectLandGeometryAndDistinctPalette(t *testing.T) {
	gs := &state.GameState{TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
		{ID: "source", OffMap: true, Links: []world.TradeCenterLink{{RegionID: "center", Type: world.TradeRouteSea}}},
		{ID: "center"},
	}}}
	centers := []tradeCenterVisual{{id: "source"}, {id: "center"}}
	types := tradeCenterLinkTypes(gs, centers)
	sourceLinks := tradeCenterSourceLinks(gs, centers)
	if got := types["0|1"]; got != world.TradeRouteLand {
		t.Fatalf("source link geometry = %q, want %q", got, world.TradeRouteLand)
	}
	if !sourceLinks["0|1"] {
		t.Fatal("source link style was not marked")
	}
	landGlow, landCore, landArrow := tradeRoutePalette(world.TradeRouteLand)
	sourceGlow, sourceCore, sourceArrow := tradeSourceRoutePalette()
	if sourceGlow == landGlow || sourceCore == landCore || sourceArrow == landArrow {
		t.Fatalf("source palette should differ from normal land palette: source=(%v,%v,%v), land=(%v,%v,%v)", sourceGlow, sourceCore, sourceArrow, landGlow, landCore, landArrow)
	}
}

func TestTradeCorridorDetailsFilterToCenterFactions(t *testing.T) {
	veniceRoute := &economy.TradeRoute{FromFactionID: "venice", ToFactionID: "alexandria"}
	unrelatedRoute := &economy.TradeRoute{FromFactionID: "saruhan", ToFactionID: "mentese"}
	corridor := tradeCorridorInfo{
		centerFactions: [2]string{"venice", "alexandria"},
		routeDetails: []tradeCorridorRouteDetail{
			{key: "venice->alexandria", route: veniceRoute},
			{key: "saruhan->mentese", route: unrelatedRoute},
			{key: "historical", historical: true},
		},
	}
	details := tradeCorridorDetailsForCenter(corridor)
	if got, want := len(details), 2; got != want {
		t.Fatalf("merkezle ilgili popup detayı = %d, want %d", got, want)
	}
	for _, detail := range details {
		if detail.route == unrelatedRoute {
			t.Fatal("Saruhan-Menteşe rotası merkez popup'ına sızdı")
		}
	}
}

func TestTradeCenterEndpointRegionsSelectsSingleNearestSea(t *testing.T) {
	regions := map[world.RegionID]*world.Region{
		"center":   {ID: "center", WorldX: 100, WorldY: 100, Neighbors: []world.RegionID{"far_sea", "near_sea"}},
		"far_sea":  {ID: "far_sea", IsSea: true, WorldX: 300, WorldY: 100},
		"near_sea": {ID: "near_sea", IsSea: true, WorldX: 110, WorldY: 100},
	}
	center := tradeCenterVisual{id: "center", regionID: "center", worldX: 100, worldY: 100}
	got := tradeCenterEndpointRegions(center, world.TradeRouteSea, regions)
	if len(got) != 1 || got[0] != "near_sea" {
		t.Fatalf("merkez deniz girişi = %v, want [near_sea]", got)
	}
}

func TestTradeCorridorRouteDetailsDeduplicateRoutes(t *testing.T) {
	route := &economy.TradeRoute{FromFactionID: "a", ToFactionID: "b"}
	details := appendTradeCorridorRouteDetail(nil, tradeCorridorRouteDetail{key: route.AssignmentKey()})
	details = appendTradeCorridorRouteDetail(details, tradeCorridorRouteDetail{key: route.AssignmentKey(), good: "Demir"})
	details = appendTradeCorridorRouteDetail(details, tradeCorridorRouteDetail{key: "b->a", good: "Tahıl"})
	if got, want := len(details), 2; got != want {
		t.Fatalf("koridor rota detayı sayısı = %d, want %d", got, want)
	}
}

func TestHistoricalTradeCorridorDetailsGroupGoodsAndSumAmounts(t *testing.T) {
	details := appendTradeCorridorRouteDetail(nil, tradeCorridorRouteDetail{
		key: "historical", good: "Tahıl", amount: 4, historical: true,
	})
	details = appendTradeCorridorRouteDetail(details, tradeCorridorRouteDetail{
		key: "historical", good: "Tahıl", amount: 3, historical: true,
	})
	details = appendTradeCorridorRouteDetail(details, tradeCorridorRouteDetail{
		key: "historical", good: "Baharat", amount: 2, historical: true,
	})
	if got, want := len(details), 1; got != want {
		t.Fatalf("tarihsel grup sayısı = %d, want %d", got, want)
	}
	if got, want := len(details[0].historicalGoods), 2; got != want {
		t.Fatalf("tarihsel mal sayısı = %d, want %d", got, want)
	}
	for _, total := range details[0].historicalGoods {
		if total.good == "Tahıl" && total.amount != 7 {
			t.Fatalf("tahıl toplamı = %d, want 7", total.amount)
		}
	}
}

func TestTradeCorridorDetailsPutHistoricalFirst(t *testing.T) {
	details := []tradeCorridorRouteDetail{
		{key: "normal", historical: false},
		{key: "historical", historical: true},
		{key: "normal-2", historical: false},
	}
	sortTradeCorridorRouteDetails(details)
	if !details[0].historical || details[1].historical || details[2].historical {
		t.Fatalf("tarihsel rota sırası = %+v, tarihsel kayıt en üstte olmalı", details)
	}
}

func TestTradeCorridorTooltipHeightIncludesAllRoutes(t *testing.T) {
	corridor := tradeCorridorInfo{routeDetails: make([]tradeCorridorRouteDetail, 4)}
	if got, want := tradeCorridorTooltipHeight(corridor), 224.0; got != want {
		t.Fatalf("çoklu rota tooltip yüksekliği = %v, want %v", got, want)
	}
}

func TestSplitTradePhysicalPathKeepsSharedSegmentsTogether(t *testing.T) {
	points := []tradeOverlayPoint{{x: 0}, {x: 10}, {x: 20}, {x: 30}, {x: 40}}
	keys := []string{"connector:a|sea:x#0", "connector:a|sea:x#1", "sea:x|sea:y#0", "sea:x|sea:y#1"}
	segments := splitTradePhysicalPath(points, keys)
	if got, want := len(segments), 2; got != want {
		t.Fatalf("fiziksel segment sayısı = %d, want %d", got, want)
	}
	if got, want := segments[0].key, "connector:a|sea:x"; got != want {
		t.Fatalf("ilk fiziksel segment = %q, want %q", got, want)
	}
	if got, want := len(segments[0].points), 3; got != want {
		t.Fatalf("ilk segment nokta sayısı = %d, want %d", got, want)
	}
	if got, want := segments[1].key, "sea:x|sea:y"; got != want {
		t.Fatalf("ikinci fiziksel segment = %q, want %q", got, want)
	}
}

func TestTradeCenterAnchorPrefersFirstPortSettlement(t *testing.T) {
	portRegion := &world.Region{
		ID: "port_region",
		Settlements: []world.Settlement{
			{ID: "first_port", Type: world.SettlementPort},
			{ID: "second_port", Type: world.SettlementPort},
		},
	}
	wm := &WorldMap{
		settlementAnchor: map[settlementAnchorKey][2]int{
			{Region: "port_region", Index: 0}: {100, 200},
			{Region: "port_region", Index: 1}: {300, 400},
		},
		primarySettlement: map[world.RegionID][2]int{
			"port_region": {500, 600},
		},
	}
	r := &Renderer{worldMap: wm, camScale: 1}
	gotX, gotY := r.tradePortScreenPos(portRegion, "")
	wantX, wantY := r.worldToScreen(100, 200)
	if gotX != wantX || gotY != wantY {
		t.Fatalf("ticaret merkezi port anchor = (%v,%v), want (%v,%v)", gotX, gotY, wantX, wantY)
	}

	landRegion := &world.Region{ID: "land_region", Settlements: []world.Settlement{{ID: "center", IsCenter: true}}}
	wm.primarySettlement[landRegion.ID] = [2]int{700, 800}
	gotX, gotY = r.tradePortScreenPos(landRegion, "")
	wantX, wantY = r.worldToScreen(700, 800)
	if gotX != wantX || gotY != wantY {
		t.Fatalf("ticaret merkezi merkez anchor = (%v,%v), want (%v,%v)", gotX, gotY, wantX, wantY)
	}
}

func sameRegionIDs(got, want []world.RegionID) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
