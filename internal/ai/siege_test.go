package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiSiegeTestState(withSiegeUnit bool) *state.GameState {
	units := []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}
	if withSiegeUnit {
		units = append(units, army.Unit{TypeID: "siege", CurrentHP: 100})
	}
	return &state.GameState{
		Turn:            3,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"src": {ID: "src", OwnerID: "ai_1", Neighbors: []world.RegionID{"fort"}},
			"fort": {
				ID:          "fort",
				OwnerID:     "player",
				Neighbors:   []world.RegionID{"src"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fortress", Type: world.SettlementFortress, NameTR: "Hisar"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {
				ID:            "ai_army",
				OwnerID:       "ai_1",
				RegionID:      "src",
				Units:         units,
				MovePoints:    2,
				MaxMovePoints: 2,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -70, Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":   {ID: "inf", Category: army.CategoryInfantry, Attack: 14, Defense: 12, Morale: 55},
			"siege": {ID: "siege", Category: army.CategorySiege, Tier: 1, Attack: 8, Defense: 4, Morale: 30},
		},
	}
}

func TestChooseBestMoveCanSelectFortifiedTargetWithoutSiegeUnit(t *testing.T) {
	gs := aiSiegeTestState(false)

	target := chooseBestMove(gs, gs.Armies["ai_army"])

	if target != "fort" {
		t.Fatalf("siege birimi olmayan AI tahkimli hedefi kuşatma adayı olarak seçebilmeliydi, got=%s", target)
	}
}

func TestExecuteMoveStartsSiegeOnFortifiedTarget(t *testing.T) {
	gs := aiSiegeTestState(false)
	a := gs.Armies["ai_army"]

	outcome := executeMove(gs, a, "fort", "ai_1")

	if !outcome.survived {
		t.Fatal("kuşatma başlatan AI ordusu hayatta kalmalıydı")
	}
	if gs.SiegeAt("fort") == nil {
		t.Fatal("AI tahkimli hedefte kuşatma başlatmalıydı")
	}
	if a.RegionID != "src" {
		t.Fatalf("kuşatma başlatan AI ordusu hedefe girmemeli, got=%s", a.RegionID)
	}
	if a.MovePoints != 0 {
		t.Fatalf("kuşatma sonrası hareket puanı bitmeliydi, got=%d", a.MovePoints)
	}
}

func TestExecuteMoveDefersAISortieAgainstPlayerSiege(t *testing.T) {
	gs := aiSiegeTestState(false)
	gs.Regions["fort"].OwnerID = "ai_1"
	gs.Armies["ai_army"].RegionID = "fort"
	gs.Armies["ai_army"].MovePoints = 2
	gs.Armies["player_siege"] = &army.Army{
		ID: "player_siege", OwnerID: "player", RegionID: "fort",
		Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}, MovePoints: 0, MaxMovePoints: 2,
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"fort": {RegionID: "fort", AttackerArmyID: "player_siege", AttackerFactionID: "player", DefenderArmyID: "ai_army"},
	}

	outcome := executeMove(gs, gs.Armies["ai_army"], "src", "ai_1")

	if !outcome.survived || outcome.step.Kind != TurnStepSortie {
		t.Fatalf("oyuncu kuşatmasına karşı AI huruç kararı bekletilmeliydi: outcome=%+v", outcome)
	}
	if gs.SiegeAt("fort") == nil {
		t.Fatal("oyuncu kararı gelmeden kuşatma çözülmemeliydi")
	}
	if len(gs.Armies["ai_army"].Units) != 2 || len(gs.Armies["player_siege"].Units) != 1 {
		t.Fatal("oyuncu kararı gelmeden huruç savaşı çözülmemeliydi")
	}
}

func TestExecuteMoveCreatesContactBeforeSiegeOnFortifiedTargetWithDefender(t *testing.T) {
	gs := aiSiegeTestState(false)
	gs.Armies["defender"] = &army.Army{
		ID:            "defender",
		OwnerID:       "player",
		RegionID:      "fort",
		Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
		MovePoints:    2,
		MaxMovePoints: 2,
	}
	a := gs.Armies["ai_army"]

	outcome := executeMove(gs, a, "fort", "ai_1")

	if !outcome.survived || gs.PendingLandContact == nil {
		t.Fatalf("AI tahkimli hedefte önce kara teması oluşturmalıydı: outcome=%+v contact=%+v", outcome, gs.PendingLandContact)
	}
	if gs.SiegeAt("fort") != nil {
		t.Fatal("temas kararı verilmeden tahkimli hedefte doğrudan kuşatma başlamamalı")
	}
	if a.RegionID != "fort" {
		t.Fatalf("temas sırasında AI ordusu hedefte görünmeli, got=%s", a.RegionID)
	}
	if a.MovePoints != 1 {
		t.Fatalf("temas popup'ı açılırken hareket puanı tüketilmeli, got=%d", a.MovePoints)
	}
}

