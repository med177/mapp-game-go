package state

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
)

func TestCommanderPoolAssignsAndReleasesUniqueCommander(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "player"},
			"a2": {ID: "a2", OwnerID: "player"},
		},
	}
	gs.InitializePlayerCommanders()
	available := gs.AvailableCommanders("player")
	if len(available) != InitialPlayerCommanderPool {
		t.Fatalf("ilk komutan havuzu yanlış: got=%d want=%d", len(available), InitialPlayerCommanderPool)
	}
	commanderID := available[0].ID
	if !gs.AssignCommanderToArmy(commanderID, "a1") {
		t.Fatal("komutan ilk orduya atanamadı")
	}
	if len(gs.AvailableCommanders("player")) != InitialPlayerCommanderPool-1 {
		t.Fatal("atanan komutan havuzdan düşmedi")
	}
	if gs.AssignCommanderToArmy(commanderID, "a2") {
		t.Fatal("aynı komutan ikinci orduya atanabildi")
	}
	if !gs.UnassignCommanderFromArmy("a1") || len(gs.AvailableCommanders("player")) != InitialPlayerCommanderPool {
		t.Fatal("komutan ayrılınca havuza dönmedi")
	}
}

func TestRecruitPlayerCommanderUsesEnteredNameAndDeterministicProgression(t *testing.T) {
	gs := &GameState{
		PlayerFactionID:  "player",
		NextCommanderSeq: 8,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 900, Grain: 180},
		},
	}
	commander, ok := gs.RecruitPlayerCommander("  Aylin Hatun  ")
	if !ok || commander == nil {
		t.Fatal("oyuncu komutanı oluşturulamadı")
	}
	if commander.ID != "commander_player_9" || commander.OwnerID != "player" || commander.Name != "Aylin Hatun" {
		t.Fatalf("oluşturulan komutanın kimliği yanlış: %+v", commander)
	}
	if commander.PortraitAsset != army.DefaultPortraitAsset || commander.Experience < 0 || commander.Experience >= army.CommanderLevel5XP {
		t.Fatalf("rastgele başlangıç profili sınır dışında: %+v", commander)
	}
	expected := recruitedCommanderExperience("player", 9)
	if commander.Experience != expected {
		t.Fatalf("başlangıç XP deterministik değil: got=%d want=%d", commander.Experience, expected)
	}
	if commander.Level >= 3 && !commander.HasTrait(army.CommanderTraitTactician) {
		t.Fatalf("XP'nin açtığı taktik uzmanlığı eksik: %+v", commander)
	}
	if commander.Level >= 4 && !commander.HasTrait(army.CommanderTraitDefender) {
		t.Fatalf("XP'nin açtığı savunma uzmanlığı eksik: %+v", commander)
	}
	if got := gs.Factions["player"]; got.Gold != 400 || got.Grain != 80 {
		t.Fatalf("komutan maliyeti uygulanmadı: %+v", got)
	}
}

func TestRecruitPlayerCommanderRejectsBlankAndTooLongNames(t *testing.T) {
	gs := &GameState{PlayerFactionID: "player", Factions: map[faction.FactionID]*faction.Faction{"player": {ID: "player", Gold: 500, Grain: 100}}}
	if commander, ok := gs.RecruitPlayerCommander(" \t "); ok || commander != nil {
		t.Fatal("boş komutan adı kabul edilmemeliydi")
	}
	if commander, ok := gs.RecruitPlayerCommander(strings.Repeat("a", PlayerCommanderMaxNameRunes+1)); ok || commander != nil {
		t.Fatal("fazla uzun komutan adı kabul edilmemeliydi")
	}
}

func TestRecruitPlayerCommanderRequiresGoldAndGrain(t *testing.T) {
	gs := &GameState{PlayerFactionID: "player", Factions: map[faction.FactionID]*faction.Faction{"player": {ID: "player", Gold: 500, Grain: 99}}}
	if commander, ok := gs.RecruitPlayerCommander("Yetersiz Tahıl"); ok || commander != nil {
		t.Fatal("yetersiz tahılla komutan oluşturulmamalı")
	}
	if got := gs.Factions["player"]; got.Gold != 500 || got.Grain != 99 {
		t.Fatalf("başarısız üretim kaynak tüketmemeli: %+v", got)
	}
}

