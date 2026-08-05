package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func siegeTestState() *state.GameState {
	return &state.GameState{
		Turn:            5,
		PlayerFactionID: "p1",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", Religion: "sunni"},
			"p2": {ID: "p2", Religion: "catholic"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {ID: "src", OwnerID: "p1", Neighbors: []world.RegionID{"dst"}},
			"dst": {
				ID:          "dst",
				OwnerID:     "p2",
				Neighbors:   []world.RegionID{"src"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress, NameTR: "Kale"}},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":   {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
			"siege": {ID: "siege", Category: army.CategorySiege, Tier: 1, Attack: 8, Defense: 4, Morale: 30},
		},
	}
}

func siegeSupportTestState() *state.GameState {
	return &state.GameState{
		Turn:            5,
		PlayerFactionID: "p1",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1":   {ID: "p1", Religion: "sunni"},
			"ally": {ID: "ally", Religion: "sunni"},
			"p3":   {ID: "p3", Religion: "catholic"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "p1",
				Neighbors: []world.RegionID{"dst"},
			},
			"ally_src": {
				ID:        "ally_src",
				OwnerID:   "ally",
				Neighbors: []world.RegionID{"dst"},
			},
			"dst": {
				ID:          "dst",
				OwnerID:     "p3",
				Neighbors:   []world.RegionID{"src", "ally_src"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress, NameTR: "Kale"}},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p3"):   {FactionA: "p1", FactionB: "p3", Stance: faction.StanceWar},
			faction.RelationKey("ally", "p3"): {FactionA: "ally", FactionB: "p3", Stance: faction.StanceWar},
			faction.RelationKey("p1", "ally"): {FactionA: "p1", FactionB: "ally", Stance: faction.StanceAllied},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":   {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
			"siege": {ID: "siege", Category: army.CategorySiege, Tier: 1, Attack: 8, Defense: 4, Morale: 30},
		},
	}
}

func sortieTestState(defenderUnits, besiegerUnits int) *state.GameState {
	gs := &state.GameState{
		Turn:            5,
		Month:           6,
		PlayerFactionID: "p2",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1":   {ID: "p1", Religion: "sunni"},
			"p2":   {ID: "p2", Religion: "catholic"},
			"ally": {ID: "ally", Religion: "catholic"},
		},
		Regions: map[world.RegionID]*world.Region{
			"besieged": {ID: "besieged", OwnerID: "p2", Neighbors: []world.RegionID{"exit"}},
			"exit":     {ID: "exit", OwnerID: "p2", Neighbors: []world.RegionID{"besieged"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p2"):   {FactionA: "p1", FactionB: "p2", Stance: faction.StanceWar},
			faction.RelationKey("p1", "ally"): {FactionA: "p1", FactionB: "ally", Stance: faction.StanceWar},
			faction.RelationKey("p2", "ally"): {FactionA: "p2", FactionB: "ally", Stance: faction.StanceAllied},
		},
		Armies: map[army.ArmyID]*army.Army{
			"besieger": {
				ID: "besieger", OwnerID: "p1", RegionID: "besieged", MovePoints: 2, MaxMovePoints: 2,
				Units: repeatedUnits("inf", besiegerUnits, 100),
			},
			"defender": {
				ID: "defender", OwnerID: "p2", RegionID: "besieged", MovePoints: 2, MaxMovePoints: 2,
				Units: repeatedUnits("inf", defenderUnits, 100),
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"besieged": {RegionID: "besieged", AttackerArmyID: "besieger", AttackerFactionID: "p1", FortLevel: 2},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	return gs
}

func TestStartSiegeCreatesStateWithoutSiegeUnit(t *testing.T) {
	gs := siegeTestState()
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if !g.startSiegeForArmy("atk", "dst", false) {
		t.Fatal("kuşatma başlatılamadı")
	}

	if gs.Armies["atk"].RegionID != "dst" {
		t.Fatalf("kuşatma başlatan ordu hedefe yerleşmeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.Regions["dst"].OwnerID != "p2" {
		t.Fatalf("kuşatma başlatmak sahipliği değiştirmemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
	if gs.SiegeAt("dst") == nil {
		t.Fatal("kuşatma kaydı oluşmalıydı")
	}
}

func TestMoveArmyToFortifiedRegionOpensSiegeDecisionWithoutSiegeUnit(t *testing.T) {
	gs := siegeTestState()
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
		},
	}
	r := render.New(gs)
	g := &Game{gs: gs, renderer: r}

	g.moveArmyWithStance("atk", "dst", combat.BattleStanceBalanced)

	if !r.ConfirmDialogVisible() {
		t.Fatal("tahkimli hedefe doğrudan hareket emri kuşatma kararını açmalıydı")
	}
	if gs.Armies["atk"].RegionID != "src" || gs.Armies["atk"].MovePoints != 2 {
		t.Fatalf("karar verilmeden ordu hareket etmemeli: %+v", gs.Armies["atk"])
	}
	if gs.SiegeAt("dst") != nil {
		t.Fatal("karar verilmeden kuşatma başlatılmamalı")
	}
}

func TestStartSiegeCreatesStateWithoutSiegeUnitAgainstDefenderArmy(t *testing.T) {
	gs := siegeTestState()
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
		},
		"def": {
			ID:            "def",
			OwnerID:       "p2",
			RegionID:      "dst",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if !g.startSiegeForArmy("atk", "dst", false) {
		t.Fatal("kuşatma savunma ordusu varken de başlatılabilmeliydi")
	}
	if gs.SiegeAt("dst") == nil {
		t.Fatal("savunma ordusu varken kuşatma kaydı oluşmalıydı")
	}
	if gs.SiegeAt("dst").DefenderArmyID != "def" {
		t.Fatalf("kuşatma savunma ordusunu bağlamalıydı, got=%+v", gs.SiegeAt("dst"))
	}
	if gs.Armies["atk"].RegionID != "dst" {
		t.Fatalf("kuşatma başlatan ordu hedefe yerleşmeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
}

func TestStartSiegeCreatesStateAndConsumesMove(t *testing.T) {
	gs := siegeTestState()
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "siege", CurrentHP: 100}},
		},
		"def": {
			ID:            "def",
			OwnerID:       "p2",
			RegionID:      "dst",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
		},
	}
	g := &Game{gs: gs}

	if !g.startSiegeForArmy("atk", "dst", false) {
		t.Fatal("kuşatma başlatılamadı")
	}
	siege := gs.SiegeAt("dst")
	if siege == nil {
		t.Fatal("kuşatma kaydı oluşmalıydı")
	}
	if siege.AttackerArmyID != "atk" || siege.DefenderArmyID != "def" {
		t.Fatalf("kuşatma eşleşmesi hatalı: %+v", siege)
	}
	if gs.Armies["atk"].MovePoints != 0 {
		t.Fatalf("kuşatma başlatan ordu hareketini bitirmeliydi, got=%d", gs.Armies["atk"].MovePoints)
	}
}

func TestCommanderSiegeBonusesIncreaseProgressAndBreachGain(t *testing.T) {
	gs := siegeTestState()
	baseAttacker := &army.Army{
		ID:            "atk_base",
		OwnerID:       "p1",
		RegionID:      "src",
		MovePoints:    2,
		MaxMovePoints: 2,
		Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "siege", CurrentHP: 100}},
	}
	commandedAttacker := &army.Army{
		ID:            "atk_cmd",
		OwnerID:       "p1",
		RegionID:      "src",
		MovePoints:    2,
		MaxMovePoints: 2,
		Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "siege", CurrentHP: 100}},
	}
	aggressor := army.NewCommander("cmd_aggressor", "Saldırgan")
	aggressor.Experience = army.CommanderLevel5XP
	aggressor.Normalize()
	commandedAttacker.AssignCommander(aggressor)

	baseProgress := siegeProgressGain(gs, baseAttacker, gs.Regions["dst"], nil)
	baseBreach := siegeBreachGain(gs, baseAttacker, gs.Regions["dst"], nil)
	commandedProgress := siegeProgressGain(gs, commandedAttacker, gs.Regions["dst"], nil)
	commandedBreach := siegeBreachGain(gs, commandedAttacker, gs.Regions["dst"], nil)

	if commandedProgress != baseProgress+1 {
		t.Fatalf("saldırgan komutan kuşatma ilerlemesine +1 vermeliydi: base=%d commanded=%d", baseProgress, commandedProgress)
	}
	if commandedBreach != baseBreach+1 {
		t.Fatalf("saldırgan komutan gedik kazanımına +1 vermeliydi: base=%d commanded=%d", baseBreach, commandedBreach)
	}
}

