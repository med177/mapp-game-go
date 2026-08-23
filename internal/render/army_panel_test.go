package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestScoutedEnemyRevealCountUsesSeventyFivePercentForSiegeIntel(t *testing.T) {
	if got := scoutedEnemyRevealCount(8, false, 0.75); got != 6 {
		t.Fatalf("8 birim için %%75 görünürlük 6 olmalıydı, got=%d", got)
	}
	if got := scoutedEnemyRevealCount(1, false, 0.75); got != 1 {
		t.Fatalf("tek birim için en az 1 görünürlük korunmalıydı, got=%d", got)
	}
}

func TestArmyPanelUnitIndexGroupsUnitsByCategory(t *testing.T) {
	types := map[string]*army.UnitType{
		"cav":   {ID: "cav", Category: army.CategoryCavalry},
		"siege": {ID: "siege", Category: army.CategorySiege},
		"inf":   {ID: "inf", Category: army.CategoryInfantry},
	}
	units := []army.Unit{
		{TypeID: "cav"},
		{TypeID: "siege"},
		{TypeID: "inf"},
		{TypeID: "inf"},
	}

	want := []int{2, 3, 0, 1}
	for displayIndex, wantIndex := range want {
		if got := armyPanelUnitIndex(units, types, displayIndex); got != wantIndex {
			t.Fatalf("gösterim index'i %d için gerçek birim index'i %d olmalıydı, got=%d", displayIndex, wantIndex, got)
		}
	}
	if got := armyPanelUnitIndex(units, types, len(units)); got != -1 {
		t.Fatalf("dolu birimlerin sonrasında boş slot için -1 bekleniyordu, got=%d", got)
	}
}

func TestDisbandButtonIsLeftOfArmyActions(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() { ScreenWidth, ScreenHeight = oldW, oldH }()
	ScreenWidth, ScreenHeight = 1280, 720
	gs := &state.GameState{
		PlayerFactionID: "player",
		UnitTypes:       map[string]*army.UnitType{"inf": {ID: "inf", Category: army.CategoryInfantry}},
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", Units: []army.Unit{{TypeID: "inf"}, {TypeID: "inf"}}},
		},
	}
	selected := map[int]bool{0: true}
	disband, ok := buildDisbandArmyButton(gs, "army", selected)
	if !ok {
		t.Fatal("seçili birim varken terhis düğmesi oluşturulmalıydı")
	}
	split, ok := buildSplitArmyButton(gs, "army", selected)
	if !ok {
		t.Fatal("iki birimli orduda böl düğmesi oluşturulmalıydı")
	}
	if disband.X+disband.W+float64(actionBtnGap) != split.X {
		t.Fatalf("terhis düğmesi böl düğmesinin hemen solunda olmalıydı: disband=%+v split=%+v", disband, split)
	}
}

func TestArmyUpkeepSummaryShowsGrainAndGoldSalary(t *testing.T) {
	gs := &state.GameState{
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 4, GoldUpkeep: 7},
		},
	}
	a := &army.Army{
		Units:         []army.Unit{{TypeID: "inf"}},
		EmbarkedUnits: []army.Unit{{TypeID: "inf"}},
	}
	grainText, goldText := armyUpkeepSummaryStrings(gs, a)
	if grainText != "Tahıl ihtiyacı: 4 / tur" {
		t.Fatalf("tahıl etiketi hatalı: %q", grainText)
	}
	if goldText != "Asker Maaşı: 14/tur" {
		t.Fatalf("kara + taşınan birlik maaşı birlikte gösterilmeli: %q", goldText)
	}
}

func TestArmyPanelTransportFooterTextShowsLoadAndCapacity(t *testing.T) {
	gs := &state.GameState{
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 8},
		},
	}
	fleet := &army.Army{
		IsNaval:       true,
		Units:         []army.Unit{{TypeID: "transport"}},
		EmbarkedUnits: []army.Unit{{TypeID: "infantry"}, {TypeID: "cavalry"}},
	}
	if got := armyTransportFooterText(gs, fleet); got != "Taşıma: 2/8" {
		t.Fatalf("nakliye doluluk metni yanlış: got=%q", got)
	}
}

