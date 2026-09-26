package game

import (
	"testing"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestAutoStartResearchIfIdleStartsNextResearchableTech(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:   "player",
				Gold: 25,
				Research: faction.ResearchState{
					Completed: map[string]bool{"root": true},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"root": {
				ID:       "root",
				Category: tech.CategoryMilitary,
			},
			"smithing": {
				ID:            "smithing",
				NameTR:        "Demircilik",
				Category:      tech.CategoryMilitary,
				Requires:      []string{"root"},
				GoldCost:      20,
				TurnsRequired: 4,
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if started := g.autoStartResearchIfIdle(); !started {
		t.Fatal("uygun teknoloji varken otomatik baslatma calismaliydi")
	}

	player := gs.Factions["player"]
	if player.Research.ActiveID != "smithing" {
		t.Fatalf("beklenen aktif research smithing olmaliydi, got=%q", player.Research.ActiveID)
	}
	if player.Research.TurnsLeft != 4 {
		t.Fatalf("turn sayaci teknoloji turu kadar ayarlanmaliydi, got=%d", player.Research.TurnsLeft)
	}
	if player.Gold != 5 {
		t.Fatalf("gold otomatik baslatmada dusmeliydi, got=%d", player.Gold)
	}
}

func TestMoveDockedFleetToSamePortDoesNotConsumeMovement(t *testing.T) {
	const (
		sea  = world.RegionID("sea")
		port = world.RegionID("port")
	)

	fleet := &army.Army{
		ID:             "fleet",
		OwnerID:        "player",
		IsNaval:        true,
		RegionID:       sea,
		DockedRegionID: port,
		MovePoints:     2,
		MaxMovePoints:  2,
	}
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Regions: map[world.RegionID]*world.Region{
			sea:  {ID: sea, IsSea: true, Neighbors: []world.RegionID{port}},
			port: {ID: port, OwnerID: "player", Neighbors: []world.RegionID{sea}},
		},
		Armies: map[army.ArmyID]*army.Army{fleet.ID: fleet},
	}

	(&Game{gs: gs}).moveArmyToSettlementWithStance(fleet.ID, port, "", combat.BattleStanceBalanced)

	if fleet.MovePoints != 2 {
		t.Fatalf("aynı limana hareket emri hareket puanını değiştirdi: got=%d want=2", fleet.MovePoints)
	}
	if fleet.DockedRegionID != port {
		t.Fatalf("filo aynı limandan ayrıldı: got=%q want=%q", fleet.DockedRegionID, port)
	}
}

func TestShouldOfferPostWarVassalizationRejectsSeaRegion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "venice",
		Factions: map[faction.FactionID]*faction.Faction{
			"venice":      {ID: "venice"},
			"hospitaller": {ID: "hospitaller"},
		},
		Regions: map[world.RegionID]*world.Region{
			"rhodes": {ID: "rhodes", OwnerID: "hospitaller"},
		},
	}
	game := &Game{gs: gs}

	sea := &world.Region{ID: "dardanelles_strait", IsSea: true, OwnerID: "hospitaller"}
	if game.shouldOfferPostWarVassalization("venice", "hospitaller", sea) {
		t.Fatal("deniz bölgesi için savaş sonrası kara fetih kararı üretildi")
	}
}

func TestQuickTurnSuppressesOnlyRelationshipNotifications(t *testing.T) {
	gs := &state.GameState{
		DiplomaticOffers: []state.DiplomaticOffer{
			{Action: string(diplomacy.ActionImproveRelations)},
			{Action: string(diplomacy.ActionSendGift)},
			{Action: string(diplomacy.ActionProposePeace)},
			{Action: string(diplomacy.ActionProposeTrade)},
			{Action: string(diplomacy.ActionOfferVassalization)},
		},
	}
	r := render.New(gs)
	r.CurrentSettings.FastAITurns = true
	g := &Game{gs: gs, renderer: r}

	g.suppressQuickTurnRelationshipNotifications()

	if got, want := len(gs.DiplomaticOffers), 3; got != want {
		t.Fatalf("hızlı turda kalan teklif sayısı: got=%d want=%d", got, want)
	}
	for _, offer := range gs.DiplomaticOffers {
		if diplomacy.IsRelationshipNotification(diplomacy.Action(offer.Action)) {
			t.Fatalf("ilişki bildirimi hızlı tur kuyruğunda kaldı: %q", offer.Action)
		}
	}
}

