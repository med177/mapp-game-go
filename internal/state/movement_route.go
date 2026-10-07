package state

import (
	"container/heap"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// MovementRouteNode, bir ordunun bu tur ulaşabildiği tek bir bölgeyi ve
// başlangıçtan itibaren seçilen en ucuz yolun bilgisini taşır.
type MovementRouteNode struct {
	RegionID world.RegionID
	Cost     int
	Previous world.RegionID
}

// MovementReachability, oyuncu hareket önizlemesi ile gerçek çok adımlı
// hareketin ortak kullandığı deterministik bölge grafiğidir.
type MovementReachability struct {
	Start  world.RegionID
	Budget int
	Nodes  map[world.RegionID]MovementRouteNode
}

// PathTo başlangıç bölgesi dahil olmak üzere hedefe giden yolu döndürür.
// Hedef erişilemiyorsa nil döner.
func (r MovementReachability) PathTo(target world.RegionID) []world.RegionID {
	if target == "" || len(r.Nodes) == 0 {
		return nil
	}
	if _, ok := r.Nodes[target]; !ok {
		return nil
	}

	path := make([]world.RegionID, 0, len(r.Nodes))
	current := target
	for current != "" {
		path = append(path, current)
		if current == r.Start {
			break
		}
		node, ok := r.Nodes[current]
		if !ok || node.Previous == "" {
			return nil
		}
		current = node.Previous
	}
	if len(path) == 0 || path[len(path)-1] != r.Start {
		return nil
	}
	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}
	return path
}

// MovementReachableForArmy, seçili ordunun bu turda ulaşabileceği bölgeleri
// hareket puanı bütçesiyle hesaplar. Kara orduları için deniz hedefi bir
// embark eylemi olduğundan rotaya dahil edilmez; filolar için kara hedefi
// yalnızca son durak olabilir.
func (s *GameState) MovementReachableForArmy(a *army.Army) MovementReachability {
	reachability := MovementReachability{Nodes: make(map[world.RegionID]MovementRouteNode)}
	if s == nil || a == nil || a.RegionID == "" || a.MovePoints <= 0 {
		return reachability
	}
	if s.Regions[a.RegionID] == nil {
		return reachability
	}

	reachability.Start = a.RegionID
	reachability.Budget = a.MovePoints
	reachability.Nodes[a.RegionID] = MovementRouteNode{RegionID: a.RegionID}

	queue := &movementRouteQueue{{regionID: a.RegionID}}
	heap.Init(queue)
	for queue.Len() > 0 {
		item := heap.Pop(queue).(movementRouteQueueItem)
		currentNode, ok := reachability.Nodes[item.regionID]
		if !ok || currentNode.Cost != item.cost {
			continue
		}
		current := s.Regions[item.regionID]
		// Başlangıç denizinde rakip filo bulunması, filonun oradan ayrılmasını
		// engellemez. Rakip filo bulunan sonraki deniz düğümü ise aşağıdaki
		// transit kontrolüyle son durak olarak kalır; temas/çatışma kararını
		// gerçek hareket çözümlemesi verir.
		if current == nil || (current.ID != a.RegionID && !s.movementRegionCanTransit(a, current)) {
			continue
		}

		neighbors := world.SortedRegionIDs(current.Neighbors)
		for _, neighborID := range neighbors {
			target := s.Regions[neighborID]
			cost, allowed := s.movementEntryCost(a, current.ID, target)
			if !allowed || currentNode.Cost+cost > a.MovePoints {
				continue
			}

			nextCost := currentNode.Cost + cost
			existing, exists := reachability.Nodes[neighborID]
			if exists && !movementRouteNodeBetter(nextCost, current.ID, existing) {
				continue
			}
			reachability.Nodes[neighborID] = MovementRouteNode{
				RegionID: neighborID,
				Cost:     nextCost,
				Previous: current.ID,
			}
			heap.Push(queue, movementRouteQueueItem{regionID: neighborID, cost: nextCost})
		}
	}
	return reachability
}

// MovementRouteForArmy, hedef bu tur ulaşılabilir olduğu sürece başlangıçtan
// hedefe kadar olan bölge zincirini döndürür.
func (s *GameState) MovementRouteForArmy(a *army.Army, target world.RegionID) []world.RegionID {
	if s == nil || a == nil || target == "" {
		return nil
	}
	reachability := s.MovementReachableForArmy(a)
	if !s.AirSortieAllowed(a, target, reachability) {
		return nil
	}
	return reachability.PathTo(target)
}

