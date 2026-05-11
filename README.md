# Mapp — Orta Çağ Strateji Oyunu

> Akdeniz'in hâkimi sen olacaksın.

Mapp, **1300–1600 yılları** arasını kapsayan, sıra tabanlı bir harita strateji oyunudur. Total War serisinin kampanya haritasından ilham alarak geliştirilmiştir — taktik savaş sahnesi yoktur; tüm çarpışmalar harita üzerinde otomatik hesaplanır.

---

## Özellikler

- **İzometrik Dünya Haritası** — Akdeniz havzasını kapsayan, Voronoi tabanlı bölge sistemi
- **Sıra Tabanlı Strateji** — 1 tur = 1 ay; mevsimler haritaya yansır, kış ordularınızı eritir
- **9 Oynanabilir Fraksiyon** — Osmanlı, Venedik, Fransa, İngiltere, Memlük, Safevi, Rusya, Aragon, Portekiz
- **Din & Diplomasi Sistemi** — Mezhep farklılıkları ilişkileri etkiler; ittifak, ticaret, savaş ilanı
- **Ekonomi & Şehir Geliştirme** — Vergi oranı, bina üretimi, ticaret malları, isyan mekaniği
- **Teknoloji Ağacı** — Askeri, ekonomi, denizcilik, din kategorileri; bina ve bölge bağımlılıkları
- **Tarihsel Olaylar** — Veba, kıtlık, taht krizleri, Reformasyon; gerçek tarihe yakın tetikleme
- **Yapay Zeka** — 3 zorluk seviyesi, fırsatçı/ekonomik/diplomatik strateji
- **Senaryo Sistemi** — Farklı başlangıç koşulları ve zafer hedefleriyle birden fazla senaryo
- **Kayıt/Yükleme** — JSON tabanlı, insan okunabilir kayıt dosyaları

---

## Senaryolar

| Senaryo | Yıl | Açıklama |
|---|---|---|
| **Osmanlı'nın Yükselişi** | 1300 | Küçük bir Anadolu beyliğinden Akdeniz imparatorluğuna |
| **Konstantinopolis** | 1444 | Doğu Roma'nın son günleri; Fetih eşiğinde |

### Zafer Koşulları (seçilebilir)
- **Toprak Hakimiyeti** — 20+ bölge ve kritik şehirleri ele geçir
- **Ekonomik Güç** — Tur başı 500+ altın geliri 5 tur koru
- **Askeri Üstünlük** — Büyük fraksiyonları yenilgiye uğrat
- **Dinî Zafer** — Kudüs, Roma ve Mekke'yi aynı anda tut

---

## Gereksinimler

- **Go 1.21+**
- **Windows 10/11** (öncelikli platform; cross-platform mimari)
- Bağımlılıklar `go.sum` üzerinden otomatik indirilir

### Bağımlılıklar

| Paket | Sürüm | Kullanım |
|---|---|---|
| `github.com/hajimehoshi/ebiten/v2` | v2.9.9 | Oyun motoru (2D render, input) |
| `golang.org/x/image` | v0.39.0 | Görüntü işleme |
| `github.com/joho/godotenv` | v1.5.1 | Ortam değişkeni yönetimi |

---

## Kurulum & Çalıştırma

```bash
# Repoyu klonla
git clone https://github.com/med177/mapp-game-go.git
cd mapp-game-go

# Bağımlılıkları indir
go mod download

# Derle ve çalıştır
go run ./cmd/game

# Ya da derleyip bin/ altına yaz
go build -o bin/game.exe ./cmd/game
./bin/game.exe
```

---

## Proje Yapısı