func TestArmyPanelTransportFooterTextHiddenWithoutTransportCapacity(t *testing.T) {
	gs := &state.GameState{UnitTypes: map[string]*army.UnitType{}}
	fleet := &army.Army{IsNaval: true, Units: []army.Unit{{TypeID: "warship"}}}
	if got := armyTransportFooterText(gs, fleet); got != "" {
		t.Fatalf("nakliye kapasitesi olmayan filo için metin çizilmemeliydi: got=%q", got)
	}
}

func TestArmyTransportFooterTextUsesInfoButtonGeometry(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	gs := &state.GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 20},
			"infantry":  {ID: "infantry", Category: army.CategoryInfantry},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID:            "fleet",
				OwnerID:       "player",
				IsNaval:       true,
				Units:         []army.Unit{{TypeID: "transport"}},
				EmbarkedUnits: []army.Unit{{TypeID: "infantry"}, {TypeID: "infantry"}, {TypeID: "infantry"}, {TypeID: "infantry"}},
			},
		},
	}

	footer := armyPanelFooterLayoutFor(gs, gs.Armies["fleet"], armyPanelGeometry())
	button, ok := armyTransportInfoButtonFromFooter(footer)
	if !ok {
		t.Fatal("taşıma kapasitesi metni bilgi butonu üretmeliydi")
	}
	if button.Label != "Taşıma: 4/20" {
		t.Fatalf("taşıma butonu etiketi yanlış: got=%q", button.Label)
	}
	textW := float32(MeasureText(button.Label, FaceSmall))
	if button.W < float64(textW+20) {
		t.Fatalf("taşıma butonu yatay padding'i yetersiz: width=%.1f text=%.1f", button.W, textW)
	}
	if got, want := button.X+button.W/2, float64(footer.transportX+textW/2); got != want {
		t.Fatalf("taşıma etiketi yatayda footer metin merkezine oturmalı: got=%.1f want=%.1f", got, want)
	}
	if transportInfoButtonStyle.TextOffsetY != 0 {
		t.Fatal("taşıma butonu ortak dikey merkezleme için sabit offset kullanmamalı")
	}
	if button.Y != float64(footer.footerY+merchantRouteFooterButtonGap) || button.H != float64(merchantRouteFooterButtonH) {
		t.Fatalf("taşıma butonu footer içinde üst-alt margin ile durmalı: y=%.1f h=%.1f", button.Y, button.H)
	}
	if loaded := transportInfoButtonStyleFor(gs.Armies["fleet"]); loaded.BG == transportInfoButtonStyle.BG || loaded.Border == transportInfoButtonStyle.Border {
		t.Fatal("taşınan birlik bulunan filo farklı aktif buton rengi kullanmalı")
	}
	if !armyTransportInfoButtonHit(button.X+button.W/2, button.Y+button.H/2, gs, "fleet") {
		t.Fatal("taşıma bilgi butonu kendi merkezinde tıklanabilir olmalı")
	}
}

