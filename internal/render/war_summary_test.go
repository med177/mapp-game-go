package render

import "testing"

func TestShowBattleReportQueuesWhileWarSummaryVisible(t *testing.T) {
	r := &Renderer{
		warSummary: warSummaryState{
			show: true,
			data: WarSummaryReport{Title: "Savaş Özeti"},
		},
	}

	r.ShowBattleReport(BattleReport{Scene: BattleSceneLand, Title: "Muharebe"})

	if r.battleReport.show {
		t.Fatal("war summary açıkken battle report doğrudan görünmemeli")
	}
	if !r.queuedBattleReport.show {
		t.Fatal("battle report kuyruğa alınmalıydı")
	}
}

func TestHideWarSummaryShowsQueuedBattleReport(t *testing.T) {
	r := &Renderer{
		warSummary: warSummaryState{
			show: true,
			data: WarSummaryReport{Title: "Savaş Özeti"},
		},
		queuedBattleReport: battleReportState{
			show: true,
			data: BattleReport{Scene: BattleSceneLand, Title: "Muharebe"},
		},
	}

	r.HideWarSummary()

	if r.warSummary.show {
		t.Fatal("war summary kapanmalıydı")
	}
	if !r.battleReport.show {
		t.Fatal("kuyruktaki battle report görünür olmalıydı")
	}
	if r.queuedBattleReport.show {
		t.Fatal("queued battle report boşaltılmalıydı")
	}
}
