---
name: scenario-validation
description: Use when validating Mapp scenario JSON changes, checking region references, running regionlint, or verifying scenario loading and save compatibility.
user-invocable: true
---

# Senaryo doğrulama

1. Önce `assets/scenarios/` altındaki mevcut senaryo dizinlerini keşfet; tek bir senaryonun veya bölgenin her zaman var olduğunu varsayma.
2. Değişen JSON dosyasının şemasını [`wiki/dev/data-format.md`](../../../wiki/dev/data-format.md) ile karşılaştır. Optional alanları zorunlu hale getirme ve mevcut alanları silme.
3. Bölge verisi değiştiyse ilgili senaryoda `go run ./tools/regionlint -file <scenario>/data/regions.json` çalıştır.
4. Referansları kontrol et: faction, owner, region, neighbor, asset ve cross-file ID'leri aynı senaryonun gerçek kayıtlarına çözülmeli.
5. Senaryo yükleme ile ilgili hedefli testleri, ardından gerekirse `go test ./... -count=1` çalıştır. Save/state etkisi varsa serialize/deserialize ve eski kayıt yükleme testlerini de çalıştır.
6. Üretilmiş harita veya veri çıktısı değiştiyse kullanılan `tools/` aracını ve çıktının doğrulamasını raporla; generated dosyayı elle düzeltme.
7. Sonuçta doğrulanan senaryoları, çalıştırılan kontrolleri, başarısızlıkları ve doğrulanamayan noktaları kısa şekilde raporla.

Veri davranışı veya şema değiştiyse ilgili wiki sayfasını ve özellik tamamlandıysa [`wiki/dev/progress.md`](../../../wiki/dev/progress.md) dosyasını güncelle.