func TestNormalizeEmptyArmiesRemovesOnlyEmptyStacks(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Gold: 1000, Grain: 200},
		},
		Armies: map[army.ArmyID]*army.Army{
			"empty":    {ID: "empty", OwnerID: "ai"},
			"land":     {ID: "land", OwnerID: "ai", Units: []army.Unit{{TypeID: "inf"}}},
			"embarked": {ID: "embarked", OwnerID: "ai", IsNaval: true, EmbarkedUnits: []army.Unit{{TypeID: "inf"}}},
		},
	}
	if got := gs.NormalizeEmptyArmies(); got != 1 {
		t.Fatalf("yalnız boş stack temizlenmeliydi: removed=%d", got)
	}
	if gs.Armies["empty"] != nil || gs.Armies["land"] == nil || gs.Armies["embarked"] == nil {
		t.Fatalf("geçerli ordular da temizlendi: %+v", gs.Armies)
	}
}

func TestCommanderPoolUsesScenarioTemplatesBeforeFallback(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "ottoman",
		CommanderTemplates: map[string][]*army.Commander{
			"ottoman": {
				{
					ID:            "commander_ottoman_test",
					OwnerID:       "ottoman",
					Name:          "Senaryo Komutanı",
					PortraitAsset: "commanders/test.png",
					Traits:        []army.CommanderTrait{army.CommanderTraitDefender},
				},
			},
		},
	}

	gs.InitializePlayerCommanders()
	commander := gs.Commanders["commander_ottoman_test"]
	if commander == nil || commander.Name != "Senaryo Komutanı" || commander.PortraitAsset != "commanders/test.png" || !commander.HasTrait(army.CommanderTraitDefender) {
		t.Fatalf("senaryo komutanı havuza alınmadı: %+v", commander)
	}
}

func TestCommanderAvailabilityAddsArrivalsAndRetiresExpiredAssignments(t *testing.T) {
	commander := &army.Commander{
		ID:        "cmd_new",
		OwnerID:   "player",
		Name:      "Yeni Komutan",
		StartYear: 1305,
		EndYear:   1310,
	}
	gs := &GameState{
		Year:               1305,
		PlayerFactionID:    "player",
		Commanders:         map[string]*army.Commander{},
		CommanderTemplates: map[string][]*army.Commander{"player": {commander}},
		Armies:             map[army.ArmyID]*army.Army{"a1": {ID: "a1", OwnerID: "player"}},
	}
	arrivals := gs.SyncCommanderAvailability()
	if len(arrivals) != 1 || gs.Commanders["cmd_new"] == nil {
		t.Fatalf("aktif komutan havuza/bildirime alınmadı: arrivals=%d commanders=%+v", len(arrivals), gs.Commanders)
	}
	if !gs.AssignCommanderToArmy("cmd_new", "a1") {
		t.Fatal("aktif komutan atanamadı")
	}

	gs.Year = 1310
	gs.RetireExpiredCommanders()
	if gs.Commanders["cmd_new"] != nil || gs.Armies["a1"].Commander != nil {
		t.Fatalf("süresi dolan komutan emekli olmadı: commanders=%+v army=%+v", gs.Commanders, gs.Armies["a1"])
	}
	if gs.AssignCommanderToArmy("cmd_new", "a1") {
		t.Fatal("süresi dolan komutan yeniden atanabildi")
	}
}

func TestCommanderActiveYearHasExclusiveEndBoundary(t *testing.T) {
	commander := &army.Commander{StartYear: 1300, EndYear: 1310}
	if !commander.ActiveInYear(1300) || !commander.ActiveInYear(1309) {
		t.Fatal("komutan başlangıç ve bitişten önceki yılda aktif olmalı")
	}
	if commander.ActiveInYear(1299) || commander.ActiveInYear(1310) {
		t.Fatal("komutan start_year öncesi veya end_year başlangıcında aktif olmamalı")
	}
}

