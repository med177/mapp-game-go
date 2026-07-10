package diplomacy

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

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
	rel.Score = 20

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
	rel.Score = 20

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
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceTrade
	rel.Score = 25
	gs.TradeRoutes = nil

	EnsureTradeRoutesForActiveRelations(gs)

	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("trade stance için iki yönlü 2 rota kurulmalıydı, got=%d", len(gs.TradeRoutes))
	}
}

func TestEnsureTradeRoutesForActiveRelationsRemovesStaleEliminatedRoutes(t *testing.T) {
	gs := testGameState()
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

func TestAssessTradeProposalBlocksLowScore(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 5

	assessment := AssessTradeProposal(gs, rel, "a", "b")
	if assessment.BlockReason != "İlişki puanı 10 altı" {
		t.Fatalf("beklenen düşük skor engeli, got=%+v", assessment)
	}
	if assessment.Accepted() {
		t.Fatalf("düşük skor kabul edilmemeli: %+v", assessment)
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

func TestOfferVassalizationMakesTargetVassalAndBlocksThirdParties(t *testing.T) {
	gs := testGameState()
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic, Gold: 90, Grain: 60}
	gs.Regions["c_cap"] = &world.Region{ID: "c_cap", OwnerID: "c", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4}
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
	gs.Factions["b"].OverlordID = "a"
	gs.Factions["c"] = &faction.Faction{ID: "c", NameTR: "C", Religion: religion.Catholic}

	if reason := ActionBlockReason(gs, "a", "b", ActionAnnexVassal); reason != "" {
		t.Fatalf("doğrudan overlord ilhak edebilmeli, got=%q", reason)
	}
	if reason := ActionBlockReason(gs, "c", "b", ActionAnnexVassal); reason == "" {
		t.Fatal("üçüncü taraf vassalı ilhak edememeli")
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
