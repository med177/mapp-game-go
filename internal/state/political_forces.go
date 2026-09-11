package state

import (
	"sort"
	"strconv"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

const DefaultSuccessorRevivalMilitia = 5

// MarkFactionSubjugation, bir faction'ın başka bir faction tarafından
// elenmesi veya vassal yapılması bilgisini veri tabanlı event çözümüne taşır.
func (s *GameState) MarkFactionSubjugation(actor, target faction.FactionID) {
	if s == nil || actor == "" || target == "" || actor == target {
		return
	}
	s.LastSubjugationActorID = actor
	s.LastSubjugatedFactionID = target
}

// ConsumeFactionSubjugation son üstünlük bilgisini tek event taramasında
// tüketir; böylece aynı vassallık veya fetih birden fazla event başlatamaz.
func (s *GameState) ConsumeFactionSubjugation() (faction.FactionID, faction.FactionID) {
	if s == nil {
		return "", ""
	}
	actor, target := s.LastSubjugationActorID, s.LastSubjugatedFactionID
	s.LastSubjugationActorID = ""
	s.LastSubjugatedFactionID = ""
	return actor, target
}

// ReviveSuccessorAtRegion event veya savaş sonrası ardıl devlet akışının ortak
// state işlemini yürütür. Bölge sahibi değişir, faction etkinleşir ve temel
// bir kuruluş ordusu oluşturulur.
func (s *GameState) ReviveSuccessorAtRegion(regionID world.RegionID, successorID faction.FactionID, militiaCount int) (army.ArmyID, bool) {
	if s == nil || s.Regions == nil || s.Factions == nil || regionID == "" || successorID == "" {
		return "", false
	}
	region := s.Regions[regionID]
	successor := s.Factions[successorID]
	if region == nil || region.IsSea || successor == nil || !successor.IsEliminated || len(s.LandRegionsOwnedBy(successorID)) != 0 {
		return "", false
	}
	if militiaCount <= 0 {
		militiaCount = DefaultSuccessorRevivalMilitia
	}
	successor.IsEliminated = false
	successor.IsVirtual = false
	successor.PendingCapitalSettlementID = ""
	successor.PendingCapitalTurns = 0
	region.OwnerID = string(successorID)
	if s.Armies == nil {
		s.Armies = make(map[army.ArmyID]*army.Army)
	}
	s.NextArmySeq++
	armyID := army.ArmyID("army_" + string(successorID) + "_" + strconv.Itoa(s.NextArmySeq))
	for s.Armies[armyID] != nil {
		s.NextArmySeq++
		armyID = army.ArmyID("army_" + string(successorID) + "_" + strconv.Itoa(s.NextArmySeq))
	}
	newArmy := &army.Army{
		ID:            armyID,
		OwnerID:       string(successorID),
		RegionID:      regionID,
		Units:         army.MakeUnits("militia", militiaCount),
		MaxMovePoints: army.DefaultArmyMovePoints,
		MovePoints:    army.DefaultArmyMovePoints,
	}
	if s.UnitTypes != nil {
		newArmy.MaxMovePoints = newArmy.BaseMovePoints(s.UnitTypes)
		newArmy.MovePoints = newArmy.MaxMovePoints
	}
	s.Armies[armyID] = newArmy
	s.NormalizeFactionCapitals()
	return armyID, true
}

// PoliticalForceTransferReport siyasi dönüşüm sırasında aktarılan kuvvetleri
// özetler. Filo ve kara ordusu aynı Army state'i kullandığı için rapor ikisini
// IsNaval ile ayırır.
type PoliticalForceTransferReport struct {
	TransferredArmies int
	TransferredFleets int
	ByFaction         map[faction.FactionID]int
}

// ActivatePoliticalMembers, dönüşüm event'iyle ortaya çıkan faction'ları
// çalışır hale getirir ve kaynak faction'ın stoklarını eşit paylara ayırır.
// Kaynak faction'ın bir payı kendisinde bırakılır; böylece iç savaş başlarken
// merkez devlet tamamen boşaltılmaz.
func (s *GameState) ActivatePoliticalMembers(sourceID faction.FactionID, memberIDs []faction.FactionID) int {
	if s == nil || s.Factions == nil || sourceID == "" || len(memberIDs) == 0 {
		return 0
	}
	source := s.Factions[sourceID]
	if source == nil {
		return 0
	}
	members := make([]*faction.Faction, 0, len(memberIDs))
	seen := map[faction.FactionID]struct{}{}
	for _, memberID := range memberIDs {
		if memberID == "" || memberID == sourceID {
			continue
		}
		if _, ok := seen[memberID]; ok {
			continue
		}
		member := s.Factions[memberID]
		if member == nil {
			continue
		}
		seen[memberID] = struct{}{}
		members = append(members, member)
	}
	if len(members) == 0 {
		return 0
	}
	parts := len(members) + 1
	for _, member := range members {
		member.IsEliminated = false
		member.IsVirtual = false
		member.Gold += source.Gold / parts
		member.Grain += source.Grain / parts
		member.Iron += source.Iron / parts
		member.Timber += source.Timber / parts
		member.Stone += source.Stone / parts
		member.Spice += source.Spice / parts
		member.Cloth += source.Cloth / parts
		member.Research.Completed = cloneResearchCompleted(source.Research.Completed)
	}
	source.Gold = source.Gold / parts
	source.Grain = source.Grain / parts
	source.Iron = source.Iron / parts
	source.Timber = source.Timber / parts
	source.Stone = source.Stone / parts
	source.Spice = source.Spice / parts
	source.Cloth = source.Cloth / parts
	return len(members)
}

func cloneResearchCompleted(completed map[string]bool) map[string]bool {
	if completed == nil {
		return nil
	}
	clone := make(map[string]bool, len(completed))
	for id, done := range completed {
		clone[id] = done
	}
	return clone
}

// MergeFactionForces bir veya daha fazla faction'ın kara ordusu ve filolarını
// resultFaction'a bağlar. Ordu nesneleri korunur; yalnız sahiplik değişir.
// Böylece hareket konumu, komutan, filo görevi ve taşınan kara birlikleri
// dönüşüm sırasında kaybolmaz.
func (s *GameState) MergeFactionForces(memberIDs []faction.FactionID, resultFaction faction.FactionID) PoliticalForceTransferReport {
	report := PoliticalForceTransferReport{ByFaction: map[faction.FactionID]int{}}
	if s == nil || resultFaction == "" || len(memberIDs) == 0 {
		return report
	}
	members := make(map[string]struct{}, len(memberIDs))
	for _, memberID := range memberIDs {
		if memberID != "" {
			members[string(memberID)] = struct{}{}
		}
	}
	for _, armyID := range s.politicalArmyIDs() {
		current := s.Armies[armyID]
		if current == nil {
			continue
		}
		if _, ok := members[current.OwnerID]; !ok {
			continue
		}
		current.OwnerID = string(resultFaction)
		normalizePoliticalCommanderOwners(current)
		if current.IsNaval {
			report.TransferredFleets++
		} else {
			report.TransferredArmies++
		}
		report.ByFaction[resultFaction]++
	}
	return report
}

// UnitePoliticalFactions tüm üyelerin bölgelerini, kuvvetlerini ve
// kaynaklarını sonuç faction'ında toplar. Üye faction'lar tarihsel dönüşümün
// sonucunda pasif/elenmiş duruma alınır; sonuç faction'ı korunur.
func (s *GameState) UnitePoliticalFactions(memberIDs []faction.FactionID, resultFaction faction.FactionID) PoliticalForceTransferReport {
	report := s.MergeFactionForces(memberIDs, resultFaction)
	if s == nil || s.Factions == nil || resultFaction == "" {
		return report
	}
	if result := s.Factions[resultFaction]; result != nil {
		result.IsEliminated = false
		result.IsVirtual = false
	}
	seen := make(map[faction.FactionID]struct{}, len(memberIDs))
	for _, memberID := range memberIDs {
		if memberID == "" || memberID == resultFaction {
			continue
		}
		seen[memberID] = struct{}{}
		member := s.Factions[memberID]
		result := s.Factions[resultFaction]
		if member == nil || result == nil {
			continue
		}
		result.Gold += member.Gold
		result.Grain += member.Grain
		result.Iron += member.Iron
		result.Timber += member.Timber
		result.Stone += member.Stone
		result.Spice += member.Spice
		result.Cloth += member.Cloth
		member.Gold = 0
		member.Grain = 0
		member.Iron = 0
		member.Timber = 0
		member.Stone = 0
		member.Spice = 0
		member.Cloth = 0
		member.IsEliminated = true
		member.IsVirtual = false
		member.OverlordID = ""
		member.CapitalSettlementID = ""
		member.PendingCapitalSettlementID = ""
		member.PendingCapitalTurns = 0
	}
	for _, region := range s.Regions {
		if region == nil || region.IsSea {
			continue
		}
		if _, ok := seen[faction.FactionID(region.OwnerID)]; ok {
			region.OwnerID = string(resultFaction)
		}
	}
	s.NormalizeEliminatedFactionRelations()
	s.NormalizeFactionCapitals()
	return report
}

// SplitFactionForces dağıtılacak faction'ların kuvvetlerini bölge sahipliği
// veya explicit regionTargets bilgisine göre hedef faction'lara aktarır.
// Bölge eşleşmesi olmayan açık deniz filoları ve konumsuz kuvvetler, hedefler
// arasında deterministik round-robin ile dağıtılır; böylece event çözümleyicisi
// ayrıca filo için özel bir branch yazmak zorunda kalmaz.
func (s *GameState) SplitFactionForces(
	sourceIDs []faction.FactionID,
	targetIDs []faction.FactionID,
	regionTargets map[world.RegionID]faction.FactionID,
) PoliticalForceTransferReport {
	report := PoliticalForceTransferReport{ByFaction: map[faction.FactionID]int{}}
	if s == nil || len(sourceIDs) == 0 || len(targetIDs) == 0 {
		return report
	}
	sources := make(map[string]struct{}, len(sourceIDs))
	for _, sourceID := range sourceIDs {
		if sourceID != "" {
			sources[string(sourceID)] = struct{}{}
		}
	}
	targets := make([]faction.FactionID, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		if targetID == "" {
			continue
		}
		if s.Factions != nil && s.Factions[targetID] == nil {
			continue
		}
		targets = append(targets, targetID)
	}
	if len(targets) == 0 {
		return report
	}
	fallback := 0
	for _, armyID := range s.politicalArmyIDs() {
		current := s.Armies[armyID]
		if current == nil {
			continue
		}
		if _, ok := sources[current.OwnerID]; !ok {
			continue
		}
		target, ok := politicalForceTargetForRegion(s, current, regionTargets)
		if !ok {
			target = targets[fallback%len(targets)]
			fallback++
		}
		current.OwnerID = string(target)
		normalizePoliticalCommanderOwners(current)
		if current.IsNaval {
			report.TransferredFleets++
		} else {
			report.TransferredArmies++
		}
		report.ByFaction[target]++
	}
	return report
}

// DissolvePoliticalFaction bir birleşik faction'ı üyelerine geri ayırır.
// RegionTargets açıkça verilmiş bölgelerde dağılımı belirler; eksik bölgeler
// birleşik faction üyelerinden biri ise aynı üyede bırakılır, değilse
// deterministik round-robin ile paylaştırılır.
func (s *GameState) DissolvePoliticalFaction(
	sourceID faction.FactionID,
	targetIDs []faction.FactionID,
	regionTargets map[world.RegionID]faction.FactionID,
) PoliticalForceTransferReport {
	report := PoliticalForceTransferReport{ByFaction: map[faction.FactionID]int{}}
	if s == nil || s.Factions == nil || sourceID == "" || len(targetIDs) == 0 {
		return report
	}
	source := s.Factions[sourceID]
	if source == nil {
		return report
	}
	targets := make([]faction.FactionID, 0, len(targetIDs))
	seen := make(map[faction.FactionID]struct{}, len(targetIDs))
	for _, targetID := range targetIDs {
		if targetID == "" || s.Factions[targetID] == nil {
			continue
		}
		if _, ok := seen[targetID]; ok {
			continue
		}
		seen[targetID] = struct{}{}
		targets = append(targets, targetID)
		s.Factions[targetID].IsEliminated = false
		s.Factions[targetID].IsVirtual = false
	}
	if len(targets) == 0 {
		return report
	}

	// Kaynak stoklarının tamamı hedeflere aktarılır; birleşik faction hedefler
	// arasında değilse kendi payını kaybetmez, doğrudan pasifleştirilir.
	gold, grain := source.Gold, source.Grain
	iron, timber := source.Iron, source.Timber
	stone, spice := source.Stone, source.Spice
	cloth := source.Cloth
	source.Gold, source.Grain, source.Iron = 0, 0, 0
	source.Timber, source.Stone, source.Spice, source.Cloth = 0, 0, 0, 0
	for i, targetID := range targets {
		s.Factions[targetID].Gold += splitPoliticalResource(gold, i, len(targets))
		s.Factions[targetID].Grain += splitPoliticalResource(grain, i, len(targets))
		s.Factions[targetID].Iron += splitPoliticalResource(iron, i, len(targets))
		s.Factions[targetID].Timber += splitPoliticalResource(timber, i, len(targets))
		s.Factions[targetID].Stone += splitPoliticalResource(stone, i, len(targets))
		s.Factions[targetID].Spice += splitPoliticalResource(spice, i, len(targets))
		s.Factions[targetID].Cloth += splitPoliticalResource(cloth, i, len(targets))
	}

	// Önce açık bölge eşleşmelerini uygula; tanımsız source bölgeleri için
	// source zaten hedefse toprağı onda tut, değilse üyeler arasında dağıt.
	fallback := 0
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(sourceID) {
			continue
		}
		targetID, ok := regionTargets[region.ID]
		if _, valid := seen[targetID]; !valid {
			if !ok {
				if _, sourceRemains := seen[sourceID]; sourceRemains {
					targetID = sourceID
				} else {
					targetID = targets[fallback%len(targets)]
					fallback++
				}
			} else {
				targetID = targets[fallback%len(targets)]
				fallback++
			}
		}
		if _, valid := seen[targetID]; valid {
			regionTargets[region.ID] = targetID
			region.OwnerID = string(targetID)
		}
	}

	report = s.SplitFactionForces([]faction.FactionID{sourceID}, targets, regionTargets)
	if _, remains := seen[sourceID]; !remains {
		source.IsEliminated = true
		source.IsVirtual = false
		source.OverlordID = ""
		source.CapitalSettlementID = ""
		source.PendingCapitalSettlementID = ""
		source.PendingCapitalTurns = 0
	}
	s.NormalizeEliminatedFactionRelations()
	s.NormalizeFactionCapitals()
	return report
}

