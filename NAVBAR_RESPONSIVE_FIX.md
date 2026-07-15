# Evaluasi & Perbaikan Responsif Navbar (Mobile Menu)
**Project:** `ml-topup-v2`  
**File yang diubah:** `style.css`

Dokumen ini berisi **evaluasi responsif** pada navbar (menu navigasi mobile) beserta **langkah perbaikan** agar tampilan menu saat dibuka di HP konsisten dengan tema (Light/Dark Mode). Dokumen ini dirancang agar mudah dieksekusi oleh **Junior Programmer** atau **AI Model Murah**.

---

## 📌 1. Analisis Masalah (Evaluasi Navbar Responsif)

Pada file `style.css`, blok menu navigasi khusus mobile (`.nav-links` di dalam `@media(max-width:768px)`) saat ini menggunakan warna background statis:
`background: rgba(8,8,15,.97);`

**Masalah:**
* Warna `rgba(8,8,15,.97)` adalah warna gelap kehitaman. 
* Meskipun website dalam status **Light Mode** (Mode Terang), menu yang muncul saat hamburger icon di-klik akan tetap berwarna gelap. Hal ini menimbulkan inkonsistensi tema dan menurunkan pengalaman pengguna (UX).
* Saat layar mengecil, karena `.nav-links` ditarik keluar dari aliran (position: fixed), tombol Theme Toggle dan Hamburger menumpuk di sisi kiri, berdekatan dengan logo. Sehingga terlihat "tidak rapi".

---

## 📌 2. Solusi Desain

Agar background menu mengikuti tema (Light/Dark Mode) dan layout sejajar rapi:
1. **Default (Light Mode):** Ubah background menu mobile di `@media(max-width:768px)` menjadi terang, misalnya `rgba(255, 255, 255, 0.97)`.
2. **Override (Dark Mode):** Tambahkan aturan baru untuk `body.dark-mode .nav-links` agar menjadi gelap `rgba(8, 8, 15, 0.97)` saat dark mode aktif.
3. **Layout Tombol:** Tambahkan `margin-left: auto` pada `.theme-toggle-btn` di dalam `@media(max-width:768px)` untuk mendorong tombol tema dan hamburger ke ujung kanan.

---

## 📌 3. Template Prompt Eksekusi (Untuk AI Murah / Junior Programmer)

Salin dan gunakan prompt di bawah ini untuk menugaskan AI model murah melakukan modifikasi file CSS:

```text
Kamu adalah asisten pemrograman CSS. Tugasmu adalah memperbaiki konsistensi warna background menu navbar mobile di file style.css agar mendukung Light/Dark Mode dengan benar.

Instruksi Perubahan:

1. Edit Background Default (Light Mode):
   - Cari blok `@media(max-width:768px)` di dalam style.css.
   - Di dalamnya, temukan selector `.nav-links`.
   - Ubah nilai properti `background: rgba(8,8,15,.97);` menjadi `background: rgba(255,255,255,.97);`.
   - (Jangan hapus properti lain di dalam selector tersebut, seperti position, top, width, dll).

2. Tambahkan Override untuk Dark Mode:
   - Cari blok yang berisi aturan-aturan dark mode, misalnya di sekitar baris yang diawali dengan `body.dark-mode`.
   - Tambahkan aturan CSS berikut di akhir kumpulan gaya dark mode tersebut:
     
     body.dark-mode .nav-links {
       background: rgba(8, 8, 15, 0.97) !important;
       border-left: 1px solid rgba(255, 255, 255, 0.08) !important;
     }

3. Perbaiki Tata Letak (Layout) Tombol Mobile:
   - Cari blok kode CSS untuk `.theme-toggle-btn` (sekitar baris ke-800).
   - Temukan properti `margin-left: 10px;`.
   - Ubah nilainya menjadi `margin-left: auto;` agar pada layar HP, tombol tema dan menu hamburger otomatis terdorong secara rapi ke ujung kanan layar.

4. Pastikan tidak ada karakter atau penutup kurung kurawal yang tertinggal/hilang. Berikan kode akhir yang rapi dan benar.
```

---

## 📌 4. Prosedur Pengujian

Setelah kode diubah dan disimpan:
1. Refresh halaman website di HP atau menggunakan fitur Device Toolbar (Inspect Element -> Responsive) di browser.
2. Klik ikon Hamburger (garis tiga) untuk membuka menu navbar mobile.
3. Pastikan background menu berwarna putih keabu-abuan saat dalam **Light Mode**.
4. Aktifkan **Dark Mode** (via tombol switch bulan/matahari). Buka kembali menu navbar, dan pastikan background-nya berubah menjadi gelap.
