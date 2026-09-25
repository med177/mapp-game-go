package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

const (
	navalEscortDefenseBonusPerFleet = 0.15
	navalEscortDefenseBonusCap      = 0.30
)

// CanAssignNavalMission merkezi oyuncu filo görevi doğrulamasıdır. Renderer
// yalnız adayları gösterir; gerçek state değişikliği bu kapıdan geçer.
func (s *GameState) CanAssignNavalMission(fleetID army.ArmyID, mission army.NavalMission) (bool, string) {
	if s == nil || fleetID == "" {
		return false, "Geçersiz filo."
	}
	fleet := s.Armies[fleetID]
	if fleet == nil || !fleet.IsNaval {
		return false, "Seçilen ordu bir donanma değil."
	}
	if mission.Kind == "" {
		return false, "Filo görevi seçilmedi."
	}

	switch mission.Kind {
	case army.NavalMissionPatrol:
		if !fleetHasWarship(s, fleet) {
			return false, "Bu görev için savaş gemisi gerekir."
		}
		if !fleet.IsAtSea() || mission.TargetRegionID != fleet.RegionID {
			return false, "Devriye görevi yalnızca filonun bulunduğu açık denizde atanabilir."
		}
		if !s.validSeaMissionTarget(mission.TargetRegionID) {
			return false, "Hedef deniz bölgesi geçerli değil."
		}
	case army.NavalMissionBlockade:
		if !fleetHasWarship(s, fleet) {
			return false, "Bu görev için savaş gemisi gerekir."
		}
		if !fleet.IsAtSea() || mission.TargetRegionID != fleet.RegionID {
			return false, "Abluka görevi yalnızca filonun bulunduğu açık denizde atanabilir."
		}
		if !s.IsValidNavalBlockadeTarget(fleet, mission.TargetRegionID) {
			return false, "Abluka yalnızca düşman kıyısına komşu deniz bölgesine atanabilir."
		}
	case army.NavalMissionEscort:
		if !fleetHasWarship(s, fleet) {
			return false, "Escort görevi için savaş gemisi gerekir."
		}
		target := s.Armies[mission.TargetFleetID]
		if target == nil || target.ID == fleet.ID || target.OwnerID != fleet.OwnerID || !target.IsNaval || target.TransportCapacity(s.UnitTypes) <= 0 {
			return false, "Escort hedefi aynı devlete ait nakliye filosu olmalı."
		}
		if !fleet.IsAtSea() || !target.IsAtSea() || target.RegionID != fleet.RegionID {
			return false, "Escort hedefi, savaş filosuyla aynı açık deniz bölgesinde olmalı."
		}
	case army.NavalMissionTransport:
		if fleet.TransportCapacity(s.UnitTypes) <= 0 {
			return false, "Bu görev için nakliye kapasitesi gerekir."
		}
		if len(fleet.EmbarkedUnits) == 0 {
			return false, "Nakliye görevi için filoda taşınan kara ordusu yok."
		}
		target := s.Regions[mission.TargetRegionID]
		if target == nil || target.IsSea || !target.CanLandEnter() || !target.IsCoastal(s.Regions) {
			return false, "Nakliye hedefi kıyı kara bölgesi olmalı."
		}
	case army.NavalMissionSupplyArmy:
		if fleet.TransportCapacity(s.UnitTypes) <= 0 {
			return false, "İkmal görevi için nakliye gemisi gerekir."
		}
		if fleet.SupplyCargo.Grain <= 0 {
			return false, "Filo önce merkez limanında ikmal malları yüklemeli."
		}
		target := s.Armies[mission.TargetArmyID]
		if target == nil || target.ID == fleet.ID || target.OwnerID != fleet.OwnerID || target.IsNaval || len(target.Units) == 0 {
			return false, "İkmal hedefi aynı devlete ait, kara üzerinde bulunan bir ordu olmalı."
		}
		targetRegion := s.Regions[target.RegionID]
		if targetRegion == nil || targetRegion.IsSea || !targetRegion.IsCoastal(s.Regions) {
			return false, "İkmal edilecek ordu kıyı bölgesinde bulunmalı."
		}
		if !fleet.IsAtSea() || !regionsShareSeaNeighbor(s, targetRegion, fleet.RegionID) {
			return false, "Filo, ordunun kıyı bölgesine komşu denizde olmalı."
		}
	default:
		return false, "Bilinmeyen filo görevi."
	}
	return true, ""
}

