# Personal Union (Kişisel Birlik) Planı

## Amaç

Personal union, iki ayrı devletin aynı hükümdar veya hanedan tarafından
yönetilmesini temsil eder. Devletler hukuken ayrı kalır; ancak hükümdarın
ölümü, mirasçı durumu, hanedan değişimi veya yerel soyluların kararı birliğin
devam edip etmeyeceğini belirler.

Mevcut oyunda personal union için ayrı bir state yoktur. `allied` kaydı askerî
ve diplomatik yakınlığı, `OverlordID` ise vassallığı temsil eder. Personal union'ı
doğrudan vassallık olarak kullanmak haraç ve diplomasi kısıtları nedeniyle fazla
ağır; yalnızca `allied` olarak tutmak ise ortak hükümdar ve veraset davranışını
modelleyemez. Bu nedenle ileride ayrı bir mekanik eklenmelidir.

## Tarihsel başlangıç adayları

- Fransa – Navarra: 1285–1328 arasındaki kişisel birlik.
- Macaristan – Hırvatistan: ortak taç altında ayrı kurumların korunması.
- León – Kastilya: 1230 sonrası ortak taç; 1310'da personal union'dan kalıcı
  birleşmeye yaklaşmış bir yapı olarak değerlendirilmelidir.
- Yeni eklenecek personal union'lar yalnızca senaryo verisi ve tarihsel
  hükümdar/veraset bilgisi doğrulandıktan sonra tanımlanmalıdır.

## Önerilen veri modeli

Fraksiyonlara veya ayrı bir `personal_unions.json` dosyasına şu alanlar
eklenebilir:

```json
{
  "id": "france_navarra_union",
  "senior_faction_id": "france",
  "junior_faction_id": "navarra",
  "start_year": 1285,
  "start_month": 1,
  "end_year": 1328,
  "succession_mode": "shared_monarch",
  "inheritance_rule": "dynastic",
  "military_obligation": "automatic_defense",
  "economic_integration": "separate_treasuries",
  "historical_source": "..."
}
```

Önerilen runtime state alanları:

- `PersonalUnionID`
- `SeniorFactionID`
- `JuniorFactionID`
- ortak hükümdar veya hanedan kimliği
- birliğin kurulduğu tur
- bir sonraki veraset kontrol tarihi
- ayrılma ihtimali ve ayrılma gerekçesi
- senior/junior tarafının veraset iddiaları

Personal union state'i `Relation.Stance` içine gömülmemelidir. Vassallıkta olduğu
gibi ayrı bir siyasi bağ olarak tutulmalı; normal relation matrisi askerî
ittifak, ticaret ve düşmanlık gibi bağımsız ilişkileri taşımaya devam etmelidir.

## Diplomasi davranışı

Personal union aktifken:

- İki devlet doğrudan birbirine savaş ilan edemez.
- Senior devletin ilan ettiği savaşa junior devlet otomatik katılır.
- Junior devletin dış diplomasi kapasitesi sınırlı olabilir; ancak vassal kadar
  tamamen kapatılmamalıdır.
- Junior devlet ayrı hazine, ekonomi, bölge sahipliği ve yerel ordu state'ini
  korur.
- Senior devlet junior devletin tüm gelirini otomatik alamaz; ayrı bir katkı
  veya ortak savaş masrafı modeli kullanılabilir.
- Personal union, doğrudan ittifak kapasitesine sayılmamalıdır.
- Personal union içindeki iki devletin üçüncü taraflarla yaptığı ittifaklar
  ayrı ayrı değerlendirilmelidir.
- Diplomasi panelinde `İttifak` veya `Vassallık` yerine `Personal Union` /
  `Kişisel Birlik` durumu gösterilmelidir.

## Hükümdar ölümü ve veraset akışı

Hükümdarın ölümü personal union'ı otomatik olarak bitirmemelidir. Veraset
kontrolü şu sırayla yapılmalıdır:

1. Senior ve junior tacın aynı mirasçıya geçip geçmediği kontrol edilir.
2. Aynı mirasçı iki tacı da alırsa union devam eder ve yeni hükümdar kaydedilir.
3. Taclardan biri farklı hanedana veya farklı mirasçıya geçerse union sona erer.
4. Yerel meclis, soylular veya tarihsel event ayrı hükümdar seçebiliyorsa bu
   seçim union'ın devamını veya ayrılmasını etkiler.
