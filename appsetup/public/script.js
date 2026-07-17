// public/script.js — Shared utilities + Customer page logic

/* ================================
   PARTICLES BACKGROUND
   ================================ */
function initParticles() {
  const canvas = document.getElementById('bg-canvas');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  let W, H, parts;

  function resize() { W = canvas.width = innerWidth; H = canvas.height = innerHeight; }

  class P {
    constructor() { this.reset(true); }
    reset(init = false) {
      this.x = Math.random() * W;
      this.y = init ? Math.random() * H : H + 5;
      this.vx = (Math.random() - .5) * .25;
      this.vy = -(Math.random() * .5 + .15);
      this.r = Math.random() * 1.2 + .3;
      this.alpha = Math.random() * .4 + .1;
      this.life = 0;
      this.maxLife = Math.random() * 400 + 200;
      this.color = ['#f59e0b','#fbbf24','#06b6d4','#7c3aed','#ef4444','#22c55e'][Math.floor(Math.random() * 6)];
    }
    update() {
      this.x += this.vx; this.y += this.vy; this.life++;
      this.a = this.alpha * (1 - this.life / this.maxLife);
      if (this.life >= this.maxLife || this.y < -5) this.reset();
    }
    draw() {
      ctx.save(); ctx.globalAlpha = this.a; ctx.fillStyle = this.color;
      ctx.beginPath(); ctx.arc(this.x, this.y, this.r, 0, Math.PI * 2); ctx.fill();
      ctx.restore();
    }
  }

  function init() { resize(); parts = Array.from({length: 90}, () => new P()); }
  function loop() { ctx.clearRect(0, 0, W, H); parts.forEach(p => { p.update(); p.draw(); }); requestAnimationFrame(loop); }

  window.addEventListener('resize', resize);
  init(); loop();
}

/* ================================
   NAVBAR + BTT
   ================================ */
function initNavbar() {
  const nav = document.getElementById('navbar');
  const ham = document.getElementById('hamburger');
  const nl  = document.getElementById('nav-links');
  const btt = document.getElementById('btt');

  window.addEventListener('scroll', () => {
    if (nav) nav.classList.toggle('scrolled', scrollY > 60);
    if (btt) btt.classList.toggle('hidden', scrollY < 400);
  }, { passive: true });

  if (ham) ham.addEventListener('click', () => {
    ham.classList.toggle('open');
    nl.classList.toggle('open');
  });
  if (nl) nl.querySelectorAll('a').forEach(a => a.addEventListener('click', () => {
    ham.classList.remove('open'); nl.classList.remove('open');
  }));

  // Smooth scroll
  document.querySelectorAll('a[href^="#"]').forEach(a => {
    a.addEventListener('click', e => {
      const id = a.getAttribute('href');
      if (id === '#') return;
      const el = document.querySelector(id);
      if (!el) return;
      e.preventDefault();
      const offset = (nav?.offsetHeight || 70) + 8;
      window.scrollTo({ top: el.getBoundingClientRect().top + scrollY - offset, behavior: 'smooth' });
    });
  });
}

function initBtt() {
  const btt = document.getElementById('btt');
  if (btt) btt.addEventListener('click', () => window.scrollTo({ top: 0, behavior: 'smooth' }));
}

/* ================================
   TOAST
   ================================ */
function showToast(msg, icon = 'check-circle', dur = 3000) {
  const t = document.getElementById('toast');
  if (!t) return;
  const iconEl = document.getElementById('toast-icon');
  if (iconEl) {
    iconEl.innerHTML = `<i data-lucide="${icon}"></i>`;
    if (window.lucide) window.lucide.createIcons();
  }
  document.getElementById('toast-msg').textContent = msg;
  t.classList.remove('hidden');
  clearTimeout(t._timer);
  t._timer = setTimeout(() => t.classList.add('hidden'), dur);
}

/* ================================
   FAQ ACCORDION
   ================================ */
function initFaq() {
  document.querySelectorAll('.faq-item').forEach(item => {
    item.querySelector('.faq-q')?.addEventListener('click', () => {
      const open = item.classList.contains('open');
      document.querySelectorAll('.faq-item').forEach(i => i.classList.remove('open'));
      if (!open) item.classList.add('open');
    });
  });
}

/* ================================
   FALLBACK PACKAGES (saat API gagal)
   ================================ */
const FALLBACK_PACKAGES = [
  { product_id:'AGML022',  product_name:'22 Diamond',    price:5900,   diamond:22,   bonus:'', cat:'starter' },
  { product_id:'AGML056',  product_name:'56 Diamond',    price:13900,  diamond:56,   bonus:'', cat:'starter' },
  { product_id:'AGML086',  product_name:'86 Diamond',    price:21900,  diamond:86,   bonus:'', cat:'starter' },
  { product_id:'AGML172',  product_name:'172 Diamond',   price:43900,  diamond:172,  bonus:'', cat:'starter' },
  { product_id:'AGML257',  product_name:'257 Diamond',   price:64900,  diamond:257,  bonus:'+20%', cat:'mid' },
  { product_id:'AGML344',  product_name:'344 Diamond',   price:84900,  diamond:344,  bonus:'+20%', cat:'mid' },
  { product_id:'AGML514',  product_name:'514 Diamond',   price:124900, diamond:514,  bonus:'+20%', cat:'mid' },
  { product_id:'AGML600',  product_name:'600 Diamond',   price:144900, diamond:600,  bonus:'+25%', cat:'mid' },
  { product_id:'AGML878',  product_name:'878 Diamond',   price:209900, diamond:878,  bonus:'+25%', cat:'pro' },
  { product_id:'AGML1195', product_name:'1195 Diamond',  price:279900, diamond:1195, bonus:'+30%', cat:'pro' },
  { product_id:'AGML2010', product_name:'2010 Diamond',  price:469900, diamond:2010, bonus:'+35%', cat:'whale' },
  { product_id:'AGML3688', product_name:'3688 Diamond',  price:849900, diamond:3688, bonus:'+40%', cat:'whale' },
  { product_id:'AGML5532', product_name:'5532 Diamond',  price:1249900,diamond:5532, bonus:'+40%', cat:'whale' },
];

