import re
import json
import os
from collections import defaultdict

print("=========================================")
print("  Dabrong Store — Product Excel Syncer   ")
print("=========================================")

xls_path = "Produk Tokovoucher.xls"
output_path = "database/products.json"

if not os.path.exists(xls_path):
    print(f"ERROR: File {xls_path} tidak ditemukan!")
    exit(1)

print(f"Membaca {xls_path}...")
with open(xls_path, 'r', encoding='utf-8') as f:
    html_content = f.read()

# Temukan baris tabel
rows = re.findall(r'<tr.*?>(.*?)</tr>', html_content, re.DOTALL)
print(f"Ditemukan {len(rows)} baris tabel (termasuk header).")

# Kategori game yang didukung beserta nama aslinya untuk pembersihan
categories_map = {
    'AGFF': 'Free Fire',
    'AGML': 'Mobile Legends',
    'AGPUBG': 'PUBG Mobile',
    'AGCODM': 'CODM',
    'AGVALO': 'Valorant',
    'AGGI': 'Genshin Impact',
    'AGHOK': 'Honor of Kings',
    'AGHSR': 'Honkai Star Rail',
    'AGPBL': 'Point Blank',
    'AGRBLX': 'Roblox',
    'AGSTM': 'Steam Wallet',
    'AGAOV': 'Arena of Valor',
    'AGGS': 'Garena Shells',
    'AGHAGO': 'Hago',
    'AGTOF': 'Tower of Fantasy'
}

def extract_number(name):
    # Cari angka seperti 5, 10, 1000, dll.
    nums = re.findall(r'\b\d[\d,\.]*\b', name)
    if not nums:
        return 0
    # Bersihkan koma/titik pemisah ribuan
    num_str = nums[0].replace(',', '').replace('.', '')
    try:
        return int(num_str)
    except ValueError:
        return 0

def is_subscription(name):
    name_l = name.lower()
    keywords = ["pass", "welkin", "starlight", "twilight", "card", "membership", "package", "pack", "supply", "level up", "coupon", "rookie", "gift card"]
    return any(k in name_l for k in keywords)

def clean_product_name(raw_name, cat, game_name):
    # 1. Bersihkan nama game dan kata kunci umum
    clean_name = raw_name
    clean_name = re.sub(game_name, '', clean_name, flags=re.IGNORECASE)
    
    # 2. Kasus khusus HSR / Star Rail
    if cat == 'AGHSR':
        clean_name = re.sub(r'Honkai\s*:\s*Star\s*Rail', '', clean_name, flags=re.IGNORECASE)
        clean_name = re.sub(r'Honkai\s+Star\s+Rail', '', clean_name, flags=re.IGNORECASE)

    # 3. Kasus khusus MLBB Twilight/Weekly Pass
    if cat == 'AGML':
        if "weekly" in raw_name.lower() and "diamond" in raw_name.lower():
            return "Weekly Diamond Pass"
        if "twilight" in raw_name.lower():
            return "Twilight Pass"
            
    # 4. Hapus kata-kata satuan mata uang game
    clean_name = re.sub(r'\b(diamonds?|points?|vouchers?|crystals?|shells?|oneiric shards?|candy|tokens?|cash|uc|robux)\b', '', clean_name, flags=re.IGNORECASE)
    
    # 5. Bersihkan simbol strip dan spasi berlebih
    clean_name = re.sub(r'\s*-\s*', ' ', clean_name)
    clean_name = re.sub(r'\s+', ' ', clean_name).strip()
    
    # 6. Bersihkan tanda kurung kosong yang tersisa dari pembersihan sebelumnya
    clean_name = re.sub(r'\(\s*\)', '', clean_name)
    clean_name = re.sub(r'\[\s*\]', '', clean_name)
    
    # 7. Bersihkan spasi ganda lagi setelah tanda kurung dihapus
    clean_name = re.sub(r'\s+', ' ', clean_name).strip()
    
    return clean_name if clean_name else raw_name

