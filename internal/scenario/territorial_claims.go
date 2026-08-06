package scenario

import (
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

const (
	defaultCoreClaimValue   = 100
	defaultTargetClaimValue = 50
)

// ApplyInitialTerritorialClaims senaryo base state'i kurulurken başlangıç
// sahipliğini core'a, AI stratejisindeki bölge hedeflerini de claim'e çevirir.
// Bu fonksiyon yalnızca base state yüklemesinde çağrılmalıdır; fetih sonrası
// güncel sahiplikler core/claim listesini geriye dönük değiştirmemelidir.
func ApplyInitialTerritorialClaims(
	regions map[world.RegionID]*world.Region,
	factions map[faction.FactionID]*faction.Faction,
	strategies map[string]AIFactionStrategy,
) {
	if len(regions) == 0 || len(factions) == 0 {
		return
	}

	initialRegionsByOwner := make(map[faction.FactionID][]world.RegionID)
	for regionID, region := range regions {
		if region == nil || region.IsSea || region.OwnerID == "" {
			continue
		}
		ownerID := faction.FactionID(region.OwnerID)
		if factions[ownerID] == nil {
			continue
		}
		initialRegionsByOwner[ownerID] = append(initialRegionsByOwner[ownerID], regionID)
	}
	for ownerID := range initialRegionsByOwner {
		sort.Slice(initialRegionsByOwner[ownerID], func(i, j int) bool {
			return initialRegionsByOwner[ownerID][i] < initialRegionsByOwner[ownerID][j]
		})
	}

	for ownerID, fx := range factions {
		if strategy, ok := strategies[string(ownerID)]; ok && strategy.ExpansionTargets != nil {
			fx.AIExpansionTargets = make([]faction.FactionID, 0, len(strategy.ExpansionTargets))
			for _, targetID := range strategy.ExpansionTargets {
				if targetID != "" && factions[faction.FactionID(targetID)] != nil && faction.FactionID(targetID) != ownerID {
					fx.AIExpansionTargets = append(fx.AIExpansionTargets, faction.FactionID(targetID))
				}
			}
		}
		claims := make(map[world.RegionID]faction.TerritorialClaim, len(fx.TerritorialClaims))
		for _, claim := range fx.TerritorialClaims {
			regionID := world.RegionID(claim.RegionID)
			if regions[regionID] == nil || regions[regionID].IsSea {
				continue
			}
			if claim.Value <= 0 {
				claim.Value = defaultTargetClaimValue
			}
			if claim.Value > 100 {
				claim.Value = 100
			}
			claims[regionID] = claim
		}
		if strategy, ok := strategies[string(ownerID)]; ok {
			for _, claim := range strategy.TerritorialClaims {
				regionID := world.RegionID(claim.RegionID)
				if region := regions[regionID]; region != nil && !region.IsSea {
					mergeTerritorialClaim(claims, regionID, claim.Value, false)
				}
			}
			for _, objective := range strategy.Objectives {
				for _, claim := range objective.TerritorialClaims {
					regionID := world.RegionID(claim.RegionID)
					if region := regions[regionID]; region != nil && !region.IsSea {
						mergeTerritorialClaim(claims, regionID, claim.Value, false)
					}
				}
			}
		}

		// Every region owned at scenario start is an unconditional core. An
		// explicit claim on the same region is retained as an override, but
		// can never turn an initial core back into a non-core claim.
		for _, regionID := range initialRegionsByOwner[ownerID] {
			mergeTerritorialClaim(claims, regionID, defaultCoreClaimValue, true)
		}

		claimIDs := make([]world.RegionID, 0, len(claims))
		for regionID := range claims {
			claimIDs = append(claimIDs, regionID)
		}
		sort.Slice(claimIDs, func(i, j int) bool { return claimIDs[i] < claimIDs[j] })
		fx.TerritorialClaims = fx.TerritorialClaims[:0]
		for _, regionID := range claimIDs {
			fx.TerritorialClaims = append(fx.TerritorialClaims, claims[regionID])
		}
	}
}

func mergeTerritorialClaim(
	claims map[world.RegionID]faction.TerritorialClaim,
	regionID world.RegionID,
	value int,
	core bool,
) {
	if regionID == "" {
		return
	}
	if value <= 0 {
		value = defaultTargetClaimValue
	}
	claim := claims[regionID]
	claim.RegionID = string(regionID)
	if value > claim.Value {
		claim.Value = value
	}
	if core {
		claim.Core = true
		claim.Value = defaultCoreClaimValue
	}
	claims[regionID] = claim
}
