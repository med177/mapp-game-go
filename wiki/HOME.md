---
type: index
tags: [home, navigation]
last_updated: 2026-05-10
---

# Mapp Game Go — Wiki

Orta Çağ temalı (1300–1600) sıra tabanlı strateji oyunu. Total War kampanya haritası tarzı — taktik savaş yok, tüm çarpışmalar otomatik hesaplanır.

> **Bakım notu:** Bu wiki LLM tarafından güncel tutulur. Kod değişince ilgili sayfa güncellenir; yeni sistem eklenince yeni sayfa açılır. Hiçbir bilgi hardcode değil — JSON veri dosyalarından veya koddan sentezlenmiştir.

---

## Mimari

| Sayfa | Konu |
|---|---|
| [[architecture/game-loop]] | Ebitengine döngüsü, Phase state machine, tur akışı |
| [[architecture/state-management]] | `GameState` merkezi yapısı, serialize/deserialize |
| [[architecture/render-pipeline]] | Render katmanları, kamera sistemi, input yönetimi |

## Oyun Sistemleri

| Sayfa | Konu |
|---|---|
| [[systems/combat]] | Çarpışma motoru, arazi bonusu, kayıp hesabı |
| [[systems/diplomacy]] | İlişki puanı, duruşlar, diplomatik eylemler |
| [[systems/economy]] | Altın, gelir kaynakları, ticaret güzergahları |
| [[systems/seasons]] | Mevsim mekaniği, 1 tur = 1 ay, ceza/bonuslar |
| [[systems/events]] | Tarihsel olaylar, tetikleme koşulları |
| [[systems/tech-tree]] | Teknoloji araştırma, etkiler, bağımlılıklar |
| [[systems/victory]] | 4 zafer tipi, kontrol mantığı |
| [[systems/ai]] | AI tur mantığı, koalisyon, birim alımı, teknoloji |

## Dünya

| Sayfa | Konu |
|---|---|
| [[world/regions]] | Bölge yapısı, arazi tipleri, komşuluk |
| [[world/factions]] | 9 oynanabilir fraksiyon, din sistemi |

## Geliştirme

| Sayfa | Konu |
|---|---|
| [[dev/progress]] | Tamamlanan/eksik sistemler, bilinen sorunlar |
| [[dev/data-format]] | JSON veri şemaları, assets/data/ yapısı |

---

## Senaryo Sistemi

`internal/scenario/scenario.go` — `assets/scenarios/` altındaki senaryo klasörlerini yükler.

Her senaryo: `scenario.json` (meta + zafer koşulları + müzik playlistleri) + `data/` (regions, factions, armies, events, units, buildings, technologies, country_shapes) + `maps/` + `sprites/` + `musics/`

Paylaşılan oyun içi efektler senaryodan bağımsızdır ve `assets/sounds/` altından yüklenir.

Mevcut senaryolar: `1300_ottoman_rise` · `1444_constantinople`

---

## Hızlı Referans

**Akış:** `PhaseMainMenu` → `PhaseScenarioSelect` → `PhaseFactionSelect` → `PhaseVictorySelect` → oyun

**Tur sırası:** `PhasePlayerTurn` → `PhaseAITurn` → `PhaseTurnResolution` → `PhasePlayerTurn`

**Tur çözümleme sırası** (`internal/game/game.go:230`):
`applySeasonEffects` → `applyEconomyTick` → `applyTechTicks` → `applyReligionConversion` → `checkRegionUnlocks` → `checkRebellions` → `checkEliminations` → `applyRelationDecay` → `victory.Check` → `events.Tick`

**Klavye kısayolları:** `Enter/Space` tur sonu · `Tab` diplomasi · `T` teknoloji · `R` asker al · `N` gemi inşa · `1-6` bina · `S/L` kaydet/yükle · `F11` tam ekran