func TestArmyPanelMixedFleetFooterColumnsStaySeparated(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	gs := &state.GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
			"warship":       {ID: "warship", Category: army.CategoryNavalWar},
			"transport":     {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 8},
		},
	}
	fleet := &army.Army{
		OwnerID: "player",
		IsNaval: true,
		Units: []army.Unit{
			{TypeID: "merchant_ship"},
			{TypeID: "warship"},
			{TypeID: "transport"},
		},
		EmbarkedUnits: []army.Unit{{TypeID: "infantry"}},
	}

	layout := armyPanelFooterLayoutFor(gs, fleet, armyPanelGeometry())
	if layout.routeButton.X+layout.routeButton.W > layout.missionButton.X {
		t.Fatalf("ROTA ATA ve GÖREV düğmeleri örtüşüyor: route=%+v mission=%+v", layout.routeButton, layout.missionButton)
	}
	if layout.routeStatus.X+layout.routeStatus.W > layout.missionStatus.X {
		t.Fatalf("rota ve görev bilgi alanları örtüşüyor: route=%+v mission=%+v", layout.routeStatus, layout.missionStatus)
	}
	if layout.missionStatus.X+layout.missionStatus.W > float64(layout.bonusX)-12 {
		t.Fatalf("görev durumu gemi bonusu alanına taşıyor: mission=%+v bonusX=%.1f", layout.missionStatus, layout.bonusX)
	}
	if float64(layout.bonusX)+float64(MeasureText(layout.bonusText, FaceSmall)) > float64(layout.powerX)-12 {
		t.Fatalf("gemi bonusu güç alanına taşıyor: bonusX=%.1f powerX=%.1f", layout.bonusX, layout.powerX)
	}
}

func TestArmyPanelReplenishmentBadgeActivatesForDamagedFleetInOwnPort(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"sea":  {ID: "sea", IsSea: true},
			"home": {ID: "home", OwnerID: "p1", Buildings: []string{"port"}},
		},
	}
	fleet := &army.Army{
		OwnerID:        "p1",
		RegionID:       "sea",
		DockedRegionID: "home",
		IsNaval:        true,
		Units:          []army.Unit{{TypeID: "warship", CurrentHP: 60}},
	}

	if !armyCanRenderReplenishment(gs, fleet) {
		t.Fatal("kendi limanındaki hasarlı filonun tamirat göstergesi aktif olmalıydı")
	}
	fleet.DockedRegionID = ""
	if armyCanRenderReplenishment(gs, fleet) {
		t.Fatal("açık denizdeki hasarlı filonun tamirat göstergesi görünmemeliydi")
	}
	fleet.DockedRegionID = "home"
	fleet.Units[0].CurrentHP = army.MaxUnitHP
	fleet.EmbarkedUnits = []army.Unit{{TypeID: "inf", CurrentHP: 60}}
	if !armyCanRenderReplenishment(gs, fleet) {
		t.Fatal("yalnız taşınan kara birimi hasarlıysa dock edilmiş filonun tamirat göstergesi aktif olmalıydı")
	}
}

func TestArmyPanelReplenishmentBadgeActivatesInVassalRegion(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"vassal": {ID: "vassal", OverlordID: "player"},
		},
		Regions: map[world.RegionID]*world.Region{
			"vassal_land": {ID: "vassal_land", OwnerID: "vassal"},
		},
	}
	armyRef := &army.Army{
		OwnerID:  "player",
		RegionID: "vassal_land",
		Units:    []army.Unit{{TypeID: "inf", CurrentHP: 60}},
	}

	if !armyCanRenderReplenishment(gs, armyRef) {
		t.Fatal("vassal bölgesindeki hasarlı kara ordusunun toparlanma rozeti aktif olmalıydı")
	}
}

