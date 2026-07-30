package render

import "testing"

func TestHideBattleReportShowsQueuedChoiceDialog(t *testing.T) {
	r := &Renderer{
		battleReport: battleReportState{
			show: true,
			data: BattleReport{Scene: BattleSceneLand},
		},
	}

	r.QueueChoiceDialogAfterBattleReport(
		"Savaş Sonrası Düzen",
		"Karar metni",
		"İlhak Et",
		"Vassal Yap",
		InputAction{Kind: ActionAnnexDefeatedFaction},
		InputAction{Kind: ActionVassalizeDefeatedFaction},
	)
	r.HideBattleReport()

	if r.battleReport.show {
		t.Fatal("battle report kapanmalıydı")
	}
	if !r.confirmDialog.show {
		t.Fatal("kuyruklanan seçim diyaloğu görünür olmalıydı")
	}
	if r.confirmDialog.pendingAction.Kind != ActionAnnexDefeatedFaction {
		t.Fatalf("sol aksiyon annex olmalıydı, got=%s", r.confirmDialog.pendingAction.Kind)
	}
	if !r.confirmDialog.declineActs || r.confirmDialog.declineAction.Kind != ActionVassalizeDefeatedFaction {
		t.Fatalf("sağ aksiyon vassal olmalıydı, got=%+v", r.confirmDialog)
	}
	if r.queuedConfirmDialog.show {
		t.Fatal("kuyruk boşaltılmalıydı")
	}
}

func TestHideBattleReportShowsQueuedThreeChoiceSuccessorDialog(t *testing.T) {
	r := &Renderer{battleReport: battleReportState{show: true, data: BattleReport{Scene: BattleSceneLand}}}
	r.QueueThreeChoiceDialogAfterBattleReport(
		"Ardıl Devlet Kararı",
		"Bölgenin kaderi",
		"İlhak Et",
		"Serbest Bırak",
		"Vassal Yap",
		InputAction{Kind: ActionAnnexSuccessor},
		InputAction{Kind: ActionReleaseSuccessor},
		InputAction{Kind: ActionVassalizeSuccessor},
	)
	r.HideBattleReport()

	if !r.confirmDialog.show || !r.confirmDialog.declineActs {
		t.Fatal("üç seçenekli ardıl kararı görünür olmalıydı")
	}
	if r.confirmDialog.pendingAction.Kind != ActionAnnexSuccessor || r.confirmDialog.thirdAction.Kind != ActionReleaseSuccessor || r.confirmDialog.declineAction.Kind != ActionVassalizeSuccessor {
		t.Fatalf("ardıl karar aksiyonları yanlış: %+v", r.confirmDialog)
	}
}
