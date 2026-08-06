package render

import (
	"image/color"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestSortedFactionsSkipsPlayerAndEliminated(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"a":      {ID: "a"},
			"b":      {ID: "b", IsEliminated: true},
			"c":      {ID: "c"},
		},
	}

	got := sortedFactions(gs)
	if len(got) != 2 {
		t.Fatalf("beklenen 2 aktif fraksiyon, got=%d (%v)", len(got), got)
	}
	if got[0] != "a" || got[1] != "c" {
		t.Fatalf("beklenen [a c], got=%v", got)
	}
}

func TestDiplomacyFactionSortsByPowerRanking(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"strong": {ID: "strong"},
			"middle": {ID: "middle"},
			"weak":   {ID: "weak"},
			"dead":   {ID: "dead", IsEliminated: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"strong-army": {ID: "strong-army", OwnerID: "strong", Units: []army.Unit{{}, {}, {}}},
			"middle-army": {ID: "middle-army", OwnerID: "middle", Units: []army.Unit{{}, {}}},
			"weak-army":   {ID: "weak-army", OwnerID: "weak", Units: []army.Unit{{}}},
			"dead-army":   {ID: "dead-army", OwnerID: "dead", Units: []army.Unit{{}, {}, {}, {}}},
		},
	}

	want := []faction.FactionID{"strong", "middle", "weak"}
	got := sortedDiplomacyFactions(gs, diplomacyListSortPowerRanking)
	if len(got) != len(want) {
		t.Fatalf("aktif hedef sayısı yanlış: got=%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("güç sıralaması yanlış: got=%v want=%v", got, want)
		}
	}

	power, rank, count := factionMilitaryPowerStanding(gs, "middle")
	if power != 20 || rank != 2 || count != 4 {
		t.Fatalf("liste metriği yanlış: power=%d rank=%d count=%d", power, rank, count)
	}
}

func TestDiplomacyFactionRelationSortPrefersAdjacentFactionOnScoreTie(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":   {ID: "player"},
			"remote":   {ID: "remote"},
			"border":   {ID: "border"},
			"friendly": {ID: "friendly"},
		},
		Regions: map[world.RegionID]*world.Region{
			"player-region": {ID: "player-region", OwnerID: "player", Neighbors: []world.RegionID{"border-region"}},
			"border-region": {ID: "border-region", OwnerID: "border", Neighbors: []world.RegionID{"player-region"}},
			"remote-region": {ID: "remote-region", OwnerID: "remote"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "remote"): {
				FactionA: "player",
				FactionB: "remote",
				Score:    30,
			},
			faction.RelationKey("player", "border"): {
				FactionA: "player",
				FactionB: "border",
				Score:    30,
			},
			faction.RelationKey("player", "friendly"): {
				FactionA: "player",
				FactionB: "friendly",
				Score:    60,
			},
		},
	}

	got := sortedDiplomacyFactions(gs, diplomacyListSortRelation)
	want := []faction.FactionID{"friendly", "border", "remote"}
	if len(got) != len(want) {
		t.Fatalf("ilişki sıralamasında hedef sayısı yanlış: got=%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ilişki sıralaması yanlış: got=%v want=%v", got, want)
		}
	}
}

func TestDiplomacyListSortButtonsUpdateRendererState(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"a":      {ID: "a"},
			"b":      {ID: "b"},
		},
	}
	r := &Renderer{gs: gs, showDiplomacy: true, diplomacyFocus: 1, diplomacyScroll: 1}
	buttons := buildDiplomacyListSortButtons(diplomacyListLayoutForScreen())
	input := gameui.InputState{
		MouseX:          buttons[1].Button.X + 4,
		MouseY:          buttons[1].Button.Y + 4,
		LeftJustPressed: true,
	}

	r.handleDiplomacyInput(input)
	if r.diplomacyListSort != diplomacyListSortRelation {
		t.Fatalf("ilişki tuşu sıralamayı değiştirmedi: got=%d", r.diplomacyListSort)
	}
	if buttons[1].Button.Label != "İlişki" {
		t.Fatalf("ikinci sıralama düğmesi İlişki olarak adlandırılmalı: got=%q", buttons[1].Button.Label)
	}
	if len(buttons) != 4 || buttons[3].Button.Label != "Ekonomik Sıralama" {
		t.Fatalf("ekonomik sıralama düğmesi eksik: buttons=%d label=%q", len(buttons), buttons[3].Button.Label)
	}
	if r.diplomacyFocus != 0 || r.diplomacyScroll != 0 {
		t.Fatalf("sıralama değişince liste odağı sıfırlanmalı: focus=%d scroll=%d", r.diplomacyFocus, r.diplomacyScroll)
	}
}

