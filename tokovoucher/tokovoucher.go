package tokovoucher

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const BASE_URL = "https://api.tokovoucher.net/v1"

func GetMemberCode() string {
	return os.Getenv("TOKOVOUCHER_MEMBER_ID")
}

func GetSecretKey() string {
	return os.Getenv("TOKOVOUCHER_SECRET")
}



type TokovoucherResponse struct {
	Status  interface{} `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Cek username / nickname akun game
func CheckUsername(game, userID, serverID string) (any, error) {
	g := strings.ToLower(game)
	isML := strings.Contains(g, "legend") || g == "ml" || g == "mlbb"
	isFF := g == "ff" || g == "free fire" || g == "freefire"

	if isML || isFF {
		var url string
		if isML {
			url = fmt.Sprintf("https://api.isan.eu.org/nickname/ml?id=%s&zone=%s", userID, serverID)
		} else {
			url = fmt.Sprintf("https://api.isan.eu.org/nickname/ff?id=%s", userID)
		}
		
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var apiResp map[string]any
			if json.Unmarshal(body, &apiResp) == nil {
				// Cek jika success = true
				if success, ok := apiResp["success"].(bool); ok && success {
					name, _ := apiResp["name"].(string)
					// Kalau FF mungkin tidak return name secara explicit, fallback ke "Valid"
					if name == "" {
						name = "ID Valid"
					}
					return map[string]any{
						"status": "success",
						"rc":     200,
						"data": map[string]any{
							"username": name,
						},
					}, nil
				} else {
					// Berarti ID tidak ditemukan dari API ini
					return map[string]any{
						"status": "success",
						"rc":     200,
						"data": map[string]any{
							"username": "", // Kosong agar trigger notif "tidak ditemukan"
						},
					}, nil
				}
			}
			return nil, fmt.Errorf("invalid response from API")
		}
		return nil, err
	}

	// Untuk game lain atau jika pengecekan gagal, lempar error agar masuk mode "bypass" di frontend.
	return nil, fmt.Errorf("unsupported game or verification failed")
}

// Transaksi (Topup) menggunakan Tokovoucher
func SendTransaction(orderNo, productID, userID, serverID string) (any, error) {
	memberCode := GetMemberCode()
	secret := GetSecretKey()

	apiURL := fmt.Sprintf("%s/transaksi", BASE_URL)
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("ref_id", orderNo)
	q.Set("produk", productID)
	q.Set("tujuan", userID)
	q.Set("secret", secret)
	q.Set("member_code", memberCode)
	
	if serverID != "" && serverID != "-" {
		q.Set("server_id", serverID)
	}

	u.RawQuery = q.Encode()

	log.Printf("[TOKOVOUCHER TRANSAKSI] Calling URL: %s\n", u.String())

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("[TOKOVOUCHER TRANSAKSI-RESPONSE] Raw: %s\n", string(body))

	var genericMap map[string]any
	if err2 := json.Unmarshal(body, &genericMap); err2 == nil {
		return genericMap, nil
	}
	return nil, fmt.Errorf("failed to parse response: %v, body: %s", err, string(body))
}

func CheckTransactionStatus(refID string) (any, error) {
	memberCode := GetMemberCode()
	secret := GetSecretKey()
	
	plain := fmt.Sprintf("%s:%s:%s", memberCode, secret, refID)
	hash := md5.Sum([]byte(plain))
	signature := fmt.Sprintf("%x", hash)
	
	apiURL := fmt.Sprintf("%s/transaksi/status", BASE_URL)
	
	payload := map[string]string{
		"ref_id":      refID,
		"member_code": memberCode,
		"signature":   signature,
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	log.Printf("[TOKOVOUCHER STATUS-CHECK] Calling URL: %s, Payload: %s\n", apiURL, string(jsonPayload))
	
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	log.Printf("[TOKOVOUCHER STATUS-RESPONSE] Raw: %s\n", string(body))
	
	var genericMap map[string]any
	if err2 := json.Unmarshal(body, &genericMap); err2 == nil {
		return genericMap, nil
	}
	return nil, fmt.Errorf("failed to parse status check response: %v, body: %s", err, string(body))
}

func CheckBalance() (any, error) {
	memberCode := GetMemberCode()
	secret := GetSecretKey()
	
	plain := fmt.Sprintf("%s:%s", memberCode, secret)
	hash := md5.Sum([]byte(plain))
	signature := fmt.Sprintf("%x", hash)
	
	apiURL := fmt.Sprintf("%s/member?member_code=%s&signature=%s", BASE_URL, memberCode, signature)
	
	log.Printf("[TOKOVOUCHER SALDO] Calling URL: %s\n", apiURL)
	
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	log.Printf("[TOKOVOUCHER SALDO-RESPONSE] Raw: %s\n", string(body))
	
	var genericMap map[string]any
	if err2 := json.Unmarshal(body, &genericMap); err2 == nil {
		return genericMap, nil
	}
	return nil, fmt.Errorf("failed to parse balance response: %v, body: %s", err, string(body))
}

// Mengambil semua produk dari API Tokovoucher
func GetAllProducts() ([]map[string]any, error) {
	memberCode := GetMemberCode()
	secret := GetSecretKey()
	
	plain := fmt.Sprintf("%s:%s", memberCode, secret)
	hash := md5.Sum([]byte(plain))
	signature := fmt.Sprintf("%x", hash)
	
	apiURL := fmt.Sprintf("%s/member/produk/list?member_code=%s&signature=%s", "https://api.tokovoucher.net", memberCode, signature)
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var apiResp struct {
		Status  int              `json:"status"`
		Message string           `json:"message"`
		Data    []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse products response: %v", err)
	}
	
	return apiResp.Data, nil
}