def is_foreign_region(code, name):
    code_upper = code.upper()
    name_lower = name.lower()
    
    currencies = ["MYR", "PHP", "SGD", "VND", "THB", "USD", "HKD", "EUR", "GBP", "MY", "SG", "PH", "VN", "TH", "KH", "BR", "TR", "BD"]
    for curr in currencies:
        if curr == "SG" and "VGS" in code_upper:
            if code_upper.count("SG") > 1 or code_upper.endswith("SG") or "_SG" in code_upper:
                return True
            continue
        if curr in code_upper:
            return True
            
    foreign_keywords = ["malaysia", "singapore", "philippines", "vietnam", "thailand", "cambodia", "brazil", "turkey", "bangladesh", "hong kong", "myr", "sgd", "usd", "hkd", "php"]
    for kw in foreign_keywords:
        if kw in name_lower:
            return True
            
    return False

# Parsing data produk
parsed_by_cat = defaultdict(list)
skipped_count = 0

for i, row in enumerate(rows[1:]): # Lewati baris header pertama
    cols = re.findall(r'<td.*?>(.*?)</td>', row, re.DOTALL)
    if not cols:
        continue
    # Hapus tag HTML di dalam kolom
    cols_clean = [re.sub(r'<[^>]*>', '', col).strip() for col in cols]
    if len(cols_clean) >= 8:
        kategori, jenis_id, kode_produk, nama_produk, deskripsi, harga, harga_vip, harga_vvip = cols_clean[:8]
        
        # Hanya ambil kategori Topup Game dan Voucher Game
        if kategori not in ('Topup Game', 'Voucher Game'):
            continue
            
        code_upper = kode_produk.upper()
        name_lower = nama_produk.lower()
        
        # Saring produk luar negeri / valuta asing secara global
        if is_foreign_region(kode_produk, nama_produk):
            skipped_count += 1
            continue
            
        # Logika pengelompokan berdasarkan kode dan nama produk
        cat = None
        if "free fire" in name_lower or " ff " in name_lower or code_upper.startswith("FF") or code_upper.startswith("PFF") or code_upper.startswith("MFF") or code_upper.startswith("UPFF"):
            cat = 'AGFF'
        elif ("mobile legend" in name_lower or "mlbb" in name_lower or code_upper.startswith("MLBB") or code_upper.startswith("MCGG") or code_upper.startswith("KPMLBB") or code_upper == "MLWP" or code_upper == "MLTP") and not code_upper.startswith("MLA"):
            cat = 'AGML'
        elif "pubg" in name_lower or code_upper.startswith("PMI") or code_upper.startswith("PUBGM") or code_upper.startswith("PMG") or code_upper.startswith("UPPB"):
            # Bedakan dengan Point Blank
            if "cash" in name_lower or "pb cash" in name_lower or code_upper.startswith("PBC") or code_upper.startswith("VGPB"):
                cat = 'AGPBL'
            else:
                cat = 'AGPUBG'
        elif "codm" in name_lower or "call of duty" in name_lower or code_upper.startswith("CODM") or code_upper.startswith("UPCOM"):
            cat = 'AGCODM'
        elif "valorant" in name_lower or code_upper.startswith("VALO") or code_upper.startswith("UPVL") or code_upper.startswith("VVAL"):
            cat = 'AGVALO'
        elif "genshin" in name_lower or code_upper.startswith("GI") or code_upper.startswith("GIR") or code_upper.startswith("GIP") or code_upper.startswith("UPGI"):
            cat = 'AGGI'
        elif "honor of kings" in name_lower or "hok" in name_lower or code_upper.startswith("HOK") or code_upper.startswith("UPHOK"):
            cat = 'AGHOK'
        elif "star rail" in name_lower or "hsr" in name_lower or code_upper.startswith("HKISR") or code_upper.startswith("UPHSR"):
            cat = 'AGHSR'
        elif "point blank" in name_lower or code_upper.startswith("PBC") or code_upper.startswith("VGPB") or code_upper.startswith("UPPB"):
            cat = 'AGPBL'
        elif "roblox" in name_lower or "robux" in name_lower or code_upper.startswith("ROBGC") or code_upper.startswith("ROBUX"):
            cat = 'AGRBLX'
        elif "steam" in name_lower or code_upper.startswith("STM"):
            cat = 'AGSTM'
        elif "aov" in name_lower or "arena of valor" in name_lower or code_upper.startswith("AOV"):
            cat = 'AGAOV'
        elif ("shell" in name_lower or "shells" in name_lower or code_upper.startswith("VGS")) and not any(x in name_lower for x in ("undawn", "free fire", "ff", "aov", "codm", "speed drifter", "arena of valor")):
            cat = 'AGGS'
        elif "hago" in name_lower or code_upper.startswith("KPHAGO"):
            cat = 'AGHAGO'
        elif "tower of fantasy" in name_lower or code_upper.startswith("TOF"):
            cat = 'AGTOF'
            
        if cat:
            price_val = int(harga) if harga.isdigit() else 0
            
            # Hitung markup harga (misal profit 3%, minimal 500)
            margin = price_val * 0.03
            if margin < 500:
                margin = 500
            selling_price = price_val + margin
            selling_price = ((int(selling_price) + 99) // 100) * 100
            
            # Saring produk "Cek ID" (harga sangat murah/kurang dari 100) atau produk tidak valid
            if price_val < 100 or "cek" in name_lower or "check" in name_lower:
                skipped_count += 1
                continue
                
            parsed_by_cat[cat].append({
                "product_id": kode_produk,
                "product_name": nama_produk,
                "category": cat,
                "original_price": price_val,
                "price": int(selling_price),
                "status": "active"
            })

# Pengelompokan denominasi untuk menyaring duplikat & mengambil harga termurah
final_products = []
product_id_counter = 1

for cat, items in parsed_by_cat.items():
    denom_groups = defaultdict(list)
    for item in items:
        # Tentukan kunci denominasi unik
        name_l = item["product_name"].lower()
        if cat == "AGML" and "weekly" in name_l and ("card" in name_l or "pass" in name_l or "diamond" in name_l) and "elite" not in name_l:
            key = "agml_weekly_diamond_pass"
        elif is_subscription(item["product_name"]):
            key = "sub_" + item["product_name"].lower().strip()
        else:
            extracted_num = extract_number(item["product_name"])
            if extracted_num == 0:
                key = "other_" + clean_product_name(item["product_name"], cat, categories_map[cat]).lower().strip()
            else:
                key = extracted_num

        denom_groups[key].append(item)

    print(f"Kategori {cat} ({categories_map[cat]}): {len(items)} produk mentah -> {len(denom_groups)} denominasi unik.")

    special_items = []
    numeric_items = []
    for key, group in denom_groups.items():
        if key == "agml_weekly_diamond_pass":
            cheapest = min(group, key=lambda x: x["original_price"])
            cheapest["product_name"] = "Weekly Diamond Pass"
            cheapest["price"] = 28000
            special_items.append(cheapest)
        else:
            cheapest = min(group, key=lambda x: x["original_price"])
            cheapest["product_name"] = clean_product_name(cheapest["product_name"], cat, categories_map[cat])

            if is_subscription(cheapest["product_name"]):
                special_items.append(cheapest)
            else:
                num = extract_number(cheapest["product_name"])
                if num == 0:
                    special_items.append(cheapest)
                else:
                    numeric_items.append((num, cheapest))

    numeric_items.sort(key=lambda x: (x[0], x[1]["price"]))

    filtered_numeric = []
    prev_num = None
    for num, item in numeric_items:
        if prev_num is None:
            filtered_numeric.append(item)
            prev_num = num
        else:
            if num <= 50: min_diff, min_ratio = 10, 1.30
            elif num <= 200: min_diff, min_ratio = 35, 1.30
            elif num <= 1000: min_diff, min_ratio = 180, 1.30
            elif num <= 5000: min_diff, min_ratio = 700, 1.30
            else: min_diff, min_ratio = 1800, 1.25

            if (num - prev_num >= min_diff) and (num / prev_num >= min_ratio):
                filtered_numeric.append(item)
                prev_num = num

    cat_final = special_items + filtered_numeric
    print(f"   -> Setelah penyeleksian (100% subscription dipertahankan): {len(cat_final)} produk disimpan.")

    for item in cat_final:
        item["id"] = product_id_counter
        item["cached_at"] = ""
        product_id_counter += 1
        final_products.append(item)

print(f"\nSelesai memproses! Total produk unik yang akan di-save: {len(final_products)} (Disaring {skipped_count} produk Cek ID/Junk).")

# Tulis hasil ke database/products.json
print(f"Menulis data ke {output_path}...")
os.makedirs(os.path.dirname(output_path), exist_ok=True)
with open(output_path, 'w', encoding='utf-8') as f:
    json.dump(final_products, f, indent=2, ensure_ascii=False)

print("✅ Sinkronisasi produk Excel selesai dengan sukses!")
