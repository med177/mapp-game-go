---
name: mapp-scenario-data-guidelines
description: Mapp senaryo JSON'u, veri yükleyicisi ve veri araçlarını değiştirirken geriye dönük uyumluluk ve senaryo kaynak kurallarını uygula.
applyTo: "assets/scenarios/**/*.json,internal/scenario/**/*.go,tools/**/*.go"
---

# Senaryo ve veri kuralları

- Veri şeması için [`wiki/dev/data-format.md`](../../wiki/dev/data-format.md), senaryo yükleme ve akış için [`wiki/HOME.md`](../../wiki/HOME.md) ile [`README.md`](../../README.md) dosyalarını kaynak kabul et.
- `assets/scenarios/<id>/` altındaki `scenario.json` ve `data/` dosyalarını kaynak veri kabul et; senaryoya özgü değerleri Go koduna hardcode etme.
- Önce mevcut senaryo klasörlerini ve ilgili şemayı keşfet. Testleri tek bir senaryo veya bölge kesin varmış gibi yazma; mevcut dizinleri dolaş ve optional dosya/alan yokluğunu destekle.
- Optional alanları gereksiz zorunlu yapma. Eski senaryoların ve mevcut save dosyalarının yüklenmesini koru; yeni alanlar için güvenli varsayılan ve gerekiyorsa serialize/deserialize testi ekle.
- ID, owner, region, neighbor, asset ve cross-file referanslarını gerçek senaryo kayıtlarına çözümle; sessizce geçersiz referans üretme.
- Üretilen veya dönüştürülen harita/veri dosyalarını elle düzenleme; ilgili `tools/` aracını kullan ve aracın çıktısını doğrula.
- Bölge verisi değişince uygun `regionlint` kontrolünü ve ilgili senaryo/yükleme testlerini çalıştır. State/save etkisi varsa serialize/deserialize ve eski kayıt uyumluluğunu da kontrol et.
- JSON değişikliği oyun davranışını veya veri sözleşmesini etkiliyorsa ilgili wiki sayfasını ve özellik tamamlandıysa `wiki/dev/progress.md` dosyasını güncelle.
