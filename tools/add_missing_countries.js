// Eksik ülkeleri country_shapes.json ve regions.json'a ekler.
// Koordinatlar: 1920x1080 piksel uzayı (MapScale öncesi)
// Projeksiyon: px = 20*lon + 476, py = -18.98*lat + 1234.8
// Kullanım: node tools/add_missing_countries.js

const fs = require('fs');

function lonToPx(lon) { return Math.round(20 * lon + 476); }
function latToPy(lat) { return Math.round(-18.98 * lat + 1234.8); }
function ll(lon, lat) { return [lonToPx(lon), latToPy(lat)]; }

// ──────────────────────────────────────────────────────────────────────────────
// Ülke poligonları  (CW, piksel uzayı)
// Mevcut şekil sınırlarıyla hizalanmış kritik noktalar manüel düzenlenmiştir.
// ──────────────────────────────────────────────────────────────────────────────

const NEW_SHAPES = {

  // ── Slovakya ──────────────────────────────────────────────────────────────
  // Kuzey: Polonya güneyi (y≈308-311), Batı: Çekya doğusu (x≈820),
  // Güney: Macaristan kuzeyi (y≈320-327)
  SVK: [[
    [820, 307], [838, 306], [857, 305], [875, 303],
    [896, 302], [912, 301], [926, 302],
    [928, 310], [928, 320], [922, 327],
    [900, 328], [876, 325], [852, 323], [830, 321], [820, 320],
    [820, 307],
  ]],

  // ── Estonya ───────────────────────────────────────────────────────────────
  // Güney: Letonya kuzeyi (y≈144), Doğu: Rusya batısı (x≈1036-1038)
  EST: [[
    [900, 144], [920, 143], [940, 142], [960, 142],
    [985, 141], [1010, 140], [1036, 139],
    [1037, 126], [1037, 110],
    [1025, 100], [1000, 96], [975, 96], [950, 97],
    [930, 100], [912, 106], [905, 118], [900, 130],
    [900, 144],
  ]],

  // ── Letonya ───────────────────────────────────────────────────────────────
  // Güney: Litvanya kuzeyi (y=166-170), Kuzey: Estonya güneyi (y≈143-144)
  LVA: [[
    [896, 168], [910, 167], [923, 166], [940, 166],
    [960, 166], [985, 167], [1005, 168], [1022, 170],
    [1030, 165], [1036, 157], [1037, 144],
    [1010, 143], [985, 142], [960, 143],
    [940, 143], [920, 144], [900, 144],
    [896, 148], [893, 158], [896, 168],
  ]],

  // ── Beyaz Rusya ───────────────────────────────────────────────────────────
  // Batı: Polonya doğusu (x≈965), Kuzey: Litvanya/Letonya güneyi
  // Doğu: Rusya batısı, Güney: Ukrayna kuzeyi (y≈248)
  BLR: [[
    [965, 215], [975, 213], [987, 211], [995, 209],
    [1001, 208], [1010, 190], [1022, 170], [1036, 163],
    [1052, 168], [1067, 173], [1082, 178],
    [1103, 181], [1121, 211], [1136, 218], [1140, 246],
    [1119, 248], [1080, 249], [1040, 250],
    [1000, 249], [965, 248],
    [965, 232], [965, 215],
  ]],

  // ── Norveç ────────────────────────────────────────────────────────────────
  // Harita tepesinde kırpılmış (lat≈65°N → y≈0), Doğu: İsveç batısı
  NOR: [[
    [638, 140],
    [645, 132], [638, 122], [626, 112], [614, 102],
    [600, 90], [588, 78], [578, 65], [574, 52],
    [577, 38], [585, 24], [596, 11], [610, 2],
    [628, 0], [650, 0], [675, 0], [698, 0],
    [714, 8], [718, 22], [714, 38], [710, 54],
    [706, 70], [702, 86], [698, 102], [694, 116],
    [688, 128], [682, 138], [670, 142], [656, 141],
    [645, 140], [638, 140],
  ]],

  // ── İsveç ─────────────────────────────────────────────────────────────────
  // Güney kıyısı Danimarka'nın hemen doğusunda, Doğu: Baltık/Botniya körfezi
  SWE: [[
    [698, 0], [714, 8], [718, 22], [714, 38], [710, 54],
    [706, 70], [702, 86], [698, 102], [694, 116],
    [688, 128], [682, 138], [698, 152], [706, 165],
    [712, 178], [720, 188],
    [740, 184], [762, 180], [786, 178], [808, 178],
    [828, 182], [840, 188],
    [844, 170], [846, 150], [844, 130], [844, 110],
    [844, 90], [842, 70], [840, 50], [840, 30],
    [840, 10], [840, 0],
    [810, 0], [780, 0], [750, 0], [720, 0], [698, 0],
  ]],

  // ── Finlandiya ────────────────────────────────────────────────────────────
  // Güney: Finlandiya körfezi (Estonya'nın hemen kuzeyi), Doğu: Rusya
  // Batı: Botniya körfezi (İsveç'ten denizle ayrılır)
  FIN: [[
    [906, 96], [930, 93], [955, 92], [980, 92],
    [1006, 93], [1028, 95], [1037, 97],
    [1039, 84], [1041, 68], [1043, 52],
    [1046, 36], [1048, 20], [1050, 6], [1052, 0],
    [1025, 0], [995, 0], [965, 0], [935, 0],
    [905, 0], [875, 0], [848, 0],
    [846, 12], [844, 28], [844, 44], [844, 60],
    [844, 76], [848, 90], [870, 94], [890, 95],
    [906, 96],
  ]],
};

