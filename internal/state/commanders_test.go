package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestDismissCommanderRemovesAssignedCommanderFromState(t *testing.T) {
	commander := &army.Commander{ID: "commander_vassal_1", OwnerID: "player", Name: "Commander"}
	currentArmy := &army.Army{ID: "army_1", OwnerID: "player", Commander: commander}
	gs := &GameState{
		Armies:     map[army.ArmyID]*army.Army{currentArmy.ID: currentArmy},
		Commanders: map[string]*army.Commander{commander.ID: commander},
	}

	if !gs.DismissCommander(commander.ID) {
		t.Fatal("DismissCommander() rejected an existing commander")
	}
	if currentArmy.Commander != nil {
		t.Fatal("dismissed commander remained assigned to the army")
	}
	if commander.AssignedArmyID != "" {
		t.Fatalf("dismissed commander kept army assignment %q", commander.AssignedArmyID)
	}
	if _, exists := gs.Commanders[commander.ID]; exists {
		t.Fatal("dismissed commander remained in the canonical commander pool")
	}
}

func TestPayCommanderRansomTransfersGoldAndReleases(t *testing.T) {
	commander := &army.Commander{ID: "cmd_ransom", OwnerID: "captive_owner", CaptiveAtSettlementID: "holding_main"}
	gs := &GameState{
		CommanderRansomGold: 250,
		Commanders:          map[string]*army.Commander{commander.ID: commander},
		Factions: map[faction.FactionID]*faction.Faction{
			"captive_owner": {ID: "captive_owner", Gold: 500},
			"captor":        {ID: "captor", Gold: 100},
		},
	}
	if !gs.PayCommanderRansom(commander.ID, "captor") {
		t.Fatal("fidye ödemesi uygulanamadı")
	}
	if commander.CaptiveAtSettlementID != "" {
		t.Fatal("komutan fidye sonrası tutsak kaldı")
	}
	if got := gs.Factions["captive_owner"].Gold; got != 250 {
		t.Fatalf("ödeyen altını = %d, want 250", got)
	}
	if got := gs.Factions["captor"].Gold; got != 350 {
		t.Fatalf("tutsak sahibi altını = %d, want 350", got)
	}
}

func TestCommanderCaptiveIsReleasedWhenOwnerCapturesHoldingRegion(t *testing.T) {
	commander := &army.Commander{ID: "cmd_captive", OwnerID: "lannister", Name: "Tutsak Komutan"}
	gs := &GameState{
		FiredEventIDs: map[string]bool{"flag:commander_captured": true},
		Commanders:    map[string]*army.Commander{commander.ID: commander},
		Regions: map[world.RegionID]*world.Region{
			"holding": {
				ID:          "holding",
				OwnerID:     "arryn",
				Settlements: []world.Settlement{{ID: "holding_main", IsCenter: true}},
			},
		},
	}
	if !gs.CaptureCommanderAtSettlement(commander.ID, "holding_main", []string{"commander_captured"}) {
		t.Fatal("komutan tutsak olarak kaydedilemedi")
	}
	if got := gs.CommanderUnavailableReason(commander); got != "Tutsak" {
		t.Fatalf("tutsak komutan nedeni = %q, want %q", got, "Tutsak")
	}
	commander.CaptivityKeepPenaltyApplied = true
	gs.Regions["holding"].OwnerID = "lannister"
	if released := gs.ReleaseCaptivesInOwnedSettlements(); len(released) != 1 || released[0] != commander.ID {
		t.Fatalf("tutsak otomatik serbest bırakılmadı: %#v", released)
	}
	if commander.CaptiveAtSettlementID != "" || commander.CaptivityKeepPenaltyApplied || gs.FiredEventIDs["flag:commander_captured"] {
		t.Fatal("tutsaklık state'i temizlenmedi")
	}
	if got := gs.CommanderUnavailableReason(commander); got != "" {
		t.Fatalf("serbest komutan hâlâ pasif: %q", got)
	}
}