func TestSyncCommanderLinksUsesPoolAsCanonicalPointer(t *testing.T) {
	armyCommander := army.NewCommander("c1", "Komutan")
	pooledCommander := army.NewCommander("c1", "Havuz Komutanı")
	pooledCommander.OwnerID = "player"
	gs := &GameState{
		Commanders: map[string]*army.Commander{"c1": pooledCommander},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "player", Commander: armyCommander},
		},
	}
	gs.SyncCommanderLinks()
	if gs.Armies["a1"].Commander != pooledCommander || pooledCommander.AssignedArmyID != "a1" {
		t.Fatal("havuz komutanı orduya canonical pointer olarak bağlanmadı")
	}
}

func TestEnsureFactionCommandersAssignsAIFieldArmies(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Gold: 1000, Grain: 200},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "ai", RegionID: "r1", Units: []army.Unit{{TypeID: "inf"}}},
			"a2": {ID: "a2", OwnerID: "ai", RegionID: "r2", Units: []army.Unit{{TypeID: "inf"}}},
			"g1": {ID: "g1", OwnerID: "ai", IsGarrison: true, RegionID: "r3"},
			"p1": {ID: "p1", OwnerID: "player", RegionID: "r4"},
		},
	}
	gs.EnsureFactionCommanders("ai")

	if gs.Armies["a1"].Commander == nil || gs.Armies["a2"].Commander == nil {
		t.Fatal("AI saha ordularına komutan atanmadı")
	}
	if gs.Armies["g1"].Commander != nil || gs.Armies["p1"].Commander != nil {
		t.Fatal("AI komutanı kapsam dışı orduya atandı")
	}
	if len(gs.Commanders) != 2 {
		t.Fatalf("AI komutan havuzu yanlış boyutta: got=%d", len(gs.Commanders))
	}
	if gs.Armies["a1"].Commander == gs.Armies["a2"].Commander {
		t.Fatal("aynı AI komutanı iki orduya atanmış")
	}
	for _, commander := range gs.Commanders {
		if commander.OwnerID != "ai" || commander.AssignedArmyID == "" {
			t.Fatalf("AI komutan bağlantısı eksik: %+v", commander)
		}
		if commander.PortraitAsset != army.DefaultPortraitAsset {
			t.Fatalf("AI fallback komutan portresi default olmaliydi: %+v", commander)
		}
	}
	if got := gs.Factions["ai"]; got.Gold != 0 || got.Grain != 0 {
		t.Fatalf("AI komutan maliyeti uygulanmadı: %+v", got)
	}
}

func TestEnsureFactionCommandersLeavesAIArmiesUncommandedWithoutResources(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{"ai": {ID: "ai", Gold: 499, Grain: 100}},
		Armies:   map[army.ArmyID]*army.Army{"a1": {ID: "a1", OwnerID: "ai", Units: []army.Unit{{TypeID: "inf"}}}},
	}
	gs.EnsureFactionCommanders("ai")
	if gs.Armies["a1"].Commander != nil || len(gs.Commanders) != 0 {
		t.Fatalf("yetersiz kaynaklı AI komutan üretmemeli: armies=%+v commanders=%+v", gs.Armies, gs.Commanders)
	}
}

func TestEnsureFactionCommandersSkipsNonMilitaryFleetsAndReusesAvailableCommander(t *testing.T) {
	commander := army.NewCommander("cmd_idle", "Boşta Komutan")
	commander.OwnerID = "ai"
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Gold: 500, Grain: 100},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Category: army.CategoryInfantry},
			"merchant":  {ID: "merchant", Category: army.CategoryNavalTrade},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
		},
		Commanders: map[string]*army.Commander{commander.ID: commander},
		Armies: map[army.ArmyID]*army.Army{
			"land": {
				ID: "land", OwnerID: "ai", Units: []army.Unit{{TypeID: "inf"}},
			},
			"merchant_fleet": {
				ID: "merchant_fleet", OwnerID: "ai", IsNaval: true,
				Units: []army.Unit{{TypeID: "merchant"}}, Commander: commander,
			},
			"transport_fleet": {
				ID: "transport_fleet", OwnerID: "ai", IsNaval: true,
				Units: []army.Unit{{TypeID: "transport"}},
			},
		},
	}

	gs.EnsureFactionCommanders("ai")

	if gs.Armies["merchant_fleet"].Commander != nil || gs.Armies["transport_fleet"].Commander != nil {
		t.Fatal("ticaret ve nakliye filolarına AI komutanı atanmamalı")
	}
	if gs.Armies["land"].Commander != commander {
		t.Fatal("boşta kalan mevcut komutan askerî orduya atanmalı")
	}
	if got := gs.Factions["ai"]; got.Gold != 500 || got.Grain != 100 {
		t.Fatalf("boşta komutan varken yeni üretim maliyeti uygulanmamalı: %+v", got)
	}
}

