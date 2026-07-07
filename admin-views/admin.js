// public/admin/admin.js — Admin Panel Logic

document.addEventListener('DOMContentLoaded', () => {
  initParticles();
  checkLogin();
});

/* ================================
   AUTH
   ================================ */
async function checkLogin() {
  try {
    const res  = await fetch('/api/admin/me');
    const data = await res.json();
    if (data.success) {
      showPanel(); loadDashboard();
    } else {
      showLogin();
    }
  } catch {
    showLogin();
  }
}

function showLogin() {
  document.getElementById('login-overlay').style.display  = 'flex';
  document.getElementById('admin-wrapper').classList.add('hidden');
}

function showPanel() {
  document.getElementById('login-overlay').style.display  = 'none';
  document.getElementById('admin-wrapper').classList.remove('hidden');
}

// Login form
document.getElementById('btn-login').addEventListener('click', async () => {
  const user = document.getElementById('adm-user').value.trim();
  const pass = document.getElementById('adm-pass').value;
  const errEl = document.getElementById('login-err');
  errEl.classList.add('hidden');

  if (!user || !pass) {
    errEl.textContent = 'Username dan password wajib diisi';
    errEl.classList.remove('hidden'); return;
  }

  const btn = document.getElementById('btn-login');
  btn.disabled = true; btn.textContent = 'Masuk...';

  try {
    const res  = await fetch('/api/admin/login', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: user, password: pass }),
    });
    const data = await res.json();

    if (data.success) {
      showPanel(); loadDashboard();
    } else {
      errEl.textContent = data.message || 'Login gagal';
      errEl.classList.remove('hidden');
    }
  } catch {
    errEl.textContent = 'Gagal terhubung ke server';
    errEl.classList.remove('hidden');
  } finally {
    btn.disabled = false; btn.textContent = 'Login →';
  }
});

document.getElementById('adm-pass').addEventListener('keypress', e => {
  if (e.key === 'Enter') document.getElementById('btn-login').click();
});

// Logout
document.getElementById('btn-logout').addEventListener('click', async () => {
  await fetch('/api/admin/logout', { method: 'POST' });
  showLogin();
});

/* ================================
   PAGE NAVIGATION
   ================================ */
const pages = { dashboard: 'page-dashboard', orders: 'page-orders', pending: 'page-pending', products: 'page-products' };

document.querySelectorAll('.snav-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    const page = btn.dataset.page;
    document.querySelectorAll('.snav-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    document.querySelectorAll('.admin-page').forEach(p => p.classList.remove('active'));
    document.getElementById(pages[page])?.classList.add('active');

    // Load data for page
    if (page === 'dashboard') loadDashboard();
    else if (page === 'orders') loadAllOrders();
    else if (page === 'pending') loadPendingOrders();
    else if (page === 'products') loadProducts();
  });
});

/* ================================
   DASHBOARD
   ================================ */
document.getElementById('btn-refresh-dash').addEventListener('click', loadDashboard);

async function loadDashboard() {
  try {
    const res  = await fetch('/api/admin/stats');
    const data = await res.json();

    if (!data.success) return;

    const s = data.stats;
    document.getElementById('sc-total').textContent   = s.total || 0;
    document.getElementById('sc-pending').textContent = s.pending || 0;
    document.getElementById('sc-proses').textContent  = s.proses || 0;
    document.getElementById('sc-sukses').textContent  = s.sukses || 0;
    document.getElementById('sc-gagal').textContent   = s.gagal || 0;
    document.getElementById('sc-revenue').textContent = fmt(s.revenue || 0);
    document.getElementById('pending-count').textContent = s.pending > 0 ? s.pending : '';

    // Balance
    const bal = data.balance;
    document.getElementById('bc-amount').textContent = bal
      ? `Rp ${parseInt(bal.saldo || bal.balance || bal.deposit || 0).toLocaleString('id-ID')}`
      : 'Tidak tersedia';

    // Recent orders
    const ordRes  = await fetch('/api/admin/orders?limit=10');
    const ordData = await ordRes.json();
    renderRecentTable(ordData.orders || []);
  } catch (err) {
    console.error('loadDashboard error', err);
  }
}

