package diplomacy

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestMilitaryPowerBreakdownIncludesNavalStrength(t *testing.T) {
	gs := &state.GameState{
		Armies: map[army.ArmyID]*army.Army{
			"land":  {ID: "land", OwnerID: "a", Units: []army.Unit{{}, {}}},
			"fleet": {ID: "fleet", OwnerID: "a", IsNaval: true, Units: []army.Unit{{}, {}, {}}},
		},
	}

	land, naval := MilitaryPowerBreakdown(gs, "a")
	if land != 20 || naval != 30 {
		t.Fatalf("kara/deniz gücü yanlış: land=%d naval=%d", land, naval)
	}
	if total := MilitaryPower(gs, "a"); total != 50 {
		t.Fatalf("toplam askerî güce donanma dahil olmalı, got=%d", total)
	}
}

func TestProposePeaceRejectedOutsideWar(t *testing.T) {
	gs := testGameState()

	result := Execute(gs, "a", "b", ActionProposePeace)

	if result.Accepted || result.Applied {
		t.Fatalf("barış teklifi savaş dışındayken uygulanmamalı: %+v", result)
	}
}

func TestProposeAllianceRejectedOnLowScore(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 10

	result := Execute(gs, "a", "b", ActionProposeAlliance)

	if result.Accepted || result.Applied {
		t.Fatalf("düşük skorlu ittifak reddedilmeliydi: %+v", result)
	}
	if rel.Stance != faction.StancePeace {
		t.Fatalf("stance peace kalmalı, got=%s", rel.Stance)
	}
}

func TestProposeAllianceRejectedAgainstCurrentAllyWarEnemy(t *testing.T) {
	gs := testGameState()
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic}
	ally := EnsureRelation(gs, "a", "b")
	ally.Stance = faction.StanceAllied
	ally.Score = 60
	war := EnsureRelation(gs, "b", "c")
	war.Stance = faction.StanceWar
	war.Score = -80
	target := EnsureRelation(gs, "a", "c")
	target.Score = 60

	assessment := AssessAllianceProposal(gs, target, "a", "c")
	if !strings.Contains(assessment.BlockReason, "savaş halinde olan devlete") {
		t.Fatalf("müttefikin savaş düşmanına ittifak engellenmeliydi: %+v", assessment)
	}
	if result := Execute(gs, "a", "c", ActionProposeAlliance); result.Applied {
		t.Fatalf("müttefikin savaş düşmanıyla ittifak uygulanmamalıydı: %+v", result)
	}
}

func TestQueuedAllianceOfferExpiresWhenAllyWarConflictAppears(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "c"
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic}
	enableABLandTrade(gs)
	target := EnsureRelation(gs, "a", "c")
	target.Score = 60
	if !QueueOffer(gs, "a", "c", ActionProposeAlliance) {
		t.Fatal("ittifak teklifi kuyruğa alınmalıydı")
	}
	ally := EnsureRelation(gs, "a", "b")
	ally.Stance = faction.StanceAllied
	war := EnsureRelation(gs, "b", "c")
	war.Stance = faction.StanceWar
	war.Score = -80

	result := ResolveOffer(gs, 0, true)
	if result.Applied {
		t.Fatalf("müttefikin savaşı sonradan oluşunca bekleyen ittifak uygulanmamalıydı: %+v", result)
	}
	if target.Stance != faction.StancePeace {
		t.Fatalf("geçersiz bekleyen teklif sonrası hedef ilişki barışta kalmalıydı: %s", target.Stance)
	}
}

func TestDiplomacyOfferQuotaBlocksAfterThreeDirectActions(t *testing.T) {
	gs := testGameState()
	gs.Factions["a"].Gold = 200

	for i := 0; i < 3; i++ {
		result := Execute(gs, "a", "b", ActionImproveRelations)
		if !result.Applied || !result.Accepted {
			t.Fatalf("üç teklifin ilki uygulanmalıydı, i=%d result=%+v", i, result)
		}
	}
	if got := gs.DiplomacyOfferQuotaRemaining("a"); got != 0 {
		t.Fatalf("üç teklif sonrası kalan hak 0 olmalıydı, got=%d", got)
	}

	result := Execute(gs, "a", "b", ActionImproveRelations)
	if result.Applied || result.Accepted {
		t.Fatalf("dördüncü teklif reddedilmeliydi, result=%+v", result)
	}
	if result.Message != diplomacyOfferQuotaBlockReasonTR {
		t.Fatalf("beklenen quota mesajı gelmeliydi, got=%q", result.Message)
	}
}

func TestQueueOfferWithMetaRespectsTurnQuota(t *testing.T) {
	gs := testGameState()
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic}
	gs.Factions["d"] = &faction.Faction{ID: "d", NameTR: "D", Religion: religion.Catholic}
	gs.Factions["e"] = &faction.Faction{ID: "e", NameTR: "E", Religion: religion.Catholic}

	if !QueueOfferWithMeta(gs, "a", "b", ActionProposePeace, 10, "") {
		t.Fatal("ilk teklif kuyruğa düşmeliydi")
	}
	if !QueueOfferWithMeta(gs, "a", "c", ActionProposeAlliance, 10, "") {
		t.Fatal("ikinci teklif kuyruğa düşmeliydi")
	}
	if !QueueOfferWithMeta(gs, "a", "d", ActionProposeTrade, 10, "") {
		t.Fatal("üçüncü teklif kuyruğa düşmeliydi")
	}
	if QueueOfferWithMeta(gs, "a", "e", ActionProposePeace, 10, "") {
		t.Fatal("dördüncü teklif kuyruklanmamalıydı")
	}
	if got := gs.DiplomacyOfferQuotaUsed("a"); got != 3 {
		t.Fatalf("kullanılan teklif hakkı 3 olmalıydı, got=%d", got)
	}
	if got := len(gs.DiplomaticOffers); got != 3 {
		t.Fatalf("kuyrukta 3 teklif kalmalıydı, got=%d", got)
	}
}

func TestQueueSurrenderOfferDoesNotSpendDiplomacyQuota(t *testing.T) {
	gs := testGameState()
	gs.DiplomacyOfferCounts = map[faction.FactionID]int{
		"a": state.MaxDiplomacyOffersPerTurn,
		"b": state.MaxDiplomacyOffersPerTurn,
	}
	gs.Regions["a_cap"].Neighbors = []world.RegionID{"b_cap"}
	gs.Regions["b_cap"].Neighbors = []world.RegionID{"a_cap"}
	gs.Relations[faction.RelationKey("a", "b")] = &faction.Relation{
		FactionA: "a",
		FactionB: "b",
		Stance:   faction.StanceWar,
		Score:    -80,
	}
	gs.Armies["a1"].RegionID = "b_cap"
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"b_cap": {
			RegionID:          "b_cap",
			AttackerArmyID:    "a1",
			AttackerFactionID: "a",
		},
	}

	if !QueueSurrenderOffer(gs, "a", "b", "b_cap", 155, "kuşatma baskısı") {
		t.Fatal("kuşatanın teslimiyet teklifi kotası doluyken de kuyruğa alınmalıydı")
	}
	if got := gs.DiplomacyOfferQuotaUsed("a"); got != state.MaxDiplomacyOffersPerTurn {
		t.Fatalf("kuşatanın elçi hakkı teslimiyet teklifinde değişmemeli, got=%d", got)
	}

	if !QueueSurrenderOffer(gs, "b", "a", "b_cap", 175, "savunma baskısı") {
		t.Fatal("kuşatılanın teslimiyet teklifi kotası doluyken de kuyruğa alınmalıydı")
	}
	if got := gs.DiplomacyOfferQuotaUsed("b"); got != state.MaxDiplomacyOffersPerTurn {
		t.Fatalf("kuşatılanın elçi hakkı teslimiyet teklifinde değişmemeli, got=%d", got)
	}
}

