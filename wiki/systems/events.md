---
type: system
tags: [events, historical, trigger, notification]
last_updated: 2026-07-22
related: [world/regions, systems/economy, architecture/game-loop, architecture/state-management, architecture/render-pipeline]
---

# Tarihsel Olaylar Sistemi

**Kaynak:** `internal/events/events.go`, `assets/scenarios/*/data/events.json`

## Olay Yapısı

Olaylar JSON'dan yüklenir ve `events.LoadEvents()` ile `[]*Event` listesine dönüştürülür.

Her olay `events.Tick()` içinde yalnızca **tespit edilir**; uygulama sonrasında `events.Apply()` / `events.ApplyChoice()` ile yapılır. Tek seferlik event'ler `gs.FiredEventIDs` içinde takip edilir.

Deterministik simülasyon için `random_region` adayları ile `all_factions`, `all_armies` ve fraksiyon-sahipliği tabanlı aktif event bölge listeleri `RegionID` sırasına sokulur. Rastgele indeks ve `ActiveRegionEvents` sıra düzeni Go map iterasyonundan etkilenmez.

---

## Tetikleme Koşulları

`events.Tick(gs, evts)` — tur çözümleme sırasında çağrılır.

Tetikleme kriterleri:
- **Yıl/ay:** `historical_year` + opsiyonel `historical_month`
- **Rastgele olay:** `probability > 0` ve `min_turn` eşiği
- **Tek seferlik olay:** `one_shot=true` ise tekrar tetiklenmez

---

## Etki Modeli

Base event etkisi eski alanlarla (`target`, `sat_delta`, `gold_delta`, `grain_delta`, `army_hp_mod`, `affected_faction`) tutulur. Hasat/kıtlık/kuraklık gibi bölgesel tahıl olayları ayrıca aktif olay süresi boyunca yüzde modifiyeri taşıyabilir:

- `grain_production_percent`: bölgenin efektif tahıl üretimine eklenen yüzde puan
- `grain_demand_percent`: sivil tahıl talebine eklenen yüzde puan

Bu alanlar `ActiveRegionEvents` içindeki `RegionEventStatus` kaydına kopyalanır. `GameState.RegionProductionSummary()`, ekonomi tick'i, bölgesel ordu lojistiği ve AI aynı aktif kayıtları okur; `TurnsLeft` sıfırlandığında etki kendiliğinden kalkar. Böylece olayın anlık `grain_delta` etkisi ile birkaç tur süren üretim/tüketim baskısı birbirinden ayrılır. `drought`, `bad_harvest`, `famine` ve `harvest` adlandırmaları event ikon tipine otomatik bağlanır.

Yeni choice katmanı:

```json
{
  "choice_prompt_tr": "Devlet nasıl yanıt verecek?",
  "choices": [
    {
      "label_tr": "Karantina Uygula",
      "desc_tr": "Ticaret daralır ama kayıplar sınırlanır.",
      "ai_weight": 8,
      "effect": {
        "target": "player_faction",
        "gold_delta": -180,
        "sat_delta": 8,
        "army_hp_mod": 0.92,
        "relation_delta_all": -5
      }
    }
  ]
}
```

`relation_delta_all`, etkilenen fraksiyonun tüm aktif ilişkilerine doğrudan score delta uygular; bu olay seçimlerinin diplomasi etkisini taşır.

## Zincir Tetikleme

Choice sonuçları artık doğrudan follow-up event açabilir.