function renderRecentTable(orders) {
  const tbody = document.getElementById('recent-tbody');
  if (!orders.length) {
    tbody.innerHTML = '<tr class="empty-row"><td colspan="6">Belum ada order</td></tr>';
    return;
  }
  tbody.innerHTML = orders.map(o => `
    <tr>
      <td><span class="tbl-order">${o.order_no}</span></td>
      <td>${o.player_name || '—'}<br/><small style="color:var(--muted)">${o.player_id} / Server ${o.server_id}</small></td>
      <td>${o.product_name}<br/><small style="color:var(--muted)">${o.diamond} 💎</small></td>
      <td style="color:var(--gold);font-family:var(--fgame)">${fmt(o.total_bayar)}</td>
      <td><span class="status-badge status-${o.status.toLowerCase()}">${o.status}</span></td>
      <td class="tbl-actions">
        ${o.status === 'PENDING' ? `<button class="btn-sm btn-sm-green" onclick="confirmOrder('${o.order_no}')">✅ Konfirmasi</button>` : ''}
        <button class="btn-sm btn-sm-blue" onclick="viewOrder('${o.order_no}')">🔍 Detail</button>
      </td>
    </tr>
  `).join('');
}

/* ================================
   ALL ORDERS
   ================================ */
document.getElementById('btn-refresh-orders').addEventListener('click', loadAllOrders);
document.getElementById('filter-status').addEventListener('change', loadAllOrders);

async function loadAllOrders() {
  const status = document.getElementById('filter-status').value;
  const url    = status ? `/api/admin/orders?status=${status}` : '/api/admin/orders?limit=100';
  try {
    const res  = await fetch(url);
    const data = await res.json();
    renderAllOrdersTable(data.orders || []);
  } catch (err) { console.error(err); }
}

function renderAllOrdersTable(orders) {
  const tbody = document.getElementById('all-orders-tbody');
  if (!orders.length) {
    tbody.innerHTML = '<tr class="empty-row"><td colspan="7">Tidak ada order</td></tr>';
    return;
  }
  tbody.innerHTML = orders.map(o => `
    <tr>
      <td><span class="tbl-order">${o.order_no}</span><br/><small style="color:var(--muted)">${o.created_at}</small></td>
      <td>${o.player_name || '—'}<br/><small style="color:var(--muted)">ID: ${o.player_id} | Srv: ${o.server_id}</small></td>
      <td>${o.product_name}<br/><small style="color:var(--muted)">${o.diamond} 💎</small></td>
      <td style="color:var(--gold);font-family:var(--fgame)">${fmt(o.total_bayar)}</td>
      <td><a href="https://wa.me/62${o.whatsapp?.replace(/^0/,'')}" target="_blank" style="color:var(--green);font-size:.8rem" title="Chat WhatsApp">${maskWhatsApp(o.whatsapp)}</a></td>
      <td><span class="status-badge status-${o.status.toLowerCase()}">${o.status}</span></td>
      <td class="tbl-actions">
        ${o.status === 'PENDING' ? `<button class="btn-sm btn-sm-green" onclick="confirmOrder('${o.order_no}')">✅</button>` : ''}
        ${['PENDING','PROSES'].includes(o.status) ? `<button class="btn-sm btn-sm-red" onclick="rejectOrder('${o.order_no}')">❌</button>` : ''}
        <button class="btn-sm btn-sm-blue" onclick="viewOrder('${o.order_no}')">🔍</button>
      </td>
    </tr>
  `).join('');
}

/* ================================
   PENDING ORDERS
   ================================ */
document.getElementById('btn-refresh-pending').addEventListener('click', loadPendingOrders);

async function loadPendingOrders() {
  try {
    const res  = await fetch('/api/admin/orders?status=PENDING');
    const data = await res.json();
    renderPendingCards(data.orders || []);
    document.getElementById('pending-count').textContent = data.orders?.length > 0 ? data.orders.length : '';
  } catch (err) { console.error(err); }
}