// ──────────────────────────────────────────────────────────────────────────────
// Yeni bölge tanımları (regions.json'a eklenecek)
// ──────────────────────────────────────────────────────────────────────────────
const NEW_REGIONS = [
  {
    id: "slovakia", name: "Slovakia", name_tr: "Slovakya",
    terrain: "mountain", owner_id: "",
    neighbors: ["bohemia", "poland", "hungary", "austria"],
    world_x: 875, world_y: 314,
    is_sea: false, is_locked: false, unlock_turn: 0,
    base_gold_income: 25, base_grain_output: 30, trade_capacity: 2,
    satisfaction: 65, tax_rate: 30, population: 500,
    religion: "catholic", active_event_id: "", shape_id: "SVK",
  },
  {
    id: "estonia", name: "Estonia", name_tr: "Estonya",
    terrain: "plain", owner_id: "",
    neighbors: ["latvia", "_sea_baltic", "_sea_north"],
    world_x: 968, world_y: 120,
    is_sea: false, is_locked: false, unlock_turn: 0,
    base_gold_income: 20, base_grain_output: 20, trade_capacity: 2,
    satisfaction: 60, tax_rate: 25, population: 200,
    religion: "catholic", active_event_id: "", shape_id: "EST",
  },
  {
    id: "latvia", name: "Latvia", name_tr: "Letonya",
    terrain: "plain", owner_id: "",
    neighbors: ["estonia", "lithuania", "belarus", "_sea_baltic"],
    world_x: 966, world_y: 155,
    is_sea: false, is_locked: false, unlock_turn: 0,
    base_gold_income: 22, base_grain_output: 25, trade_capacity: 2,
    satisfaction: 60, tax_rate: 25, population: 250,
    religion: "catholic", active_event_id: "", shape_id: "LVA",
  },
  {
    id: "belarus", name: "Belarus", name_tr: "Beyaz Rusya",
    terrain: "plain", owner_id: "",
    neighbors: ["poland", "lithuania", "latvia", "moscow", "wallachia"],
    world_x: 1052, world_y: 208,
    is_sea: false, is_locked: false, unlock_turn: 0,
    base_gold_income: 28, base_grain_output: 35, trade_capacity: 2,
    satisfaction: 60, tax_rate: 30, population: 600,
    religion: "orthodox", active_event_id: "", shape_id: "BLR",
  },
  {
    id: "norway", name: "Norway", name_tr: "Norveç",
    terrain: "mountain", owner_id: "",
    neighbors: ["sweden", "_sea_north", "_sea_atlantic"],
    world_x: 650, world_y: 68,
    is_sea: false, is_locked: false, unlock_turn: 0,
    base_gold_income: 30, base_grain_output: 15, trade_capacity: 3,
    satisfaction: 65, tax_rate: 30, population: 350,
    religion: "catholic", active_event_id: "", shape_id: "NOR",
  },
  {
    id: "sweden", name: "Sweden", name_tr: "İsveç",
    terrain: "forest", owner_id: "",
    neighbors: ["norway", "denmark", "finland", "_sea_baltic"],
    world_x: 768, world_y: 88,
    is_sea: false, is_locked: false, unlock_turn: 0,
    base_gold_income: 35, base_grain_output: 30, trade_capacity: 3,
    satisfaction: 65, tax_rate: 30, population: 700,
    religion: "catholic", active_event_id: "", shape_id: "SWE",
  },
  {
    id: "finland", name: "Finland", name_tr: "Finlandiya",
    terrain: "forest", owner_id: "",
    neighbors: ["sweden", "estonia", "_sea_baltic", "_sea_north"],
    world_x: 945, world_y: 46,
    is_sea: false, is_locked: false, unlock_turn: 0,
    base_gold_income: 20, base_grain_output: 20, trade_capacity: 2,
    satisfaction: 65, tax_rate: 25, population: 300,
    religion: "catholic", active_event_id: "", shape_id: "FIN",
  },
];

// ──────────────────────────────────────────────────────────────────────────────
// country_shapes.json güncelle
// ──────────────────────────────────────────────────────────────────────────────
const shapesPath = 'assets/data/generated/country_shapes.json';
const shapeFile = JSON.parse(fs.readFileSync(shapesPath, 'utf8'));

const existingIDs = new Set(shapeFile.shapes.map(s => s.id));
let addedShapes = 0;
for (const [id, rings] of Object.entries(NEW_SHAPES)) {
  if (existingIDs.has(id)) {
    console.log('ATLANDI (zaten var):', id);
    continue;
  }
  shapeFile.shapes.push({ id, name: id, rings });
  addedShapes++;
  console.log('Shape eklendi:', id, '- ring nokta sayısı:', rings[0].length);
}
fs.writeFileSync(shapesPath, JSON.stringify(shapeFile, null, 2));
console.log(`\n${addedShapes} shape eklendi → ${shapesPath}`);

// ──────────────────────────────────────────────────────────────────────────────
// regions.json güncelle
// ──────────────────────────────────────────────────────────────────────────────
const regionsPath = 'assets/data/regions.json';
const regions = JSON.parse(fs.readFileSync(regionsPath, 'utf8'));
const existingRegionIDs = new Set(regions.map(r => r.id));

let addedRegions = 0;
for (const region of NEW_REGIONS) {
  if (existingRegionIDs.has(region.id)) {
    console.log('ATLANDI (bölge zaten var):', region.id);
    continue;
  }
  regions.push(region);
  addedRegions++;
  console.log('Bölge eklendi:', region.id, '(', region.name_tr, ')');
}
fs.writeFileSync(regionsPath, JSON.stringify(regions, null, 2));
console.log(`\n${addedRegions} bölge eklendi → ${regionsPath}`);
