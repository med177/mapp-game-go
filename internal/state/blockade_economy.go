package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// BlockadeEconomicEffect, liman ablukasının bölge sahibi ve ablukacı için
// ekonomi katsayılarını tek sözleşmede taşır.
type BlockadeEconomicEffect struct {
	BlockadePercent        int
	OutputRetentionPercent int
	BlockaderLootPercent   int
}

func scaleBlockadeOutput(amount, retentionPercent int) int {
	if amount <= 0 {
		return 0
	}
	if retentionPercent < 0 {
		retentionPercent = 0
	}
	if retentionPercent > 100 {
		retentionPercent = 100
	}
	return amount * retentionPercent / 100
}

// ScaleBlockadeOutputForEconomy, game çözümlemesinin state'teki aynı abluka
// retention hesabını kullanabilmesi için dışa açık küçük arayüzdür.
func ScaleBlockadeOutputForEconomy(amount, retentionPercent int) int {
	return scaleBlockadeOutput(amount, retentionPercent)
}

// RegionBlockadeEconomicEffect, onaylanan oranları uygular:
// %50 abluka bölge çıktısının %75'ini, %100 abluka %50'sini bırakır;
// ablukacının loot oranı ise sırasıyla %5 ve %10'dur.
func (s *GameState) RegionBlockadeEconomicEffect(region *world.Region) BlockadeEconomicEffect {
	if s == nil || region == nil || region.IsSea || region.OwnerID == "" || !region.HasPort() {
		return BlockadeEconomicEffect{OutputRetentionPercent: 100}
	}
	percent := s.RegionBlockadePercent(region, region.OwnerID)
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return BlockadeEconomicEffect{
		BlockadePercent:        percent,
		OutputRetentionPercent: 100 - percent/2,
		BlockaderLootPercent:   percent / 10,
	}
}

func (s *GameState) RegionBlockadeOutputRetentionPercent(region *world.Region) int {
	return s.RegionBlockadeEconomicEffect(region).OutputRetentionPercent
}

// blockadeWarshipContributionInSea, geçici map oluşturmadan belirli bir
// ablukacının etkili gemi sayısını ve denizdeki toplam etkili gemi sayısını
// döner. Devriye gemileri önce ID'si küçük ablukacıdan başlayarak düşülür; bu
// kural çoklu ablukacılarda loot paylaşımını deterministik tutar.
func (s *GameState) blockadeWarshipContributionInSea(seaID world.RegionID, targetOwner string, blockader faction.FactionID) (total, contribution int) {
	if s == nil || seaID == "" || targetOwner == "" || blockader == "" {
		return 0, 0
	}
	hostile := 0
	requested := 0
	lessThanRequested := 0
	for _, fleet := range s.Armies {
		if fleet == nil || !fleet.IsAtSea() || fleet.RegionID != seaID || fleet.OwnerID == "" || fleet.OwnerID == targetOwner || !s.atWar(fleet.OwnerID, targetOwner) || !s.fleetCountsAsBlockade(fleet, seaID) {
			continue
		}
		warships := s.fleetWarshipCount(fleet)
		if warships <= 0 {
			continue
		}
		hostile += warships
		switch {
		case fleet.OwnerID == string(blockader):
			requested += warships
		case fleet.OwnerID < string(blockader):
			lessThanRequested += warships
		}
	}

	patrol := s.patrolWarshipCountInSea(seaID, targetOwner)
	if patrol >= hostile {
		return 0, 0
	}
	effectivePatrol := patrol - lessThanRequested
	if effectivePatrol < 0 {
		effectivePatrol = 0
	}
	if effectivePatrol > requested {
		effectivePatrol = requested
	}
	return hostile - patrol, requested - effectivePatrol
}