function renderPendingCards(orders) {
  const list = document.getElementById('pending-list');
  if (!orders.length) {
    list.innerHTML = '<div style="color:var(--muted);text-align:center;padding:40px">✅ Tidak ada order pending saat ini.</div>';
    return;
  }
  list.innerHTML = orders.map(o => `
    <div class="pending-card" id="pcard-${o.order_no}">
      <div class="pc-header">
        <span class="pc-order">${o.order_no}</span>
        <span class="pc-time">⏰ ${o.created_at}</span>
      </div>
      <div class="pc-grid">
        <div class="pc-field"><label>👤 Player</label><strong>${o.player_name || '—'} (ID: ${o.player_id} | Srv: ${o.server_id})</strong></div>
        <div class="pc-field"><label>💎 Paket</label><strong>${o.product_name} (${o.diamond} 💎)</strong></div>
        <div class="pc-field"><label>💰 Total Transfer</label><strong style="color:var(--gold);font-family:var(--fgame)">${fmt(o.total_bayar)}</strong></div>
        <div class="pc-field"><label>💳 Metode</label><strong>${o.payment_method === 'qris' ? 'QRIS' : `${o.bank_name} ${o.bank_account}`}</strong></div>
        <div class="pc-field"><label>📱 WhatsApp</label><a href="https://wa.me/62${o.whatsapp?.replace(/^0/,'')}" target="_blank" style="color:var(--green);font-weight:600" title="Chat WhatsApp">${maskWhatsApp(o.whatsapp)}</a></div>
        <div class="pc-field"><label>🔑 Produk API</label><strong style="font-family:monospace;font-size:.8rem">${o.product_id}</strong></div>
      </div>
      <div class="pc-actions">
        <button class="btn btn-gold" onclick="confirmOrder('${o.order_no}', true)">✅ Konfirmasi Bayar & Kirim Diamond</button>
        <button class="btn btn-ghost" onclick="checkTrx('${o.order_no}')">🔍 Cek Status Apigames</button>
        <button class="btn btn-ghost" style="color:var(--red)" onclick="rejectOrder('${o.order_no}')">❌ Batalkan</button>
      </div>
    </div>
  `).join('');
}

/* ================================
   PRODUCTS
   ================================ */
document.getElementById('btn-refresh-products').addEventListener('click', async () => {
  document.getElementById('admin-products-grid').innerHTML = '<div style="color:var(--muted);padding:20px">Memuat...</div>';
  await loadProducts();
});

// Bind manual add product button
document.getElementById('btn-add-product').addEventListener('click', () => {
  showProductModal();
});

async function loadProducts() {
  try {
    const res  = await fetch('/api/products');
    const data = await res.json();
    renderProducts(data.products || []);
  } catch (err) {
    console.error('loadProducts error:', err);
  }
}

function renderProducts(products) {
  const grid = document.getElementById('admin-products-grid');
  if (!products.length) {
    grid.innerHTML = '<div style="color:var(--muted);padding:20px">Belum ada produk. Silakan tambah produk baru.</div>';
    return;
  }
  grid.innerHTML = products.map(p => {
    const modalHarga = p.original_price || 0;
    const margin = modalHarga > 0 ? (p.price - modalHarga) : 0;
    return `
    <div class="prod-card" style="display:flex;flex-direction:column;justify-content:space-between;min-height:220px">
      <div>
        <div class="prod-id">${p.product_id}</div>
        <div class="prod-name">${p.product_name}</div>
        <div class="prod-price" style="margin-bottom:4px" title="Harga Jual ke Pembeli">💸 Jual: ${fmt(p.price)}</div>
        <div style="font-size:0.75rem;color:var(--muted);margin-bottom:8px">
          📦 Modal: ${modalHarga > 0 ? fmt(modalHarga) : '—'}<br/>
          📈 Profit: <span style="color:var(--green)">${margin > 0 ? '+' + fmt(margin) : '—'}</span>
        </div>
        <span class="prod-status status-${p.status === 'active' ? 'sukses' : 'gagal'}">${p.status || 'active'}</span>
      </div>
      <div style="display:flex;gap:8px;margin-top:16px">
        <button class="btn-sm btn-sm-blue" style="flex:1;justify-content:center" onclick="showProductModal('${p.product_id}', '${p.product_name}', ${p.price}, '${p.status}', '${p.category}', ${p.original_price || 0})">📝 Edit</button>
        <button class="btn-sm btn-sm-red" style="padding:5px 8px" onclick="deleteProduct('${p.product_id}')">🗑️</button>
      </div>
    </div>`;
  }).join('');
}