func TestRejectedSurrenderOfferCannotRepeatInSameRegionThisTurn(t *testing.T) {
	gs := testGameState()
	gs.Regions["a_cap"].Neighbors = []world.RegionID{"b_cap"}
	gs.Regions["b_cap"].Neighbors = []world.RegionID{"a_cap"}
	gs.Relations[faction.RelationKey("a", "b")] = &faction.Relation{
		FactionA: "a",
		FactionB: "b",
		Stance:   faction.StanceWar,
	}
	gs.Armies["a1"].RegionID = "b_cap"
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"b_cap": {RegionID: "b_cap", AttackerArmyID: "a1", AttackerFactionID: "a"},
	}
	gs.MarkDiplomaticOfferRejectedForRegion("a", "b", string(ActionProposeSurrender), "b_cap")

	if QueueSurrenderOffer(gs, "a", "b", "b_cap", 155, "aynı tur tekrar denemesi") {
		t.Fatal("aynı tur reddedilen bölgeye teslimiyet teklifi tekrar gönderilmemeliydi")
	}
	if got := len(gs.DiplomaticOffers); got != 0 {
		t.Fatalf("reddedilen bölge için teklif kuyruğa eklenmemeliydi, got=%d", got)
	}
}

func TestResolveQueuedAllianceOfferDoesNotSpendQuotaTwice(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "b"
	gs.DiplomacyOfferCounts = map[faction.FactionID]int{"a": 2}
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 45

	if !QueueOffer(gs, "a", "b", ActionProposeAlliance) {
		t.Fatal("ittifak teklifi kuyruğa alınmalıydı")
	}
	if got := gs.DiplomacyOfferQuotaUsed("a"); got != 3 {
		t.Fatalf("teklif kuyruğa alınırken üçüncü hak kullanılmalıydı, got=%d", got)
	}

	result := ResolveOffer(gs, 0, true)
	if !result.Accepted || !result.Applied {
		t.Fatalf("kota dolu olsa da bekleyen teklif uygulanmalıydı: %+v", result)
	}
	if got := gs.DiplomacyOfferQuotaUsed("a"); got != 3 {
		t.Fatalf("kabul sırasında teklif kotası ikinci kez harcanmamalıydı, got=%d", got)
	}
	if rel.Stance != faction.StanceAllied {
		t.Fatalf("kabul sonrası ittifak kurulmalıydı, got=%s", rel.Stance)
	}
}

func TestResolveQueuedAllianceOfferKeepsTermsAfterStrategicStateChanges(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "b"
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 25

	if !QueueOffer(gs, "a", "b", ActionProposeAlliance) {
		t.Fatal("ittifak teklifi kuyruğa alınmalıydı")
	}
	// Tekliften sonra aynı AI hazırlık akışında koşulların değişmesini taklit et.
	gs.Regions["a_cap"].Neighbors = nil
	gs.Regions["b_cap"].Neighbors = nil
	rel.Score = 10

	result := ResolveOffer(gs, 0, true)
	if !result.Accepted || !result.Applied {
		t.Fatalf("koşullar değişse de daha önce sunulmuş teklif uygulanmalıydı: %+v", result)
	}
	if rel.Stance != faction.StanceAllied {
		t.Fatalf("kabul sonrası ittifak kurulmalıydı, got=%s", rel.Stance)
	}
}

func TestProposeAllianceAcceptedDespiteDirectThreatWithCommonEnemy(t *testing.T) {
	gs := testGameState()
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic}
	gs.Regions["c_cap"] = &world.Region{ID: "c_cap", OwnerID: "c", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	gs.Relations[faction.RelationKey("a", "c")] = &faction.Relation{FactionA: "a", FactionB: "c", Stance: faction.StanceWar, Score: -80}
	gs.Relations[faction.RelationKey("b", "c")] = &faction.Relation{FactionA: "b", FactionB: "c", Stance: faction.StanceWar, Score: -80}
	gs.Regions["a_cap"].Neighbors = []world.RegionID{"b_cap"}
	gs.Regions["b_cap"].Neighbors = []world.RegionID{"a_cap"}
	gs.Armies["a1"].Units = append(gs.Armies["a1"].Units, army.Unit{TypeID: "inf", CurrentHP: 100})

	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 25

	if !HasDirectThreat(gs, "a", "b") {
		t.Fatal("test kurulumu doğrudan tehdit üretmeliydi")
	}
	assessment := AssessAllianceProposal(gs, rel, "a", "b")
	if assessment.BlockReason != "" {
		t.Fatalf("ortak düşman varken ittifak block olmamalıydı: %+v", assessment)
	}
	if !assessment.Accepted() {
		t.Fatalf("ortak düşman doğrudan tehdidi telafi etmeliydi: %+v", assessment)
	}

	result := Execute(gs, "a", "b", ActionProposeAlliance)
	if !result.Accepted || !result.Applied {
		t.Fatalf("ittifak kabul edilmeliydi: %+v", result)
	}
}

func TestProposeAllianceAcceptedWithSharedMajorThreat(t *testing.T) {
	gs := testGameState()
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic}
	gs.Regions["c_west"] = &world.Region{ID: "c_west", OwnerID: "c", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4, Neighbors: []world.RegionID{"a_cap"}}
	gs.Regions["c_east"] = &world.Region{ID: "c_east", OwnerID: "c", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4, Neighbors: []world.RegionID{"b_cap"}}
	gs.Regions["a_cap"].Neighbors = []world.RegionID{"b_cap", "c_west"}
	gs.Regions["b_cap"].Neighbors = []world.RegionID{"a_cap", "c_east"}
	gs.Armies["a1"].Units = append(gs.Armies["a1"].Units, army.Unit{TypeID: "inf", CurrentHP: 100})
	gs.Armies["c1"] = &army.Army{
		ID:       "c1",
		OwnerID:  "c",
		RegionID: "c_west",
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}

	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 25

	if !HasSharedMajorThreat(gs, "a", "b") {
		t.Fatal("test kurulumu ortak büyük tehdit üretmeliydi")
	}
	assessment := AssessAllianceProposal(gs, rel, "a", "b")
	if assessment.BlockReason != "" {
		t.Fatalf("ortak büyük tehdit varken ittifak block olmamalıydı: %+v", assessment)
	}
	if !assessment.Accepted() {
		t.Fatalf("ortak büyük tehdit doğrudan tehdidi telafi etmeliydi: %+v", assessment)
	}
}

func TestProposeAllianceRejectedWithoutStrategicBasis(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 28

	assessment := AssessAllianceProposal(gs, rel, "a", "b")
	if assessment.BlockReason != "İttifak için coğrafi veya stratejik yakınlık yok" {
		t.Fatalf("uzak ve ilgisiz devletler engellenmeliydi, got=%+v", assessment)
	}

	result := Execute(gs, "a", "b", ActionProposeAlliance)
	if result.Accepted || result.Applied {
		t.Fatalf("stratejik taban yokken ittifak kurulmamaliydi: %+v", result)
	}
}

func TestAllianceReligionAffinityBonusPrefersSameFaith(t *testing.T) {
	sameFaith := allianceReligionAffinityBonus(religion.Catholic, religion.Catholic)
	crossFaith := allianceReligionAffinityBonus(religion.Catholic, religion.Sunni)
	hostileFaith := allianceReligionAffinityBonus(religion.Sunni, religion.Shia)

	if sameFaith <= crossFaith {
		t.Fatalf("aynı din bonusu farklı dinden yüksek olmalıydı, same=%d cross=%d", sameFaith, crossFaith)
	}
	if hostileFaith >= crossFaith {
		t.Fatalf("sert mezhep ayrımı en düşük affinity olmalıydı, hostile=%d cross=%d", hostileFaith, crossFaith)
	}
}

func TestProposeTradeRejectedDuringWar(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -80

	result := Execute(gs, "a", "b", ActionProposeTrade)

	if result.Accepted || result.Applied {
		t.Fatalf("savaşta ticaret reddedilmeliydi: %+v", result)
	}
}

