package render

import (
	"image/color"
	"sort"
	"strconv"
	"strings"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	aiDiagnosticPanelW       = float32(980)
	aiDiagnosticPanelH       = float32(590)
	aiDiagnosticVisibleLines = 17
)

func aiDiagnosticPanelRect() gameui.Rect {
	return gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, float64(aiDiagnosticPanelW), float64(aiDiagnosticPanelH), gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
}

func (r *Renderer) toggleAIDiagnostic() {
	if r == nil || r.gs == nil || !r.gs.DevelopmentMode {
		return
	}
	if r.showAIDiagnostic {
		r.showAIDiagnostic = false
		r.aiDiagnosticSnapshot = nil
		return
	}
	r.aiDiagnosticFaction = r.firstDiagnosticFaction()
	r.aiDiagnosticScroll = 0
	r.refreshAIDiagnostic()
	r.showAIDiagnostic = r.aiDiagnosticSnapshot != nil
}

func (r *Renderer) firstDiagnosticFaction() faction.FactionID {
	if r == nil || r.gs == nil {
		return ""
	}
	ids := r.diagnosticFactionIDs()
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

func (r *Renderer) diagnosticFactionIDs() []faction.FactionID {
	if r == nil || r.gs == nil {
		return nil
	}
	ids := make([]faction.FactionID, 0, len(r.gs.Factions))
	for fid, candidate := range r.gs.Factions {
		if candidate == nil || candidate.IsEliminated || fid == r.gs.PlayerFactionID {
			continue
		}
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (r *Renderer) refreshAIDiagnostic() {
	if r == nil || r.gs == nil || r.aiDiagnosticFaction == "" {
		r.aiDiagnosticSnapshot = nil
		return
	}
	r.aiDiagnosticSnapshot = ai.BuildAIDiagnosticSnapshot(r.gs, r.aiDiagnosticFaction)
}

func (r *Renderer) cycleAIDiagnosticFaction(delta int) {
	ids := r.diagnosticFactionIDs()
	if len(ids) == 0 {
		r.aiDiagnosticFaction = ""
		r.aiDiagnosticSnapshot = nil
		return
	}
	index := sort.Search(len(ids), func(index int) bool { return ids[index] >= r.aiDiagnosticFaction })
	if index >= len(ids) || ids[index] != r.aiDiagnosticFaction {
		index = 0
	}
	index = (index + delta) % len(ids)
	if index < 0 {
		index += len(ids)
	}
	r.aiDiagnosticFaction = ids[index]
	r.aiDiagnosticScroll = 0
	r.refreshAIDiagnostic()
}

func (r *Renderer) handleAIDiagnosticInput() InputAction {
	if r == nil || !r.showAIDiagnostic {
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyF3) {
		r.toggleAIDiagnostic()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyTab) {
		r.cycleAIDiagnosticFaction(1)
		return InputAction{}
	}
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		r.aiDiagnosticScroll = clampAIDiagnosticScroll(r.aiDiagnosticScroll-int(wheelY), r.aiDiagnosticLineCount())
	}
	return InputAction{}
}

func clampAIDiagnosticScroll(scroll, lineCount int) int {
	maxScroll := lineCount - aiDiagnosticVisibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll < 0 {
		return 0
	}
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func (r *Renderer) aiDiagnosticLineCount() int {
	if r == nil || r.aiDiagnosticSnapshot == nil {
		return 0
	}
	return 10 + len(r.aiDiagnosticSnapshot.Fronts) + len(r.aiDiagnosticSnapshot.BlockReasons) + r.aiDiagnosticHistoryLineCount()
}

func (r *Renderer) aiDiagnosticHistoryLineCount() int {
	if r == nil || r.gs == nil || r.aiDiagnosticFaction == "" {
		return 0
	}
	count := 0
	for _, entry := range r.gs.AIDiagnosticHistory {
		if entry.FactionID == r.aiDiagnosticFaction {
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return count + 1
}

func (r *Renderer) drawAIDiagnostic(screen *ebiten.Image) {
	if r == nil || r.aiDiagnosticSnapshot == nil {
		return
	}
	modal := aiDiagnosticPanelRect()
	drawRoundedRect(screen, float32(modal.X), float32(modal.Y), float32(modal.W), float32(modal.H), 10, colorRGBA(13, 17, 24, 246))
	drawPanelBorder(screen, float32(modal.X), float32(modal.Y), float32(modal.W), float32(modal.H))
	vector.FillRect(screen, float32(modal.X), float32(modal.Y), float32(modal.W), 3, colorRGBA(205, 168, 72, 255), false)

	snapshot := r.aiDiagnosticSnapshot
	factionName := string(snapshot.FactionID)
	if f := r.gs.Factions[snapshot.FactionID]; f != nil && f.NameTR != "" {
		factionName = f.NameTR
	}
	DrawText(screen, "AI TEŞHİSİ — "+factionName, modal.X+22, modal.Y+20, FaceMed, ColorGold)
	DrawText(screen, "F3 / ESC: kapat   TAB: devlet değiştir   Mouse tekeri: kaydır", modal.X+22, modal.Y+46, FaceSmall, ColorGray)

	lines := r.aiDiagnosticLines()
	r.aiDiagnosticScroll = clampAIDiagnosticScroll(r.aiDiagnosticScroll, len(lines))
	lineY := modal.Y + 78
	for index := r.aiDiagnosticScroll; index < len(lines) && index < r.aiDiagnosticScroll+aiDiagnosticVisibleLines; index++ {
		line := lines[index]
		face := FaceSmall
		textColor := ColorWhite
		if strings.HasPrefix(line, "CEPHE") || strings.HasPrefix(line, "ENGEL") {
			face = FaceTiny
			textColor = ColorGray
		}
		DrawText(screen, line, modal.X+24, lineY, face, textColor)
		lineY += 25
	}
	DrawText(screen, "Geliştirme modu teşhis görünümü", modal.X+22, modal.Y+modal.H-20, FaceTiny, ColorGray)
}

func (r *Renderer) aiDiagnosticLines() []string {
	snapshot := r.aiDiagnosticSnapshot
	if snapshot == nil {
		return nil
	}
	plan := string(snapshot.PlanKind)
	if plan == "" {
		plan = "yok"
	}
	lines := []string{
		"Tur: " + strconv.Itoa(snapshot.Turn) + "   Plan: " + plan + "   Hedef devlet: " + diagnosticFactionName(r.gs, snapshot.PlanTargetFactionID),
		"Hedef bölgeler: " + strings.Join(regionIDsToStrings(snapshot.PlanTargetRegionIDs), ", "),
		"Yedek: %" + strconv.Itoa(snapshot.ReservePercent) + "   hedef güç " + strconv.Itoa(snapshot.ReserveTargetPower) + "   ayrılan " + strconv.Itoa(snapshot.ReserveAssignedPower),
		"Kritik tehdit: " + diagnosticBool(snapshot.CriticalThreat),
		"Ordu rolleri: " + diagnosticRoleText(snapshot.ArmyRoleCounts),
		"DONANMA: " + diagnosticNavalText(r.gs, snapshot),
		"CEPHELER",
	}
	history := r.aiDiagnosticHistoryLines()
	if len(history) > 0 {
		lines = append(lines, history...)
	}
	for _, front := range snapshot.Fronts {
		lines = append(lines, "CEPHE "+diagnosticFactionName(r.gs, front.EnemyFactionID)+" → "+diagnosticRegionName(r.gs, front.TargetRegionID)+" | dost "+strconv.Itoa(front.FriendlyPower)+" / düşman "+strconv.Itoa(front.EnemyPower)+" | tehdit "+strconv.Itoa(front.ThreatScore))
	}
	if len(snapshot.Fronts) == 0 {
		lines = append(lines, "CEPHE yok")
	}
	lines = append(lines, "ENGELLER")
	if len(snapshot.BlockReasons) == 0 {
		lines = append(lines, "ENGEL yok — karar akışı açık")
	} else {
		for _, reason := range snapshot.BlockReasons {
			lines = append(lines, "ENGEL: "+reason)
		}
	}
	return lines
}

func diagnosticNavalText(gs *state.GameState, snapshot *ai.AIDiagnosticSnapshot) string {
	if snapshot == nil || snapshot.NavalFleetCount == 0 {
		return "filo yok"
	}
	kind := snapshot.NavalMissionKind
	if kind == "" {
		kind = "bekleme"
	}
	target := "yok"
	if snapshot.NavalTargetRegionID != "" {
		target = diagnosticRegionName(gs, snapshot.NavalTargetRegionID)
	}
	return "filo " + strconv.Itoa(snapshot.NavalFleetCount) + " | liman " + strconv.Itoa(snapshot.NavalDockedFleetCount) + " | görev " + kind + " | hedef " + target
}

func (r *Renderer) aiDiagnosticHistoryLines() []string {
	if r == nil || r.gs == nil || r.aiDiagnosticFaction == "" {
		return nil
	}
	entries := make([]state.AIDiagnosticHistoryEntry, 0, 5)
	for _, entry := range r.gs.AIDiagnosticHistory {
		if entry.FactionID == r.aiDiagnosticFaction {
			entries = append(entries, entry)
		}
	}
	if len(entries) == 0 {
		return nil
	}
	lines := []string{"İLK 5 AI TURU — kayıt " + strconv.Itoa(len(entries)) + "/5"}
	for _, entry := range entries {
		entryPlan := string(entry.PlanKind)
		if entryPlan == "" {
			entryPlan = "yok"
		}
		blockCount := len(entry.BlockReasons)
		lines = append(lines, "TUR "+strconv.Itoa(entry.Turn)+" | plan "+entryPlan+" | hedef "+diagnosticRegionName(r.gs, entry.TargetRegionID)+" | cephe "+strconv.Itoa(entry.FrontCount)+" savaş "+strconv.Itoa(entry.ActiveWarCount)+" engel "+strconv.Itoa(blockCount))
	}
	return lines
}

func diagnosticFactionName(gs *state.GameState, fid faction.FactionID) string {
	if fid == "" {
		return "yok"
	}
	if gs != nil && gs.Factions[fid] != nil && gs.Factions[fid].NameTR != "" {
		return gs.Factions[fid].NameTR
	}
	return string(fid)
}

func diagnosticRegionName(gs *state.GameState, rid world.RegionID) string {
	if rid == "" {
		return "yok"
	}
	if gs != nil && gs.Regions[rid] != nil && gs.Regions[rid].NameTR != "" {
		return gs.Regions[rid].NameTR
	}
	return string(rid)
}

func regionIDsToStrings(ids []world.RegionID) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, string(id))
	}
	return result
}

func diagnosticBool(value bool) string {
	if value {
		return "evet"
	}
	return "hayır"
}

func diagnosticRoleText(roles map[ai.AIArmyRole]int) string {
	if len(roles) == 0 {
		return "yok"
	}
	ordered := []ai.AIArmyRole{ai.AIArmyRoleAssault, ai.AIArmyRoleSiege, ai.AIArmyRoleDefense, ai.AIArmyRoleReserve, ai.AIArmyRoleRelief, ai.AIArmyRoleTransport, ai.AIArmyRoleEscort}
	parts := make([]string, 0, len(roles))
	for _, role := range ordered {
		if count := roles[role]; count > 0 {
			parts = append(parts, string(role)+"="+strconv.Itoa(count))
		}
	}
	return strings.Join(parts, " ")
}

func colorRGBA(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}
