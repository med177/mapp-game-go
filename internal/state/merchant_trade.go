package state

import (
	"math"
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
	FromRegionID     world.RegionID
	ToRegionID       world.RegionID
	FromSettlementID string
	ToSettlementID   string
	FromSeaID        world.RegionID
	ToSeaID          world.RegionID
}

// MerchantFleetTradeStatus, tek bir merchant filosunun mevcut rota konumunda
// göstereceği bonusu ve hedefe henüz ulaşmamış olma durumunu taşır.
type MerchantFleetTradeStatus struct {
	Bonus   int
	Pending bool
}

// MerchantFleetTradeStatuses tüm atanmış merchant filolarının durumunu rota
// bazında toplu hesaplar. Aynı rotanın hedef denizi ve aktif gemi toplamı bir
// kez çözülür; renderer bu snapshot'ı aynı input/draw geçişindeki rozet,
// hit-test ve tooltip sorgularında paylaşabilir.
//
// dst verilirse map yeniden kullanılır ve önce temizlenir.
func (s *GameState) MerchantFleetTradeStatuses(dst map[army.ArmyID]MerchantFleetTradeStatus) map[army.ArmyID]MerchantFleetTradeStatus {
	if dst == nil {
		dst = make(map[army.ArmyID]MerchantFleetTradeStatus)
	} else {
		clear(dst)
	}
	if s == nil {
		return dst
	}

	routes := make(map[string]*economy.TradeRoute, len(s.TradeRoutes))
	for _, route := range s.TradeRoutes {
		if route == nil {
			continue
		}
		key := route.AssignmentKey()
		if key == "" || routes[key] != nil {
			continue
		}
		routes[key] = route
	}

	targetSeas := make(map[string]world.RegionID)
	targetResolved := make(map[string]bool)
	activeShips := make(map[string]int)
	fleetShips := make(map[army.ArmyID]int)
	supported := make(map[army.ArmyID]bool)

	for fleetID, fleet := range s.Armies {
		if fleet == nil || !fleet.IsNaval || fleet.TradeRouteKey == "" {
			continue
		}
		route := routes[fleet.TradeRouteKey]
		if route == nil {
			continue
		}
		dst[fleetID] = MerchantFleetTradeStatus{Pending: true}

		if !targetResolved[fleet.TradeRouteKey] {
			targetResolved[fleet.TradeRouteKey] = true
			if seaID, ok := s.MerchantTradeRouteTargetSeaRegion(route); ok {
				targetSeas[fleet.TradeRouteKey] = seaID
			}
		}
		targetSea := targetSeas[fleet.TradeRouteKey]
		if targetSea == "" || !fleet.IsAtSea() || fleet.OwnerID != route.FromFactionID || fleet.RegionID != targetSea {
			continue
		}

		count := s.merchantShipCount(fleet)
		fleetShips[fleetID] = count
		supported[fleetID] = true
		activeShips[fleet.TradeRouteKey] += count
	}

	for fleetID := range dst {
		if !supported[fleetID] {
			continue
		}
		fleet := s.Armies[fleetID]
		route := routes[fleet.TradeRouteKey]
		count := fleetShips[fleetID]
		remaining := s.MerchantBonusCapacity(route) - (activeShips[fleet.TradeRouteKey] - count)
		bonus := 0
		if remaining >= 0 {
			bonus = count
			if bonus > remaining {
				bonus = remaining
			}
		}
		dst[fleetID] = MerchantFleetTradeStatus{Bonus: bonus}
	}
	return dst
}

// MerchantTradeRouteTargetSeaRegion rotanın tek kanonik hedef denizini döner.
// Merchant rotaları yalnızca gerçek liman çiftiyle veya açıkça deniz olarak
// tanımlanmış tarihsel merkez bağlantısıyla geçerlidir. Kara merkez
// bağlantıları merchant filosuna deniz rotası gibi sunulmaz.
func (s *GameState) MerchantTradeRouteTargetSeaRegion(route *economy.TradeRoute) (world.RegionID, bool) {
	if s == nil || route == nil || route.SuspendedTurns > 0 || route.AssignmentKey() == "" {
		return "", false
	}
	// A historical center connection is the authored route contract. If that
	// contract resolves only to land, do not reinterpret the same agreement as
	// a sea route merely because both factions also own a port.
	if s.merchantTradeCenterDirectPathExists(route, world.TradeRouteLand) ||
		s.merchantTradeCenterPathExists(route, world.TradeRouteLand) &&
			!s.merchantTradeCenterPathExists(route, world.TradeRouteSea) {
		return "", false
	}
	if pairs := s.MerchantTradeRoutePortPairs(route); len(pairs) > 0 {
		if pairs[0].ToSeaID != "" {
			return pairs[0].ToSeaID, true
		}
	}

	fromCenters, toCenters, centers, adjacency := s.merchantTradeEndpointCenters(route)
	for _, toID := range toCenters {
		for _, fromID := range fromCenters {
			if !tradeCentersConnected(fromID, toID, adjacency) {
				continue
			}
			seaIDs := s.tradeCenterSeaIDs(centers[toID])
			if len(seaIDs) > 0 {
				return seaIDs[0], true
			}
		}
	}
	return "", false
}

