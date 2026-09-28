package diplomacy

import (
	"fmt"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// PrivilegeRevocationRelationPenalty, egemen devletin imtiyazı tek taraflı
// kaldırmasının kullanım sahibiyle ilişkiye verdiği puan cezasıdır.
const PrivilegeRevocationRelationPenalty = 10

// RevokeMinorPrivilege, egemen devletin kendi imtiyazlı minor bölgesindeki
// kullanım hakkını kaldırır. Kullanım sahibiyle savaş başlatmaz; yalnızca
// ilişki puanını düşürür.
func RevokeMinorPrivilege(gs *state.GameState, actor faction.FactionID, rid world.RegionID) Result {
	if gs == nil || actor == "" || rid == "" {
		return Result{Message: "İmtiyaz kaldırma isteği geçersiz."}
	}
	region := gs.Regions[rid]
	if region == nil || region.IsSea || !region.IsMinorRegion || !region.IsPrivileged {
		return Result{Message: "Bu bölgenin aktif imtiyazı yok."}
	}
	if gs.SovereignOwnerID(region) != string(actor) {
		return Result{Message: "İmtiyazı yalnızca bölgenin egemen sahibi kaldırabilir."}
	}
	operator := faction.FactionID(region.OwnerID)
	if operator == "" || operator == actor {
		return Result{Message: "Bu bölgenin ayrı bir imtiyaz sahibi yok."}
	}

	region.IsPrivileged = false
	ForceRelation(gs, actor, operator, "", -PrivilegeRevocationRelationPenalty)
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  fmt.Sprintf("%s bölgesindeki imtiyaz kaldırıldı. %s ile ilişki -%d.", region.NameTR, factionLabel(gs, operator), PrivilegeRevocationRelationPenalty),
	}
}
