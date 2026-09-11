package scenario

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// PoliticalTransformation birden fazla faction'ın tarihsel bir siyasi
// dönüşümle birleşmesi veya ayrışması için veri sözleşmesidir. Type alanı
// birleşme, iç savaşla bölünme veya mevcut siyasi yapının dağılması gibi
// çözümleme davranışını seçer.
type PoliticalTransformation struct {
	ID                string                 `json:"id"`
	Type              string                 `json:"type"`
	Members           []string               `json:"members"`
	ResultFaction     string                 `json:"result_faction"`
	Trigger           PoliticalTrigger       `json:"trigger"`
	Transfer          PoliticalTransferRules `json:"transfer"`
	RegionAssignments map[string]string      `json:"region_assignments,omitempty"`
	DescriptionTR     string                 `json:"description_tr,omitempty"`
}

// PoliticalTrigger dönüşümün hangi tarih veya event ile etkinleşeceğini
// tanımlar. EventID verilirse tarih yalnızca yardımcı metadata olarak kalır.
type PoliticalTrigger struct {
	EventID string `json:"event_id,omitempty"`
	Year    int    `json:"year,omitempty"`
	Month   int    `json:"month,omitempty"`
}

// PoliticalTransferRules dönüşüm sırasında hangi state parçalarının
// birleştirileceğini belirtir. Bu alanlar sonraki çözümleme adımlarının ortak
// davranış sözleşmesidir.
type PoliticalTransferRules struct {
	Regions   bool   `json:"regions"`
	Armies    bool   `json:"armies"`
	Navies    bool   `json:"navies"`
	Resources bool   `json:"resources"`
	Relations string `json:"relations,omitempty"`
}

// LoadPoliticalTransformations senaryonun siyasi dönüşüm kayıtlarını yükler.
// Dosyanın bulunmaması, henüz dönüşüm tanımlamayan eski senaryolar için
// uyumluluk amacıyla boş kayıt olarak kabul edilir.
func LoadPoliticalTransformations(path string) ([]PoliticalTransformation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("siyasi dönüşümler okunamadı: %w", err)
	}
	var transformations []PoliticalTransformation
	if err := json.Unmarshal(data, &transformations); err != nil {
		return nil, fmt.Errorf("siyasi dönüşümler parse edilemedi: %w", err)
	}
	for i, transformation := range transformations {
		if transformation.ID == "" {
			return nil, fmt.Errorf("siyasi dönüşüm %d: id boş", i)
		}
		if transformation.Type == "" {
			return nil, fmt.Errorf("siyasi dönüşüm %s: type boş", transformation.ID)
		}
		if len(transformation.Members) < 2 {
			return nil, fmt.Errorf("siyasi dönüşüm %s: en az iki member gerekli", transformation.ID)
		}
		if transformation.ResultFaction == "" {
			return nil, fmt.Errorf("siyasi dönüşüm %s: result_faction boş", transformation.ID)
		}
	}
	return transformations, nil
}