func TestMoveArmyWhileBesiegingClearsSiegeAndMoves(t *testing.T) {
	gs := siegeTestState()
	gs.Regions["src"].Neighbors = []world.RegionID{"dst", "ally"}
	gs.Regions["ally"] = &world.Region{
		ID:        "ally",
		OwnerID:   "p1",
		Neighbors: []world.RegionID{"src"},
	}
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "siege", CurrentHP: 100}},
		},
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			FortLevel:         2,
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("atk", "ally", "")

	if gs.Armies["atk"].RegionID != "ally" {
		t.Fatalf("kuşatmayı kaldırıp komşu dost bölgeye yürümeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.SiegeAt("dst") != nil {
		t.Fatal("ordu başka bölgeye yürüyünce eski kuşatma temizlenmeliydi")
	}
}

func TestMoveArmyFromBesiegedRegionRequiresSortieAndMovesAfterVictory(t *testing.T) {
	gs := sortieTestState(4, 1)
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("defender", "exit", "")

	defender := gs.Armies["defender"]
	if defender == nil || defender.RegionID != "exit" {
		t.Fatalf("huruç kazanınca ordu çıkış bölgesine ilerlemeliydi, got=%+v", defender)
	}
	if defender.MovePoints != 1 {
		t.Fatalf("huruç sonrası hareket puanı bir azalmalıydı, got=%d", defender.MovePoints)
	}
	if gs.SiegeAt("besieged") != nil {
		t.Fatal("huruç kazanınca aktif kuşatma temizlenmeliydi")
	}
	if _, ok := gs.Armies["besieger"]; ok {
		t.Fatal("yenilen kuşatan ordu haritadan kaldırılmalıydı")
	}
}

func TestSortieSiegeActionResolvesInPlace(t *testing.T) {
	gs := sortieTestState(4, 1)
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if !g.sortieSiegeWithStance("defender", "besieged", "") {
		t.Fatal("savunma panelindeki huruç emri çözülmeliydi")
	}
	defender := gs.Armies["defender"]
	if defender == nil || defender.RegionID != "besieged" {
		t.Fatalf("yerinde huruç zaferinde savunmacı kuşatılan bölgede kalmalıydı: %+v", defender)
	}
	if defender.MovePoints < 1 {
		t.Fatalf("başarılı huruç sonrası savunmacının en az 1 hareket hakkı kalmalıydı, got=%d", defender.MovePoints)
	}
	if gs.SiegeAt("besieged") != nil {
		t.Fatal("başarılı huruç kuşatmayı kaldırmalıydı")
	}
}