// Show Add/Edit Modal
function showProductModal(id = '', name = '', price = '', status = 'active', category = 'AGML', originalPrice = 0) {
  const isEdit = id !== '';
  const title = document.getElementById('modal-title');
  const body = document.getElementById('modal-body');
  const acts = document.getElementById('modal-actions');

  title.textContent = isEdit ? '📝 Edit Produk' : '➕ Tambah Produk Baru';
  
  body.innerHTML = `
    <div class="fg">
      <label>Kode Produk (Apigames)</label>
      <input type="text" id="prod-id-inp" placeholder="Contoh: AGML22" value="${id}" ${isEdit ? 'disabled' : ''} style="text-transform:uppercase"/>
      <span class="hint">Harus sama dengan kode produk dari Apigames</span>
    </div>
    <div class="fg">
      <label>Nama Produk (Tampilan)</label>
      <input type="text" id="prod-name-inp" placeholder="Contoh: 22 Diamond" value="${name}"/>
    </div>
    <div class="fg">
      <label>Kategori Game</label>
      <select id="prod-category-inp">
        <option value="AGML" ${category === 'AGML' ? 'selected' : ''}>Mobile Legends (AGML)</option>
        <option value="AGFF" ${category === 'AGFF' ? 'selected' : ''}>Free Fire (AGFF)</option>
        <option value="AGPB" ${category === 'AGPB' ? 'selected' : ''}>PUBG Mobile (AGPB)</option>
        <option value="AGGI" ${category === 'AGGI' ? 'selected' : ''}>Genshin Impact (AGGI)</option>
        <option value="AGVL" ${category === 'AGVL' ? 'selected' : ''}>Valorant (AGVL)</option>
        <option value="AGGS" ${category === 'AGGS' ? 'selected' : ''}>Garena Shells (AGGS)</option>
      </select>
    </div>
    <div class="fg">
      <label>Harga Modal / Beli di Apigames (Rp)</label>
      <input type="number" id="prod-original-price-inp" placeholder="Contoh: 5000" value="${originalPrice}"/>
      <span class="hint">Kosongkan atau isi 0 jika tidak ingin menggunakan markup otomatis</span>
    </div>
    <div class="fg">
      <label>Harga Jual Manual (Rp)</label>
      <input type="number" id="prod-price-inp" placeholder="Contoh: 5900" value="${price}"/>
      <span class="hint">Hanya digunakan jika Harga Modal di atas diisi 0 (tanpa markup)</span>
    </div>
    <div class="fg">
      <label>Status</label>
      <select id="prod-status-inp">
        <option value="active" ${status === 'active' ? 'selected' : ''}>Active</option>
        <option value="inactive" ${status === 'inactive' ? 'selected' : ''}>Inactive</option>
      </select>
    </div>
  `;

  acts.innerHTML = `
    <button class="btn btn-gold" onclick="saveProduct('${id}', ${isEdit})">💾 Simpan</button>
    <button class="btn btn-ghost" onclick="closeModal()">Batal</button>
  `;

  // Auto fill selling price on modal change if markup exists
  const origInp = document.getElementById('prod-original-price-inp');
  const priceInp = document.getElementById('prod-price-inp');
  
  origInp.addEventListener('input', () => {
    const origVal = parseFloat(origInp.value) || 0;
    if (origVal > 0) {
      // 0.07% markup auto-calculation preview
      const markup = 0.07; // default 0.07%
      priceInp.value = Math.ceil(origVal * (1 + (markup / 100)));
      priceInp.disabled = true; // disable direct editing when modal price is used
      priceInp.title = 'Harga jual dihitung otomatis menggunakan markup 0.07%';
    } else {
      priceInp.disabled = false;
    }
  });
  
  // Trigger input once on load
  if (originalPrice > 0) {
    priceInp.disabled = true;
  }

  document.getElementById('modal-overlay').classList.remove('hidden');
}

