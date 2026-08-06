package world

// EnsureRequiredSettlementBuildings, yerleşim tiplerinden ve ulusal başkent
// statüsünden türeyen minimum bina kurallarını uygular. Mevcut binalar korunur;
// yalnızca eksik minimum seviyeler eklenir.
func EnsureRequiredSettlementBuildings(region *Region, isCapitalRegion bool) bool {
	if region == nil || region.IsSea {
		return false
	}

	var required [6]string
	requiredCount := 0
	if region.HasFortressSettlement() {
		required[requiredCount] = "walls"
		requiredCount++
	}
	if hasSettlementType(region, SettlementPort) {
		required[requiredCount] = "port"
		requiredCount++
	}
	if isCapitalRegion {
		for _, buildingID := range [...]string{"barracks", "granary", "temple", "market"} {
			required[requiredCount] = buildingID
			requiredCount++
		}
	}

	changed := false
	for i := 0; i < requiredCount; i++ {
		buildingID := required[i]
		if region.HasBuilding(buildingID) {
			continue
		}
		region.Buildings = append(region.Buildings, buildingID)
		changed = true
	}
	return changed
}

func hasSettlementType(region *Region, settlementType SettlementType) bool {
	for _, settlement := range region.Settlements {
		if settlement.Type == settlementType {
			return true
		}
	}
	return false
}