// merchantTradeCenterDirectPathExists, iki faction merkezinin doğrudan
// bağlantısında yazılmış rota türünü kontrol eder. Doğrudan kara bağlantısı,
// aynı merkezler arasında dolaylı bir deniz çevrimi bulunsa bile canonical
// rota kabul edilir.
func (s *GameState) merchantTradeCenterDirectPathExists(route *economy.TradeRoute, routeType world.TradeRouteType) bool {
	if s == nil || route == nil || routeType == "" {
		return false
	}
	active := make(map[world.RegionID]world.TradeCenterDef)
	for _, def := range s.TradeCenters.Centers {
		if def.ID == "" || !def.ActiveInYear(s.Year) {
			continue
		}
		region := s.Regions[def.ID]
		if region == nil || region.IsSea || !region.IsCoastal(s.Regions) {
			continue
		}
		active[def.ID] = def
	}
	for id, def := range active {
		fromRegion := s.Regions[id]
		for _, link := range def.Links {
			linked, ok := active[link.RegionID]
			if !ok {
				continue
			}
			linkType := link.Type
			if linkType == "" {
				linkType = world.TradeRouteLand
			}
			if linkType != routeType {
				continue
			}
			toRegion := s.Regions[linked.ID]
			if fromRegion.OwnerID == route.FromFactionID && toRegion.OwnerID == route.ToFactionID ||
				fromRegion.OwnerID == route.ToFactionID && toRegion.OwnerID == route.FromFactionID {
				return true
			}
		}
	}
	return false
}

// merchantTradeCenterPathExists, iki faction merkezleri arasında yalnızca
// verilen fiziksel rota türünü kullanan bir merkez bağlantısı arar. Boş link
// türü veri sözleşmesinde kara kabul edilir; merkez bağlantıları ekonomik
// grafikte iki yönlü olduğundan gerçek merkezler arasındaki ters kenar da
// eklenir.
func (s *GameState) merchantTradeCenterPathExists(route *economy.TradeRoute, routeType world.TradeRouteType) bool {
	if s == nil || route == nil || routeType == "" {
		return false
	}
	type centerEdge struct {
		to    world.RegionID
		type_ world.TradeRouteType
	}
	active := make(map[world.RegionID]world.TradeCenterDef)
	from := make([]world.RegionID, 0)
	targets := make(map[world.RegionID]struct{})
	for _, def := range s.TradeCenters.Centers {
		if def.ID == "" || !def.ActiveInYear(s.Year) {
			continue
		}
		region := s.Regions[def.ID]
		if region == nil || region.IsSea || !region.IsCoastal(s.Regions) {
			continue
		}
		active[def.ID] = def
		if region.OwnerID == route.FromFactionID {
			from = append(from, def.ID)
		}
		if region.OwnerID == route.ToFactionID {
			targets[def.ID] = struct{}{}
		}
	}
	if len(from) == 0 || len(targets) == 0 {
		return false
	}
	adjacency := make(map[world.RegionID][]centerEdge, len(active))
	for id, def := range active {
		for _, link := range def.Links {
			if _, ok := active[link.RegionID]; !ok {
				continue
			}
			linkType := link.Type
			if linkType == "" {
				linkType = world.TradeRouteLand
			}
			if linkType != routeType {
				continue
			}
			adjacency[id] = append(adjacency[id], centerEdge{to: link.RegionID, type_: linkType})
			adjacency[link.RegionID] = append(adjacency[link.RegionID], centerEdge{to: id, type_: linkType})
		}
	}
	queue := append([]world.RegionID(nil), from...)
	seen := make(map[world.RegionID]struct{}, len(queue))
	for _, id := range queue {
		seen[id] = struct{}{}
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, ok := targets[current]; ok {
			return true
		}
		for _, edge := range adjacency[current] {
			if _, ok := seen[edge.to]; ok || edge.type_ != routeType {
				continue
			}
			seen[edge.to] = struct{}{}
			queue = append(queue, edge.to)
		}
	}
	return false
}

// MergeMerchantTradeFleets, aynı ticaret rotasının hedef denizine ulaşmış
// merchant filolarını mevcut en dolu stack'te toplar. Hareket anındaki yeni
// filo aktarımı MergeMerchantTradeFleetAtRoute tarafından doğrudan daha önce
// orada bulunan filoya yapılır. Bir filo 20 gemiye ulaştığında sonraki filo
// ayrı kalır; böylece kapasite sınırı aşılmaz.
func (s *GameState) MergeMerchantTradeFleets() int {
	if s == nil || len(s.Armies) < 2 {
		return 0
	}

	type mergeGroupKey struct {
		routeKey string
		seaID    world.RegionID
	}
	fleetIDs := make([]army.ArmyID, 0, len(s.Armies))
	for fleetID := range s.Armies {
		fleetIDs = append(fleetIDs, fleetID)
	}
	sort.Slice(fleetIDs, func(i, j int) bool { return fleetIDs[i] < fleetIDs[j] })

	groups := make(map[mergeGroupKey][]army.ArmyID)
	for _, fleetID := range fleetIDs {
		fleet := s.Armies[fleetID]
		route, seaID, ok := s.merchantTradeMergeRoute(fleet)
		if !ok {
			continue
		}
		key := mergeGroupKey{routeKey: route.AssignmentKey(), seaID: seaID}
		groups[key] = append(groups[key], fleetID)
	}

	removed := 0
	for key, group := range groups {
		if len(group) < 2 {
			continue
		}
		sort.SliceStable(group, func(i, j int) bool {
			left := len(s.Armies[group[i]].Units)
			right := len(s.Armies[group[j]].Units)
			if left != right {
				return left > right
			}
			return group[i] < group[j]
		})
		target := s.Armies[group[0]]
		var route *economy.TradeRoute
		for _, candidate := range s.TradeRoutes {
			if candidate != nil && candidate.AssignmentKey() == key.routeKey {
				route = candidate
				break
			}
		}
		if !s.merchantTradeMergeEligible(target, route, key.seaID) {
			continue
		}
		for _, sourceID := range group[1:] {
			source := s.Armies[sourceID]
			if source == nil || !s.merchantTradeMergeEligible(source, route, key.seaID) {
				continue
			}
			if s.mergeMerchantTradeFleetUnits(target, source) {
				if _, exists := s.Armies[sourceID]; !exists {
					removed++
				}
			}
		}
	}
	return removed
}