// Save Product (Create or Update)
async function saveProduct(originalId, isEdit) {
  const name = document.getElementById('prod-name-inp').value.trim();
  const price = document.getElementById('prod-price-inp').value.trim();
  const originalPrice = document.getElementById('prod-original-price-inp').value.trim();
  const category = document.getElementById('prod-category-inp').value;
  const status = document.getElementById('prod-status-inp').value;

  if (!name || (!price && !originalPrice)) {
    alert('Nama dan Harga produk wajib diisi!');
    return;
  }

  let url = '/api/admin/products';
  let method = 'POST';
  let body = { 
    product_name: name, 
    category: category,
    price: price ? parseInt(price) : 0, 
    original_price: originalPrice ? parseInt(originalPrice) : 0,
    status 
  };

  if (isEdit) {
    url += '/' + originalId;
    method = 'PUT';
  } else {
    const id = document.getElementById('prod-id-inp').value.trim().toUpperCase();
    if (!id) {
      alert('Kode Produk wajib diisi!');
      return;
    }
    body.product_id = id;
  }

  showToast('⏳ Menyimpan...', '⏳');

  try {
    const res = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
    const data = await res.json();

    if (data.success) {
      showToast(data.message || '✅ Produk berhasil disimpan!');
      closeModal();
      loadProducts();
    } else {
      showToast(data.message || '❌ Gagal menyimpan produk', '❌');
    }
  } catch (err) {
    showToast('Gagal terhubung ke server', '❌');
  }
}

// Delete Product
async function deleteProduct(productId) {
  if (!confirm(`Apakah Anda yakin ingin menghapus produk ${productId}?`)) return;

  showToast('⏳ Menghapus...', '⏳');

  try {
    const res = await fetch(`/api/admin/products/${productId}`, {
      method: 'DELETE'
    });
    const data = await res.json();

    if (data.success) {
      showToast('✅ Produk berhasil dihapus!');
      loadProducts();
    } else {
      showToast(data.message || '❌ Gagal menghapus produk', '❌');
    }
  } catch (err) {
    showToast('Gagal terhubung ke server', '❌');
  }
}

/* ================================
   ORDER ACTIONS
   ================================ */

// View detail
async function viewOrder(orderNo) {
  const res  = await fetch(`/api/admin/order/${orderNo}`);
  const data = await res.json();
  if (!data.success) return showToast('Order tidak ditemukan', '❌');

  const o    = data.order;
  const body = document.getElementById('modal-body');
  const acts = document.getElementById('modal-actions');

  body.innerHTML = `
    <div class="md-row"><span>No. Order</span><span style="font-family:var(--fgame);color:var(--gold)">${o.order_no}</span></div>
    <div class="md-row"><span>Player</span><span>${o.player_name || '—'} (${o.player_id} / Srv${o.server_id})</span></div>
    <div class="md-row"><span>Produk ID</span><span style="font-family:monospace">${o.product_id}</span></div>
    <div class="md-row"><span>Paket</span><span>${o.product_name} (${o.diamond} 💎)</span></div>
    <div class="md-row md-total"><span>Total Transfer</span><span>${fmt(o.total_bayar)}</span></div>
    <div class="md-row"><span>Metode</span><span>${o.payment_method === 'qris' ? 'QRIS Dabrong Store' : `${o.bank_name} ${o.bank_account} a/n ${o.bank_holder}`}</span></div>
    <div class="md-row"><span>WhatsApp</span><span style="cursor:pointer;color:var(--green)" onclick="this.textContent = '${o.whatsapp || ''}'" title="Klik untuk lihat nomor lengkap">${maskWhatsApp(o.whatsapp)}</span></div>
    <div class="md-row"><span>Status</span><span class="status-badge status-${o.status.toLowerCase()}">${o.status}</span></div>
    ${o.apigames_message ? `<div class="md-row"><span>Info Apigames</span><span>${o.apigames_message}</span></div>` : ''}
    <div class="md-row"><span>Dibuat</span><span>${o.created_at}</span></div>
    ${o.completed_at ? `<div class="md-row"><span>Selesai</span><span>${o.completed_at}</span></div>` : ''}
  `;

  acts.innerHTML = o.status === 'PENDING'
    ? `<button class="btn btn-gold" onclick="confirmOrder('${o.order_no}', true);closeModal()">✅ Konfirmasi Bayar</button>
       <button class="btn btn-ghost" onclick="closeModal()">Tutup</button>`
    : `<button class="btn btn-ghost" onclick="closeModal()">Tutup</button>`;

  document.getElementById('modal-overlay').classList.remove('hidden');
}