func splitPoliticalResource(total, index, parts int) int {
	if parts <= 0 {
		return 0
	}
	share := total / parts
	if index < total%parts {
		share++
	}
	return share
}

func (s *GameState) politicalArmyIDs() []army.ArmyID {
	if len(s.ArmyOrder) > 0 {
		return append([]army.ArmyID(nil), s.ArmyOrder...)
	}
	ids := make([]army.ArmyID, 0, len(s.Armies))
	for id := range s.Armies {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func politicalForceTargetForRegion(s *GameState, current *army.Army, regionTargets map[world.RegionID]faction.FactionID) (faction.FactionID, bool) {
	if current == nil {
		return "", false
	}
	regionID := current.RegionID
	if current.IsNaval && current.DockedRegionID != "" {
		regionID = current.DockedRegionID
	}
	if target, ok := regionTargets[regionID]; ok && target != "" {
		if s.Factions == nil || s.Factions[target] != nil {
			return target, true
		}
	}
	return "", false
}

func normalizePoliticalCommanderOwners(current *army.Army) {
	if current == nil {
		return
	}
	if current.Commander != nil {
		current.Commander.OwnerID = current.OwnerID
		current.Commander.AssignedArmyID = current.ID
	}
	if current.EmbarkedCommander != nil {
		current.EmbarkedCommander.OwnerID = current.OwnerID
		current.EmbarkedCommander.AssignedArmyID = current.ID
	}
}