func TestDiplomacyEconomicSortUsesIncomeThenTreasury(t *testing.T) {
	gs := &state.GameState{
		Month:           6,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":          {ID: "player"},
			"income_leader":   {ID: "income_leader", Gold: 10},
			"treasury_leader": {ID: "treasury_leader", Gold: 999},
			"treasury_second": {ID: "treasury_second", Gold: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"income_leader_home":   {ID: "income_leader_home", OwnerID: "income_leader", BaseGoldIncome: 300, TaxRate: 100, Satisfaction: 50},
			"treasury_leader_home": {ID: "treasury_leader_home", OwnerID: "treasury_leader", BaseGoldIncome: 200, TaxRate: 100, Satisfaction: 50},
			"treasury_second_home": {ID: "treasury_second_home", OwnerID: "treasury_second", BaseGoldIncome: 200, TaxRate: 100, Satisfaction: 50},
		},
	}

	got := sortedDiplomacyFactions(gs, diplomacyListSortEconomicRanking)
	want := []faction.FactionID{"income_leader", "treasury_leader", "treasury_second"}
	if len(got) != len(want) {
		t.Fatalf("ekonomik sıralama hedef sayısı yanlış: got=%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ekonomik sıralama yanlış: got=%v want=%v", got, want)
		}
	}
	if got := factionTreasuryLabel(gs, "income_leader"); got != "300/10" {
		t.Fatalf("hazine etiketi gelir/altın biçiminde olmalı: got=%q", got)
	}
}

func TestDiplomacyListColumnRectsReserveSquareFactionFlag(t *testing.T) {
	row := gameui.Rect{X: 100, Y: 40, W: 728, H: diplomRowH - 10}

	nameRect, relationRect := diplomacyListColumnRects(row)
	contentX := row.X + 18
	wantNameX := contentX + diplomFactionFlagSize + diplomFactionFlagGap
	contentW := row.W - 36
	wantNameW := contentW - diplomColumnGap - 380
	if wantNameW > diplomNameColumnW {
		wantNameW = diplomNameColumnW
	}
	wantRelationX := contentX + wantNameW + diplomColumnGap

	if nameRect.X != wantNameX {
		t.Fatalf("devlet adı bayrak sonrasından başlamalı: got=%.1f want=%.1f", nameRect.X, wantNameX)
	}
	if relationRect.X != wantRelationX {
		t.Fatalf("ilişki kolonu bayrak eklenince kaymamalı: got=%.1f want=%.1f", relationRect.X, wantRelationX)
	}
	if nameRect.W != wantNameW-diplomFactionFlagSize-diplomFactionFlagGap {
		t.Fatalf("devlet adı genişliği bayrak alanını düşmeli: got=%.1f", nameRect.W)
	}
}

func TestDiplomacyRelationCategoriesUseStanceTradeRoutesAndRealm(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":  {ID: "player"},
			"subject": {ID: "subject"},
			"enemy":   {ID: "enemy"},
			"ally":    {ID: "ally"},
			"vassal":  {ID: "vassal", OverlordID: "subject"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("subject", "enemy"):  {FactionA: "subject", FactionB: "enemy", Stance: faction.StanceWar},
			faction.RelationKey("subject", "ally"):   {FactionA: "subject", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("subject", "vassal"): {FactionA: "subject", FactionB: "vassal", Stance: faction.StanceAllied},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "subject", ToFactionID: "ally"},
		},
	}
	factions := sortedFactions(gs)

	if got := diplomacyRelationCategoryCount(gs, "subject", factions, diplomacyRelationWar); got != 1 {
		t.Fatalf("savaş listesi yanlış: got=%d want=1", got)
	}
	if got := diplomacyRelationCategoryCount(gs, "subject", factions, diplomacyRelationAlliance); got != 1 {
		t.Fatalf("realm içi vassal gerçek ittifaka karışmamalı: got=%d want=1", got)
	}
	if got := diplomacyRelationCategoryCount(gs, "subject", factions, diplomacyRelationTrade); got != 1 {
		t.Fatalf("ticaret listesi aktif rotadan okunmalı: got=%d want=1", got)
	}
	if got := directVassalCount(gs, "subject"); got != 1 {
		t.Fatalf("bağlı devlet sayısı yanlış: got=%d want=1", got)
	}
}

func TestHandleDiplomacyInputScrollsOnPanelWheel(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
	}
	for i := 0; i < 12; i++ {
		id := faction.FactionID("f" + itoa(i))
		gs.Factions[id] = &faction.Faction{ID: id, NameTR: "Devlet " + itoa(i)}
	}

	r := &Renderer{
		gs:                   gs,
		showDiplomacy:        true,
		diplomacyScroll:      0,
		diplomacyFocus:       0,
		diplomacyActionFocus: 0,
	}

	layout := diplomacyListLayoutForScreen()
	input := gameui.InputState{
		MouseX: layout.panelRect.X + 20,
		MouseY: layout.panelRect.Y + 20,
		WheelY: -1,
	}

	r.handleDiplomacyInput(input)
	if r.diplomacyScroll != 1 {
		t.Fatalf("wheel aşağı kaydırınca scroll 1 olmalı, got=%d", r.diplomacyScroll)
	}

	input.WheelY = -1
	r.handleDiplomacyInput(input)
	if r.diplomacyScroll != 2 {
		t.Fatalf("ardışık wheel aşağı kaydırmada scroll 2 olmalı, got=%d", r.diplomacyScroll)
	}

	input.WheelY = 1
	r.handleDiplomacyInput(input)
	if r.diplomacyScroll != 1 {
		t.Fatalf("wheel yukarı kaydırınca scroll 1 olmalı, got=%d", r.diplomacyScroll)
	}
}

func TestDiplomacyActionDisabledReasonAllowsPeaceDuringWar(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player",
				FactionB: "enemy",
				Stance:   faction.StanceWar,
				Score:    -80,
			},
		},
	}

	if reason := diplomacyActionDisabledReason(gs, "enemy", ActionProposePeace); reason != "" {
		t.Fatalf("savaş halindeki hedef için barış aksiyonu disabled olmamalı, got=%q", reason)
	}
}

