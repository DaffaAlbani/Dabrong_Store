// routes/api.js — Public API routes (untuk halaman customer)
const express  = require('express');
const router   = express.Router();
const apigames = require('../apigames');
const db       = require('../database');

// ============================================================
//  GET /api/check-user
//  Verifikasi User ID + Server → tampilkan nickname player
//  Query: ?user_id=123&server_id=3&product_id=AGML
// ============================================================
router.get('/check-user', async (req, res) => {
  const { user_id, server_id } = req.query;

  if (!user_id || !server_id) {
    return res.json({ success: false, message: 'user_id dan server_id wajib diisi' });
  }
  if (!/^\d+$/.test(user_id.trim())) {
    return res.json({ success: false, message: 'User ID hanya boleh berupa angka' });
  }

  console.log(`[CHECK-USER] user_id=${user_id} server_id=${server_id}`);

  const result = await apigames.checkUsername({
    user_id:   user_id.trim(),
    server_id: server_id.trim(),
  });

  // ─── Berhasil dapat nickname ───────────────────────────────
  if (result.success) {
    const apiData  = result.data;
    const nickname =
      apiData?.username         ||
      apiData?.nickname         ||
      apiData?.name             ||
      apiData?.data?.username   ||
      apiData?.data?.nickname   ||
      apiData?.data?.name       ||
      apiData?.result?.username ||
      apiData?.result?.name;

    console.log(`[CHECK-USER] Raw response:`, JSON.stringify(apiData, null, 2));

    if (nickname) {
      return res.json({ success: true, nickname, user_id, server_id });
    }
  }

  // ─── API gagal / signature error → fallback mode ──────────
  // Catat error tapi tetap izinkan user lanjut (tanpa nickname)
  console.warn(`[CHECK-USER] Apigames gagal: ${result.message}`);
  console.warn(`[CHECK-USER] Raw error:`, JSON.stringify(result.raw));

  // Cek apakah errornya "invalid signature" → kemungkinan config salah
  const errMsg = (result.message || '').toLowerCase();
  const isConfigError = errMsg.includes('signature') || errMsg.includes('unauthorized') || result.status === 401;

  if (isConfigError) {
    // Config belum benar tapi tetap izinkan proceed (mode bypass)
    console.warn('[CHECK-USER] ⚠️  Signature error — bypass mode aktif');
    return res.json({
      success:  true,
      nickname: `ID ${user_id}`,   // nickname sementara
      user_id,
      server_id,
      warning:  'Verifikasi nickname tidak tersedia saat ini. Pastikan ID & Server benar sebelum melanjutkan.',
      bypass:   true,
    });
  }

  return res.json({
    success: false,
    message: result.message || 'ID Player tidak ditemukan.',
  });
});

const DEFAULT_PRODUCTS = [
  { product_id: 'AGML022', product_name: '22 Diamond', category: 'AGML', price: 5900, status: 'active' },
  { product_id: 'AGML056', product_name: '56 Diamond', category: 'AGML', price: 13900, status: 'active' },
  { product_id: 'AGML086', product_name: '86 Diamond', category: 'AGML', price: 21900, status: 'active' },
  { product_id: 'AGML172', product_name: '172 Diamond', category: 'AGML', price: 43900, status: 'active' },
  { product_id: 'AGML257', product_name: '257 Diamond', category: 'AGML', price: 64900, status: 'active' },
  { product_id: 'AGML344', product_name: '344 Diamond', category: 'AGML', price: 84900, status: 'active' },
  { product_id: 'AGML514', product_name: '514 Diamond', category: 'AGML', price: 124900, status: 'active' },
  { product_id: 'AGML600', product_name: '600 Diamond', category: 'AGML', price: 144900, status: 'active' },
  { product_id: 'AGML878', product_name: '878 Diamond', category: 'AGML', price: 209900, status: 'active' },
  { product_id: 'AGML1195', product_name: '1195 Diamond', category: 'AGML', price: 279900, status: 'active' },
  { product_id: 'AGML2010', product_name: '2010 Diamond', category: 'AGML', price: 469900, status: 'active' },
  { product_id: 'AGML3688', product_name: '3688 Diamond', category: 'AGML', price: 849900, status: 'active' },
  { product_id: 'AGML5532', product_name: '5532 Diamond', category: 'AGML', price: 1249900, status: 'active' },
];

// ============================================================
//  GET /api/products
//  Ambil daftar produk ML dari database / cache
// ============================================================
router.get('/products', async (req, res) => {
  let cached = db.getCachedProducts();
  
  // Jika database kosong, inisialisasi dengan produk default
  if (cached.length === 0) {
    console.log('[PRODUCTS] Inisialisasi database produk dengan data default...');
    db.cacheProducts(DEFAULT_PRODUCTS);
    cached = db.getCachedProducts();
  }

  // Ambil persentase markup dari .env (default 0.07%)
  const markupPercent = parseFloat(process.env.PRICE_MARKUP_PERCENT || '0.07');
  
  // Hitung harga jual otomatis jika Harga Modal (original_price) diisi
  const products = cached.map(p => {
    const originalPrice = p.original_price || 0;
    if (originalPrice > 0) {
      // Hitung harga jual: Modal * (1 + markup/100)
      const calculatedPrice = Math.ceil(originalPrice * (1 + (markupPercent / 100)));
      return {
        ...p,
        price: calculatedPrice // Ganti harga dengan harga setelah markup
      };
    }
    return p;
  });

  return res.json({ success: true, products });
});

