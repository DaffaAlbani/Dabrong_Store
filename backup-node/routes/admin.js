// routes/admin.js — Admin Panel API routes (dilindungi session auth)
const express  = require('express');
const router   = express.Router();
const apigames = require('../apigames');
const db       = require('../database');

// ============================================================
//  Middleware autentikasi admin
// ============================================================
function requireAdmin(req, res, next) {
  if (req.session && req.session.isAdmin) return next();
  return res.status(401).json({ success: false, message: 'Unauthorized. Silakan login admin.' });
}

// ============================================================
//  POST /api/admin/login
// ============================================================
router.post('/login', (req, res) => {
  const { username, password } = req.body;

  if (
    username === process.env.ADMIN_USERNAME &&
    password === process.env.ADMIN_PASSWORD
  ) {
    req.session.isAdmin = true;
    req.session.username = username;
    return res.json({ success: true, message: 'Login berhasil' });
  }

  return res.status(401).json({ success: false, message: 'Username atau password salah' });
});

// ============================================================
//  POST /api/admin/logout
// ============================================================
router.post('/logout', (req, res) => {
  req.session.destroy();
  return res.json({ success: true });
});

// ============================================================
//  GET /api/admin/me — Cek apakah sudah login
// ============================================================
router.get('/me', requireAdmin, (req, res) => {
  return res.json({ success: true, username: req.session.username });
});

// ============================================================
//  GET /api/admin/stats — Statistik dashboard
// ============================================================
router.get('/stats', requireAdmin, async (req, res) => {
  const stats   = db.getOrderStats();
  const balance = await apigames.checkBalance();

  return res.json({
    success: true,
    stats,
    balance: balance.success ? balance.data : null,
  });
});

// ============================================================
//  GET /api/admin/orders — Semua order dengan filter
//  Query: ?status=PENDING&limit=50&offset=0
// ============================================================
router.get('/orders', requireAdmin, (req, res) => {
  const { status, limit = 50, offset = 0 } = req.query;

  const orders = status
    ? db.getOrdersByStatus(status.toUpperCase())
    : db.getAllOrders(parseInt(limit), parseInt(offset));

  return res.json({ success: true, orders });
});

// ============================================================
//  GET /api/admin/order/:order_no — Detail order
// ============================================================
router.get('/order/:order_no', requireAdmin, (req, res) => {
  const order = db.getOrderByOrderNo(req.params.order_no.toUpperCase());
  if (!order) return res.json({ success: false, message: 'Order tidak ditemukan' });
  return res.json({ success: true, order });
});

// ============================================================
//  POST /api/admin/confirm
//  Konfirmasi bayar → trigger Apigames → kirim diamond
//  Body: { order_no }
// ============================================================
router.post('/confirm', requireAdmin, async (req, res) => {
  const { order_no } = req.body;
  if (!order_no) return res.json({ success: false, message: 'order_no wajib diisi' });

  const order = db.getOrderByOrderNo(order_no.toUpperCase());
  if (!order) return res.json({ success: false, message: 'Order tidak ditemukan' });

  if (order.status !== 'PENDING') {
    return res.json({
      success: false,
      message: `Order sudah dalam status ${order.status}, tidak bisa dikonfirmasi ulang.`,
    });
  }

  // Update status → PROSES
  db.updateOrderStatus(order.order_no, 'PROSES');
  console.log(`[ADMIN-CONFIRM] ${order.order_no} → PROSES`);

  // Kirim ke Apigames
  const result = await apigames.sendTransaction({
    order_no:   order.order_no,
    product_id: order.product_id,
    user_id:    order.player_id,
    server_id:  order.server_id,
  });

  console.log(`[APIGAMES-RESPONSE] ${order.order_no}`, JSON.stringify(result.data));

  if (result.success) {
    const apiStatus = result.data?.status || result.data?.trx_status || 'Sukses';
    const updatedOrder = db.updateOrderStatus(order.order_no, 'SUKSES', {
      apigames_ref_id:  order.order_no,
      apigames_status:  apiStatus,
      apigames_message: result.data?.message || 'Transaksi berhasil',
    });

    return res.json({
      success: true,
      message: `✅ Diamond berhasil dikirim ke ${order.player_name || order.player_id}!`,
      order: updatedOrder,
      apigames: result.data,
    });
  } else {
    // Apigames gagal → update status GAGAL
    const updatedOrder = db.updateOrderStatus(order.order_no, 'GAGAL', {
      apigames_ref_id:  order.order_no,
      apigames_status:  'GAGAL',
      apigames_message: result.message || 'Transaksi gagal di Apigames',
    });

    return res.json({
      success: false,
      message: `❌ Gagal mengirim diamond: ${result.message}`,
      order: updatedOrder,
    });
  }
});

