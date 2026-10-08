package world

import "testing"

func TestEnsureRequiredSettlementBuildings(t *testing.T) {
	tests := []struct {
		name          string
		region        *Region
		wantBuildings []string
		wantChanged   bool
	}{
		{
			name:          "fortress adds walls",
			region:        &Region{ID: "castle_region", Settlements: []Settlement{{Type: SettlementFortress}}},
			wantBuildings: []string{"walls"},
			wantChanged:   true,
		},
		{
			name:          "port adds port building",
			region:        &Region{ID: "port_region", Settlements: []Settlement{{Type: SettlementPort}}},
			wantBuildings: []string{"port"},
			wantChanged:   true,
		},
		{
			name: "existing buildings remain unchanged",
			region: &Region{
				ID:          "valid_region",
				Buildings:   []string{"walls", "port"},
				Settlements: []Settlement{{Type: SettlementFortress}, {Type: SettlementPort}},
			},
			wantBuildings: []string{"walls", "port"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if changed := EnsureRequiredSettlementBuildings(tt.region, false); changed != tt.wantChanged {
				t.Fatalf("EnsureRequiredSettlementBuildings() changed = %v, want %v", changed, tt.wantChanged)
			}
			if len(tt.region.Buildings) != len(tt.wantBuildings) {
				t.Fatalf("bina sayısı = %v, want %v", tt.region.Buildings, tt.wantBuildings)
			}
			for i := range tt.region.Buildings {
				if tt.region.Buildings[i] != tt.wantBuildings[i] {
					t.Fatalf("binalar = %v, want %v", tt.region.Buildings, tt.wantBuildings)
				}
			}
		})
	}
}
