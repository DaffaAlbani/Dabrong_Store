// server.js — Entry point utama aplikasi Express
require('dotenv').config();

const express        = require('express');
const session        = require('express-session');
const cors           = require('cors');
const path           = require('path');

const apiRoutes      = require('./routes/api');
const adminRoutes    = require('./routes/admin');

const app  = express();
const PORT = process.env.PORT || 3000;

// ============================================================
//  MIDDLEWARE
// ============================================================
app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Session untuk admin panel
app.use(session({
  secret:            process.env.SESSION_SECRET || 'ml-topup-secret',
  resave:            false,
  saveUninitialized: false,
  cookie: {
    secure:   false,   // set true jika pakai HTTPS
    httpOnly: true,
    maxAge:   24 * 60 * 60 * 1000, // 24 jam
  },
}));

// Static files (frontend)
app.use(express.static(path.join(__dirname, 'public')));

// ============================================================
//  ROUTES
// ============================================================
app.use('/api',        apiRoutes);
app.use('/api/admin',  adminRoutes);

// Halaman customer
app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

// Halaman cek status order
app.get('/status', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'status.html'));
});

// Admin panel — hanya serve file HTML, auth via API
app.get('/admin', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'admin', 'index.html'));
});

// Catch-all 404 untuk API
app.use('/api/*', (req, res) => {
  res.status(404).json({ success: false, message: 'Endpoint tidak ditemukan' });
});

// ============================================================
//  ERROR HANDLER
// ============================================================
app.use((err, req, res, next) => {
  console.error('[ERROR]', err.stack);
  res.status(500).json({ success: false, message: 'Internal server error' });
});

// ============================================================
//  START SERVER
// ============================================================
app.listen(PORT, () => {
  console.log('\n╔════════════════════════════════════════╗');
  console.log('║   💎 ML Top-Up Server Berjalan!        ║');
  console.log('╠════════════════════════════════════════╣');
  console.log(`║  🌐 Customer : http://localhost:${PORT}   ║`);
  console.log(`║  📋 Status   : http://localhost:${PORT}/status ║`);
  console.log(`║  🔐 Admin    : http://localhost:${PORT}/admin  ║`);
  console.log('╠════════════════════════════════════════╣');
  console.log(`║  Merchant ID : ${(process.env.APIGAMES_MERCHANT_ID || '').slice(0,20)}... ║`);
  console.log('╚════════════════════════════════════════╝\n');
});