func TestSurrenderSiegeCapturesRegionAndWithdrawsDefenders(t *testing.T) {
	gs := sortieTestState(1, 1)
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if !g.surrenderSiege("defender", "besieged") {
		t.Fatal("teslimiyet emri çözülmeliydi")
	}
	if gs.SiegeAt("besieged") != nil {
		t.Fatal("teslimiyet kuşatma kaydını temizlemeliydi")
	}
	if gs.Regions["besieged"].OwnerID != "p1" {
		t.Fatalf("teslimiyet sonrası bölge kuşatana geçmeliydi, got=%s", gs.Regions["besieged"].OwnerID)
	}
	defender := gs.Armies["defender"]
	if defender == nil || defender.RegionID != "exit" {
		t.Fatalf("teslimiyet savunma ordusunu en yakın dost bölgeye çekmeliydi: %+v", defender)
	}
	if defender.Morale != 85 {
		t.Fatalf("teslimiyet geri çekilen ordunun moralini düşürmeliydi, got=%d", defender.Morale)
	}
	if gs.Armies["besieger"].RegionID != "besieged" {
		t.Fatalf("kuşatan ordu teslimiyet sonrası bölgede kalmalıydı: %+v", gs.Armies["besieger"])
	}
}

func TestAcceptedLastRegionSiegeVassalizationKeepsRegionAndEndsSiege(t *testing.T) {
	gs := sortieTestState(1, 1)
	gs.PlayerFactionID = "p1"
	delete(gs.Regions, "exit")
	gs.DiplomaticOffers = []state.DiplomaticOffer{
		{FromFactionID: "p2", ToFactionID: "p1", Action: string(diplomacy.ActionProposeSiegeVassalization), RegionID: "besieged", CreatedTurn: gs.Turn, Priority: 175},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	_, result, ok := g.resolveDiplomacyOffer(0, true)
	if !ok || !result.Accepted || !result.Applied {
		t.Fatalf("son toprak vassallığı kabul edilmeliydi: ok=%t result=%+v", ok, result)
	}
	if gs.Factions["p2"].OverlordID != "p1" {
		t.Fatalf("son toprağı savunan devlet oyuncunun vassalı olmalıydı: %+v", gs.Factions["p2"])
	}
	if gs.Regions["besieged"].OwnerID != "p2" {
		t.Fatalf("vassallıkta bölge sahibi değişmemeliydi, got=%s", gs.Regions["besieged"].OwnerID)
	}
	if gs.SiegeAt("besieged") != nil || len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("vassallık kuşatmayı bitirip teklifi tüketmeliydi: siege=%+v offers=%+v", gs.SiegeAt("besieged"), gs.DiplomaticOffers)
	}
}

func TestPlayerCanOfferSiegeVassalizationToLastRegion(t *testing.T) {
	gs := sortieTestState(1, 2)
	gs.PlayerFactionID = "p1"
	gs.DiplomacyOfferCounts = map[faction.FactionID]int{"p1": state.MaxDiplomacyOffersPerTurn}
	delete(gs.Regions, "exit")
	siege := gs.SiegeAt("besieged")
	siege.TurnsElapsed = 3
	siege.BreachLevel = 2
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.proposeSiegeSurrender("besieger", "besieged")
	if got := gs.DiplomacyOfferQuotaUsed("p1"); got != state.MaxDiplomacyOffersPerTurn {
		t.Fatalf("oyuncunun elçi hakkı kuşatma vassallığında değişmemeli, got=%d", got)
	}
	if len(g.pendingConquestDecisions) != 0 {
		t.Fatalf("son toprak vassallığı savaş sonrası düzen kararı açmamalıydı, got=%d", len(g.pendingConquestDecisions))
	}
	if gs.SiegeAt("besieged") != nil || len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("son toprak vassallığı kabul edilince kuşatma bitip teklif tüketilmeliydi: siege=%+v offers=%+v", gs.SiegeAt("besieged"), gs.DiplomaticOffers)
	}
	if gs.Regions["besieged"].OwnerID != "p2" {
		t.Fatalf("vassallıkta bölge sahibi değişmemeliydi, got=%s", gs.Regions["besieged"].OwnerID)
	}
	if gs.Factions["p2"].OverlordID != "p1" {
		t.Fatalf("AI son toprağıyla oyuncunun vassalı olmalıydı: %+v", gs.Factions["p2"])
	}
}

func TestMoveArmyFromBesiegedRegionStaysWithLossesAndDoesNotReplenish(t *testing.T) {
	gs := sortieTestState(1, 4)
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("defender", "exit", "")

	defender := gs.Armies["defender"]
	if defender == nil {
		t.Fatal("huruç yenilgisi kalan savunmacı orduyu silmemeliydi")
	}
	if defender.RegionID != "besieged" {
		t.Fatalf("huruç kaybeden ordu kuşatılan bölgede kalmalıydı, got=%s", defender.RegionID)
	}
	if defender.MovePoints != 0 {
		t.Fatalf("huruç yenilgisi hareket puanını bitirmeliydi, got=%d", defender.MovePoints)
	}
	if defender.Units[0].CurrentHP >= army.MaxUnitHP {
		t.Fatal("huruç kaybında savunmacı HP kaybetmeliydi")
	}
	if gs.SiegeAt("besieged") == nil {
		t.Fatal("huruç kaybında kuşatma devam etmeliydi")
	}

	beforeHP := defender.Units[0].CurrentHP
	applySeasonEffects(gs)
	if defender.Units[0].CurrentHP != beforeHP {
		t.Fatalf("kuşatma altındaki savunmacı iyileşmemeliydi, before=%d after=%d", beforeHP, defender.Units[0].CurrentHP)
	}
}

