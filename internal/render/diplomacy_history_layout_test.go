package render

import (
	"testing"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"

	gameui "mapp-game-go/internal/ui"
)

func TestDiplomacyPanelOffersAndCancelsTransitPassage(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":     {ID: "ai"},
			"player": {ID: "player"},
		},
	}
	index := -1
	for i, action := range diplomActions {
		if action.action == ActionProposeTransitPassage {
			index = i
			break
		}
	}
	if index < 0 {
		t.Fatal("diplomasi panelinde geçiş hakkı teklifi bulunamadı")
	}
	if got := diplomacyActionForTarget(gs, "ai", index); got != ActionProposeTransitPassage {
		t.Fatalf("anlaşma yokken panel aksiyonu = %q", got)
	}
	if reason := diplomacyActionDisabledReason(gs, "ai", ActionProposeTransitPassage); reason != "" {
		t.Fatalf("geçerli geçiş teklifi pasif: %s", reason)
	}

	gs.AddTransitAgreement("player", "ai")
	if got := diplomacyActionForTarget(gs, "ai", index); got != ActionCancelTransitPassage {
		t.Fatalf("anlaşma varken panel aksiyonu = %q, iptal bekleniyordu", got)
	}
	if got, want := diplomacyActionLabel(gs, "ai", index), diplomacy.ActionLabelTR(diplomacy.ActionCancelTransitPassage); got != want {
		t.Fatalf("aktif anlaşma düğmesi = %q, %q bekleniyordu", got, want)
	}
	if reason := diplomacyActionDisabledReason(gs, "ai", ActionCancelTransitPassage); reason != "" {
		t.Fatalf("aktif geçiş anlaşması iptal edilemiyor: %s", reason)
	}
}