```
mapp-game-go/
├── cmd/game/main.go          # Uygulama giriş noktası
├── assets/
│   ├── sounds/               # Paylaşılan oyun efektleri
│   └── scenarios/            # Senaryo veri paketleri
│       ├── scenarios.json
│       ├── 1300_ottoman_rise/
│       │   ├── scenario.json
│       │   ├── data/         # regions, factions, armies, events …
│       │   ├── maps/         # Harita PNG dosyaları
│       │   └── musics/       # scenario.json playlist müzikleri
│       └── 1444_constantinople/
└── internal/
    ├── game/                 # Oyun döngüsü, tur yönetimi
    ├── state/                # Merkezi oyun durumu (GameState)
    ├── render/               # Ebitengine render katmanı
    ├── world/                # Harita, bölge, arazi
    ├── faction/              # Fraksiyon verisi, ilişkiler
    ├── army/                 # Ordu, birlik, hareket
    ├── combat/               # Çarpışma hesaplama motoru
    ├── economy/              # Kaynak, vergi, ticaret
    ├── city/                 # Şehir, kale, bina sistemi
    ├── tech/                 # Teknoloji ağacı
    ├── events/               # Tarihsel olaylar motoru
    ├── ai/                   # Yapay zeka stratejisi
    ├── season/               # Mevsim mekaniği
    ├── victory/              # Zafer koşulları kontrolü
    ├── scenario/             # Senaryo yükleyici
    └── save/                 # Kayıt/yükleme
```

Her sistem kendi `internal/` paketinde izole edilmiştir. Render katmanı ile oyun mantığı ayrı tutulmuş; veri hardcode yerine JSON'dan okunur.

---

## Harita Sistemi

Harita, **Voronoi tabanlı** bölge sistemidir. Her piksel, en yakın bölge merkezine (`world_x`, `world_y`) göre renklenir. Komşuluk listesi yalnızca ordu hareketi ve ticaret için kullanılır — görsel sınırları etkilemez.

Bölge tipleri:

| Tip | Etki |
|---|---|
| Ova | Serbest geçiş |
| Orman | Görüş kısıtlı, yavaş geçiş |
| Dağ | Geçilemez; sadece dar geçitler |
| Deniz | Yalnızca deniz birlikleriyle |
| Geçit | Pusu noktası, stratejik tıkama |

---

## Ordu & Çarpışma

- Bir ordu en fazla **20 birim** taşır.
- Çarpışmalar otomatik hesaplanır; sonucu belirleyen faktörler:
  - Birim sayısı × güç
  - Arazi tipi savunma çarpanı
  - Mevsim cezası/bonusu
  - Pusu bonusu (geçit noktasında hazır bekleyen ordu)

| Kategori | Temel | Orta | Elit |
|---|---|---|---|
| Piyade | Milis | Piyade | Yeniçeri / Şövalye |
| Süvari | Hafif Süvari | Süvari | Ağır Süvari |
| Topçu | Mancınık | Bombarda | Top |

---

## Geliştirme Durumu

Proje aktif geliştirme aşamasındadır. Tamamlanan ve planlanan özellikler için [wiki/dev/progress.md](wiki/dev/progress.md) sayfasına bakın.

---

## Katkıda Bulunma

Bu proje şu an kişisel geliştirme aşamasındadır. Hata bildirimi veya öneri için issue açabilirsiniz.

---

## Lisans

Henüz lisanslanmamıştır. Proje tamamlandığında uygun bir açık kaynak lisansı seçilecektir.

---

## AI Katkısı
- Bu proje genel itibariyle AI modellerinden faydalanılarak geliştirilmektedir. Kod, veri formatları ve dokümantasyon LLM tarafından güncellenir ve sürdürülür.
- Büyük ölçüde GPT-5.4 ve GPT-5.5 modelleri kullanılmıştır. Kodun %90'ından fazlası LLM tarafından yazılmıştır; insan katkısı esas olarak rehberlik, inceleme ve testtir.
- LLM yapısı gereği, kodun bazı bölümleri stil veya yapısal olarak tutarsız olabilir. Ancak, genel mimari ve veri formatları tutarlı kalmaya çalışılmıştır.
- LLM tarafından üretilen kodun doğruluğu ve güvenilirliği için kapsamlı testler yapılmıştır. Ancak, bazı hatalar veya tutarsızlıklar olabilir; bu nedenle kullanıcı geri bildirimi önemlidir.
- Çok az yerde AI tarafından yapılamayan ya da insan müdahalesi gerektiren bölümler olabilir. Bu durumlarda, kodun geri kalanıyla uyumlu olacak şekilde manuel olarak müdahale edilmiştir.