func TestAIBesiegerCanOfferPlayerSiegeSurrender(t *testing.T) {
	gs := aiSiegeTestState(false)
	gs.Regions["player_second"] = &world.Region{ID: "player_second", OwnerID: "player"}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"fort": {RegionID: "fort", AttackerArmyID: "ai_army", AttackerFactionID: "ai_1", TurnsElapsed: 3, BreachLevel: 1, FortLevel: 2},
	}
	gs.Armies["ai_army"].RegionID = "fort"
	gs.Armies["defender"] = &army.Army{
		ID: "defender", OwnerID: "player", RegionID: "fort", MovePoints: 0, MaxMovePoints: 2,
		Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}

	for turn := 1; turn <= 100; turn++ {
		gs.Turn = turn
		gs.DiplomaticOffers = nil
		gs.DiplomacyOfferCounts = nil
		gs.OfferRejectionTurns = nil
		var steps []TurnStep
		aiHandleDiplomacyWithSteps(gs, "ai_1", &steps)
		if len(gs.DiplomaticOffers) == 0 {
			continue
		}
		foundSurrender := false
		for _, offer := range gs.DiplomaticOffers {
			if offer.Action == string(diplomacy.ActionProposeSurrender) && offer.RegionID == "fort" && offer.ToFactionID == "player" {
				foundSurrender = true
				break
			}
		}
		if !foundSurrender {
			continue
		}
		focused := false
		for _, step := range steps {
			if step.Kind == TurnStepDiplomacy && step.TargetRegion == "fort" && step.FocusRegion == "fort" {
				focused = true
				break
			}
		}
		if !focused {
			t.Fatal("AI teslimiyet teklifinin turn adımı kuşatılan bölgeyi kamera odağı olarak taşımalıydı")
		}
		return
	}
	t.Fatal("AI kuşatanı uygun koşullarda oyuncuya teslimiyet teklifi üretmedi")
}

func TestAIPeaceOfferIsQueuedBeforeSiegeSurrenderOffer(t *testing.T) {
	gs := aiSiegeTestState(false)
	gs.Turn = 8
	gs.BeginWarLedger("ai_1", "player")
	gs.WarLedgerFor("ai_1", "player").StartedTurn = 0
	gs.Regions["player_second"] = &world.Region{ID: "player_second", OwnerID: "player"}
	gs.Relations[faction.RelationKey("ai_1", "player")].Score = -100
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"fort": {RegionID: "fort", AttackerArmyID: "ai_army", AttackerFactionID: "ai_1", TurnsElapsed: 3, BreachLevel: 1, FortLevel: 2},
	}
	gs.Armies["ai_army"].RegionID = "fort"
	gs.Armies["defender"] = &army.Army{
		ID: "defender", OwnerID: "player", RegionID: "fort", MovePoints: 0, MaxMovePoints: 2,
		Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}

	for turn := 1; turn <= 100; turn++ {
		gs.Turn = turn
		gs.DiplomaticOffers = nil
		gs.DiplomacyOfferCounts = nil
		var steps []TurnStep
		aiHandleDiplomacyWithSteps(gs, "ai_1", &steps)
		if len(gs.DiplomaticOffers) != 2 {
			continue
		}
		if gs.DiplomaticOffers[0].Action != string(diplomacy.ActionProposePeace) || gs.DiplomaticOffers[1].Action != string(diplomacy.ActionProposeSurrender) {
			t.Fatalf("barış önce, teslimiyet sonra kuyruğa alınmalıydı: %+v", gs.DiplomaticOffers)
		}
		return
	}
	t.Fatal("aynı savaşta barış ve kuşatma teslimiyeti tekliflerinin birlikte üretildiği tur bulunamadı")
}