// Confirm + send diamond
async function confirmOrder(orderNo, isPending = false) {
  if (!confirm(`Konfirmasi pembayaran dan kirim diamond untuk order ${orderNo}?`)) return;

  showToast('⏳ Memproses...', '⏳', 60000);

  try {
    const res  = await fetch('/api/admin/confirm', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ order_no: orderNo }),
    });
    const data = await res.json();

    if (data.success) {
      showToast(data.message || '✅ Diamond terkirim!', '💎');
      // Remove pending card if on pending page
      document.getElementById(`pcard-${orderNo}`)?.remove();
      loadDashboard();
    } else {
      showToast(data.message || '❌ Gagal mengirim diamond', '❌');
    }
  } catch {
    showToast('Gagal terhubung', '❌');
  }
}

// Reject / cancel
async function rejectOrder(orderNo) {
  const reason = prompt(`Alasan pembatalan order ${orderNo}:`, 'Pembayaran tidak diterima');
  if (reason === null) return;

  const res  = await fetch('/api/admin/reject', {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ order_no: orderNo, reason }),
  });
  const data = await res.json();
  showToast(data.success ? '✅ Order dibatalkan' : data.message, data.success ? '✅' : '❌');
  if (data.success) {
    document.getElementById(`pcard-${orderNo}`)?.remove();
    loadDashboard();
  }
}

// Check transaction at Apigames
async function checkTrx(orderNo) {
  showToast('⏳ Mengecek di Apigames...', '⏳', 10000);
  const res  = await fetch('/api/admin/check-trx', {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ order_no: orderNo }),
  });
  const data = await res.json();
  if (data.success) {
    const status = data.data?.status || data.data?.trx_status || 'Tidak diketahui';
    showToast(`Status Apigames: ${status}`, '🔍');
  } else {
    showToast(data.message || 'Gagal cek', '⚠️');
  }
}

// Modal close
function closeModal() {
  document.getElementById('modal-overlay').classList.add('hidden');
}
document.getElementById('modal-close').addEventListener('click', closeModal);
document.getElementById('modal-overlay').addEventListener('click', e => {
  if (e.target.id === 'modal-overlay') closeModal();
});

// Expose global functions
window.viewOrder        = viewOrder;
window.confirmOrder     = confirmOrder;
window.rejectOrder      = rejectOrder;
window.checkTrx         = checkTrx;
window.closeModal       = closeModal;
window.showProductModal = showProductModal;
window.saveProduct      = saveProduct;
window.deleteProduct    = deleteProduct;

/* ================================
   HELPERS
   ================================ */
function fmt(n) { return 'Rp ' + parseInt(n || 0).toLocaleString('id-ID'); }

function maskWhatsApp(num) {
  if (!num) return '—';
  const clean = num.trim();
  if (clean.length < 8) return clean;
  // Sensor 4 digit di tengah
  return clean.slice(0, 5) + '****' + clean.slice(-4);
}
