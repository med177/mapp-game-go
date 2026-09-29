package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/game"
)

func main() {
	scenario := flag.String("scenario", "assets/scenarios/1300_ottoman_rise", "Senaryo klasörü")
	player := flag.String("player", "", "Oyuncu devleti; boşsa ilk oynanabilir devlet")
	turns := flag.Int("turns", 100, "Çalıştırılacak stratejik tur sayısı")
	seed := flag.Uint64("seed", 13100301, "Deterministik AI karar seed'i")
	difficulty := flag.Int("difficulty", 2, "Zorluk: 1 kolay, 2 normal, 3 zor")
	format := flag.String("format", "markdown", "Çıktı biçimi: markdown veya json")
	out := flag.String("out", "", "Çıktı dosyası; boşsa stdout")
	flag.Parse()

	report, err := game.RunHeadlessSimulation(game.HeadlessSimulationOptions{
		ScenarioPath:    *scenario,
		PlayerFactionID: faction.FactionID(*player),
		Turns:           *turns,
		Seed:            *seed,
		Difficulty:      *difficulty,
		Progress: func(completedTurns, totalTurns int) {
			fmt.Fprintf(os.Stderr, "\r%d/%d Tur", completedTurns, totalTurns)
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "simülasyon hatası:", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr)

	var data []byte
	switch strings.ToLower(*format) {
	case "json":
		data, err = json.MarshalIndent(report, "", "  ")
		data = append(data, '\n')
	case "markdown", "md":
		data = []byte(renderMarkdown(report))
	default:
		fmt.Fprintln(os.Stderr, "geçersiz format:", *format)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "rapor oluşturulamadı:", err)
		os.Exit(1)
	}

	if *out == "" {
		_, _ = os.Stdout.Write(data)
		return
	}
	if dir := filepath.Dir(*out); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "rapor klasörü oluşturulamadı:", err)
			os.Exit(1)
		}
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "rapor yazılamadı:", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "rapor yazıldı:", *out)
}