func TestAutoStartResearchIfIdleIgnoresPausedTechsWhenAnotherTechIsAvailable(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:   "player",
				Gold: 25,
				Research: faction.ResearchState{
					Completed:   map[string]bool{"root": true},
					PausedTurns: map[string]int{"paused": 3},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"root": {
				ID:       "root",
				Category: tech.CategoryMilitary,
			},
			"paused": {
				ID:            "paused",
				NameTR:        "Duraklatilmis",
				Category:      tech.CategoryMilitary,
				Requires:      []string{"root"},
				GoldCost:      20,
				TurnsRequired: 4,
			},
			"smithing": {
				ID:            "smithing",
				NameTR:        "Demircilik",
				Category:      tech.CategoryMilitary,
				Requires:      []string{"root"},
				GoldCost:      20,
				TurnsRequired: 4,
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if started := g.autoStartResearchIfIdle(); !started {
		t.Fatal("duraklatilmis tech olsa da uygun baska tech varken otomatik baslatma calismaliydi")
	}

	player := gs.Factions["player"]
	if player.Research.ActiveID != "smithing" {
		t.Fatalf("paused tech otomatik secimi kilitlememeliydi, got=%q", player.Research.ActiveID)
	}
	if player.Gold != 5 {
		t.Fatalf("gold otomatik baslatmada dusmeliydi, got=%d", player.Gold)
	}
	if player.Research.PausedTurns["paused"] != 3 {
		t.Fatalf("duraklatilmis tech kaydi korunmaliydi, got=%d", player.Research.PausedTurns["paused"])
	}
}

func TestFindRecruitableLandArmyUsesManpowerInsteadOfArmySlotLimit(t *testing.T) {
	const fid = faction.FactionID("test")
	regions := map[world.RegionID]*world.Region{
		"home":  {ID: "home", OwnerID: string(fid)},
		"front": {ID: "front", OwnerID: string(fid)},
	}
	makeArmy := func(id army.ArmyID, rid world.RegionID) *army.Army {
		return &army.Army{
			ID: id, OwnerID: string(fid), RegionID: rid,
			Units: []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
		}
	}
	gs := &state.GameState{
		Regions: regions,
		Armies: map[army.ArmyID]*army.Army{
			"army-1": makeArmy("army-1", "home"),
			"army-2": makeArmy("army-2", "front"),
		},
	}
	g := &Game{gs: gs}

	if _, canCreate := g.findRecruitableLandArmy("home", fid); !canCreate {
		t.Fatalf("boş savaşçı kapasitesi varken ordu slotu sınırı üretimi engellememeli")
	}

	for _, a := range gs.Armies {
		for len(a.Units) < gs.ManpowerCap(fid)/2 {
			a.Units = append(a.Units, army.Unit{TypeID: "militia", CurrentHP: army.MaxUnitHP})
		}
	}
	if _, canCreate := g.findRecruitableLandArmy("home", fid); canCreate {
		t.Fatalf("maksimum savaşçı kapasitesi doluyken yeni ordu açılabilmemeli")
	}
}

func TestCheckRegionUnlocksDoesNotUnlockTerrainAreas(t *testing.T) {
	source := &world.Region{
		ID:        "florence",
		Neighbors: []world.RegionID{"area::apennin", "hidden_region"},
	}
	terrain := &world.Region{
		ID:            "area::apennin",
		IsTerrainArea: true,
		TerrainAreaID: "apennin",
		IsLocked:      true,
	}
	normal := &world.Region{ID: "hidden_region", IsLocked: true}
	armyRef := &army.Army{ID: "army", OwnerID: "florence_rep", RegionID: source.ID}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			source.ID:  source,
			terrain.ID: terrain,
			normal.ID:  normal,
		},
		Armies: map[army.ArmyID]*army.Army{armyRef.ID: armyRef},
	}

	unlocked := checkRegionUnlocks(gs)
	if !terrain.IsLocked {
		t.Fatal("terrain area was unlocked by a neighboring army")
	}
	if len(unlocked) != 1 || unlocked[0] != normal.ID {
		t.Fatalf("unexpected unlocked regions: %#v", unlocked)
	}
}

