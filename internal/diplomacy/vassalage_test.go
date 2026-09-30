package diplomacy

import (
	"strings"
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAnnexationTurnsRemainingMatchesAnnexationBlock(t *testing.T) {
	gs := &state.GameState{
		Turn:            16,
		PlayerFactionID: "overlord",
		Factions: map[faction.FactionID]*faction.Faction{
			"overlord": {ID: "overlord"},
			"vassal":   {ID: "vassal", OverlordID: "overlord", VassalizedTurn: 12, NameTR: "Vassal"},
		},
	}

	if got, want := AnnexationTurnsRemaining(gs, "overlord", "vassal"), 8; got != want {
		t.Fatalf("kalan ilhak turu: got=%d want=%d", got, want)
	}
	reason := ActionBlockReason(gs, "overlord", "vassal", ActionAnnexVassal)
	if !strings.HasSuffix(reason, "(8 tur kaldı).") {
		t.Fatalf("ilhak engel nedeni kalan turu içermiyor: %q", reason)
	}

	gs.Turn = 24
	if got := AnnexationTurnsRemaining(gs, "overlord", "vassal"); got != 0 {
		t.Fatalf("süre dolduktan sonra kalan ilhak turu: got=%d", got)
	}
	if reason := ActionBlockReason(gs, "overlord", "vassal", ActionAnnexVassal); reason != "" {
		t.Fatalf("süre dolduktan sonra ilhak engellendi: %s", reason)
	}
}

func TestRelationImprovementMessageNamesSenderAndTarget(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"sender": {ID: "sender", NameTR: "Gönderen Devlet"},
			"target": {ID: "target", NameTR: "Osmanoğulları Beyliği"},
		},
		Relations: map[string]*faction.Relation{},
	}

	result := applyRelationImprovement(gs, "sender", "target", 40, 8, 0, "diplomatik heyet")
	want := "Gönderen Devlet devleti, Osmanoğulları Beyliği için diplomatik heyet gönderdi. İlişki +8."
	if result.Message != want {
		t.Fatalf("ilişki geliştirme mesajı: got=%q want=%q", result.Message, want)
	}
}

func TestVassalizationRejectsTargetWithMoreThanThreeRegions(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"actor":  {ID: "actor", NameTR: "Üst Devlet"},
			"target": {ID: "target", NameTR: "Hedef Devlet"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("actor", "target"): {
				FactionA: "actor", FactionB: "target", Score: 80, Stance: faction.StancePeace,
			},
		},
		Regions: map[world.RegionID]*world.Region{},
	}
	for i := 0; i < 4; i++ {
		id := world.RegionID("target_" + string(rune('a'+i)))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "target"}
	}

	assessment := AssessVassalizationProposal(gs, Relation(gs, "actor", "target"), "actor", "target")
	if assessment.BlockReason == "" {
		t.Fatal("dört bölgeli hedef için vassallık kabulü engellenmedi")
	}
	if reason := ActionBlockReason(gs, "actor", "target", ActionOfferVassalization); reason == "" {
		t.Fatal("dört bölgeli hedefe vassallık teklifi gönderimi engellenmedi")
	}
	gs.Relations[faction.RelationKey("actor", "target")].Stance = faction.StanceWar
	if reason := ActionBlockReason(gs, "actor", "target", ActionOfferVassalization); reason == "" {
		t.Fatal("savaş halindeki dört bölgeli hedefe vassallık teklifi gönderimi engellenmedi")
	}
}

func TestCancelTradeWithVassalDoesNotRequireOverlordDiplomacy(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "outside",
		Factions: map[faction.FactionID]*faction.Faction{
			"outside":  {ID: "outside", NameTR: "Dış Devlet"},
			"overlord": {ID: "overlord", NameTR: "Sahip Devlet"},
			"vassal":   {ID: "vassal", NameTR: "Vassal", OverlordID: "overlord"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("outside", "vassal"): {
				FactionA: "outside", FactionB: "vassal", Stance: faction.StanceTrade,
			},
		},
		TradeRoutes: []*economy.TradeRoute{{FromFactionID: "outside", ToFactionID: "vassal"}},
	}

	if reason := ActionBlockReason(gs, "outside", "vassal", ActionCancelTrade); reason != "" {
		t.Fatalf("vassalla ticareti bitirme engellendi: %s", reason)
	}
	result := Execute(gs, "outside", "vassal", ActionCancelTrade)
	if !result.Applied || HasTradeRouteBetween(gs, "outside", "vassal") {
		t.Fatalf("vassalla ticaret iptali uygulanmadı: applied=%v routes=%d", result.Applied, len(gs.TradeRoutes))
	}
}

