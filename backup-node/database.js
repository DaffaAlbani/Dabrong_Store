// database.js — SQLite setup dan semua query database
const Database = require('better-sqlite3');
const path = require('path');

const DB_PATH = path.join(__dirname, 'orders.db');
const db = new Database(DB_PATH);

// Aktifkan WAL mode untuk performa lebih baik
db.pragma('journal_mode = WAL');

// Buat tabel jika belum ada
db.exec(`
  CREATE TABLE IF NOT EXISTS orders (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    order_no    TEXT    UNIQUE NOT NULL,
    player_id   TEXT    NOT NULL,
    server_id   TEXT    NOT NULL,
    player_name TEXT,
    product_id  TEXT    NOT NULL,
    product_name TEXT   NOT NULL,
    diamond     INTEGER NOT NULL,
    price       INTEGER NOT NULL,
    unique_code INTEGER NOT NULL DEFAULT 0,
    total_bayar INTEGER NOT NULL,
    bank_name   TEXT,
    bank_account TEXT,
    bank_holder TEXT,
    whatsapp    TEXT,
    payment_method TEXT DEFAULT 'bank_transfer',
    status      TEXT    NOT NULL DEFAULT 'PENDING',
    apigames_ref_id  TEXT,
    apigames_status  TEXT,
    apigames_message TEXT,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    confirmed_at TEXT,
    completed_at TEXT
  );

  CREATE TABLE IF NOT EXISTS products_cache (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id   TEXT UNIQUE NOT NULL,
    product_name TEXT NOT NULL,
    category     TEXT,
    price        INTEGER NOT NULL,
    original_price INTEGER DEFAULT 0,
    status       TEXT,
    cached_at    TEXT NOT NULL DEFAULT (datetime('now', 'localtime'))
  );
`);

// Jalankan migrasi kolom original_price jika belum ada
try {
  db.exec("ALTER TABLE products_cache ADD COLUMN original_price INTEGER DEFAULT 0");
  console.log("[DB] Migration: Kolom original_price berhasil ditambahkan ke tabel products_cache.");
} catch(err) {
  // Kolom sudah ada
}

// ============ ORDER QUERIES ============

function createOrder(data) {
  const stmt = db.prepare(`
    INSERT INTO orders (
      order_no, player_id, server_id, player_name,
      product_id, product_name, diamond, price,
      unique_code, total_bayar,
      bank_name, bank_account, bank_holder, whatsapp
    ) VALUES (
      @order_no, @player_id, @server_id, @player_name,
      @product_id, @product_name, @diamond, @price,
      @unique_code, @total_bayar,
      @bank_name, @bank_account, @bank_holder, @whatsapp
    )
  `);
  const result = stmt.run(data);
  return getOrderById(result.lastInsertRowid);
}

function getOrderById(id) {
  return db.prepare('SELECT * FROM orders WHERE id = ?').get(id);
}

function getOrderByOrderNo(order_no) {
  return db.prepare('SELECT * FROM orders WHERE order_no = ?').get(order_no);
}

function getAllOrders(limit = 100, offset = 0) {
  return db.prepare(`
    SELECT * FROM orders ORDER BY created_at DESC LIMIT ? OFFSET ?
  `).all(limit, offset);
}

function getOrdersByStatus(status) {
  return db.prepare(`
    SELECT * FROM orders WHERE status = ? ORDER BY created_at DESC
  `).all(status);
}

function updateOrderStatus(order_no, status, extra = {}) {
  const now = new Date().toISOString().replace('T', ' ').slice(0, 19);
  const fields = ['status = @status', 'updated_at = @now'];
  const params = { order_no, status, now };

  if (extra.apigames_ref_id) {
    fields.push('apigames_ref_id = @apigames_ref_id');
    params.apigames_ref_id = extra.apigames_ref_id;
  }
  if (extra.apigames_status) {
    fields.push('apigames_status = @apigames_status');
    params.apigames_status = extra.apigames_status;
  }
  if (extra.apigames_message) {
    fields.push('apigames_message = @apigames_message');
    params.apigames_message = extra.apigames_message;
  }
  if (status === 'PROSES' || status === 'KONFIRMASI') {
    fields.push('confirmed_at = @now');
  }
  if (status === 'SUKSES') {
    fields.push('completed_at = @now');
  }

  db.prepare(`UPDATE orders SET ${fields.join(', ')} WHERE order_no = @order_no`).run(params);
  return getOrderByOrderNo(order_no);
}

function countOrders() {
  return db.prepare('SELECT COUNT(*) as total FROM orders').get();
}

function getOrderStats() {
  return db.prepare(`
    SELECT
      COUNT(*) as total,
      SUM(CASE WHEN status='SUKSES' THEN 1 ELSE 0 END) as sukses,
      SUM(CASE WHEN status='PENDING' THEN 1 ELSE 0 END) as pending,
      SUM(CASE WHEN status='PROSES' THEN 1 ELSE 0 END) as proses,
      SUM(CASE WHEN status='GAGAL' THEN 1 ELSE 0 END) as gagal,
      SUM(CASE WHEN status='SUKSES' THEN total_bayar ELSE 0 END) as revenue
    FROM orders
  `).get();
}

// ============ PRODUCTS CACHE / MANUAL QUERIES ============

function cacheProducts(products) {
  const insert = db.prepare(`
    INSERT OR REPLACE INTO products_cache (product_id, product_name, category, price, status)
    VALUES (@product_id, @product_name, @category, @price, @status)
  `);
  const insertMany = db.transaction((prods) => {
    for (const p of prods) insert.run(p);
  });
  insertMany(products);
}

function getCachedProducts() {
  return db.prepare('SELECT * FROM products_cache ORDER BY price ASC').all();
}

function clearProductsCache() {
  db.prepare('DELETE FROM products_cache').run();
}

function addProduct(data) {
  const stmt = db.prepare(`
    INSERT INTO products_cache (product_id, product_name, category, price, original_price, status)
    VALUES (@product_id, @product_name, @category, @price, @original_price, @status)
  `);
  stmt.run(data);
  return db.prepare('SELECT * FROM products_cache WHERE product_id = ?').get(data.product_id);
}

function updateProduct(product_id, data) {
  const fields = [];
  const params = { product_id };
  
  if (data.product_name !== undefined) { fields.push('product_name = @product_name'); params.product_name = data.product_name; }
  if (data.category !== undefined) { fields.push('category = @category'); params.category = data.category; }
  if (data.price !== undefined) { fields.push('price = @price'); params.price = data.price; }
  if (data.original_price !== undefined) { fields.push('original_price = @original_price'); params.original_price = data.original_price; }
  if (data.status !== undefined) { fields.push('status = @status'); params.status = data.status; }

  if (fields.length === 0) return;

  db.prepare(`UPDATE products_cache SET ${fields.join(', ')} WHERE product_id = @product_id`).run(params);
  return db.prepare('SELECT * FROM products_cache WHERE product_id = ?').get(product_id);
}

function deleteProduct(product_id) {
  db.prepare('DELETE FROM products_cache WHERE product_id = ?').run(product_id);
}

module.exports = {
  db,
  createOrder,
  getOrderById,
  getOrderByOrderNo,
  getAllOrders,
  getOrdersByStatus,
  updateOrderStatus,
  countOrders,
  getOrderStats,
  cacheProducts,
  getCachedProducts,
  clearProductsCache,
  addProduct,
  updateProduct,
  deleteProduct,
};