// MergeMerchantTradeFleetAtRoute, yeni ulaşan merchant filosunu aynı rota ve
// hedef denizdeki daha önce mevcut filolara aktarır. Dönen ID, kaynak filo
// tamamen aktarıldıysa hayatta kalan hedefi; kısmi aktarımda kaynağın kendisini
// gösterir.
func (s *GameState) MergeMerchantTradeFleetAtRoute(fleetID army.ArmyID) (army.ArmyID, bool) {
	if s == nil {
		return fleetID, false
	}
	source := s.Armies[fleetID]
	if source == nil {
		return fleetID, false
	}
	route, seaID, ok := s.merchantTradeMergeRoute(source)
	if !ok {
		return fleetID, false
	}

	targetIDs := make([]army.ArmyID, 0, len(s.Armies))
	for targetID := range s.Armies {
		if targetID != fleetID {
			targetIDs = append(targetIDs, targetID)
		}
	}
	sort.Slice(targetIDs, func(i, j int) bool { return targetIDs[i] < targetIDs[j] })
	for _, targetID := range targetIDs {
		target := s.Armies[targetID]
		if !s.merchantTradeMergeEligible(target, route, seaID) {
			continue
		}
		if s.mergeMerchantTradeFleetUnits(target, source) {
			if _, exists := s.Armies[fleetID]; !exists {
				return targetID, true
			}
			if len(source.Units) == 0 {
				return targetID, true
			}
			return fleetID, true
		}
	}
	return fleetID, false
}

func (s *GameState) merchantTradeMergeRoute(fleet *army.Army) (*economy.TradeRoute, world.RegionID, bool) {
	if s == nil || fleet == nil || !fleet.IsAtSea() || fleet.TradeRouteKey == "" || len(fleet.EmbarkedUnits) > 0 || fleet.NavalMission != nil {
		return nil, "", false
	}
	if s.merchantShipCount(fleet) != len(fleet.Units) || len(fleet.Units) == 0 {
		return nil, "", false
	}
	for _, route := range s.TradeRoutes {
		if route == nil || route.AssignmentKey() != fleet.TradeRouteKey || route.FromFactionID != fleet.OwnerID {
			continue
		}
		seaID, ok := s.MerchantTradeRouteTargetSeaRegion(route)
		if ok && fleet.RegionID == seaID {
			return route, seaID, true
		}
	}
	return nil, "", false
}

func (s *GameState) merchantTradeMergeEligible(fleet *army.Army, route *economy.TradeRoute, seaID world.RegionID) bool {
	if fleet == nil || route == nil || fleet.RegionID != seaID || fleet.TradeRouteKey != route.AssignmentKey() {
		return false
	}
	_, _, ok := s.merchantTradeMergeRoute(fleet)
	return ok
}

func (s *GameState) mergeMerchantTradeFleetUnits(target, source *army.Army) bool {
	if target == nil || source == nil || target.ID == source.ID {
		return false
	}
	capacity := army.MaxArmySize - len(target.Units)
	if capacity <= 0 {
		return false
	}
	transferCount := len(source.Units)
	if transferCount > capacity {
		transferCount = capacity
	}
	if transferCount <= 0 {
		return false
	}
	target.Units = append(target.Units, source.Units[:transferCount]...)
	source.Units = source.Units[transferCount:]
	if len(source.Units) > 0 {
		return true
	}
	if source.Commander != nil {
		if target.Commander == nil {
			target.Commander = source.Commander
			target.Commander.AssignedArmyID = target.ID
		} else {
			source.Commander.AssignedArmyID = ""
		}
	}
	s.RemoveArmy(source.ID)
	return true
}

// MerchantTradeRouteSeaRegions, geriye dönük ortak API sözleşmesi olarak
// rotanın tek hedef denizini slice içinde döner. Bonus, AI hareketi ve abluka
// bu aynı kanonik hedefi kullanır.
func (s *GameState) MerchantTradeRouteSeaRegions(route *economy.TradeRoute) []world.RegionID {
	seaID, ok := s.MerchantTradeRouteTargetSeaRegion(route)
	if !ok {
		return nil
	}
	return []world.RegionID{seaID}
}

// MerchantTradeRoutePortPairs, anlaşmanın taraflarının canonical limanları
// arasındaki deniz bağlantılarını döner. Her taraf için başkent bölgesindeki
// kullanılabilir liman, yoksa başkente en yakın kullanılabilir liman seçilir;
// aynı sonuç görsel rota, merchant hedef denizi ve AI üretiminde kullanılır.
func (s *GameState) MerchantTradeRoutePortPairs(route *economy.TradeRoute) []MerchantTradePortPair {
	return s.merchantTradeRoutePortPairs(route, s.merchantTradePortEndpoints)
}

// MerchantTradePortRegion, fraksiyonun mevcut canonical ana ticaret portu
// bölgesini döner. Görsel marker gibi state dışı tüketiciler de rota hesabıyla
// aynı başkent/mesafe seçimini kullanır.
func (s *GameState) MerchantTradePortRegion(ownerID string) *world.Region {
	return s.merchantTradePortRegion(ownerID)
}