func renderMarkdown(report *game.HeadlessSimulationReport) string {
	names := game.HeadlessReportFactionNames(report)
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — %d Tur Simülasyon Raporu\n\n", report.ScenarioID, report.CompletedTurns)
	fmt.Fprintf(&b, "- Tarih: %d/%02d → %d/%02d\n", report.StartYear, report.StartMonth, report.EndYear, report.EndMonth)
	fmt.Fprintf(&b, "- Oyuncu devleti: %s\n", factionName(names, report.PlayerFactionID))
	fmt.Fprintf(&b, "- Seed: `%d`\n", report.Seed)
	fmt.Fprintf(&b, "- Zorluk: %d\n", report.Difficulty)
	fmt.Fprintf(&b, "- Karar politikası: %s\n\n", report.DecisionPolicy)
	if report.StopReason != "" {
		fmt.Fprintf(&b, "- Durdurma nedeni: `%s`\n\n", report.StopReason)
	}

	b.WriteString("## Nihai devlet durumu\n\n")
	b.WriteString("| Devlet | Bölge | Ordu | Kara birimi | Donanma birimi | Üretim emri | Bina emri | Bina | Teknoloji | Altın | Tahıl | Durum |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|\n")
	finalFactions := append([]game.HeadlessFactionSummary(nil), report.Factions...)
	sort.Slice(finalFactions, func(i, j int) bool {
		if finalFactions[i].FinalRegions != finalFactions[j].FinalRegions {
			return finalFactions[i].FinalRegions > finalFactions[j].FinalRegions
		}
		return finalFactions[i].FinalLandUnits > finalFactions[j].FinalLandUnits
	})
	for _, f := range finalFactions {
		status := "ayakta"
		if f.Eliminated {
			status = "ELENDİ"
		}
		fmt.Fprintf(&b, "| %s | %d (%+d) | %d | %d | %d | %d | %d | %d→%d | %+d | %d | %d | %s |\n",
			md(f.NameTR), f.FinalRegions, f.FinalRegions-f.InitialRegions, f.FinalArmies, f.FinalLandUnits,
			f.FinalNavalUnits, f.ProductionOrders, f.BuildingOrders, f.InitialBuildings, f.FinalBuildings,
			f.TechnologiesCompleted, f.FinalGold, f.FinalGrain, status)
	}
	b.WriteString("\nNot: “Üretim emri”, simülasyon boyunca AI'nin kuyruğa aldığı birim üretim emirlerini; mevcut birim sayısı ise kayıp ve takviyeler sonrası nihai aktif birlikleri gösterir.\n\n")

	b.WriteString("## Toprak değişimleri\n\n")
	for _, f := range finalFactions {
		if len(f.RegionsGained) == 0 && len(f.RegionsLost) == 0 {
			continue
		}
		fmt.Fprintf(&b, "- **%s**: +%s; -%s\n", md(f.NameTR), joinOrDash(f.RegionsGained), joinOrDash(f.RegionsLost))
	}
	b.WriteString("\n")

	b.WriteString("## Savaşlar\n\n")
	if len(report.Wars) == 0 {
		b.WriteString("Gözlenen savaş yok.\n\n")
	} else {
		b.WriteString("| Taraflar | Başlangıç | Bitiş | Durum | Bölge kazanımı | Kayıp |\n|---|---:|---:|---|---:|---:|\n")
		for _, w := range report.Wars {
			end := "devam ediyor"
			if w.EndedTurn > 0 {
				end = fmt.Sprintf("%d. tur", w.EndedTurn)
			}
			fmt.Fprintf(&b, "| %s — %s | %d | %s | %s | %d / %d | %d / %d |\n", factionName(names, w.FactionA), factionName(names, w.FactionB), w.StartedTurn, end, boolLabel(w.Active, "aktif", "bitti"), w.RegionsCapturedA, w.RegionsCapturedB, w.CasualtiesA, w.CasualtiesB)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Ticaret ve ittifak hareketleri\n\n")
	if len(report.TradeChanges) == 0 && len(report.AllianceChanges) == 0 {
		b.WriteString("Simülasyon boyunca yeni ticaret/ittifak değişimi gözlenmedi.\n\n")
	} else {
		for _, event := range report.TradeChanges {
			fmt.Fprintf(&b, "- %d. tur: **%s** — %s ↔ %s%s\n", event.Turn, event.Action, factionName(names, event.FactionA), factionName(names, event.FactionB), suffix(event.Good))
		}
		for _, event := range report.AllianceChanges {
			fmt.Fprintf(&b, "- %d. tur: **%s** — %s ↔ %s\n", event.Turn, event.Action, factionName(names, event.FactionA), factionName(names, event.FactionB))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Elenen devletler\n\n")
	if len(report.Eliminations) == 0 {
		fmt.Fprintf(&b, "%d tur içinde elenen devlet olmadı.\n\n", report.CompletedTurns)
	} else {
		for _, e := range report.Eliminations {
			fmt.Fprintf(&b, "- %d. tur: **%s**\n", e.Turn, md(e.NameTR))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Tur checkpoint'leri\n\n")
	for _, checkpoint := range report.Checkpoints {
		fmt.Fprintf(&b, "### %d. tur (%d/%02d)\n\n", checkpoint.Turn, checkpoint.Year, checkpoint.Month)
		limit := len(checkpoint.Rankings)
		if limit > 5 {
			limit = 5
		}
		for i := 0; i < limit; i++ {
			ranking := checkpoint.Rankings[i]
			fmt.Fprintf(&b, "%d. %s — %d bölge, %d ordu, %d kara birimi, güç %d\n", i+1, md(ranking.NameTR), ranking.Regions, ranking.Armies, ranking.LandUnits, ranking.MilitaryPower)
		}
		b.WriteString("\n")
	}

	if len(report.Events) > 0 {
		b.WriteString("## Tetiklenen olaylar\n\n")
		for _, event := range report.Events {
			fmt.Fprintf(&b, "- %d. tur: **%s** (`%s`)\n", event.Turn, md(event.NameTR), event.ID)
		}
	}
	return b.String()
}

func factionName(names map[faction.FactionID]string, id faction.FactionID) string {
	if name := names[id]; name != "" {
		return name
	}
	return string(id)
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "—"
	}
	return strings.Join(values, ", ")
}

func suffix(value string) string {
	if value == "" {
		return ""
	}
	return " (" + value + ")"
}

func boolLabel(value bool, yes, no string) string {
	if value {
		return yes
	}
	return no
}

func md(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "|", "\\|"), "\n", " ")
}
