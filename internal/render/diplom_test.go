package render

import (
	"testing"

	"mapp-game-go/internal/army"
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
	if status != "İlişki puanı 10 altı" {
		t.Fatalf("beklenen gerçek engel metni, got=%q", status)
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
			faction.RelationKey("player", "ally"):  {FactionA: "player", FactionB: "ally", Stance: faction.StancePeace, Score: 20},
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

func TestHandleDiplomacyOfferInputHistoryFiltersUpdateState(t *testing.T) {
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
	r := &Renderer{
		gs: gs,
	}

	layout := diplomacyOfferLayoutForScreen()
	buttons := buildDiplomacyHistoryFilterButtons(layout.historyRect, diplomacyHistoryDirectionAll, ActionNone)

	if !r.applyDiplomacyHistoryFilterHit(layout.historyRect, buttons[2].Button.X+1, buttons[2].Button.Y+1) {
		t.Fatal("offer modal history direction filtresi tıklaması consumed edilmedi")
	}
	if r.diplomacyHistoryDirectionFilter != diplomacyHistoryDirectionOutgoing {
		t.Fatalf("offer modal history direction filtresi seçilemedi, got=%v", r.diplomacyHistoryDirectionFilter)
	}
	if r.diplomacyHistoryActionFilter != ActionNone {
		t.Fatalf("direction filtresi action filtresini değiştirmemeli, got=%s", r.diplomacyHistoryActionFilter)
	}

	if !r.applyDiplomacyHistoryFilterHit(layout.historyRect, buttons[4].Button.X+1, buttons[4].Button.Y+1) {
		t.Fatal("offer modal action filtresi tıklaması consumed edilmedi")
	}
	if r.diplomacyHistoryDirectionFilter != diplomacyHistoryDirectionOutgoing {
		t.Fatalf("action filtresi direction filtresini korumalı, got=%v", r.diplomacyHistoryDirectionFilter)
	}
	if r.diplomacyHistoryActionFilter != ActionProposeTrade {
		t.Fatalf("offer modal ticaret filtresi seçilemedi, got=%s", r.diplomacyHistoryActionFilter)
	}
}

func TestHandleDiplomacyOfferInputStateHistoryCardOpensBrowsePanel(t *testing.T) {
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
		DiplomaticOffers: []state.DiplomaticOffer{
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_trade"},
		},
		DiplomaticOfferHistory: []state.DiplomaticOfferHistoryEntry{
			{FromFactionID: "ai_2", ToFactionID: "player", Action: "propose_alliance", CreatedTurn: 1, ResolvedTurn: 2, Accepted: true, Applied: true},
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_trade", CreatedTurn: 3, ResolvedTurn: 4, Accepted: true, Applied: true},
		},
	}
	r := &Renderer{
		gs:       gs,
		prevKeys: make(map[ebiten.Key]bool),
	}

	layout := diplomacyOfferLayoutForScreen()
	cardRect := diplomacyOfferHistoryCardRect(layout.historyRect, 0)
	input := gameui.InputState{
		MouseX:          cardRect.X + 6,
		MouseY:          cardRect.Y + 6,
		LeftJustPressed: true,
	}

	act := r.handleDiplomacyOfferInputState(0, input)
	if act.Kind != ActionNone {
		t.Fatalf("history kartı tıklaması doğrudan oyun aksiyonu üretmemeli, got=%s", act.Kind)
	}
	if !r.showDiplomacy {
		t.Fatal("history kartı tıklaması genel diplomasi panelini açmalı")
	}
	if r.diplomacyOfferHistoryBrowse != "ai_1" {
		t.Fatalf("browse modu ilgili fraksiyonu tutmalı, got=%q", r.diplomacyOfferHistoryBrowse)
	}
	if r.diplomacyTargetFaction != "ai_1" {
		t.Fatalf("browse modu hedef fraksiyonu tutmalı, got=%q", r.diplomacyTargetFaction)
	}
	if r.diplomacyActionFocus != 3 {
		t.Fatalf("history teklifinde action focus trade olmalı, got=%d", r.diplomacyActionFocus)
	}
	if _, ok := r.playerDiplomacyOfferIndex(); ok {
		t.Fatal("browse modu açıkken offer modal aktif olmamalı")
	}
}

func TestHandleDiplomacyOfferInputStateHistoryCardIgnoredDuringAITurn(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Phase:           state.PhaseAITurn,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
			"ai_2":   {ID: "ai_2", NameTR: "AI 2"},
		},
		DiplomaticOffers: []state.DiplomaticOffer{
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_trade"},
		},
		DiplomaticOfferHistory: []state.DiplomaticOfferHistoryEntry{
			{FromFactionID: "ai_2", ToFactionID: "player", Action: "propose_alliance", CreatedTurn: 1, ResolvedTurn: 2, Accepted: true, Applied: true},
		},
	}
	r := &Renderer{
		gs:       gs,
		prevKeys: make(map[ebiten.Key]bool),
	}

	layout := diplomacyOfferLayoutForScreen()
	cardRect := diplomacyOfferHistoryCardRect(layout.historyRect, 0)
	input := gameui.InputState{
		MouseX:          cardRect.X + 6,
		MouseY:          cardRect.Y + 6,
		LeftJustPressed: true,
	}

	act := r.handleDiplomacyOfferInputState(0, input)
	if act.Kind != ActionNone {
		t.Fatalf("AI turunda history karti oyun aksiyonu uretmemeli, got=%s", act.Kind)
	}
	if r.showDiplomacy {
		t.Fatal("AI turunda history karti genel diplomasi panelini acmamali")
	}
	if r.diplomacyOfferHistoryBrowse != "" || r.diplomacyTargetFaction != "" {
		t.Fatalf("AI turunda browse state degismemeli, got browse=%q target=%q", r.diplomacyOfferHistoryBrowse, r.diplomacyTargetFaction)
	}
	if _, ok := r.playerDiplomacyOfferIndex(); !ok {
		t.Fatal("offer modal AI turunda aktif kalmali")
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