// MerchantTradePortSettlementID, seçilen canonical ana ticaret portu
// bölgesindeki port settlement'ının ID'sini döner. Settlement verisi olmayan
// eski state'lerde boş dönebilir; deniz endpoint'i yine bölge merkezinden
// deterministik olarak seçilir.
func (s *GameState) MerchantTradePortSettlementID(ownerID string) string {
	if s == nil || ownerID == "" {
		return ""
	}
	region := s.merchantTradePortRegion(ownerID)
	if region == nil {
		return ""
	}
	endpoint := s.merchantTradePortEndpointsForRegion(region)
	if len(endpoint) == 0 {
		return ""
	}
	return endpoint[0].settlementID
}

func (s *GameState) merchantTradeRoutePortPairs(route *economy.TradeRoute, endpointFn func(string) []merchantTradePortEndpoint) []MerchantTradePortPair {
	if s == nil || route == nil || route.SuspendedTurns > 0 || route.AssignmentKey() == "" {
		return nil
	}
	fromPorts := endpointFn(route.FromFactionID)
	toPorts := endpointFn(route.ToFactionID)
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
				FromRegionID:     from.regionID,
				ToRegionID:       to.regionID,
				FromSettlementID: from.settlementID,
				ToSettlementID:   to.settlementID,
				FromSeaID:        from.seaID,
				ToSeaID:          to.seaID,
			})
		}
	}
	distanceCache := make(map[string]int, len(pairs))
	pairDistance := func(pair MerchantTradePortPair) int {
		key := string(pair.FromSeaID) + "|" + string(pair.ToSeaID)
		if distance, ok := distanceCache[key]; ok {
			return distance
		}
		distance := s.merchantTradeSeaDistance(pair.FromSeaID, pair.ToSeaID)
		distanceCache[key] = distance
		return distance
	}
	sort.Slice(pairs, func(i, j int) bool {
		distanceI := pairDistance(pairs[i])
		distanceJ := pairDistance(pairs[j])
		if distanceI != distanceJ {
			return distanceI < distanceJ
		}
		if pairs[i].FromRegionID != pairs[j].FromRegionID {
			return pairs[i].FromRegionID < pairs[j].FromRegionID
		}
		if pairs[i].ToRegionID != pairs[j].ToRegionID {
			return pairs[i].ToRegionID < pairs[j].ToRegionID
		}
		if pairs[i].FromSettlementID != pairs[j].FromSettlementID {
			return pairs[i].FromSettlementID < pairs[j].FromSettlementID
		}
		if pairs[i].ToSettlementID != pairs[j].ToSettlementID {
			return pairs[i].ToSettlementID < pairs[j].ToSettlementID
		}
		if pairs[i].FromSeaID != pairs[j].FromSeaID {
			return pairs[i].FromSeaID < pairs[j].FromSeaID
		}
		return pairs[i].ToSeaID < pairs[j].ToSeaID
	})
	return pairs
}

type merchantTradePortEndpoint struct {
	regionID     world.RegionID
	settlementID string
	seaID        world.RegionID
}

func (s *GameState) merchantTradePortEndpoints(ownerID string) []merchantTradePortEndpoint {
	if s == nil || ownerID == "" {
		return nil
	}
	region := s.merchantTradePortRegion(ownerID)
	if region == nil {
		return nil
	}
	return s.merchantTradePortEndpointsForRegion(region)
}

func (s *GameState) merchantTradePortEndpointsForRegion(region *world.Region) []merchantTradePortEndpoint {
	if s == nil || region == nil {
		return nil
	}
	portSettlement, hasPortSettlement := s.merchantTradePortSettlement(region)
	portX, portY := region.WorldX, region.WorldY
	settlementID := ""
	if hasPortSettlement {
		portX = portSettlement.X
		portY = portSettlement.Y
		settlementID = portSettlement.ID
	}
	seaID := s.merchantTradePortFacingSea(region, portX, portY)
	if seaID == "" {
		return nil
	}
	return []merchantTradePortEndpoint{{
		regionID:     region.ID,
		settlementID: settlementID,
		seaID:        seaID,
	}}
}

func (s *GameState) merchantTradePortSettlement(region *world.Region) (world.Settlement, bool) {
	if s == nil || region == nil {
		return world.Settlement{}, false
	}
	targetX, targetY := region.WorldX, region.WorldY
	if capitalRegion, capitalSettlement, _, ok := s.FactionCapital(faction.FactionID(region.OwnerID)); ok && capitalRegion != nil && capitalRegion.ID == region.ID && capitalSettlement != nil {
		targetX = capitalSettlement.X
		targetY = capitalSettlement.Y
	}

	var best world.Settlement
	bestDistance := int64(0)
	found := false
	for _, settlement := range region.Settlements {
		if settlement.Type != world.SettlementPort {
			continue
		}
		dx := int64(settlement.X - targetX)
		dy := int64(settlement.Y - targetY)
		distance := dx*dx + dy*dy
		if !found || distance < bestDistance || (distance == bestDistance && settlement.ID < best.ID) {
			best = settlement
			bestDistance = distance
			found = true
		}
	}
	return best, found
}