func TestHandleDiplomacyInputBlocksInvalidPeaceOffer(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ally":   {ID: "ally", NameTR: "Müttefik"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"): {
				FactionA: "player",
				FactionB: "ally",
				Stance:   faction.StancePeace,
				Score:    15,
			},
		},
	}
	r := &Renderer{
		gs:                     gs,
		showDiplomacy:          true,
		diplomacyTargetFaction: "ally",
		diplomacyActionFocus:   1,
		prevKeys:               make(map[ebiten.Key]bool),
		prevMouse:              make(map[ebiten.MouseButton]bool),
	}

	input := gameui.InputState{
		MouseX:          0,
		MouseY:          0,
		LeftJustPressed: false,
	}
	act := r.handleDiplomacyInput(input)
	if act.Kind != ActionNone {
		t.Fatalf("invalid barış teklifinde aksiyon üretilmemeli, got=%s", act.Kind)
	}

	sendBtn := buildDiplomacySendButton()
	input.MouseX = sendBtn.X + 5
	input.MouseY = sendBtn.Y + 5
	input.LeftJustPressed = true

	act = r.handleDiplomacyInput(input)
	if act.Kind != ActionNone {
		t.Fatalf("barış dışı stance'ta gönderim bloklanmalı, got=%s", act.Kind)
	}
	if !r.showDiplomacy {
		t.Fatal("invalid gönderimde diplomasi paneli açık kalmalı")
	}
	if r.diplomacyTargetFaction != "ally" {
		t.Fatalf("invalid gönderimde hedef korunmalı, got=%q", r.diplomacyTargetFaction)
	}
	if r.combatLog != "Barış teklifi sadece savaşta yapılır." {
		t.Fatalf("beklenen uyarı combatLog'a yazılmalı, got=%q", r.combatLog)
	}
}

func TestHandleDiplomacyInputVassalManagementAndDisabledActions(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 100},
			"vassal": {ID: "vassal", NameTR: "Bağlı Devlet", OverlordID: "player"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "vassal"): {FactionA: "player", FactionB: "vassal", Stance: faction.StanceAllied, Score: 40},
		},
	}
	r := &Renderer{
		gs:                     gs,
		showDiplomacy:          true,
		diplomacyTargetFaction: "vassal",
		diplomacyActionFocus:   4,
	}

	warButton := buildDiplomacyActionButtons(gs, "vassal")[0].Button
	r.handleDiplomacyInput(gameui.InputState{MouseX: warButton.X + 1, MouseY: warButton.Y + 1, LeftJustPressed: true})
	if r.diplomacyActionFocus != 4 {
		t.Fatalf("pasif savaş düğmesi seçimi değiştirmemeli, got=%d", r.diplomacyActionFocus)
	}

	management := buildDiplomacyVassalManagementLayout()
	r.handleDiplomacyInput(gameui.InputState{MouseX: management.releaseButton.X + 1, MouseY: management.releaseButton.Y + 1, LeftJustPressed: true})
	if !r.confirmDialog.show {
		t.Fatal("vasallığı bitirme düğmesi onay penceresi açmalı")
	}
	if r.confirmDialog.pendingAction.Kind != ActionReleaseVassal || r.confirmDialog.pendingAction.TargetFaction != "vassal" {
		t.Fatalf("yanlış vasallık yönetim aksiyonu: %+v", r.confirmDialog.pendingAction)
	}
}

func TestDiplomacyActionSelectionUsesHighlightedBorder(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player",
				FactionB: "enemy",
				Stance:   faction.StancePeace,
			},
		},
	}
	r := &Renderer{
		gs:                     gs,
		showDiplomacy:          true,
		diplomacyTargetFaction: "enemy",
		diplomacyActionFocus:   2,
	}

	buttons := buildDiplomacyActionButtons(gs, "enemy")
	warButton := buttons[0].Button
	r.handleDiplomacyInput(gameui.InputState{
		MouseX:          warButton.X + 1,
		MouseY:          warButton.Y + 1,
		LeftJustPressed: true,
	})
	if r.diplomacyActionFocus != 0 {
		t.Fatalf("tıklanan teklif türü seçili duruma geçmeli, got=%d", r.diplomacyActionFocus)
	}

	selected := diplomacyActionButtonStyle(color.RGBA{50, 120, 180, 220}, false, true)
	ordinary := diplomacyActionButtonStyle(color.RGBA{50, 120, 180, 220}, false, false)
	if selected.Border != (color.RGBA{242, 198, 82, 255}) {
		t.Fatalf("seçili teklif border'ı sarı olmalı, got=%v", selected.Border)
	}
	if selected.BorderWidth <= ordinary.BorderWidth {
		t.Fatalf("seçili teklif border'ı belirginleşmeli: selected=%v ordinary=%v", selected.BorderWidth, ordinary.BorderWidth)
	}
	if disabled := diplomacyActionButtonStyle(color.RGBA{50, 120, 180, 220}, true, true); disabled.Border == (color.RGBA{242, 198, 82, 255}) {
		t.Fatal("pasif teklif seçili sarı border stilini kullanmamalı")
	}
}