func TestAILastRegionSiegeOffersVassalization(t *testing.T) {
	gs := aiSiegeTestState(false)
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"fort": {RegionID: "fort", AttackerArmyID: "ai_army", AttackerFactionID: "ai_1", TurnsElapsed: 3, BreachLevel: 1, FortLevel: 2},
	}
	gs.Armies["ai_army"].RegionID = "fort"
	gs.Armies["defender"] = &army.Army{
		ID: "defender", OwnerID: "player", RegionID: "fort", MovePoints: 0, MaxMovePoints: 2,
		Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}

	for turn := 1; turn <= 100; turn++ {
		gs.Turn = turn
		gs.DiplomaticOffers = nil
		gs.DiplomacyOfferCounts = nil
		gs.OfferRejectionTurns = nil
		var steps []TurnStep
		aiHandleDiplomacyWithSteps(gs, "ai_1", &steps)
		for _, offer := range gs.DiplomaticOffers {
			if offer.Action == string(diplomacy.ActionProposeSiegeVassalization) && offer.RegionID == "fort" {
				return
			}
		}
	}
	t.Fatalf("son toprak için kuşatma vassallığı teklifi üretilmeliydi: %+v", gs.DiplomaticOffers)
}

func TestAISiegeSurrenderRetryCooldownIsRegionScoped(t *testing.T) {
	gs := &state.GameState{
		Turn: 5,
		OfferRejectionTurns: map[string]int{
			state.DiplomaticOfferRegionRejectionKey("ai_1", "player", string(diplomacy.ActionProposeSurrender), "north"): 5,
		},
	}
	if aiDiplomacyOfferRetryAllowedForRegion(gs, "ai_1", "player", diplomacy.ActionProposeSurrender, "north") {
		t.Fatal("reddedilen aynı kuşatma bölgesi cooldown sırasında tekrar teklif edilmemeli")
	}
	if !aiDiplomacyOfferRetryAllowedForRegion(gs, "ai_1", "player", diplomacy.ActionProposeSurrender, "south") {
		t.Fatal("başka kuşatma bölgesi reddedilen bölgenin cooldown'undan etkilenmemeli")
	}
}

func TestMoveArmyStopsAfterFailedGeneralAssault(t *testing.T) {
	gs := aiSiegeTestState(true)
	attacker := gs.Armies["ai_army"]
	attacker.Units = attacker.Units[:0]
	for i := 0; i < 20; i++ {
		unitType := "inf"
		if i == 0 {
			unitType = "siege"
		}
		attacker.Units = append(attacker.Units, army.Unit{TypeID: unitType, CurrentHP: 100})
	}
	gs.UnitTypes["inf"].Attack = 1
	gs.UnitTypes["siege"].Attack = 1
	gs.UnitTypes["elite"] = &army.UnitType{ID: "elite", Category: army.CategoryInfantry, Attack: 100, Defense: 100, Morale: 100}
	defenderUnits := make([]army.Unit, 20)
	for i := range defenderUnits {
		defenderUnits[i] = army.Unit{TypeID: "elite", CurrentHP: 100}
	}
	gs.Armies["defender"] = &army.Army{
		ID:       "defender",
		OwnerID:  "player",
		RegionID: "fort",
		Units:    defenderUnits,
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"fort": {
			RegionID:          "fort",
			AttackerArmyID:    attacker.ID,
			AttackerFactionID: attacker.OwnerID,
			DefenderArmyID:    "defender",
			FortLevel:         1,
			BreachLevel:       2,
		},
	}

	moveArmyWithSteps(gs, attacker, "ai_1", nil)

	if gs.Armies[attacker.ID] == nil {
		t.Fatal("başarısız ilk genel hücumdan sonra hayatta kalan ordu aynı turda tekrar tekrar saldırmamalı")
	}
	if attacker.MovePoints != 0 {
		t.Fatalf("başarısız genel hücum AI hareketini bitirmeli, got=%d", attacker.MovePoints)
	}
	if len(attacker.Units) == 0 {
		t.Fatal("test ordusu ilk başarısız hücumdan sonra kısmen hayatta kalmalı")
	}
}