func (s *GameState) merchantTradePortFacingSea(region *world.Region, portX, portY int) world.RegionID {
	if s == nil || region == nil {
		return ""
	}
	portVectorX := float64(portX - region.WorldX)
	portVectorY := float64(portY - region.WorldY)
	portVectorLength := math.Hypot(portVectorX, portVectorY)

	seaIDs := make([]world.RegionID, 0, len(region.Neighbors))
	for _, neighborID := range region.Neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor == nil || !neighbor.IsSea || neighbor.IsLocked {
			continue
		}
		seaIDs = append(seaIDs, neighborID)
	}
	sort.Slice(seaIDs, func(i, j int) bool { return seaIDs[i] < seaIDs[j] })
	var bestID world.RegionID
	bestAlignment := -math.MaxFloat64
	bestDistance := math.MaxFloat64
	for _, seaID := range seaIDs {
		sea := s.Regions[seaID]
		seaVectorX := float64(sea.WorldX - region.WorldX)
		seaVectorY := float64(sea.WorldY - region.WorldY)
		seaVectorLength := math.Hypot(seaVectorX, seaVectorY)
		alignment := 0.0
		if portVectorLength > 0 && seaVectorLength > 0 {
			alignment = (portVectorX*seaVectorX + portVectorY*seaVectorY) / (portVectorLength * seaVectorLength)
		}
		distanceX := float64(sea.WorldX - portX)
		distanceY := float64(sea.WorldY - portY)
		distance := distanceX*distanceX + distanceY*distanceY
		if bestID == "" || alignment > bestAlignment+1e-9 ||
			(math.Abs(alignment-bestAlignment) <= 1e-9 && (distance < bestDistance ||
				(math.Abs(distance-bestDistance) <= 1e-9 && seaID < bestID))) {
			bestID = seaID
			bestAlignment = alignment
			bestDistance = distance
		}
	}
	return bestID
}

// merchantTradePortRegion, tüm merchant rota tüketicilerinde kullanılacak tek
// canonical kara limanı bölgesini seçer. Başkentte kullanılabilir liman varsa
// öncelik ondadır; aksi halde başkente WorldX/WorldY karesel mesafesi en küçük
// liman seçilir. Başkent çözümlenemeyen eski/eksik state'lerde ID sırası
// fallback'tir.
func (s *GameState) merchantTradePortRegion(ownerID string) *world.Region {
	if s == nil || ownerID == "" {
		return nil
	}

	capital, _, _, _ := s.FactionCapital(faction.FactionID(ownerID))
	if s.isMerchantTradePortRegion(capital, ownerID) {
		return capital
	}

	regionIDs := make([]world.RegionID, 0, len(s.Regions))
	for regionID := range s.Regions {
		regionIDs = append(regionIDs, regionID)
	}
	sort.Slice(regionIDs, func(i, j int) bool { return regionIDs[i] < regionIDs[j] })

	var best *world.Region
	var bestDistance int64
	for _, regionID := range regionIDs {
		region := s.Regions[regionID]
		if !s.isMerchantTradePortRegion(region, ownerID) {
			continue
		}

		distance := int64(0)
		if capital != nil {
			dx := int64(region.WorldX - capital.WorldX)
			dy := int64(region.WorldY - capital.WorldY)
			distance = dx*dx + dy*dy
		}
		if best == nil || distance < bestDistance || (distance == bestDistance && region.ID < best.ID) {
			best = region
			bestDistance = distance
		}
	}
	return best
}

func (s *GameState) isMerchantTradePortRegion(region *world.Region, ownerID string) bool {
	if s == nil || region == nil || region.OwnerID != ownerID || region.IsSea || region.IsLocked || !region.HasPort() {
		return false
	}
	for _, neighborID := range region.Neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor != nil && neighbor.IsSea && !neighbor.IsLocked {
			return true
		}
	}
	return false
}

func (s *GameState) merchantTradeSeasConnected(start, target world.RegionID) bool {
	return s.merchantTradeSeaDistance(start, target) < int(^uint(0)>>1)
}

func (s *GameState) merchantTradeSeaDistance(start, target world.RegionID) int {
	if s == nil || start == "" || target == "" {
		return int(^uint(0) >> 1)
	}
	if start == target {
		return 0
	}
	seen := map[world.RegionID]struct{}{start: {}}
	type seaStep struct {
		id       world.RegionID
		distance int
	}
	queue := []seaStep{{id: start}}
	for len(queue) > 0 {
		currentStep := queue[0]
		queue = queue[1:]
		current := s.Regions[currentStep.id]
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
				return currentStep.distance + 1
			}
			seen[neighborID] = struct{}{}
			queue = append(queue, seaStep{id: neighborID, distance: currentStep.distance + 1})
		}
	}
	return int(^uint(0) >> 1)
}

// MerchantFleetSupportsTradeRoute filonun kendi ihracat rotasının ucundaki
// geçerli ticaret merkezi veya liman denizinde bulunup bulunmadığını bildirir.
// Merchant filoları yalnızca sahibi olduğu FromFactionID'nin mal sattığı
// yönde çalışır; karşı tarafın ihracat rotası oyuncunun gelir rotası değildir.
func (s *GameState) MerchantFleetSupportsTradeRoute(fleet *army.Army, route *economy.TradeRoute) bool {
	if s == nil || fleet == nil || route == nil || !fleet.IsAtSea() || fleet.TradeRouteKey != route.AssignmentKey() {
		return false
	}
	if fleet.OwnerID != route.FromFactionID {
		return false
	}
	for _, seaID := range s.MerchantTradeRouteSeaRegions(route) {
		if fleet.RegionID == seaID {
			return true
		}
	}
	return false
}

// MerchantTradeRouteAssignedMerchantShips rotaya atanmış merchant
// gemilerinin toplamını döner. excludeFleetID, mevcut filonun kendi
// atamasını kapasite kontrolünden çıkarmak için kullanılır.
func (s *GameState) MerchantTradeRouteAssignedMerchantShips(route *economy.TradeRoute, excludeFleetID army.ArmyID) int {
	if s == nil || route == nil || route.AssignmentKey() == "" {
		return 0
	}
	count := 0
	for fleetID, fleet := range s.Armies {
		if fleetID == excludeFleetID || fleet == nil || fleet.TradeRouteKey != route.AssignmentKey() {
			continue
		}
		count += s.merchantShipCount(fleet)
	}
	return count
}