func TestResolveSiegesForcesSurrenderWhenDurationExpiresWithDefender(t *testing.T) {
	const (
		attackerID = faction.FactionID("venice")
		defenderID = faction.FactionID("arnavut_des")
	)
	target := &world.Region{
		ID:        "albania",
		NameTR:    "Arnavutluk",
		OwnerID:   string(defenderID),
		Neighbors: []world.RegionID{"arnavut_home"},
		Buildings: []string{"walls", "walls", "granary"},
	}
	home := &world.Region{
		ID:        "arnavut_home",
		OwnerID:   string(defenderID),
		Neighbors: []world.RegionID{target.ID},
		WorldX:    -100,
	}
	attacker := &army.Army{
		ID:       "venice_army",
		OwnerID:  string(attackerID),
		RegionID: target.ID,
		Units:    []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
	}
	defender := &army.Army{
		ID:        "albania_garrison",
		OwnerID:   string(defenderID),
		RegionID:  target.ID,
		Commander: &army.Commander{ID: "albania_commander", OwnerID: string(defenderID)},
		Units:     []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
	}
	siege := &state.SiegeState{
		RegionID:          target.ID,
		AttackerArmyID:    attacker.ID,
		DefenderArmyID:    defender.ID,
		AttackerFactionID: string(attackerID),
		TurnsElapsed:      4,
		FortLevel:         2,
		GranaryLevel:      1,
	}
	gs := &state.GameState{
		PlayerFactionID: attackerID,
		Regions:         map[world.RegionID]*world.Region{target.ID: target, home.ID: home},
		Armies:          map[army.ArmyID]*army.Army{attacker.ID: attacker, defender.ID: defender},
		Factions: map[faction.FactionID]*faction.Faction{
			attackerID: {ID: attackerID},
			defenderID: {ID: defenderID},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(attackerID, defenderID): {
				FactionA: attackerID,
				FactionB: defenderID,
				Stance:   faction.StanceWar,
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{target.ID: siege},
	}

	if got := siege.TurnsUntilSurrender(); got != 1 {
		t.Fatalf("son çözümleme turu öncesi 1 tur kalmalıydı, got=%d", got)
	}

	(&Game{gs: gs}).resolveSieges()

	if _, ok := gs.Sieges[target.ID]; ok {
		t.Fatal("zorunlu teslimiyet sonrasında kuşatma kaydı kalmamalıydı")
	}
	if target.OwnerID != string(attackerID) {
		t.Fatalf("süre dolunca bölge kuşatmacıya geçmeliydi, owner=%q", target.OwnerID)
	}
	transferred := gs.Armies[defender.ID]
	if transferred == nil {
		t.Fatal("teslim olan savunma ordusu kazananın kuvvetlerine devredilmeliydi")
	}
	if transferred.OwnerID != string(attackerID) {
		t.Fatalf("savunma ordusunun sahibi kuşatan olmalıydı, owner=%q", transferred.OwnerID)
	}
	if transferred.RegionID != target.ID {
		t.Fatalf("savunma ordusu teslim çözümlemesinde otomatik kaçmamalıydı, region=%q", transferred.RegionID)
	}
	if transferred.Commander == nil || transferred.Commander.OwnerID != string(attackerID) {
		t.Fatal("devredilen ordunun komutan sahipliği de güncellenmeliydi")
	}
}

func TestResolveSiegesClearsStaleDefenderReference(t *testing.T) {
	attackerID := faction.FactionID("venice")
	defenderID := faction.FactionID("epir")
	target := &world.Region{
		ID:        "epirus",
		OwnerID:   string(defenderID),
		Neighbors: []world.RegionID{"yanya"},
	}
	otherRegion := &world.Region{
		ID:        "yanya",
		OwnerID:   string(defenderID),
		Neighbors: []world.RegionID{target.ID},
	}
	attacker := &army.Army{
		ID:       "venice_army",
		OwnerID:  string(attackerID),
		RegionID: target.ID,
		Units:    []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
	}
	defender := &army.Army{
		ID:       "epirus_army",
		OwnerID:  string(defenderID),
		RegionID: otherRegion.ID,
		Units:    []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
	}
	siege := &state.SiegeState{
		RegionID:          target.ID,
		AttackerArmyID:    attacker.ID,
		DefenderArmyID:    defender.ID,
		AttackerFactionID: string(attackerID),
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			target.ID:      target,
			otherRegion.ID: otherRegion,
		},
		Armies: map[army.ArmyID]*army.Army{
			attacker.ID: attacker,
			defender.ID: defender,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			attackerID: {ID: attackerID},
			defenderID: {ID: defenderID},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(attackerID, defenderID): {
				FactionA: attackerID,
				FactionB: defenderID,
				Stance:   faction.StanceWar,
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{target.ID: siege},
	}

	(&Game{gs: gs}).resolveSieges()

	if siege.DefenderArmyID != "" {
		t.Fatalf("bölgeden ayrılmış savunmacı bağlantısı temizlenmeli, got=%q", siege.DefenderArmyID)
	}
}

func TestExecutePlayerNavalMissionsConvertsInvalidBlockadeToPatrol(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"enemy_port": {
				ID:          "enemy_port",
				OwnerID:     "enemy",
				Neighbors:   []world.RegionID{"sea"},
				Settlements: []world.Settlement{{ID: "enemy_harbor", Type: world.SettlementPort}},
			},
			"sea": {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"enemy_port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID:         "fleet",
				OwnerID:    "player",
				RegionID:   "sea",
				IsNaval:    true,
				MovePoints: 2,
				NavalMission: &army.NavalMission{
					Kind:           army.NavalMissionBlockade,
					TargetRegionID: "sea",
				},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player",
				FactionB: "enemy",
				Stance:   faction.StanceWar,
			},
		},
	}
	game := &Game{gs: gs}

	gs.Regions["enemy_port"].OwnerID = "player"
	game.executePlayerNavalMissions()

	mission := gs.Armies["fleet"].NavalMission
	if mission == nil || mission.Kind != army.NavalMissionPatrol || mission.TargetRegionID != "sea" {
		t.Fatalf("geçersiz abluka devriyeye dönüşmeli, got %#v", mission)
	}
}

func TestBuildWarSummaryIncludesCoalitionMetrics(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"attacker":      {ID: "attacker", NameTR: "Saldıran", Gold: 120, Grain: 80},
			"attacker_ally": {ID: "attacker_ally", NameTR: "Saldıran Müttefiki", Gold: 60, Grain: 40},
			"player":        {ID: "player", NameTR: "Oyuncu", Gold: 200, Grain: 100},
			"player_ally":   {ID: "player_ally", NameTR: "Oyuncu Müttefiki", Gold: 90, Grain: 50},
		},
		Regions: map[world.RegionID]*world.Region{
			"attacker_home":      {ID: "attacker_home", OwnerID: "attacker"},
			"attacker_ally_home": {ID: "attacker_ally_home", OwnerID: "attacker_ally"},
			"player_home":        {ID: "player_home", OwnerID: "player"},
			"player_ally_home":   {ID: "player_ally_home", OwnerID: "player_ally"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"attacker_army":      {ID: "attacker_army", OwnerID: "attacker", Units: []army.Unit{{}, {}}},
			"attacker_fleet":     {ID: "attacker_fleet", OwnerID: "attacker", IsNaval: true, Units: []army.Unit{{}}},
			"attacker_ally_army": {ID: "attacker_ally_army", OwnerID: "attacker_ally", Units: []army.Unit{{}, {}, {}}},
			"player_army":        {ID: "player_army", OwnerID: "player", Units: []army.Unit{{}, {}}},
			"player_ally_army":   {ID: "player_ally_army", OwnerID: "player_ally", Units: []army.Unit{{}, {}, {}}},
		},
	}
	game := &Game{gs: gs}
	report := game.buildWarSummaryFor("attacker", "player", diplomacy.WarDeclarationResult{
		PlayerCalls: []diplomacy.WarCallOutcome{{FactionID: "attacker_ally", NameTR: "Saldıran Müttefiki", Joined: true}},
		EnemyCalls:  []diplomacy.WarCallOutcome{{FactionID: "player_ally", NameTR: "Oyuncu Müttefiki", Joined: true}},
	})

	if got, want := len(report.Attacker.Participants), 2; got != want {
		t.Fatalf("saldıran katılımcı sayısı: got=%d want=%d", got, want)
	}
	if got, want := report.Attacker.TotalLandUnits, 5; got != want {
		t.Fatalf("saldıran kara birimleri: got=%d want=%d", got, want)
	}
	if got, want := report.Attacker.TotalNavalUnits, 1; got != want {
		t.Fatalf("saldıran deniz birimleri: got=%d want=%d", got, want)
	}
	if got, want := report.Attacker.TotalRegions, 2; got != want {
		t.Fatalf("saldıran bölge toplamı: got=%d want=%d", got, want)
	}
	if got, want := report.Attacker.TotalGold, 180; got != want {
		t.Fatalf("saldıran altın toplamı: got=%d want=%d", got, want)
	}
	if got, want := report.Defender.TotalLandUnits, 5; got != want {
		t.Fatalf("savunan kara birimleri: got=%d want=%d", got, want)
	}
	if report.Attacker.Participants[0].FactionID == "" || report.Defender.Participants[0].FactionID == "" {
		t.Fatal("bayrak çizimi için faction ID snapshot'a taşınmadı")
	}
}

