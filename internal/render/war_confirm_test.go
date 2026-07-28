package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func buttonRectsOverlap(a, b gameui.Button) bool {
	return a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H
}

func TestHandleDiplomacyInputOpensWarConfirmModal(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StancePeace, Score: -40},
		},
	}
	r := &Renderer{
		gs:                     gs,
		showDiplomacy:          true,
		diplomacyTargetFaction: "enemy",
		diplomacyActionFocus:   0,
		prevKeys:               map[ebiten.Key]bool{},
		prevMouse:              map[ebiten.MouseButton]bool{},
	}

	sendBtn := buildDiplomacySendButton()
	act := r.handleDiplomacyInput(gameui.InputState{
		MouseX:          sendBtn.X + 5,
		MouseY:          sendBtn.Y + 5,
		LeftJustPressed: true,
	})

	if act.Kind != ActionNone {
		t.Fatalf("savaş ilanı önce modal açmalı, doğrudan aksiyon üretmemeli: %+v", act)
	}
	if !r.warConfirm.show {
		t.Fatal("war confirm modalı açılmalıydı")
	}
	if r.warConfirm.factionID != "enemy" {
		t.Fatalf("yanlış hedef modalı açıldı: %q", r.warConfirm.factionID)
	}
	if r.showDiplomacy {
		t.Fatal("modal açılınca diplomasi paneli kapanmalıydı")
	}
}

func TestFinalizeWarConfirmCarriesSelectedAllies(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: "player"}
	r := &Renderer{gs: gs}
	wc := warConfirmState{
		factionID: "enemy",
		preview: diplomacy.WarDeclarationPreview{
			Attacker: diplomacy.WarSidePreview{
				CallableAllies: []diplomacy.WarParticipantPreview{
					{FactionID: "ally_a"},
					{FactionID: "ally_b"},
				},
			},
		},
		selectedAllies: map[faction.FactionID]bool{
			"ally_b": true,
		},
		battleContext: combat.BattleContextLand,
	}

	act := r.finalizeWarConfirm(wc)

	if act.Kind != ActionDeclareWar {
		t.Fatalf("pending hareket yokken yalnız savaş ilanı dönmeliydi, got=%s", act.Kind)
	}
	if len(act.WarAllies) != 1 || act.WarAllies[0] != "ally_b" {
		t.Fatalf("seçili müttefik listesi aksiyona taşınmalıydı, got=%v", act.WarAllies)
	}
}