// ============================================================
//  POST /api/admin/check-trx
//  Cek status transaksi langsung ke Apigames
//  Body: { order_no }
// ============================================================
router.post('/check-trx', requireAdmin, async (req, res) => {
  const { order_no } = req.body;
  if (!order_no) return res.json({ success: false, message: 'order_no wajib diisi' });

  const order = db.getOrderByOrderNo(order_no.toUpperCase());
  if (!order) return res.json({ success: false, message: 'Order tidak ditemukan' });

  const result = await apigames.checkTransactionStatus(order.order_no);

  if (result.success) {
    // Update status jika ada perubahan
    const remoteStatus = result.data?.status || result.data?.trx_status;
    if (remoteStatus === 'Sukses' || remoteStatus === 'Success') {
      db.updateOrderStatus(order.order_no, 'SUKSES', {
        apigames_status: remoteStatus,
        apigames_message: result.data?.message,
      });
    }
  }

  return res.json({ success: result.success, data: result.data, message: result.message });
});

// ============================================================
//  POST /api/admin/reject
//  Tolak / batalkan order
//  Body: { order_no, reason }
// ============================================================
router.post('/reject', requireAdmin, (req, res) => {
  const { order_no, reason = 'Dibatalkan oleh admin' } = req.body;
  if (!order_no) return res.json({ success: false, message: 'order_no wajib diisi' });

  const order = db.getOrderByOrderNo(order_no.toUpperCase());
  if (!order) return res.json({ success: false, message: 'Order tidak ditemukan' });

  if (['SUKSES', 'GAGAL'].includes(order.status)) {
    return res.json({ success: false, message: 'Order sudah final, tidak bisa dibatalkan' });
  }

  const updated = db.updateOrderStatus(order.order_no, 'GAGAL', {
    apigames_message: reason,
  });

  return res.json({ success: true, message: 'Order berhasil dibatalkan', order: updated });
});

// ============================================================
//  POST /api/admin/refresh-products
//  Hapus cache produk & ambil ulang dari Apigames
// ============================================================
router.post('/refresh-products', requireAdmin, async (req, res) => {
  db.clearProductsCache();
  const result = await apigames.getProducts('AGML');

  if (!result.success) {
    return res.json({ success: false, message: result.message });
  }

  const raw      = result.data?.data || result.data || [];
  const products = Array.isArray(raw) ? raw.map(p => ({
    product_id:   p.produk || p.product_id || p.kode,
    product_name: p.nama || p.product_name || p.name,
    category:     p.kategori || p.category || 'AGML',
    price:        parseInt(p.harga || p.price || 0),
    status:       p.status || 'active',
  })) : [];

  if (products.length > 0) db.cacheProducts(products);

  return res.json({
    success: true,
    message: `${products.length} produk berhasil diperbarui`,
    products,
  });
});

// ============ MANUAL PRODUCTS MANAGEMENT ============

// POST /api/admin/products — Tambah produk baru
router.post('/products', requireAdmin, (req, res) => {
  const { product_id, product_name, category = 'AGML', price, original_price = 0, status = 'active' } = req.body;

  if (!product_id || !product_name || !price) {
    return res.json({ success: false, message: 'Data tidak lengkap' });
  }

  try {
    const product = db.addProduct({
      product_id: product_id.trim().toUpperCase(),
      product_name: product_name.trim(),
      category,
      price: parseInt(price),
      original_price: parseInt(original_price),
      status
    });
    return res.json({ success: true, message: 'Produk berhasil ditambahkan', product });
  } catch (err) {
    return res.json({ success: false, message: 'Gagal menambah produk: ' + err.message });
  }
});

// PUT /api/admin/products/:product_id — Edit produk
router.put('/products/:product_id', requireAdmin, (req, res) => {
  const { product_id } = req.params;
  const { product_name, category, price, original_price, status } = req.body;

  try {
    const updated = db.updateProduct(product_id.toUpperCase(), {
      product_name,
      category,
      price: price !== undefined ? parseInt(price) : undefined,
      original_price: original_price !== undefined ? parseInt(original_price) : undefined,
      status
    });
    return res.json({ success: true, message: 'Produk berhasil diupdate', product: updated });
  } catch (err) {
    return res.json({ success: false, message: 'Gagal update produk: ' + err.message });
  }
});

// DELETE /api/admin/products/:product_id — Hapus produk
router.delete('/products/:product_id', requireAdmin, (req, res) => {
  const { product_id } = req.params;

  try {
    db.deleteProduct(product_id.toUpperCase());
    return res.json({ success: true, message: 'Produk berhasil dihapus' });
  } catch (err) {
    return res.json({ success: false, message: 'Gagal menghapus produk: ' + err.message });
  }
});

module.exports = router;
