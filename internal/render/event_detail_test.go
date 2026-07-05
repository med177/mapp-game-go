package render

import "testing"

func TestEventDetailSourceLabel(t *testing.T) {
	tests := []struct {
		name  string
		title string
		body  []string
		want  string
	}{
		{
			name:  "historical event",
			title: "[OLAY] Büyük Salgın",
			want:  "Kaynak: Olay kaydı",
		},
		{
			name:  "historical choice",
			title: "[KARAR] Büyük Salgın: Karantina",
			want:  "Kaynak: Karar kaydı",
		},
		{
			name:  "active region trace",
			title: "Veba",
			body:  []string{"", "Kaynak: Harita izi", "Bölge: Test"},
			want:  "Kaynak: Harita izi",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := eventDetailSourceLabel(tc.title, tc.body); got != tc.want {
				t.Fatalf("beklenen kaynak etiketi %q, got=%q", tc.want, got)
			}
		})
	}
}