func TestArmyPanelCapitalReplenishmentUsesSecondBadgeOnTheLeft(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_city"},
			"vassal": {ID: "vassal", CapitalSettlementID: "vassal_city", OverlordID: "player"},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {
				ID:          "capital",
				OwnerID:     "player",
				Settlements: []world.Settlement{{ID: "capital_city", IsCenter: true}},
			},
			"field": {ID: "field", OwnerID: "player"},
			"vassal_capital": {
				ID:          "vassal_capital",
				OwnerID:     "vassal",
				Settlements: []world.Settlement{{ID: "vassal_city", IsCenter: true}},
			},
		},
	}
	capitalArmy := &army.Army{OwnerID: "player", RegionID: "capital"}
	normalArmy := &army.Army{OwnerID: "player", RegionID: "field"}
	playerVassalArmy := &army.Army{OwnerID: "player", RegionID: "vassal_capital"}
	fleet := &army.Army{OwnerID: "player", IsNaval: true, DockedRegionID: "capital"}

	if !armyHasCapitalReplenishmentBonus(gs, capitalArmy) {
		t.Fatal("başkentteki kara ordusu çift toparlanma rozeti kullanmalıydı")
	}
	if armyHasCapitalReplenishmentBonus(gs, normalArmy) {
		t.Fatal("normal bölgedeki ordu ikinci toparlanma rozetini kullanmamalıydı")
	}
	if armyHasCapitalReplenishmentBonus(gs, playerVassalArmy) {
		t.Fatal("oyuncu ordusu vassal başkentinde ikinci toparlanma rozetini kullanmamalıydı")
	}
	if armyHasCapitalReplenishmentBonus(gs, fleet) {
		t.Fatal("donanma başkentte olsa da ikinci kara toparlanma rozetini kullanmamalıydı")
	}

	right := armyReplenishmentBadgeRect(100, 200, 0)
	left := armyReplenishmentBadgeRect(100, 200, 1)
	if left.X+left.W+float64(armyReplenishmentBadgeGap) != right.X {
		t.Fatalf("ikinci rozet mevcut rozetin solunda sabit boşlukla durmalı: left=%+v right=%+v", left, right)
	}
	if left.Y != right.Y || left.W != right.W || left.H != right.H {
		t.Fatalf("iki toparlanma rozeti aynı ölçü ve dikey anchor'ı paylaşmalı: left=%+v right=%+v", left, right)
	}
}

func TestSplitSelectionRequiresOneUnitToRemain(t *testing.T) {
	a := &army.Army{Units: []army.Unit{{TypeID: "inf"}, {TypeID: "cav"}, {TypeID: "siege"}}}

	if !splitSelectionCanBeApplied(a, nil) {
		t.Fatal("seçim yokken varsayılan bölme kullanılabilmeliydi")
	}
	if !splitSelectionCanBeApplied(a, map[int]bool{0: true, 2: true}) {
		t.Fatal("ana orduda bir birlik kaldığında seçili bölme yapılabilmeliydi")
	}
	if splitSelectionCanBeApplied(a, map[int]bool{0: true, 1: true, 2: true}) {
		t.Fatal("tüm birlikler seçiliyken bölme yapılmamalıydı")
	}
}

func TestArmyPanelUnitHoverIDUsesDisplayedCardOrder(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		UnitTypes: map[string]*army.UnitType{
			"cav":   {ID: "cav", Category: army.CategoryCavalry},
			"siege": {ID: "siege", Category: army.CategorySiege},
			"inf":   {ID: "inf", Category: army.CategoryInfantry},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {
				ID:      "a1",
				OwnerID: "osm",
				Units:   []army.Unit{{TypeID: "cav"}, {TypeID: "siege"}, {TypeID: "inf"}},
			},
		},
	}

	layout := armyPanelGeometry()
	cardCenter := func(displayIndex int) (float64, float64) {
		cardX, cardY := armyPanelUnitPosition(layout, displayIndex)
		cx := cardX + cardW/2
		cy := cardY + cardH/2
		return float64(cx), float64(cy)
	}

	for displayIndex, want := range []string{"inf", "cav", "siege"} {
		mx, my := cardCenter(displayIndex)
		if got := ArmyPanelUnitHoverID(mx, my, gs, "a1"); got != want {
			t.Fatalf("görünen kart %d için %q bekleniyordu, got=%q", displayIndex, want, got)
		}
	}
	emptyX, emptyY := cardCenter(3)
	if got := ArmyPanelUnitHoverID(emptyX, emptyY, gs, "a1"); got != "" {
		t.Fatalf("boş kart hover'ı birim döndürmemeliydi, got=%q", got)
	}

	gs.Armies["a1"].OwnerID = "enemy"
	mx, my := cardCenter(0)
	if got := ArmyPanelUnitHoverID(mx, my, gs, "a1"); got != "" {
		t.Fatalf("düşman ordusunda birim detayı açılmamalıydı, got=%q", got)
	}

	gs.Factions = map[faction.FactionID]*faction.Faction{
		"osm":    {ID: "osm"},
		"vassal": {ID: "vassal", OverlordID: "osm"},
	}
	gs.Armies["a1"].OwnerID = "vassal"
	if got := ArmyPanelUnitHoverID(mx, my, gs, "a1"); got != "inf" {
		t.Fatalf("oyuncu vassalının birim detayı görünür olmalıydı, got=%q", got)
	}
}