func TestAITurnWarDeclarationShowsSummaryDuringQuickTurn(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"attacker": {ID: "attacker", NameTR: "Saldıran"},
			"player":   {ID: "player", NameTR: "Oyuncu"},
		},
	}
	renderer := render.New(gs)
	renderer.CurrentSettings.FastAITurns = true
	game := &Game{gs: gs, renderer: renderer, aiTurn: &aiTurnState{}}
	game.handleAITurnStep(ai.TurnStep{
		FactionID:     "attacker",
		Kind:          ai.TurnStepDiplomacy,
		TargetFaction: "player",
		WarDeclaration: &diplomacy.WarDeclarationResult{
			Result: diplomacy.Result{Applied: true, Message: "Savaş başladı."},
		},
	})

	if !renderer.WarSummaryVisible() {
		t.Fatal("hızlı turda oyuncuya savaş ilanı özeti açılmadı")
	}
}

func TestQueueConquestDecisionRejectsOwnRegionEvenWithRestorableSuccessor(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":    {ID: "player"},
			"successor": {ID: "successor", IsEliminated: true},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:                 "home",
				OwnerID:            "player",
				SuccessorFactionID: "successor",
			},
		},
	}

	// Renderer yalnızca gerçek kuyruklama akışını çalıştırabilmek için
	// gereklidir; kararın açılıp açılmayacağı state koşuluyla belirlenir.
	g := &Game{gs: gs, renderer: render.New(gs)}
	if g.queueConquestDecision("player", gs.Regions["home"], false) {
		t.Fatal("kendi bölgesini kurtaran ordu için fetih kararı kuyruğa alındı")
	}
	if len(g.pendingConquestDecisions) != 0 {
		t.Fatalf("kendi bölgesi için bekleyen fetih kararı: %#v", g.pendingConquestDecisions)
	}
}