// MerchantTradeRouteActiveMerchantShips yalnızca hedef denizine ulaşmış ve
// gerçek bonus üretebilen merchant gemilerini sayar. Atanmış fakat yoldaki
// filolar kapasiteyi tüketmez.
func (s *GameState) MerchantTradeRouteActiveMerchantShips(route *economy.TradeRoute, excludeFleetID army.ArmyID) int {
	if s == nil || route == nil || route.AssignmentKey() == "" {
		return 0
	}
	count := 0
	for fleetID, fleet := range s.Armies {
		if fleetID == excludeFleetID || fleet == nil || fleet.TradeRouteKey != route.AssignmentKey() || !s.MerchantFleetSupportsTradeRoute(fleet, route) {
			continue
		}
		count += s.merchantShipCount(fleet)
	}
	return count
}

// MerchantTradeRouteHasCapacityForFleet bildirir; seçili filo mevcut
// atamasını koruyarak rotaya atanabilir ve bonus üretebilir mi?
func (s *GameState) MerchantTradeRouteHasCapacityForFleet(fleet *army.Army, route *economy.TradeRoute) bool {
	if s == nil || fleet == nil || route == nil {
		return false
	}
	return s.MerchantTradeRouteActiveMerchantShips(route, fleet.ID) < s.MerchantBonusCapacity(route)
}

// MerchantBonusCapacity, bir ticaret rotasının merchant filosuna açtığı
// ek hacim tavanını döner. Temel rota hacmi korunur; rotanın iki ucundaki
// aktif ticaret merkezleri, merchant gemilerinin kullanabileceği ek ticari
// kapasite sağlar.
func (s *GameState) MerchantBonusCapacity(route *economy.TradeRoute) int {
	if route == nil {
		return 0
	}
	base := economy.MerchantBonusCapacity(route)
	if s == nil || base <= 0 {
		return base
	}
	return base + s.merchantRouteCenterCapacityBonus(route)
}

func (s *GameState) merchantRouteCenterCapacityBonus(route *economy.TradeRoute) int {
	if s == nil || route == nil {
		return 0
	}
	config := s.TradeCenters.ApplyDefaultBonuses()
	bonusForFaction := func(fid string) int {
		best := 0
		for _, center := range config.Centers {
			region := s.Regions[center.ID]
			if region == nil || region.OwnerID != fid || !center.ActiveInYear(s.Year) {
				continue
			}
			bonus := config.SecondaryMerchantCapacityBonus
			if center.Tier == world.TradeCenterPrimary {
				bonus = config.PrimaryMerchantCapacityBonus
			}
			if center.MerchantCapacityBonus > 0 {
				bonus = center.MerchantCapacityBonus
			}
			if bonus > best {
				best = bonus
			}
		}
		return best
	}
	return bonusForFaction(route.FromFactionID) + bonusForFaction(route.ToFactionID)
}

// MerchantFleetTradeRouteCapacityBonus, filonun rota kapasitesinden alabileceği
// bonusu konumdan bağımsız döner. MerchantFleetTradeRouteBonus bunun konum
// uygunluğu doğrulanmış halidir.
func (s *GameState) MerchantFleetTradeRouteCapacityBonus(fleet *army.Army, route *economy.TradeRoute) int {
	if s == nil || fleet == nil || route == nil {
		return 0
	}
	count := s.merchantShipCount(fleet)
	capacity := s.MerchantBonusCapacity(route) - s.MerchantTradeRouteActiveMerchantShips(route, fleet.ID)
	if capacity < 0 {
		return 0
	}
	if count > capacity {
		return capacity
	}
	return count
}

// MerchantTradeRouteActiveBonus, ekonomi çözümlemesine yan etki vermeden,
// hedef denizine ulaşmış merchant gemilerinin rotaya sağlayacağı hacmi döner.
// RefreshMerchantTradeBonuses aynı sözleşmeyi runtime alanına yazar.
func (s *GameState) MerchantTradeRouteActiveBonus(route *economy.TradeRoute) int {
	if s == nil || route == nil || route.SuspendedTurns > 0 || route.AssignmentKey() == "" {
		return 0
	}
	capacity := s.MerchantBonusCapacity(route)
	if capacity <= 0 {
		return 0
	}
	fleetIDs := make([]army.ArmyID, 0, len(s.Armies))
	for fleetID := range s.Armies {
		fleetIDs = append(fleetIDs, fleetID)
	}
	sort.Slice(fleetIDs, func(i, j int) bool { return fleetIDs[i] < fleetIDs[j] })
	total := 0
	for _, fleetID := range fleetIDs {
		fleet := s.Armies[fleetID]
		if fleet == nil || !s.MerchantFleetSupportsTradeRoute(fleet, route) {
			continue
		}
		remaining := capacity - total
		if remaining <= 0 {
			break
		}
		count := s.merchantShipCount(fleet)
		if count > remaining {
			count = remaining
		}
		total += count
	}
	return total
}

// MerchantTradeRouteEffectiveAmount, preview katmanının henüz runtime rota
// bonusu yenilenmemiş olsa bile mevcut filo konumunu doğru görmesini sağlar.
func (s *GameState) MerchantTradeRouteEffectiveAmount(route *economy.TradeRoute) int {
	if s == nil || route == nil {
		return 0
	}
	bonus := s.MerchantTradeRouteActiveBonus(route)
	amount := route.AmountPerTurn + bonus
	if amount < 0 {
		amount = 0
	}
	blockade := route.BlockadePercent
	if blockade < 0 {
		blockade = 0
	}
	if blockade > economy.MaxTradeRouteBlockadePercent {
		blockade = economy.MaxTradeRouteBlockadePercent
	}
	return amount * (economy.MaxTradeRouteBlockadePercent - blockade) / economy.MaxTradeRouteBlockadePercent
}