// AirSortieRequired, uçan ordunun hedef bölgede kalamayacağı hareketleri
// sortie olarak işaretler. Bu hedeflerde ordu hareket puanının yarısıyla gider
// ve hareket çözümlemesi tamamlandığında başlangıç bölgesine döner.
func (s *GameState) AirSortieRequired(a *army.Army, targetID world.RegionID) bool {
	if s == nil || a == nil || !a.UsesAirMovement(s.UnitTypes) {
		return false
	}
	target := s.Regions[targetID]
	if target == nil || target.IsSea {
		return true
	}
	ownerID := s.SovereignOwnerID(target)
	return ownerID == "" || (ownerID != a.OwnerID && !s.movementOwnerCanTransit(a.OwnerID, ownerID))
}

// AirSortieBudget, yerleşilemeyen hava hedefleri için tek yön menzilini döner.
func (s *GameState) AirSortieBudget(a *army.Army) int {
	if s == nil || a == nil || a.MovePoints <= 0 {
		return 0
	}
	budget := a.MovePoints / 2
	if budget == 0 {
		budget = 1
	}
	return budget
}

// AirSortieAllowed, hedefin normal hava hareketi veya yarım menzilli sortie
// olarak seçilebilir olup olmadığını doğrular.
func (s *GameState) AirSortieAllowed(a *army.Army, targetID world.RegionID, reachability MovementReachability) bool {
	node, ok := reachability.Nodes[targetID]
	if !ok || !s.AirSortieRequired(a, targetID) {
		return ok
	}
	return node.Cost <= s.AirSortieBudget(a)
}

// BeginAirSortie, yerleşilemeyen hava hedefi için dönüş bağlamını başlatır.
func (s *GameState) BeginAirSortie(a *army.Army, targetID world.RegionID) bool {
	if s == nil || a == nil || !s.AirSortieRequired(a, targetID) {
		return false
	}
	if a.AirSortieOriginID == "" {
		a.AirSortieOriginID = a.RegionID
		a.AirSortieTargetID = targetID
		a.AirSortieOutboundCost = 0
	}
	return a.AirSortieOriginID != "" && a.AirSortieTargetID == targetID
}

// RecordAirSortieStep, sortie rotasının tükettiği tek yön hareket puanını
// kaydeder.
func (s *GameState) RecordAirSortieStep(a *army.Army, cost int) {
	if s == nil || a == nil || a.AirSortieOriginID == "" || cost <= 0 {
		return
	}
	a.AirSortieOutboundCost += cost
}

// FinishAirSortie, uçan ordu hedefe ulaştığında onu başlangıç bölgesine döner.
func (s *GameState) FinishAirSortie(a *army.Army, targetID world.RegionID) bool {
	if !s.QueueAirSortieReturn(a, targetID) {
		return false
	}
	return s.CompleteAirSortieReturn(a)
}

// QueueAirSortieReturn, sortie hedefindeki çözümleme tamamlandıktan sonra
// dönüşün görsel hareket animasyonuyla yapılması için dönüşü beklemeye alır.
func (s *GameState) QueueAirSortieReturn(a *army.Army, targetID world.RegionID) bool {
	if s == nil || a == nil || a.AirSortieOriginID == "" || a.AirSortieTargetID != targetID {
		return false
	}
	a.AirSortieReturnPending = true
	return true
}

// CompleteAirSortieReturn, bekleyen sortie dönüşünü state üzerinde tamamlar.
func (s *GameState) CompleteAirSortieReturn(a *army.Army) bool {
	if s == nil || a == nil || !a.AirSortieReturnPending || a.AirSortieOriginID == "" {
		return false
	}
	origin := a.AirSortieOriginID
	returnCost := a.AirSortieOutboundCost
	a.PreviousRegionID = a.RegionID
	a.RegionID = origin
	a.DockedRegionID = ""
	a.DockedSettlementID = ""
	a.MovePoints -= returnCost
	if a.MovePoints < 0 {
		a.MovePoints = 0
	}
	a.AirSortieOriginID = ""
	a.AirSortieTargetID = ""
	a.AirSortieOutboundCost = 0
	a.AirSortieReturnPending = false
	return true
}

func movementRouteNodeBetter(cost int, previous world.RegionID, existing MovementRouteNode) bool {
	return cost < existing.Cost || (cost == existing.Cost && previous < existing.Previous)
}

