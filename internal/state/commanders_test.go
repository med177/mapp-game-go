package state

import (
	"testing"

	"mapp-game-go/internal/army"
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
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "ai", RegionID: "r1"},
			"a2": {ID: "a2", OwnerID: "ai", RegionID: "r2"},
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
