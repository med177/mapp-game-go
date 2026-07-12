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
	}
	minorThreshold, _ := siegeBreachThresholds(3)
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

func TestAssaultSiegeWithoutSiegeUnitIsBlocked(t *testing.T) {
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
		t.Fatalf("kuşatma birimi olmadan genel hücum kale fethi vermemeliydi, got=%s", gs.Regions["dst"].OwnerID)
	}
	if gs.Armies["atk"].RegionID != "src" {
		t.Fatalf("kuşatma birimi olmadan genel hücumda ordu hedefe girmemeliydi, got=%s", gs.Armies["atk"].RegionID)
	}
	if gs.SiegeAt("dst") == nil {
		t.Fatal("aktif kuşatma assault engellense de sürmeliydi")
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
