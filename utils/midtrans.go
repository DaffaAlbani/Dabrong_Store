package utils

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type SnapRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	ItemDetails        []ItemDetail       `json:"item_details,omitempty"`
	CustomerDetails    CustomerDetails    `json:"customer_details,omitempty"`
}

type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int    `json:"gross_amount"`
}

type ItemDetail struct {
	ID       string `json:"id"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
}

type CustomerDetails struct {
	FirstName string `json:"first_name"`
	Phone     string `json:"phone"`
}

type SnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
	ErrorMessages []string `json:"error_messages,omitempty"`
}

func IsMidtransProduction() bool {
	return strings.ToLower(os.Getenv("MIDTRANS_IS_PRODUCTION")) == "true"
}

func GetMidtransServerKey() string {
	return os.Getenv("MIDTRANS_SERVER_KEY")
}

func GetMidtransClientKey() string {
	return os.Getenv("MIDTRANS_CLIENT_KEY")
}

// CreateSnapTransaction generates a Midtrans Snap token & URL for an order
func CreateSnapTransaction(orderNo string, grossAmount int, productName string, customerName string, customerPhone string) (*SnapResponse, error) {
	serverKey := GetMidtransServerKey()
	if serverKey == "" {
		return nil, fmt.Errorf("MIDTRANS_SERVER_KEY tidak dikonfigurasi di .env")
	}

	var apiURL string
	if IsMidtransProduction() {
		apiURL = "https://app.midtrans.com/snap/v1/transactions"
	} else {
		apiURL = "https://app.sandbox.midtrans.com/snap/v1/transactions"
	}

	// Truncate product name if too long for Midtrans
	cleanName := productName
	if len(cleanName) > 50 {
		cleanName = cleanName[:50]
	}

	reqBody := SnapRequest{
		TransactionDetails: TransactionDetails{
			OrderID:     orderNo,
			GrossAmount: grossAmount,
		},
		ItemDetails: []ItemDetail{
			{
				ID:       orderNo,
				Price:    grossAmount,
				Quantity: 1,
				Name:     cleanName,
			},
		},
		CustomerDetails: CustomerDetails{
			FirstName: customerName,
			Phone:     customerPhone,
		},
	}

	jsonPayload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	// Basic Auth with ServerKey (password empty)
	req.SetBasicAuth(serverKey, "")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("[MIDTRANS-ERROR] status=%d body=%s\n", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("gagal membuat transaksi Midtrans (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var snapResp SnapResponse
	if err := json.Unmarshal(bodyBytes, &snapResp); err != nil {
		return nil, err
	}

	return &snapResp, nil
}

// VerifyMidtransSignature validates SHA512 signature from Midtrans webhooks
func VerifyMidtransSignature(orderID string, statusCode string, grossAmount string, signatureKey string) bool {
	serverKey := GetMidtransServerKey()
	if serverKey == "" {
		return false
	}

	raw := orderID + statusCode + grossAmount + serverKey
	hasher := sha512.New()
	hasher.Write([]byte(raw))
	expectedSignature := hex.EncodeToString(hasher.Sum(nil))

	return strings.ToLower(expectedSignature) == strings.ToLower(signatureKey)
}