func TestArmyOfBesiegedRegionAllyIsAlsoSortieDefender(t *testing.T) {
	gs := sortieTestState(4, 1)
	gs.PlayerFactionID = "ally"
	gs.Armies["defender"].OwnerID = "ally"
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if !gs.IsArmyDefendingSiegedRegion(gs.Armies["defender"]) {
		t.Fatal("bölge sahibinin müttefik ordusu savunmacı kabul edilmeliydi")
	}
	g.moveArmyWithStance("defender", "exit", "")
	if gs.Armies["defender"] == nil || gs.Armies["defender"].RegionID != "exit" {
		t.Fatalf("müttefik savunmacı huruç zaferi sonrası çıkabilmeliydi, got=%+v", gs.Armies["defender"])
	}
}

func TestSplitBesiegingArmyKeepsSiegeWithRemainingUnit(t *testing.T) {
	gs := siegeTestState()
	gs.Regions["dst"].Neighbors = []world.RegionID{"src", "ally"}
	gs.Regions["ally"] = &world.Region{
		ID:        "ally",
		OwnerID:   "p1",
		Neighbors: []world.RegionID{"dst"},
	}
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "dst",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units: []army.Unit{
				{TypeID: "inf", CurrentHP: 100},
				{TypeID: "inf", CurrentHP: 100},
			},
		},
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			FortLevel:         2,
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.splitArmy("atk")
	newArmyID := army.ArmyID("army_p1_1")
	if g.renderer.SelectedArmy != newArmyID {
		t.Fatalf("split sonrası yeni parça seçili kalmalıydı, got=%s", g.renderer.SelectedArmy)
	}
	if len(gs.Armies["atk"].Units) != 1 || len(gs.Armies[newArmyID].Units) != 1 {
		t.Fatalf("split iki tarafta da birer birim bırakmalıydı: eski=%d yeni=%d", len(gs.Armies["atk"].Units), len(gs.Armies[newArmyID].Units))
	}

	// Hareket emri kuşatma ID'sine gitse bile bölgede kalan tek birimlik kardeş
	// ordu kuşatmayı sürdürmelidir.
	g.moveArmyWithStance("atk", "ally", "")

	if gs.Armies["atk"].RegionID != "ally" {
		t.Fatalf("hareket eden parça dost bölgeye çıkmalıydı, got=%s", gs.Armies["atk"].RegionID)
	}
	siege := gs.SiegeAt("dst")
	if siege == nil || siege.AttackerArmyID != newArmyID {
		t.Fatalf("kalan birim kuşatmayı devralmalıydı, got=%+v", siege)
	}
}

