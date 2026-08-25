package world

// TerrainType arazi tipini tanımlar.
type TerrainType string

const (
	TerrainPlain       TerrainType = "plain"        // ova — serbest geçiş
	TerrainForest      TerrainType = "forest"       // orman — yavaş, görüş kısıtlı
	TerrainDenseForest TerrainType = "dense_forest" // sık orman
	TerrainDesert      TerrainType = "desert"       // çöl
	TerrainLake        TerrainType = "lake"         // göl
	TerrainRiver       TerrainType = "river"        // nehir
	TerrainSwamp       TerrainType = "swamp"        // bataklık
	TerrainMountain    TerrainType = "mountain"     // dağ — geçilemez blok
	TerrainPass        TerrainType = "pass"         // dar geçit — pusu noktası
	TerrainCoast       TerrainType = "coast"        // kıyı — kara+deniz geçişi
	TerrainSea         TerrainType = "sea"          // deniz — sadece gemi
)

// TerrainProps arazi özelliklerini tutar.
type TerrainProps struct {
	NameTR          string
	MoveCost        int  // 1 tur harcanan hareket puanı
	DefenseBonus    int  // savunma çarpanı yüzdesi (0 = +0%)
	VisibilityRange int  // görüş mesafesi (bölge sayısı)
	Passable        bool // kara orduları geçebilir mi
	SeaPassable     bool // deniz birlikleri geçebilir mi
	AmbushBonus     int  // pusu saldırısı bonusu yüzdesi
}

var TerrainData = map[TerrainType]TerrainProps{
	TerrainPlain: {
		NameTR:   "Ova",
		MoveCost: 1, DefenseBonus: 0, VisibilityRange: 3,
		Passable: true, SeaPassable: false, AmbushBonus: 0,
	},
	TerrainForest: {
		NameTR:   "Orman",
		MoveCost: 2, DefenseBonus: 15, VisibilityRange: 1,
		Passable: true, SeaPassable: false, AmbushBonus: 25,
	},
	TerrainDenseForest: {
		NameTR: "Sık Orman", MoveCost: 3, DefenseBonus: 25, VisibilityRange: 1,
		Passable: true, SeaPassable: false, AmbushBonus: 35,
	},
	TerrainDesert: {
		NameTR: "Çöl", MoveCost: 2, DefenseBonus: 5, VisibilityRange: 4,
		Passable: true, SeaPassable: false, AmbushBonus: 5,
	},
	TerrainLake: {
		NameTR: "Göl", MoveCost: 99, DefenseBonus: 0, VisibilityRange: 3,
		Passable: false, SeaPassable: true, AmbushBonus: 0,
	},
	TerrainRiver: {
		NameTR: "Nehir", MoveCost: 2, DefenseBonus: 10, VisibilityRange: 3,
		Passable: true, SeaPassable: false, AmbushBonus: 10,
	},
	TerrainSwamp: {
		NameTR: "Bataklık", MoveCost: 3, DefenseBonus: 20, VisibilityRange: 1,
		Passable: true, SeaPassable: false, AmbushBonus: 30,
	},
	TerrainMountain: {
		NameTR:   "Dağ",
		MoveCost: 99, DefenseBonus: 30, VisibilityRange: 4,
		Passable: false, SeaPassable: false, AmbushBonus: 0,
	},
	TerrainPass: {
		NameTR:   "Geçit",
		MoveCost: 3, DefenseBonus: 40, VisibilityRange: 2,
		Passable: true, SeaPassable: false, AmbushBonus: 50,
	},
	TerrainCoast: {
		NameTR:   "Kıyı",
		MoveCost: 1, DefenseBonus: 0, VisibilityRange: 3,
		Passable: true, SeaPassable: true, AmbushBonus: 0,
	},
	TerrainSea: {
		NameTR:   "Deniz",
		MoveCost: 1, DefenseBonus: 0, VisibilityRange: 4,
		Passable: false, SeaPassable: true, AmbushBonus: 0,
	},
}

func (t TerrainType) LabelTR() string {
	if def, ok := TerrainData[t]; ok && def.NameTR != "" {
		return def.NameTR
	}
	return string(t)
}