func TestTradeCreatesUniqueRoutesAndWarRemovesThem(t *testing.T) {
	gs := testGameState()
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 15

	result := Execute(gs, "a", "b", ActionProposeTrade)
	if !result.Accepted || !result.Applied {
		t.Fatalf("ticaret kabul edilmeliydi: %+v", result)
	}
	if rel.Stance != faction.StanceTrade {
		t.Fatalf("stance trade olmalı, got=%s", rel.Stance)
	}
	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("iki yönlü 2 rota bekleniyordu, got=%d", len(gs.TradeRoutes))
	}

	result = Execute(gs, "a", "b", ActionProposeTrade)
	if result.Accepted || result.Applied {
		t.Fatalf("tekrar ticaret aynı rotaları çoğaltmamalı: %+v", result)
	}
	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("rota sayısı 2 kalmalı, got=%d", len(gs.TradeRoutes))
	}

	result = Execute(gs, "a", "b", ActionDeclareWar)
	if !result.Accepted || !result.Applied {
		t.Fatalf("savaş ilanı uygulanmalı: %+v", result)
	}
	if len(gs.TradeRoutes) != 0 {
		t.Fatalf("savaşta ticaret yolları kapanmalı, got=%d", len(gs.TradeRoutes))
	}
}

func TestProposeTradeWhileAlliedKeepsAllianceAndAddsRoutes(t *testing.T) {
	gs := testGameState()
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceAllied
	rel.Score = 30

	result := Execute(gs, "a", "b", ActionProposeTrade)
	if !result.Accepted || !result.Applied {
		t.Fatalf("müttefikle ticaret kabul edilmeliydi: %+v", result)
	}
	if rel.Stance != faction.StanceAllied {
		t.Fatalf("müttefiklik korunmalıydı, got=%s", rel.Stance)
	}
	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("iki yönlü 2 rota bekleniyordu, got=%d", len(gs.TradeRoutes))
	}
}

func TestCancelAlliancePreservesTradeAgreement(t *testing.T) {
	gs := testGameState()
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceAllied
	rel.Score = 60
	ensureTradeRoutesBetween(gs, "a", "b")

	result := Execute(gs, "a", "b", ActionCancelAlliance)

	if !result.Accepted || !result.Applied {
		t.Fatalf("ittifak iptali uygulanmalıydı: %+v", result)
	}
	if rel.Stance != faction.StanceTrade {
		t.Fatalf("ticaret sürerken ilişki trade durumuna inmeli, got=%s", rel.Stance)
	}
	if !HasTradeRouteBetween(gs, "a", "b") || len(gs.TradeRoutes) != 2 {
		t.Fatalf("ittifak iptali ticaret rotalarını korumalı, got=%+v", gs.TradeRoutes)
	}
}

func TestCancelTradePreservesAlliance(t *testing.T) {
	gs := testGameState()
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceAllied
	rel.Score = 60
	ensureTradeRoutesBetween(gs, "a", "b")

	result := Execute(gs, "a", "b", ActionCancelTrade)

	if !result.Accepted || !result.Applied {
		t.Fatalf("ticaret iptali uygulanmalıydı: %+v", result)
	}
	if rel.Stance != faction.StanceAllied {
		t.Fatalf("ticaret iptali ittifakı korumalı, got=%s", rel.Stance)
	}
	if HasTradeRouteBetween(gs, "a", "b") || len(gs.TradeRoutes) != 0 {
		t.Fatalf("ticaret rotaları kaldırılmalı, got=%+v", gs.TradeRoutes)
	}
}

func TestVassalInternalAgreementsCannotBeCancelledDirectly(t *testing.T) {
	gs := testGameState()
	gs.Factions["b"].OverlordID = "a"
	NormalizeVassalage(gs)

	if reason := ActionBlockReason(gs, "a", "b", ActionCancelAlliance); reason == "" {
		t.Fatal("vassal bağı ittifak iptaliyle kaldırılamamalı")
	}
	if reason := ActionBlockReason(gs, "a", "b", ActionCancelTrade); reason == "" {
		t.Fatal("zorunlu vassal ticareti doğrudan iptal edilememeli")
	}
}

func TestProposeTradeAcceptedDespiteDirectThreatWhenOverallChanceHigh(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 30
	gs.Regions["a_cap"].Neighbors = []world.RegionID{"b_cap"}
	gs.Regions["b_cap"].Neighbors = []world.RegionID{"a_cap"}
	gs.Armies["a1"].Units = append(gs.Armies["a1"].Units, army.Unit{TypeID: "inf", CurrentHP: 100})

	if !HasDirectThreat(gs, "a", "b") {
		t.Fatal("test kurulumu doğrudan tehdit üretmeliydi")
	}
	assessment := AssessTradeProposal(gs, rel, "a", "b")
	if assessment.BlockReason != "" {
		t.Fatalf("doğrudan tehdit artık sert engel olmamalıydı: %+v", assessment)
	}
	if !assessment.Accepted() {
		t.Fatalf("yüksek toplam puanda ticaret kabul edilebilir olmalıydı: %+v", assessment)
	}

	result := Execute(gs, "a", "b", ActionProposeTrade)
	if !result.Accepted || !result.Applied {
		t.Fatalf("ticaret kabul edilmeliydi: %+v", result)
	}
}

func TestForceRelationToAlliancePreservesExistingTradeRoutes(t *testing.T) {
	gs := testGameState()
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 15
	if result := Execute(gs, "a", "b", ActionProposeTrade); !result.Accepted || !result.Applied {
		t.Fatalf("önce ticaret açılmalıydı: %+v", result)
	}

	ForceRelation(gs, "a", "b", faction.StanceAllied, 5)

	if rel.Stance != faction.StanceAllied {
		t.Fatalf("stance allied olmalıydı, got=%s", rel.Stance)
	}
	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("ittifakta mevcut rotalar korunmalıydı, got=%d", len(gs.TradeRoutes))
	}
}

func TestEnsureTradeRoutesForActiveRelationsBuildsMissingRoutes(t *testing.T) {
	gs := testGameState()
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceTrade
	rel.Score = 25
	gs.TradeRoutes = nil

	EnsureTradeRoutesForActiveRelations(gs)

	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("trade stance için iki yönlü 2 rota kurulmalıydı, got=%d", len(gs.TradeRoutes))
	}
}

func TestEnsureTradeRoutesForActiveRelationsEnforcesPartnerLimit(t *testing.T) {
	gs := testGameState()
	for _, partnerID := range []faction.FactionID{"p1", "p2", "p3", "p4", "p5"} {
		regionID := world.RegionID(partnerID + "_cap")
		gs.Factions[partnerID] = &faction.Faction{ID: partnerID, NameTR: string(partnerID), Religion: religion.Catholic}
		gs.Regions[regionID] = &world.Region{
			ID:            regionID,
			OwnerID:       string(partnerID),
			TaxRate:       50,
			Satisfaction:  50,
			TradeCapacity: 4,
			Neighbors:     []world.RegionID{"a_cap"},
		}
		gs.Regions["a_cap"].Neighbors = append(gs.Regions["a_cap"].Neighbors, regionID)
		rel := EnsureRelation(gs, "a", partnerID)
		rel.Stance = faction.StanceTrade
		rel.Score = 25
	}

	EnsureTradeRoutesForActiveRelations(gs)

	if got := ActiveTradePartnerCount(gs, "a"); got != MaxTradePartners {
		t.Fatalf("a için partner limiti %d olmalı, got=%d", MaxTradePartners, got)
	}
	if HasTradeRouteBetween(gs, "a", "p5") {
		t.Fatalf("deterministik sınırdan sonra beşinci partner için rota kurulmamış olmalı")
	}
	if len(gs.TradeRoutes) != MaxTradePartners*2 {
		t.Fatalf("partner limiti iki yönlü rota sayısını sınırlandırmalı, got=%d", len(gs.TradeRoutes))
	}
}