func TestDiplomacyEstablishedAgreementsBecomeCancellationActions(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ally":   {ID: "ally", NameTR: "Müttefik"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"): {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied, Score: 60},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "player", ToFactionID: "ally"},
			{FromFactionID: "ally", ToFactionID: "player"},
		},
	}

	if got := diplomacyActionForTarget(gs, "ally", 2); got != ActionCancelAlliance {
		t.Fatalf("kurulu ittifak iptal aksiyonuna dönüşmeli, got=%s", got)
	}
	if got := diplomacyActionForTarget(gs, "ally", 3); got != ActionCancelTrade {
		t.Fatalf("kurulu ticaret iptal aksiyonuna dönüşmeli, got=%s", got)
	}
	buttons := buildDiplomacyActionButtons(gs, "ally")
	if buttons[2].Button.Label != "İttifakı Bitir" || buttons[3].Button.Label != "Ticareti Bitir" {
		t.Fatalf("iptal düğmesi etiketleri yanlış: alliance=%q trade=%q", buttons[2].Button.Label, buttons[3].Button.Label)
	}

	r := &Renderer{gs: gs, showDiplomacy: true, diplomacyTargetFaction: "ally"}
	r.handleDiplomacyInput(gameui.InputState{MouseX: buttons[2].Button.X + 1, MouseY: buttons[2].Button.Y + 1, LeftJustPressed: true})
	send := buildDiplomacySendButtonForAction(ActionCancelAlliance)
	action := r.handleDiplomacyInput(gameui.InputState{MouseX: send.X + 1, MouseY: send.Y + 1, LeftJustPressed: true})
	if action.Kind != ActionCancelAlliance || action.TargetFaction != "ally" {
		t.Fatalf("ittifak iptal input aksiyonu yanlış: %+v", action)
	}
}

func TestDiplomacyTradeChanceUsesRealAcceptanceRules(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"ally":   {ID: "ally", NameTR: "Müttefik"},
		},
		Regions: map[world.RegionID]*world.Region{
			"p1": {ID: "p1", OwnerID: "player", TradeCapacity: 4},
			"a1": {ID: "a1", OwnerID: "ally", TradeCapacity: 4},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"): {
				FactionA: "player",
				FactionB: "ally",
				Stance:   faction.StancePeace,
				Score:    5,
			},
		},
	}

	chance, status := estimateDiplomacyChance(gs, "ally", ActionProposeTrade)
	if chance != 0 {
		t.Fatalf("düşük skorda ticaret şansı 0 gösterilmeli, got=%d", chance)
	}
	if status != "Ticaret için ilişki puanı 15 altı" {
		t.Fatalf("beklenen gerçek engel metni, got=%q", status)
	}
}

func TestDiplomacyPeaceChanceUsesRealAcceptanceRules(t *testing.T) {
	gs := &state.GameState{
		ScenarioID:      "1300_ottoman_rise",
		Turn:            1,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"enemy":  {ID: "enemy", NameTR: "Düşman", Gold: 0, Grain: 80},
		},
		Regions: map[world.RegionID]*world.Region{
			"p1": {ID: "p1", OwnerID: "player"},
			"e1": {ID: "e1", OwnerID: "enemy"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player_army": {ID: "player_army", OwnerID: "player", RegionID: "p1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"enemy_army":  {ID: "enemy_army", OwnerID: "enemy", RegionID: "e1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player",
				FactionB: "enemy",
				Stance:   faction.StanceWar,
				Score:    -44,
			},
		},
	}
	for i := 0; i < 11; i++ {
		id := world.RegionID("p_extra_" + itoa(i))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "player"}
	}
	gs.BeginWarLedger("player", "enemy")

	want := diplomacy.AssessPeaceProposal(gs, "player", "enemy")
	chance, status := estimateDiplomacyChance(gs, "enemy", ActionProposePeace)
	if chance != want.Chance {
		t.Fatalf("ekran ve gerçek barış değerlendirmesi ayrıştı: screen=%d want=%d", chance, want.Chance)
	}
	if chance >= 100 || status == "Kesin kabul" {
		t.Fatalf("reddedilecek barış teklifi ekranda kesin görünmemeli: chance=%d status=%q", chance, status)
	}
}

