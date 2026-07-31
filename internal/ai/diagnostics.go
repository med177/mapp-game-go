package ai

import (
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// AIDiagnosticFront debug/telemetry tüketicilerinin cephe kararını aynı
// runtime context'ten okuyabilmesi için kullanılan küçük snapshot'tır.
type AIDiagnosticFront struct {
	EnemyFactionID faction.FactionID
	TargetRegionID world.RegionID
	FriendlyPower  int
	EnemyPower     int
	ThreatScore    int
	AtWar          bool
	CapitalThreat  bool
	CriticalThreat bool
}

// AIDiagnosticSnapshot AI turundaki plan, cephe, rol ve bloklanma nedenlerini
// tek yerde toplar. Save'e yazılmaz; debug paneli veya tempo raporu doğrudan
// bunu tüketebilir.
type AIDiagnosticSnapshot struct {
	FactionID             faction.FactionID
	Turn                  int
	PlanKind              state.AIObjectiveKind
	PlanTargetFactionID   faction.FactionID
	PlanTargetRegionIDs   []world.RegionID
	ReservePercent        int
	ReserveTargetPower    int
	ReserveAssignedPower  int
	CriticalThreat        bool
	Fronts                []AIDiagnosticFront
	ArmyRoleCounts        map[AIArmyRole]int
	NavalMissionKind      string
	NavalTargetRegionID   world.RegionID
	NavalFleetCount       int
	NavalDockedFleetCount int
	BlockReasons          []string
}

// BuildAIDiagnosticSnapshot mevcut AI karar akışını tekrar kullanır; ayrı bir
// teşhis algoritması üretmediği için debug gösterimi ile gerçek runtime kararı
// arasında drift oluşmaz.
func BuildAIDiagnosticSnapshot(gs *state.GameState, fid faction.FactionID) *AIDiagnosticSnapshot {
	snapshot := &AIDiagnosticSnapshot{FactionID: fid, ArmyRoleCounts: make(map[AIArmyRole]int)}
	if gs == nil || fid == "" {
		snapshot.BlockReasons = []string{"geçersiz oyun veya fraksiyon"}
		return snapshot
	}
	ctx := prepareStrategicContext(gs, fid)
	snapshot.Turn = gs.Turn
	snapshot.ReservePercent = ctx.ReservePercent
	snapshot.ReserveTargetPower = ctx.ReserveTargetPower
	snapshot.ReserveAssignedPower = ctx.ReserveAssignedPower
	snapshot.CriticalThreat = ctx.CriticalThreat
	if plan := gs.AIPlans[fid]; plan != nil {
		snapshot.PlanKind = plan.Kind
		snapshot.PlanTargetFactionID = plan.TargetFactionID
		snapshot.PlanTargetRegionIDs = append([]world.RegionID(nil), plan.TargetRegionIDs...)
	} else {
		snapshot.BlockReasons = append(snapshot.BlockReasons, "stratejik plan yok")
	}
	for _, front := range ctx.Fronts {
		snapshot.Fronts = append(snapshot.Fronts, AIDiagnosticFront{
			EnemyFactionID: front.EnemyFactionID,
			TargetRegionID: front.TargetRegionID,
			FriendlyPower:  front.FriendlyPower,
			EnemyPower:     front.EnemyPower,
			ThreatScore:    front.ThreatScore,
			AtWar:          front.AtWar,
			CapitalThreat:  front.CapitalThreat,
			CriticalThreat: front.CriticalThreat,
		})
	}
	for _, assignment := range ctx.ArmyAssignments {
		snapshot.ArmyRoleCounts[assignment.Role]++
	}
	for _, fleet := range aiSortedArmies(gs) {
		if fleet == nil || fleet.OwnerID != string(fid) || !fleet.IsNaval {
			continue
		}
		snapshot.NavalFleetCount++
		if fleet.IsDocked() {
			snapshot.NavalDockedFleetCount++
		}
	}
	if ctx.navalMission != nil {
		snapshot.NavalMissionKind = string(ctx.navalMission.Kind)
		snapshot.NavalTargetRegionID = ctx.navalMission.TargetRegionID
		if ctx.navalMission.MissingCapacity > 0 {
			snapshot.BlockReasons = append(snapshot.BlockReasons, "deniz görevi için nakliye kapasitesi eksik")
		}
	} else if snapshot.NavalFleetCount > 0 {
		snapshot.NavalMissionKind = "patrol"
	}
	if snapshot.NavalDockedFleetCount > 0 {
		snapshot.BlockReasons = append(snapshot.BlockReasons, "donanmanın bir kısmı limanda; ilk hareket denize çıkış")
	}
	if ctx.CriticalThreat {
		snapshot.BlockReasons = append(snapshot.BlockReasons, "kritik tehdit")
	}
	if ctx.ReserveAssignedPower < ctx.ReserveTargetPower {
		snapshot.BlockReasons = append(snapshot.BlockReasons, "yetersiz yedek kuvvet")
	}
	if aiWarLogisticsPolicyActive(gs) && !aiWarLogisticsReady(gs, fid) {
		snapshot.BlockReasons = append(snapshot.BlockReasons, "tahıl/altın lojistik rezervi yetersiz")
	}
	for _, front := range ctx.Fronts {
		if front.AtWar && front.TargetRegionID == "" {
			snapshot.BlockReasons = append(snapshot.BlockReasons, "aktif savaşta geçerli hedef yok")
		}
	}
	return snapshot
}

// RecordAIDiagnosticRound, geliştirme save'i yüklendikten sonra başlayan bir
// tam AI fazının başlangıç snapshot'ını kaydeder. Kayıt AI kararları state'i
// değiştirmeden önce alındığı için beş tur arasındaki plan/hedef/cephe farkını
// aynı başlangıç noktalarından karşılaştırır.
func RecordAIDiagnosticRound(gs *state.GameState) bool {
	if gs == nil || !gs.DevelopmentMode || gs.AIDiagnosticCaptureTurnsRemain <= 0 {
		return false
	}

	ids := make([]faction.FactionID, 0, len(gs.Factions))
	for fid, candidate := range gs.Factions {
		if fid == gs.PlayerFactionID || candidate == nil || candidate.IsEliminated {
			continue
		}
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	for _, fid := range ids {
		snapshot := BuildAIDiagnosticSnapshot(gs, fid)
		entry := state.AIDiagnosticHistoryEntry{
			Turn:                 snapshot.Turn,
			FactionID:            snapshot.FactionID,
			PlanKind:             snapshot.PlanKind,
			PlanTargetFactionID:  snapshot.PlanTargetFactionID,
			FrontCount:           len(snapshot.Fronts),
			ReservePercent:       snapshot.ReservePercent,
			ReserveTargetPower:   snapshot.ReserveTargetPower,
			ReserveAssignedPower: snapshot.ReserveAssignedPower,
			BlockReasons:         append([]string(nil), snapshot.BlockReasons...),
		}
		for _, front := range snapshot.Fronts {
			if !front.AtWar {
				continue
			}
			entry.ActiveWarCount++
			if entry.TargetRegionID == "" {
				entry.TargetRegionID = front.TargetRegionID
			}
		}
		gs.AIDiagnosticHistory = append(gs.AIDiagnosticHistory, entry)
	}
	gs.AIDiagnosticCaptureTurnsRemain--
	return true
}
