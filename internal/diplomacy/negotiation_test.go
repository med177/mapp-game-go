package diplomacy

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestTransferOfferAppliesRegionResourceArmyAndDirectionalRelation(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "a",
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", NameTR: "A", Grain: 100},
			"b": {ID: "b", NameTR: "B", Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"a_land": {ID: "a_land", OwnerID: "a", WorldX: 1, WorldY: 1},
			"b_land": {ID: "b_land", OwnerID: "b", WorldX: 2, WorldY: 1, BaseGoldIncome: 20},
			"b_rear": {ID: "b_rear", OwnerID: "b", WorldX: 3, WorldY: 1},
		},
		Armies: map[army.ArmyID]*army.Army{
			"b_army": {ID: "b_army", OwnerID: "b", RegionID: "b_land", Units: []army.Unit{
				{TypeID: "infantry", CurrentHP: army.MaxUnitHP},
				{TypeID: "infantry", CurrentHP: army.MaxUnitHP},
				{TypeID: "infantry", CurrentHP: army.MaxUnitHP},
				{TypeID: "infantry", CurrentHP: army.MaxUnitHP},
			}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", NameTR: "Piyade", Attack: 10, Defense: 10, Morale: 10},
		},
	}
	EnsureRelation(gs, "a", "b")
	requested := []state.DiplomaticTransfer{
		{Kind: transferKindRegion, ID: "b_land", Amount: 1},
		{Kind: transferKindResource, ID: "grain", Amount: 5},
		{Kind: transferKindArmy, ID: "infantry", Amount: 3},
	}
	offered := []state.DiplomaticTransfer{{Kind: transferKindResource, ID: "grain", Amount: 5}}
	if !QueueTransferOffer(gs, "a", "b", requested, offered, 50, "test") {
		t.Fatal("pazarlık kuyruğa eklenemedi")
	}
	result := ResolveOffer(gs, 0, true)
	if !result.Applied {
		t.Fatalf("pazarlık uygulanmadı: %s", result.Message)
	}
	if got := gs.Regions["b_land"].OwnerID; got != "a" {
		t.Fatalf("bölge sahibi = %q, a bekleniyordu", got)
	}
	if gs.Factions["a"].Grain != 100 || gs.Factions["b"].Grain != 100 {
		t.Fatalf("karşılıklı tahıl aktarımı netlenmedi: a=%d b=%d", gs.Factions["a"].Grain, gs.Factions["b"].Grain)
	}
	if got := unitCountForFaction(gs, "a", "infantry"); got != 3 {
		t.Fatalf("alıcı asker sayısı = %d, 3 bekleniyordu", got)
	}
	if got := unitCountForFaction(gs, "b", "infantry"); got != 1 {
		t.Fatalf("veren asker sayısı = %d, 1 bekleniyordu", got)
	}
	if got := gs.Armies["b_army"].RegionID; got != "b_rear" {
		t.Fatalf("bölgeyi veren ordunun konumu = %q, b_rear bekleniyordu", got)
	}
	if got := RelationScore(gs, "a", "b"); got != transferRelationGain {
		t.Fatalf("a->b ilişki puanı = %d, %d bekleniyordu", got, transferRelationGain)
	}
	if got := RelationScore(gs, "b", "a"); got != transferRelationGain {
		t.Fatalf("b->a ilişki puanı = %d, %d bekleniyordu", got, transferRelationGain)
	}
}

func TestQueueTransferOfferRejectsUnavailableRequestedResource(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a"},
			"b": {ID: "b", Grain: 2},
		},
		Relations: make(map[string]*faction.Relation),
	}
	EnsureRelation(gs, "a", "b")
	if QueueTransferOffer(gs, "a", "b", []state.DiplomaticTransfer{{Kind: transferKindResource, ID: "grain", Amount: 3}}, nil, 0, "") {
		t.Fatal("mevcut olmayan kaynakla pazarlık kuyruğa girdi")
	}
}

func TestExecutePlayerTransferOfferReturnsAIResourceCounterOffer(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":     {ID: "ai", Grain: 100},
			"player": {ID: "player", Gold: 100},
		},
	}
	EnsureRelation(gs, "ai", "player")
	requested := []state.DiplomaticTransfer{{Kind: transferKindResource, ID: "grain", Amount: 2}}
	offered := []state.DiplomaticTransfer{{Kind: transferKindResource, ID: "gold", Amount: 1}}
	if accepted, _ := AssessTransferOffer(gs, "player", "ai", requested, offered); accepted {
		t.Fatal("ilk şartların reddedilmesi gerekiyordu")
	}

	result := ExecuteTransferOffer(gs, "player", "ai", requested, offered)
	if result.Applied || result.Accepted {
		t.Fatalf("AI karşı teklifi transferleri hemen uygulamamalı: %+v", result)
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("bekleyen karşı teklif sayısı = %d, want 1", len(gs.DiplomaticOffers))
	}
	counter := gs.DiplomaticOffers[0]
	if counter.FromFactionID != "ai" || counter.ToFactionID != "player" || counter.PriorityReason != "AI karşı teklifi" {
		t.Fatalf("AI karşı teklifi yanlış yön/metadata ile oluşturuldu: %+v", counter)
	}
	if len(counter.RequestedTransfers) != 1 || counter.RequestedTransfers[0].ID != "gold" || counter.RequestedTransfers[0].Amount <= offered[0].Amount {
		t.Fatalf("AI karşı teklifi oyuncudan daha fazla altın istemeli: %+v", counter.RequestedTransfers)
	}
	if len(counter.OfferedTransfers) != 1 || counter.OfferedTransfers[0].ID != "grain" || counter.OfferedTransfers[0].Amount > requested[0].Amount {
		t.Fatalf("AI karşı teklifi oyuncunun istediğinden fazla tahıl sunmamalı: %+v", counter.OfferedTransfers)
	}
	if accepted, _ := AssessTransferOffer(gs, "player", "ai", counter.OfferedTransfers, counter.RequestedTransfers); !accepted {
		t.Fatalf("AI kendi karşı teklifini kabul etmiyor: %+v", counter)
	}
	if gs.Factions["ai"].Grain != 100 || gs.Factions["player"].Gold != 100 {
		t.Fatal("karşı teklif gönderilirken kaynak aktarımı erken uygulandı")
	}
}