func regionsShareSeaNeighbor(s *GameState, land *world.Region, seaID world.RegionID) bool {
	if s == nil || land == nil || seaID == "" {
		return false
	}
	for _, neighborID := range land.Neighbors {
		if neighborID == seaID && s.Regions[neighborID] != nil && s.Regions[neighborID].IsSea {
			return true
		}
	}
	return false
}

func supplyCargoTotal(cargo economy.ResourceCost) int {
	return cargo.Grain + cargo.Iron + cargo.Timber + cargo.Stone + cargo.Spice + cargo.Cloth
}

// CanLoadSupplyCargoAtCapital, yalnızca merkez limanında ve asker taşımayan
// filonun devlet stokundan ikmal yüklemesine izin verir.
func (s *GameState) CanLoadSupplyCargoAtCapital(fleetID army.ArmyID, cargo economy.ResourceCost) (bool, string) {
	if s == nil || fleetID == "" {
		return false, "Geçersiz filo."
	}
	fleet := s.Armies[fleetID]
	if fleet == nil || !fleet.IsNaval || fleet.OwnerID != string(s.PlayerFactionID) {
		return false, "Yalnız oyuncu filosu yüklenebilir."
	}
	capital, _, _, ok := s.FactionCapital(factionID(fleet.OwnerID))
	if !ok || capital == nil || !capital.HasPort() || fleet.DockedRegionID != capital.ID {
		return false, "İkmal yükü yalnızca devletin merkez limanında yüklenebilir."
	}
	if len(fleet.EmbarkedUnits) > 0 {
		return false, "Asker taşıyan filo ikmal yükü alamaz."
	}
	if fleet.TransportCapacity(s.UnitTypes) <= 0 {
		return false, "İkmal taşımak için nakliye gemisi gerekir."
	}
	if cargo.Gold != 0 || supplyCargoTotal(cargo) <= 0 {
		return false, "En az bir ikmal malı seçilmeli; altın ikmal yükü değildir."
	}
	if supplyCargoTotal(fleet.SupplyCargo)+supplyCargoTotal(cargo) > s.SupplyCargoCapacityForTurns(fleetID, 5) {
		return false, "Bu filo için belirlenen ikmal kargo kapasitesi aşılıyor."
	}
	if !cargo.CanAfford(s.Factions[factionID(fleet.OwnerID)]) {
		return false, "Devletin stoklarında bu yük için yeterli mal yok."
	}
	return true, ""
}

// LoadSupplyCargoAtCapital devlet stokunu azaltıp ikmal yükünü filoya koyar.
func (s *GameState) LoadSupplyCargoAtCapital(fleetID army.ArmyID, cargo economy.ResourceCost) (bool, string) {
	if ok, reason := s.CanLoadSupplyCargoAtCapital(fleetID, cargo); !ok {
		return false, reason
	}
	fleet := s.Armies[fleetID]
	faction := s.Factions[factionID(fleet.OwnerID)]
	cargo.Apply(faction)
	fleet.SupplyCargo = addSupplyCargo(fleet.SupplyCargo, cargo)
	return true, ""
}