func TestChooseBestMoveAndExecuteMoveCanBreakSiegeWithoutConquestInAlliedRegion(t *testing.T) {
	gs := &state.GameState{
		Turn:            3,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "ai_1",
				Neighbors: []world.RegionID{"fort"},
			},
			"fort": {
				ID:          "fort",
				OwnerID:     "ally",
				Neighbors:   []world.RegionID{"src"},
				Buildings:   []string{"walls"},
				Terrain:     world.TerrainPlain,
				Settlements: []world.Settlement{{ID: "fortress", Type: world.SettlementFortress, NameTR: "Hisar"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {
				ID:            "ai_army",
				OwnerID:       "ai_1",
				RegionID:      "src",
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
			"besieger": {
				ID:            "besieger",
				OwnerID:       "player",
				RegionID:      "fort",
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
			"ally":   {ID: "ally", NameTR: "Müttefik", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -70, Stance: faction.StanceWar},
			faction.RelationKey("ai_1", "ally"):   {FactionA: "ai_1", FactionB: "ally", Score: 60, Stance: faction.StanceAllied},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {
				RegionID:          "fort",
				AttackerArmyID:    "besieger",
				AttackerFactionID: "player",
				StartedTurn:       3,
				FortLevel:         2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":   {ID: "inf", Category: army.CategoryInfantry, Attack: 14, Defense: 12, Morale: 55},
			"siege": {ID: "siege", Category: army.CategorySiege, Tier: 1, Attack: 8, Defense: 4, Morale: 30},
		},
	}

	target := chooseBestMove(gs, gs.Armies["ai_army"])
	if target != "fort" {
		t.Fatalf("AI kuşatma altındaki müttefik bölgede düşman besiegerı hedefleyebilmeliydi, got=%s", target)
	}

	outcome := executeMove(gs, gs.Armies["ai_army"], target, "ai_1")

	if !outcome.survived {
		t.Fatal("kuşatma savunmasını kıran AI ordusu hayatta kalmalıydı")
	}
	if gs.Armies["ai_army"].RegionID != "fort" {
		t.Fatalf("AI orduyu hedefe sokmalıydı, got=%s", gs.Armies["ai_army"].RegionID)
	}
	if gs.Regions["fort"].OwnerID != "ally" {
		t.Fatalf("kuşatma kaldırılırken bölge sahipliği değişmemeliydi, got=%s", gs.Regions["fort"].OwnerID)
	}
	if gs.SiegeAt("fort") != nil {
		t.Fatal("besieging army yenildiğinde kuşatma kaydı temizlenmeliydi")
	}
}

func TestExecuteMoveBlocksAllyTransitIntoBesiegedRegionWithoutWarWithBesieger(t *testing.T) {
	gs := &state.GameState{
		Turn:            3,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "ally",
				Neighbors: []world.RegionID{"fort"},
			},
			"fort": {
				ID:          "fort",
				OwnerID:     "p2",
				Neighbors:   []world.RegionID{"src"},
				Buildings:   []string{"walls"},
				Terrain:     world.TerrainPlain,
				Settlements: []world.Settlement{{ID: "fortress", Type: world.SettlementFortress, NameTR: "Hisar"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ally_army": {
				ID:            "ally_army",
				OwnerID:       "ally",
				RegionID:      "src",
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ally": {ID: "ally", NameTR: "Müttefik", Religion: religion.Sunni},
			"p1":   {ID: "p1", NameTR: "Besieger", Religion: religion.Catholic},
			"p2":   {ID: "p2", NameTR: "Hedef", Religion: religion.Catholic},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ally", "p2"): {FactionA: "ally", FactionB: "p2", Score: 60, Stance: faction.StanceAllied},
			faction.RelationKey("ally", "p1"): {FactionA: "ally", FactionB: "p1", Score: 0, Stance: faction.StancePeace},
			faction.RelationKey("p1", "p2"):   {FactionA: "p1", FactionB: "p2", Score: -70, Stance: faction.StanceWar},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {
				RegionID:          "fort",
				AttackerArmyID:    "besieger",
				AttackerFactionID: "p1",
				StartedTurn:       3,
				FortLevel:         2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 14, Defense: 12, Morale: 55},
		},
	}

	target := chooseBestMove(gs, gs.Armies["ally_army"])
	if target != "" {
		t.Fatalf("AI kuşatan tarafla savaşta olmayan müttefik geçişi seçmemeliydi, got=%s", target)
	}

	outcome := executeMove(gs, gs.Armies["ally_army"], "fort", "ally")

	if !outcome.survived {
		t.Fatal("engellenen transit ordusu hayatta kalmalıydı")
	}
	if gs.Armies["ally_army"].RegionID != "src" {
		t.Fatalf("kuşatanla savaşta olmayan müttefik kuşatılmış bölgeye girememeliydi, got=%s", gs.Armies["ally_army"].RegionID)
	}
	if gs.Regions["fort"].OwnerID != "p2" {
		t.Fatalf("engellenen transit sahipliği değiştirmemeliydi, got=%s", gs.Regions["fort"].OwnerID)
	}
	if gs.SiegeAt("fort") == nil || gs.SiegeAt("fort").AttackerArmyID != "besieger" {
		t.Fatalf("aktif kuşatma korunmalıydı, got=%+v", gs.SiegeAt("fort"))
	}
}

func TestExecuteMoveAlliedSiegeSupportCannotConquerBesiegedRegion(t *testing.T) {
	gs := aiSiegeTestState(false)
	gs.Regions["src"].OwnerID = "ally"
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Müttefik", Religion: religion.Sunni}
	gs.Armies["ai_army"].OwnerID = "ally"
	gs.Armies["ai_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}
	gs.Armies["besieger"] = &army.Army{
		ID: "besieger", OwnerID: "ai_1", RegionID: "src",
		Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}
	gs.Armies["defender"] = &army.Army{
		ID: "defender", OwnerID: "player", RegionID: "fort",
		Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"fort": {RegionID: "fort", AttackerArmyID: "besieger", AttackerFactionID: "ai_1", FortLevel: 2},
	}
	gs.Relations[faction.RelationKey("ally", "player")] = &faction.Relation{
		FactionA: "ally", FactionB: "player", Stance: faction.StanceWar,
	}
	gs.Relations[faction.RelationKey("ally", "ai_1")] = &faction.Relation{
		FactionA: "ally", FactionB: "ai_1", Stance: faction.StanceAllied,
	}
	gs.Relations[faction.RelationKey("ai_1", "player")] = &faction.Relation{
		FactionA: "ai_1", FactionB: "player", Stance: faction.StanceWar,
	}

	outcome := executeMove(gs, gs.Armies["ai_army"], "fort", "ally")

	if !outcome.survived {
		t.Fatal("müttefik destek ordusu savunma zaferinden sonra hayatta kalmalıydı")
	}
	if got := gs.Regions["fort"].OwnerID; got != "player" {
		t.Fatalf("müttefik destek ordusu kuşatılmış bölgeyi fethetmemeliydi, got=%s", got)
	}
	if siege := gs.SiegeAt("fort"); siege == nil || siege.AttackerArmyID != "besieger" {
		t.Fatalf("ilk kuşatma korunmalıydı, got=%+v", siege)
	}
}

func TestExecuteMoveAlliedDefenderSortiesAndLeavesBesiegedRegionAfterVictory(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"fort": {ID: "fort", OwnerID: "player", Neighbors: []world.RegionID{"exit"}},
			"exit": {ID: "exit", OwnerID: "ally", Neighbors: []world.RegionID{"fort"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"defender": {
				ID: "defender", OwnerID: "ally", RegionID: "fort", MovePoints: 2, MaxMovePoints: 2,
				Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
			},
			"besieger": {
				ID: "besieger", OwnerID: "enemy", RegionID: "fort", MovePoints: 2, MaxMovePoints: 2,
				Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Religion: religion.Catholic},
			"ally":   {ID: "ally", Religion: religion.Catholic},
			"enemy":  {ID: "enemy", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ally", "player"): {FactionA: "ally", FactionB: "player", Stance: faction.StanceAllied},
			faction.RelationKey("ally", "enemy"):  {FactionA: "ally", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {RegionID: "fort", AttackerArmyID: "besieger", AttackerFactionID: "enemy"},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}

	outcome := executeMove(gs, gs.Armies["defender"], "exit", "ally")
	if !outcome.survived {
		t.Fatal("müttefik savunmacı huruç zaferinde hayatta kalmalıydı")
	}
	if gs.Armies["defender"].RegionID != "exit" {
		t.Fatalf("AI müttefik savunmacı huruç sonrası çıkmalıydı, got=%s", gs.Armies["defender"].RegionID)
	}
	if gs.SiegeAt("fort") != nil {
		t.Fatal("huruç zaferinde aktif kuşatma temizlenmeliydi")
	}
}
