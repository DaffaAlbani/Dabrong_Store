// apigames.js — Integrasi Apigames.id API
const axios = require('axios');
const md5   = require('md5');

const BASE_URL     = 'https://v1.apigames.id/v2';
const MERCHANT_ID  = process.env.APIGAMES_MERCHANT_ID;
const SECRET_KEY   = process.env.APIGAMES_SECRET_KEY;

/**
 * Generate signature Apigames
 * Formula: md5(merchant_id + ":" + secret_key + ":" + ref_id)
 */
function generateSignature(ref_id) {
  return md5(`${MERCHANT_ID}:${SECRET_KEY}:${ref_id}`);
}

/**
 * Generate unique ref_id berdasarkan timestamp + random
 */
function generateRefId() {
  return `REF${Date.now()}${Math.floor(Math.random() * 1000)}`;
}

/**
 * Cek username / nickname akun game ML via Apigames
 * Endpoint: GET https://v1.apigames.id/merchant/{merchant_id}/cek-username/mobilelegend
 * Signature: md5(merchant_id + secret_key)   ← BEDA dengan transaksi!
 */
async function checkUsername({ user_id, server_id }) {
  const signature = md5(`${MERCHANT_ID}${SECRET_KEY}`);

  // URL cek-username TIDAK pakai /v2/ — beda dari endpoint transaksi
  const url = `https://v1.apigames.id/merchant/${MERCHANT_ID}/cek-username/mobilelegend`;

  const params = {
    user_id:   `${user_id}${server_id}`, // Gabungkan user ID + server ID (contoh: 532082832011)
    signature,
  };

  console.log(`\n[APIGAMES CEK-USERNAME]`);
  console.log(`  URL       : ${url}`);
  console.log(`  user_id   : ${user_id}${server_id}`);
  console.log(`  signature : ${signature}`);

  try {
    const response = await axios.get(url, { params, timeout: 12000 });
    console.log(`  RESPONSE  :`, JSON.stringify(response.data));
    return { success: true, data: response.data };
  } catch (err) {
    const errData = err.response?.data;
    const status  = err.response?.status;
    console.error(`  ERROR ${status} :`, errData || err.message);
    return {
      success: false,
      message: errData?.message || errData?.msg || err.message,
      raw: errData,
      status,
    };
  }
}

/**
 * Ambil daftar produk / katalog dari Apigames
 * Bisa difilter per kategori (misal: AGML untuk Mobile Legends)
 */
async function getProducts(category = 'AGML') {
  const ref_id    = generateRefId();
  const signature = generateSignature(ref_id);

  try {
    const response = await axios.get(`${BASE_URL}/produk`, {
      params: {
        merchant_id: MERCHANT_ID,
        ref_id,
        signature,
        kategori: category,
      },
      timeout: 15000,
    });

    return { success: true, data: response.data };
  } catch (err) {
    const msg = err.response?.data?.message || err.message;
    return { success: false, message: msg };
  }
}

/**
 * Kirim transaksi top-up ke Apigames
 */
async function sendTransaction({ order_no, product_id, user_id, server_id }) {
  const ref_id    = order_no; // gunakan order_no sebagai ref_id agar unik & mudah tracking
  const signature = generateSignature(ref_id);

  try {
    const params = {
      merchant_id: MERCHANT_ID,
      ref_id,
      produk: product_id,
      tujuan: user_id,
      server_id: server_id || '',
      signature,
    };

    const response = await axios.get(`${BASE_URL}/transaksi`, {
      params,
      timeout: 30000,
    });

    return { success: true, data: response.data };
  } catch (err) {
    const errData = err.response?.data;
    return {
      success: false,
      message: errData?.message || err.message,
      data: errData,
    };
  }
}

/**
 * Cek status transaksi yang sudah dikirim
 */
async function checkTransactionStatus(ref_id) {
  const signature = generateSignature(ref_id);

  try {
    const url = 'https://v1.apigames.id/v2/transaksi/status';
    const response = await axios.get(url, {
      params: {
        merchant_id: MERCHANT_ID,
        ref_id,
        signature,
      },
      timeout: 15000,
    });

    return { success: true, data: response.data };
  } catch (err) {
    const msg = err.response?.data?.message || err.message;
    return { success: false, message: msg };
  }
}

/**
 * Cek saldo / deposit di Apigames
 */
async function checkBalance() {
  const signature = md5(`${MERCHANT_ID}:${SECRET_KEY}`);

  try {
    const url = `https://v1.apigames.id/merchant/${MERCHANT_ID}`;
    const response = await axios.get(url, {
      params: {
        signature,
      },
      timeout: 10000,
    });

    return { success: true, data: response.data };
  } catch (err) {
    const msg = err.response?.data?.message || err.message;
    return { success: false, message: msg };
  }
}

module.exports = {
  generateSignature,
  generateRefId,
  checkUsername,
  getProducts,
  sendTransaction,
  checkTransactionStatus,
  checkBalance,
};