func TestDeclareWarOnVassalDoesNotRequireOverlordDiplomacy(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "outside",
		Factions: map[faction.FactionID]*faction.Faction{
			"outside":  {ID: "outside", NameTR: "Dış Devlet"},
			"overlord": {ID: "overlord", NameTR: "Sahip Devlet"},
			"vassal":   {ID: "vassal", NameTR: "Vassal", OverlordID: "overlord"},
		},
	}

	if reason := ActionBlockReason(gs, "outside", "vassal", ActionDeclareWar); reason != "" {
		t.Fatalf("vassala savaş ilanı üst devlete yönlendirildi: %s", reason)
	}
	result := ExecuteWarDeclaration(gs, "outside", "vassal", nil)
	if !result.Applied {
		t.Fatalf("vassala savaş ilanı uygulanmadı: %s", result.Message)
	}
	if !IsWar(gs, "outside", "overlord") {
		t.Fatal("vassala savaş ilanı realm üst devletiyle savaş ilişkisi oluşturmadı")
	}
}

func TestNormalizeVassalagePreservesExternalVassalTrade(t *testing.T) {
	gs := &state.GameState{
		Turn:            6,
		PlayerFactionID: "venice",
		Factions: map[faction.FactionID]*faction.Faction{
			"venice":   {ID: "venice", NameTR: "Venedik"},
			"overlord": {ID: "overlord", NameTR: "Sahip Devlet"},
			"vassal":   {ID: "vassal", NameTR: "Flandre", OverlordID: "overlord", VassalizedTurn: 1},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("venice", "vassal"): {
				FactionA: "venice", FactionB: "vassal", Score: 30, Stance: faction.StanceTrade,
			},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "venice", ToFactionID: "vassal"},
			{FromFactionID: "vassal", ToFactionID: "venice"},
		},
	}

	NormalizeVassalage(gs)

	rel := Relation(gs, "venice", "vassal")
	if rel == nil || rel.Stance != faction.StanceTrade {
		t.Fatalf("vassal dış ticaret ilişkisi normalize edildi: %#v", rel)
	}
	if !HasTradeRouteBetween(gs, "venice", "vassal") {
		t.Fatal("vassal dış ticaret rotası normalize edilirken silindi")
	}
}

func TestApplyVassalizationPromotesNestedVassalsToNewOverlord(t *testing.T) {
	gs := &state.GameState{
		Turn: 7,
		Factions: map[faction.FactionID]*faction.Faction{
			"new_overlord":    {ID: "new_overlord", NameTR: "Yeni Üst Devlet"},
			"former_overlord": {ID: "former_overlord", NameTR: "Eski Üst Devlet"},
			"vassal":          {ID: "vassal", NameTR: "Vassal", OverlordID: "former_overlord", VassalizedTurn: 2},
			"nested":          {ID: "nested", NameTR: "Alt Vassal", OverlordID: "vassal", VassalizedTurn: 3},
		},
	}

	result := applyVassalization(gs, "new_overlord", "former_overlord")
	if !result.Applied {
		t.Fatalf("vassallık uygulanmadı: %s", result.Message)
	}
	for _, fid := range []faction.FactionID{"former_overlord", "vassal", "nested"} {
		f := gs.Factions[fid]
		if f.OverlordID != "new_overlord" {
			t.Fatalf("%s yeni üst devletin doğrudan vassalı olmadı: got=%s", fid, f.OverlordID)
		}
		if f.VassalizedTurn != gs.Turn {
			t.Fatalf("%s yeni vassallık turunu taşımadı: got=%d want=%d", fid, f.VassalizedTurn, gs.Turn)
		}
	}
}

func TestApplyVassalizationAddsVassalRegionsToOverextension(t *testing.T) {
	gs := &state.GameState{
		Turn:                         7,
		AggressiveExpansionLastTurns: 12,
		Factions: map[faction.FactionID]*faction.Faction{
			"overlord": {ID: "overlord", NameTR: "Üst Devlet"},
			"vassal":   {ID: "vassal", NameTR: "Vassal"},
		},
		Regions: map[world.RegionID]*world.Region{
			"vassal_a": {ID: "vassal_a", OwnerID: "vassal"},
			"vassal_b": {ID: "vassal_b", OwnerID: "vassal"},
		},
	}
	for index := 0; index < 10; index++ {
		id := world.RegionID("overlord_" + string(rune('a'+index)))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "overlord"}
	}

	result := applyVassalization(gs, "overlord", "vassal")
	if !result.Applied {
		t.Fatalf("vassallık uygulanmadı: %s", result.Message)
	}
	if got, want := gs.RecentRegionGain("overlord"), 2; got != want {
		t.Fatalf("vassal bölge kazanımı = %d, want %d", got, want)
	}
	if got, want := gs.OverextensionScore("overlord"), 40; got != want {
		t.Fatalf("vassal bölgeleri sonrası aşırı genişleme = %d, want %d", got, want)
	}
}

func TestNormalizeVassalageFlattensNestedVassals(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"overlord": {ID: "overlord"},
			"vassal":   {ID: "vassal", OverlordID: "overlord"},
			"nested":   {ID: "nested", OverlordID: "vassal"},
		},
	}

	NormalizeVassalage(gs)

	if got := gs.Factions["nested"].OverlordID; got != "overlord" {
		t.Fatalf("nested vassal kök üst devlete bağlanmadı: got=%s", got)
	}
}