func TestSplitArmyWithSelectedUnitsMovesOnlyThoseUnits(t *testing.T) {
	gs := siegeTestState()
	gs.NextArmySeq = 0
	gs.Armies = map[army.ArmyID]*army.Army{
		"main": {
			ID:       "main",
			OwnerID:  "p1",
			RegionID: "src",
			Units: []army.Unit{
				{TypeID: "inf", CurrentHP: 91},
				{TypeID: "cav", CurrentHP: 82},
				{TypeID: "siege", CurrentHP: 73},
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.splitArmy("main", 0, 2)

	newArmy := gs.Armies["army_p1_1"]
	if newArmy == nil {
		t.Fatal("seçilen birliklerden yeni ordu oluşturulmalıydı")
	}
	if len(gs.Armies["main"].Units) != 1 || gs.Armies["main"].Units[0].TypeID != "cav" {
		t.Fatalf("seçilmeyen birlik ana orduda kalmalıydı: %+v", gs.Armies["main"].Units)
	}
	if len(newArmy.Units) != 2 || newArmy.Units[0].TypeID != "inf" || newArmy.Units[1].TypeID != "siege" {
		t.Fatalf("seçilen birlikler yeni orduya aynı sırayla taşınmalıydı: %+v", newArmy.Units)
	}
}

func TestMergeBesiegingArmyKeepsSiegeWithSurvivingArmy(t *testing.T) {
	newGame := func() (*Game, army.ArmyID) {
		gs := siegeTestState()
		gs.Armies = map[army.ArmyID]*army.Army{
			"atk": {
				ID:            "atk",
				OwnerID:       "p1",
				RegionID:      "dst",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units: []army.Unit{
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
				},
			},
			"split": {
				ID:            "split",
				OwnerID:       "p1",
				RegionID:      "dst",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		}
		gs.Sieges = map[world.RegionID]*state.SiegeState{
			"dst": {
				RegionID:          "dst",
				AttackerArmyID:    "atk",
				AttackerFactionID: "p1",
				FortLevel:         2,
			},
		}
		return &Game{gs: gs, renderer: &render.Renderer{}}, "split"
	}

	t.Run("kuşatan ordu silinirse", func(t *testing.T) {
		g, survivingID := newGame()
		g.mergeArmiesManual("atk")

		if g.gs.Armies["atk"] != nil {
			t.Fatal("kuşatan ordu birleşme sonrası silinmeliydi")
		}
		if siege := g.gs.SiegeAt("dst"); siege == nil || siege.AttackerArmyID != survivingID {
			t.Fatalf("kuşatma hayatta kalan orduya devredilmeliydi, got=%+v", g.gs.SiegeAt("dst"))
		}
	})

	t.Run("kuşatan ordu hayatta kalırsa", func(t *testing.T) {
		g, _ := newGame()
		g.mergeArmiesManual("split")

		if g.gs.Armies["split"] != nil {
			t.Fatal("birleşen parça silinmeliydi")
		}
		if siege := g.gs.SiegeAt("dst"); siege == nil || siege.AttackerArmyID != "atk" {
			t.Fatalf("kuşatma mevcut kuşatan orduda kalmalıydı, got=%+v", g.gs.SiegeAt("dst"))
		}
	})
}

func TestMoveArmyWithStanceAllowsAlliedSiegeSupport(t *testing.T) {
	gs := siegeSupportTestState()
	gs.PlayerFactionID = "ally"
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "siege", CurrentHP: 100}},
		},
		"support": {
			ID:            "support",
			OwnerID:       "ally",
			RegionID:      "ally_src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
		},
		"def": {
			ID:            "def",
			OwnerID:       "p3",
			RegionID:      "dst",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
		},
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			StartedTurn:       5,
			FortLevel:         2,
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("support", "dst", "")

	if gs.Armies["support"].RegionID != "dst" {
		t.Fatalf("müttefik destek ordusu kuşatmaya katılabilmeliydi, got=%s", gs.Armies["support"].RegionID)
	}
	if gs.Armies["support"].MovePoints != 1 {
		t.Fatalf("müttefik destek ordusu bir hareket puanı harcamalıydı, got=%d", gs.Armies["support"].MovePoints)
	}
	if gs.Sieges["dst"] == nil || gs.Sieges["dst"].AttackerArmyID != "atk" {
		t.Fatalf("mevcut kuşatma korunmalıydı, got=%+v", gs.Sieges["dst"])
	}
	if gs.Regions["dst"].OwnerID != "p3" {
		t.Fatalf("kuşatma desteği bölgeyi fethetmemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
	if gs.Armies["def"] == nil || gs.Armies["def"].RegionID != "dst" {
		t.Fatalf("savunan ordu yerinde kalmalıydı, got=%+v", gs.Armies["def"])
	}
}

func TestResolveSiegesUsesSiegeUnitArrivingAfterSiegeStarted(t *testing.T) {
	gs := siegeSupportTestState()
	gs.PlayerFactionID = "ally"
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID: "atk", OwnerID: "p1", RegionID: "src", MovePoints: 2, MaxMovePoints: 2,
			Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
		},
		"support": {
			ID: "support", OwnerID: "ally", RegionID: "ally_src", MovePoints: 2, MaxMovePoints: 2,
			Units: []army.Unit{{TypeID: "siege", CurrentHP: 100}},
		},
		"def": {
			ID: "def", OwnerID: "p3", RegionID: "dst",
			Units: make([]army.Unit, 20),
		},
	}
	for i := range gs.Armies["def"].Units {
		gs.Armies["def"].Units[i] = army.Unit{TypeID: "inf", CurrentHP: 100}
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if !g.startSiegeForArmy("atk", "dst", false) {
		t.Fatal("kuşatma birimi olmayan ana ordu kuşatma başlatabilmeliydi")
	}
	g.resolveSieges()
	if siege := gs.SiegeAt("dst"); siege == nil || siege.BreachProgress != 0 {
		t.Fatalf("kuşatma birimi gelmeden gedik ilerlememeliydi, got=%+v", siege)
	}

	// Müttefik kuşatma ordusu aktif kuşatma bölgesine ayrı ordu olarak gelir.
	g.moveArmyWithStance("support", "dst", "")
	if gs.Armies["support"].RegionID != "dst" {
		t.Fatalf("müttefik kuşatma ordusu aktif kuşatmaya girebilmeliydi, got=%s", gs.Armies["support"].RegionID)
	}
	if gs.Armies["atk"].HasSiegeUnits(gs.UnitTypes) {
		t.Fatal("test kurulumu ana orduda kuşatma birimi olmadan başlamalıydı")
	}

	minorThreshold, _ := siegeBreachThresholds(gs.Regions["dst"].FortificationLevel())
	for i := 0; i < minorThreshold; i++ {
		g.resolveSieges()
	}

	siege := gs.SiegeAt("dst")
	if siege == nil {
		t.Fatal("gedik açılırken kuşatma kaydı korunmalıydı")
	}
	if siege.BreachLevel < 1 {
		t.Fatalf("sonradan gelen kuşatma birimi gedik ilerlemesine katılmalıydı, progress=%d level=%d", siege.BreachProgress, siege.BreachLevel)
	}
}

func TestResolveSiegesGranaryReducesDefenderAttrition(t *testing.T) {
	buildState := func(withGranary bool) *state.GameState {
		gs := siegeTestState()
		gs.Armies = map[army.ArmyID]*army.Army{
			"atk": {
				ID: "atk", OwnerID: "p1", RegionID: "src",
				Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
			"def": {
				ID: "def", OwnerID: "p2", RegionID: "dst",
				Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		}
		gs.Sieges = map[world.RegionID]*state.SiegeState{
			"dst": {
				RegionID: "dst", AttackerArmyID: "atk", DefenderArmyID: "def",
				AttackerFactionID: "p1", FortLevel: 2,
			},
		}
		if withGranary {
			gs.Regions["dst"].Buildings = append(gs.Regions["dst"].Buildings, "granary")
		}
		return gs
	}

	withoutGranary := buildState(false)
	withGranary := buildState(true)
	(&Game{gs: withoutGranary}).resolveSieges()
	(&Game{gs: withGranary}).resolveSieges()

	withoutHP := withoutGranary.Armies["def"].Units[0].CurrentHP
	withHP := withGranary.Armies["def"].Units[0].CurrentHP
	if withHP <= withoutHP {
		t.Fatalf("ambar savunucu kuşatma yıpranmasını azaltmalıydı: ambarsız=%d ambarlı=%d", withoutHP, withHP)
	}
}

func TestMoveArmyWithStanceCanBreakSiegeInAlliedRegionWithoutConquest(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1":   {ID: "p1", Religion: "sunni"},
			"ally": {ID: "ally", Religion: "sunni"},
			"p3":   {ID: "p3", Religion: "catholic"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "p1",
				Neighbors: []world.RegionID{"dst"},
			},
			"dst": {
				ID:        "dst",
				OwnerID:   "ally",
				Neighbors: []world.RegionID{"src"},
				Terrain:   world.TerrainPlain,
				Buildings: []string{"walls"},
				Settlements: []world.Settlement{
					{ID: "fort", Type: world.SettlementFortress, NameTR: "Kale"},
				},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "ally"): {FactionA: "p1", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("p1", "p3"):   {FactionA: "p1", FactionB: "p3", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"atk": {
				ID:            "atk",
				OwnerID:       "p1",
				RegionID:      "src",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units: []army.Unit{
					{TypeID: "elite", CurrentHP: 100},
					{TypeID: "elite", CurrentHP: 100},
					{TypeID: "elite", CurrentHP: 100},
					{TypeID: "elite", CurrentHP: 100},
				},
			},
			"besieger": {
				ID:            "besieger",
				OwnerID:       "p3",
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
				AttackerFactionID: "p3",
				StartedTurn:       5,
				FortLevel:         2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"elite": {ID: "elite", Category: army.CategoryInfantry, Attack: 40, Defense: 35, Morale: 90},
			"inf":   {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("atk", "dst", "")

	if gs.Armies["atk"].RegionID != "dst" {
		t.Fatalf("AI savaşı kazanıp bölgeye girmeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.Regions["dst"].OwnerID != "ally" {
		t.Fatalf("kuşatma kaldırılırken sahiplik değişmemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
	if gs.SiegeAt("dst") != nil {
		t.Fatal("düşman kuşatma yapan ordu yenildiğinde kuşatma temizlenmeliydi")
	}
}

func TestMoveArmyWithStanceBlocksAlliedTransitIntoBesiegedRegionWithoutWarWithBesieger(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p4",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", Religion: "sunni"},
			"p3": {ID: "p3", Religion: "catholic"},
			"p4": {ID: "p4", Religion: "orthodox"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "p4",
				Neighbors: []world.RegionID{"dst"},
			},
			"dst": {
				ID:          "dst",
				OwnerID:     "p3",
				Neighbors:   []world.RegionID{"src"},
				Terrain:     world.TerrainPlain,
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress, NameTR: "Kale"}},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p4", "p3"): {FactionA: "p4", FactionB: "p3", Stance: faction.StanceAllied},
			faction.RelationKey("p4", "p1"): {FactionA: "p4", FactionB: "p1", Stance: faction.StancePeace},
			faction.RelationKey("p1", "p3"): {FactionA: "p1", FactionB: "p3", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"atk": {
				ID:            "atk",
				OwnerID:       "p4",
				RegionID:      "src",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
			"besieger": {
				ID:            "besieger",
				OwnerID:       "p1",
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
				AttackerFactionID: "p1",
				StartedTurn:       5,
				FortLevel:         2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("atk", "dst", "")

	if gs.Armies["atk"].RegionID != "src" {
		t.Fatalf("kuşatılmış dost bölgeye savaşta olmayan müttefik girememeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.Armies["atk"].MovePoints != 2 {
		t.Fatalf("engellenen giriş hareket puanı tüketmemeliydi, got=%d", gs.Armies["atk"].MovePoints)
	}
	if gs.SiegeAt("dst") == nil || gs.SiegeAt("dst").AttackerArmyID != "besieger" {
		t.Fatalf("kuşatma korunmalıydı, got=%+v", gs.SiegeAt("dst"))
	}
	if gs.Regions["dst"].OwnerID != "p3" {
		t.Fatalf("bölge sahipliği değişmemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
}

func TestMoveArmyWithStanceAllowsOverlordTransitToVassalRegion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "lord",
		Factions: map[faction.FactionID]*faction.Faction{
			"lord":   {ID: "lord", Religion: "sunni"},
			"vassal": {ID: "vassal", Religion: "sunni", OverlordID: "lord"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {ID: "src", OwnerID: "lord", Neighbors: []world.RegionID{"dst"}},
			"dst": {ID: "dst", OwnerID: "vassal", Neighbors: []world.RegionID{"src"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"atk": {
				ID:            "atk",
				OwnerID:       "lord",
				RegionID:      "src",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("atk", "dst", "")

	if gs.Armies["atk"].RegionID != "dst" {
		t.Fatalf("overlord ordusu vassal toprağına savaşsız girebilmeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.Regions["dst"].OwnerID != "vassal" {
		t.Fatalf("askeri geçişte bölge sahibi değişmemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
}

func TestMoveArmyWithStanceAllowsAlliedTransitIntoFortifiedRegionWithoutSiege(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1":   {ID: "p1", Religion: "sunni"},
			"ally": {ID: "ally", Religion: "sunni"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "p1",
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
			faction.RelationKey("p1", "ally"): {FactionA: "p1", FactionB: "ally", Stance: faction.StanceAllied},
		},
		Armies: map[army.ArmyID]*army.Army{
			"atk": {
				ID:            "atk",
				OwnerID:       "p1",
				RegionID:      "src",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("atk", "dst", "")

	if gs.Armies["atk"].RegionID != "dst" {
		t.Fatalf("müttefik tahkimli bölgeye normal hareket etmeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.Armies["atk"].MovePoints != 1 {
		t.Fatalf("başarılı hareket bir puan harcamalıydı, got=%d", gs.Armies["atk"].MovePoints)
	}
	if gs.Regions["dst"].OwnerID != "ally" {
		t.Fatalf("normal geçişte bölge sahibi değişmemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
}

func TestMoveArmyWithStanceAllowsSameRealmSiegeSupport(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "vassal",
		Factions: map[faction.FactionID]*faction.Faction{
			"lord":   {ID: "lord", Religion: "sunni"},
			"vassal": {ID: "vassal", Religion: "sunni", OverlordID: "lord"},
			"enemy":  {ID: "enemy", Religion: "catholic"},
		},
		Regions: map[world.RegionID]*world.Region{
			"lord_src":   {ID: "lord_src", OwnerID: "lord", Neighbors: []world.RegionID{"enemy_dst"}},
			"vassal_src": {ID: "vassal_src", OwnerID: "vassal", Neighbors: []world.RegionID{"enemy_dst"}},
			"enemy_dst": {
				ID:          "enemy_dst",
				OwnerID:     "enemy",
				Neighbors:   []world.RegionID{"lord_src", "vassal_src"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress, NameTR: "Kale"}},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("lord", "enemy"):   {FactionA: "lord", FactionB: "enemy", Stance: faction.StanceWar},
			faction.RelationKey("vassal", "enemy"): {FactionA: "vassal", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"atk": {
				ID:            "atk",
				OwnerID:       "lord",
				RegionID:      "lord_src",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "siege", CurrentHP: 100}},
			},
			"support": {
				ID:            "support",
				OwnerID:       "vassal",
				RegionID:      "vassal_src",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
			"def": {
				ID:            "def",
				OwnerID:       "enemy",
				RegionID:      "enemy_dst",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"enemy_dst": {
				RegionID:          "enemy_dst",
				AttackerArmyID:    "atk",
				AttackerFactionID: "lord",
				StartedTurn:       5,
				FortLevel:         2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":   {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
			"siege": {ID: "siege", Category: army.CategorySiege, Tier: 1, Attack: 8, Defense: 4, Morale: 30},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("support", "enemy_dst", "")

	if gs.Armies["support"].RegionID != "enemy_dst" {
		t.Fatalf("aynı vassal zincirindeki destek ordusu kuşatmaya katılabilmeliydi, got=%s", gs.Armies["support"].RegionID)
	}
	if gs.Regions["enemy_dst"].OwnerID != "enemy" {
		t.Fatalf("kuşatma desteği sahipliği değiştirmemeliydi, got=%s", gs.Regions["enemy_dst"].OwnerID)
	}
}

func TestMoveArmyWithStanceBlocksNonAlliedSiegeSupport(t *testing.T) {
	gs := siegeSupportTestState()
	gs.PlayerFactionID = "p4"
	gs.Factions["p4"] = &faction.Faction{ID: "p4", Religion: "orthodox"}
	gs.Regions["third_src"] = &world.Region{
		ID:        "third_src",
		OwnerID:   "p4",
		Neighbors: []world.RegionID{"dst"},
	}
	gs.Relations[faction.RelationKey("p4", "p3")] = &faction.Relation{FactionA: "p4", FactionB: "p3", Stance: faction.StanceWar}
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "siege", CurrentHP: 100}},
		},
		"third": {
			ID:            "third",
			OwnerID:       "p4",
			RegionID:      "third_src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
		},
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			StartedTurn:       5,
			FortLevel:         2,
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("third", "dst", "")

	if gs.Armies["third"].RegionID != "third_src" {
		t.Fatalf("müttefik olmayan üçüncü devlet kuşatmaya girememeliydi, got=%s", gs.Armies["third"].RegionID)
	}
	if gs.Armies["third"].MovePoints != 2 {
		t.Fatalf("başarısız girişte hareket puanı harcanmamalıydı, got=%d", gs.Armies["third"].MovePoints)
	}
	if gs.Sieges["dst"] == nil || gs.Sieges["dst"].AttackerArmyID != "atk" {
		t.Fatalf("mevcut kuşatma korunmalıydı, got=%+v", gs.Sieges["dst"])
	}
}

func TestResolveSiegesCapturesBreachedFortifiedRegion(t *testing.T) {
	gs := siegeTestState()
	_, majorThreshold := siegeBreachThresholds(2)
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "siege", CurrentHP: 100}},
		},
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			StartedTurn:       4,
			TurnsElapsed:      2,
			FortLevel:         2,
			BreachProgress:    majorThreshold,
			BreachLevel:       2,
		},
	}
	g := &Game{gs: gs}

	updates := g.resolveSieges()

	if len(updates) == 0 {
		t.Fatal("kuşatma çözümlemesi en az bir bildirim üretmeliydi")
	}
	if gs.Regions["dst"].OwnerID != "p1" {
		t.Fatalf("gedik açılmış ve savunucu yoksa bölge teslim olmalıydı, got=%s", gs.Regions["dst"].OwnerID)
	}
	if gs.Armies["atk"].RegionID != "dst" {
		t.Fatalf("kazanan ordu tahkimli bölgeye girmeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.SiegeAt("dst") != nil {
		t.Fatal("teslimiyet sonrası kuşatma kaydı temizlenmeliydi")
	}
}

func TestResolveSiegesOpensBreachVerySlowlyWithInsufficientSiegeTier(t *testing.T) {
	gs := siegeTestState()
	gs.Regions["dst"].Buildings = []string{"walls", "walls"}
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "siege", CurrentHP: 100}},
		},
		"def": {
			ID: "def", OwnerID: "p2", RegionID: "dst",
			Units: make([]army.Unit, 20),
		},
	}
	for i := range gs.Armies["def"].Units {
		gs.Armies["def"].Units[i] = army.Unit{TypeID: "inf", CurrentHP: 100}
	}
	minorThreshold, _ := siegeBreachThresholds(3)
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			DefenderArmyID:    "def",
			StartedTurn:       4,
			FortLevel:         3,
		},
	}
	g := &Game{gs: gs}

	for i := 0; i < minorThreshold-1; i++ {
		g.resolveSieges()
	}

	siege := gs.SiegeAt("dst")
	if siege == nil {
		t.Fatal("kuşatma kaydı korunmalıydı")
	}
	if siege.BreachProgress >= minorThreshold {
		t.Fatalf("gedik eşiği dolmadan breach açılmamalıydı, got progress=%d threshold=%d", siege.BreachProgress, minorThreshold)
	}

	g.resolveSieges()

	siege = gs.SiegeAt("dst")
	if siege == nil {
		t.Fatal("kuşatma kaydı korunmalıydı")
	}
	if siege.BreachLevel < 1 {
		t.Fatalf("yetersiz siege tier ile gedik çok yavaş da olsa açılmalıydı, got progress=%d level=%d", siege.BreachProgress, siege.BreachLevel)
	}
}

func TestResolveSiegesCanStarveFortWithoutBreach(t *testing.T) {
	gs := siegeTestState()
	gs.Regions["dst"].Buildings = []string{"walls", "walls"}
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
		},
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			StartedTurn:       1,
			TurnsElapsed:      siegeSurrenderTurns(3) - 1,
			FortLevel:         3,
		},
	}
	g := &Game{gs: gs}

	g.resolveSieges()

	if gs.Regions["dst"].OwnerID != "p1" {
		t.Fatalf("uzun kuşatma açlık teslimiyeti getirmeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
}

func TestAssaultSiegeWithoutSiegeUnitIsAllowed(t *testing.T) {
	gs := siegeTestState()
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
		},
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			StartedTurn:       4,
			TurnsElapsed:      1,
			FortLevel:         2,
			BreachProgress:    0,
			BreachLevel:       0,
		},
	}
	g := &Game{gs: gs}

	g.assaultSiegeWithStance("atk", "dst", "")

	if gs.Regions["dst"].OwnerID != "p2" {
		t.Fatalf("gedik yokken kuşatma birimi olmadan genel hücum kale fethi vermemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
	if gs.Armies["atk"].RegionID != "src" {
		t.Fatalf("gedik yokken kuşatma birimi olmadan genel hücumda ordu hedefe girmemeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.SiegeAt("dst") == nil {
		t.Fatal("gediksiz genel hücumdan sonra aktif kuşatma sürmeliydi")
	}
}

func TestAssaultSiegeWithoutBreachCannotCaptureFortifiedRegion(t *testing.T) {
	gs := siegeTestState()
	gs.UnitTypes["elite"] = &army.UnitType{ID: "elite", Category: army.CategoryInfantry, Attack: 90, Defense: 90, Morale: 100}
	gs.Armies = map[army.ArmyID]*army.Army{
		"atk": {
			ID:            "atk",
			OwnerID:       "p1",
			RegionID:      "src",
			MovePoints:    2,
			MaxMovePoints: 2,
			Units: []army.Unit{
				{TypeID: "elite", CurrentHP: 100},
				{TypeID: "elite", CurrentHP: 100},
				{TypeID: "elite", CurrentHP: 100},
				{TypeID: "elite", CurrentHP: 100},
				{TypeID: "elite", CurrentHP: 100},
				{TypeID: "elite", CurrentHP: 100},
				{TypeID: "siege", CurrentHP: 100},
			},
		},
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			StartedTurn:       4,
			TurnsElapsed:      1,
			FortLevel:         2,
			BreachProgress:    0,
			BreachLevel:       0,
		},
	}
	g := &Game{gs: gs}

	g.assaultSiegeWithStance("atk", "dst", "")

	if gs.Regions["dst"].OwnerID != "p2" {
		t.Fatalf("gedik yokken genel hücum kale fethi vermemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
	if gs.Armies["atk"].RegionID != "src" {
		t.Fatalf("başarısız gediksiz hücum sonrası ordu hedefe girmemeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.SiegeAt("dst") == nil {
		t.Fatal("aktif kuşatma gediksiz hücumdan sonra sürmeliydi")
	}
}

func TestSiegeAssaultAttackerDamageDropsAsBreachGrows(t *testing.T) {
	noBreach := siegeAssaultAttackerDamage(3, 0)
	minorBreach := siegeAssaultAttackerDamage(3, 1)
	majorBreach := siegeAssaultAttackerDamage(3, 2)

	if noBreach <= minorBreach {
		t.Fatalf("gedik yokken hücum kaybı küçük gedikten yüksek olmalıydı, got no_breach=%d minor=%d", noBreach, minorBreach)
	}
	if minorBreach <= majorBreach {
		t.Fatalf("küçük gedik hücum kaybı büyük gedikten yüksek olmalıydı, got minor=%d major=%d", minorBreach, majorBreach)
	}
}