// LoadSupplyCargoForTurns, mevcut kara ordularının tahıl ihtiyacını seçilen
// tur sayısı için merkez limanında yükler. İlk sürümde gerçek lojistik hesabı
// tahıl üzerinden yürüdüğü için otomatik hesaplanan yük tahıldır.
func (s *GameState) LoadSupplyCargoForTurns(fleetID army.ArmyID, turns int) (economy.ResourceCost, bool, string) {
	if s == nil || turns < 1 {
		return economy.ResourceCost{}, false, "İkmal süresi en az bir tur olmalı."
	}
	fleet := s.Armies[fleetID]
	if fleet == nil {
		return economy.ResourceCost{}, false, "Geçersiz filo."
	}
	capacity := s.SupplyCargoCapacityForTurns(fleetID, turns)
	remaining := capacity - supplyCargoTotal(fleet.SupplyCargo)
	if remaining <= 0 {
		return economy.ResourceCost{}, false, "Bu filo seçilen süre için azami ikmal yükünü zaten taşıyor."
	}
	cargo := economy.ResourceCost{Grain: remaining}
	if ok, reason := s.LoadSupplyCargoAtCapital(fleetID, cargo); !ok {
		return economy.ResourceCost{}, false, reason
	}
	return cargo, true, ""
}

// SupplyCargoCapacityForTurns, filonun seçilen süre boyunca taşıyabileceği
// toplam ikmal malı miktarını döner. Her asker taşıma kapasitesi puanı, en
// yüksek tahıl tüketimli kara biriminin tur başı tüketimi kadar yük taşır.
func (s *GameState) SupplyCargoCapacityForTurns(fleetID army.ArmyID, turns int) int {
	if s == nil || turns <= 0 {
		return 0
	}
	fleet := s.Armies[fleetID]
	if fleet == nil || !fleet.IsNaval {
		return 0
	}
	maxUpkeep := 0
	for _, unitType := range s.UnitTypes {
		if unitType == nil || unitType.Category == army.CategoryNavalWar || unitType.Category == army.CategoryNavalTrans || unitType.Category == army.CategoryNavalTrade {
			continue
		}
		if unitType.GrainUpkeep > maxUpkeep {
			maxUpkeep = unitType.GrainUpkeep
		}
	}
	return fleet.TransportCapacity(s.UnitTypes) * maxUpkeep * turns
}

// SupplyCargoLoadedAmount, filodaki tüm ikmal mallarının toplam miktarını
// yükleme paneli ve filo bilgi ekranı için döner.
func (s *GameState) SupplyCargoLoadedAmount(fleetID army.ArmyID) int {
	if s == nil || s.Armies[fleetID] == nil {
		return 0
	}
	return supplyCargoTotal(s.Armies[fleetID].SupplyCargo)
}

// UnloadSupplyCargoAtCapital, merkez limanında kalan yükü devlet stokuna iade eder.
func (s *GameState) UnloadSupplyCargoAtCapital(fleetID army.ArmyID) (economy.ResourceCost, bool, string) {
	if s == nil {
		return economy.ResourceCost{}, false, "Geçersiz oyun durumu."
	}
	fleet := s.Armies[fleetID]
	if fleet == nil || !fleet.IsNaval {
		return economy.ResourceCost{}, false, "İkmal yükü yalnızca merkez limanında boşaltılabilir."
	}
	capital, _, _, ok := s.FactionCapital(factionID(fleet.OwnerID))
	if !ok || capital == nil || fleet.DockedRegionID != capital.ID {
		return economy.ResourceCost{}, false, "İkmal yükü yalnızca merkez limanında boşaltılabilir."
	}
	cargo := fleet.SupplyCargo
	if supplyCargoTotal(cargo) <= 0 {
		return economy.ResourceCost{}, false, "Filoda boşaltılacak ikmal yükü yok."
	}
	cargo.Gold = 0
	cargo.Refund(s.Factions[factionID(fleet.OwnerID)])
	fleet.SupplyCargo = economy.ResourceCost{}
	if fleet.NavalMission != nil && fleet.NavalMission.Kind == army.NavalMissionSupplyArmy {
		fleet.NavalMission = nil
	}
	return cargo, true, ""
}