func TestPlayerCanSeeArmyDetailsIncludesVassalChain(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm":       {ID: "osm"},
			"vassal":    {ID: "vassal", OverlordID: "osm"},
			"subvassal": {ID: "subvassal", OverlordID: "vassal"},
			"enemy":     {ID: "enemy"},
		},
	}

	for _, owner := range []string{"osm", "vassal", "subvassal"} {
		if !playerCanSeeArmyDetails(gs, &army.Army{OwnerID: owner}) {
			t.Fatalf("%s ordusu tam istihbarat kapsamında olmalıydı", owner)
		}
	}
	if playerCanSeeArmyDetails(gs, &army.Army{OwnerID: "enemy"}) {
		t.Fatal("oyuncu dışı düşman ordusu vassal görünürlüğü kapsamına girmemeliydi")
	}
}

func TestArmyPanelUnitHoverReturnsCurrentHPAndTypeCount(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry},
			"cav": {ID: "cav", Category: army.CategoryCavalry},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {
				ID:      "a1",
				OwnerID: "osm",
				Units: []army.Unit{
					{TypeID: "inf", CurrentHP: 73},
					{TypeID: "inf", CurrentHP: 41},
					{TypeID: "cav", CurrentHP: 88},
				},
			},
		},
	}

	layout := armyPanelGeometry()
	mx := float64(layout.gridX + cardW/2)
	my := float64(layout.gridY + cardH/2)
	unit, count, ok := armyPanelUnitHover(mx, my, gs, "a1")
	if !ok {
		t.Fatal("ordu birim kartı hover verisi bekleniyordu")
	}
	if unit.TypeID != "inf" || unit.CurrentHP != 73 {
		t.Fatalf("kartın gerçek birim örneği dönmeliydi: unit=%+v", unit)
	}
	if count != 2 {
		t.Fatalf("aynı tipin ordu içindeki adedi 2 olmalıydı, got=%d", count)
	}
}

func TestArmyPanelTooltipOnlyActiveOnTopLayer(t *testing.T) {
	gs := &state.GameState{Phase: state.PhasePlayerTurn}
	r := &Renderer{gs: gs, SelectedArmy: "army"}
	if !r.armyPanelTooltipActive() {
		t.Fatal("başka bir üst panel yokken ordu popup'ı aktif olmalıydı")
	}

	r.showDiplomacy = true
	if r.armyPanelTooltipActive() {
		t.Fatal("diplomasi paneli açıkken arkadaki ordu popup'ı görünmemeliydi")
	}
	r.showDiplomacy = false
	r.showMerchantRoutePanel = true
	if r.armyPanelTooltipActive() {
		t.Fatal("merchant route paneli açıkken arkadaki ordu popup'ı görünmemeliydi")
	}
	r.showMerchantRoutePanel = false
	r.showActiveWars = true
	if !r.armyPanelTooltipActive() {
		t.Fatal("aktif savaşlar paneli açıkken harita pasifleşmediği için ordu popup'ı görünmeliydi")
	}
	r.showMerchantRoutePanel = true
	if r.armyPanelTooltipActive() {
		t.Fatal("aktif savaşlar paneliyle birlikte merchant route paneli açıkken arkadaki ordu popup'ı görünmemeliydi")
	}
}

