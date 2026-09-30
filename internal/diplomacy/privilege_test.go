package diplomacy

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestPrivilegedMinorAutomaticallyCreatesExemptTradeRoutes(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"grantor":  {ID: "grantor", NameTR: "İmtiyaz Veren"},
			"operator": {ID: "operator", NameTR: "İmtiyaz Alan"},
		},
		Regions: map[world.RegionID]*world.Region{
			"parent": {ID: "parent", OwnerID: "grantor"},
			"minor": {
				ID: "minor", OwnerID: "operator", IsMinorRegion: true,
				IsPrivileged: true, ParentRegionID: "parent",
			},
		},
	}

	EnsureTradeRoutesForActiveRelations(gs)

	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("imtiyazlı minor için rota sayısı = %d, 2 bekleniyordu", len(gs.TradeRoutes))
	}
	for _, route := range gs.TradeRoutes {
		if route == nil || !route.IsPrivilegedMinor {
			t.Fatalf("otomatik rota imtiyaz rotası olarak işaretlenmedi: %#v", route)
		}
		if route.GoldPerUnit != 0 {
			t.Fatalf("imtiyaz rotası normal birim fiyatı taşıyor: %d", route.GoldPerUnit)
		}
	}
	if got := ActiveTradePartnerCount(gs, "grantor"); got != 0 {
		t.Fatalf("imtiyaz rotası verenin partner hakkını tüketti: %d", got)
	}
	if got := ActiveTradePartnerCount(gs, "operator"); got != 0 {
		t.Fatalf("imtiyaz rotası alanın partner hakkını tüketti: %d", got)
	}
	if got := TradeRouteCapacityUsage(gs, "grantor"); got != 0 {
		t.Fatalf("imtiyaz rotası verenin rota kapasitesini tüketti: %d", got)
	}
	if got := TradeRouteCapacityUsage(gs, "operator"); got != 0 {
		t.Fatalf("imtiyaz rotası alanın rota kapasitesini tüketti: %d", got)
	}
	if !HasTradeRouteBetween(gs, "grantor", "operator") {
		t.Fatal("imtiyazlı minor devletleri arasında otomatik rota açılmadı")
	}
}

func TestRevokingPrivilegeRemovesAutomaticRouteWithoutRemovingNormalTrade(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"grantor":  {ID: "grantor", NameTR: "İmtiyaz Veren"},
			"operator": {ID: "operator", NameTR: "İmtiyaz Alan"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("grantor", "operator"): {
				FactionA: "grantor", FactionB: "operator", Stance: faction.StanceTrade,
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"parent": {ID: "parent", OwnerID: "grantor"},
			"minor": {
				ID: "minor", OwnerID: "operator", IsMinorRegion: true,
				IsPrivileged: true, ParentRegionID: "parent",
			},
		},
	}

	EnsureTradeRoutesForActiveRelations(gs)
	if len(gs.TradeRoutes) != 2 || !gs.TradeRoutes[0].IsPrivilegedMinor {
		t.Fatalf("başlangıç imtiyaz rotası kurulmadı: %#v", gs.TradeRoutes)
	}
	gs.Regions["minor"].IsPrivileged = false
	EnsurePrivilegedMinorTradeRoutes(gs)

	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("imtiyaz kaldırılınca normal rota korunmadı: %d", len(gs.TradeRoutes))
	}
	for _, route := range gs.TradeRoutes {
		if route.IsPrivilegedMinor || route.GoldPerUnit == 0 {
			t.Fatalf("normal ticaret rotası imtiyaz işaretini/fiyatını korudu: %#v", route)
		}
	}
	if got := ActiveTradePartnerCount(gs, "grantor"); got != 1 {
		t.Fatalf("normal rota imtiyaz kaldırıldıktan sonra partner sayısı = %d, 1 bekleniyordu", got)
	}
}

func TestWarAgainstPrivilegedSovereignPenalizesOperatorWithoutWar(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "venice",
		Factions: map[faction.FactionID]*faction.Faction{
			"venice":    {ID: "venice", NameTR: "Venedik"},
			"east_rome": {ID: "east_rome", NameTR: "Doğu Roma"},
			"genoa":     {ID: "genoa", NameTR: "Ceneviz"},
		},
		Relations: map[string]*faction.Relation{},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", OwnerID: "east_rome"},
			"galata": {
				ID: "galata", OwnerID: "genoa", IsMinorRegion: true,
				IsPrivileged: true, ParentRegionID: "constantinople",
			},
		},
	}

	result := ExecuteWarDeclaration(gs, "venice", "east_rome", nil)
	if !result.Applied {
		t.Fatalf("savaş ilanı uygulanmadı: %s", result.Message)
	}
	if !IsWar(gs, "venice", "east_rome") {
		t.Fatal("imtiyazlı alt bölgenin egemen sahibine savaş açılmadı")
	}
	if IsWar(gs, "venice", "genoa") {
		t.Fatal("imtiyaz kullanım sahibine doğrudan savaş açıldı")
	}
	rel := Relation(gs, "venice", "genoa")
	if rel == nil || RelationScore(gs, "venice", "genoa") != -privilegeAttackRelationPenalty || rel.Stance != faction.StancePeace {
		t.Fatalf("imtiyaz kullanım sahibine ilişki cezası yanlış: %#v", rel)
	}
}