// BlockadeLootForFaction, ablukacı fraksiyonun sonraki tur ekonomi önizlemesi
// ve gerçek ekonomi tick'i için alacağı yerel çıktı payını döner.
func (s *GameState) BlockadeLootForFaction(fid faction.FactionID) RegionProductionSummary {
	if s == nil || fid == "" {
		return RegionProductionSummary{}
	}
	var loot RegionProductionSummary
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.OwnerID == "" || region.OwnerID == string(fid) || s.SiegeAt(region.ID) != nil {
			continue
		}
		effect := s.RegionBlockadeEconomicEffect(region)
		if effect.BlockaderLootPercent <= 0 {
			continue
		}
		bestTotal, bestContribution := 0, 0
		for _, neighborID := range region.Neighbors {
			sea := s.Regions[neighborID]
			if sea == nil || !sea.IsSea {
				continue
			}
			total, contribution := s.blockadeWarshipContributionInSea(neighborID, region.OwnerID, fid)
			if total > bestTotal {
				bestTotal, bestContribution = total, contribution
			}
		}
		if bestTotal <= 0 || bestContribution <= 0 {
			continue
		}
		base := s.UnblockedRegionProductionSummary(region)
		lootPercent := effect.BlockaderLootPercent * bestContribution / bestTotal
		loot.Gold += scaleBlockadeOutput(base.Gold, lootPercent)
		loot.Grain += scaleBlockadeOutput(base.Grain, lootPercent)
		loot.Iron += scaleBlockadeOutput(base.Iron, lootPercent)
		loot.Timber += scaleBlockadeOutput(base.Timber, lootPercent)
		loot.Stone += scaleBlockadeOutput(base.Stone, lootPercent)
		loot.Spice += scaleBlockadeOutput(base.Spice, lootPercent)
		loot.Cloth += scaleBlockadeOutput(base.Cloth, lootPercent)
	}
	return loot
}

// BlockadeLootGoldForFleet, hedef denizdeki tek bir abluka filosunun mevcut
// kıyı üretiminden alacağı altın ganimetini döner. Tooltip ve ekonomi önizleme
// aynı %5/%10 oranını kullanır; filo hedef denizde değilse sonuç sıfırdır.
func (s *GameState) BlockadeLootGoldForFleet(fleet *army.Army) int {
	if s == nil || fleet == nil || !fleet.IsNaval || !fleet.IsAtSea() || fleet.OwnerID == "" || fleet.NavalMission == nil {
		return 0
	}
	mission := fleet.NavalMission
	if mission.Kind != army.NavalMissionBlockade || mission.TargetRegionID == "" || fleet.RegionID != mission.TargetRegionID {
		return 0
	}
	fleetWarships := s.fleetWarshipCount(fleet)
	if fleetWarships <= 0 {
		return 0
	}

	gold := 0
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.OwnerID == "" || region.OwnerID == fleet.OwnerID || s.SiegeAt(region.ID) != nil {
			continue
		}
		adjacent := false
		for _, neighborID := range region.Neighbors {
			if neighborID == fleet.RegionID {
				adjacent = true
				break
			}
		}
		if !adjacent {
			continue
		}
		effect := s.RegionBlockadeEconomicEffect(region)
		if effect.BlockaderLootPercent <= 0 {
			continue
		}
		total, factionContribution := s.blockadeWarshipContributionInSea(fleet.RegionID, region.OwnerID, faction.FactionID(fleet.OwnerID))
		if total <= 0 || factionContribution <= 0 {
			continue
		}
		requested := 0
		for _, candidate := range s.Armies {
			if candidate == nil || candidate.OwnerID != fleet.OwnerID || !candidate.IsAtSea() || candidate.RegionID != fleet.RegionID || !s.fleetCountsAsBlockade(candidate, fleet.RegionID) {
				continue
			}
			requested += s.fleetWarshipCount(candidate)
		}
		if requested <= 0 {
			continue
		}
		fleetContribution := fleetWarships
		if factionContribution < requested {
			fleetContribution = fleetWarships * factionContribution / requested
		}
		if fleetContribution <= 0 {
			continue
		}
		lootPercent := effect.BlockaderLootPercent * fleetContribution / total
		gold += scaleBlockadeOutput(s.UnblockedRegionProductionSummary(region).Gold, lootPercent)
	}
	return gold
}