func TestAllianceChanceAllowsDirectThreatWhenCommonEnemyExists(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"ally":   {ID: "ally", NameTR: "Müttefik"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Regions: map[world.RegionID]*world.Region{
			"p1": {ID: "p1", OwnerID: "player", Neighbors: []world.RegionID{"a1"}},
			"a1": {ID: "a1", OwnerID: "ally", Neighbors: []world.RegionID{"p1"}},
			"e1": {ID: "e1", OwnerID: "enemy"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"):  {FactionA: "player", FactionB: "ally", Stance: faction.StancePeace, Score: 25},
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar, Score: -80},
			faction.RelationKey("ally", "enemy"):   {FactionA: "ally", FactionB: "enemy", Stance: faction.StanceWar, Score: -80},
		},
		Armies: map[army.ArmyID]*army.Army{
			"p_army": {ID: "p_army", OwnerID: "player", RegionID: "p1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
			"a_army": {ID: "a_army", OwnerID: "ally", RegionID: "a1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
	}

	if reason := diplomacyActionDisabledReason(gs, "ally", ActionProposeAlliance); reason != "" {
		t.Fatalf("ortak düşman doğrudan tehdidi block reason'a dönüştürmemeli, got=%q", reason)
	}
	chance, status := estimateDiplomacyChance(gs, "ally", ActionProposeAlliance)
	if chance <= 0 {
		t.Fatalf("ittifak şansı sıfır olmamalıydı, got=%d status=%q", chance, status)
	}
}

func TestHandleDiplomacyInputSelectsOnFirstClickAndOpensOnSecond(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"a":      {ID: "a", NameTR: "A Devleti"},
			"b":      {ID: "b", NameTR: "B Devleti"},
		},
	}
	r := &Renderer{
		gs:             gs,
		showDiplomacy:  true,
		diplomacyFocus: 0,
		prevKeys:       make(map[ebiten.Key]bool),
		prevMouse:      make(map[ebiten.MouseButton]bool),
	}

	list := buildDiplomacyListView(gs, r.diplomacyFocus, r.diplomacyScroll)
	rowY := list.Rect.Y + diplomRowH + 8
	rowX := list.Rect.X + 20

	click := gameui.InputState{
		MouseX:          rowX,
		MouseY:          rowY,
		LeftJustPressed: true,
	}

	r.handleDiplomacyInput(click)
	if r.diplomacyFocus != 1 {
		t.Fatalf("ilk tık ikinci satırı seçmeli, got=%d", r.diplomacyFocus)
	}
	if r.diplomacyTargetFaction != "" {
		t.Fatalf("ilk tık teklif paneli açmamalı, got=%q", r.diplomacyTargetFaction)
	}

	r.handleDiplomacyInput(click)
	if r.diplomacyTargetFaction != "b" {
		t.Fatalf("aynı satıra ikinci tık teklif panelini açmalı, got=%q", r.diplomacyTargetFaction)
	}
}

func TestHandleDiplomacyInputVassalDoubleClickOpensOverlord(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":   {ID: "player"},
			"a_lord":   {ID: "a_lord", NameTR: "Himaye Eden Devlet"},
			"b_vassal": {ID: "b_vassal", NameTR: "Bağlı Devlet", OverlordID: "a_lord"},
		},
	}
	r := &Renderer{
		gs:            gs,
		showDiplomacy: true,
		prevKeys:      make(map[ebiten.Key]bool),
		prevMouse:     make(map[ebiten.MouseButton]bool),
	}

	list := buildDiplomacyListView(gs, r.diplomacyFocus, r.diplomacyScroll)
	click := gameui.InputState{
		MouseX:          list.Rect.X + 20,
		MouseY:          list.Rect.Y + diplomRowH + 8,
		LeftJustPressed: true,
	}

	r.handleDiplomacyInput(click)
	if r.diplomacyFocus != 1 || r.diplomacyTargetFaction != "" {
		t.Fatalf("vassal ilk tıkta yalnız seçilmeli: focus=%d target=%q", r.diplomacyFocus, r.diplomacyTargetFaction)
	}

	r.handleDiplomacyInput(click)
	if r.diplomacyTargetFaction != "a_lord" {
		t.Fatalf("vassal çift tıklamasında overlord hedeflenmeli, got=%q", r.diplomacyTargetFaction)
	}
	if r.diplomacyFocus != 0 {
		t.Fatalf("teklif paneli açılırken focus overlord satırına taşınmalı, got=%d", r.diplomacyFocus)
	}
}

func TestHandleDiplomacyInputHistoryClickOpensRelevantFaction(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
			"ai_2":   {ID: "ai_2", NameTR: "AI 2"},
		},
		DiplomaticOfferHistory: []state.DiplomaticOfferHistoryEntry{
			{FromFactionID: "ai_2", ToFactionID: "ai_1", Action: "propose_alliance", CreatedTurn: 1, ResolvedTurn: 2, Accepted: true, Applied: true},
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_trade", CreatedTurn: 3, ResolvedTurn: 4, Accepted: true, Applied: true},
		},
	}
	r := &Renderer{
		gs:                      gs,
		showDiplomacy:           true,
		diplomacyHistoryVisible: true,
		prevKeys:                make(map[ebiten.Key]bool),
		prevMouse:               make(map[ebiten.MouseButton]bool),
	}

	layout := diplomacyListLayoutForScreen()
	cardRect := diplomacyOfferHistoryCardRect(layout.historyRect, 0)
	input := gameui.InputState{
		MouseX:          cardRect.X + 6,
		MouseY:          cardRect.Y + 6,
		LeftJustPressed: true,
	}

	act := r.handleDiplomacyInput(input)
	if act.Kind != ActionNone {
		t.Fatalf("history tıklamasında doğrudan oyun aksiyonu üretilmemeli, got=%s", act.Kind)
	}
	if r.diplomacyTargetFaction != "ai_1" {
		t.Fatalf("history tıklaması ilgili karşı fraksiyonu açmalıydı, got=%q", r.diplomacyTargetFaction)
	}
	if r.diplomacyActionFocus != 3 {
		t.Fatalf("trade history teklifinde action focus trade olmalıydı, got=%d", r.diplomacyActionFocus)
	}
}

