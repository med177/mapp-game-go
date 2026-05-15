package diplomacy

import (
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

type Action string

const (
	ActionDeclareWar      Action = "declare_war"
	ActionProposePeace    Action = "propose_peace"
	ActionProposeAlliance Action = "propose_alliance"
	ActionProposeTrade    Action = "propose_trade"
)

type Result struct {
	Accepted bool
	Applied  bool
	Message  string
}

func Execute(gs *state.GameState, actor, target faction.FactionID, action Action) Result {
	if gs == nil {
		return Result{Message: "Diplomasi durumu yok."}
	}
	if actor == "" || target == "" || actor == target {
		return Result{Message: "Geçersiz diplomasi hedefi."}
	}
	actorFaction := gs.Factions[actor]
	targetFaction := gs.Factions[target]
	if actorFaction == nil || targetFaction == nil {
		return Result{Message: "Fraksiyon bulunamadı."}
	}
	if actorFaction.IsEliminated || targetFaction.IsEliminated {
		return Result{Message: "Elenmiş fraksiyonlarla diplomasi kurulamaz."}
	}

	rel := EnsureRelation(gs, actor, target)
	switch action {
	case ActionDeclareWar:
		if rel.Stance == faction.StanceWar {
			return Result{Message: factionLabel(gs, target) + " ile zaten savaş halindesiniz."}
		}
		rel.Stance = faction.StanceWar
		rel.Score = -80
		removeTradeRoutesBetween(gs, actor, target)
		return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " ile savaş başladı."}

	case ActionProposePeace:
		if rel.Stance != faction.StanceWar {
			return Result{Message: "Barış teklifi yalnızca savaş halindeyken yapılabilir."}
		}
		if !acceptPeace(gs, rel, actor, target) {
			return Result{Message: factionLabel(gs, target) + " barışı reddetti."}
		}
		rel.Stance = faction.StancePeace
		rel.Score = -20
		removeTradeRoutesBetween(gs, actor, target)
		return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " barışı kabul etti."}

	case ActionProposeTrade:
		if rel.Stance == faction.StanceWar {
			return Result{Message: "Savaş halindeyken ticaret yapılamaz."}
		}
		if rel.Stance == faction.StanceTrade {
			return Result{Message: "Zaten aktif bir ticaret anlaşması var."}
		}
		if rel.Stance == faction.StanceAllied {
			return Result{Message: "Bu fraksiyon zaten müttefik statüsünde."}
		}
		if !acceptTrade(gs, rel, actor, target) {
			return Result{Message: factionLabel(gs, target) + " ticaret teklifini reddetti."}
		}
		rel.Stance = faction.StanceTrade
		rel.Score = clamp(rel.Score+15, -100, 100)
		ensureTradeRoutesBetween(gs, actor, target)
		return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " ile ticaret anlaşması imzalandı."}

	case ActionProposeAlliance:
		if rel.Stance == faction.StanceWar {
			return Result{Message: "Savaş halindeyken ittifak kurulamaz."}
		}
		if rel.Stance == faction.StanceAllied {
			return Result{Message: "Zaten müttefiksiniz."}
		}
		if !acceptAlliance(gs, rel, actor, target) {
			return Result{Message: factionLabel(gs, target) + " ittifak teklifini reddetti."}
		}
		rel.Stance = faction.StanceAllied
		rel.Score = clamp(rel.Score+20, -100, 100)
		return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " ile ittifak kuruldu."}
	}

	return Result{Message: "Bilinmeyen diplomasi aksiyonu."}
}

func EnsureRelation(gs *state.GameState, a, b faction.FactionID) *faction.Relation {
	if gs.Relations == nil {
		gs.Relations = make(map[string]*faction.Relation)
	}
	key := faction.RelationKey(a, b)
	if rel := gs.Relations[key]; rel != nil {
		return rel
	}
	rel := &faction.Relation{
		FactionA: a,
		FactionB: b,
		Score:    0,
		Stance:   faction.StancePeace,
	}
	gs.Relations[key] = rel
	return rel
}