func TestFinalizeWarConfirmOpensSiegeDecisionWithoutSiegeUnit(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "player",
				Neighbors: []world.RegionID{"dst"},
			},
			"dst": {
				ID:          "dst",
				OwnerID:     "enemy",
				Neighbors:   []world.RegionID{"src"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress, NameTR: "Kale"}},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"atk": {
				ID:            "atk",
				OwnerID:       "player",
				RegionID:      "src",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
			"def": {
				ID:            "def",
				OwnerID:       "enemy",
				RegionID:      "dst",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	r := &Renderer{gs: gs}
	wc := warConfirmState{
		factionID:       "enemy",
		pendingArmy:     "atk",
		pendingDest:     "dst",
		pendingEnemy:    "def",
		opensBattlePlan: true,
		battleAction:    ActionMoveArmy,
		preview: diplomacy.WarDeclarationPreview{
			Attacker: diplomacy.WarSidePreview{},
		},
		battleContext: combat.BattleContextLand,
	}

	act := r.finalizeWarConfirm(wc)

	if act.Kind != ActionDeclareWar {
		t.Fatalf("savaş ilanı geri dönmeliydi, got=%s", act.Kind)
	}
	if !r.confirmDialog.show {
		t.Fatal("kuşatma kararı modalı açılmalıydı")
	}
	if r.confirmDialog.thirdLabel != "Genel Hücum" {
		t.Fatalf("kuşatma birimi olmasa da genel hücum düğmesi görünmeli, got=%q", r.confirmDialog.thirdLabel)
	}
	if r.confirmDialog.pendingAction.Kind != ActionStartSiege {
		t.Fatalf("kuşatma kararı start siege üretmeliydi, got=%s", r.confirmDialog.pendingAction.Kind)
	}
	if r.battlePlan.show {
		t.Fatal("kuşatma kararı modalı doğrudan battle plan açmamalıydı")
	}
}

func TestFinalizeWarConfirmRoutesBesiegingArmyFightInsteadOfSiegeDecision(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"ally":   {ID: "ally", NameTR: "Müttefik"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "player",
				Neighbors: []world.RegionID{"dst"},
			},
			"dst": {
				ID:          "dst",
				OwnerID:     "ally",
				Neighbors:   []world.RegionID{"src"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress, NameTR: "Kale"}},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"):  {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"atk": {
				ID:            "atk",
				OwnerID:       "player",
				RegionID:      "src",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
			"besieger": {
				ID:            "besieger",
				OwnerID:       "enemy",
				RegionID:      "dst",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"dst": {
				RegionID:          "dst",
				AttackerArmyID:    "besieger",
				AttackerFactionID: "enemy",
				StartedTurn:       6,
				FortLevel:         2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	r := &Renderer{gs: gs}
	wc := warConfirmState{
		factionID:       "enemy",
		pendingArmy:     "atk",
		pendingDest:     "dst",
		pendingEnemy:    "besieger",
		opensBattlePlan: true,
		battleAction:    ActionMoveArmy,
		battleContext:   combat.BattleContextLand,
		preview: diplomacy.WarDeclarationPreview{
			Attacker: diplomacy.WarSidePreview{},
		},
	}

	act := r.finalizeWarConfirm(wc)

	if act.Kind != ActionDeclareWar {
		t.Fatalf("savaş ilanı geri dönmeliydi, got=%s", act.Kind)
	}
	if r.confirmDialog.show {
		t.Fatal("kuşatma kararı açılmamalıydı")
	}
	if !r.battlePlan.show {
		t.Fatal("aktif kuşatma altındaki düşman ordu için savaş planı açılmalıydı")
	}
	if r.battlePlan.pendingEnemy != "besieger" {
		t.Fatalf("battle plan yanlış düşmanı hedefledi: %q", r.battlePlan.pendingEnemy)
	}
}

func TestOpenSiegeDecisionIncludesCommanderOperationalSummary(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"dst": {
				ID:          "dst",
				OwnerID:     "enemy",
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress, NameTR: "Kale"}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":  {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
			"bomb": {ID: "bomb", Category: army.CategorySiege, Tier: 1, Attack: 8, Defense: 4, Morale: 40},
		},
	}
	attacker := &army.Army{
		ID:      "atk",
		OwnerID: "player",
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "bomb", CurrentHP: 100},
		},
		Commander: &army.Commander{
			ID:     "cmd",
			Name:   "Osman Bey",
			Level:  4,
			Traits: []army.CommanderTrait{army.CommanderTraitTactician, army.CommanderTraitAggressor},
		},
	}
	r := &Renderer{gs: gs}

	r.openSiegeDecision(attacker, gs.Regions["dst"])

	if !r.confirmDialog.show {
		t.Fatal("kuşatma kararı modalı açılmalıydı")
	}
	if !strings.Contains(r.confirmDialog.message, "Osman Bey") {
		t.Fatalf("komutan adı modal mesajında görünmeli, got=%q", r.confirmDialog.message)
	}
	if !strings.Contains(r.confirmDialog.message, "Hareket +1") || !strings.Contains(r.confirmDialog.message, "Kuşatma +1/+1") {
		t.Fatalf("operasyon bonusları modal mesajında görünmeli, got=%q", r.confirmDialog.message)
	}
}