func TestEnsureFactionCommandersAssignsWarshipButNotMerchantFleet(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Gold: 500, Grain: 100},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship":  {ID: "warship", Category: army.CategoryNavalWar},
			"merchant": {ID: "merchant", Category: army.CategoryNavalTrade},
		},
		Armies: map[army.ArmyID]*army.Army{
			"war_fleet": {
				ID: "war_fleet", OwnerID: "ai", IsNaval: true,
				Units: []army.Unit{{TypeID: "warship"}},
			},
			"merchant_fleet": {
				ID: "merchant_fleet", OwnerID: "ai", IsNaval: true,
				Units: []army.Unit{{TypeID: "merchant"}},
			},
		},
	}

	gs.EnsureFactionCommanders("ai")

	if gs.Armies["war_fleet"].Commander == nil || gs.Armies["merchant_fleet"].Commander != nil {
		t.Fatal("savaş filosu komutan almalı, ticaret filosu almamalı")
	}
	if got := gs.Factions["ai"]; got.Gold != 0 || got.Grain != 0 {
		t.Fatalf("savaş filosu komutanının maliyeti uygulanmalı: %+v", got)
	}
}

func TestSyncCommanderLinksRepairsGeneratedCommanderPortraitFromSave(t *testing.T) {
	gs := &GameState{
		Commanders: map[string]*army.Commander{
			"commander_ai_7": {
				ID:             "commander_ai_7",
				OwnerID:        "ai",
				Name:           "Komutan 7",
				AssignedArmyID: "a1",
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "ai"},
		},
	}

	gs.SyncCommanderLinks()

	if got := gs.Commanders["commander_ai_7"].PortraitAsset; got != army.DefaultPortraitAsset {
		t.Fatalf("save'den gelen fallback komutan default portre almaliydi: got=%q", got)
	}
}

func TestCommanderTransportAndRemovalLifecycle(t *testing.T) {
	commander := army.NewCommander("c1", "Nakliye Komutanı")
	commander.OwnerID = "player"
	gs := &GameState{
		Armies: map[army.ArmyID]*army.Army{
			"land":  {ID: "land", OwnerID: "player", Commander: commander},
			"fleet": {ID: "fleet", OwnerID: "player", IsNaval: true},
			"land2": {ID: "land2", OwnerID: "player"},
		},
	}

	gs.MoveCommanderIntoFleet("land", "fleet")
	if gs.Armies["land"].Commander != nil || gs.Armies["fleet"].EmbarkedCommander != commander || commander.AssignedArmyID != "fleet" {
		t.Fatal("komutan filoya taşınmadı")
	}
	gs.RemoveArmy("land")
	gs.MoveEmbarkedCommanderToArmy("fleet", "land2")
	if gs.Armies["land2"].Commander != commander || gs.Armies["fleet"].EmbarkedCommander != nil || commander.AssignedArmyID != "land2" {
		t.Fatal("taşınan komutan yeni kara ordusuna dönmedi")
	}

	gs.RemoveArmy("land2")
	if commander.AssignedArmyID != "" {
		t.Fatal("silinen ordunun komutanı havuza bırakılmadı")
	}
}

func TestTransferArmyOwnershipUpdatesCommanderOwner(t *testing.T) {
	commander := army.NewCommander("c1", "Devralınan Komutan")
	commander.OwnerID = "old"
	gs := &GameState{Armies: map[army.ArmyID]*army.Army{
		"fleet": {ID: "fleet", OwnerID: "old", IsNaval: true, Commander: commander},
	}}

	gs.TransferArmyOwnership(gs.Armies["fleet"], "new")
	if gs.Armies["fleet"].OwnerID != "new" || commander.OwnerID != "new" {
		t.Fatalf("ordu/komutan sahipliği birlikte aktarılmadı: army=%q commander=%q", gs.Armies["fleet"].OwnerID, commander.OwnerID)
	}
}
