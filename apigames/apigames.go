package apigames

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const BASE_URL = "https://v1.apigames.id/v2"

func GetMerchantID() string {
	return os.Getenv("APIGAMES_MERCHANT_ID")
}

func GetSecretKey() string {
	return os.Getenv("APIGAMES_SECRET_KEY")
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GenerateSignature(refID string) string {
	return getMD5Hash(fmt.Sprintf("%s:%s:%s", GetMerchantID(), GetSecretKey(), refID))
}

func GenerateRefID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("REF%d%d", time.Now().UnixNano()/int64(time.Millisecond), r.Intn(1000))
}

type ApiResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
	Data    any            `json:"data,omitempty"`
	Status  int            `json:"status,omitempty"`
}

// Cek username / nickname akun game via Apigames (ML, FF, dll.)
func CheckUsername(game, userID, serverID string) (any, error) {
	merchantID := GetMerchantID()
	secretKey := GetSecretKey()
	signature := getMD5Hash(fmt.Sprintf("%s%s", merchantID, secretKey))

	var targetGame string
	var queryUserID string

	g := strings.ToLower(strings.TrimSpace(game))
	if strings.Contains(g, "legend") || g == "ml" || g == "mlbb" {
		targetGame = "mobilelegend"
		queryUserID = fmt.Sprintf("%s%s", userID, serverID)
	} else if strings.Contains(g, "freefire") || g == "ff" {
		targetGame = "freefire"
		queryUserID = userID
	} else if strings.Contains(g, "higgs") {
		targetGame = "higgs"
		queryUserID = userID
	} else if strings.Contains(g, "cod") || g == "codm" {
		targetGame = "codm"
		queryUserID = userID
	} else if strings.Contains(g, "starrail") || g == "hsr" || strings.Contains(g, "honkai") {
		targetGame = "starrail"
		queryUserID = fmt.Sprintf("%s%s", userID, serverID)
	} else {
		// Game lain tidak didukung cek username resmi di Apigames, kembalikan bypass virtual
		return map[string]any{
			"status": 1,
			"rc":     200,
			"data": map[string]any{
				"username": fmt.Sprintf("ID %s", userID),
			},
		}, nil
	}

	apiURL := fmt.Sprintf("https://v1.apigames.id/merchant/%s/cek-username/%s", merchantID, targetGame)
	
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("user_id", queryUserID)
	q.Set("signature", signature)
	u.RawQuery = q.Encode()

	log.Printf("[APIGAMES CEK-USERNAME] Calling URL: %s\n", u.String())

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Ambil daftar produk / katalog dari Apigames (Bypass / Optional)
func GetProducts(category string) (any, error) {
	refID := GenerateRefID()
	signature := GenerateSignature(refID)

	apiURL := fmt.Sprintf("%s/produk", BASE_URL)
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("merchant_id", GetMerchantID())
	q.Set("ref_id", refID)
	q.Set("signature", signature)
	q.Set("kategori", category)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Kirim transaksi top-up ke Apigames
func SendTransaction(orderNo, productID, userID, serverID string) (any, error) {
	signature := GenerateSignature(orderNo) // Gunakan orderNo sebagai refID

	apiURL := fmt.Sprintf("%s/transaksi", BASE_URL)
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("merchant_id", GetMerchantID())
	q.Set("ref_id", orderNo)
	q.Set("produk", productID)
	q.Set("tujuan", userID)
	if serverID != "" {
		q.Set("server_id", serverID)
	}
	q.Set("signature", signature)
	u.RawQuery = q.Encode()

	log.Printf("[APIGAMES TRANSAKSI] Calling URL: %s\n", u.String())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Cek status transaksi yang sudah dikirim
func CheckTransactionStatus(refID string) (any, error) {
	signature := GenerateSignature(refID)

	apiURL := "https://v1.apigames.id/v2/transaksi/status"
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("merchant_id", GetMerchantID())
	q.Set("ref_id", refID)
	q.Set("signature", signature)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Cek saldo / deposit di Apigames
func CheckBalance() (any, error) {
	merchantID := GetMerchantID()
	secretKey := GetSecretKey()
	signature := getMD5Hash(fmt.Sprintf("%s:%s", merchantID, secretKey))

	apiURL := fmt.Sprintf("https://v1.apigames.id/merchant/%s", merchantID)
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("signature", signature)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
