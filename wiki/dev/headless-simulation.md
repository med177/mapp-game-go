---
type: dev
tags: [simulation, reporting, headless]
last_updated: 2026-09-28
related: [progress, architecture/game-loop, systems/ai]
---

# Headless kampanya raporu

`cmd/simreport`, Ebitengine oyun penceresi açmadan normal AI tur sırasını ve
`GameState.resolveTurn()` çözümlemesini çalıştırır. Böylece bir senaryonun
uzun dönemli savaş, toprak, ekonomi, üretim, diplomasi ve olay sonuçları tek
bir raporda incelenebilir.

Örnek:

```bash
go run ./cmd/simreport \
  -scenario assets/scenarios/1300_ottoman_rise \
  -turns 100 -player ottoman -seed 13100301 -difficulty 2 \
  -out reports/headless_100_turn_report.md
```

`-format json` makine tarafından işlenebilir çıktı üretir. `-out` verilmezse
çıktı stdout'a yazılır. Aynı seed ile AI kararları, savaş zarları ve rastgele
olay seçimleri tekrarlanabilir.

Simülasyon politikasında tüm devletler AI tarafından yönetilir. Oyuncu adına
gelen barış, ittifak ve ticaret teklifleri kabul edilir; teslim olma ve
kuşatma kaynaklı vassallık teklifleri reddedilir. Tarihsel olaylarda senaryonun
otomatik seçim kuralı kullanılır. Bu nedenle çıktı, insan oyuncunun yaptığı
seçimlerin kaydı değil, karşılaştırılabilir bir AI kampanya koşusudur.