5. Birlik sona erdiğinde iki devlet bağımsızlaşır; mevcut ittifak otomatik
   olarak korunmamalı, tarihsel event açıkça istemiyorsa normal barış relation'ı
   oluşturulmalıdır.

## Event ve tarihsel akış

Gerekli event türleri:

- personal union kurulması
- ortak hükümdarın tahta çıkması
- senior hükümdarın ölümü
- junior hükümdarın ölümü
- aynı mirasçının iki tacı alması
- farklı mirasçıların taçları ayırması
- yerel soyluların union'ı reddetmesi
- personal union'ın kalıcı birleşmeye dönüşmesi
- personal union'ın savaş veya isyan sonucunda sona ermesi

Event sonuçları doğrudan ilişki kaydı yazmak yerine ortak personal union helper'ı
kullanmalıdır. Böylece save/load, AI, diplomasi paneli, savaş koalisyonu ve
veraset akışı aynı state'i görür.

## AI ve savaş davranışı

- Senior devlet junior devletin ana savunma garantörü olarak değerlendirilmelidir.
- Junior devletin savaş hedefleri bağımsız kalabilir; senior otomatik olarak
  bütün junior claim'lerini kendi claim'i gibi almamalıdır.
- Junior devlete saldırı, union köküne karşı savaş olarak çözümlenmelidir.
- Savaş çağrısı ekranında junior devlet `vassal` değil, `personal union partner`
  olarak gösterilmelidir.
- AI, union devam ederken senior ve junior arasında savaş veya düşmanca
  diplomasi üretmemelidir.
- Junior'ın bağımsızlık arzusu, hanedan uyumsuzluğu, düşük meşruiyet, yerel
  soylu direnci ve ekonomik baskı ayrılma ihtimalini etkileyebilir.

## UI ve save/load

- Faction detay panelinde `Üst Devlet` yerine `Kişisel Birlik: Fransa` gibi
  ayrı bir satır gösterilmelidir.
- Personal union, vassallık haraç satırını göstermemelidir.
- Diplomasi panelinde ortak hükümdar, senior/junior rolü ve bir sonraki veraset
  kontrolü gösterilmelidir.
- Save/load personal union state'ini, hükümdar/hane bilgilerini, başlangıç
  turunu ve olası ayrılma event durumunu korumalıdır.
- Eski save'lerde `OverlordID` ile yaklaşık modellenen personal union kayıtları
  açık migration kuralı olmadan otomatik dönüştürülmemelidir.

## Uygulama sırası

1. Personal union veri şemasını ve loader doğrulamasını ekle.
2. Runtime state, save/load ve legacy migration desteğini ekle.
3. Senior/junior savaş ve diplomasi kurallarını ortak helper'lara bağla.
4. Hükümdar ölümü ve veraset event altyapısını ekle.
5. AI savaş, diplomasi ve ayrılma kararlarını bağla.
6. Diplomasi paneli, faction detayları ve savaş özeti gösterimini güncelle.
7. Fransa–Navarra, Macaristan–Hırvatistan ve León–Kastilya için tarihsel
   başlangıç/veraset testleri yaz.
8. Gerçek senaryo ile save/load, savaş koalisyonu, ayrılma ve devam etme testleri
   çalıştır.

## Kabul kriterleri

- Personal union vassal haraç ve vassal diplomasi kısıtlarını uygulamaz.
- Aynı hükümdarın ölümü union'ı tek başına bitirmez.
- Farklı veraset sonucunda union güvenli biçimde ayrılır.
- Senior/junior tarafları ayrı devlet ve ekonomi olarak kalır.
- Savaş koalisyonu ve UI personal union'ı vassallıktan ayırır.
- Save/load sonrasında union state'i ve veraset tarihi kaybolmaz.
- Senaryo kaynak JSON'larında yalnız tarihsel olarak doğrulanmış union kayıtları
  bulunur; runtime varsayılanları kaynak dosyaya yazılmaz.
