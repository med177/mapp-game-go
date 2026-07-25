package state

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

const blockadePercentPerWarship = 50

// MerchantTradePortPair, tarihsel ticaret merkezi bulunmayan ancak iki tarafın
// aktif limanları arasında deniz bağlantısı olan merchant rotasının görsel ve
// lojistik uçlarını taşır.
type MerchantTradePortPair struct {
	FromRegionID world.RegionID
	ToRegionID   world.RegionID
	FromSeaID    world.RegionID
	ToSeaID      world.RegionID
}

// MerchantTradeRouteSeaRegions, merchant filosunun çalışabileceği kıyı
// denizlerini döner. Öncelik tarihsel ticaret merkezi/link ağıdır; bu ağ rota
// ucu üretmiyorsa anlaşma taraflarının sahip olduğu, denize bağlı aktif
// limanlar fallback olarak kullanılır.
func (s *GameState) MerchantTradeRouteSeaRegions(route *economy.TradeRoute) []world.RegionID {
	if s == nil || route == nil || route.SuspendedTurns > 0 || route.AssignmentKey() == "" {
		return nil
	}
	fromCenters, toCenters, centers, adjacency := s.merchantTradeEndpointCenters(route)
	if len(fromCenters) == 0 && len(toCenters) == 0 {
		return s.merchantTradePortSeaRegions(route)
	}

	validFrom := make(map[world.RegionID]bool, len(fromCenters))
	validTo := make(map[world.RegionID]bool, len(toCenters))
	if len(fromCenters) == 0 || len(toCenters) == 0 {
		for _, centerID := range fromCenters {
			validFrom[centerID] = true
		}
		for _, centerID := range toCenters {
			validTo[centerID] = true
		}
	} else {
		for _, fromID := range fromCenters {
			for _, toID := range toCenters {
				if !tradeCentersConnected(fromID, toID, adjacency) {
					continue
				}
				validFrom[fromID] = true
				validTo[toID] = true
			}
		}
	}
	if len(validFrom) == 0 && len(validTo) == 0 {
		return s.merchantTradePortSeaRegions(route)
	}

	seen := make(map[world.RegionID]struct{})
	for centerID := range validFrom {
		s.addTradeCenterSeas(seen, centers[centerID])
	}
	for centerID := range validTo {
		s.addTradeCenterSeas(seen, centers[centerID])
	}
	result := make([]world.RegionID, 0, len(seen))
	for seaID := range seen {
		result = append(result, seaID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

// MerchantTradeRoutePortPairs, tarihsel merkez ağı rota uçlarını
// sağlayamadığında anlaşmanın taraflarına ait gerçek limanlar arasındaki
// deniz bağlantılarını döner. Limanlar yalnızca gerçek deniz bölgeleriyle
// kıyı komşuluğu varsa aday kabul edilir.
func (s *GameState) MerchantTradeRoutePortPairs(route *economy.TradeRoute) []MerchantTradePortPair {
	if s == nil || route == nil || route.SuspendedTurns > 0 || route.AssignmentKey() == "" {
		return nil
	}
	fromPorts := s.merchantTradePortEndpoints(route.FromFactionID)
	toPorts := s.merchantTradePortEndpoints(route.ToFactionID)
	if len(fromPorts) == 0 || len(toPorts) == 0 {
		return nil
	}

	pairs := make([]MerchantTradePortPair, 0, len(fromPorts)*len(toPorts))
	for _, from := range fromPorts {
		for _, to := range toPorts {
			if !s.merchantTradeSeasConnected(from.seaID, to.seaID) {
				continue
			}
			pairs = append(pairs, MerchantTradePortPair{
				FromRegionID: from.regionID,
				ToRegionID:   to.regionID,
				FromSeaID:    from.seaID,
				ToSeaID:      to.seaID,
			})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].FromRegionID != pairs[j].FromRegionID {
			return pairs[i].FromRegionID < pairs[j].FromRegionID
		}
		if pairs[i].ToRegionID != pairs[j].ToRegionID {
			return pairs[i].ToRegionID < pairs[j].ToRegionID
		}
		if pairs[i].FromSeaID != pairs[j].FromSeaID {
			return pairs[i].FromSeaID < pairs[j].FromSeaID
		}
		return pairs[i].ToSeaID < pairs[j].ToSeaID
	})
	return pairs
}

type merchantTradePortEndpoint struct {
	regionID world.RegionID
	seaID    world.RegionID
}

func (s *GameState) merchantTradePortEndpoints(ownerID string) []merchantTradePortEndpoint {
	if s == nil || ownerID == "" {
		return nil
	}
	result := make([]merchantTradePortEndpoint, 0)
	regionIDs := make([]world.RegionID, 0, len(s.Regions))
	for regionID := range s.Regions {
		regionIDs = append(regionIDs, regionID)
	}
	sort.Slice(regionIDs, func(i, j int) bool { return regionIDs[i] < regionIDs[j] })
	for _, regionID := range regionIDs {
		region := s.Regions[regionID]
		if region == nil || region.OwnerID != ownerID || region.IsSea || region.IsLocked || !region.HasPort() {
			continue
		}
		seaIDs := make([]world.RegionID, 0, len(region.Neighbors))
		for _, neighborID := range region.Neighbors {
			neighbor := s.Regions[neighborID]
			if neighbor == nil || !neighbor.IsSea || neighbor.IsLocked {
				continue
			}
			seaIDs = append(seaIDs, neighborID)
		}
		sort.Slice(seaIDs, func(i, j int) bool { return seaIDs[i] < seaIDs[j] })
		for _, seaID := range seaIDs {
			result = append(result, merchantTradePortEndpoint{regionID: regionID, seaID: seaID})
		}
	}
	return result
}

func (s *GameState) merchantTradePortSeaRegions(route *economy.TradeRoute) []world.RegionID {
	seen := make(map[world.RegionID]struct{})
	for _, pair := range s.MerchantTradeRoutePortPairs(route) {
		seen[pair.FromSeaID] = struct{}{}
		seen[pair.ToSeaID] = struct{}{}
	}
	result := make([]world.RegionID, 0, len(seen))
	for seaID := range seen {
		result = append(result, seaID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (s *GameState) merchantTradeSeasConnected(start, target world.RegionID) bool {
	if s == nil || start == "" || target == "" {
		return false
	}
	if start == target {
		return true
	}
	seen := map[world.RegionID]struct{}{start: {}}
	queue := []world.RegionID{start}
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		current := s.Regions[currentID]
		if current == nil || !current.IsSea || current.IsLocked {
			continue
		}
		for _, neighborID := range current.Neighbors {
			if _, ok := seen[neighborID]; ok {
				continue
			}
			neighbor := s.Regions[neighborID]
			if neighbor == nil || !neighbor.IsSea || neighbor.IsLocked {
				continue
			}
			if neighborID == target {
				return true
			}
			seen[neighborID] = struct{}{}
			queue = append(queue, neighborID)
		}
	}
	return false
}

// MerchantFleetSupportsTradeRoute filonun rota ucundaki geçerli ticaret
// merkezi veya liman denizinde bulunup bulunmadığını bildirir.
func (s *GameState) MerchantFleetSupportsTradeRoute(fleet *army.Army, route *economy.TradeRoute) bool {
	if s == nil || fleet == nil || route == nil || !fleet.IsNaval || fleet.TradeRouteKey != route.AssignmentKey() {
		return false
	}
	if fleet.OwnerID != route.FromFactionID && fleet.OwnerID != route.ToFactionID {
		return false
	}
	for _, seaID := range s.MerchantTradeRouteSeaRegions(route) {
		if fleet.RegionID == seaID {
			return true
		}
	}
	return false
}

// MerchantFleetTradeRouteBonus, seçili merchant filosunun mevcut konumunda
// rotaya sağlayacağı hacim bonusunu döner. Her rota en fazla iki merchant
// gemisinden yararlanır; filo rota denizinde değilse bonus sıfırdır.
func (s *GameState) MerchantFleetTradeRouteBonus(fleet *army.Army, route *economy.TradeRoute) int {
	if s == nil || fleet == nil || route == nil || !s.MerchantFleetSupportsTradeRoute(fleet, route) {
		return 0
	}
	count := s.merchantShipCount(fleet)
	if count > economy.MaxMerchantAmountBonusPerRoute {
		count = economy.MaxMerchantAmountBonusPerRoute
	}
	return count
}

// MerchantTradeRoutesForFleet oyuncunun/AI'ın merchant filosuna atanabilecek
// aktif rotaları döner. Rota yalnızca filonun sahibi rota uçlarından biriyse
// ve rota en az bir geçerli ticaret merkezi veya liman denizine sahipse
// listelenir.
func (s *GameState) MerchantTradeRoutesForFleet(fleet *army.Army) []*economy.TradeRoute {
	if s == nil || fleet == nil || !fleet.IsNaval || s.merchantShipCount(fleet) == 0 {
		return nil
	}

	ownerID := fleet.OwnerID
	routes := make([]*economy.TradeRoute, 0, len(s.TradeRoutes))
	for _, route := range s.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 || route.AssignmentKey() == "" {
			continue
		}
		if route.FromFactionID != ownerID && route.ToFactionID != ownerID {
			continue
		}
		if len(s.MerchantTradeRouteSeaRegions(route)) == 0 {
			continue
		}
		routes = append(routes, route)
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].GoldPerUnit != routes[j].GoldPerUnit {
			return routes[i].GoldPerUnit > routes[j].GoldPerUnit
		}
		return routes[i].AssignmentKey() < routes[j].AssignmentKey()
	})
	return routes
}

// SetMerchantTradeRoute merchant filosunun ticaret görevini doğrulayarak
// değiştirir. Boş rota anahtarı görevi kaldırır.
func (s *GameState) SetMerchantTradeRoute(fleetID army.ArmyID, routeKey string) bool {
	if s == nil || fleetID == "" {
		return false
	}
	fleet := s.Armies[fleetID]
	if fleet == nil || !fleet.IsNaval || s.merchantShipCount(fleet) == 0 {
		return false
	}
	if routeKey == "" {
		fleet.TradeRouteKey = ""
		return true
	}
	for _, route := range s.MerchantTradeRoutesForFleet(fleet) {
		if route.AssignmentKey() == routeKey {
			fleet.TradeRouteKey = routeKey
			return true
		}
	}
	return false
}

// RefreshMerchantTradeBonuses runtime rota hacmini gerçek fleet assignment ve
// konumundan yeniden türetir. Bir rota en fazla iki merchant gemisinden yararlanır.
func (s *GameState) RefreshMerchantTradeBonuses() {
	if s == nil {
		return
	}
	routes := make(map[string]*economy.TradeRoute, len(s.TradeRoutes))
	for _, route := range s.TradeRoutes {
		if route == nil {
			continue
		}
		route.MerchantAmountBonus = 0
		if route.SuspendedTurns <= 0 && route.AssignmentKey() != "" {
			routes[route.AssignmentKey()] = route
		}
	}
	for _, fleet := range s.Armies {
		if fleet == nil || !fleet.IsNaval || fleet.TradeRouteKey == "" {
			continue
		}
		route := routes[fleet.TradeRouteKey]
		if route == nil || !s.MerchantFleetSupportsTradeRoute(fleet, route) {
			continue
		}
		count := s.MerchantFleetTradeRouteBonus(fleet, route)
		remaining := economy.MaxMerchantAmountBonusPerRoute - route.MerchantAmountBonus
		if count > remaining {
			count = remaining
		}
		if count > 0 {
			route.MerchantAmountBonus += count
		}
	}
}

// RefreshTradeRouteBlockades deniz savaş filolarının aktif ticaret rotalarına
// verdiği kesintiyi gerçek konum ve savaş ilişkilerinden yeniden türetir.
// Bir savaş gemisi rota ucundaki denizdeyse rota %50, iki veya daha fazla
// savaş gemisi varsa tamamen ablukalı kabul edilir.
func (s *GameState) RefreshTradeRouteBlockades() {
	if s == nil {
		return
	}
	for _, route := range s.TradeRoutes {
		if route == nil {
			continue
		}
		route.BlockadePercent = 0
		if route.SuspendedTurns > 0 || route.AssignmentKey() == "" {
			continue
		}
		for _, seaID := range s.MerchantTradeRouteSeaRegions(route) {
			warships := s.hostileWarshipCountInSea(seaID, route.FromFactionID, route.ToFactionID)
			blockade := warships * blockadePercentPerWarship
			if blockade > route.BlockadePercent {
				route.BlockadePercent = blockade
			}
		}
		if route.BlockadePercent > economy.MaxTradeRouteBlockadePercent {
			route.BlockadePercent = economy.MaxTradeRouteBlockadePercent
		}
	}
}

// RegionBlockadePercent bir liman bölgesinin komşu denizindeki düşman savaş
// gemilerinin yerleşim/rezerv ikmal tamponunu ne kadar kestiğini döner.
func (s *GameState) RegionBlockadePercent(region *world.Region, ownerID string) int {
	if s == nil || region == nil || region.IsSea || ownerID == "" || !region.HasPort() {
		return 0
	}
	maxBlockade := 0
	for _, neighborID := range region.Neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor == nil || !neighbor.IsSea {
			continue
		}
		blockade := s.hostileWarshipCountInSea(neighborID, ownerID) * blockadePercentPerWarship
		if blockade > maxBlockade {
			maxBlockade = blockade
		}
	}
	if maxBlockade > economy.MaxTradeRouteBlockadePercent {
		return economy.MaxTradeRouteBlockadePercent
	}
	return maxBlockade
}