func TestCommanderPortraitHitRectStaysInsideCommanderColumn(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "osm", RegionID: "ankara", Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}},
		},
	}

	rect, ok := commanderPortraitHitRect(gs, "a1")
	if !ok {
		t.Fatal("commander portrait rect bekleniyordu")
	}
	layout := armyPanelGeometry()
	if rect.X < float64(layout.commanderX) || rect.X+rect.W > float64(layout.commanderX+layout.commanderW) {
		t.Fatalf("portrait rect commander kolonunu tasiyor: rect=%+v layout=%+v", rect, layout)
	}
	if rect.Y < float64(layout.commanderY) || rect.Y+rect.H > float64(layout.commanderY+layout.commanderH) {
		t.Fatalf("portrait rect commander kartini tasiyor: rect=%+v layout=%+v", rect, layout)
	}
}

func TestArmyPanelCommanderUnassignButtonUsesPrimaryCommanderAndOwnArmy(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Armies: map[army.ArmyID]*army.Army{
			"owned": {
				ID: "owned", OwnerID: "osm", Commander: &army.Commander{ID: "cmd_1"},
			},
			"empty": {ID: "empty", OwnerID: "osm"},
			"enemy": {
				ID: "enemy", OwnerID: "byz", Commander: &army.Commander{ID: "cmd_2"},
			},
		},
	}

	button, ok := buildArmyPanelCommanderUnassignButton(gs, "owned")
	if !ok {
		t.Fatal("kendi ana komutanı için ayırma düğmesi oluşturulmalıydı")
	}
	if !button.HitTest(button.X+button.W/2, button.Y+button.H/2) ||
		!ArmyPanelCommanderUnassignHitTest(button.X+button.W/2, button.Y+button.H/2, gs, "owned") {
		t.Fatal("ordu panelindeki Komutanı Ayır düğmesi kendi rect'inde hit olmalıydı")
	}
	if _, ok := buildArmyPanelCommanderUnassignButton(gs, "empty"); ok {
		t.Fatal("komutansız ordu için ayırma düğmesi çizilmemeliydi")
	}
	if _, ok := buildArmyPanelCommanderUnassignButton(gs, "enemy"); ok {
		t.Fatal("düşman ordusu için ayırma düğmesi çizilmemeliydi")
	}
}

func TestScoutedEnemyRevealCountUsesFullIntelWhenEnabled(t *testing.T) {
	if got := scoutedEnemyRevealCount(8, true, 0.75); got != 8 {
		t.Fatalf("tam istihbaratta tüm birimler görünmeliydi, got=%d", got)
	}
}

func TestArmyPanelStrengthUsesCurrentHPForPlayerArmy(t *testing.T) {
	types := map[string]*army.UnitType{
		"inf": {ID: "inf", Attack: 20, Defense: 10, Morale: 20},
	}
	units := []army.Unit{{TypeID: "inf", CurrentHP: 50, Experience: 20}}

	attack, defense := armyPanelStrength(units, types, true)
	if attack != 11 || defense != 5 {
		t.Fatalf("oyuncu ordusu gücü mevcut HP/XP ile hesaplanmalıydı: attack=%d defense=%d", attack, defense)
	}
}

func TestScoutedEnemyArmyStrengthOnlyUsesVisibleUnitTypes(t *testing.T) {
	types := map[string]*army.UnitType{
		"cav": {ID: "cav", Category: army.CategoryCavalry, Attack: 20, Defense: 15},
		"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 10, Defense: 8},
	}
	armyRef := &army.Army{Units: []army.Unit{
		{TypeID: "cav", CurrentHP: 25},
		{TypeID: "inf", CurrentHP: 25},
		{TypeID: "inf", CurrentHP: 25},
		{TypeID: "cav", CurrentHP: 25},
	}}
	gs := &state.GameState{UnitTypes: types}

	attack, defense := scoutedEnemyArmyStrength(gs, armyRef, false, 0.5)
	if attack != 20 || defense != 16 {
		t.Fatalf("düşman gücü gizli birimleri içermemeliydi: attack=%d defense=%d", attack, defense)
	}
}