func TestCommanderUnavailableFlagUnassignsAndBlocksAssignment(t *testing.T) {
	commander := &army.Commander{
		ID:      "cmd_tyrion",
		OwnerID: "lannister",
		Name:    "Tyrion Lannister",
		UnavailableStatuses: []army.CommanderUnavailableStatus{
			{Flag: "tyrion_captured", ReasonTR: "Tutsak"},
			{Flag: "tyrion_executed", ReasonTR: "İnfaz edildi", HideFromRoster: true, Permanent: true},
		},
	}
	assignedArmy := &army.Army{ID: "army_assigned", OwnerID: "lannister", Commander: commander}
	targetArmy := &army.Army{ID: "army_target", OwnerID: "lannister"}
	commander.AssignedArmyID = assignedArmy.ID
	gs := &GameState{
		Armies: map[army.ArmyID]*army.Army{
			assignedArmy.ID: assignedArmy,
			targetArmy.ID:   targetArmy,
		},
		Commanders:         map[string]*army.Commander{commander.ID: commander},
		CommanderTemplates: map[string][]*army.Commander{"lannister": {commander}},
		FiredEventIDs:      map[string]bool{"flag:tyrion_captured": true},
	}

	gs.SyncCommanderLinks()
	if got := gs.CommanderUnavailableReason(commander); got != "Tutsak" {
		t.Fatalf("pasif komutan nedeni = %q, want %q", got, "Tutsak")
	}
	if assignedArmy.Commander != nil || commander.AssignedArmyID != "" {
		t.Fatal("pasif komutan mevcut ordudan ayrılmadı")
	}
	if len(gs.AvailableCommanders("lannister")) != 0 {
		t.Fatal("pasif komutan atanabilirler listesinde kaldı")
	}
	if gs.AssignCommanderToArmy(commander.ID, targetArmy.ID) {
		t.Fatal("pasif komutan yeni orduya atanabildi")
	}
	if len(gs.CommanderRoster("lannister")) != 1 {
		t.Fatal("pasif komutan roster'dan kaldırıldı")
	}

	delete(gs.FiredEventIDs, "flag:tyrion_captured")
	if !gs.AssignCommanderToArmy(commander.ID, targetArmy.ID) {
		t.Fatal("flag temizlendikten sonra komutan atanamadı")
	}
	if gs.CommanderHiddenFromRoster(commander) {
		t.Fatal("tutsak komutan roster'dan gizlendi")
	}
	gs.FiredEventIDs["flag:tyrion_executed"] = true
	gs.RemovePermanentlyUnavailableCommanders()
	gs.SyncCommanderLinks()
	if _, exists := gs.Commanders[commander.ID]; exists || !gs.DismissedCommanderIDs[commander.ID] || len(gs.CommanderRoster("lannister")) != 0 {
		t.Fatal("infaz edilen komutan roster'dan çıkarılmadı")
	}
}

func TestCommanderInjuryExpiresByTurn(t *testing.T) {
	commander := &army.Commander{ID: "cmd_injured", OwnerID: "player", Name: "Yaralı Komutan"}
	currentArmy := &army.Army{ID: "army_current", OwnerID: "player", Commander: commander}
	targetArmy := &army.Army{ID: "army_target", OwnerID: "player"}
	commander.AssignedArmyID = currentArmy.ID
	gs := &GameState{
		Turn:                         10,
		CommanderInjuryTurnsOnDefeat: 2,
		Armies: map[army.ArmyID]*army.Army{
			currentArmy.ID: currentArmy,
			targetArmy.ID:  targetArmy,
		},
		Commanders: map[string]*army.Commander{commander.ID: commander},
	}

	if !gs.InjureCommander(commander.ID, gs.CommanderInjuryDuration(commander), "Yaralı") {
		t.Fatal("komutan yaralanma durumu yazılamadı")
	}
	if currentArmy.Commander != nil || commander.AssignedArmyID != "" {
		t.Fatal("yaralanan komutan ordudan ayrılmadı")
	}
	if got := gs.CommanderUnavailableReason(commander); got != "Yaralı" {
		t.Fatalf("yaralı komutan nedeni = %q, want %q", got, "Yaralı")
	}
	if gs.AssignCommanderToArmy(commander.ID, targetArmy.ID) {
		t.Fatal("yaralı komutan atanabildi")
	}

	gs.Turn = 12
	if got := gs.CommanderUnavailableReason(commander); got != "" {
		t.Fatalf("iyileşen komutan nedeni = %q, want boş", got)
	}
	if !gs.AssignCommanderToArmy(commander.ID, targetArmy.ID) {
		t.Fatal("iyileşen komutan atanamadı")
	}
}

