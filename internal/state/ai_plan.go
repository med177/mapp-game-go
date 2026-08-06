package state

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// AIObjectiveKind kalıcı stratejik niyetin genel türünü belirtir. Bu değerler
// kararın kendisidir; güç haritası, yol ve lojistik gibi türetilmiş hesaplar
// save state'e yazılmaz.
type AIObjectiveKind string

const (
	AIObjectiveExpand      AIObjectiveKind = "expand"
	AIObjectiveDefend      AIObjectiveKind = "defend"
	AIObjectiveConsolidate AIObjectiveKind = "consolidate"
)

// AIPlanState bir fraksiyonun save/load arasında korunması gereken stratejik
// niyetini taşır. TargetRegionIDs öncelik sırasındadır. RallyRegionID ile
// RallyDeadlineTurn çok turlu hazırlığın save/load sırasında sıfırlanmasını
// engeller; runtime rol, güç ve path cache'leri her AI turunda yeniden üretilir.
type AIPlanState struct {
	ObjectiveID        string            `json:"objective_id"`
	Kind               AIObjectiveKind   `json:"kind"`
	TargetFactionID    faction.FactionID `json:"target_faction_id,omitempty"`
	TargetRegionIDs    []world.RegionID  `json:"target_region_ids,omitempty"`
	StartedTurn        int               `json:"started_turn"`
	ReassessTurn       int               `json:"reassess_turn"`
	RallyRegionID      world.RegionID    `json:"rally_region_id,omitempty"`
	RallyDeadlineTurn  int               `json:"rally_deadline_turn,omitempty"`
	Commitment         int               `json:"commitment"`
	AllowVassalization bool              `json:"allow_vassalization,omitempty"`
	Reason             string            `json:"reason,omitempty"`
}