func TestRebalanceTradeRouteCapacitiesSharesEffectiveCapacity(t *testing.T) {
	gs := testGameState()
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic}
	gs.Regions["a_cap"].TradeCapacity = 5
	gs.Regions["b_cap"].TradeCapacity = 4
	gs.Regions["c_cap"] = &world.Region{ID: "c_cap", OwnerID: "c", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}

	ensureTradeRoutesBetween(gs, "a", "b")
	ensureTradeRoutesBetween(gs, "a", "c")

	if got := tradeAgreementAmountForTest(gs, "a", "b"); got != 3 {
		t.Fatalf("5 kapasitenin ilk partner payı 3 olmalı, got=%d", got)
	}
	if got := tradeAgreementAmountForTest(gs, "a", "c"); got != 2 {
		t.Fatalf("5 kapasitenin ikinci partner payı 2 olmalı, got=%d", got)
	}
	if got := TradeRouteCapacityUsage(gs, "a"); got != 5 {
		t.Fatalf("rota kullanımı efektif kapasiteyi tüketmeli, got=%d", got)
	}

	gs.Regions["a_cap"].TradeCapacity = 2
	RebalanceTradeRouteCapacities(gs)
	if got := tradeAgreementAmountForTest(gs, "a", "b"); got != 1 {
		t.Fatalf("kapasite düştüğünde ilk rota 1 olmalı, got=%d", got)
	}
	if got := tradeAgreementAmountForTest(gs, "a", "c"); got != 1 {
		t.Fatalf("kapasite düştüğünde ikinci rota 1 olmalı, got=%d", got)
	}
}

func tradeAgreementAmountForTest(gs *state.GameState, a, b faction.FactionID) int {
	for _, route := range gs.TradeRoutes {
		if route == nil {
			continue
		}
		if (route.FromFactionID == string(a) && route.ToFactionID == string(b)) ||
			(route.FromFactionID == string(b) && route.ToFactionID == string(a)) {
			return route.AmountPerTurn
		}
	}
	return -1
}

func TestBuildTradeRoutePrioritizesGrainForStrategicDemand(t *testing.T) {
	gs := testGameState()
	gs.GrainEconomy = map[faction.FactionID]state.GrainEconomyStatus{
		"a": {FactionID: "a", StorageCapacity: 100},
		"b": {FactionID: "b", TotalDemand: 40},
	}

	route := buildTradeRoute(gs, "a", "b")
	if route.Good != economy.GoodGrain {
		t.Fatalf("rezerv açığı olan hedef için kaynak tahıl rotası kurmalıydı, route=%+v", route)
	}
}

func TestEnsureTradeRoutesForActiveRelationsRemovesStaleEliminatedRoutes(t *testing.T) {
	gs := testGameState()
	enableABLandTrade(gs)
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic, IsEliminated: true}
	gs.TradeRoutes = []*economy.TradeRoute{
		{FromFactionID: "a", ToFactionID: "c", Good: economy.GoodCloth, AmountPerTurn: 2, GoldPerUnit: 8},
		{FromFactionID: "c", ToFactionID: "a", Good: economy.GoodSpice, AmountPerTurn: 2, GoldPerUnit: 12},
	}

	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceTrade
	rel.Score = 25

	EnsureTradeRoutesForActiveRelations(gs)

	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("stale rota temizlenip a-b için iki yönlü rota kurulmalıydı, got=%d", len(gs.TradeRoutes))
	}
	for _, route := range gs.TradeRoutes {
		if route.FromFactionID == "c" || route.ToFactionID == "c" {
			t.Fatalf("elenmiş fraksiyon rotası kalmamalıydı, got=%+v", route)
		}
	}
}

func TestProposePeaceAcceptedUnderWarPressure(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -100
	gs.Factions["b"].Gold = 40

	result := Execute(gs, "a", "b", ActionProposePeace)

	if !result.Accepted || !result.Applied {
		t.Fatalf("yüksek savaş baskısında barış kabul edilmeliydi: %+v", result)
	}
	if rel.Stance != faction.StancePeace || rel.Score != -20 {
		t.Fatalf("barış sonrası ilişki güncellenmedi: %+v", rel)
	}
}

func TestAcceptedPeaceEvacuatesNavalLandingSiege(t *testing.T) {
	gs := testGameState()
	gs.Regions["b_cap"].WorldX = 100
	gs.Regions["b_cap"].WorldY = 0
	gs.Regions["a_cap"].WorldX = 0
	gs.Regions["a_cap"].WorldY = 100
	gs.Regions["nearland"] = &world.Region{ID: "nearland", OwnerID: "a", WorldX: 110, WorldY: 0}
	gs.Regions["nearsea"] = &world.Region{ID: "nearsea", IsSea: true, WorldX: 100, WorldY: 20}
	gs.Armies["a1"].RegionID = "b_cap"
	gs.Armies["fleet"] = &army.Army{
		ID: "fleet", OwnerID: "a", RegionID: "nearsea", IsNaval: true,
		Units: []army.Unit{{TypeID: "transport"}},
	}
	gs.UnitTypes["transport"] = &army.UnitType{ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"b_cap": {RegionID: "b_cap", AttackerArmyID: "a1", AttackerFactionID: "a", NavalLanding: true},
	}
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -100

	result := Execute(gs, "a", "b", ActionProposePeace)
	if !result.Applied {
		t.Fatalf("barış uygulanmalıydı: %+v", result)
	}
	if _, ok := gs.Armies["a1"]; ok || len(gs.Armies["fleet"].EmbarkedUnits) != 1 {
		t.Fatalf("barış sonrası çıkarma ordusu en yakın nakliye filosuna binmeliydi: armies=%+v", gs.Armies)
	}
}

func TestAcceptedPeaceEvacuatesLandArmyRegardlessOfMovePoints(t *testing.T) {
	gs := testGameState()
	gs.Regions["a_cap"].WorldX = 0
	gs.Regions["a_cap"].WorldY = 0
	gs.Regions["b_cap"].WorldX = 100
	gs.Regions["b_cap"].WorldY = 0
	gs.Regions["vassal_land"] = &world.Region{ID: "vassal_land", OwnerID: "vassal", WorldX: 90, WorldY: 0}
	gs.Factions["vassal"] = &faction.Faction{ID: "vassal", NameTR: "Vassal", Religion: religion.Catholic, OverlordID: "a"}
	gs.Armies["a1"].RegionID = "b_cap"
	gs.Armies["a1"].MovePoints = 0
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"b_cap": {RegionID: "b_cap", AttackerArmyID: "a1", AttackerFactionID: "a"},
	}
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -100

	result := Execute(gs, "a", "b", ActionProposePeace)
	if !result.Applied {
		t.Fatalf("barış uygulanmalıydı: %+v", result)
	}
	if got := gs.Armies["a1"].RegionID; got != "vassal_land" {
		t.Fatalf("hareket puanı olmasa da ordu en yakın vassal bölgesine çekilmeli: %s", got)
	}
	if gs.SiegeAt("b_cap") != nil {
		t.Fatal("barışla terk edilen düşman bölgesindeki kuşatma kaldırılmalı")
	}
}

func TestAcceptedPeaceCreatesTemporaryTruce(t *testing.T) {
	gs := testGameState()
	gs.Turn = 10
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -100
	gs.BeginWarLedger("a", "b")
	gs.Factions["b"].Gold = 40

	if result := Execute(gs, "a", "b", ActionProposePeace); !result.Applied {
		t.Fatalf("barış uygulanmalıydı: %+v", result)
	}
	if remaining := gs.TruceRemaining("a", "b"); remaining != 6 {
		t.Fatalf("barış sonrası altı turluk ateşkes bekleniyordu: remaining=%d", remaining)
	}
	if result := Execute(gs, "a", "b", ActionDeclareWar); result.Applied {
		t.Fatalf("ateşkes sürerken yeniden savaş ilan edilmemeliydi: %+v", result)
	}

	gs.Turn += 6
	if result := Execute(gs, "a", "b", ActionDeclareWar); !result.Applied {
		t.Fatalf("ateşkes bitince savaş ilanı yeniden açılmalıydı: %+v", result)
	}
}

