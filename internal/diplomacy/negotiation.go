package diplomacy

import (
	"fmt"
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	transferKindRegion   = "region"
	transferKindResource = "resource"
	transferKindArmy     = "army"
	transferRelationGain = 10
	maxTransferItems     = 6
)

// QueueTransferOffer, iki taraflı kaynak/bölge/asker pazarlığını teklif
// kuyruğuna ekler. Kalemler çözümleme anına kadar state'te korunur.
func QueueTransferOffer(gs *state.GameState, from, to faction.FactionID, requested, offered []state.DiplomaticTransfer, priority int, reason string) bool {
	if gs == nil || from == "" || to == "" || from == to || len(requested) == 0 && len(offered) == 0 {
		return false
	}
	if gs.Factions[from] == nil || gs.Factions[to] == nil || gs.Factions[from].IsEliminated || gs.Factions[to].IsEliminated {
		return false
	}
	if rel := Relation(gs, from, to); rel != nil && rel.Stance == faction.StanceWar {
		return false
	}
	if !validTransferList(gs, to, requested) || !validTransferList(gs, from, offered) {
		return false
	}
	if len(requested) > maxTransferItems || len(offered) > maxTransferItems || !gs.SpendDiplomacyOfferQuota(from) {
		return false
	}
	gs.DiplomaticOffers = append(gs.DiplomaticOffers, state.DiplomaticOffer{
		FromFactionID:      from,
		ToFactionID:        to,
		Action:             string(ActionProposeTransfer),
		RequestedTransfers: cloneTransfers(requested),
		OfferedTransfers:   cloneTransfers(offered),
		CreatedTurn:        gs.Turn,
		Priority:           priority,
		PriorityReason:     reason,
	})
	return true
}

// ExecuteTransferOffer, oyuncunun AI devletine gönderdiği pazarlığı aynı turda
// değerlendirir; oyuncuya gönderilen teklifler ise normal karar kuyruğunda kalır.
func ExecuteTransferOffer(gs *state.GameState, from, to faction.FactionID, requested, offered []state.DiplomaticTransfer) Result {
	if !QueueTransferOffer(gs, from, to, requested, offered, 50, "karşılıklı diplomatik pazarlık") {
		return Result{Message: "Pazarlık teklifi geçerli değil veya teklif hakkı yok."}
	}
	idx := len(gs.DiplomaticOffers) - 1
	if to == gs.PlayerFactionID {
		return Result{Message: factionLabel(gs, to) + " devletine pazarlık teklifi gönderildi."}
	}
	accepted, chance := AssessTransferOffer(gs, from, to, requested, offered)
	if !accepted && from == gs.PlayerFactionID {
		counterRequested, counterOffered, ok := acceptedTransferCounterOffer(gs, to, from, requested, offered)
		if ok && QueueTransferOffer(gs, to, from, counterRequested, counterOffered, 50, "AI karşı teklifi") {
			gs.DiplomaticOffers = append(gs.DiplomaticOffers[:idx], gs.DiplomaticOffers[idx+1:]...)
			return Result{Message: factionLabel(gs, to) + " pazarlığı kabul etmedi ve karşı teklif gönderdi."}
		}
	}
	result := ResolveOffer(gs, idx, accepted)
	if !accepted && result.Message == "" {
		result.Message = fmt.Sprintf("Pazarlık reddedildi (%d%% kabul olasılığı).", chance)
	}
	return result
}

