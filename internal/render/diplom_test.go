package render

import (
	"testing"

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
