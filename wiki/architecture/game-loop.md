---
type: architecture
tags: [game-loop, phases, ebitengine, turn-system]
last_updated: 2026-07-22
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
    ↓ (AI fraksiyonları `FactionOrder` sırasıyla, adım adım çözülür) │
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
2. `applyEconomyTick(gs)` — vergi geliri, ticaret, abluka kesintili ticaret rotaları, bölge tahıl üretimi, nüfus bazlı sivil tahıl tüketimi ve hareket/kuşatma katsayılı ordu tahıl bakımı; stratejik tahıl talebi fiyat sinyaline yazılır, kapasite üstü tahılla kara ordusu yenilemesi ve açık tercihe göre otomatik ihracat işlenir; Kasım ayında stabil rezerv fazlası uygun bölgelere nüfus büyümesi olarak yatırılır; bölgesel ikmal baskısı da aynı efektif talep hesabını kullanır. Oyuncunun bölge panelindeki tahıl yardımı ayrı bir state mutasyonu olarak tur içinde uygulanır → [[systems/economy]]
3. `applyTechTicks(gs)` — aktif araştırma ilerleme sayacı → [[systems/tech-tree]]
4. `applyProductionTicks()` — bina ve birim üretim kuyruğunu ilerletir; aktif kuşatma altındaki bölge emirleri duraklatılır, kuşatma kalkınca aynı `TurnsLeft` ile devam eder; bölge el değiştirince o bölgedeki üretim emirleri temizlenir; tamamlanan oyuncu üretimleri popup/event log bildirimi üretir
5. `applyReligionConversion(gs)` — ele geçirilmiş bölgelerde yavaş din dönüşümü
6. `resolveSieges()` — aktif kuşatmalarda gedik ilerlemesi, savunucu yıpranması ve teslimiyet/hücum sonucu
7. `checkRegionUnlocks(gs)` — kilitli bölgeleri açma koşulları
8. `checkRebellions(gs)` — düşük memnuniyet → isyan kontrolü
9. `checkEliminations(gs)` — bölgesi kalmayan fraksiyon elenir
10. `applyRelationDecay(gs)` — ilişki puanlarını sıfıra doğru çekme
11. `victory.Check(gs)` — zafer/yenilgi koşulu kontrolü → [[systems/victory]]
12. `events.Tick(gs, evts)` — tarihsel olayları tetikle → [[systems/events]]
13. `events.Apply()` / `events.ApplyChoice()` — olayın anlık etkilerini uygula ve aktif bölge olayına üretim/tüketim modifiyerlerini taşı; bu geçici etkiler bir sonraki ekonomi tick'inde okunur
14. `gs.AdvanceTurn()` — ay/yıl ilerlet

## AI Tur Akışı

`PhaseAITurn` artık tek frame'de tüm AI'yi bitirmez. `internal/game/game.go` içindeki AI sıra denetleyicisi:

1. Oyuncu kameranın anlık konumunu saklar.
2. AI fraksiyonlarını `FactionOrder` tabanlı deterministik sıraya dizer.
3. Her fraksiyon için `ai.TurnStepper` oluşturur.
4. Prelude safhasında diplomasi uygulanır. `1300_ottoman_rise` için acil rezerv ve
   plan türüne bağlı runtime harcama bütçesi üretilir; araştırma, ekonomi, donanma ve
   ordu hazırlıkları soft kategori paylarıyla çözülür. Araştırma adımı plan profili,
   gerçek teknoloji efekti, birim açılımı, üretim/stok darboğazı, istikrar, kıyı erişimi,
   maliyet ve süreyi puanlar; aktif araştırmayı değiştirmez. Ekonomi adımı bina adaylarını
   marjinal ROI, kaynak darboğazı, cephe/objective, istikrar, süre ve kuyrukla puanlar;
   zayıf adayda harcamayı pas geçer. Ordu adımı planın piyade/süvari/kuşatma bileşim
   açığını; gerçek düşman profili, hedef arazisi, kuşatma desteği, maliyet, bakım ve
   üretim süresiyle birlikte puanlar. Seçilen birimin üretim bölgesi kalan throughput,
   kuyruk, güvenlik, üretim sonrası ikmal ve cephe/rally/savunma anchor'ına ağırlıklı
   rota maliyetiyle seçilir. Kullanılmayan kategori payı aynı tur sonraki hazırlığa
   aktarılır. Diğer senaryolar mevcut sabit rezerv, araştırma, birim ve recruit bölgesi
   sırasını korur.
5. Prelude sonunda 1300 senaryosu için cephe, rally ve runtime ordu rolleri yeniden
   üretilir. Yıpranmış/yerel olarak ezilen kara orduları güvenli ikmal anchor'ına
   `retreat`; düşük memnuniyetli fetih bölgelerini koruyacak uygun ordular `security`
   rolü alabilir. Kara rol mesafeleri ve uzun menzilli ilk adım arazi, erişim, tehdit ve
   ikmal ağırlıklı Dijkstra alanından seçilir → [[systems/ai]].