// ============================================================
//  POST /api/order
//  Buat order baru
//  Body: { player_id, server_id, player_name, product_id,
//          product_name, diamond, price, whatsapp }
// ============================================================
router.post('/order', async (req, res) => {
  const {
    player_id, server_id, player_name,
    product_id, product_name, diamond, price,
    whatsapp,
  } = req.body;

  // Validasi input
  if (!player_id || !server_id || !product_id || !price || !diamond) {
    return res.json({ success: false, message: 'Data tidak lengkap' });
  }

  if (!whatsapp || !/^08\d{8,12}$/.test(whatsapp.trim()) && !/^628\d{8,12}$/.test(whatsapp.trim())) {
    return res.json({ success: false, message: 'Nomor WhatsApp tidak valid (contoh: 08123456789)' });
  }

  // Generate order number & kode unik (3 digit terakhir random agar nominal transfer unik)
  const orderNo    = `DML${Date.now().toString().slice(-10)}`;
  const uniqueCode = Math.floor(Math.random() * 900) + 100; // 100–999
  const totalBayar = parseInt(price) + uniqueCode;

  const bankName    = process.env.BANK_NAME    || 'BCA';
  const bankAccount = process.env.BANK_ACCOUNT || '0000000000';
  const bankHolder  = process.env.BANK_HOLDER  || 'Admin';

  try {
    const order = db.createOrder({
      order_no:     orderNo,
      player_id:    player_id.trim(),
      server_id:    server_id.trim(),
      player_name:  player_name || '',
      product_id:   product_id,
      product_name: product_name || `${diamond} Diamond`,
      diamond:      parseInt(diamond),
      price:        parseInt(price),
      unique_code:  uniqueCode,
      total_bayar:  totalBayar,
      bank_name:    bankName,
      bank_account: bankAccount,
      bank_holder:  bankHolder,
      whatsapp:     whatsapp.trim(),
    });

    console.log(`[ORDER-CREATED] ${orderNo} | ${player_name} (${player_id}/${server_id}) | ${diamond}💎 | Rp${totalBayar}`);

    // Buat pesan WhatsApp otomatis
    const waNumber = process.env.WHATSAPP_NUMBER || '6281234567890';
    const waMsg    = encodeURIComponent(
      `Halo Admin, saya sudah transfer untuk order top-up ML:\n\n` +
      `📋 No. Order: ${orderNo}\n` +
      `👤 Player: ${player_name || player_id} (ID: ${player_id} | Server: ${server_id})\n` +
      `💎 Paket: ${product_name || diamond + ' Diamond'}\n` +
      `💰 Total Transfer: Rp ${totalBayar.toLocaleString('id-ID')}\n` +
      `🏦 Ke: ${bankName} ${bankAccount} a/n ${bankHolder}\n\n` +
      `Mohon segera dikonfirmasi. Terima kasih!`
    );
    const waUrl = `https://wa.me/${waNumber}?text=${waMsg}`;

    return res.json({
      success: true,
      order: {
        order_no:     order.order_no,
        player_name:  order.player_name,
        player_id:    order.player_id,
        server_id:    order.server_id,
        product_name: order.product_name,
        diamond:      order.diamond,
        price:        order.price,
        unique_code:  order.unique_code,
        total_bayar:  order.total_bayar,
        bank_name:    order.bank_name,
        bank_account: order.bank_account,
        bank_holder:  order.bank_holder,
        status:       order.status,
        created_at:   order.created_at,
      },
      whatsapp_url: waUrl,
    });
  } catch (err) {
    console.error('[ORDER-ERROR]', err.message);
    return res.json({ success: false, message: 'Gagal membuat order: ' + err.message });
  }
});

// ============================================================
//  GET /api/order/:order_no
//  Cek status order (oleh customer)
// ============================================================
router.get('/order/:order_no', (req, res) => {
  const order = db.getOrderByOrderNo(req.params.order_no.toUpperCase());

  if (!order) {
    return res.json({ success: false, message: 'Order tidak ditemukan' });
  }

  // Sembunyikan data sensitif untuk customer
  return res.json({
    success: true,
    order: {
      order_no:     order.order_no,
      player_name:  order.player_name,
      player_id:    order.player_id,
      product_name: order.product_name,
      diamond:      order.diamond,
      total_bayar:  order.total_bayar,
      bank_name:    order.bank_name,
      bank_account: order.bank_account,
      bank_holder:  order.bank_holder,
      status:       order.status,
      created_at:   order.created_at,
      confirmed_at: order.confirmed_at,
      completed_at: order.completed_at,
    },
  });
});

module.exports = router;
