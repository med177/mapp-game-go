package world

// RequiredInfrastructureBuildingIDs, yerleşim tipleri ve ulusal başkent
// statüsünden türeyen minimum bina setini döndürür. Dizi kullanımı, bu
// kuralların bakım hesabında her tur gereksiz allocation üretmemesini sağlar.
// Count kadar eleman geçerlidir.
func RequiredInfrastructureBuildingIDs(region *Region, isCapitalRegion bool) ([6]string, int) {
	var required [6]string
	if region == nil || region.IsSea {
		return required, 0
	}

	count := 0
	if region.HasFortressSettlement() {
		required[count] = "walls"
		count++
	}
	if hasSettlementType(region, SettlementPort) {
		required[count] = "port"
		count++
	}
	if isCapitalRegion && !region.IsMinorRegion {
		for _, buildingID := range [...]string{"barracks", "granary", "temple", "market"} {
			required[count] = buildingID
			count++
		}
	}
	return required, count
}

// EnsureRequiredSettlementBuildings, yerleşim tiplerinden ve ulusal başkent
// statüsünden türeyen minimum bina kurallarını uygular. Mevcut binalar korunur;
// yalnızca eksik minimum seviyeler eklenir.
func EnsureRequiredSettlementBuildings(region *Region, isCapitalRegion bool) bool {
	required, requiredCount := RequiredInfrastructureBuildingIDs(region, isCapitalRegion)

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

// EnsureSuccessorFoundingBuildings, tarihsel event ile yeniden kurulan tek
// bölgeli devletin kuruluş bölgesine gerekli ilk askerî ve tahıl altyapısını
// ekler. Mevcut binaları korur ve eksik olanların yalnızca birinci seviyesini
// temsil eden tek kaydı ekler.
func EnsureSuccessorFoundingBuildings(region *Region) bool {
	if region == nil || region.IsSea || region.IsMinorRegion {
		return false
	}
	changed := false
	for _, buildingID := range [...]string{"barracks", "granary"} {
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