function extractDiamondGlobal(name = '') {
  const m = name.match(/(\d[\d,\.]*)/);
  return m ? parseInt(m[1].replace(/[,\.]/g, '')) : 0;
}

function categorize(pkg) {
  const diamond = pkg.diamond !== undefined ? pkg.diamond : extractDiamondGlobal(pkg.product_name || '');
  
  const nameLower = (pkg.product_name || '').toLowerCase();
  const isSub = !diamond || diamond === 0 || 
                nameLower.includes('pass') || 
                nameLower.includes('weekly') || 
                nameLower.includes('monthly') || 
                nameLower.includes('starlight') || 
                nameLower.includes('membership') || 
                nameLower.includes('premium');
                
  if (isSub) {
    return 'subscription';
  }
  const p = pkg.price || 0;
  if (p <= 50000)  return 'starter';
  if (p <= 150000) return 'mid';
  if (p <= 400000) return 'pro';
  return 'whale';
}

/* ================================
   CUSTOMER PAGE LOGIC
   ================================ */
document.addEventListener('DOMContentLoaded', () => {
  initNavbar(); initParticles(); initFaq(); initBtt();
  if (window.lucide) window.lucide.createIcons();

  // Only run customer logic if on main page
  if (!document.getElementById('panel-combined')) return;

  // State
  let s = {
    step: 1,
    selectedGame: 'mobilelegends',
    selectedCategory: 'AGML',
    playerId: '', serverId: '', playerName: '',
    selectedPkg: null,
    packages: [],
    pkgFilter: 'all',
    paymentMethod: 'bank_transfer',
  };

  const GAMES_CONFIG = {
    mobilelegends: {
      category: 'AGML',
      title: 'Mobile Legends',
      idLabel: 'User ID',
      idPlaceholder: 'Masukkan User ID',
      idHint: 'Lihat di profil game Anda',
      hasServer: true,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'User ID tidak valid';
        if (!srv || srv.length < 1) return 'Server ID / Zone tidak valid';
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    freefire: {
      category: 'AGFF',
      title: 'Free Fire',
      idLabel: 'Player ID',
      idPlaceholder: 'Masukkan Player ID',
      idHint: 'Lihat di profil game Anda',
      hasServer: false,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'Player ID tidak valid';
        
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    honorofkings: {
      category: 'AGHOK',
      title: 'Honor of Kings',
      idLabel: 'User ID',
      idPlaceholder: 'Masukkan User ID',
      idHint: 'Lihat di profil game Anda',
      hasServer: true,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'User ID tidak valid';
        if (!srv || srv.length < 1) return 'Server ID / Zone tidak valid';
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    pubg: {
      category: 'AGPUBG',
      title: 'PUBG Mobile',
      idLabel: 'Player ID',
      idPlaceholder: 'Masukkan Player ID',
      idHint: 'Lihat di profil game Anda',
      hasServer: false,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'Player ID tidak valid';
        
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    valorant: {
      category: 'AGVALO',
      title: 'Valorant',
      idLabel: 'Player ID',
      idPlaceholder: 'Masukkan Player ID',
      idHint: 'Lihat di profil game Anda',
      hasServer: false,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'Player ID tidak valid';
        
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    genshin: {
      category: 'AGGI',
      title: 'Genshin Impact',
      idLabel: 'UID',
      idPlaceholder: 'Masukkan UID',
      idHint: 'Lihat di profil game Anda',
      hasServer: true,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'UID tidak valid';
        if (!srv || srv.length < 1) return 'Server ID / Zone tidak valid';
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    starrail: {
      category: 'AGHSR',
      title: 'Honkai: Star Rail',
      idLabel: 'UID',
      idPlaceholder: 'Masukkan UID',
      idHint: 'Lihat di profil game Anda',
      hasServer: true,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'UID tidak valid';
        if (!srv || srv.length < 1) return 'Server ID / Zone tidak valid';
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    codm: {
      category: 'AGCODM',
      title: 'CODM',
      idLabel: 'Player ID',
      idPlaceholder: 'Masukkan Player ID',
      idHint: 'Lihat di profil game Anda',
      hasServer: false,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'Player ID tidak valid';
        
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    aov: {
      category: 'AGAOV',
      title: 'Arena of Valor',
      idLabel: 'User ID',
      idPlaceholder: 'Masukkan User ID',
      idHint: 'Lihat di profil game Anda',
      hasServer: true,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'User ID tidak valid';
        if (!srv || srv.length < 1) return 'Server ID / Zone tidak valid';
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    pointblank: {
      category: 'AGPBL',
      title: 'Point Blank',
      idLabel: 'Player ID',
      idPlaceholder: 'Masukkan Player ID',
      idHint: 'Lihat di profil game Anda',
      hasServer: false,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'Player ID tidak valid';
        
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    toweroffantasy: {
      category: 'AGTOF',
      title: 'Tower of Fantasy',
      idLabel: 'UID',
      idPlaceholder: 'Masukkan UID',
      idHint: 'Lihat di profil game Anda',
      hasServer: true,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'UID tidak valid';
        if (!srv || srv.length < 1) return 'Server ID / Zone tidak valid';
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    hago: {
      category: 'AGHAGO',
      title: 'Hago',
      idLabel: 'Player ID',
      idPlaceholder: 'Masukkan Player ID',
      idHint: 'Lihat di profil game Anda',
      hasServer: false,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'Player ID tidak valid';
        
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    roblox: {
      category: 'AGRBLX',
      title: 'Roblox',
      idLabel: 'Nomor HP / Email',
      idPlaceholder: 'Masukkan Nomor HP / Email',
      idHint: 'Kode voucher akan dikirimkan ke kontak ini',
      hasServer: false,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'Nomor HP / Email tidak valid';
        
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    steam: {
      category: 'AGSTM',
      title: 'Steam Wallet',
      idLabel: 'Nomor HP / Email',
      idPlaceholder: 'Masukkan Nomor HP / Email',
      idHint: 'Kode voucher akan dikirimkan ke kontak ini',
      hasServer: false,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'Nomor HP / Email tidak valid';
        
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
    garena: {
      category: 'AGGS',
      title: 'Garena Shells',
      idLabel: 'Nomor HP / Email',
      idPlaceholder: 'Masukkan Nomor HP / Email',
      idHint: 'Kode voucher akan dikirimkan ke kontak ini',
      hasServer: false,
      hasServerSelect: false,
      validation: (uid, srv) => {
        if (!uid || uid.length < 3) return 'Nomor HP / Email tidak valid';
        
        return null;
      },
      theme: {
        accent: '#8b5cf6',
        accentL: '#a78bfa',
        accentD: '#6d28d9',
        glow: 'rgba(139, 92, 246, 0.4)',
        gradient: 'linear-gradient(135deg, #6d28d9, #3b82f6)'
      }
    },
  };

  // Handle game card selection
  document.querySelectorAll('.game-card').forEach(card => {
    card.addEventListener('click', () => {
      document.querySelectorAll('.game-card').forEach(c => c.classList.remove('active'));
      card.classList.add('active');

      const gameKey = card.dataset.game;
      s.selectedGame = gameKey;
      s.selectedCategory = card.dataset.category;

      // Update theme and UI inputs
      applyGameConfig(gameKey);
      
      // Reset step 1
      resetStep1();
      
      // Load packages for selected game
      loadPackages();

      // Smooth scroll to top-up form
      const topup = document.getElementById('topup');
      if (topup) {
        const off = document.getElementById('navbar')?.offsetHeight || 70;
        window.scrollTo({ top: topup.getBoundingClientRect().top + scrollY - off - 8, behavior: 'smooth' });
      }
    });
  });

  function applyGameConfig(gameKey) {
    const config = GAMES_CONFIG[gameKey];
    if (!config) return;

    // 1. Update text and titles
    document.getElementById('form-game-title').textContent = config.title;
    document.getElementById('lbl-uid').textContent = config.idLabel + ' *';
    document.getElementById('inp-uid').placeholder = config.idPlaceholder;
    document.getElementById('hint-uid').textContent = config.idHint;

    // 2. Manage visibility of server inputs
    const fgSrv = document.getElementById('fg-srv');
    const fgSrvSelect = document.getElementById('fg-srv-select');

    if (config.hasServer) {
      fgSrv.classList.remove('hidden');
    } else {
      fgSrv.classList.add('hidden');
    }

    if (config.hasServerSelect) {
      fgSrvSelect.classList.remove('hidden');
    } else {
      fgSrvSelect.classList.add('hidden');
    }

    // 3. Dynamic colors via root elements
    document.documentElement.style.setProperty('--gold', config.theme.accent);
    document.documentElement.style.setProperty('--gold-l', config.theme.accentL);
    document.documentElement.style.setProperty('--gold-d', config.theme.accentD);
    document.documentElement.style.setProperty('--gold-gw', config.theme.glow);
    document.documentElement.style.setProperty('--grad-g', config.theme.gradient);

    // 4. Update Game Header (Left Column)
    const ghTitle = document.getElementById('gh-title');
    if (ghTitle) ghTitle.textContent = config.title;
    
    const activeCard = document.querySelector('.game-card.active');
    if (activeCard) {
      const cardBanner = activeCard.querySelector('.game-banner');
      const cardLogo = activeCard.querySelector('img');
      const cardSub = activeCard.querySelector('.game-sub');
      
      const ghBanner = document.getElementById('gh-banner');
      const ghLogo = document.getElementById('gh-logo');
      const ghDesc = document.getElementById('gh-desc');
      
      if (cardBanner && ghBanner) ghBanner.style.background = cardBanner.style.background;
      if (cardLogo && ghLogo) {
        ghLogo.src = cardLogo.src;
        ghLogo.className = cardLogo.className;
      }
      if (cardSub && ghDesc) ghDesc.textContent = cardSub.textContent;
    }

    // Refresh saved accounts for this game if member is logged in
    if (memberUser) {
      loadSavedAccounts(memberUser.username);
    }
  }

  function resetStep1() {
    vCard.classList.add('hidden');
    btnS1Nxt.disabled = true;
    s.playerId = ''; s.serverId = ''; s.playerName = ''; s.selectedPkg = null; s.packages = [];
    s.paymentMethod = 'bank_transfer';
    document.querySelectorAll('.payment-method-card').forEach(c => c.classList.toggle('active', c.dataset.method === 'bank_transfer'));
    uidInp.value = ''; srvInp.value = '';
    document.getElementById('inp-srv-select').selectedIndex = 0;
  }

  /* ---- Step helpers ---- */
  
  function showInvoice() {
    document.getElementById('panel-combined').classList.remove('active');
    document.getElementById('panel-invoice').classList.add('active');
  }
  function resetForm() {
    document.getElementById('panel-invoice').classList.remove('active');
    document.getElementById('panel-combined').classList.add('active');
    const topup = document.getElementById('topup');
    if (topup) setTimeout(() => {
      const off = document.getElementById('navbar')?.offsetHeight || 70;
      window.scrollTo({ top: topup.getBoundingClientRect().top + scrollY - off - 8, behavior: 'smooth' });
    }, 100);
  }

  /* ---- STEP 1: Verifikasi ID ---- */
  const uidInp   = document.getElementById('inp-uid');
  const srvInp   = document.getElementById('inp-srv');
  const btnVer   = document.getElementById('btn-verify');
  const btnS1Nxt = document.getElementById('btn-s1-next') || document.createElement('button');
  const vCard    = document.getElementById('verified-card');
  const vcName   = document.getElementById('vc-name');
  const vcId     = document.getElementById('vc-id');

  btnVer.addEventListener('click', async () => {
    const config = GAMES_CONFIG[s.selectedGame];
    if (!config) return;

    let uid = uidInp.value.trim();
    let srv = '';

    if (config.hasServer) {
      srv = srvInp.value.trim();
    } else if (config.hasServerSelect) {
      srv = document.getElementById('inp-srv-select').value;
    }

    // Run client validation
    const err = config.validation(uid, srv);
    if (err) {
      showToast(err, 'alert-triangle'); return;
    }

    document.getElementById('verify-txt').textContent = '⏳ Memverifikasi...';
    btnVer.disabled = true;

    try {
      const res  = await fetch(`/api/check-user?game=${s.selectedGame}&user_id=${encodeURIComponent(uid)}&server_id=${encodeURIComponent(srv)}`);
      const data = await res.json();

      if (data.success) {
        s.playerId = uid; s.serverId = srv; s.playerName = data.nickname;
        vcName.textContent = data.nickname;
        vcId.textContent   = srv ? `ID: ${uid} | Server ${srv}` : `ID: ${uid}`;
        vCard.classList.remove('hidden');
        btnS1Nxt.disabled = false;

        if (data.bypass) {
          vCard.style.borderColor = 'rgba(245,158,11,0.4)';
          document.querySelector('.badge-ok').innerHTML = '<i data-lucide="alert-triangle" class="icon-inline"></i> Lanjutkan — pastikan ID benar!';
          document.querySelector('.badge-ok').style.background = 'rgba(245,158,11,.15)';
          document.querySelector('.badge-ok').style.color = '#f59e0b';
        } else {
          vCard.style.borderColor = '';
          document.querySelector('.badge-ok').innerHTML = '<i data-lucide="check-circle" class="icon-inline"></i> Akun Ditemukan';
          document.querySelector('.badge-ok').style = '';
          showToast(`Akun "${data.nickname}" ditemukan!`, 'check-circle');
        }
        if (window.lucide) window.lucide.createIcons();
      } else {
        showToast(data.message || 'ID tidak ditemukan. Periksa kembali.', 'x-circle');
        vCard.classList.add('hidden');
        btnS1Nxt.disabled = true;
      }
    } catch {
      showToast('Gagal terhubung ke server', 'x-circle');
    } finally {
      document.getElementById('verify-txt').textContent = '🔍 Verifikasi ID Player';
      btnVer.disabled = false;
    }
  });

  // Enter to verify
  uidInp.addEventListener('keypress', e => { if (e.key === 'Enter') btnVer.click(); });

  // Change ID
  document.getElementById('vc-change').addEventListener('click', () => {
    resetStep1();
  });

  // How to find ID
  document.getElementById('howid-btn').addEventListener('click', () => {
    document.getElementById('howid-list').classList.toggle('hidden');
  });

  // Step 1 Next
  btnS1Nxt.addEventListener('click', () => {
    goStep(2);
    loadPackages();
  });

  /* ---- STEP 2: Pilih Paket ---- */
  const pkgGrid  = document.getElementById('pkg-grid');
  const pkgLoad  = document.getElementById('pkg-loading');
  const pkgErr   = document.getElementById('pkg-error');
  const selPkg   = document.getElementById('sel-pkg');
  const btnS2Bk  = document.getElementById('btn-s2-back') || document.createElement('button');
  const btnS2Nxt = document.getElementById('btn-submit-order') || document.getElementById('btn-s2-next') || document.createElement('button');

  async function loadPackages() {
    if (s.packages.length > 0) { renderPkgs(); return; }
    pkgLoad.classList.remove('hidden');
    pkgGrid.innerHTML = '';

    try {
      const res  = await fetch(`/api/products?category=${s.selectedCategory}&t=${Date.now()}`);
      const data = await res.json();

      if (data.success && data.products?.length > 0) {
        s.packages = data.products.map(p => ({
          ...p,
          diamond: extractDiamondGlobal(p.product_name),
          cat: categorize(p),
        }));
      } else {
        throw new Error('empty');
      }
    } catch {
      pkgErr.classList.remove('hidden');
      if (s.selectedCategory === 'AGML') {
        s.packages = FALLBACK_PACKAGES;
      } else {
        s.packages = [];
      }
    } finally {
      pkgLoad.classList.add('hidden');
      updateFilterButtons();
      renderPkgs();
      updateCheckoutSummary();
    }
  }

  function extractDiamondGlobal(name = '') {
    const m = name.match(/(\d[\d,\.]*)/);
    return m ? parseInt(m[1].replace(/[,\.]/g, '')) : 0;
  }

  function updateFilterButtons() {
    const catsPresent = new Set(s.packages.map(p => p.cat));
    document.querySelectorAll('.pfp').forEach(btn => {
      const filter = btn.dataset.cat;
      if (filter === 'all') return;
      if (catsPresent.has(filter)) {
        btn.classList.remove('hidden');
      } else {
        btn.classList.add('hidden');
      }
    });
  }

  function updateCheckoutSummary() {
    const summaryItem = document.getElementById('summary-item');
    const summaryPayment = document.getElementById('summary-payment');
    const checkoutTotal = document.getElementById('checkout-total-price');

    if (summaryItem) {
      if (s.selectedPkg) {
        const isPass = !s.selectedPkg.diamond || s.selectedPkg.diamond === 0;
        summaryItem.textContent = isPass ? s.selectedPkg.product_name : `💎 ${(s.selectedPkg.diamond||0).toLocaleString('id-ID')} Diamond`;
      } else {
        summaryItem.textContent = '—';
      }
    }

    if (summaryPayment) {
      if (s.paymentMethod === 'bank_transfer') {
        summaryPayment.textContent = '🏦 Transfer Bank BCA';
      } else if (s.paymentMethod === 'qris') {
        summaryPayment.textContent = '🛡️ QRIS (E-Wallet & Bank)';
      } else {
        summaryPayment.textContent = '—';
      }
    }

    if (checkoutTotal) {
      if (s.selectedPkg) {
        checkoutTotal.textContent = fmt(s.selectedPkg.price);
      } else {
        checkoutTotal.textContent = 'Rp 0';
      }
    }
  }

  function renderPkgs() {
    const filtered = s.pkgFilter === 'all' ? s.packages : s.packages.filter(p => p.cat === s.pkgFilter);
    pkgGrid.innerHTML = filtered.map(pkg => {
      const isSel = s.selectedPkg?.product_id === pkg.product_id;
      const badgeTxt = pkg.product_name?.toLowerCase().includes('1195') || pkg.product_name?.toLowerCase().includes('514') ? 'POPULAR' : '';
      const badgeCls = badgeTxt === 'HOT' ? 'hot' : badgeTxt === 'POPULAR' ? 'pop' : '';
      
      // If product has 0 diamonds (like passes, memberships, etc.), show the cleaned product name
      const rawName = pkg.product_name || pkg.product_id;
      const isPass = !pkg.diamond || pkg.diamond === 0;
      const displayAmt = isPass ? rawName : pkg.diamond.toLocaleString('id-ID');
      const gemIcon = isPass ? '🎫' : '💎';

      return `
        <div class="pkg-card ${isSel ? 'selected' : ''}"
             onclick="selectPkg('${pkg.product_id}')"
             id="pk-${pkg.product_id.replace(/[^a-z0-9]/gi,'')}"
             title="${pkg.product_name}">
          ${badgeTxt ? `<div class="pkg-badge ${badgeCls}">${badgeTxt}</div>` : ''}
          <div class="pkg-gem">${gemIcon}</div>
          <div class="pkg-info-col">
            <div class="pkg-amt ${isPass ? 'pass-text' : ''}">${displayAmt}</div>
            <div class="pkg-price">${fmt(pkg.price)}</div>
          </div>
          ${pkg.bonus && pkg.bonus !== '\u00a0' ? `<div class="pkg-bonus">${pkg.bonus}</div>` : ''}
        </div>`;
    }).join('');
  }

  // Expose for inline onclick
  window.selectPkg = function(id) {
    s.selectedPkg = s.packages.find(p => p.product_id === id) || null;
    renderPkgs();
    if (s.selectedPkg) {
      selPkg.classList.remove('hidden');
      const isPass = !s.selectedPkg.diamond || s.selectedPkg.diamond === 0;
      const displayInfo = isPass ? s.selectedPkg.product_name : `💎 ${(s.selectedPkg.diamond||0).toLocaleString('id-ID')} Diamond${s.selectedPkg.bonus ? ' ' + s.selectedPkg.bonus : ''}`;
      document.getElementById('sel-pkg-info').textContent = displayInfo;
      document.getElementById('sel-pkg-price').textContent = fmt(s.selectedPkg.price);
      checkStep2Next();
    } else {
      selPkg.classList.add('hidden');
      btnS2Nxt.disabled = true;
    }
  };

  function checkStep2Next() {
    const wa = document.getElementById('inp-wa')?.value.trim();
    btnS2Nxt.disabled = !s.selectedPkg || !wa || !/^0\d{9,12}$/.test(wa);
  }

  document.getElementById('inp-wa').addEventListener('input', checkStep2Next);
  
  // Load initial packages for default game (Mobile Legends) on page load
  loadPackages();

  // Payment Method Selection
  document.querySelectorAll('.payment-method-card').forEach(card => {
    card.addEventListener('click', () => {
      document.querySelectorAll('.payment-method-card').forEach(c => c.classList.remove('active'));
      card.classList.add('active');
      s.paymentMethod = card.dataset.method;
      updateCheckoutSummary();
    });
  });

  document.querySelectorAll('.pfp').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.pfp').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      s.pkgFilter = btn.dataset.cat;
      renderPkgs();
    });
  });

  

  btnS2Nxt.addEventListener('click', async () => {
    if (!s.selectedPkg) { showToast('Pilih nominal top-up dahulu', 'alert-triangle'); return; }
    const wa = document.getElementById('inp-wa').value.trim();
    if (!wa || !/^0\d{9,12}$/.test(wa)) { showToast('Nomor WhatsApp tidak valid', 'alert-triangle'); return; }

    btnS2Nxt.disabled = true;
    btnS2Nxt.textContent = '⏳ Membuat order...';

    try {
      const res  = await fetch('/api/order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          player_id:      s.playerId,
          server_id:      s.serverId,
          player_name:    s.playerName,
          product_id:     s.selectedPkg.product_id,
          product_name:   s.selectedPkg.product_name,
          diamond:        s.selectedPkg.diamond,
          price:          s.selectedPkg.price,
          whatsapp:       wa,
          payment_method: s.paymentMethod || 'bank_transfer',
        }),
      });
      const data = await res.json();

      if (!data.success) {
        showToast(data.message || 'Gagal membuat order', 'x-circle');
        return;
      }

      fillPaymentInstructions(data.order, data.whatsapp_url);
      showInvoice();
      showToast('Pesanan berhasil dibuat!', 'check-circle');
    } catch {
      showToast('Gagal terhubung ke server', 'x-circle');
    } finally {
      btnS2Nxt.disabled = false;
      btnS2Nxt.textContent = 'Pesan Sekarang →';
    }
  });

  /* ---- STEP 3: Instruksi Bayar ---- */
  function fillPaymentInstructions(order, waUrl) {
    currentOrderNo = order.order_no;
    document.getElementById('pi-order-no').textContent = order.order_no;
    document.getElementById('pi-player').textContent   = `${order.player_name || '—'} (ID: ${order.player_id})`;
    document.getElementById('pi-package').textContent  = `${order.product_name} (${order.diamond})`;

    // Toggle QRIS vs Bank UI details
    const bankRow = document.getElementById('pi-bank-row');
    const qrisRow = document.getElementById('pi-qris-row');
    const noteText = document.getElementById('pi-note-text');

    if (order.payment_method === 'qris') {
      bankRow.classList.add('hidden');
      qrisRow.classList.remove('hidden');
      noteText.innerHTML = `⚠️ Pindai QRIS di atas lalu bayar <strong>tepat</strong> sesuai nominal di bawah ini agar order terverifikasi otomatis.`;
      
      try {
        const dynamicQRIS = generateDynamicQRIS(order.total_bayar);
        document.getElementById('pi-qris-img').src = `https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=${encodeURIComponent(dynamicQRIS)}`;
      } catch (e) {
        document.getElementById('pi-qris-img').src = `/images/qris.jpg`;
      }
    } else {
      bankRow.classList.remove('hidden');
      qrisRow.classList.add('hidden');
      noteText.innerHTML = `⚠️ Transfer <strong>tepat</strong> sesuai nominal di atas (termasuk kode unik 3 digit terakhir) agar order terverifikasi otomatis.`;
    }

    document.getElementById('pi-bank').innerHTML       =
      `<strong>${order.bank_name}</strong><br/>${order.bank_account}<br/>a/n ${order.bank_holder}`;
    document.getElementById('pi-total').textContent    = fmt(order.total_bayar);
    document.getElementById('btn-wa').href             = waUrl;

    // Sesi saldo check
    const saldoPayContainer = document.getElementById('saldo-pay-container');
    if (memberUser && memberUser.saldo >= order.total_bayar) {
      saldoPayContainer.classList.remove('hidden');
    } else {
      saldoPayContainer.classList.add('hidden');
    }

    // Copy buttons
    setupCopy('copy-order-no', order.order_no, 'Nomor pesanan berhasil disalin');
    setupCopy('copy-account', order.bank_account, 'Nomor rekening berhasil disalin');
    setupCopy('copy-total', String(order.total_bayar), 'Nominal transfer berhasil disalin');
  }

  function setupCopy(btnId, text, msg) {
    const btn = document.getElementById(btnId);
    if (!btn) return;
    btn.addEventListener('click', () => {
      navigator.clipboard.writeText(text).then(() => showToast(msg, 'copy'));
    });
  }

  // ============================================================
  //  MEMBER LOGIN / REGISTER & PROFILE LOGIC
  // ============================================================
  let memberUser = null;
  let currentOrderNo = null;

  async function checkMemberSession() {
    try {
      const res = await fetch('/api/member/profile');
      const data = await res.json();
      if (data.success && data.user) {
        memberUser = data.user;
        document.getElementById('nav-member-btn').classList.add('hidden');
        
        const navProfile = document.getElementById('nav-member-profile');
        navProfile.classList.remove('hidden');
        document.getElementById('nav-username').textContent = memberUser.username;

        // Tampilkan widget di step 1
        const widget = document.getElementById('member-widget');
        widget.classList.remove('hidden');
        document.getElementById('widget-username').textContent = memberUser.username;
        document.getElementById('widget-saldo').textContent = fmt(memberUser.saldo);

        // Tampilkan tombol simpan akun
        document.getElementById('vc-save-acc').classList.remove('hidden');

        loadSavedAccounts(memberUser.username);
      } else {
        clearMemberSessionUI();
      }
    } catch {
      clearMemberSessionUI();
    }
  }

  function clearMemberSessionUI() {
    memberUser = null;
    document.getElementById('nav-member-btn').classList.remove('hidden');
    
    const navProfile = document.getElementById('nav-member-profile');
    navProfile.classList.add('hidden');

    const widget = document.getElementById('member-widget');
    widget.classList.add('hidden');

    document.getElementById('vc-save-acc').classList.add('hidden');
  }

  function loadSavedAccounts(username) {
    const list = document.getElementById('saved-accounts-list');
    if (!list) return;
    
    const key = 'saved_accounts_' + username;
    const accounts = JSON.parse(localStorage.getItem(key) || '[]');
    
    if (accounts.length === 0) {
      list.innerHTML = '<span style="font-size:0.8rem;color:var(--muted)">Belum ada akun tersimpan.</span>';
      return;
    }

    list.innerHTML = accounts.map(acc => `
      <button class="btn-acc" style="background:rgba(255,255,255,0.05);border:1px solid rgba(255,255,255,0.1);color:#fff;padding:4px 8px;border-radius:4px;font-size:0.75rem;cursor:pointer" data-id="${acc.id}" data-srv="${acc.server}">
        👤 ${acc.name} (${acc.id})
      </button>
    `).join('');

    // Bind click events
    list.querySelectorAll('.btn-acc').forEach(btn => {
      btn.addEventListener('click', () => {
        uidInp.value = btn.dataset.id;
        srvInp.value = btn.dataset.srv;
        btnVer.click(); // Langsung verifikasi
      });
    });
  }

  // Simpan akun terverifikasi ke profil
  document.getElementById('vc-save-acc').addEventListener('click', () => {
    if (!memberUser || !s.playerId) return;
    
    const name = prompt('Masukkan nama label untuk akun ini (contoh: Akun Utama):');
    if (!name || name.trim() === '') return;

    const username = memberUser.username;
    const key = 'saved_accounts_' + username;
    const accounts = JSON.parse(localStorage.getItem(key) || '[]');

    // Cek apakah sudah ada
    const exists = accounts.some(acc => acc.id === s.playerId && acc.server === s.serverId);
    if (exists) {
      showToast('Akun ini sudah disimpan!', 'alert-triangle');
      return;
    }

    accounts.push({
      name: name.trim(),
      id: s.playerId,
      server: s.serverId
    });

    localStorage.setItem(key, JSON.stringify(accounts));
    loadSavedAccounts(username);
    showToast('Akun berhasil disimpan!', 'check-circle');
  });

  // Modal handlers
  const memberOverlay = document.getElementById('member-modal-overlay');
  const loginBox = document.getElementById('member-login-box');
  const regBox = document.getElementById('member-register-box');
  const profBox = document.getElementById('member-profile-box');

  document.getElementById('nav-member-btn').addEventListener('click', (e) => {
    e.preventDefault();
    memberOverlay.classList.remove('hidden');
    loginBox.classList.remove('hidden');
    regBox.classList.add('hidden');
    profBox.classList.add('hidden');
  });

  // Click on Profile username to open dashboard modal
  document.getElementById('nav-username').addEventListener('click', async (e) => {
    e.preventDefault();
    if (!memberUser) return;

    // Ambil info profile & transaksi terupdate
    try {
      const res = await fetch('/api/member/profile');
      const data = await res.json();
      if (data.success && data.user) {
        memberUser = data.user;
        document.getElementById('prof-username').textContent = memberUser.username;
        document.getElementById('prof-email').textContent = memberUser.email;
        document.getElementById('prof-whatsapp').textContent = memberUser.whatsapp;
        document.getElementById('prof-saldo').textContent = fmt(memberUser.saldo);
      }
    } catch {}

    memberOverlay.classList.remove('hidden');
    loginBox.classList.add('hidden');
    regBox.classList.add('hidden');
    profBox.classList.remove('hidden');

    // Ambil riwayat order
    renderMemberOrders();
  });

  async function renderMemberOrders() {
    const tbody = document.getElementById('member-orders-table-body');
    tbody.innerHTML = '<tr><td colspan="4" style="text-align:center;padding:20px;color:var(--muted)">Memuat riwayat...</td></tr>';
    
    try {
      const res = await fetch('/api/member/orders');
      const data = await res.json();
      if (data.success && data.orders?.length > 0) {
        tbody.innerHTML = data.orders.map(o => `
          <tr style="border-bottom:1px solid rgba(255,255,255,0.05)">
            <td style="padding:8px;font-family:monospace">${o.order_no}</td>
            <td style="padding:8px">${o.product_name}</td>
            <td style="padding:8px;color:var(--gold)">${fmt(o.total_bayar)}</td>
            <td style="padding:8px"><span class="status-badge status-${o.status.toLowerCase()}">${o.status}</span></td>
          </tr>
        `).join('');
      } else {
        tbody.innerHTML = '<tr><td colspan="4" style="text-align:center;padding:20px;color:var(--muted)">Belum ada transaksi.</td></tr>';
      }
    } catch {
      tbody.innerHTML = '<tr><td colspan="4" style="text-align:center;padding:20px;color:var(--red)">Gagal memuat riwayat.</td></tr>';
    }
  }

  // Switch between Login and Register views
  document.getElementById('switch-to-register').addEventListener('click', (e) => {
    e.preventDefault();
    loginBox.classList.add('hidden');
    regBox.classList.remove('hidden');
  });

  document.getElementById('switch-to-login').addEventListener('click', (e) => {
    e.preventDefault();
    regBox.classList.add('hidden');
    loginBox.classList.remove('hidden');
  });

  // Close buttons
  const closes = ['login-close', 'register-close', 'profile-close'];
  closes.forEach(id => {
    document.getElementById(id).addEventListener('click', () => {
      memberOverlay.classList.add('hidden');
    });
  });

  // Login submit
  document.getElementById('btn-member-login-submit').addEventListener('click', async () => {
    const identity = document.getElementById('login-identity').value.trim();
    const password = document.getElementById('login-password').value;
    const errorEl = document.getElementById('login-error-msg');

    if (!identity || !password) {
      errorEl.textContent = 'Username/Email dan Password wajib diisi';
      return;
    }

    errorEl.textContent = 'Memproses...';

    try {
      const res = await fetch('/api/member/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ identity, password })
      });
      const data = await res.json();
      if (data.success) {
        showToast('Login Berhasil!', 'check-circle');
        memberOverlay.classList.add('hidden');
        checkMemberSession();
      } else {
        errorEl.textContent = data.message || 'Login gagal';
      }
    } catch {
      errorEl.textContent = 'Gagal menghubungi server';
    }
  });

  // Register submit
  document.getElementById('btn-member-register-submit').addEventListener('click', async () => {
    const username = document.getElementById('reg-username').value.trim();
    const email = document.getElementById('reg-email').value.trim();
    const whatsapp = document.getElementById('reg-whatsapp').value.trim();
    const password = document.getElementById('reg-password').value;
    const errorEl = document.getElementById('register-error-msg');

    if (!username || !email || !whatsapp || !password) {
      errorEl.textContent = 'Semua kolom wajib diisi';
      return;
    }

    if (password.length < 6) {
      errorEl.textContent = 'Password minimal terdiri dari 6 karakter';
      return;
    }

    errorEl.textContent = 'Mendaftar...';

    try {
      const res = await fetch('/api/member/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, email, whatsapp, password })
      });
      const data = await res.json();
      if (data.success) {
        showToast('Pendaftaran Berhasil!', 'check-circle');
        regBox.classList.add('hidden');
        loginBox.classList.remove('hidden');
        document.getElementById('login-identity').value = username;
      } else {
        errorEl.textContent = data.message || 'Pendaftaran gagal';
      }
    } catch {
      errorEl.textContent = 'Gagal menghubungi server';
    }
  });

  // Logout click
  document.getElementById('nav-logout-btn').addEventListener('click', async () => {
    try {
      const res = await fetch('/api/member/logout', { method: 'POST' });
      const data = await res.json();
      if (data.success) {
        showToast('Logout Berhasil', 'check-circle');
        clearMemberSessionUI();
      }
    } catch {}
  });

  // Pay with saldo submit
  document.getElementById('btn-pay-saldo').addEventListener('click', async () => {
    if (!currentOrderNo) return;
    
    if (!confirm('Konfirmasi pembayaran menggunakan Saldo Akun? Sisa saldo Anda akan berkurang secara otomatis.')) return;
    
    const btn = document.getElementById('btn-pay-saldo');
    btn.disabled = true;
    btn.textContent = '⏳ Memproses...';

    try {
      const formData = new FormData();
      formData.append('order_no', currentOrderNo);

      const res = await fetch('/api/member/pay-with-saldo', {
        method: 'POST',
        body: formData
      });
      const data = await res.json();
      if (data.success) {
        showToast('Pembayaran Saldo Berhasil!', 'check-circle');
        document.getElementById('saldo-pay-container').classList.add('hidden');
        
        // Update saldo UI
        if (memberUser) {
          memberUser.saldo = data.saldo;
          document.getElementById('widget-saldo').textContent = fmt(memberUser.saldo);
        }

        // Tampilkan status lunas di layar
        document.getElementById('pi-total').innerHTML = 
          document.getElementById('pi-total').textContent + ' <br/><span style="background:var(--green);color:#fff;font-size:0.75rem;padding:2px 6px;border-radius:4px;display:inline-block;margin-top:4px">LUNAS (Saldo)</span>';
        
        // Ganti button WA menjadi Cek Status
        const btnWa = document.getElementById('btn-wa');
        btnWa.href = '/status';
        btnWa.innerHTML = '<i data-lucide="clipboard-list" class="btn-icon"></i> Cek Status Pesanan';
        btnWa.style.background = 'var(--gold)';
        btnWa.style.color = '#000';
        if (window.lucide) window.lucide.createIcons();
      } else {
        showToast(data.message || 'Gagal membayar dengan saldo', 'x-circle');
      }
    } catch {
      showToast('Gagal terhubung ke server', 'x-circle');
    } finally {
      btn.disabled = false;
      btn.textContent = 'Bayar Pakai Saldo (Instan)';
    }
  });

  // Check login session on load
  checkMemberSession();
});

/* ================================
   HELPERS
   ================================ */
function fmt(n) {
  return 'Rp ' + parseInt(n || 0).toLocaleString('id-ID');
}

/* ================================
   DYNAMIC QRIS GENERATOR
   ================================ */
function generateDynamicQRIS(amount) {
  const staticQRIS = "00020101021126570011ID.DANA.WWW011893600915303309664402090330966440303UMI51440014ID.CO.QRIS.WWW0215ID10265451123550303UMI5204654053033605802ID5913Dabrong Store6015Kota Probolingg61056723763046FF0";
  
  // Remove CRC16 at the end
  let payload = staticQRIS.slice(0, -4);
  
  // Replace 010211 (static) with 010212 (dynamic)
  payload = payload.replace("010211", "010212");
  
  // Format amount tag: tag 54 + length of amount + amount
  const amountStr = amount.toString();
  const amountTag = "54" + amountStr.length.toString().padStart(2, "0") + amountStr;
  
  // Insert amount tag right after currency tag 5303360
  const currencyTag = "5303360";
  const pos = payload.indexOf(currencyTag);
  if (pos !== -1) {
    payload = payload.slice(0, pos + currencyTag.length) + amountTag + payload.slice(pos + currencyTag.length);
  }
  
  // Calculate CRC16 CCITT FALSE
  let crc = 0xFFFF;
  const polynomial = 0x1021;
  
  for (let i = 0; i < payload.length; i++) {
    const byte = payload.charCodeAt(i);
    crc ^= (byte << 8);
    for (let j = 0; j < 8; j++) {
      if ((crc & 0x8000) !== 0) {
        crc = ((crc << 1) ^ polynomial) & 0xFFFF;
      } else {
        crc = (crc << 1) & 0xFFFF;
      }
    }
  }
  
  const crcHex = crc.toString(16).toUpperCase().padStart(4, "0");
  return payload + crcHex;
}


/* ================================
   THEME TOGGLE (LIGHT/DARK)
   ================================ */
function initThemeToggle() {
  const btn = document.getElementById('theme-toggle');
  const icon = document.getElementById('theme-icon');
  if (!btn || !icon) return;

  const currentTheme = localStorage.getItem('theme');
  if (currentTheme === 'light-mode') {
    document.body.classList.add('light-mode');
    icon.setAttribute('data-lucide', 'sun');
  } else {
    icon.setAttribute('data-lucide', 'moon');
  }
  
  if (window.lucide) window.lucide.createIcons();

  btn.addEventListener('click', () => {
    document.body.classList.toggle('light-mode');
    let theme = 'dark-mode';
    if (document.body.classList.contains('light-mode')) {
      theme = 'light-mode';
      icon.setAttribute('data-lucide', 'sun');
    } else {
      icon.setAttribute('data-lucide', 'moon');
    }
    localStorage.setItem('theme', theme);
    if (window.lucide) window.lucide.createIcons();
  });
}

document.addEventListener('DOMContentLoaded', initThemeToggle);