func TestDiplomacyPanelDisablesTransitOfferForAlliesAndVassals(t *testing.T) {
	const (
		player = faction.FactionID("player")
		ally   = faction.FactionID("ally")
		vassal = faction.FactionID("vassal")
	)
	gs := &state.GameState{
		PlayerFactionID: player,
		Factions: map[faction.FactionID]*faction.Faction{
			player:    {ID: player},
			ally:      {ID: ally},
			vassal:    {ID: vassal, OverlordID: player},
			"neutral": {ID: "neutral"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(player, ally): {
				FactionA: player,
				FactionB: ally,
				Stance:   faction.StanceAllied,
			},
		},
	}
	transitIndex := -1
	for i, action := range diplomActions {
		if action.action == ActionProposeTransitPassage {
			transitIndex = i
			break
		}
	}
	if transitIndex < 0 {
		t.Fatal("diplomasi panelinde geçiş hakkı teklifi bulunamadı")
	}

	for _, target := range []faction.FactionID{ally, vassal} {
		if diplomacyActionEnabled(gs, target, transitIndex) {
			t.Errorf("%s için gereksiz geçiş hakkı teklifi etkin", target)
		}
		if reason := diplomacyActionDisabledReason(gs, target, ActionProposeTransitPassage); reason == "" {
			t.Errorf("%s için geçiş teklifi devre dışı bırakma nedeni gösterilmedi", target)
		}
	}
	if !diplomacyActionEnabled(gs, "neutral", transitIndex) {
		t.Fatal("ittifak/vassallık bağı olmayan devlete geçiş teklifi engellendi")
	}
}

func TestDiplomacyOfferActionsPairPeaceAndAllianceWithoutOverlap(t *testing.T) {
	for _, action := range diplomActions {
		if action.action == ActionDeclareWar {
			t.Fatal("savaş aksiyonu teklif panelinde görünmeye devam ediyor")
		}
	}
	if len(diplomActions) < 3 {
		t.Fatalf("diplomasi panelinde beklenen aksiyonlar eksik: %d", len(diplomActions))
	}

	layout := diplomacyOfferLayoutForScreen()
	buttons := buildDiplomacyActionButtons(nil, "")
	buttonRect := func(button gameui.Button) gameui.Rect {
		return gameui.Rect{X: button.X, Y: button.Y, W: button.W, H: button.H}
	}
	peace := buttonRect(buttons[0].Button)
	alliance := buttonRect(buttons[1].Button)
	if peace.Y != alliance.Y || peace.H != diplomActionPairButtonH || alliance.H != diplomActionPairButtonH {
		t.Fatalf("barış ve ittifak aynı satırda/kompakt değil: barış=%+v ittifak=%+v", peace, alliance)
	}
	if peace.X+peace.W+diplomActionGap > alliance.X {
		t.Fatalf("barış ve ittifak düğmeleri üst üste geliyor: barış=%+v ittifak=%+v", peace, alliance)
	}
	next := buttonRect(buttons[2].Button)
	if next.Y < peace.Y+peace.H+diplomActionGap {
		t.Fatalf("sonraki teklif satırı ikili satıra taşıyor: ikili=%+v sonraki=%+v", peace, next)
	}
	for i, button := range buttons {
		rect := buttonRect(button.Button)
		if rect.X < layout.actionsRect.X || rect.X+rect.W > layout.actionsRect.X+layout.actionsRect.W ||
			rect.Y < layout.actionsRect.Y || rect.Y+rect.H > layout.actionsRect.Y+layout.actionsRect.H {
			t.Errorf("aksiyon %d teklif alanı dışına taşıyor: %+v alan=%+v", i, rect, layout.actionsRect)
		}
	}
	if bottom, limit := layout.actionsRect.Y+layout.actionsRect.H, layout.selectedRect.Y+layout.selectedRect.H; bottom > limit {
		t.Errorf("teklif satırları seçili alanı örtüyor: satırlar=%.1f seçili alt=%.1f", bottom, limit)
	}
	if gap := layout.backRect.Y - (layout.actionsRect.Y + layout.actionsRect.H); gap < diplomOfferFooterGap {
		t.Errorf("teklif düğmeleriyle alt çerçeve düğmeleri arasında yeterli boşluk yok: %.1f, en az %.1f olmalı", gap, diplomOfferFooterGap)
	}
}

func TestDiplomacyHistoryActionFiltersUseSpaciousCenteredRows(t *testing.T) {
	panel := gameui.Rect{X: 100, Y: 200, W: diplomHistoryPanelW, H: diplomHistoryPanelH}
	buttons := buildDiplomacyHistoryFilterButtons(panel, diplomacyHistoryDirectionAll, ActionNone)

	for i := 3; i < len(buttons); i++ {
		button := buttons[i].Button
		if button.W < 75 {
			t.Errorf("action filter %q width = %.1f, want at least 75", button.Label, button.W)
		}
		if button.X < panel.X || button.X+button.W > panel.X+panel.W {
			t.Errorf("action filter %q extends beyond panel: %+v", button.Label, button)
		}
	}

	if got, want := buttons[7].Button.X, buttons[6].Button.X+buttons[6].Button.W+6; got != want {
		t.Errorf("second action row is not centered/contiguous: last button starts at %.1f, want %.1f", got, want)
	}
	if buttons[7].Button.Y != buttons[6].Button.Y {
		t.Errorf("second action row buttons have different Y positions: %.1f and %.1f", buttons[6].Button.Y, buttons[7].Button.Y)
	}
}

func TestDiplomacyHistoryVisibleEntriesStayInsidePanel(t *testing.T) {
	for _, tc := range []struct {
		name       string
		panelH     float64
		maxEntries int
		want       int
	}{
		{name: "legacy height clips fourth card", panelH: 324, maxEntries: 4, want: 3},
		{name: "expanded panel fits four cards", panelH: 400, maxEntries: 4, want: 4},
		{name: "respects caller limit", panelH: 400, maxEntries: 3, want: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			panel := gameui.Rect{X: 40, Y: 60, W: diplomHistoryPanelW, H: tc.panelH}
			if got := diplomacyHistoryVisibleEntries(panel, tc.maxEntries); got != tc.want {
				t.Fatalf("visible entries = %d, want %d", got, tc.want)
			}
			count := diplomacyHistoryVisibleEntries(panel, tc.maxEntries)
			if count == 0 {
				return
			}
			last := diplomacyOfferHistoryCardRect(panel, count-1)
			if bottom, limit := last.Y+last.H, panel.Y+panel.H-8; bottom > limit {
				t.Errorf("last history card ends at %.1f, beyond inset panel bottom %.1f", bottom, limit)
			}
		})
	}
}