func addSupplyCargo(left, right economy.ResourceCost) economy.ResourceCost {
	return economy.ResourceCost{Grain: left.Grain + right.Grain, Iron: left.Iron + right.Iron, Timber: left.Timber + right.Timber, Stone: left.Stone + right.Stone, Spice: left.Spice + right.Spice, Cloth: left.Cloth + right.Cloth}
}

func factionID(owner string) faction.FactionID { return faction.FactionID(owner) }

func (s *GameState) validSeaMissionTarget(regionID world.RegionID) bool {
	if s == nil || regionID == "" {
		return false
	}
	region := s.Regions[regionID]
	return region != nil && region.IsSea && !region.IsLocked
}

// IsValidNavalBlockadeTarget, ablukanın yalnızca savaş halindeki düşmanın
// kıyı kara bölgelerine komşu denizlerde kurulabilmesini sağlar.
func (s *GameState) IsValidNavalBlockadeTarget(fleet *army.Army, seaID world.RegionID) bool {
	if s == nil || fleet == nil || !fleet.IsNaval || !s.validSeaMissionTarget(seaID) {
		return false
	}
	for _, land := range s.Regions {
		if land == nil || land.IsSea || land.OwnerID == "" || land.OwnerID == fleet.OwnerID || !s.atWar(fleet.OwnerID, land.OwnerID) {
			continue
		}
		for _, neighborID := range land.Neighbors {
			if neighborID == seaID {
				return true
			}
		}
	}
	return false
}

// ConvertInvalidNavalBlockadesToPatrol, komşu kıyılarda artık abluka yapılacak
// düşman bölgesi kalmayan filoları aynı denizde devriyeye çevirir. Hedef deniz
// bölgesi de düzeltilir; böylece doğrudan atanmış veya eski kayıttan gelen
// geçersiz hedef, devriye görevinin kanonik state'ine taşınır.
func (s *GameState) ConvertInvalidNavalBlockadesToPatrol() int {
	if s == nil {
		return 0
	}

	converted := 0
	for _, fleet := range s.Armies {
		if fleet == nil || !fleet.IsAtSea() || fleet.NavalMission == nil || fleet.NavalMission.Kind != army.NavalMissionBlockade {
			continue
		}
		if s.IsValidNavalBlockadeTarget(fleet, fleet.RegionID) && fleet.NavalMission.TargetRegionID == fleet.RegionID {
			continue
		}
		if !s.validSeaMissionTarget(fleet.RegionID) {
			continue
		}

		fleet.NavalMission.Kind = army.NavalMissionPatrol
		fleet.NavalMission.TargetRegionID = fleet.RegionID
		fleet.NavalMission.TargetFleetID = ""
		converted++
	}
	return converted
}

// AssignNavalMission doğrulanmış görevi filoya kopyalar.
func (s *GameState) AssignNavalMission(fleetID army.ArmyID, mission army.NavalMission) (bool, string) {
	if ok, reason := s.CanAssignNavalMission(fleetID, mission); !ok {
		return false, reason
	}
	fleet := s.Armies[fleetID]
	missionCopy := mission
	fleet.NavalMission = &missionCopy
	return true, ""
}

func (s *GameState) ClearNavalMission(fleetID army.ArmyID) bool {
	if s == nil {
		return false
	}
	fleet := s.Armies[fleetID]
	if fleet == nil || !fleet.IsNaval {
		return false
	}
	fleet.NavalMission = nil
	return true
}

// ClearNavalMissionAfterRelocation, devriye veya abluka görevi taşıyan filo
// gerçek konumundan ayrıldığında görevi temizler. LocationID kullanılır;
// böylece aynı deniz ankrajında limana girme/limandan çıkma da yeni konum
// sayılır.
func (s *GameState) ClearNavalMissionAfterRelocation(fleet *army.Army, previousLocation string) bool {
	if fleet == nil || !fleet.IsNaval || fleet.NavalMission == nil || previousLocation == "" || fleet.LocationID() == previousLocation {
		return false
	}
	if fleet.NavalMission.Kind != army.NavalMissionPatrol && fleet.NavalMission.Kind != army.NavalMissionBlockade {
		return false
	}
	fleet.NavalMission = nil
	return true
}