func (s *GameState) hostileWarshipCountInSea(seaID world.RegionID, targetOwners ...string) int {
	if s == nil || seaID == "" || len(targetOwners) == 0 {
		return 0
	}
	count := 0
	for _, fleet := range s.Armies {
		if fleet == nil || !fleet.IsNaval || fleet.RegionID != seaID {
			continue
		}
		warships := s.fleetWarshipCount(fleet)
		if warships <= 0 {
			continue
		}
		for _, targetOwner := range targetOwners {
			if targetOwner != "" && targetOwner != fleet.OwnerID && s.atWar(fleet.OwnerID, targetOwner) {
				count += warships
				break
			}
		}
	}
	return count
}

func (s *GameState) fleetWarshipCount(fleet *army.Army) int {
	if s == nil || fleet == nil || !fleet.IsNaval {
		return 0
	}
	count := 0
	for _, unit := range fleet.Units {
		unitType := s.UnitTypes[unit.TypeID]
		if unitType != nil && unitType.Category == army.CategoryNavalWar && unit.CurrentHP > 0 {
			count++
		}
	}
	return count
}

func (s *GameState) atWar(a, b string) bool {
	if s == nil || a == "" || b == "" || a == b {
		return false
	}
	relation := s.Relations[faction.RelationKey(faction.FactionID(a), faction.FactionID(b))]
	return relation != nil && relation.Stance == faction.StanceWar
}

