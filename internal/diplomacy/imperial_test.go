package diplomacy

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func imperialTestState() *state.GameState {
	return &state.GameState{
		Turn: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			"hre":   {ID: "hre", NameTR: "HRE", Religion: religion.Catholic, Gold: 100, Grain: 100},
			"milan": {ID: "milan", NameTR: "Milano", Religion: religion.Catholic, Gold: 100, Grain: 100},
			"enemy": {ID: "enemy", NameTR: "Düşman", Religion: religion.Catholic, Gold: 100, Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"hre_home":   {ID: "hre_home", OwnerID: "hre", Neighbors: []world.RegionID{"enemy_home"}},
			"milan_home": {ID: "milan_home", OwnerID: "milan", Neighbors: []world.RegionID{"enemy_home"}},
			"enemy_home": {ID: "enemy_home", OwnerID: "enemy", Neighbors: []world.RegionID{"hre_home", "milan_home"}},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		Relations: map[string]*faction.Relation{},
		Imperial: &state.ImperialState{
			EmpireID:  "hre",
			EmperorID: "hre",
			Authority: 90,
			Members: map[faction.FactionID]*state.ImperialMember{
				"milan": {FactionID: "milan", Status: state.ImperialMemberPrince, Loyalty: 100, Autonomy: 0, MilitaryCommitment: 100, ElectorWeight: 1},
			},
		},
	}
}

func TestImperialWarPreviewIncludesIndependentMembers(t *testing.T) {
	gs := imperialTestState()

	preview := BuildWarDeclarationPreview(gs, "hre", "enemy")

	var found bool
	for _, entry := range append(preview.Attacker.AutoParticipants, preview.Attacker.CallableAllies...) {
		if entry.FactionID == "milan" {
			found = true
			if !entry.ImperialMember {
				t.Fatal("üye satırı ImperialMember olarak işaretlenmeli")
			}
		}
	}
	if !found {
		t.Fatalf("Milano HRE savaş önizlemesinde görünmeli: %+v / %+v", preview.Attacker.AutoParticipants, preview.Attacker.CallableAllies)
	}
}

func TestImperialWarCallJoinsHighCommitmentMember(t *testing.T) {
	gs := imperialTestState()

	result := ExecuteWarDeclaration(gs, "hre", "enemy", nil)

	if !result.Applied {
		t.Fatalf("HRE savaşı uygulanmalıydı: %+v", result)
	}
	if rel := Relation(gs, "milan", "enemy"); rel == nil || rel.Stance != faction.StanceWar {
		t.Fatalf("yüksek bağlılıklı imparatorluk üyesi savaşa katılmalı: %+v", rel)
	}
	if gs.Imperial.LastWarCall == nil || gs.Imperial.LastWarCall.EnemyID != "enemy" {
		t.Fatalf("son imparatorluk çağrısı save state'e yazılmalı: %+v", gs.Imperial.LastWarCall)
	}
}

func TestImperialWarCallCanSendLimitedSupportWithoutEnteringWar(t *testing.T) {
	gs := imperialTestState()
	member := gs.Imperial.Members["milan"]
	member.Loyalty = 35
	member.Autonomy = 100
	member.MilitaryCommitment = 0
	delete(gs.Regions, "milan_home")
	gs.Regions["milan_home"] = &world.Region{ID: "milan_home", OwnerID: "milan"}
	gs.Armies["milan_army"] = &army.Army{ID: "milan_army", OwnerID: "milan", RegionID: "milan_home", Units: []army.Unit{{TypeID: "infantry"}}}

	result := ExecuteWarDeclaration(gs, "hre", "enemy", nil)

	var limited bool
	for _, outcome := range result.PlayerCalls {
		if outcome.FactionID == "milan" {
			limited = outcome.LimitedSupport
			if !limited {
				t.Fatalf("Milano sınırlı destek vermeli: %+v", outcome)
			}
			if outcome.SupportGold == 0 || outcome.SupportGrain == 0 {
				t.Fatalf("sınırlı destek miktarları raporlanmalı: %+v", outcome)
			}
		}
	}
	if limited {
		if rel := Relation(gs, "milan", "enemy"); rel != nil && rel.Stance == faction.StanceWar {
			t.Fatal("sınırlı destek veren üye savaşa girmemeli")
		}
		return
	}
	t.Fatal("Milano için sınırlı destek sonucu bulunamadı")
}

func TestImperialElectionUpdatesEmperorAndResetsAuthority(t *testing.T) {
	gs := imperialTestState()
	gs.Imperial.Authority = 40
	gs.Imperial.Members["milan"].ElectorWeight = 1
	gs.Relations[faction.RelationKey("milan", "hre")] = &faction.Relation{FactionA: "milan", FactionB: "hre", Score: -80, Stance: faction.StancePeace}
	result := HoldImperialElection(gs)

	if result.WinnerID == "" {
		t.Fatal("imparatorluk seçimi kazanan üretmeli")
	}
	if gs.Imperial.EmperorID != result.WinnerID {
		t.Fatalf("imparator seçimi state'e yazılmalı: state=%q result=%q", gs.Imperial.EmperorID, result.WinnerID)
	}
	if gs.Imperial.Authority != 48 {
		t.Fatalf("seçim otoriteyi 8 artırmalı: got=%d", gs.Imperial.Authority)
	}
}

func TestPlayerImperialPoliticsCreatesAndResolvesDietDecision(t *testing.T) {
	gs := imperialTestState()
	gs.PlayerFactionID = "hre"
	gs.Turn = 12
	gs.Imperial.NextDietTurn = 12

	report := AdvanceImperialPolitics(gs)
	if !report.Pending || gs.Imperial.PendingDecision == nil || gs.Imperial.PendingDecision.Kind != state.ImperialDecisionDiet {
		t.Fatalf("oyuncu Diyet kararı beklemeli: report=%+v imperial=%+v", report, gs.Imperial)
	}
	if ok, _ := ResolveImperialDiet(gs, 2); !ok {
		t.Fatal("oyuncu Diyet kararı çözülebilmeli")
	}
	if gs.Imperial.PendingDecision != nil {
		t.Fatalf("Diyet kararı sonrasında pending state temizlenmedi: %+v", gs.Imperial.PendingDecision)
	}
}

func TestPlayerImperialElectionUsesSelectedValidCandidate(t *testing.T) {
	gs := imperialTestState()
	gs.PlayerFactionID = "hre"
	gs.Turn = 97
	gs.Imperial.ElectionDueTurn = 97

	report := AdvanceImperialPolitics(gs)
	if !report.Pending || gs.Imperial.PendingDecision == nil || gs.Imperial.PendingDecision.Kind != state.ImperialDecisionElection {
		t.Fatalf("oyuncu seçim kararı beklemeli: report=%+v imperial=%+v", report, gs.Imperial)
	}
	if ok, _ := ResolveImperialElection(gs, "milan"); !ok {
		t.Fatal("geçerli aday seçilebilmeli")
	}
	if gs.Imperial.EmperorID != "milan" || gs.Imperial.PendingDecision != nil {
		t.Fatalf("seçim state'i yanlış çözüldü: emperor=%q pending=%+v", gs.Imperial.EmperorID, gs.Imperial.PendingDecision)
	}
}
