#!/bin/bash
# install.sh — Jalankan sekali untuk install Node.js + dependencies

echo "====================================="
echo " DiamondML — Setup Script"
echo "====================================="

# Install Node.js pkg
echo ""
echo "1. Menginstall Node.js v22 LTS..."
if [ -f "/tmp/node-install.pkg" ]; then
  sudo installer -pkg /tmp/node-install.pkg -target /
  echo "   ✅ Node.js terinstall!"
else
  echo "   ⚠️  File installer tidak ditemukan di /tmp/node-install.pkg"
  echo "   Download manual dari: https://nodejs.org/dist/v22.17.0/node-v22.17.0.pkg"
  exit 1
fi

# Tambah ke PATH
export PATH="/usr/local/bin:$PATH"
echo ""
echo "2. Verifikasi Node.js..."
node -v && npm -v

# Install dependencies
echo ""
echo "3. Install dependencies npm..."
cd /Users/daffaalbani/Documents/ml-topup-v2
npm install

echo ""
echo "====================================="
echo " ✅ Setup selesai!"
echo "====================================="
echo ""
echo " Jalankan server:"
echo "   cd /Users/daffaalbani/Documents/ml-topup-v2"
echo "   npm start"
echo ""
echo " Buka browser:"
echo "   Customer: http://localhost:3000"
echo "   Admin:    http://localhost:3000/admin"
echo ""
echo " PENTING: Edit file .env untuk:"
echo "   - BANK_ACCOUNT = nomor rekening kamu"
echo "   - BANK_HOLDER  = nama rekening"
echo "   - WHATSAPP_NUMBER = no WA admin (format: 628xxx)"
echo "   - ADMIN_PASSWORD = ganti dari default!"
echo "====================================="