func (s *GameState) movementEntryCost(a *army.Army, from world.RegionID, target *world.Region) (int, bool) {
	if s == nil || a == nil || target == nil {
		return 0, false
	}
	airborne := a.UsesAirMovement(s.UnitTypes)
	if target.IsLocked && !(airborne && target.IsTerrainArea) {
		return 0, false
	}
	if airborne {
		return 1, true
	}
	if a.IsNaval {
		if target.CanNavalEnter() {
			return 1, true
		}
		if target.CanLandEnter() && s.navalLandMovementAllowed(a, target) {
			return 1, true
		}
		return 0, false
	}
	if target.IsSea {
		return 0, false
	}
	return s.LandRegionEntryCost(from, target)
}

func (s *GameState) movementRegionCanTransit(a *army.Army, region *world.Region) bool {
	if s == nil || a == nil || region == nil {
		return false
	}
	airborne := a.UsesAirMovement(s.UnitTypes)
	if region.IsLocked && !(airborne && region.IsTerrainArea) {
		return false
	}
	if !airborne && a.IsNaval {
		if !region.IsSea {
			return false
		}
		for _, candidate := range s.Armies {
			if candidate == nil || candidate.ID == a.ID || candidate.RegionID != region.ID || candidate.OwnerID == a.OwnerID {
				continue
			}
			// Savaş dışı/görevsiz filolar aynı denizde bulunabilir ve rota
			// birbirlerinin içinden geçebilir. Yalnızca gerçek savaş ilişkisi
			// hedef denizi temasın son durağı yapar.
			relation := s.Relations[faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(candidate.OwnerID))]
			if relation != nil && relation.Stance == faction.StanceWar {
				return false
			}
		}
		return true
	}
	if !airborne && region.IsSea {
		return false
	}
	regionOwnerID := s.SovereignOwnerID(region)
	if regionOwnerID != "" && regionOwnerID != a.OwnerID &&
		(!airborne || s.AirspaceEnabled) && !s.movementOwnerCanTransit(a.OwnerID, regionOwnerID) {
		return false
	}
	// Başka bir ordunun bulunduğu bölge, rota için güvenli bir transit noktası
	// değildir: oraya varış mevcut savaş/temas çözümlemesini tetikleyebilir.
	for _, candidate := range s.Armies {
		if candidate == nil || candidate.ID == a.ID || candidate.RegionID != region.ID || candidate.OwnerID == a.OwnerID {
			continue
		}
		return false
	}
	return true
}

func (s *GameState) movementOwnerCanTransit(ownerID, regionOwnerID string) bool {
	if ownerID == "" || regionOwnerID == "" || ownerID == regionOwnerID {
		return true
	}
	if stateSameRealm(s, faction.FactionID(ownerID), faction.FactionID(regionOwnerID)) {
		return true
	}
	relation := s.Relations[faction.RelationKey(faction.FactionID(ownerID), faction.FactionID(regionOwnerID))]
	return relation != nil && relation.Stance == faction.StanceAllied
}

func (s *GameState) navalLandMovementAllowed(fleet *army.Army, target *world.Region) bool {
	if s == nil || fleet == nil || target == nil || !fleet.IsNaval || target.IsSea || !target.CanLandEnter() {
		return false
	}
	hasCargo := len(fleet.EmbarkedUnits) > 0
	targetOwnerID := s.SovereignOwnerID(target)
	if targetOwnerID == "" {
		return hasCargo
	}
	if targetOwnerID == fleet.OwnerID || s.movementOwnerCanTransit(fleet.OwnerID, targetOwnerID) {
		return target.HasPort() || hasCargo
	}
	// Düşman kıyıya çıkarma, mevcut hareket akışında savaş ilanı/temas
	// kararına bırakılan son duraktır; rotanın içinden geçiş noktası değildir.
	return hasCargo
}

type movementRouteQueueItem struct {
	regionID world.RegionID
	cost     int
}

type movementRouteQueue []movementRouteQueueItem

func (q movementRouteQueue) Len() int { return len(q) }

func (q movementRouteQueue) Less(i, j int) bool {
	if q[i].cost != q[j].cost {
		return q[i].cost < q[j].cost
	}
	return q[i].regionID < q[j].regionID
}

func (q movementRouteQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

func (q *movementRouteQueue) Push(value any) {
	*q = append(*q, value.(movementRouteQueueItem))
}

func (q *movementRouteQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[:n-1]
	return item
}
