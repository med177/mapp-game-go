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
// Gerçek liman çiftleri varsa rota yönündeki hedef limanın denizi kullanılır;
// böylece örneğin Gemlik -> Özi rotası Karadeniz Açık 4'te çalışır. Gerçek
// liman çifti bulunamazsa aktif tarihsel ticaret merkezlerinin hedef tarafı
// deterministik fallback olarak kullanılır.
func (s *GameState) MerchantTradeRouteTargetSeaRegion(route *economy.TradeRoute) (world.RegionID, bool) {
	if s == nil || route == nil || route.SuspendedTurns > 0 || route.AssignmentKey() == "" {
		return "", false
	}
	if pairs := s.MerchantTradeRoutePortPairs(route); len(pairs) > 0 {
		if pairs[0].ToSeaID != "" {
			return pairs[0].ToSeaID, true
		}
	}

	fromCenters, toCenters, centers, adjacency := s.merchantTradeEndpointCenters(route)
	if len(toCenters) == 0 {
		return "", false
	}
	for _, toID := range toCenters {
		connected := len(fromCenters) == 0
		if !connected {
			for _, fromID := range fromCenters {
				if tradeCentersConnected(fromID, toID, adjacency) {
					connected = true
					break
				}
			}
		}
		if !connected {
			continue
		}
		seaIDs := s.tradeCenterSeaIDs(centers[toID])
		if len(seaIDs) > 0 {
			return seaIDs[0], true
		}
	}
	return "", false
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

// MerchantTradeRoutePortPairs, anlaşmanın taraflarına ait gerçek limanlar
// arasındaki deniz bağlantılarını döner. Limanlar yalnızca gerçek deniz
// bölgeleriyle kıyı komşuluğu varsa aday kabul edilir.
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