func TestDismissCommanderSuppressesScenarioTemplateReappearance(t *testing.T) {
	commander := &army.Commander{ID: "commander_template", OwnerID: "player", Name: "Tarihsel Komutan"}
	gs := &GameState{
		Year:                    1300,
		PlayerFactionID:         "player",
		Commanders:              map[string]*army.Commander{commander.ID: commander},
		CommanderTemplates:      map[string][]*army.Commander{"player": {commander}},
		CommanderArrivalNotices: map[string]bool{commander.ID: true},
	}

	if !gs.DismissCommander(commander.ID) {
		t.Fatal("DismissCommander() rejected a template commander")
	}
	gs.SyncCommanderAvailability()
	if _, exists := gs.Commanders[commander.ID]; exists {
		t.Fatal("dismissed template commander reappeared during availability sync")
	}
	if !gs.DismissedCommanderIDs[commander.ID] {
		t.Fatal("dismissed template commander was not persisted as dismissed")
	}
}

func TestUpdateCommanderProfileChangesGeneratedCommanderOnlyProfile(t *testing.T) {
	commander := &army.Commander{
		ID:            "commander_vassal_2",
		OwnerID:       "vassal",
		Name:          "Commander",
		PortraitAsset: army.DefaultPortraitAsset,
		Experience:    300,
	}
	currentArmy := &army.Army{ID: "army_vassal", OwnerID: "vassal", Commander: commander}
	gs := &GameState{
		Armies:     map[army.ArmyID]*army.Army{currentArmy.ID: currentArmy},
		Commanders: map[string]*army.Commander{commander.ID: commander},
	}
	gs.TransferArmyOwnership(currentArmy, "player")
	if commander.OwnerID != "player" {
		t.Fatalf("transferred commander owner = %q", commander.OwnerID)
	}

	if !gs.IsGeneratedCommander(commander.ID) {
		t.Fatal("runtime commander was not identified as generated")
	}
	if !gs.UpdateCommanderProfile(commander.ID, "Köse Mihal", "commander_ottoman_kose.png") {
		t.Fatal("UpdateCommanderProfile() rejected a valid profile")
	}
	if commander.Name != "Köse Mihal" || commander.PortraitAsset != "commander_ottoman_kose.png" {
		t.Fatalf("profile was not updated: %#v", commander)
	}
	if commander.Experience != 300 {
		t.Fatal("profile update changed commander career progression")
	}
}

func TestAssignCommanderToArmyRejectsNonMilitaryFleets(t *testing.T) {
	tests := []struct {
		name      string
		units     []army.Unit
		canAssign bool
	}{
		{
			name:  "yalnız nakliye",
			units: []army.Unit{{TypeID: "transport", CurrentHP: army.MaxUnitHP}},
		},
		{
			name:  "yalnız tüccar",
			units: []army.Unit{{TypeID: "merchant_ship", CurrentHP: army.MaxUnitHP}},
		},
		{
			name: "savaş ve nakliye",
			units: []army.Unit{
				{TypeID: "warship", CurrentHP: army.MaxUnitHP},
				{TypeID: "transport", CurrentHP: army.MaxUnitHP},
			},
			canAssign: true,
		},
		{
			name: "savaş ve tüccar",
			units: []army.Unit{
				{TypeID: "warship", CurrentHP: army.MaxUnitHP},
				{TypeID: "merchant_ship", CurrentHP: army.MaxUnitHP},
			},
			canAssign: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const fleetID army.ArmyID = "fleet_1"
			commander := army.NewCommander("commander_1", "Komutan")
			commander.OwnerID = "player"
			fleet := &army.Army{
				ID:      fleetID,
				OwnerID: "player",
				IsNaval: true,
				Units:   tt.units,
			}
			gs := &GameState{
				Armies:     map[army.ArmyID]*army.Army{fleetID: fleet},
				Commanders: map[string]*army.Commander{commander.ID: commander},
				UnitTypes: map[string]*army.UnitType{
					"transport":     {ID: "transport", Category: army.CategoryNavalTrans},
					"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
					"warship":       {ID: "warship", Category: army.CategoryNavalWar},
				},
			}

			if got := gs.CanAssignCommanderToArmy(fleetID); got != tt.canAssign {
				t.Fatalf("CanAssignCommanderToArmy() = %v, want %v", got, tt.canAssign)
			}
			if got := gs.AssignCommanderToArmy(commander.ID, fleetID); got != tt.canAssign {
				t.Fatalf("AssignCommanderToArmy() = %v, want %v", got, tt.canAssign)
			}
			if tt.canAssign && fleet.Commander != commander {
				t.Fatal("savaş gemisi içeren filoya komutan atanmadı")
			}
			if !tt.canAssign && (fleet.Commander != nil || commander.AssignedArmyID != "") {
				t.Fatal("saf nakliye/tüccar filosuna komutan ataması state'e yazıldı")
			}
		})
	}
}