- `Effect.set_flags[]` seçilen kararın state flag'ini yazar
- `Effect.clear_flags[]` eski veya rakip branch flag'ini temizler
- `Effect.complete_techs[]` ilgili fraksiyon için teknoloji tamamlar
- `Effect.start_research_tech` aktif araştırma boşsa ücretsiz yönlendirilmiş araştırma başlatır
- `Effect.relations[]` belirli fraksiyonlarla doğrudan `score_delta` ve opsiyonel `stance` uygular
- `Effect.capital_settlement_id` + opsiyonel `capital_move_turns`, ilgili fraksiyon için doğrudan anlık taşımak yerine başkent taşıma kuyruğu başlatır
- `Event.requires_owned_regions[]` follow-up event gelmeden önce ilgili fraksiyonun belirli bölgeleri hâlâ elinde tuttuğunu doğrular
- `Event.requires_techs[]` ilgili fraksiyonun belirli teknolojileri zaten tamamlamış olmasını ister
- `Event.blocks_techs[]` ilgili teknoloji zaten açıksa follow-up event'i bastırır
- `Event.relation_requirements[]` belirli fraksiyonla stance ve score koşulu ister; `any_of_stances`, `blocks_stances`, `min_score`, `max_score` desteklenir
- `Event.requires_flags[]` olmadan event tetiklenmez
- `Event.blocks_flags[]` varsa ilgili flag set iken event bastırılır

Uygulama notu:
- Ayrı bir save yapısı eklenmedi
- flag'ler `gs.FiredEventIDs` içinde `flag:<id>` anahtarıyla tutulur
- bu yüzden zincir state'i save/load ile doğal olarak korunur
- başkent taşıma kuyruğu fraksiyon state'inde saklandığı için event kaynaklı başkent değişimleri de save/load ile doğal olarak korunur

## Olay Tipleri

| Tip | Efekt |
|---|---|
| Veba | Bölge nüfus/üretim düşüşü, komşulara yayılma riski |
| Kıtlık | Tahıl üretimi sıfır, isyan riski artar |
| Taht krizi | Fraksiyon içi isyan veya geçici zayıflık |
| Suikast | Lider/komutan kaybı |
| Dini hareket | Reformasyon, mezhep çatışması |
| Keşif | Yeni bölge açılımı tetikleyici |

---

## Bildirim

Olay tetiklendiğinde:
- `renderer.ShowCombatResult("OLAY: ...")` → kısa bildirim
- Tarihsel event varsa `ShowHistoricalEvent(...)` → tam ekran popup
- Choice varsa historical modal içinde A/B butonları açılır ve sonuç ayrı `KARAR` event log kaydı üretir
- Historical modal artık choice etkisi yanında:
  - açılacak follow-up event adını / tarihini
  - ilgili follow-up için bölge, teknoloji ve diplomasi koşullarını
  aynı ekranda önizleme olarak gösterir
- Event log'da historical event veya karar satırına tıklanınca detay popup içinde aynı zincir özeti sonradan tekrar okunabilir
- Aktif event'ler ana harita ve minimap üzerinde marker olarak çizilmez; event kaydı ilgili bölge seçildiğinde bölge bilgi panelindeki `OLAYLAR` sekmesinde gösterilir
- Bölge panelindeki `OLAYLAR` sekmesi aktif olayları ve `Komşu Bölgeler` listesini `BİNALAR` kartlarıyla aynı içerik alanında gösterir. Event satırı olay adını, tipini ve kalan tur sayısını taşır; aynı bölgede birden fazla aktif event varsa ayrı satırlarda listelenir.
- Event satırına tıklamak mevcut detay popup'ını açar; olay/komşu viewport'u kendi scrollbar'ını ve mouse wheel akışını korur. Komşu başlığındaki `[Daralt] / [Tümünü Göster]` kontrolü kara bölge olaylar sekmesinde de çalışır.
- Detay popup artık başlık, kaynak etiketi ve satır bazlı iz bloğunu ayrı gösterir; `[OLAY]`, `[KARAR]` ve harita izi birbirine karışmaz
- Event log panelindeki `Kodex` düğmesi, oyuncu fraksiyonu için bekleyen tarihsel event havuzunu açar:
  - `Hazir`: tarih ve koşullar uygun
  - `Takvim`: koşullar uygun ama event tarihi henüz gelmedi
  - `Kilitli`: flag, bölge, tech veya diplomasi koşulu eksik
- Kodex popup'ında `Tümü / Hazır / Takvim / Kilitli` filtre sekmeleri bulunur; ok tuşları veya mouse ile değiştirilebilir
- Kodex listesi artık event başına:
  - takvime kalan ay sayısını
  - kilitliyse ilk ve en kritik eksik koşulu
  gösterir