func (s *GameState) merchantShipCount(fleet *army.Army) int {
	if s == nil || fleet == nil {
		return 0
	}
	count := 0
	for _, unit := range fleet.Units {
		unitType := s.UnitTypes[unit.TypeID]
		if unit.TypeID == "merchant_ship" || unitType != nil && unitType.Category == army.CategoryNavalTrade {
			count++
		}
	}
	return count
}

func (s *GameState) merchantTradeEndpointCenters(route *economy.TradeRoute) ([]world.RegionID, []world.RegionID, map[world.RegionID]*world.Region, map[world.RegionID][]world.RegionID) {
	centers := make(map[world.RegionID]*world.Region)
	adjacency := make(map[world.RegionID][]world.RegionID)
	var fromCenters []world.RegionID
	var toCenters []world.RegionID
	activeDefs := make(map[world.RegionID]world.TradeCenterDef)
	for _, def := range s.TradeCenters.Centers {
		if def.ID == "" || !def.ActiveInYear(s.Year) {
			continue
		}
		region := s.Regions[def.ID]
		if region == nil || region.IsSea || !region.IsCoastal(s.Regions) {
			continue
		}
		activeDefs[def.ID] = def
		centers[def.ID] = region
		if region.OwnerID == route.FromFactionID {
			fromCenters = append(fromCenters, def.ID)
		}
		if region.OwnerID == route.ToFactionID {
			toCenters = append(toCenters, def.ID)
		}
	}
	for id, def := range activeDefs {
		for _, linkedID := range def.Links {
			if _, ok := activeDefs[linkedID]; !ok {
				continue
			}
			adjacency[id] = appendUniqueRegionID(adjacency[id], linkedID)
			adjacency[linkedID] = appendUniqueRegionID(adjacency[linkedID], id)
		}
	}
	sort.Slice(fromCenters, func(i, j int) bool { return fromCenters[i] < fromCenters[j] })
	sort.Slice(toCenters, func(i, j int) bool { return toCenters[i] < toCenters[j] })
	return fromCenters, toCenters, centers, adjacency
}

func (s *GameState) addTradeCenterSeas(seen map[world.RegionID]struct{}, center *world.Region) {
	if s == nil || center == nil {
		return
	}
	for _, neighborID := range center.Neighbors {
		if neighbor := s.Regions[neighborID]; neighbor != nil && neighbor.IsSea {
			seen[neighborID] = struct{}{}
		}
	}
}

func tradeCentersConnected(start, target world.RegionID, adjacency map[world.RegionID][]world.RegionID) bool {
	if start == "" || target == "" {
		return false
	}
	if start == target {
		return true
	}
	seen := map[world.RegionID]bool{start: true}
	queue := []world.RegionID{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, neighborID := range adjacency[current] {
			if seen[neighborID] {
				continue
			}
			if neighborID == target {
				return true
			}
			seen[neighborID] = true
			queue = append(queue, neighborID)
		}
	}
	return false
}

func appendUniqueRegionID(list []world.RegionID, id world.RegionID) []world.RegionID {
	for _, existing := range list {
		if existing == id {
			return list
		}
	}
	return append(list, id)
}