// NavalFleetsAutoEngage, otomatik deniz savaşının tek görev kombinasyonunu
// tanımlar: aynı denizdeki Devriye filosu açıkça hedeflenmiş Abluka filosunu
// yakalar. Görevsiz filo, escort veya nakliye bu kapıdan otomatik savaş açmaz.
func (s *GameState) NavalFleetsAutoEngage(attacker, defender *army.Army) bool {
	if s == nil || attacker == nil || defender == nil || attacker.ID == defender.ID || attacker.OwnerID == defender.OwnerID || !attacker.IsAtSea() || !defender.IsAtSea() || attacker.RegionID != defender.RegionID {
		return false
	}
	return s.NavalFleetsAutoEngageAtSea(attacker, defender, attacker.RegionID)
}

// NavalFleetsAutoEngageAtSea, hareket eden filo henüz hedef denize yazılmadan
// savunucu seçilirken görev çiftini hedef deniz üzerinde değerlendirir.
func (s *GameState) NavalFleetsAutoEngageAtSea(attacker, defender *army.Army, seaID world.RegionID) bool {
	if s == nil || attacker == nil || defender == nil || seaID == "" || attacker.ID == defender.ID || attacker.OwnerID == defender.OwnerID || !attacker.IsAtSea() || !defender.IsAtSea() {
		return false
	}
	return navalMissionTargetsSea(attacker, army.NavalMissionPatrol, seaID) && navalMissionTargetsSea(defender, army.NavalMissionBlockade, seaID) ||
		navalMissionTargetsSea(attacker, army.NavalMissionBlockade, seaID) && navalMissionTargetsSea(defender, army.NavalMissionPatrol, seaID)
}

func navalMissionTargetsSea(fleet *army.Army, kind army.NavalMissionKind, seaID world.RegionID) bool {
	return fleet != nil && fleet.NavalMission != nil && fleet.NavalMission.Kind == kind && fleet.NavalMission.TargetRegionID == seaID
}

// NavalEscortDefenseBonus, aynı deniz bölgesindeki nakliye filosuna atanmış
// escort savaş gemilerinin deniz savunmasına katkısını döndürür. Bonus yalnız
// gerçek escort hedefi de aynı savaşa savunmacı olarak katılıyorsa uygulanır.
// Birden fazla escort toplamda yüzde 30 ile sınırlıdır.
func (s *GameState) NavalEscortDefenseBonus(sourceIDs []army.ArmyID, targetRegionID world.RegionID) float64 {
	if s == nil || targetRegionID == "" || len(sourceIDs) == 0 {
		return 0
	}
	escortCount := 0
	for _, sourceID := range sourceIDs {
		escort := s.Armies[sourceID]
		if escort == nil || escort.RegionID != targetRegionID || escort.NavalMission == nil || escort.NavalMission.Kind != army.NavalMissionEscort {
			continue
		}
		transport := s.Armies[escort.NavalMission.TargetFleetID]
		if transport == nil || transport.ID == escort.ID || transport.RegionID != targetRegionID || !transport.IsNaval || transport.TransportCapacity(s.UnitTypes) <= 0 {
			continue
		}
		escortCount++
	}
	bonus := float64(escortCount) * navalEscortDefenseBonusPerFleet
	if bonus > navalEscortDefenseBonusCap {
		return navalEscortDefenseBonusCap
	}
	return bonus
}

func fleetHasWarship(s *GameState, fleet *army.Army) bool {
	if s == nil || fleet == nil {
		return false
	}
	for _, unit := range fleet.Units {
		if unitType := s.UnitTypes[unit.TypeID]; unitType != nil && unitType.Category == army.CategoryNavalWar {
			return true
		}
	}
	return false
}
