---
type: architecture
tags: [game-loop, phases, ebitengine, turn-system]
last_updated: 2026-05-10
related: [state-management, render-pipeline]
---

# Oyun Döngüsü & Phase State Machine

**Kaynak:** `internal/game/game.go`

## Ebitengine Entegrasyonu

`Game` struct, Ebitengine'in `ebiten.Game` interface'ini uygular:

```
Update()  → 60 TPS — oyun mantığı
Draw()    → her frame — render
Layout()  → pencere boyutu bildirir
```

`Game` içinde üç bileşen bulunur:
- `gs *state.GameState` — tüm oyun verisi
- `renderer *render.Renderer` — görsel katman
- `evts []*events.Event` — yüklenmiş tarihsel olaylar listesi

---

## Phase State Machine

```
PhaseMainMenu
    ↓ YeniOyun
PhaseScenarioSelect
    ↓ SenaryoSeç
    ├─ EDIT_MODE=true → PhaseEditMode
    ↓ normal oyun
PhaseFactionSelect
    ↓ FraksiyonSeç
PhaseVictorySelect
    ↓ ZaferKoşuluSeç
PhasePlayerTurn ←──────────────────────┐
    ↓ TurSonu                          │
PhaseAITurn                            │
    ↓ (tüm AI fraksiyonlar işlendi)    │
PhaseTurnResolution                    │
    ↓ (çözüm tamamlandı)               │
    ├─ oyun devam → PhasePlayerTurn ───┘
    └─ oyun bitti → PhaseGameOver
```

**Ayrıca:** `PhaseSettings` (ana menüden, ana menüye döner) · `PhasePauseMenu` (ESC ile) · `PhaseEditMode` (`.env` içinde `EDIT_MODE=true` ise senaryo seçildikten sonra açılır)

## Edit Mode

`EDIT_MODE=true` ile senaryo seçildikten sonra oyun bağımsız harita düzenleyici açılır. İlk araçlar settlement ve bölge merkezi düzenleme içindir:

| Aksiyon | Tetikleyici | Açıklama |
|---|---|---|
| Yerleşim seç | Sol tık | En yakın settlement noktasını seçer |
| Yerleşim taşı | Sol tık sürükle | `regions.json` içindeki settlement `x/y` değerlerini canlı günceller; başka kara bölgeye sürüklenirse settlement o bölgenin `settlements[]` listesine aktarılır |
| Yerleşim ekle | Alt + sol tık | Tıklanan kara bölgeye yeni `city` settlement ekler; ID region içinde çakışmayacak şekilde üretilir |
| Yerleşim sil | Delete | Seçili settlement'ı kaldırır; silinen settlement capital ise kalan ilk settlement capital yapılır |
| Bölge ekle | Ctrl + Alt + sol tık veya HUD `Bolge Ekle` | Tıklanan/seçili region'ın `shape_id` alanını paylaşan yeni kara region seed'i oluşturur; Voronoi cache'i yenilenir ve görsel komşular iki yönlü yazılır |
| Bölge sil | HUD `Bolge Sil` veya settlement seçili değilken Delete | Seçili kara region'ı siler; diğer region'lardan neighbor referansı ve o region'daki başlangıç orduları kaldırılır |
| Yerleşim ismi değiştir | F2 veya Enter | Seçili settlement adını düzenler; Enter kaydeder, Esc iptal eder |
| Yerleşim tipi değiştir | HUD `Tip` | Yerleşim tipi dropdown'ını açar; `city`, `town`, `fortress`, `port` değerlerinden doğrudan seçilir |
| Ana yerleşim yap | HUD `Ana Yap` | Seçili settlement'ı tek `is_capital` yerleşim yapar |
| Bölge arazisi değiştir | HUD `Arazi` | Arazi tipi dropdown'ını açar; `plain`, `forest`, `mountain`, `pass`, `coast` değerlerinden doğrudan seçilir |
| Bölge sahibi seç | HUD `Sahip` | Fraksiyon dropdown'ını açar; listeden doğrudan `owner_id` seçilir, boş sahip de seçilebilir |
| Bölge adı değiştir | HUD `Ad TR` / `Ad EN` | Region `name_tr` veya `name` alanını inline metin girişiyle düzenler |
| Bölge kilidi düzenle | HUD `Kilit`, `-10 Tur`, `+10 Tur` | `is_locked` ve `unlock_turn` alanlarını düzenler |
| Komşuları senkronize et | HUD `Komsu Sync` | Seçili region'ın raster/Voronoi görsel komşularını JSON `neighbors` listesine yazar ve karşı tarafı iki yönlü günceller |
| Geniş veri düzenle | HUD `Veri` sekmesi | Faction ekleme/düzenleme formu açılır; ID, adlar, din, renk, playable, kaynaklar, AI değeri ve seçili hedef faction ile başlangıç diplomasi `stance/score` değeri formdan girilir. Faction silme ve seçili ordunun başlangıç region/owner alanı da buradadır |
| Bölge merkezi taşı | Shift + sol tık sürükle | Tıklanan kara bölgenin `world_x/world_y` koordinatlarını taşır; Voronoi harita cache'i fare bırakıldığında yeniden kurulur |
| Voronoi debug aç/kapat | V | Seçili veya hover bölgenin görsel Voronoi komşularını JSON `neighbors` listesiyle karşılaştıran overlay'i açar/kapatır |
| Geri al / ileri al | Ctrl+Z / Ctrl+Y veya Ctrl+Shift+Z | Edit command stack üzerinden settlement, bölge merkezi ve temel alan değişikliklerini geri alır veya yeniden uygular |
| Senaryo kaydet | Ctrl+S | Aktif senaryonun `data/regions.json`, `data/factions.json`, `data/relations.json` ve `data/armies.json` dosyalarına yazar |
| Ana menüye dön | Esc | Değişiklik yoksa edit mode'dan çıkar; kaydedilmemiş değişiklik varsa `Kaydet`, `Kaydetmeden Cik`, `Iptal` seçenekli modal açar |