func TestOpenDiplomacyTargetSelectsOfferPage(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"a":      {ID: "a", NameTR: "A Devleti"},
			"b":      {ID: "b", NameTR: "B Devleti"},
		},
	}
	r := &Renderer{gs: gs}

	r.openDiplomacyTarget("b", 0)

	if !r.showDiplomacy {
		t.Fatal("diplomasi paneli açılmalıydı")
	}
	if r.diplomacyTargetFaction != "b" {
		t.Fatalf("hedef fraksiyon teklif sayfasında seçili olmalıydı, got=%q", r.diplomacyTargetFaction)
	}
	if r.diplomacyFocus != 1 {
		t.Fatalf("liste odağı hedef fraksiyona taşınmalıydı, got=%d", r.diplomacyFocus)
	}
	if r.diplomacyActionFocus != 0 {
		t.Fatalf("varsayılan aksiyon odağı ilk satıra dönmeli, got=%d", r.diplomacyActionFocus)
	}
}

func TestHandleDiplomacyInputHistoryFiltersUpdateState(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
		},
	}
	r := &Renderer{
		gs:                      gs,
		showDiplomacy:           true,
		diplomacyHistoryVisible: true,
		prevKeys:                make(map[ebiten.Key]bool),
		prevMouse:               make(map[ebiten.MouseButton]bool),
	}

	layout := diplomacyListLayoutForScreen()
	buttons := buildDiplomacyHistoryFilterButtons(layout.historyRect, diplomacyHistoryDirectionAll, ActionNone)

	click := gameui.InputState{
		MouseX:          buttons[1].Button.X + 1,
		MouseY:          buttons[1].Button.Y + 1,
		LeftJustPressed: true,
	}
	r.handleDiplomacyInput(click)
	if r.diplomacyHistoryDirectionFilter != diplomacyHistoryDirectionIncoming {
		t.Fatalf("gelen filtresi seçilemedi, got=%v", r.diplomacyHistoryDirectionFilter)
	}
	if r.diplomacyHistoryActionFilter != ActionNone {
		t.Fatalf("direction filtresi action filtresini değiştirmemeli, got=%s", r.diplomacyHistoryActionFilter)
	}

	click.MouseX = buttons[4].Button.X + 1
	click.MouseY = buttons[4].Button.Y + 1
	r.handleDiplomacyInput(click)
	if r.diplomacyHistoryDirectionFilter != diplomacyHistoryDirectionIncoming {
		t.Fatalf("action filtresi direction filtresini korumalı, got=%v", r.diplomacyHistoryDirectionFilter)
	}
	if r.diplomacyHistoryActionFilter != ActionProposeTrade {
		t.Fatalf("ticaret filtresi seçilemedi, got=%s", r.diplomacyHistoryActionFilter)
	}
}

func TestHandleDiplomacyInputTogglesRelationsAndHistory(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
		},
	}
	r := &Renderer{gs: gs, showDiplomacy: true}
	btn := buildDiplomacySideViewButton(diplomacyListLayoutForScreen().historyRect, false)
	click := gameui.InputState{MouseX: btn.X + 1, MouseY: btn.Y + 1, LeftJustPressed: true}

	r.handleDiplomacyInput(click)
	if !r.diplomacyHistoryVisible {
		t.Fatal("Geçmiş düğmesi geçmiş görünümünü açmalı")
	}
	btn = buildDiplomacySideViewButton(diplomacyListLayoutForScreen().historyRect, true)
	click.MouseX, click.MouseY = btn.X+1, btn.Y+1
	r.handleDiplomacyInput(click)
	if r.diplomacyHistoryVisible {
		t.Fatal("İlişkiler düğmesi aktif ilişkiler görünümüne dönmeli")
	}
}

func TestHandleDiplomacyOfferInputStateIgnoresSummaryPanelClicks(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
		},
		DiplomaticOffers: []state.DiplomaticOffer{
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_trade"},
		},
	}
	r := &Renderer{gs: gs, prevKeys: make(map[ebiten.Key]bool), prevMouse: make(map[ebiten.MouseButton]bool)}

	layout := diplomacyOfferLayoutForScreen()
	input := gameui.InputState{
		MouseX:          layout.historyRect.X + 12,
		MouseY:          layout.historyRect.Y + 12,
		LeftJustPressed: true,
	}

	act := r.handleDiplomacyOfferInputState(0, input)
	if act.Kind != ActionNone {
		t.Fatalf("summary panel tıklaması oyun aksiyonu üretmemeliydi, got=%s", act.Kind)
	}
	if r.showDiplomacy {
		t.Fatal("summary panel tıklaması genel diplomasi panelini açmamalı")
	}
	if r.diplomacyOfferHistoryBrowse != "" || r.diplomacyTargetFaction != "" {
		t.Fatalf("summary panel tıklaması browse state üretmemeliydi, got browse=%q target=%q", r.diplomacyOfferHistoryBrowse, r.diplomacyTargetFaction)
	}
}