func TestArmyPanelBoundsHitCoversWholePanel(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {
				ID:       "a1",
				OwnerID:  "osm",
				RegionID: "ankara",
				Units: []army.Unit{
					{TypeID: "infantry", CurrentHP: 100},
				},
			},
		},
	}

	rect, ok := armyDetailPanelRect(gs, "a1")
	if !ok {
		t.Fatal("ordu panel rect'i bekleniyordu")
	}
	if !ArmyPanelBoundsHit(rect.X+rect.W/2, rect.Y+rect.H/2, gs, "a1") {
		t.Fatalf("panel merkezi hit olmalıydı: rect=%+v", rect)
	}
	if ArmyPanelBoundsHit(rect.X-4, rect.Y-4, gs, "a1") {
		t.Fatalf("panel dışı hit olmamalıydı: rect=%+v", rect)
	}
}

func TestArmyPanelInteractiveHitIgnoresEmptyPanelArea(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {
				ID:       "a1",
				OwnerID:  "osm",
				RegionID: "ankara",
				Units:    []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
			},
		},
	}

	layout := armyPanelGeometry()
	fx := float64(layout.gridX + cardW*4)
	fy := float64(layout.gridY + cardH + 20)
	if !ArmyPanelBoundsHit(fx, fy, gs, "a1") {
		t.Fatal("secilen nokta panel bounds icinde olmaliydi")
	}
	if ArmyPanelInteractiveHit(fx, fy, gs, "a1") {
		t.Fatalf("bos panel alani interactive sayilmamali: x=%.1f y=%.1f", fx, fy)
	}
}

func TestEnemyArmyPanelBoundsHitUsesIntelAwareRects(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
			"byz": {ID: "byz", NameTR: "Bizans"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm", Neighbors: []world.RegionID{"bursa"}},
			"bursa":  {ID: "bursa", NameTR: "Bursa", OwnerID: "byz", Neighbors: []world.RegionID{"ankara"}},
			"izmir":  {ID: "izmir", NameTR: "İzmir", OwnerID: "byz"},
		},
		Relations: map[string]*faction.Relation{},
		Armies: map[army.ArmyID]*army.Army{
			"player": {ID: "player", OwnerID: "osm", RegionID: "ankara", MovePoints: 2, Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"near":   {ID: "near", OwnerID: "byz", RegionID: "bursa", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"far":    {ID: "far", OwnerID: "byz", RegionID: "izmir", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
	}

	nearRect, ok := armyDetailPanelRect(gs, "near")
	if !ok {
		t.Fatal("yakın düşman için rect bekleniyordu")
	}
	farRect, ok := armyDetailPanelRect(gs, "far")
	if !ok {
		t.Fatal("uzak düşman için rect bekleniyordu")
	}
	if nearRect != farRect {
		t.Fatalf("düşman panelleri artık aynı ortak geometriyi paylaşmalı: near=%+v far=%+v", nearRect, farRect)
	}
	if !ArmyPanelBoundsHit(nearRect.X+nearRect.W/2, nearRect.Y+nearRect.H/2, gs, "near") {
		t.Fatalf("yakın düşman panel merkezi hit olmalıydı: rect=%+v", nearRect)
	}
	if !ArmyPanelBoundsHit(farRect.X+farRect.W/2, farRect.Y+farRect.H/2, gs, "far") {
		t.Fatalf("uzak düşman panel merkezi hit olmalıydı: rect=%+v", farRect)
	}
}