func acceptedTransferCounterOffer(gs *state.GameState, ai, player faction.FactionID, requested, offered []state.DiplomaticTransfer) ([]state.DiplomaticTransfer, []state.DiplomaticTransfer, bool) {
	counterRequested := cloneTransfers(offered)
	counterOffered := cloneTransfers(requested)
	if len(counterRequested) == 0 || len(counterOffered) == 0 {
		return nil, nil, false
	}
	accepted := func() bool {
		ok, _ := AssessTransferOffer(gs, player, ai, counterOffered, counterRequested)
		return ok
	}
	if accepted() {
		return counterRequested, counterOffered, true
	}

	for i := 0; i < len(counterOffered); i++ {
		transfer := counterOffered[i]
		switch transfer.Kind {
		case transferKindRegion:
			if len(counterOffered) == 1 {
				continue
			}
			counterOffered = append(counterOffered[:i], counterOffered[i+1:]...)
			i--
			if accepted() {
				return counterRequested, counterOffered, true
			}
		case transferKindResource, transferKindArmy:
			if transfer.Amount <= 1 {
				continue
			}
			low, high, best := 1, transfer.Amount-1, 0
			for low <= high {
				mid := low + (high-low)/2
				counterOffered[i].Amount = mid
				if accepted() {
					best = mid
					low = mid + 1
				} else {
					high = mid - 1
				}
			}
			if best > 0 {
				counterOffered[i].Amount = best
				return counterRequested, counterOffered, true
			}
			counterOffered[i].Amount = 1
		}
	}
	if accepted() {
		return counterRequested, counterOffered, true
	}

	for i := range counterRequested {
		transfer := counterRequested[i]
		maxAmount := 0
		switch transfer.Kind {
		case transferKindResource:
			maxAmount = economy.FactionResourceAmount(gs.Factions[player], economy.ResourceKind(transfer.ID))
		case transferKindArmy:
			maxAmount = unitCountForFaction(gs, player, transfer.ID)
		default:
			continue
		}
		if maxAmount <= transfer.Amount {
			continue
		}
		low, high, best := transfer.Amount+1, maxAmount, 0
		for low <= high {
			mid := low + (high-low)/2
			counterRequested[i].Amount = mid
			if accepted() {
				best = mid
				high = mid - 1
			} else {
				low = mid + 1
			}
		}
		if best > 0 {
			counterRequested[i].Amount = best
			return counterRequested, counterOffered, true
		}
		counterRequested[i].Amount = maxAmount
	}
	if accepted() {
		return counterRequested, counterOffered, true
	}
	return nil, nil, false
}

// AssessTransferOffer, AI'nin pazarlığın maddi dengesini ve ilişki puanını
// birlikte değerlendirdiği basit, deterministik kabul hesabıdır.
func AssessTransferOffer(gs *state.GameState, from, to faction.FactionID, requested, offered []state.DiplomaticTransfer) (bool, int) {
	if gs == nil || from == "" || to == "" {
		return false, 0
	}
	requestedValue := TransferListValue(gs, to, requested)
	offeredValue := TransferListValue(gs, from, offered)
	score := RelationScore(gs, to, from)
	chance := 50 + score/2
	if requestedValue == 0 {
		chance += 30
	} else {
		chance += (offeredValue - requestedValue) * 40 / (requestedValue + 100)
	}
	if chance < 0 {
		chance = 0
	}
	if chance > 100 {
		chance = 100
	}
	return chance >= 55, chance
}

func resolveAcceptedTransferOffer(gs *state.GameState, offer state.DiplomaticOffer) Result {
	if !offerPartiesExist(gs, offer) || !validTransferList(gs, offer.ToFactionID, offer.RequestedTransfers) || !validTransferList(gs, offer.FromFactionID, offer.OfferedTransfers) {
		return Result{Message: "Pazarlık artık geçerli değil."}
	}
	if !applyTransferList(gs, offer.ToFactionID, offer.FromFactionID, offer.RequestedTransfers) || !applyTransferList(gs, offer.FromFactionID, offer.ToFactionID, offer.OfferedTransfers) {
		return Result{Message: "Pazarlık uygulanamadı; kaynaklardan biri artık mevcut değil."}
	}
	if len(offer.RequestedTransfers) > 0 {
		AddRelationScore(gs, offer.FromFactionID, offer.ToFactionID, transferRelationGain)
	}
	if len(offer.OfferedTransfers) > 0 {
		AddRelationScore(gs, offer.ToFactionID, offer.FromFactionID, transferRelationGain)
	}
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  factionLabel(gs, offer.ToFactionID) + " ile pazarlık kabul edildi ve kararlaştırılan aktarım uygulandı.",
	}
}