func TestDiplomacyOfferSummaryListsReflectState(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", NameTR: "Piyade"},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
			"ally":   {ID: "ally", NameTR: "Müttefik"},
			"trade":  {ID: "trade", NameTR: "Tüccar"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
			"lord":   {ID: "lord", NameTR: "Efendi"},
		},
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "ai_1"},
			"r2": {ID: "r2", OwnerID: "ai_1"},
			"r3": {ID: "r3", OwnerID: "ai_1", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "ai_1", Units: []army.Unit{{TypeID: "inf"}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "ally"):  {FactionA: "ai_1", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("ai_1", "trade"): {FactionA: "ai_1", FactionB: "trade", Stance: faction.StanceTrade},
			faction.RelationKey("ai_1", "enemy"): {FactionA: "ai_1", FactionB: "enemy", Stance: faction.StanceWar},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "ai_1", ToFactionID: "trade"},
		},
	}
	if got := diplomacyOfferRelationLine(gs, "ai_1", diplomacyOfferRelationAllied); got != "Müttefik" {
		t.Fatalf("ittifak listesi beklenmedik: %q", got)
	}
	if got := diplomacyOfferRelationLine(gs, "ai_1", diplomacyOfferRelationTrade); got != "Tüccar" {
		t.Fatalf("ticaret listesi beklenmedik: %q", got)
	}
	if got := diplomacyOfferRelationLine(gs, "ai_1", diplomacyOfferRelationWar); got != "Düşman" {
		t.Fatalf("savaş listesi beklenmedik: %q", got)
	}
	if got := diplomacyFactionLandRegionCount(gs, "ai_1"); got != 2 {
		t.Fatalf("beklenmeyen kara bölge sayısı: got=%d", got)
	}
	if got := diplomacyFactionOwnedRegionCount(gs, "ai_1"); got != 3 {
		t.Fatalf("beklenmeyen toplam bölge sayısı: got=%d", got)
	}
}

func TestDiplomacyOfferActionLabelTRWarJoinCall(t *testing.T) {
	if got := diplomacyOfferActionLabelTR(string(diplomacy.ActionJoinWarCall)); got != "savaşa katılım" {
		t.Fatalf("beklenmeyen savaş çağrısı etiketi: %q", got)
	}
}

func TestDiplomacyOfferMessageTRWarJoinCall(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ally":   {ID: "ally", NameTR: "Müttefik"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
			"player": {ID: "player", NameTR: "Oyuncu"},
		},
	}
	offer := state.DiplomaticOffer{
		FromFactionID:        "ally",
		ToFactionID:          "player",
		Action:               string(diplomacy.ActionJoinWarCall),
		WarDeclarerFactionID: "enemy",
		WarEnemyFactionID:    "enemy",
	}

	got := diplomacyOfferMessageTR(gs, offer)
	want := "Düşman devleti Müttefik devletine savaş ilan etti. Müttefikiniz sizi kendi safında savaşa çağırıyor."
	if got != want {
		t.Fatalf("beklenmeyen savaş çağrısı mesajı:\nwant=%q\ngot =%q", want, got)
	}
}

func TestDiplomacyOfferTruceNoticeTR(t *testing.T) {
	if got := diplomacyOfferTruceNoticeTR(state.DiplomaticOffer{Action: string(diplomacy.ActionProposePeace)}); got != "Ateşkes: Barışı kabul ederseniz 6 tur boyunca bu devlete yeniden savaş ilan edemezsiniz." {
		t.Fatalf("beklenmeyen ateşkes bildirimi: %q", got)
	}
	if got := diplomacyOfferTruceNoticeTR(state.DiplomaticOffer{Action: string(diplomacy.ActionProposeTrade)}); got != "" {
		t.Fatalf("barış dışı tekliflerde ateşkes bildirimi gösterilmemeli: %q", got)
	}
}

func TestHandleDiplomacyInputBackKeepsBrowseHighlightOnList(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
			"ai_2":   {ID: "ai_2", NameTR: "AI 2"},
		},
	}
	r := &Renderer{
		gs:                          gs,
		showDiplomacy:               true,
		diplomacyTargetFaction:      "ai_1",
		diplomacyOfferHistoryBrowse: "ai_1",
		diplomacyActionFocus:        3,
		diplomacyFocus:              1,
		prevKeys:                    make(map[ebiten.Key]bool),
		prevMouse:                   make(map[ebiten.MouseButton]bool),
	}

	backBtn := buildDiplomacyBackButton()
	input := gameui.InputState{
		MouseX:          backBtn.X + 1,
		MouseY:          backBtn.Y + 1,
		LeftJustPressed: true,
	}

	act := r.handleDiplomacyInput(input)
	if act.Kind != ActionNone {
		t.Fatalf("back tıklaması doğrudan oyun aksiyonu üretmemeli, got=%s", act.Kind)
	}
	if r.diplomacyTargetFaction != "" {
		t.Fatalf("back list view'e dönünce target temizlenmeli, got=%q", r.diplomacyTargetFaction)
	}
	if r.diplomacyOfferHistoryBrowse != "ai_1" {
		t.Fatalf("browse highlight korunmalı, got=%q", r.diplomacyOfferHistoryBrowse)
	}
	if !r.showDiplomacy {
		t.Fatal("back list görünümünde diplomasi paneli açık kalmalı")
	}
	if _, ok := r.playerDiplomacyOfferIndex(); ok {
		t.Fatal("browse highlight aktifken offer modal tekrar açılmamalı")
	}
}

