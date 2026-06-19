package game

import (
	"testing"

	"mapp-game-go/internal/army"
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

func TestMoveArmyWithStanceBlocksFortifiedRegionWithoutSiegeUnit(t *testing.T) {
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

	g.moveArmyWithStance("atk", "dst", "")

	if gs.Armies["atk"].RegionID != "src" {
		t.Fatalf("ordu tahkimata siege birimi olmadan girmemeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.Regions["dst"].OwnerID != "p2" {
		t.Fatalf("tahkimli bölge savaşsız el değiştirmemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
	if gs.SiegeAt("dst") != nil {
		t.Fatal("siege birimi olmayan ordu kuşatma başlatamamalıydı")
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

func TestResolveSiegesCapturesBreachedFortifiedRegion(t *testing.T) {
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
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			StartedTurn:       4,
			TurnsElapsed:      2,
			FortLevel:         2,
			BreachProgress:    16,
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

func TestResolveSiegesDoesNotOpenBreachWithInsufficientSiegeTier(t *testing.T) {
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
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"dst": {
			RegionID:          "dst",
			AttackerArmyID:    "atk",
			AttackerFactionID: "p1",
			StartedTurn:       4,
			FortLevel:         3,
		},
	}
	g := &Game{gs: gs}

	g.resolveSieges()

	siege := gs.SiegeAt("dst")
	if siege == nil {
		t.Fatal("kuşatma kaydı korunmalıydı")
	}
	if siege.BreachProgress != 0 || siege.BreachLevel != 0 {
		t.Fatalf("yetersiz siege tier ile 3. seviye tahkimatta gedik açılmamalıydı, got progress=%d level=%d", siege.BreachProgress, siege.BreachLevel)
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
			Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "siege", CurrentHP: 100}},
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