func TestSelectedDefensiveSiegePanelFollowsArmyAndSettlementSelection(t *testing.T) {
	gs := &state.GameState{
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"enemy":  {ID: "enemy", NameTR: "Kuşatan"},
		},
		Regions: map[world.RegionID]*world.Region{
			"fort": {ID: "fort", OwnerID: "player", Settlements: []world.Settlement{{ID: "castle", NameTR: "Kale"}}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"besieger": {ID: "besieger", OwnerID: "enemy", RegionID: "fort", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"defender": {ID: "defender", OwnerID: "player", RegionID: "fort", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {RegionID: "fort", AttackerArmyID: "besieger", DefenderArmyID: "defender", AttackerFactionID: "enemy", FortLevel: 2},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	r := &Renderer{gs: gs, SelectedArmy: "defender", SelectedRegion: "fort"}
	defender, attacker, siege, target, surrenderOffered, ok := r.selectedDefensiveSiegePanelState()
	if !ok || defender == nil || defender.ID != "defender" || attacker.ID != "besieger" || siege.RegionID != "fort" || target.ID != "fort" {
		t.Fatalf("ordu seçimi savunma kuşatma paneline bağlanmalıydı: defender=%+v attacker=%+v siege=%+v target=%+v ok=%t", defender, attacker, siege, target, ok)
	}

	r.SelectedArmy = ""
	r.SelectedRegion = "fort"
	defender, attacker, _, _, surrenderOffered, ok = r.selectedDefensiveSiegePanelState()
	if !ok || defender == nil || defender.ID != "defender" || attacker.ID != "besieger" {
		t.Fatalf("yerleşim/bölge seçimi canlı savunma ordusunu bulmalıydı: defender=%+v attacker=%+v ok=%t", defender, attacker, ok)
	}
	sortie, surrender := buildDefensiveSiegeButtons()
	if sortie.Label != "Huruç başlat" || surrender.Label != "Teslim ol" {
		t.Fatalf("savunma kuşatması düğmeleri yanlış: %q / %q", sortie.Label, surrender.Label)
	}
	assault, lift, offer := buildSelectedSiegeButtons()
	if assault.Label != "Genel Hücum" || lift.Label != "Kuşatmayı Kaldır" || offer.Label != "Teslimiyet Teklifi" {
		t.Fatalf("saldıran kuşatma düğmeleri yanlış: %q / %q / %q", assault.Label, lift.Label, offer.Label)
	}
	if buttonRectsOverlap(assault, lift) || buttonRectsOverlap(assault, offer) || buttonRectsOverlap(lift, offer) {
		t.Fatal("saldıran kuşatma düğmeleri birbirinin üzerine binmemeli")
	}
	delete(gs.Armies, "defender")
	defender, attacker, _, _, surrenderOffered, ok = r.selectedDefensiveSiegePanelState()
	if !ok || defender != nil || attacker == nil || attacker.ID != "besieger" {
		t.Fatalf("savunma ordusu kalmasa da yerleşim kuşatma paneli açık kalmalıydı: defender=%+v attacker=%+v ok=%t", defender, attacker, ok)
	}
	if surrenderOffered {
		t.Fatalf("teslimiyet teklifi yokken düğme aktif görünmemeli")
	}
	gs.DiplomaticOffers = []state.DiplomaticOffer{{
		FromFactionID: "enemy", ToFactionID: "player", Action: string(diplomacy.ActionProposeSurrender), RegionID: "fort",
	}}
	_, _, _, _, surrenderOffered, ok = r.selectedDefensiveSiegePanelState()
	if !ok || !surrenderOffered {
		t.Fatalf("AI teslimiyet teklifi geldiğinde Teslim ol düğmesi aktif olmalı")
	}
}

func TestSelectedSiegeSurrenderOfferDisabledAfterSameTurnRejection(t *testing.T) {
	gs := &state.GameState{
		Turn:            8,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Regions: map[world.RegionID]*world.Region{
			"fort":  {ID: "fort", OwnerID: "enemy"},
			"other": {ID: "other", OwnerID: "enemy"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"besieger": {ID: "besieger", OwnerID: "player", RegionID: "fort", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {RegionID: "fort", AttackerArmyID: "besieger", AttackerFactionID: "player"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
		OfferRejectionTurns: map[string]int{
			state.DiplomaticOfferRegionRejectionKey("player", "enemy", string(diplomacy.ActionProposeSurrender), "fort"): 8,
		},
	}
	r := &Renderer{gs: gs}

	if sent, canSend := r.attackerSiegeSurrenderState(gs.Armies["besieger"], gs.Regions["fort"]); sent || canSend {
		t.Fatalf("aynı tur reddedilen kuşatmada teklif düğmesi pasif olmalı: sent=%t canSend=%t", sent, canSend)
	}
	if sent, canSend := r.attackerSiegeSurrenderState(gs.Armies["besieger"], gs.Regions["other"]); sent || !canSend {
		t.Fatalf("ret diğer bölgenin teklif düğmesini etkilememeli: sent=%t canSend=%t", sent, canSend)
	}
}

func TestWarConfirmScrollHelpersClampAndHitVisibleRows(t *testing.T) {
	viewport := gameui.Rect{X: 100, Y: 200, W: 320, H: 110}
	entries := []diplomacy.WarParticipantPreview{
		{FactionID: "a"},
		{FactionID: "b"},
		{FactionID: "c"},
		{FactionID: "d"},
		{FactionID: "e"},
	}
	selected := map[faction.FactionID]bool{"c": true}

	if rows := warConfirmVisibleRows(viewport); rows != 2 {
		t.Fatalf("110px viewport için 2 görünür satır bekleniyordu, got=%d", rows)
	}
	if got := clampWarConfirmScroll(len(entries), viewport, 99); got != 3 {
		t.Fatalf("scroll max'a clamplenmeli, got=%d", got)
	}

	boxes := warConfirmCheckboxes(viewport, entries, selected, 2)
	if len(boxes) != 2 {
		t.Fatalf("scroll sonrası yalnız görünür checkbox'lar dönmeli, got=%d", len(boxes))
	}
	if !boxes[0].Checked {
		t.Fatal("scroll ile görünür hale gelen seçili entry checked olmalıydı")
	}
	if idx := warConfirmEntryIndexAt(viewport, entries, 2, boxes[0].Rect.X+2, boxes[0].Rect.Y+2); idx != 2 {
		t.Fatalf("ilk görünür checkbox üçüncü entry'ye maplenmeliydi, got=%d", idx)
	}
}

func TestWarConfirmListViewportsDoNotOverlap(t *testing.T) {
	side := gameui.Rect{X: 100, Y: 200, W: 320, H: 372}
	autoViewport := warConfirmAutoViewport(side)
	callViewport := warConfirmCallViewport(side)

	if autoViewport.Y+autoViewport.H > callViewport.Y {
		t.Fatalf("otomatik ve çağrılabilir listelerin viewport'ları örtüşmemeli: auto=%+v call=%+v", autoViewport, callViewport)
	}
	if rows := warConfirmVisibleRows(autoViewport); rows < 1 {
		t.Fatal("otomatik katılanlar viewport'u en az bir satır göstermeli")
	}
	if rows := warConfirmVisibleRows(callViewport); rows < 1 {
		t.Fatal("çağrılabilir müttefikler viewport'u en az bir satır göstermeli")
	}
	row := warConfirmEntryRowRect(autoViewport, 0)
	if row.Y+row.H > autoViewport.Y+autoViewport.H {
		t.Fatalf("katılımcı kartı viewport dışına taşmamalı: row=%+v viewport=%+v", row, autoViewport)
	}
	if row.H < warConfirmRowPitch {
		t.Fatalf("üç satırlı katılımcı kartı not satırını kapsamalı: height=%.1f", row.H)
	}
}

func TestWarConfirmScrollTargetIsLocalToViewport(t *testing.T) {
	side := gameui.Rect{X: 100, Y: 200, W: 320, H: 372}
	autoViewport := warConfirmAutoViewport(side)
	callViewport := warConfirmCallViewport(side)

	if hit, auto := warConfirmScrollTarget(side, autoViewport, callViewport, autoViewport.X+4, autoViewport.Y+4); !hit || !auto {
		t.Fatal("otomatik katılanlar viewport'u kendi scroll hedefi olmalı")
	}
	if hit, auto := warConfirmScrollTarget(side, autoViewport, callViewport, callViewport.X+4, callViewport.Y+4); !hit || auto {
		t.Fatal("çağrılabilir müttefikler viewport'u kendi scroll hedefi olmalı")
	}
	if hit, _ := warConfirmScrollTarget(side, autoViewport, callViewport, side.X+4, side.Y+70); hit {
		t.Fatal("liste dışındaki başlık alanı scroll hedefi olmamalı")
	}
}