Alt-sol bilgi HUD'u seçili bölge, settlement veya ordu özetini gösterir. `Harita` sekmesi settlement/region metadata araçlarını, `Veri` sekmesi region dışı başlangıç verisi araçlarını gösterir. `Tip`, `Arazi` ve `Sahip` seçimleri dropdown ile yapılır. Bölge seçiliyken HUD'dan eklenen settlement bölge merkezine konur ve sonradan sürüklenebilir.

Voronoi debug overlay açıkken camgöbeği pikseller seçili/hover bölgenin gerçek raster sınırını gösterir. Yeşil çizgi hem raster/Voronoi komşusu hem JSON komşusu olan bölgeyi, kırmızı çizgi sadece görsel komşu olan bölgeyi, gri çizgi ise sadece JSON `neighbors` listesinde olan bölgeyi gösterir. Sağ üst debug paneli hover edilen pixel'in `RegionAt` sonucunu ve senaryo koordinatını gösterir.

Kamera kontrolleri normal harita ile aynıdır.

---

## Tur Çözümleme Sırası

`resolveTurn()` — `internal/game/game.go:230`

1. `applySeasonEffects(gs)` — kış hasarı, ilkbahar bonusu → [[systems/seasons]]
2. `applyEconomyTick(gs)` — vergi geliri, ticaret → [[systems/economy]]
3. `applyTechTicks(gs)` — aktif araştırma ilerleme sayacı → [[systems/tech-tree]]
4. `applyProductionTicks()` — bina ve birim üretim kuyruğunu ilerletir; tamamlanan oyuncu üretimleri popup/event log bildirimi üretir
5. `applyReligionConversion(gs)` — ele geçirilmiş bölgelerde yavaş din dönüşümü
6. `checkRegionUnlocks(gs)` — kilitli bölgeleri açma koşulları
7. `checkRebellions(gs)` — düşük memnuniyet → isyan kontrolü
8. `checkEliminations(gs)` — bölgesi kalmayan fraksiyon elenir
9. `applyRelationDecay(gs)` — ilişki puanlarını sıfıra doğru çekme
10. `victory.Check(gs)` — zafer/yenilgi koşulu kontrolü → [[systems/victory]]
11. `events.Tick(gs, evts)` — tarihsel olayları tetikle → [[systems/events]]
12. `gs.AdvanceTurn()` — ay/yıl ilerlet

---

## Oyuncu Aksiyonları (PhasePlayerTurn)

| Aksiyon | Tetikleyici | Açıklama |
|---|---|---|
| `ActionEndTurn` | Enter/Space | AI turuna geç |
| `ActionMoveArmy` | Sağ tık | Orduyu komşu bölgeye taşı / savaş |
| `ActionRecruitUnit` | R | Seçili bölgede milis eğitimini üretim kuyruğuna al; aynı üretime tekrar basılırsa iptal edip altını iade eder |
| `ActionRecruitNaval` | N | Kıyı bölgede nakliye gemisi üretimini kuyruğa al; aynı üretime tekrar basılırsa iptal edip altını iade eder |
| `ActionBuild` | 1-6 | market/farm/barracks/port/walls/temple inşaatını kuyruğa al; kuyruktaki binaya tekrar basılırsa iptal edip altını iade eder |
| `ActionResearch` | Tech panelinden | Teknoloji araştır |
| `ActionDeclareWar` | Diplomasi paneli | Savaş ilan et |
| `ActionProposePeace` | Diplomasi paneli | Barış teklif et |
| `ActionProposeAlliance` | Diplomasi paneli | İttifak kur |
| `ActionProposeTrade` | Diplomasi paneli | Ticaret anlaşması |
| `ActionSave` / `ActionLoad` | S / L | Kaydet / Yükle |
| `ActionAdjustTax` | . / , | Vergi ±5% |

---

## Başlangıç Orduları

`army.LoadArmies()` — `internal/army/loader.go`

Başlangıç orduları artık kodda üretilmiyor. Her senaryo `data/armies.json` dosyasında ordu ID'si, sahip fraksiyon, başlangıç bölgesi ve birim sayımlarını tanımlar; yükleyici `count` değerlerini tek tek `army.Unit` kayıtlarına açar.