- Kodex sıralaması önceliklidir:
  - önce `Hazır`
  - sonra tarihe en yakın zincir
  - sonra daha az eksikle açılacak zincir
- Kodex satırları görsel olarak ayrışır:
  - `[+]` ve yeşil ton = `Hazır`
  - `[~]` ve altın ton = `Takvim`
  - `[!]` ve kırmızı ton = `Kilitli`
- Kodex artık iki kolonlu çalışır:
  - solda seçilebilir zincir listesi
  - sağda seçilen zincirin tam açıklaması, kalan süre ve eksik koşul dökümü

---

## Mevcut Choice Örnekleri

- `black_death_1347` → `Karantina Uygula` / `Limanları Açık Tut`
- `printing_press_1455` → `Matbaayı Destekle` / `Sansür Uygula`
- `reformation_1517` → `Tolerans Politikası` / `Baskı Uygula`

## Mevcut Zincir Örnekleri

- `fall_of_constantinople_1453`
  - `Şehri İmar Et` → `capital_rebuild_program_1454`
  - `Seferleri Finanse Et` → `arsenal_push_1454`
- `columbus_1492`
  - `Seferleri Büyüt` → `atlantic_casa_program_1494`
  - `İç Pazarı Besle` → `iberian_market_reform_1494`
- `vasco_da_gama_1498`
  - `Tekeli Donanmayla Koru` → `estado_da_india_1499`
  - `Liman Ağını Genişlet` → `feitoria_network_1499`
- `suleiman_rise_1520`
  - `Hazineyi Toparla` → `imperial_defter_reform_1521`
  - `Orduları Güçlendir` → `frontier_host_1521`

Bazı follow-up event'ler yalnız kaynak değil teknoloji de verir:

- `capital_rebuild_program_1454` → ekonomi/idarî teknoloji hattı
- `arsenal_push_1454` → kuşatma/mühendislik hattı
- `atlantic_casa_program_1494` / `estado_da_india_1499` → denizcilik hattı
- `iberian_market_reform_1494` / `feitoria_network_1499` → ticaret hattı

Bazı zincirler artık diplomasi ve araştırma yönü de kurar:

- `capital_rebuild_program_1454` → Ceneviz ile ticaret yakınlaşması + sonraki idarî araştırma
- `atlantic_casa_program_1494` → Kastilya ile ittifak yakınlaşması + kartografi hattı
- `iberian_market_reform_1494` → Portekiz ile ticaret yakınlaşması + mali araştırma
- `imperial_defter_reform_1521` → Venedik ile ticaret yakınlaşması + diplomasi/iaşe hattı

Bazı zincirler artık gerçek harita durumuna da bağlıdır:

- `capital_rebuild_program_1454` / `arsenal_push_1454` için `constantinople` hâlâ ilgili fraksiyonda olmalı
- `atlantic_casa_program_1494` / `iberian_market_reform_1494` için `granada` tutulmalı
- `estado_da_india_1499` / `feitoria_network_1499` için `portugal` ana bölgesi tutulmalı
- `imperial_defter_reform_1521` / `frontier_host_1521` için `constantinople` ve `bithynia` birlikte elde kalmalı

Bazı zincirler artık araştırma ve dış ilişki durumu ile de filtrelenir:

- follow-up ödül tech'i zaten tamamlandıysa event tekrar düşmez
- ticaret/imar zincirleri ilgili partnerle aktif savaş varsa gelmez
- saldırgan sınır zincirleri ise hedef komşuyla `trade` veya `allied` durumunda bastırılabilir

## Eklenecek / Planlanmış

- [ ] Olayların zincirleme tetiklenmesi (veba → kıtlık)
- [x] Bölge ikonu gösterimi (harita üzerinde ❗ ikon)
- [x] Aktif bölge event ikonları tıklanınca detay popup gösterimi
- [x] Event detay popup başlık/kaynak/gövde ayrımı
- [x] Choice sonuçlarını minimap/bölge ikonlarıyla görünür kıl