func TestRevokeMinorPrivilegePenalizesOperatorWithoutWar(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"east_rome": {ID: "east_rome", NameTR: "Doğu Roma"},
			"genoa":     {ID: "genoa", NameTR: "Ceneviz"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("east_rome", "genoa"): {
				FactionA: "east_rome", FactionB: "genoa", ScoreAToB: 20, ScoreBToA: 20, Stance: faction.StancePeace,
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", OwnerID: "east_rome"},
			"galata": {
				ID: "galata", NameTR: "Galata", OwnerID: "genoa", IsMinorRegion: true,
				IsPrivileged: true, ParentRegionID: "constantinople",
			},
		},
	}

	result := RevokeMinorPrivilege(gs, "east_rome", "galata")
	if !result.Applied || gs.Regions["galata"].IsPrivileged {
		t.Fatalf("imtiyaz kaldırılmadı: result=%#v region=%#v", result, gs.Regions["galata"])
	}
	if rel := Relation(gs, "east_rome", "genoa"); rel == nil || RelationScore(gs, "east_rome", "genoa") != 20-PrivilegeRevocationRelationPenalty || rel.Stance != faction.StancePeace {
		t.Fatalf("imtiyaz kaldırma ilişki cezası yanlış: %#v", rel)
	}
	if IsWar(gs, "east_rome", "genoa") {
		t.Fatal("imtiyaz kaldırma kullanım sahibiyle savaş başlattı")
	}
}

func TestMinorPrivilegeOfferRequiresAcceptanceAndCreatesRouteWithRelationBonus(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "grantor",
		Turn:            7,
		Factions: map[faction.FactionID]*faction.Faction{
			"grantor":  {ID: "grantor", NameTR: "İmtiyaz Veren"},
			"operator": {ID: "operator", NameTR: "İmtiyaz Alan"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("grantor", "operator"): {
				FactionA: "grantor", FactionB: "operator", ScoreAToB: 10, ScoreBToA: 10, Stance: faction.StancePeace,
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"parent": {ID: "parent", OwnerID: "grantor"},
			"minor":  {ID: "minor", NameTR: "Minor", OwnerID: "grantor", IsMinorRegion: true, ParentRegionID: "parent"},
		},
	}

	if !QueueMinorPrivilegeOffer(gs, "grantor", "operator", "minor", 20, "test") {
		t.Fatal("imtiyaz teklifi kuyruğa alınmadı")
	}
	if len(gs.DiplomaticOffers) != 1 || gs.DiplomaticOffers[0].RegionID != "minor" {
		t.Fatalf("bölge kimliği taşınmayan teklif: %#v", gs.DiplomaticOffers)
	}
	result := ResolveOffer(gs, 0, true)
	if !result.Applied || !gs.Regions["minor"].IsPrivileged {
		t.Fatalf("kabul edilen imtiyaz uygulanmadı: %#v", result)
	}
	if got := RelationScore(gs, "grantor", "operator"); got != 25 {
		t.Fatalf("imtiyaz kabulünde ilişki skoru = %d, 25 bekleniyordu", got)
	}
	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("imtiyaz kabulünde otomatik rota çifti = %d, 2 bekleniyordu", len(gs.TradeRoutes))
	}
}

func TestMinorPrivilegeOfferEconomicAdvantageOverridesNegativeRelation(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"grantor":  {ID: "grantor", NameTR: "İmtiyaz Veren", Cloth: 10},
			"operator": {ID: "operator", NameTR: "İmtiyaz Alan"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("grantor", "operator"): {
				FactionA: "grantor", FactionB: "operator", ScoreAToB: -80, ScoreBToA: -80, Stance: faction.StancePeace,
			},
		},
		BaseGoldValues: map[economy.GoodType]int{economy.GoodCloth: 100},
		Regions: map[world.RegionID]*world.Region{
			"parent": {ID: "parent", OwnerID: "grantor"},
			"minor":  {ID: "minor", OwnerID: "grantor", IsMinorRegion: true, ParentRegionID: "parent", BaseGoldIncome: 10, TaxRate: 100, Satisfaction: 50},
		},
	}

	assessment := AssessMinorPrivilegeOffer(gs, "grantor", "operator", "minor")
	if assessment.BlockReason != "" || !assessment.EconomicAdvantage || assessment.Chance != 60 {
		t.Fatalf("ekonomik avantaj imtiyaz teklifini ezmedi: %#v", assessment)
	}
}