func TestProposePeaceAcceptedWithPeaceTechBonus(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -88
	gs.Factions["a"].Research.Completed = map[string]bool{"diplomacy": true}
	gs.TechTypes = map[string]*tech.Technology{
		"diplomacy": {ID: "diplomacy", Effects: tech.Effects{PeaceRelationBonus: 10}},
	}

	result := Execute(gs, "a", "b", ActionProposePeace)

	if !result.Accepted || !result.Applied {
		t.Fatalf("barış tech bonusu kabul eşiğini geçirmeliydi: %+v", result)
	}
}

func TestQueueAndResolveOfferForPlayer(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "b"
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -100
	gs.Factions["b"].Gold = 40

	if !QueueOffer(gs, "a", "b", ActionProposePeace) {
		t.Fatal("teklif kuyruğa alınmalıydı")
	}
	if QueueOffer(gs, "a", "b", ActionProposePeace) {
		t.Fatal("aynı teklif ikinci kez kuyruğa alınmamalı")
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("tek bekleyen teklif bekleniyordu, got=%d", len(gs.DiplomaticOffers))
	}

	result := ResolveOffer(gs, 0, true)
	if !result.Accepted || !result.Applied {
		t.Fatalf("kabul edilen teklif uygulanmalıydı: %+v", result)
	}
	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("teklif kabul sonrası kuyruktan düşmeli, got=%d", len(gs.DiplomaticOffers))
	}
	if rel.Stance != faction.StancePeace {
		t.Fatalf("barış sonrası stance peace olmalı, got=%s", rel.Stance)
	}
}

func TestResolvePeaceOfferAcceptedByPlayerAlwaysApplies(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "b"
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -81
	gs.Factions["b"].Gold = 500

	if !QueueOffer(gs, "a", "b", ActionProposePeace) {
		t.Fatal("barış teklifi kuyruğa alınmalıydı")
	}
	result := ResolveOffer(gs, 0, true)
	if !result.Accepted || !result.Applied {
		t.Fatalf("oyuncu kabul ettiğinde barış kesin uygulanmalıydı: %+v", result)
	}
	if rel.Stance != faction.StancePeace || rel.Score != -20 {
		t.Fatalf("barış kabulünde relation peace/-20 olmalıydı: %+v", rel)
	}
}

func TestPeaceOfferPrecedesSiegeSurrenderAndAcceptanceSkipsIt(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "b"
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	gs.DiplomaticOffers = []state.DiplomaticOffer{
		{FromFactionID: "a", ToFactionID: "b", Action: string(ActionProposeSurrender), RegionID: "b_cap", Priority: 175, CreatedTurn: 5},
		{FromFactionID: "a", ToFactionID: "b", Action: string(ActionProposePeace), Priority: 10, CreatedTurn: 5},
	}

	bestIndex, ok := BestOfferIndex(gs, "b")
	if !ok || gs.DiplomaticOffers[bestIndex].Action != string(ActionProposePeace) {
		t.Fatalf("barış teklifi kuşatma teslimiyetinden önce seçilmeliydi: index=%d offers=%+v", bestIndex, gs.DiplomaticOffers)
	}

	result := ResolveOffer(gs, bestIndex, true)
	if !result.Applied || rel.Stance != faction.StancePeace {
		t.Fatalf("barış kabulü uygulanmalıydı: result=%+v relation=%+v", result, rel)
	}
	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("barış kabulünden sonra aynı savaşa ait teslimiyet teklifi atlanmalıydı: %+v", gs.DiplomaticOffers)
	}
}

func TestResolveRejectedDiplomaticOfferLowersRelationAndRecordsRetry(t *testing.T) {
	gs := testGameState()
	gs.Turn = 5
	gs.PlayerFactionID = "b"
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 40

	if !QueueOffer(gs, "a", "b", ActionProposeAlliance) {
		t.Fatal("ittifak teklifi kuyruğa alınmalıydı")
	}
	result := ResolveOffer(gs, 0, false)
	if result.Accepted || result.Applied {
		t.Fatalf("reddedilen teklif kabul edilmiş görünmemeli: %+v", result)
	}
	if rel.Score != 37 {
		t.Fatalf("ret ilişkiyi 3 puan düşürmeliydi, got=%d", rel.Score)
	}
	if !gs.DiplomaticOfferRetryBlocked("a", "b", string(ActionProposeAlliance), state.DiplomaticOfferRetryCooldownTurns) {
		t.Fatal("ret sonrası teklif retry cooldown'u aktif olmalıydı")
	}
}

func TestAssessTradeProposalBlocksLowScore(t *testing.T) {
	gs := testGameState()
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 5

	assessment := AssessTradeProposal(gs, rel, "a", "b")
	if assessment.BlockReason != "Ticaret için ilişki puanı 15 altı" {
		t.Fatalf("beklenen düşük skor engeli, got=%+v", assessment)
	}
	if assessment.Accepted() {
		t.Fatalf("düşük skor kabul edilmemeli: %+v", assessment)
	}
}

func TestAssessTradeProposalBlocksMissingRoute(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 30

	assessment := AssessTradeProposal(gs, rel, "a", "b")
	if assessment.BlockReason != "Ticaret için bağlanabilir kara veya deniz hattı yok" {
		t.Fatalf("baglanti yoksa trade block olmaliydi, got=%+v", assessment)
	}
}

func TestImproveRelationsConsumesGoldAndRaisesScore(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 12
	gs.Factions["a"].Gold = 100

	result := Execute(gs, "a", "b", ActionImproveRelations)

	if !result.Accepted || !result.Applied {
		t.Fatalf("heyet gönderimi uygulanmalıydı: %+v", result)
	}
	if got := gs.Factions["a"].Gold; got != 60 {
		t.Fatalf("heyet maliyeti altından düşmeliydi, got=%d", got)
	}
	if got := rel.Score; got != 20 {
		t.Fatalf("ilişki +8 artmalıydı, got=%d", got)
	}
}

func TestApplyRelationDecayErodesUnsupportedAlliance(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceAllied
	rel.Score = 42

	ApplyRelationDecay(gs)

	if got := rel.Score; got != 41 {
		t.Fatalf("desteksiz ittifak yavaşça aşınmalıydı, got=%d", got)
	}
}