// MerchantTradeIncomeForRoute, gerçekten taşınan merchant hacminden doğan
// aracılık gelirini döner. Aktif olmayan, abluka nedeniyle yalnızca temel
// hacmi taşıyan veya başarısız olan rotalar merchant kârı üretmez.
func (s *GameState) MerchantTradeIncomeForRoute(route *economy.TradeRoute, transportedVolume int) int {
	if s == nil || route == nil || transportedVolume <= 0 || route.SuspendedTurns > 0 {
		return 0
	}
	merchantVolume := transportedVolume - route.AmountPerTurn
	if merchantVolume <= 0 {
		return 0
	}
	activeBonus := s.MerchantTradeRouteActiveBonus(route)
	if merchantVolume > activeBonus {
		merchantVolume = activeBonus
	}
	if merchantVolume <= 0 {
		return 0
	}
	perShip := s.merchantTradeIncomePerShip()
	if perShip <= 0 {
		return 0
	}
	perShip += s.merchantRouteCenterIncomeBonus(route)
	return merchantVolume * perShip * s.MerchantTradeRouteSafetyPercent(route) / 100
}

func (s *GameState) merchantTradeIncomePerShip() int {
	if s == nil {
		return 8
	}
	for _, unitType := range s.UnitTypes {
		if unitType == nil || unitType.Category != army.CategoryNavalTrade {
			continue
		}
		if unitType.MerchantTradeIncome > 0 {
			return unitType.MerchantTradeIncome
		}
	}
	return 8
}

func (s *GameState) merchantRouteCenterIncomeBonus(route *economy.TradeRoute) int {
	if s == nil || route == nil {
		return 0
	}
	config := s.TradeCenters.ApplyDefaultBonuses()
	bonusForFaction := func(fid string) int {
		best := 0
		for _, center := range config.Centers {
			region := s.Regions[center.ID]
			if region == nil || region.OwnerID != fid || !center.ActiveInYear(s.Year) {
				continue
			}
			bonus := config.SecondaryMerchantIncomeBonus
			if center.Tier == world.TradeCenterPrimary {
				bonus = config.PrimaryMerchantIncomeBonus
			}
			if center.MerchantIncomeBonus > 0 {
				bonus = center.MerchantIncomeBonus
			}
			if bonus > best {
				best = bonus
			}
		}
		return best
	}
	return maxIntState(bonusForFaction(route.FromFactionID), bonusForFaction(route.ToFactionID))
}

// MerchantTradeRouteSafetyPercent, merchant filosunun aynı hedef denizde
// devriye veya escort desteği olup olmadığını değerlendirir. Abluka kesintisi
// ayrıca rota hacminde uygulanır; bu katsayı yalnız merchant kârını etkiler.
func (s *GameState) MerchantTradeRouteSafetyPercent(route *economy.TradeRoute) int {
	if s == nil || route == nil {
		return 0
	}
	for _, seaID := range s.MerchantTradeRouteSeaRegions(route) {
		if s.patrolWarshipCountInSea(seaID, route.FromFactionID) > 0 || s.merchantRouteHasEscort(route, seaID) {
			return 100
		}
	}
	return 75
}

func (s *GameState) merchantRouteHasEscort(route *economy.TradeRoute, seaID world.RegionID) bool {
	if s == nil || route == nil || seaID == "" {
		return false
	}
	for _, fleet := range s.Armies {
		if fleet == nil || fleet.OwnerID != route.FromFactionID || !fleet.IsAtSea() || fleet.RegionID != seaID || s.fleetWarshipCount(fleet) <= 0 || fleet.NavalMission == nil {
			continue
		}
		if fleet.NavalMission.Kind != army.NavalMissionEscort {
			continue
		}
		merchantFleet := s.Armies[fleet.NavalMission.TargetFleetID]
		if merchantFleet != nil && merchantFleet.OwnerID == route.FromFactionID && merchantFleet.TradeRouteKey == route.AssignmentKey() && merchantFleet.IsAtSea() && merchantFleet.RegionID == seaID {
			return true
		}
	}
	return false
}

func maxIntState(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MerchantFleetTradeRouteBonus, seçili merchant filosunun mevcut konumunda
// rotaya sağlayacağı hacim bonusunu döner. Her merchant gemisi +1 hacim
// sağlar; rotaya atanmış diğer filoların kullandığı kapasite düşüldükten sonra
// kalan rota kapasitesini aşamaz.
func (s *GameState) MerchantFleetTradeRouteBonus(fleet *army.Army, route *economy.TradeRoute) int {
	if s == nil || fleet == nil || route == nil || !s.MerchantFleetSupportsTradeRoute(fleet, route) {
		return 0
	}
	return s.MerchantFleetTradeRouteCapacityBonus(fleet, route)
}

// MerchantTradeRoutesForFleet oyuncunun/AI'ın merchant filosuna atanabilecek
// aktif ihracat rotalarını döner. Rota yalnızca filonun sahibi FromFactionID
// ise ve rota için kanonik bir hedef deniz bulunuyorsa listelenir.
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
		if route.FromFactionID != ownerID {
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
			if fleet.TradeRouteKey != routeKey && !s.MerchantTradeRouteHasCapacityForFleet(fleet, route) {
				return false
			}
			fleet.TradeRouteKey = routeKey
			return true
		}
	}
	return false
}

