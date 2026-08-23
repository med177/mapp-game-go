package state

import (
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

const (
	DefaultCapitalMoveTurns = 5

	CapitalRegionGoldBonus   = 35
	CapitalRegionGrainBonus  = 4
	CapitalRegionIronBonus   = 1
	CapitalRegionTimberBonus = 1
	CapitalRegionStoneBonus  = 2
	CapitalRegionSpiceBonus  = 1
	CapitalRegionClothBonus  = 1

	CapitalRegionLogisticsBonus        = 6
	CapitalArmyReplenishmentMultiplier = 2
)

type CapitalMoveProgress struct {
	FactionID      faction.FactionID
	SettlementID   string
	SettlementName string
	RemainingTurns int
	Completed      bool
	Cancelled      bool
}

func (s *GameState) FindSettlementByID(settlementID string) (*world.Region, *world.Settlement, int, bool) {
	if s == nil || settlementID == "" {
		return nil, nil, -1, false
	}
	for _, region := range s.Regions {
		if region == nil {
			continue
		}
		for i := range region.Settlements {
			if region.Settlements[i].ID != settlementID {
				continue
			}
			return region, &region.Settlements[i], i, true
		}
	}
	return nil, nil, -1, false
}

func (s *GameState) FactionCapital(fid faction.FactionID) (*world.Region, *world.Settlement, int, bool) {
	if s == nil || fid == "" {
		return nil, nil, -1, false
	}
	f := s.Factions[fid]
	if f == nil || f.CapitalSettlementID == "" {
		return nil, nil, -1, false
	}
	region, settlement, idx, ok := s.FindSettlementByID(f.CapitalSettlementID)
	if !ok || region == nil || settlement == nil || region.OwnerID != string(fid) {
		return nil, nil, -1, false
	}
	return region, settlement, idx, true
}

func (s *GameState) IsCapitalRegion(region *world.Region) bool {
	if s == nil || region == nil || region.IsSea || region.OwnerID == "" {
		return false
	}
	return s.IsFactionCapitalRegion(faction.FactionID(region.OwnerID), region)
}

// IsFactionCapitalRegion, verilen faction'ın kendi ulusal başkentinin bu
// bölgede olup olmadığını döner. Bölge başka bir faction'a aitse başkent
// settlement'ı aynı ID'yi taşısa bile false döner.
func (s *GameState) IsFactionCapitalRegion(fid faction.FactionID, region *world.Region) bool {
	if s == nil || fid == "" || region == nil || region.IsSea || region.OwnerID != string(fid) {
		return false
	}
	f := s.Factions[fid]
	if f == nil || f.CapitalSettlementID == "" {
		return false
	}
	return regionHasSettlementID(region, f.CapitalSettlementID)
}

func (s *GameState) IsFactionCapitalSettlement(fid faction.FactionID, settlementID string) bool {
	if s == nil || fid == "" || settlementID == "" {
		return false
	}
	f := s.Factions[fid]
	return f != nil && f.CapitalSettlementID == settlementID
}

func (s *GameState) SetFactionCapital(fid faction.FactionID, settlementID string) bool {
	if s == nil || fid == "" || settlementID == "" {
		return false
	}
	f := s.Factions[fid]
	if f == nil {
		return false
	}
	region, _, _, ok := s.FindSettlementByID(settlementID)
	if !ok || region == nil || region.OwnerID != string(fid) || region.IsSea {
		return false
	}
	f.CapitalSettlementID = settlementID
	f.PendingCapitalSettlementID = ""
	f.PendingCapitalTurns = 0
	return true
}

func (s *GameState) StartCapitalMove(fid faction.FactionID, settlementID string, turns int) bool {
	if s == nil || fid == "" || settlementID == "" {
		return false
	}
	if turns <= 0 {
		turns = DefaultCapitalMoveTurns
	}
	f := s.Factions[fid]
	if f == nil {
		return false
	}
	region, _, _, ok := s.FindSettlementByID(settlementID)
	if !ok || region == nil || region.IsSea || region.OwnerID != string(fid) {
		return false
	}
	if f.CapitalSettlementID == settlementID {
		f.PendingCapitalSettlementID = ""
		f.PendingCapitalTurns = 0
		return false
	}
	f.PendingCapitalSettlementID = settlementID
	f.PendingCapitalTurns = turns
	return true
}

func (s *GameState) CancelCapitalMove(fid faction.FactionID) {
	if s == nil || fid == "" {
		return
	}
	if f := s.Factions[fid]; f != nil {
		f.PendingCapitalSettlementID = ""
		f.PendingCapitalTurns = 0
	}
}

func (s *GameState) AdvanceCapitalMoves() []CapitalMoveProgress {
	if s == nil {
		return nil
	}
	ids := make([]faction.FactionID, 0, len(s.Factions))
	for fid := range s.Factions {
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	updates := make([]CapitalMoveProgress, 0, len(ids))
	for _, fid := range ids {
		f := s.Factions[fid]
		if f == nil || f.PendingCapitalSettlementID == "" || f.PendingCapitalTurns <= 0 {
			continue
		}
		region, settlement, _, ok := s.FindSettlementByID(f.PendingCapitalSettlementID)
		if !ok || region == nil || settlement == nil || region.IsSea || region.OwnerID != string(fid) {
			updates = append(updates, CapitalMoveProgress{
				FactionID:      fid,
				SettlementID:   f.PendingCapitalSettlementID,
				SettlementName: pendingCapitalSettlementName(region, settlement, f.PendingCapitalSettlementID),
				Cancelled:      true,
			})
			s.CancelCapitalMove(fid)
			continue
		}
		f.PendingCapitalTurns--
		if f.PendingCapitalTurns <= 0 {
			f.CapitalSettlementID = f.PendingCapitalSettlementID
			updates = append(updates, CapitalMoveProgress{
				FactionID:      fid,
				SettlementID:   f.CapitalSettlementID,
				SettlementName: pendingCapitalSettlementName(region, settlement, f.CapitalSettlementID),
				Completed:      true,
			})
			f.PendingCapitalSettlementID = ""
			f.PendingCapitalTurns = 0
			continue
		}
		updates = append(updates, CapitalMoveProgress{
			FactionID:      fid,
			SettlementID:   f.PendingCapitalSettlementID,
			SettlementName: pendingCapitalSettlementName(region, settlement, f.PendingCapitalSettlementID),
			RemainingTurns: f.PendingCapitalTurns,
		})
	}
	return updates
}

func (s *GameState) NormalizeFactionCapitals() {
	if s == nil || len(s.Factions) == 0 {
		return
	}
	ids := make([]faction.FactionID, 0, len(s.Factions))
	for fid := range s.Factions {
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, fid := range ids {
		f := s.Factions[fid]
		if f == nil {
			continue
		}
		if _, _, _, ok := s.FactionCapital(fid); ok {
			continue
		}
		bestSettlementID, ok := s.BestCapitalSettlementForFaction(fid)
		if !ok {
			f.CapitalSettlementID = ""
			f.PendingCapitalSettlementID = ""
			f.PendingCapitalTurns = 0
			continue
		}
		f.CapitalSettlementID = bestSettlementID
		if f.PendingCapitalSettlementID != "" {
			if region, _, _, ok := s.FindSettlementByID(f.PendingCapitalSettlementID); !ok || region == nil || region.OwnerID != string(fid) {
				f.PendingCapitalSettlementID = ""
				f.PendingCapitalTurns = 0
			}
		}
	}
}

func (s *GameState) BestCapitalSettlementForFaction(fid faction.FactionID) (string, bool) {
	if s == nil || fid == "" {
		return "", false
	}
	regions := s.LandRegionsOwnedBy(fid)
	if len(regions) == 0 {
		return "", false
	}
	sort.SliceStable(regions, func(i, j int) bool {
		leftScore := capitalDevelopmentScore(regions[i])
		rightScore := capitalDevelopmentScore(regions[j])
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return regions[i].ID < regions[j].ID
	})
	idx := primarySettlementIndex(regions[0])
	if idx < 0 {
		return "", false
	}
	return regions[0].Settlements[idx].ID, true
}

// capitalDevelopmentScore, başkent kaybedildiğinde AI'nin kalan toprakları
// arasından en gelişmiş olanı seçmek için kullanılır. Gelir tek başına
// gelişmişliği temsil etmez: yüksek üretimli fakat altyapısız bir bölgenin,
// daha fazla tamamlanmış binaya sahip bölgenin önüne geçmesini engeller.
//
// Bina seviyeleri ana ölçüttür; settlement altyapısı ve nüfus ise aynı bina
// seviyesine sahip bölgeler arasındaki farkı belirler. Skor yalnızca mevcut
// state verisinden üretildiği için deterministiktir.
func capitalDevelopmentScore(region *world.Region) int {
	if region == nil || region.IsSea {
		return 0
	}

	score := len(region.Buildings) * 1000
	score += len(region.Settlements) * 100
	score += region.Population / 10
	for _, settlement := range region.Settlements {
		switch settlement.Type {
		case world.SettlementCity:
			score += 40
		case world.SettlementFortress:
			score += 35
		case world.SettlementPort:
			score += 30
		case world.SettlementTown:
			score += 20
		}
		score += settlement.Population / 10
	}
	return score
}

func (s *GameState) CapitalRegionBonus(region *world.Region) RegionProductionSummary {
	if !s.IsCapitalRegion(region) {
		return RegionProductionSummary{}
	}
	return RegionProductionSummary{
		Gold:   CapitalRegionGoldBonus,
		Grain:  CapitalRegionGrainBonus,
		Iron:   CapitalRegionIronBonus,
		Timber: CapitalRegionTimberBonus,
		Stone:  CapitalRegionStoneBonus,
		Spice:  CapitalRegionSpiceBonus,
		Cloth:  CapitalRegionClothBonus,
	}
}

func primarySettlementIndex(region *world.Region) int {
	if region == nil || len(region.Settlements) == 0 {
		return -1
	}
	for i, settlement := range region.Settlements {
		if settlement.IsCenter {
			return i
		}
	}
	return 0
}

func pendingCapitalSettlementName(region *world.Region, settlement *world.Settlement, fallback string) string {
	if settlement != nil {
		if settlement.NameTR != "" {
			return settlement.NameTR
		}
		if settlement.Name != "" {
			return settlement.Name
		}
	}
	if region != nil && region.NameTR != "" {
		return region.NameTR
	}
	return fallback
}

func regionHasSettlementID(region *world.Region, settlementID string) bool {
	if region == nil || settlementID == "" {
		return false
	}
	for _, settlement := range region.Settlements {
		if settlement.ID == settlementID {
			return true
		}
	}
	return false
}
