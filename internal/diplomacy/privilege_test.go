package diplomacy

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

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
	if rel == nil || rel.Score != -privilegeAttackRelationPenalty || rel.Stance != faction.StancePeace {
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
				FactionA: "east_rome", FactionB: "genoa", Score: 20, Stance: faction.StancePeace,
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
	if rel := Relation(gs, "east_rome", "genoa"); rel == nil || rel.Score != 20-PrivilegeRevocationRelationPenalty || rel.Stance != faction.StancePeace {
		t.Fatalf("imtiyaz kaldırma ilişki cezası yanlış: %#v", rel)
	}
	if IsWar(gs, "east_rome", "genoa") {
		t.Fatal("imtiyaz kaldırma kullanım sahibiyle savaş başlattı")
	}
}