// NormalizeMerchantTradeAssignments geçersizleşmiş merchant görevlerini
// temizler. Özellikle eski save'lerde kara merkez rotası TradeRouteKey olarak
// kalmış olabilir; bu atama oyuncu/AI yeni bir rota seçeneğine dokunmadan da
// state'ten kaldırılmalıdır.
func (s *GameState) NormalizeMerchantTradeAssignments() int {
	if s == nil {
		return 0
	}
	cleared := 0
	for _, fleet := range s.Armies {
		if fleet == nil || fleet.TradeRouteKey == "" {
			continue
		}
		valid := false
		for _, route := range s.MerchantTradeRoutesForFleet(fleet) {
			if route != nil && route.AssignmentKey() == fleet.TradeRouteKey {
				valid = true
				break
			}
		}
		if valid {
			continue
		}
		fleet.TradeRouteKey = ""
		cleared++
	}
	return cleared
}

// RefreshMerchantTradeBonuses runtime rota hacmini gerçek fleet assignment ve
// konumundan yeniden türetir. Her gemi +1 hacim sağlar ve rota kapasitesi
// dolduğunda sonraki filolar katkı vermez.
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
	fleetIDs := make([]army.ArmyID, 0, len(s.Armies))
	for fleetID := range s.Armies {
		fleetIDs = append(fleetIDs, fleetID)
	}
	sort.Slice(fleetIDs, func(i, j int) bool { return fleetIDs[i] < fleetIDs[j] })
	for _, fleetID := range fleetIDs {
		fleet := s.Armies[fleetID]
		if fleet == nil || !fleet.IsNaval || fleet.TradeRouteKey == "" {
			continue
		}
		route := routes[fleet.TradeRouteKey]
		if route == nil || !s.MerchantFleetSupportsTradeRoute(fleet, route) {
			continue
		}
		count := s.MerchantFleetTradeRouteBonus(fleet, route)
		remaining := s.MerchantBonusCapacity(route) - route.MerchantAmountBonus
		if count > remaining {
			count = remaining
		}
		if count > 0 {
			route.MerchantAmountBonus += count
		}
	}
}

// RefreshTradeRouteBlockades açıkça Abluka görevi taşıyan deniz savaş
// filolarının aktif ticaret rotalarına
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
			warships := s.effectiveHostileWarshipCountInSea(seaID, route.FromFactionID, route.ToFactionID)
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
		blockade := s.effectiveHostileWarshipCountInSea(neighborID, ownerID) * blockadePercentPerWarship
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
		if fleet == nil || !fleet.IsAtSea() || fleet.RegionID != seaID || !s.fleetCountsAsBlockade(fleet, seaID) {
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

// effectiveHostileWarshipCountInSea, açıkça Abluka görevi verilen düşman
// gemilerinin etkisinden, aynı denizdeki sahip devriye gemilerini düşer.
// Böylece Devriye ticaret ve liman ikmalini korur; Abluka aynı konumda gerçek
// ekonomik baskı yaratır.
func (s *GameState) effectiveHostileWarshipCountInSea(seaID world.RegionID, targetOwners ...string) int {
	hostile := s.hostileWarshipCountInSea(seaID, targetOwners...)
	patrol := s.patrolWarshipCountInSea(seaID, targetOwners...)
	if patrol >= hostile {
		return 0
	}
	return hostile - patrol
}

func (s *GameState) fleetCountsAsBlockade(fleet *army.Army, seaID world.RegionID) bool {
	if fleet == nil || fleet.NavalMission == nil {
		return false
	}
	mission := fleet.NavalMission
	return mission.Kind == army.NavalMissionBlockade && mission.TargetRegionID == seaID
}

func (s *GameState) patrolWarshipCountInSea(seaID world.RegionID, owners ...string) int {
	if s == nil || seaID == "" || len(owners) == 0 {
		return 0
	}
	count := 0
	for _, fleet := range s.Armies {
		if fleet == nil || !fleet.IsAtSea() || fleet.RegionID != seaID || fleet.NavalMission == nil {
			continue
		}
		mission := fleet.NavalMission
		if mission.Kind != army.NavalMissionPatrol || mission.TargetRegionID != seaID || s.fleetWarshipCount(fleet) <= 0 {
			continue
		}
		for _, ownerID := range owners {
			if ownerID == fleet.OwnerID {
				count += s.fleetWarshipCount(fleet)
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
	for id := range activeDefs {
		for _, link := range s.TradeCenters.Centers {
			if link.ID != id {
				continue
			}
			for _, centerLink := range link.Links {
				if centerLink.Type != world.TradeRouteSea {
					continue
				}
				if _, ok := activeDefs[centerLink.RegionID]; !ok {
					continue
				}
				adjacency[id] = appendUniqueRegionID(adjacency[id], centerLink.RegionID)
			}
		}
		// TradeAdjacency is intentionally not used here: it treats all center
		// links as economic connectivity, while merchant assignment needs the
		// authored physical route type.
	}
	sort.Slice(fromCenters, func(i, j int) bool { return fromCenters[i] < fromCenters[j] })
	sort.Slice(toCenters, func(i, j int) bool { return toCenters[i] < toCenters[j] })
	return fromCenters, toCenters, centers, adjacency
}

func (s *GameState) tradeCenterSeaIDs(center *world.Region) []world.RegionID {
	if s == nil || center == nil {
		return nil
	}
	seaIDs := make([]world.RegionID, 0, len(center.Neighbors))
	for _, neighborID := range center.Neighbors {
		if neighbor := s.Regions[neighborID]; neighbor != nil && neighbor.IsSea {
			seaIDs = append(seaIDs, neighborID)
		}
	}
	sort.Slice(seaIDs, func(i, j int) bool { return seaIDs[i] < seaIDs[j] })
	return seaIDs
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
