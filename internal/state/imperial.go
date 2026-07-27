package state

import (
	"encoding/json"
	"fmt"
	"os"

	"mapp-game-go/internal/faction"
)

// ImperialMemberStatus bir imparatorluk üyesinin siyasi sınıfını belirtir.
// Vassal üyeler ayrıca Faction.OverlordID üzerinden gerçek realm kuralını korur;
// bu alan, bağımsız imparatorluk prenslerini SameRealm'e sokmadan temsil eder.
type ImperialMemberStatus string

const (
	ImperialMemberElector  ImperialMemberStatus = "elector"
	ImperialMemberPrince   ImperialMemberStatus = "prince"
	ImperialMemberFreeCity ImperialMemberStatus = "free_city"
	ImperialMemberOrder    ImperialMemberStatus = "order"
	ImperialMemberVassal   ImperialMemberStatus = "vassal"
)

// ImperialMember bir fraksiyonun imparatorluk içindeki politik konumunu ve
// çağrılara vereceği tepkiyi etkileyen save-backed değerleri taşır.
type ImperialMember struct {
	FactionID          faction.FactionID    `json:"faction_id"`
	Status             ImperialMemberStatus `json:"status"`
	Loyalty            int                  `json:"loyalty"`
	Autonomy           int                  `json:"autonomy"`
	ElectorWeight      int                  `json:"elector_weight,omitempty"`
	MilitaryCommitment int                  `json:"military_commitment"`
}

// ImperialWarCall son çağrının savaş bağlamını ve üyelerin katılımını save'de
// izlemek için kullanılır. İlişki kayıtları gerçek savaşa girenleri taşımaya
// devam eder; bu kayıt yalnızca çağrının politik kaynağını korur.
type ImperialWarCall struct {
	CallerID    faction.FactionID `json:"caller_id"`
	EnemyID     faction.FactionID `json:"enemy_id"`
	Reason      string            `json:"reason"`
	StartedTurn int               `json:"started_turn"`
}

// ImperialDecisionKind oyuncunun çözmesi gereken imparatorluk kararını belirtir.
type ImperialDecisionKind string

const (
	ImperialDecisionDiet     ImperialDecisionKind = "diet"
	ImperialDecisionElection ImperialDecisionKind = "election"
)

// ImperialPendingDecision imparatorluk kararının tur geçişinde kaybolmaması
// için save-backed olarak tutulur.
type ImperialPendingDecision struct {
	Kind        ImperialDecisionKind `json:"kind"`
	CreatedTurn int                  `json:"created_turn"`
}

// ImperialState HRE gibi seçilmiş bir imparatorluk kurumunun, onu oluşturan
// bağımsız üyelerden ayrı kampanya state'idir.
type ImperialState struct {
	EmpireID        faction.FactionID                     `json:"empire_id"`
	EmperorID       faction.FactionID                     `json:"emperor_id"`
	Authority       int                                   `json:"authority"`
	NextDietTurn    int                                   `json:"next_diet_turn,omitempty"`
	ElectionDueTurn int                                   `json:"election_due_turn,omitempty"`
	Members         map[faction.FactionID]*ImperialMember `json:"members"`
	LastWarCall     *ImperialWarCall                      `json:"last_war_call,omitempty"`
	PendingDecision *ImperialPendingDecision              `json:"pending_decision,omitempty"`
}

// Clone deep-copies the mutable imperial campaign state for save payloads.
func (s *ImperialState) Clone() *ImperialState {
	if s == nil {
		return nil
	}
	out := *s
	if s.Members != nil {
		out.Members = make(map[faction.FactionID]*ImperialMember, len(s.Members))
		for id, member := range s.Members {
			if member == nil {
				continue
			}
			copy := *member
			out.Members[id] = &copy
		}
	}
	if s.LastWarCall != nil {
		call := *s.LastWarCall
		out.LastWarCall = &call
	}
	if s.PendingDecision != nil {
		decision := *s.PendingDecision
		out.PendingDecision = &decision
	}
	return &out
}

// Clamp normalizes values loaded from older or hand-edited scenario/save data.
func (s *ImperialState) Clamp() {
	if s == nil {
		return
	}
	if s.Authority < 0 {
		s.Authority = 0
	}
	if s.Authority > 100 {
		s.Authority = 100
	}
	for _, member := range s.Members {
		if member == nil {
			continue
		}
		member.Loyalty = clampImperialValue(member.Loyalty)
		member.Autonomy = clampImperialValue(member.Autonomy)
		member.MilitaryCommitment = clampImperialValue(member.MilitaryCommitment)
		if member.ElectorWeight < 0 {
			member.ElectorWeight = 0
		}
	}
}

func clampImperialValue(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

type imperialFile struct {
	EmpireID        faction.FactionID `json:"empire_id"`
	EmperorID       faction.FactionID `json:"emperor_id"`
	Authority       int               `json:"authority"`
	NextDietTurn    int               `json:"next_diet_turn,omitempty"`
	ElectionDueTurn int               `json:"election_due_turn,omitempty"`
	Members         []ImperialMember  `json:"members"`
}

// LoadImperialState senaryo data/imperial.json dosyasını yükler. Dosya
// bulunmuyorsa senaryo imparatorluk sistemi kullanmıyor kabul edilir.
func LoadImperialState(path string, factions map[faction.FactionID]*faction.Faction) (*ImperialState, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("imparatorluk verisi okunamadı (%s): %w", path, err)
	}
	var file imperialFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("imparatorluk verisi parse hatası (%s): %w", path, err)
	}
	if file.EmpireID == "" {
		return nil, fmt.Errorf("imparatorluk verisinde empire_id eksik: %s", path)
	}
	if factions != nil && factions[file.EmpireID] == nil {
		return nil, fmt.Errorf("imparatorluk kök fraksiyonu bulunamadı: %s", file.EmpireID)
	}
	if file.EmperorID == "" {
		file.EmperorID = file.EmpireID
	}
	state := &ImperialState{
		EmpireID:        file.EmpireID,
		EmperorID:       file.EmperorID,
		Authority:       file.Authority,
		NextDietTurn:    file.NextDietTurn,
		ElectionDueTurn: file.ElectionDueTurn,
		Members:         make(map[faction.FactionID]*ImperialMember, len(file.Members)),
	}
	for _, configured := range file.Members {
		member := configured
		if member.FactionID == "" || (factions != nil && factions[member.FactionID] == nil) {
			continue
		}
		if member.Status == "" {
			member.Status = ImperialMemberPrince
		}
		state.Members[member.FactionID] = &member
	}
	state.Clamp()
	return state, nil
}
