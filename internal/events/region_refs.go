package events

import "mapp-game-go/internal/world"

// RenameRegionIDReferences, senaryo editöründe bir bölge ID'si değiştiğinde
// event tanımlarındaki tüm bölge referanslarını yeni ID'ye taşır.
func RenameRegionIDReferences(eventList []*Event, oldID, newID world.RegionID) {
	if oldID == "" || newID == "" || oldID == newID {
		return
	}
	for _, event := range eventList {
		if event == nil {
			continue
		}
		renameEventRegionReferences(event, oldID, newID)
	}
}

func renameEventRegionReferences(event *Event, oldID, newID world.RegionID) {
	renameRegionIDSlice(event.RequiresOwnedRegions, oldID, newID)
	renameRegionIDSlice(event.RequiresOwnedRegionsAny, oldID, newID)
	renameRegionIDSlice(event.RequiresUnownedRegions, oldID, newID)
	renameSuccessorRevival(event.SuccessorRevival, oldID, newID)
	for i := range event.SuccessorRevivals {
		renameSuccessorRevival(&event.SuccessorRevivals[i], oldID, newID)
	}
	for i := range event.TradeNetworkModifiers {
		renameStringRegionIDSlice(event.TradeNetworkModifiers[i].RegionIDs, oldID, newID)
	}
	for i := range event.ArmyDefections {
		renameRegionIDSlice(event.ArmyDefections[i].SourceRegionIDs, oldID, newID)
		if event.ArmyDefections[i].DestinationRegionID == oldID {
			event.ArmyDefections[i].DestinationRegionID = newID
		}
	}
	if event.DynasticSettlement != nil {
		renameRegionIDSlice(event.DynasticSettlement.RegionIDs, oldID, newID)
	}
	for i := range event.Choices {
		renameEffectRegionReferences(&event.Choices[i].Effect, oldID, newID)
	}
}

func renameEffectRegionReferences(effect *Effect, oldID, newID world.RegionID) {
	renameSuccessorRevival(effect.SuccessorRevival, oldID, newID)
	for i := range effect.SuccessorRevivals {
		renameSuccessorRevival(&effect.SuccessorRevivals[i], oldID, newID)
	}
	for i := range effect.TradeNetworkModifiers {
		renameStringRegionIDSlice(effect.TradeNetworkModifiers[i].RegionIDs, oldID, newID)
	}
	for i := range effect.ArmyDefections {
		renameRegionIDSlice(effect.ArmyDefections[i].SourceRegionIDs, oldID, newID)
		if effect.ArmyDefections[i].DestinationRegionID == oldID {
			effect.ArmyDefections[i].DestinationRegionID = newID
		}
	}
	if effect.DynasticSettlement != nil {
		renameRegionIDSlice(effect.DynasticSettlement.RegionIDs, oldID, newID)
	}
}

func renameSuccessorRevival(effect *SuccessorRevivalEffect, oldID, newID world.RegionID) {
	if effect == nil {
		return
	}
	for i := range effect.Regions {
		if effect.Regions[i] == string(oldID) {
			effect.Regions[i] = string(newID)
		}
	}
	if effect.RegionID == string(oldID) {
		effect.RegionID = string(newID)
	}
}

func renameRegionIDSlice(ids []world.RegionID, oldID, newID world.RegionID) {
	for i := range ids {
		if ids[i] == oldID {
			ids[i] = newID
		}
	}
}

func renameStringRegionIDSlice(ids []string, oldID, newID world.RegionID) {
	for i := range ids {
		if ids[i] == string(oldID) {
			ids[i] = string(newID)
		}
	}
}
