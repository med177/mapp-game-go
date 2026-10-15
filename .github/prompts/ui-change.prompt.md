---
name: Mapp UI Değişikliği
description: Mapp arayüzünde görünüm, yerleşim, hover veya tıklama davranışı değiştirirken kullan.
argument-hint: İstenen arayüz değişikliğini ve varsa görsel referansı belirt.
agent: agent
---

Bu Mapp UI isteğini uygula:

**İstek:** ${input:request:Değişecek ekran/öğe ve beklenen davranışı yaz}
**Görsel referans:** ${input:reference:Varsa görsel veya kısa tarif; yoksa "yok"}
**Kısıtlar:** ${input:constraints:Korunması gerekenler veya kapsam dışı işler; yoksa "yok"}
**Kabul ölçütü:** ${input:acceptance:İstenen son görünüm/davranış; boşsa isteğimden çıkar}

Çalışırken [proje kurallarını](../../AGENTS.md), [UI yönergelerini](../instructions/ui.instructions.md) ve [ekran rehberini](../../wiki/architecture/ui-screen-guide.md) izle; bu kuralları yeniden kopyalamak yerine kaynak kabul et.

1. Önce mevcut implementasyonu, ilgili testleri ve çalışma ağacındaki değişiklikleri incele; kullanıcı değişikliklerini koru.
2. Var olan `internal/ui` bileşenlerini ve ortak geometri helper'larını kullan. Aynı geometri kaynağı çizim, ölçüm, hover/cursor, hit-test ve click dispatch'i beslesin.
3. Görünüm için kart/panel ölçüsü, hizalama, padding, metin kırpma/sarma, görsel en-boy oranı ve viewport clipping'i; etkileşim için tıklama sonucu, disabled/hover durumu ve overlay katmanını kabul ölçütlerine göre kontrol et.
4. En küçük gerekli değişikliği yap. Oyun kuralı/state mutasyonunu render koduna koyma. Input veya geometri değişiyorsa uygun hedefli regresyon testini ekle ya da güncelle.
5. Go dosyalarını `gofmt` ile biçimlendir; yalnızca ilgili hedefli testleri çalıştır ve değişen dosyalar için Problems tanılarını kontrol et. Proje yönergeleri aksini söylemedikçe oyunu otomatik açıp görsel doğrulama yapma; görsel kontrolü kullanıcı yapacak.

Tamamlandığında değişen dosyaları, çalıştırdığın testleri ve kullanıcı tarafından görsel olarak kontrol edilmesi gereken noktaları kısa maddelerle bildir.
