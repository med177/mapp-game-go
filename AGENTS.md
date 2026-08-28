# AGENTS.md

Bu dosya, `/mnt/d/mapp-game-go` üzerinde çalışan kod ajanları için çalışma
sözleşmesidir. Proje durumu ve oyun tasarımı için `README.md` ile `wiki/`
dizinini kaynak kabul et; bu dosyada bunların kopyasını tutma.

## Dil ve çalışma ortamı

- Kodlama görevlerinde kullanıcıyla Türkçe iletişim kur.
- Dosyaları UTF-8 olarak oku ve yaz; Türkçe karakterleri ASCII'ye dönüştürme.
- Codex bu çalışma alanında WSL/Ubuntu içinde çalışır. Linux yollarını kullan;
  Windows'a özel kod veya yol gerekiyorsa platform ayrımı kullan.
- Önce ilgili kaynakları, testleri, veri şemasını ve mevcut değişiklikleri
  incele. Varsayım yerine gerçek repo içeriğine dayan.
- Kapsamı değiştiren daha iyi bir yaklaşım çıkarsa önce gerekçeyi ve etkisini
  açıkla; mevcut yaklaşım geçerliyse doğrudan uygula.

## Repo gerçekleri

- Uygulama girişi `cmd/game/main.go`; oyun mantığı `internal/` paketlerinde.
- Ortak UI komponentleri `internal/ui`, Ebitengine ekranları ve UI köprüsü
  `internal/render` altındadır.
- Senaryo paketleri `assets/scenarios/` altındadır; her senaryo kendi
  `scenario.json`, `data/`, harita ve medya dosyalarını taşıyabilir.
- Kayıtlar `saves/`, Windows derleme çıktısı `bin/game.exe` altındadır.
- Harita/veri yardımcıları `tools/` altındadır. Üretilmiş dosyaları elle
  düzenlemek yerine ilgili aracı kullan.

## Mimari ve kod kuralları

- Oyun durumu ve kurallarını render kodundan ayır. State, çözümleme, AI,
  ekonomi, diplomasi ve hareket mantığını UI çizimine bağlama.
- Mevcut canonical helper, domain API ve state akışını kullan; aynı hesabı UI,
  AI ve oyun çözümlemesinde tekrar etme.
- JSON senaryolarını kaynak kabul et; senaryoya özel değerleri Go koduna
  hardcode etme. Optional alanları gereksiz yere zorunlu yapma; mevcut veriyi
  koru.
- `Update` ve `Draw` döngülerinde gereksiz allocation, görüntü oluşturma veya
  pahalı tekrar hesaplardan kaçın. Asset'leri cache'le ve statik hesapları
  uygun yerde önceden üret.
- Kullanıcının mevcut değişikliklerini koru; ilgisiz dosyaları geri alma,
  destructive git komutları veya geniş kapsamlı otomatik dönüşüm kullanma.

## UI kuralları

- Yeni veya değiştirilen panel, modal, overlay, buton, satır, etiket ve liste
  için `internal/ui` komponentlerini ve `internal/render` helper'larını kullan.
- Komponent yetmiyorsa önce ortak komponenti/helper'ı genişlet; tek kullanımlık
  manuel `vector`/`DrawText` UI kopyası oluşturma.
- Çizim, ölçü, hit-test, cursor, tooltip ve input dispatch aynı geometry/rect
  kaynağından türemelidir. Kapatma düğmeleri ortak `IconClose` stilini kullanır.

## Değişiklik ve doğrulama akışı

1. `git status --short` ve ilgili dosyaları incele; kirli çalışma ağacındaki
   kullanıcı değişikliklerini ayır.
2. Arama için `rg`/`rg --files`, JSON/YAML için `jq`/`yq` kullan. Düzenlemeleri
   `apply_patch` ile yap.
3. Go kodunda `gofmt` çalıştır. Davranış, state, save/load, input, routing veya
   ortak helper değiştiyse ilgili regresyon testini ekle/güncelle.
4. Önce hedefli testleri, sonra kapsam uygunsa `go test ./... -count=1` çalıştır;
   iki sonucu ayrı raporla.
5. Gerekli olduğunda `go generate ./internal/buildinfo` ve
   `GOOS=windows GOARCH=amd64 go build -o bin/game.exe ./cmd/game` ile build'i
   doğrula. Veri değişikliklerinde ilgili lint ve yükleme testini de çalıştır.
6. Sonuçta değişen dosyaları, test kapsamını ve doğrulanamayan noktaları açıkça
   bildir.

## Senaryo, save ve test ilkeleri

- Senaryo testleri mevcut dizinleri keşfetmeli; sabit bir senaryo/bölge varmış
  gibi davranmamalı. Optional veri yoksa gereksiz başarısızlık üretmemeli;
  mevcutsa referansların gerçek kayda çözüldüğünü doğrulamalı.
- Save veya state değişiyorsa serialize/deserialize ve varsayılan değerleri
  birlikte kontrol et; eski kayıtların yüklenmesini bozma.
- Yeni mekanik, state, ekonomi, AI, diplomasi, hareket veya save davranışı için
  ilgili `internal/<paket>/*_test.go` altında regresyon testi yaz. Salt görsel
  polish değişikliklerinde test zorunlu değildir.
- Sentetik fixture ile gerçek senaryo verisini karıştırma; map/render kurulumu
  gerekmeyen testlerde gereksiz renderer oluşturma.

## Wiki bakımı

- Mimari veya oyun davranışı değiştiğinde ilgili `wiki/` sayfasını güncelle;
  salt formatlama ve geçici debug değişikliklerinde wiki değiştirme.
- Yeni/önemli wiki sayfaları YAML frontmatter kullanmalı, `last_updated` güncel
  olmalı ve ilgili sayfalara Obsidian wikilink eklenmeli.
- Özellik tamamlandıysa `wiki/dev/progress.md` durumunu güncelle. Kod
  konumlarını mümkün olduğunda `internal/<paket>/<dosya>.go:<satır>` biçiminde
  göster.
