package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// SendFonnteMessage sends a WhatsApp message using the Fonnte API.
func SendFonnteMessage(target string, message string) {
	token := os.Getenv("FONNTE_TOKEN")
	if token == "" {
		log.Println("[FONNTE] Token tidak ditemukan. Skip pengiriman WA ke", target)
		return
	}

	url := "https://api.fonnte.com/send"
	
	payload := map[string]string{
		"target": target,
		"message": message,
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Println("[FONNTE] Error marshalling payload:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Println("[FONNTE] Error creating request:", err)
		return
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[FONNTE] Error sending message:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("[FONNTE] Berhasil mengirim pesan WA ke", target)
	} else {
		log.Println("[FONNTE] Gagal mengirim pesan WA, status:", resp.StatusCode)
	}
}