func TestReleaseInvalidFleetCommandersRemovesOnlyCivilianFleetCommanders(t *testing.T) {
	unitTypes := map[string]*army.UnitType{
		"transport":     {ID: "transport", Category: army.CategoryNavalTrans},
		"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
		"warship":       {ID: "warship", Category: army.CategoryNavalWar},
	}
	newCommander := func(id string) *army.Commander {
		return &army.Commander{ID: id, OwnerID: "player", Name: id}
	}
	transportCommander := newCommander("commander_transport")
	merchantCommander := newCommander("commander_merchant")
	warshipCommander := newCommander("commander_warship")
	transportFleet := &army.Army{
		ID:        "fleet_transport",
		OwnerID:   "player",
		IsNaval:   true,
		Units:     []army.Unit{{TypeID: "transport", CurrentHP: army.MaxUnitHP}},
		Commander: transportCommander,
	}
	merchantFleet := &army.Army{
		ID:        "fleet_merchant",
		OwnerID:   "player",
		IsNaval:   true,
		Units:     []army.Unit{{TypeID: "merchant_ship", CurrentHP: army.MaxUnitHP}},
		Commander: merchantCommander,
	}
	warshipFleet := &army.Army{
		ID:      "fleet_warship",
		OwnerID: "player",
		IsNaval: true,
		Units: []army.Unit{
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "transport", CurrentHP: army.MaxUnitHP},
		},
		Commander: warshipCommander,
	}
	transportCommander.AssignedArmyID = transportFleet.ID
	merchantCommander.AssignedArmyID = merchantFleet.ID
	warshipCommander.AssignedArmyID = warshipFleet.ID
	gs := &GameState{
		Armies: map[army.ArmyID]*army.Army{
			transportFleet.ID: transportFleet,
			merchantFleet.ID:  merchantFleet,
			warshipFleet.ID:   warshipFleet,
		},
		Commanders: map[string]*army.Commander{
			transportCommander.ID: transportCommander,
			merchantCommander.ID:  merchantCommander,
			warshipCommander.ID:   warshipCommander,
		},
		UnitTypes: unitTypes,
	}

	if got, want := gs.ReleaseInvalidFleetCommanders(), 2; got != want {
		t.Fatalf("ReleaseInvalidFleetCommanders() = %d, want %d", got, want)
	}
	if transportFleet.Commander != nil || merchantFleet.Commander != nil {
		t.Fatal("sivil filolardaki eski komutan bağlantısı temizlenmedi")
	}
	if transportCommander.AssignedArmyID != "" || merchantCommander.AssignedArmyID != "" {
		t.Fatal("serbest bırakılan komutan eski filo bağlantısını korudu")
	}
	if warshipFleet.Commander != warshipCommander || warshipCommander.AssignedArmyID != warshipFleet.ID {
		t.Fatal("savaş gemisi içeren filonun komutanı yanlışlıkla kaldırıldı")
	}
}