func TestEconomyTickConsumesAIStyleNavalSupplyCargo(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Grain: 100},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 2},
			"soldier":   {ID: "soldier", Category: army.CategoryInfantry, GrainUpkeep: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", OwnerID: "ai", Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army": {
				ID: "army", OwnerID: "ai", RegionID: "coast",
				Units: []army.Unit{{TypeID: "soldier", CurrentHP: army.MaxUnitHP}, {TypeID: "soldier", CurrentHP: army.MaxUnitHP}},
			},
			"fleet": {
				ID: "fleet", OwnerID: "ai", RegionID: "sea", IsNaval: true,
				Units:        []army.Unit{{TypeID: "transport", CurrentHP: army.MaxUnitHP}},
				SupplyCargo:  economy.ResourceCost{Grain: 20},
				NavalMission: &army.NavalMission{Kind: army.NavalMissionSupplyArmy, TargetArmyID: "army"},
			},
		},
	}

	applyRegionalLogisticsPressure(gs)

	if got := gs.Armies["fleet"].SupplyCargo.Grain; got != 4 {
		t.Fatalf("yerel kapasite sonrası ikmal kargosu yanlış tüketildi: %d", got)
	}
	status := gs.RegionLogistics["coast"]
	if status.NavalSupplyGrainSpent != 16 || status.Overload != 0 {
		t.Fatalf("deniz ikmali lojistik durumuna yansımadı: %+v", status)
	}
	applyRegionalLogisticsPressure(gs)
	if gs.Armies["fleet"].NavalMission != nil {
		t.Fatal("kargo bitince ikmal görevi temizlenmedi")
	}
}
