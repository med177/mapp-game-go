package world

// TerrainType arazi tipini tanımlar.
type TerrainType string

const (
	TerrainPlain   TerrainType = "plain"   // ova — serbest geçiş
	TerrainForest  TerrainType = "forest"  // orman — yavaş, görüş kısıtlı
	TerrainMountain TerrainType = "mountain" // dağ — geçilemez blok
	TerrainPass    TerrainType = "pass"    // dar geçit — pusu noktası
	TerrainCoast   TerrainType = "coast"   // kıyı — kara+deniz geçişi
	TerrainSea     TerrainType = "sea"     // deniz — sadece gemi
)

// TerrainProps arazi özelliklerini tutar.
type TerrainProps struct {
	MoveCost      int  // 1 tur harcanan hareket puanı
	DefenseBonus  int  // savunma çarpanı yüzdesi (0 = +0%)
	VisibilityRange int // görüş mesafesi (bölge sayısı)
	Passable      bool // kara orduları geçebilir mi
	SeaPassable   bool // deniz birlikleri geçebilir mi
	AmbushBonus   int  // pusu saldırısı bonusu yüzdesi
}

var TerrainData = map[TerrainType]TerrainProps{
	TerrainPlain: {
		MoveCost: 1, DefenseBonus: 0, VisibilityRange: 3,
		Passable: true, SeaPassable: false, AmbushBonus: 0,
	},
	TerrainForest: {
		MoveCost: 2, DefenseBonus: 15, VisibilityRange: 1,
		Passable: true, SeaPassable: false, AmbushBonus: 25,
	},
	TerrainMountain: {
		MoveCost: 99, DefenseBonus: 30, VisibilityRange: 4,
		Passable: false, SeaPassable: false, AmbushBonus: 0,
	},
	TerrainPass: {
		MoveCost: 3, DefenseBonus: 40, VisibilityRange: 2,
		Passable: true, SeaPassable: false, AmbushBonus: 50,
	},
	TerrainCoast: {
		MoveCost: 1, DefenseBonus: 0, VisibilityRange: 3,
		Passable: true, SeaPassable: true, AmbushBonus: 0,
	},
	TerrainSea: {
		MoveCost: 1, DefenseBonus: 0, VisibilityRange: 4,
		Passable: false, SeaPassable: true, AmbushBonus: 0,
	},
}