func TestDiplomacyOfferHistorySummaryCountsRelevantEntries(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		DiplomaticOfferHistory: []state.DiplomaticOfferHistoryEntry{
			{FromFactionID: "player", ToFactionID: "ai_1", Action: "propose_peace", Accepted: true, Applied: true},
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_trade", Accepted: false, Applied: false},
			{FromFactionID: "ai_2", ToFactionID: "ai_1", Action: "propose_alliance", Accepted: true, Applied: true},
		},
	}

	total, accepted, rejected, applied := diplomacyOfferHistorySummary(gs, diplomacyHistoryDirectionAll, ActionNone)
	if total != 2 || accepted != 1 || rejected != 1 || applied != 1 {
		t.Fatalf("beklenmeyen history özeti: total=%d accepted=%d rejected=%d applied=%d", total, accepted, rejected, applied)
	}
	if got := diplomacyOfferHistoryDirectionTR(gs.DiplomaticOfferHistory[0], gs.PlayerFactionID); got != "Giden" {
		t.Fatalf("giden kayıt için yön etiketi beklenmedik: %q", got)
	}
	if got := diplomacyOfferHistoryDirectionTR(gs.DiplomaticOfferHistory[1], gs.PlayerFactionID); got != "Gelen" {
		t.Fatalf("gelen kayıt için yön etiketi beklenmedik: %q", got)
	}

	filteredTotal, filteredAccepted, filteredRejected, filteredApplied := diplomacyOfferHistorySummary(gs, diplomacyHistoryDirectionIncoming, ActionProposeTrade)
	if filteredTotal != 1 || filteredAccepted != 0 || filteredRejected != 1 || filteredApplied != 0 {
		t.Fatalf("filtreli history özeti beklenmedik: total=%d accepted=%d rejected=%d applied=%d", filteredTotal, filteredAccepted, filteredRejected, filteredApplied)
	}
}

func TestHandleDiplomacyInputHistorySelectionRespectsFilters(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
			"ai_2":   {ID: "ai_2", NameTR: "AI 2"},
		},
		DiplomaticOfferHistory: []state.DiplomaticOfferHistoryEntry{
			{FromFactionID: "player", ToFactionID: "ai_2", Action: "propose_trade", CreatedTurn: 1, ResolvedTurn: 2, Accepted: true, Applied: true},
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_peace", CreatedTurn: 3, ResolvedTurn: 4, Accepted: true, Applied: true},
			{FromFactionID: "ai_2", ToFactionID: "player", Action: "propose_trade", CreatedTurn: 5, ResolvedTurn: 6, Accepted: true, Applied: true},
		},
	}
	r := &Renderer{
		gs:                              gs,
		showDiplomacy:                   true,
		diplomacyHistoryVisible:         true,
		diplomacyHistoryDirectionFilter: diplomacyHistoryDirectionIncoming,
		diplomacyHistoryActionFilter:    ActionProposeTrade,
		prevKeys:                        make(map[ebiten.Key]bool),
		prevMouse:                       make(map[ebiten.MouseButton]bool),
	}

	layout := diplomacyListLayoutForScreen()
	cardRect := diplomacyOfferHistoryCardRect(layout.historyRect, 0)
	input := gameui.InputState{
		MouseX:          cardRect.X + 6,
		MouseY:          cardRect.Y + 6,
		LeftJustPressed: true,
	}

	act := r.handleDiplomacyInput(input)
	if act.Kind != ActionNone {
		t.Fatalf("filtreli history tıklamasında doğrudan oyun aksiyonu üretilmemeli, got=%s", act.Kind)
	}
	if r.diplomacyTargetFaction != "ai_2" {
		t.Fatalf("filtre yalnız incoming+trade kaydı açmalı, got=%q", r.diplomacyTargetFaction)
	}
	if r.diplomacyActionFocus != 3 {
		t.Fatalf("trade history kayıt action focus trade olmalı, got=%d", r.diplomacyActionFocus)
	}
}

func TestHandleDiplomacyInputHistoryHoverDoesNotMutateState(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
			"ai_2":   {ID: "ai_2", NameTR: "AI 2"},
		},
		DiplomaticOfferHistory: []state.DiplomaticOfferHistoryEntry{
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_trade", CreatedTurn: 3, ResolvedTurn: 4, Accepted: true, Applied: true},
		},
	}
	r := &Renderer{
		gs:                      gs,
		showDiplomacy:           true,
		diplomacyHistoryVisible: true,
		prevKeys:                make(map[ebiten.Key]bool),
		prevMouse:               make(map[ebiten.MouseButton]bool),
	}

	layout := diplomacyListLayoutForScreen()
	buttons := buildDiplomacyHistoryFilterButtons(layout.historyRect, diplomacyHistoryDirectionAll, ActionNone)
	filterHover := gameui.InputState{
		MouseX: buttons[1].Button.X + 1,
		MouseY: buttons[1].Button.Y + 1,
	}
	r.handleDiplomacyInput(filterHover)
	if r.diplomacyHistoryDirectionFilter != diplomacyHistoryDirectionAll {
		t.Fatalf("hover filter state degistirmemeli, got=%v", r.diplomacyHistoryDirectionFilter)
	}

	cardRect := diplomacyOfferHistoryCardRect(layout.historyRect, 0)
	cardHover := gameui.InputState{
		MouseX: cardRect.X + 6,
		MouseY: cardRect.Y + 6,
	}
	act := r.handleDiplomacyInput(cardHover)
	if act.Kind != ActionNone {
		t.Fatalf("hover history aksiyon uretmemeli, got=%s", act.Kind)
	}
	if r.diplomacyTargetFaction != "" {
		t.Fatalf("hover history browse acmamali, got=%q", r.diplomacyTargetFaction)
	}
}