func Relation(gs *state.GameState, a, b faction.FactionID) *faction.Relation {
	if gs == nil {
		return nil
	}
	return gs.Relations[faction.RelationKey(a, b)]
}

func IsWar(gs *state.GameState, a, b faction.FactionID) bool {
	rel := Relation(gs, a, b)
	return rel != nil && rel.Stance == faction.StanceWar
}

func ApplyRelationDecay(gs *state.GameState) {
	for _, rel := range gs.Relations {
		if rel == nil {
			continue
		}
		switch rel.Stance {
		case faction.StanceWar:
			rel.Score = clamp(rel.Score-1, -100, 100)
		case faction.StancePeace:
			if rel.Score < 0 {
				rel.Score++
			}
		case faction.StanceTrade:
			if rel.Score < 30 {
				rel.Score++
			}
		case faction.StanceAllied:
			if rel.Score < 50 {
				rel.Score++
			}
		}
	}
}

func MilitaryPower(gs *state.GameState, fid faction.FactionID) int {
	total := 0
	for _, a := range gs.Armies {
		if a == nil || a.OwnerID != string(fid) {
			continue
		}
		if gs.UnitTypes != nil {
			total += a.TotalStrength(gs.UnitTypes)
			continue
		}
		total += len(a.Units) * 10
	}
	return total
}

func HasCommonEnemy(gs *state.GameState, a, b faction.FactionID) bool {
	for otherID := range gs.Factions {
		if otherID == a || otherID == b {
			continue
		}
		if IsWar(gs, a, otherID) && IsWar(gs, b, otherID) {
			return true
		}
	}
	return false
}

func HasDirectThreat(gs *state.GameState, a, b faction.FactionID) bool {
	if !sharesBorder(gs, a, b) {
		return false
	}
	powerA := MilitaryPower(gs, a)
	powerB := MilitaryPower(gs, b)
	if powerA == 0 || powerB == 0 {
		return powerA != powerB
	}
	if powerA > powerB*12/10 || powerB > powerA*12/10 {
		return true
	}
	return frontierArmyCount(gs, a, b) > frontierArmyCount(gs, b, a)+1 ||
		frontierArmyCount(gs, b, a) > frontierArmyCount(gs, a, b)+1
}

func acceptPeace(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) bool {
	warPressure := 0
	if rel.Score < -80 {
		warPressure = -rel.Score - 80
	}

	actorPower := MilitaryPower(gs, actor)
	targetPower := MilitaryPower(gs, target)
	strengthPressure := 0
	if actorPower > targetPower {
		strengthPressure += min(25, (actorPower-targetPower)/8)
	} else if targetPower > actorPower {
		strengthPressure -= min(10, (targetPower-actorPower)/12)
	}

	actorRegions := len(gs.RegionsOwnedBy(actor))
	targetRegions := len(gs.RegionsOwnedBy(target))
	if actorRegions > targetRegions {
		strengthPressure += min(20, (actorRegions-targetRegions)*4)
	}

	return warPressure+strengthPressure+economicStress(gs, target) >= 18
}

func acceptTrade(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) bool {
	if rel.Score < -35 {
		return false
	}
	return landRegionCount(gs, actor) > 0 && landRegionCount(gs, target) > 0
}

func acceptAlliance(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) bool {
	if rel.Score < 20 {
		return false
	}
	if HasDirectThreat(gs, actor, target) {
		return false
	}
	return true
}

func ensureTradeRoutesBetween(gs *state.GameState, a, b faction.FactionID) {
	removeTradeRoutesBetween(gs, a, b)
	routeAB := buildTradeRoute(gs, a, b)
	routeBA := buildTradeRoute(gs, b, a)
	gs.TradeRoutes = append(gs.TradeRoutes, routeAB, routeBA)
}