func validTransferList(gs *state.GameState, owner faction.FactionID, transfers []state.DiplomaticTransfer) bool {
	if gs == nil || owner == "" || len(transfers) > maxTransferItems {
		return false
	}
	seen := make(map[string]struct{}, len(transfers))
	for _, transfer := range transfers {
		if transfer.Amount <= 0 || transfer.ID == "" {
			return false
		}
		key := transfer.Kind + ":" + transfer.ID
		if _, ok := seen[key]; ok {
			return false
		}
		seen[key] = struct{}{}
		switch transfer.Kind {
		case transferKindRegion:
			region := gs.Regions[world.RegionID(transfer.ID)]
			if region == nil || region.IsSea || region.IsTerrainArea || region.OwnerID != string(owner) || transfer.Amount != 1 {
				return false
			}
		case transferKindResource:
			if !validResourceKind(transfer.ID) || economy.FactionResourceAmount(gs.Factions[owner], economy.ResourceKind(transfer.ID)) < transfer.Amount {
				return false
			}
		case transferKindArmy:
			if unitCountForFaction(gs, owner, transfer.ID) < transfer.Amount {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func applyTransferList(gs *state.GameState, from, to faction.FactionID, transfers []state.DiplomaticTransfer) bool {
	if !validTransferList(gs, from, transfers) {
		return false
	}
	regionTransferred := false
	for _, transfer := range transfers {
		switch transfer.Kind {
		case transferKindRegion:
			region := gs.Regions[world.RegionID(transfer.ID)]
			region.OwnerID = string(to)
			gs.ClearProductionOrdersForRegion(region.ID)
			regionTransferred = true
		case transferKindResource:
			kind := economy.ResourceKind(transfer.ID)
			economy.AddFactionResource(gs.Factions[from], kind, -transfer.Amount)
			economy.AddFactionResource(gs.Factions[to], kind, transfer.Amount)
		case transferKindArmy:
			if !transferArmyUnits(gs, from, to, transfer.ID, transfer.Amount) {
				return false
			}
		}
	}
	if regionTransferred {
		// Yeni sahipliğe geçen bölgedeki eski sahip ordularını geçerli kara toprağına taşı.
		gs.EvacuateArmiesFromPeaceTerritory([]faction.FactionID{from}, []faction.FactionID{to})
	}
	gs.NormalizeFactionCapitals()
	gs.EvacuateArmiesWithoutLandAccess()
	return true
}

func transferArmyUnits(gs *state.GameState, from, to faction.FactionID, unitTypeID string, amount int) bool {
	if amount <= 0 {
		return false
	}
	destination := gs.LandRegionsOwnedBy(to)
	if len(destination) == 0 {
		return false
	}
	sort.Slice(destination, func(i, j int) bool { return destination[i].ID < destination[j].ID })
	ids := make([]army.ArmyID, 0, len(gs.Armies))
	for id, current := range gs.Armies {
		if current != nil && current.OwnerID == string(from) && !current.IsNaval && current.Commander == nil && current.EmbarkedCommander == nil {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	units := make([]army.Unit, 0, amount)
	for _, id := range ids {
		current := gs.Armies[id]
		remaining := current.Units[:0]
		for _, unit := range current.Units {
			if unit.TypeID == unitTypeID && len(units) < amount {
				units = append(units, unit)
				continue
			}
			remaining = append(remaining, unit)
		}
		current.Units = remaining
		if len(current.Units) == 0 && len(current.EmbarkedUnits) == 0 && current.Commander == nil && current.EmbarkedCommander == nil {
			gs.RemoveArmy(id)
		}
		if len(units) == amount {
			break
		}
	}
	if len(units) != amount {
		return false
	}
	gs.NextArmySeq++
	id := army.ArmyID(fmt.Sprintf("army_%s_transfer_%d", to, gs.NextArmySeq))
	for gs.Armies[id] != nil {
		gs.NextArmySeq++
		id = army.ArmyID(fmt.Sprintf("army_%s_transfer_%d", to, gs.NextArmySeq))
	}
	newArmy := &army.Army{ID: id, OwnerID: string(to), RegionID: destination[0].ID, Units: units, MaxMovePoints: army.DefaultArmyMovePoints, MovePoints: army.DefaultArmyMovePoints}
	if gs.UnitTypes != nil {
		newArmy.MaxMovePoints = newArmy.BaseMovePoints(gs.UnitTypes)
		newArmy.MovePoints = newArmy.MaxMovePoints
	}
	gs.Armies[id] = newArmy
	return true
}

func unitCountForFaction(gs *state.GameState, owner faction.FactionID, unitTypeID string) int {
	count := 0
	for _, current := range gs.Armies {
		if current == nil || current.OwnerID != string(owner) || current.IsNaval || current.Commander != nil || current.EmbarkedCommander != nil {
			continue
		}
		for _, unit := range current.Units {
			if unit.TypeID == unitTypeID {
				count++
			}
		}
	}
	return count
}

// TransferListValue returns the shared diplomatic value estimate for a list
// of resources, regions, or army units owned by the specified faction.
func TransferListValue(gs *state.GameState, owner faction.FactionID, transfers []state.DiplomaticTransfer) int {
	value := 0
	for _, transfer := range transfers {
		switch transfer.Kind {
		case transferKindRegion:
			value += RegionTenTurnEconomicValue(gs, world.RegionID(transfer.ID))
		case transferKindResource:
			kind := economy.ResourceKind(transfer.ID)
			unitValue := 1
			if kind != economy.ResourceGold {
				good := economy.ResourceDefByKind(kind).TradeGood
				if gs.MarketPrices != nil && gs.MarketPrices[good] > 0 {
					unitValue = gs.MarketPrices[good]
				} else if price := gs.BasePrice(good); price > 0 {
					unitValue = price
				}
			}
			value += transfer.Amount * unitValue
		case transferKindArmy:
			unitType := gs.UnitTypes[transfer.ID]
			unitValue := 10
			if unitType != nil {
				unitValue += unitType.Attack + unitType.Defense + unitType.Morale + int(unitType.Tier)*10
			}
			value += transfer.Amount * unitValue
		}
	}
	return value
}

// RegionTenTurnEconomicValue bir bölgenin mevcut efektif altın ve kaynak
// üretimini güncel pazar fiyatlarıyla altın eşdeğerine çevirip on turla çarpar.
// Bölge alım/satım teklifleri, üretim değerinden bağımsız sabit bir etiket
// kullanmak yerine bu alt sınırı gözetir.
func RegionTenTurnEconomicValue(gs *state.GameState, regionID world.RegionID) int {
	if gs == nil || regionID == "" {
		return 0
	}
	region := gs.Regions[regionID]
	if region == nil || region.IsSea || region.IsTerrainArea || gs.SiegeAt(regionID) != nil {
		return 0
	}
	production := gs.RegionProductionSummary(region)
	valuePerTurn := production.Gold
	resources := economy.ResourceCost{
		Grain: production.Grain, Iron: production.Iron, Timber: production.Timber,
		Stone: production.Stone, Spice: production.Spice, Cloth: production.Cloth,
	}
	for _, kind := range economy.CostResourceKinds() {
		if kind == economy.ResourceGold {
			continue
		}
		good := economy.ResourceDefByKind(kind).TradeGood
		price := 0
		if gs.MarketPrices != nil {
			price = gs.MarketPrices[good]
		}
		if price <= 0 {
			price = gs.BasePrice(good)
		}
		if price > 0 {
			valuePerTurn += resources.Amount(kind) * price
		}
	}
	if valuePerTurn <= 0 {
		return 0
	}
	return valuePerTurn * 10
}

func validResourceKind(id string) bool {
	for _, kind := range economy.AllResourceKinds() {
		if string(kind) == id {
			return true
		}
	}
	return false
}

func cloneTransfers(transfers []state.DiplomaticTransfer) []state.DiplomaticTransfer {
	return append([]state.DiplomaticTransfer(nil), transfers...)
}