6. Riskli aktif kuşatmasını terk edecek AI ordusu varsa kuşatma, normal hareketten önce
   aynı fraksiyondan kalan uygun orduya devredilir veya kaldırılır; bu işlem ayrı ve
   görünür bir `TurnStep` üretir.
7. Hareket safhasında ordular tek adım ilerler; her adım arasında kısa bekleme bırakılır.
8. Oyuncuya bekleyen diplomasi teklifi düşerse AI sıra makinesi durur ve oyuncu cevabı gelene kadar yeni step çözmez.
9. Oyuncu bölgelerine veya oyuncu ordularına graph mesafesi `<= 3` olan hamlelerde kamera ilgili bölgeye odaklanır ve popup gösterilir.
10. Uzak hamlelerde sadece AI overlay akmaya devam eder; kamera yerinde kalır.
11. Bekleyen teklif kabul edilirse, teklif sahibi aktif AI fraksiyonunun kalan turu kapatılır; aynı tur içinde yeni saldırı veya ileri hareket yapmaz.
12. AI turu bittiğinde kamera eski konumuna geri yüklenir ve `PhaseTurnResolution` başlar.

---

## Oyuncu Aksiyonları (PhasePlayerTurn)

| Aksiyon | Tetikleyici | Açıklama |
|---|---|---|
| `ActionEndTurn` | Enter/Space | Önce `autosave` slotuna kaydeder, sonra AI turuna geç |
| `ActionMoveArmy` | Sağ tık | Orduyu komşu bölgeye taşı; düşman kara ordusu varsa önce savaş planı modalında `Agresif / Dengeli / Savunmacı` seçimi alınır, sonra seçilen duruşla resolve edilir. Aynı modal düşman donanma varsa deniz savaşı için de açılır. Ordu o anda başka bir bölgeyi kuşatıyorsa ve farklı bir komşuya yürüyorsa eski kuşatma otomatik kaldırılır; aktif kuşatmaya aynı fraksiyon ya da müttefik fraksiyon destek için girebilir, ilgisiz üçüncü devletler yeni kuşatma hamlesi üretemez. Tahkimli ve zaten kuşatılmış bölgedeki besieger düşman orduya savaş açılabiliyorsa bu hareket kuşatmayı kaldırır ama yeni kuşatma açmaz |
| `ActionDisembarkArmy` | Sağ tık | Nakliye filosu düşman kıyıya savaş halinde çıkarma yapabilir; savunan ordu varsa önce `Çıkarma Muharebesi` modalı açılır ve seçilen duruş `ActionDisembarkArmy.BattleStance` alanıyla oyun katmanına taşınır. Kendi kıyısında limana dock olduktan sonra aynı kara bölgesine tekrar sağ tıklanınca gemideki birlikler doğrudan karaya indirilir |
| `ActionStartSiege` | Kuşatma modalı | Tahkimli düşman kara bölgesinde aktif kuşatma başlatır; kuşatma birimi varsa bu orduyla daha sonra genel hücum da seçilebilir, kuşatma birimi yoksa yalnız aç bırakma / teslim bekleme hattı açık kalır. Ordu hedefe girmez, hareketi biter. Başka bir devletin devam eden kuşatmasına destek ayrı `ActionMoveArmy` akışıyla gelir; burada yeni bir kuşatma kaydı açılmaz |
| `ActionAssaultSiege` | Kuşatma modalı / seçili kuşatma paneli | Aktif kuşatma üstünden tahkimata genel hücum yapar; gedik ve sur bonusu combat çözümüne katılır |
| `ActionLiftSiege` | Kuşatma modalı / seçili kuşatma paneli | Seçili ordunun yürüttüğü kuşatmayı kaldırır |
| `ActionRecruitUnit` | R | Seçili bölgede milis eğitimini üretim kuyruğuna al; aynı üretime tekrar basılırsa iptal edip altını iade eder |
| `ActionRecruitNaval` | N | Kıyı bölgede nakliye gemisi üretimini kuyruğa al; aynı üretime tekrar basılırsa iptal edip altını iade eder |
| `ActionBuild` | 1-6 | market/farm/barracks/port/walls/temple inşaatını kuyruğa al; kuyruktaki binaya tekrar basılırsa iptal edip altını iade eder |
| `ActionResearch` | Tech panelinden | Teknoloji araştır |
| `ActionDeclareWar` | Diplomasi paneli | Savaş ilan et |
| `ActionProposePeace` | Diplomasi paneli | Barış teklif et |
| `ActionProposeAlliance` | Diplomasi paneli | İttifak kur |
| `ActionProposeTrade` | Diplomasi paneli | Ticaret anlaşması |
| `ActionSave` / `ActionLoad` | S veya Ctrl+S / L | `quicksave` slotuna kaydet / `autosave` slotundan yükle |
| `ActionAdjustTax` | . / , | Vergi ±5% |

---

## Başlangıç Orduları

`army.LoadArmies()` — `internal/army/loader.go`

Başlangıç orduları artık kodda üretilmiyor. Her senaryo `data/armies.json` dosyasında ordu ID'si, sahip fraksiyon, başlangıç bölgesi ve birim sayımlarını tanımlar; yükleyici `count` değerlerini tek tek `army.Unit` kayıtlarına açar.