func removeTradeRoutesBetween(gs *state.GameState, a, b faction.FactionID) {
	if len(gs.TradeRoutes) == 0 {
		return
	}
	filtered := gs.TradeRoutes[:0]
	aStr := string(a)
	bStr := string(b)
	for _, route := range gs.TradeRoutes {
		if route == nil {
			continue
		}
		if (route.FromFactionID == aStr && route.ToFactionID == bStr) ||
			(route.FromFactionID == bStr && route.ToFactionID == aStr) {
			continue
		}
		filtered = append(filtered, route)
	}
	gs.TradeRoutes = filtered
}

func buildTradeRoute(gs *state.GameState, from, to faction.FactionID) *economy.TradeRoute {
	good := chooseExportGood(gs, from)
	return &economy.TradeRoute{
		FromFactionID: string(from),
		ToFactionID:   string(to),
		Good:          good,
		AmountPerTurn: tradeAmount(gs, from, to),
		GoldPerUnit:   economy.BaseGoldValue[good],
	}
}

func chooseExportGood(gs *state.GameState, fid faction.FactionID) economy.GoodType {
	f := gs.Factions[fid]
	if f == nil {
		return economy.GoodGrain
	}
	type goodStock struct {
		good  economy.GoodType
		stock int
	}
	options := []goodStock{
		{economy.GoodSpice, f.Spice},
		{economy.GoodCloth, f.Cloth},
		{economy.GoodIron, f.Iron},
		{economy.GoodTimber, f.Timber},
		{economy.GoodGrain, f.Grain},
	}
	best := options[len(options)-1].good
	bestScore := -1
	for _, option := range options {
		score := option.stock * economy.BaseGoldValue[option.good]
		if score > bestScore {
			bestScore = score
			best = option.good
		}
	}
	return best
}

func tradeAmount(gs *state.GameState, a, b faction.FactionID) int {
	capA := totalTradeCapacity(gs, a)
	capB := totalTradeCapacity(gs, b)
	capacity := min(capA, capB)
	switch {
	case capacity <= 0:
		return 1
	case capacity >= 8:
		return 4
	case capacity >= 5:
		return 3
	case capacity >= 2:
		return 2
	default:
		return 1
	}
}

func totalTradeCapacity(gs *state.GameState, fid faction.FactionID) int {
	total := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		total += region.TradeCapacity
	}
	return total
}

func landRegionCount(gs *state.GameState, fid faction.FactionID) int {
	count := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		count++
	}
	return count
}

func sharesBorder(gs *state.GameState, a, b faction.FactionID) bool {
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(a) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && !neighbor.IsSea && neighbor.OwnerID == string(b) {
				return true
			}
		}
	}
	return false
}

func frontierArmyCount(gs *state.GameState, owner, against faction.FactionID) int {
	count := 0
	for _, armyRef := range gs.Armies {
		if armyRef == nil || armyRef.OwnerID != string(owner) || armyRef.IsNaval {
			continue
		}
		region := gs.Regions[armyRef.RegionID]
		if region == nil {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && neighbor.OwnerID == string(against) {
				count++
				break
			}
		}
	}
	return count
}

func economicStress(gs *state.GameState, fid faction.FactionID) int {
	f := gs.Factions[fid]
	if f == nil {
		return 0
	}
	stress := 0
	if f.Gold < 80 {
		stress += 8
	}
	if f.Grain < 40 {
		stress += 8
	}
	if landRegionCount(gs, fid) <= 2 {
		stress += 6
	}
	return stress
}

func factionLabel(gs *state.GameState, fid faction.FactionID) string {
	if f := gs.Factions[fid]; f != nil {
		if f.NameTR != "" {
			return f.NameTR
		}
		if f.Name != "" {
			return f.Name
		}
	}
	return string(fid)
}

func clamp(v, minValue, maxValue int) int {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