func TestOfferVassalizationMakesTargetVassalAndBlocksThirdParties(t *testing.T) {
	gs := testGameState()
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic, Gold: 90, Grain: 60}
	gs.Regions["c_cap"] = &world.Region{ID: "c_cap", OwnerID: "c", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	for _, regionID := range []world.RegionID{"a_frontier_1", "a_frontier_2", "a_frontier_3", "a_frontier_4"} {
		gs.Regions[regionID] = &world.Region{ID: regionID, OwnerID: "a", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	}
	gs.Relations[faction.RelationKey("a", "b")] = &faction.Relation{FactionA: "a", FactionB: "b", Stance: faction.StancePeace, Score: 70}
	gs.Relations[faction.RelationKey("b", "c")] = &faction.Relation{FactionA: "b", FactionB: "c", Stance: faction.StanceTrade, Score: 35}
	gs.TradeRoutes = []*economy.TradeRoute{
		{FromFactionID: "b", ToFactionID: "c", Good: economy.GoodCloth, AmountPerTurn: 2, GoldPerUnit: 8},
		{FromFactionID: "c", ToFactionID: "b", Good: economy.GoodGrain, AmountPerTurn: 2, GoldPerUnit: 5},
	}
	gs.Armies["a1"].Units = append(gs.Armies["a1"].Units,
		army.Unit{TypeID: "inf", CurrentHP: 100},
		army.Unit{TypeID: "inf", CurrentHP: 100},
		army.Unit{TypeID: "inf", CurrentHP: 100},
		army.Unit{TypeID: "inf", CurrentHP: 100},
	)

	result := Execute(gs, "a", "b", ActionOfferVassalization)

	if !result.Accepted || !result.Applied {
		t.Fatalf("vassallık kabul edilmeliydi: %+v", result)
	}
	if got := gs.Factions["b"].OverlordID; got != "a" {
		t.Fatalf("hedef a'ya bağlanmalıydı, got=%q", got)
	}
	if rel := Relation(gs, "a", "b"); rel == nil || rel.Stance != faction.StanceAllied {
		t.Fatalf("overlord-vassal ilişkisi allied olmalıydı, got=%+v", rel)
	}
	if len(gs.TradeRoutes) != 2 || !HasTradeRouteBetween(gs, "a", "b") {
		t.Fatalf("vassalın dış rotaları kapanıp overlord ile iki yönlü ticaret açılmalıydı, got=%+v", gs.TradeRoutes)
	}
	if reason := ActionBlockReason(gs, "c", "b", ActionProposeAlliance); reason == "" {
		t.Fatal("üçüncü tarafın vassalla doğrudan diplomasi kurması engellenmeliydi")
	}
	if reason := ActionBlockReason(gs, "b", "c", ActionProposeTrade); reason == "" {
		t.Fatal("vassalın üçüncü tarafla diplomasi kurması engellenmeliydi")
	}
}

func TestAssessVassalizationDoesNotAcceptRelationScoreAlone(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 55

	assessment := AssessVassalizationProposal(gs, rel, "a", "b")
	if assessment.BlockReason == "" || assessment.Accepted() {
		t.Fatalf("55 ilişki skoru tek başına vassallık kabulü sağlamamalıydı: %+v", assessment)
	}
	if result := Execute(gs, "a", "b", ActionOfferVassalization); result.Applied {
		t.Fatalf("tehdit/üstünlük yokken vassallık uygulanmamalıydı: %+v", result)
	}
}

func TestAssessVassalizationRequiresFivefoldMilitaryAndRegionalSuperiority(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 55
	gs.Armies["a1"].Units = append(gs.Armies["a1"].Units,
		army.Unit{TypeID: "inf", CurrentHP: 100},
		army.Unit{TypeID: "inf", CurrentHP: 100},
		army.Unit{TypeID: "inf", CurrentHP: 100},
	)
	for _, regionID := range []world.RegionID{"a_extra_1", "a_extra_2", "a_extra_3"} {
		gs.Regions[regionID] = &world.Region{ID: regionID, OwnerID: "a", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	}

	assessment := AssessVassalizationProposal(gs, rel, "a", "b")
	if assessment.BlockReason == "" || assessment.Accepted() {
		t.Fatalf("5 katın altındaki askeri veya bölgesel üstünlük vassallık için yeterli olmamalıydı: %+v", assessment)
	}
}

func TestAssessVassalizationAcceptsDirectFrontierThreatWithoutRegionalSuperiority(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 55
	gs.Regions["a_cap"].Neighbors = []world.RegionID{"b_cap"}
	gs.Regions["b_cap"].Neighbors = []world.RegionID{"a_cap"}
	gs.Regions["b_rear"] = &world.Region{ID: "b_rear", OwnerID: "b", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	gs.Armies["a2"] = &army.Army{
		ID:       "a2",
		OwnerID:  "a",
		RegionID: "a_cap",
		Units:    []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}
	gs.Armies["a3"] = &army.Army{
		ID:       "a3",
		OwnerID:  "a",
		RegionID: "a_cap",
		Units:    []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}
	gs.Armies["b2"] = &army.Army{
		ID:       "b2",
		OwnerID:  "b",
		RegionID: "b_rear",
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
		},
	}

	assessment := AssessVassalizationProposal(gs, rel, "a", "b")
	if !assessment.Accepted() {
		t.Fatalf("doğrudan sınır tehdidi varken vassallık kabul edilebilmeliydi: %+v", assessment)
	}
}

func TestForceVassalizeAfterWarEndsWarAndPropagatesOverlordWars(t *testing.T) {
	gs := testGameState()
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic, Gold: 90, Grain: 60}
	gs.Regions["c_cap"] = &world.Region{ID: "c_cap", OwnerID: "c", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	gs.Relations[faction.RelationKey("a", "b")] = &faction.Relation{FactionA: "a", FactionB: "b", Stance: faction.StanceWar, Score: -80}
	gs.Relations[faction.RelationKey("a", "c")] = &faction.Relation{FactionA: "a", FactionB: "c", Stance: faction.StanceWar, Score: -80}
	gs.Relations[faction.RelationKey("b", "c")] = &faction.Relation{FactionA: "b", FactionB: "c", Stance: faction.StanceTrade, Score: 25}
	gs.TradeRoutes = []*economy.TradeRoute{
		{FromFactionID: "b", ToFactionID: "c", Good: economy.GoodCloth, AmountPerTurn: 2, GoldPerUnit: 8},
	}

	result := ForceVassalizeAfterWar(gs, "a", "b")

	if !result.Accepted || !result.Applied {
		t.Fatalf("savaş sonrası vassallık uygulanmalıydı: %+v", result)
	}
	if got := gs.Factions["b"].OverlordID; got != "a" {
		t.Fatalf("hedef a'ya bağlanmalıydı, got=%q", got)
	}
	if rel := Relation(gs, "a", "b"); rel == nil || rel.Stance != faction.StanceAllied {
		t.Fatalf("savaş sonrası ilişki allied olmalıydı, got=%+v", rel)
	}
	if rel := Relation(gs, "b", "c"); rel == nil || rel.Stance != faction.StanceWar {
		t.Fatalf("vassal overlord'un savaşına girmeliydi, got=%+v", rel)
	}
	if len(gs.TradeRoutes) != 2 || !HasTradeRouteBetween(gs, "a", "b") {
		t.Fatalf("vassalın dış rotaları kapanıp overlord ticareti açılmalıydı, got=%+v", gs.TradeRoutes)
	}
}

func TestReleaseVassalRestoresIndependenceAndKeepsTrade(t *testing.T) {
	gs := testGameState()
	gs.Factions["b"].OverlordID = "a"
	NormalizeVassalage(gs)

	result := Execute(gs, "a", "b", ActionReleaseVassal)

	if !result.Accepted || !result.Applied {
		t.Fatalf("vasallık sona erdirilmeliydi: %+v", result)
	}
	if got := gs.Factions["b"].OverlordID; got != "" {
		t.Fatalf("hedef bağımsız olmalıydı, got overlord=%q", got)
	}
	if rel := Relation(gs, "a", "b"); rel == nil || rel.Stance != faction.StanceTrade {
		t.Fatalf("bağımsızlık sonrası ticaret duruşu korunmalıydı, got=%+v", rel)
	}
	if !HasTradeRouteBetween(gs, "a", "b") || len(gs.TradeRoutes) != 2 {
		t.Fatalf("iki yönlü ticaret rotası devam etmeliydi, got=%+v", gs.TradeRoutes)
	}
}

func TestVassalManagementActionsRequireDirectOverlord(t *testing.T) {
	gs := testGameState()
	gs.Turn = 20
	gs.Factions["b"].OverlordID = "a"
	gs.Factions["b"].VassalizedTurn = 1
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic}

	if reason := ActionBlockReason(gs, "a", "b", ActionAnnexVassal); reason != "" {
		t.Fatalf("doğrudan overlord ilhak edebilmeli, got=%q", reason)
	}
	if reason := ActionBlockReason(gs, "c", "b", ActionAnnexVassal); reason == "" {
		t.Fatal("üçüncü taraf vassalı ilhak edememeli")
	}
}

func TestAnnexVassalRequiresTwelveCompletedTurns(t *testing.T) {
	gs := testGameState()
	gs.Turn = 20
	gs.Factions["b"].OverlordID = "a"
	gs.Factions["b"].VassalizedTurn = 9

	if reason := ActionBlockReason(gs, "a", "b", ActionAnnexVassal); reason == "" {
		t.Fatal("vassallığın 11. turunda ilhak engellenmeliydi")
	}

	gs.Turn = 21
	if reason := ActionBlockReason(gs, "a", "b", ActionAnnexVassal); reason != "" {
		t.Fatalf("12 tamamlanmış turdan sonra ilhak açılmalıydı: %q", reason)
	}
}

func TestVassalizationStoresStartTurnForAnnexationCooldown(t *testing.T) {
	gs := testGameState()
	gs.Turn = 7
	gs.Factions["b"].OverlordID = ""
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 100
	gs.Factions["a"].Gold = 1000
	gs.Factions["b"].Gold = 0
	gs.Factions["a"].Grain = 1000

	result := applyVassalization(gs, "a", "b")
	if !result.Applied {
		t.Fatalf("vassallık uygulanmalıydı: %+v", result)
	}
	if got := gs.Factions["b"].VassalizedTurn; got != 7 {
		t.Fatalf("vassallık başlangıç turu kaydedilmedi: got=%d", got)
	}
}

func TestNormalizeVassalageBackfillsStartTurnForExistingVassal(t *testing.T) {
	gs := testGameState()
	gs.Turn = 4
	gs.Factions["b"].OverlordID = "a"

	NormalizeVassalage(gs)

	if got := gs.Factions["b"].VassalizedTurn; got != 4 {
		t.Fatalf("mevcut vassalın başlangıç turu normalize edilmeliydi: got=%d", got)
	}
	if reason := ActionBlockReason(gs, "a", "b", ActionAnnexVassal); reason == "" {
		t.Fatal("normalize edilen vassal 12 tur dolmadan ilhak edilememeliydi")
	}
}

func TestDeclareWarPropagatesToVassalCoalitions(t *testing.T) {
	gs := testGameState()
	gs.Factions["a_v"] = &faction.Faction{ID: "a_v", NameTR: "A Vassal", Religion: religion.Catholic, OverlordID: "a"}
	gs.Factions["b_v"] = &faction.Faction{ID: "b_v", NameTR: "B Vassal", Religion: religion.Catholic, OverlordID: "b"}
	gs.Regions["a_v_cap"] = &world.Region{ID: "a_v_cap", OwnerID: "a_v", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	gs.Regions["b_v_cap"] = &world.Region{ID: "b_v_cap", OwnerID: "b_v", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	NormalizeVassalage(gs)

	result := Execute(gs, "a", "b", ActionDeclareWar)

	if !result.Accepted || !result.Applied {
		t.Fatalf("savaş ilanı uygulanmalıydı: %+v", result)
	}
	for _, pair := range [][2]faction.FactionID{
		{"a", "b"},
		{"a_v", "b"},
		{"a", "b_v"},
		{"a_v", "b_v"},
	} {
		if rel := Relation(gs, pair[0], pair[1]); rel == nil || rel.Stance != faction.StanceWar {
			t.Fatalf("war coalition bekleniyordu for %s-%s, got=%+v", pair[0], pair[1], rel)
		}
	}
}

func TestBuildWarDeclarationPreviewIncludesVassalsAndCallableAllies(t *testing.T) {
	gs := testGameState()
	gs.Factions["a_v"] = &faction.Faction{ID: "a_v", NameTR: "A Vassal", Religion: religion.Catholic, OverlordID: "a"}
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Ally", Religion: religion.Catholic}
	gs.Factions["b_v"] = &faction.Faction{ID: "b_v", NameTR: "B Vassal", Religion: religion.Catholic, OverlordID: "b"}
	gs.Regions["a_v_cap"] = &world.Region{ID: "a_v_cap", OwnerID: "a_v", TradeCapacity: 4}
	gs.Regions["ally_cap"] = &world.Region{ID: "ally_cap", OwnerID: "ally", TradeCapacity: 4}
	gs.Regions["b_v_cap"] = &world.Region{ID: "b_v_cap", OwnerID: "b_v", TradeCapacity: 4}
	EnsureRelation(gs, "a", "ally").Stance = faction.StanceAllied
	EnsureRelation(gs, "a", "ally").Score = 55
	NormalizeVassalage(gs)

	preview := BuildWarDeclarationPreview(gs, "a", "b")

	if len(preview.Attacker.AutoParticipants) != 1 || preview.Attacker.AutoParticipants[0].FactionID != "a_v" {
		t.Fatalf("saldıran vassal önizlemesi eksik: %+v", preview.Attacker.AutoParticipants)
	}
	if len(preview.Attacker.CallableAllies) != 1 || preview.Attacker.CallableAllies[0].FactionID != "ally" {
		t.Fatalf("çağrılabilir müttefik önizlemesi eksik: %+v", preview.Attacker.CallableAllies)
	}
	if len(preview.Defender.AutoParticipants) != 1 || preview.Defender.AutoParticipants[0].FactionID != "b_v" {
		t.Fatalf("savunan vassal önizlemesi eksik: %+v", preview.Defender.AutoParticipants)
	}
}

func TestExecuteWarDeclarationDecliningSelectedAllyBreaksAlliance(t *testing.T) {
	gs := testGameState()
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Ally", Religion: religion.Catholic, Gold: 120, Grain: 100}
	gs.Regions["ally_cap"] = &world.Region{ID: "ally_cap", OwnerID: "ally", TradeCapacity: 4}
	rel := EnsureRelation(gs, "a", "ally")
	rel.Stance = faction.StanceAllied
	rel.Score = 20
	enemyRel := EnsureRelation(gs, "ally", "b")
	enemyRel.Stance = faction.StancePeace
	enemyRel.Score = 30
	ensureTradeRoutesBetween(gs, "ally", "b")

	result := ExecuteWarDeclaration(gs, "a", "b", []faction.FactionID{"ally"})

	if !result.Applied {
		t.Fatalf("savaş ilanı uygulanmalıydı: %+v", result)
	}
	if len(result.PlayerCalls) != 1 || result.PlayerCalls[0].Joined {
		t.Fatalf("müttefik çağrıyı reddetmeliydi: %+v", result.PlayerCalls)
	}
	updated := Relation(gs, "a", "ally")
	if updated == nil || updated.Stance != faction.StancePeace {
		t.Fatalf("ittifak bozulup barışa düşmeliydi, got=%+v", updated)
	}
	if updated.Score != 10 {
		t.Fatalf("ilişki puanı 10'a düşmeliydi, got=%d", updated.Score)
	}
}

func TestExecuteWarDeclarationQueuesPlayerWhenAlliedAttackerCallsToWar(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "player"
	gs.Factions["player"] = &faction.Faction{ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic}
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Müttefik", Religion: religion.Catholic}
	gs.Regions["player_cap"] = &world.Region{ID: "player_cap", OwnerID: "player", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	gs.Regions["ally_cap"] = &world.Region{ID: "ally_cap", OwnerID: "ally", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}

	playerRel := EnsureRelation(gs, "ally", "player")
	playerRel.Stance = faction.StanceAllied
	playerRel.Score = 55

	result := ExecuteWarDeclaration(gs, "ally", "b", nil)

	if !result.Accepted || !result.Applied {
		t.Fatalf("AI savaş ilanı uygulanmalıydı: %+v", result)
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("oyuncuya savaş çağrısı kuyruğa düşmeliydi, got=%d", len(gs.DiplomaticOffers))
	}
	offer := gs.DiplomaticOffers[0]
	if offer.Action != string(ActionJoinWarCall) || offer.FromFactionID != "ally" || offer.ToFactionID != "player" {
		t.Fatalf("beklenmeyen savaş çağrısı teklifi: %+v", offer)
	}
	if offer.WarDeclarerFactionID != "ally" || offer.WarEnemyFactionID != "b" {
		t.Fatalf("savaş çağrısı metadata hatalı: %+v", offer)
	}
	if IsWar(gs, "player", "b") {
		t.Fatal("oyuncu kabul etmeden savaşa dahil olmamalı")
	}
	if len(result.PlayerCalls) != 1 || !result.PlayerCalls[0].PendingDecision {
		t.Fatalf("saldıran taraf için bekleyen oyuncu çağrısı görünmeliydi: %+v", result.PlayerCalls)
	}
}

func TestExecuteWarDeclarationDoesNotQueuePlayerAlreadyAtWarWithTarget(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "player"
	gs.Factions["player"] = &faction.Faction{ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic}
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Müttefik", Religion: religion.Catholic}
	gs.Regions["player_cap"] = &world.Region{ID: "player_cap", OwnerID: "player", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	gs.Regions["ally_cap"] = &world.Region{ID: "ally_cap", OwnerID: "ally", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}

	playerRel := EnsureRelation(gs, "ally", "player")
	playerRel.Stance = faction.StanceAllied
	playerRel.Score = 55
	EnsureRelation(gs, "player", "b").Stance = faction.StanceWar

	result := ExecuteWarDeclaration(gs, "ally", "b", nil)

	if !result.Accepted || !result.Applied {
		t.Fatalf("AI savaş ilanı uygulanmalıydı: %+v", result)
	}
	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("oyuncu zaten savaşta olduğu hedef için savaş çağrısı kuyruğa düşmemeliydi: %+v", gs.DiplomaticOffers)
	}
	if len(result.PlayerCalls) != 0 {
		t.Fatalf("oyuncu zaten savaşta olduğu hedef için bekleyen çağrı üretilmemeliydi: %+v", result.PlayerCalls)
	}
	if !IsWar(gs, "player", "b") {
		t.Fatal("mevcut oyuncu-b savaşı korunmalıydı")
	}
}

func TestBestOfferIndexSkipsWarJoinOfferForExistingWar(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "b"
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic}
	gs.Regions["c_cap"] = &world.Region{ID: "c_cap", OwnerID: "c", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	EnsureRelation(gs, "b", "c").Stance = faction.StanceWar
	gs.DiplomaticOffers = []state.DiplomaticOffer{
		{
			FromFactionID:        "a",
			ToFactionID:          "b",
			Action:               string(ActionJoinWarCall),
			WarDeclarerFactionID: "a",
			WarEnemyFactionID:    "c",
		},
	}

	if _, ok := BestOfferIndex(gs, "b"); ok {
		t.Fatal("oyuncu hedefle zaten savaşta olduğunda savaş çağrısı modal için seçilmemeliydi")
	}
}

func TestExecuteWarDeclarationQueuesPlayerWhenAllyIsAttacked(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "player"
	gs.Factions["player"] = &faction.Faction{ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic}
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Müttefik", Religion: religion.Catholic}
	gs.Regions["player_cap"] = &world.Region{ID: "player_cap", OwnerID: "player", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	gs.Regions["ally_cap"] = &world.Region{ID: "ally_cap", OwnerID: "ally", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}

	playerRel := EnsureRelation(gs, "ally", "player")
	playerRel.Stance = faction.StanceAllied
	playerRel.Score = 55

	result := ExecuteWarDeclaration(gs, "b", "ally", nil)

	if !result.Accepted || !result.Applied {
		t.Fatalf("AI savaş ilanı uygulanmalıydı: %+v", result)
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("savunan taraf için oyuncu çağrısı kuyruğa düşmeliydi, got=%d", len(gs.DiplomaticOffers))
	}
	offer := gs.DiplomaticOffers[0]
	if offer.Action != string(ActionJoinWarCall) || offer.FromFactionID != "ally" || offer.ToFactionID != "player" {
		t.Fatalf("beklenmeyen savunma savaş çağrısı teklifi: %+v", offer)
	}
	if offer.WarDeclarerFactionID != "b" || offer.WarEnemyFactionID != "b" {
		t.Fatalf("savunma savaş çağrısı metadata hatalı: %+v", offer)
	}
	if IsWar(gs, "player", "b") {
		t.Fatal("oyuncu kabul etmeden savunma savaşına dahil olmamalı")
	}
	if len(result.EnemyCalls) != 1 || !result.EnemyCalls[0].PendingDecision {
		t.Fatalf("savunan taraf için bekleyen oyuncu çağrısı görünmeliydi: %+v", result.EnemyCalls)
	}
}

func TestResolveAcceptedWarJoinOfferAddsPlayerToWar(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "player"
	gs.Factions["player"] = &faction.Faction{ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic}
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Müttefik", Religion: religion.Catholic}
	gs.Regions["player_cap"] = &world.Region{ID: "player_cap", OwnerID: "player", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	gs.Regions["ally_cap"] = &world.Region{ID: "ally_cap", OwnerID: "ally", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}

	playerRel := EnsureRelation(gs, "ally", "player")
	playerRel.Stance = faction.StanceAllied
	playerRel.Score = 60
	EnsureRelation(gs, "ally", "b").Stance = faction.StanceWar
	QueueWarJoinOffer(gs, "ally", "player", "b", "b", "")

	result := ResolveOffer(gs, 0, true)

	if !result.Accepted || !result.Applied {
		t.Fatalf("kabul edilen savaş çağrısı uygulanmalıydı: %+v", result)
	}
	if !IsWar(gs, "player", "b") {
		t.Fatal("oyuncu kabul sonrası hedefle savaşta olmalı")
	}
	ledger := gs.WarLedgerFor("player", "b")
	if ledger == nil || ledger.DeclarerFactionID != "b" || ledger.DefenderFactionID != "player" {
		t.Fatalf("savunan ittifaka katılım savaş ilanı yönünü korumalıydı: %+v", ledger)
	}
}

func TestResolveRejectedWarJoinOfferBreaksAlliance(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "player"
	gs.Factions["player"] = &faction.Faction{ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic}
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Müttefik", Religion: religion.Catholic}
	gs.Regions["player_cap"] = &world.Region{ID: "player_cap", OwnerID: "player", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
	gs.Regions["ally_cap"] = &world.Region{ID: "ally_cap", OwnerID: "ally", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}

	playerRel := EnsureRelation(gs, "ally", "player")
	playerRel.Stance = faction.StanceAllied
	playerRel.Score = 60
	QueueWarJoinOffer(gs, "ally", "player", "b", "b", "")

	result := ResolveOffer(gs, 0, false)

	if result.Accepted || !result.Applied {
		t.Fatalf("ret edilen savaş çağrısı alliance break ile uygulanmış sayılmalıydı: %+v", result)
	}
	updated := Relation(gs, "ally", "player")
	if updated == nil || updated.Stance != faction.StancePeace {
		t.Fatalf("ret sonrası ittifak bozulup barışa düşmeliydi, got=%+v", updated)
	}
	if updated.Score != 50 {
		t.Fatalf("ret sonrası ilişki puanı 10 düşmeliydi, got=%d", updated.Score)
	}
}

func testGameState() *state.GameState {
	return &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", NameTR: "A", Religion: religion.Catholic, Grain: 120, Iron: 40, Spice: 10},
			"b": {ID: "b", NameTR: "B", Religion: religion.Catholic, Grain: 80, Cloth: 15},
		},
		Regions: map[world.RegionID]*world.Region{
			"a_cap": {ID: "a_cap", OwnerID: "a", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4},
			"b_cap": {ID: "b_cap", OwnerID: "b", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4},
		},
		Relations:   map[string]*faction.Relation{},
		TradeRoutes: []*economy.TradeRoute{},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "a", RegionID: "a_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"b1": {ID: "b1", OwnerID: "b", RegionID: "b_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 12, Defense: 10, Morale: 60},
		},
	}
}

func enableABLandTrade(gs *state.GameState) {
	gs.Regions["a_cap"].Neighbors = []world.RegionID{"b_cap"}
	gs.Regions["b_cap"].Neighbors = []world.RegionID{"a_cap"}
}

func TestAssessTradeProposalUsesEffectiveBuildingCapacity(t *testing.T) {
	gs := testGameState()
	gs.Regions["a_cap"].TradeCapacity = 3
	gs.Regions["b_cap"].TradeCapacity = 3
	gs.Regions["a_cap"].Buildings = []string{"market"}
	gs.Regions["b_cap"].Buildings = []string{"market"}
	gs.BuildingTypes = map[string]*city.Building{
		"market": {ID: "market", TradeCapacityMod: 1.45},
	}
	enableABLandTrade(gs)
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 30

	assessment := AssessTradeProposal(gs, rel, "a", "b")
	if assessment.BlockReason != "" {
		t.Fatalf("pazarın yükselttiği efektif kapasite ticareti açmalıydı: %+v", assessment)
	}
}
