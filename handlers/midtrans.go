package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"ml-topup-v2/database"
	"ml-topup-v2/tokovoucher"
	"ml-topup-v2/utils"
)

type MidtransCallbackRequest struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status,omitempty"`
}

// POST /api/midtrans/callback
func MidtransCallback(c *fiber.Ctx) error {
	var req MidtransCallbackRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[MIDTRANS-CALLBACK-ERROR] Invalid body parser: %v\n", err)
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Invalid body"})
	}

	log.Printf("[MIDTRANS-CALLBACK] order_id=%s status=%s status_code=%s payment_type=%s\n",
		req.OrderID, req.TransactionStatus, req.StatusCode, req.PaymentType)

	// Verify signature
	if !utils.VerifyMidtransSignature(req.OrderID, req.StatusCode, req.GrossAmount, req.SignatureKey) {
		log.Printf("[MIDTRANS-CALLBACK-WARN] Signature verification failed for order %s\n", req.OrderID)
		return c.Status(403).JSON(fiber.Map{"status": "error", "message": "Invalid signature"})
	}

	order, err := database.GetOrderByOrderNo(req.OrderID)
	if err != nil {
		log.Printf("[MIDTRANS-CALLBACK-WARN] Order %s not found in database: %v\n", req.OrderID, err)
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Order not found"})
	}

	txStatus := strings.ToLower(req.TransactionStatus)
	fraudStatus := strings.ToLower(req.FraudStatus)

	isSuccess := false
	if txStatus == "capture" {
		if fraudStatus == "accept" || fraudStatus == "" {
			isSuccess = true
		}
	} else if txStatus == "settlement" {
		isSuccess = true
	}

	if isSuccess {
		if order.Status == "SUKSES" {
			log.Printf("[MIDTRANS-CALLBACK] Order %s already SUKSES. Ignoring.\n", order.OrderNo)
			return c.JSON(fiber.Map{"status": "OK", "message": "Already completed"})
		}

		log.Printf("[MIDTRANS-CALLBACK] Pembayaran LUNAS untuk %s! Memproses pengiriman otomatis...\n", order.OrderNo)
		
		// Update status to PROSES
		_ = database.UpdateOrderStatus(order.OrderNo, "PROSES", req.PaymentType, "Pembayaran berhasil diverifikasi otomatis oleh Midtrans")

		// Send transaction to Tokovoucher / Supplier
		tvRes, err := tokovoucher.SendTransaction(order.OrderNo, order.ProductID, order.PlayerID, order.ServerID)
		if err != nil {
			log.Printf("[MIDTRANS-FULFILLMENT-ERROR] Order %s gagal kirim ke supplier: %v\n", order.OrderNo, err)
			_ = database.UpdateOrderStatus(order.OrderNo, "PROSES", req.PaymentType, fmt.Sprintf("Pembayaran LUNAS (Midtrans). Gagal auto-injeksi: %v", err))
		} else {
			log.Printf("[MIDTRANS-FULFILLMENT-SUCCESS] Order %s berhasil dikirim ke supplier: %+v\n", order.OrderNo, tvRes)
			_ = database.UpdateOrderStatus(order.OrderNo, "SUKSES", req.PaymentType, "Item berhasil terkirim otomatis")
		}

		// Send WhatsApp Notification to buyer via Fonnte if phone number present
		if order.Whatsapp != "" {
			waMsg := fmt.Sprintf(
				"✅ *PEMBAYARAN & TOP-UP BERHASIL!*\n\n"+
					"📋 No. Order: %s\n"+
					"🎮 Game: %s\n"+
					"👤 Player ID: %s (Server: %s)\n"+
					"💎 Paket: %s\n"+
					"💰 Total: Rp %s\n\n"+
					"Item Anda telah berhasil dikirim otomatis oleh Dabrong Store. Terima kasih telah berbelanja!",
				order.OrderNo, order.ProductName, order.PlayerID, order.ServerID, order.ProductName,
				formatRupiah(order.TotalBayar),
			)
			go utils.SendFonnteMessage(order.Whatsapp, waMsg)
		}

		return c.JSON(fiber.Map{"status": "OK", "message": "Payment verified and item fulfilled"})
	} else if txStatus == "cancel" || txStatus == "deny" || txStatus == "expire" {
		log.Printf("[MIDTRANS-CALLBACK] Order %s dibatalkan/kadaluarsa (status: %s)\n", order.OrderNo, txStatus)
		_ = database.UpdateOrderStatus(order.OrderNo, "GAGAL", req.PaymentType, fmt.Sprintf("Pembayaran %s di Midtrans", txStatus))
		return c.JSON(fiber.Map{"status": "OK", "message": "Order updated to GAGAL"})
	}

	return c.JSON(fiber.Map{"status": "OK", "message": "Notification received"})
}

// POST /api/midtrans/create-snap
func CreateSnap(c *fiber.Ctx) error {
	type SnapReq struct {
		OrderNo string `json:"order_no"`
	}
	var req SnapReq
	if err := c.BodyParser(&req); err != nil || req.OrderNo == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "order_no wajib diisi"})
	}

	order, err := database.GetOrderByOrderNo(req.OrderNo)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Order tidak ditemukan"})
	}

	pName := order.PlayerName
	if pName == "" {
		pName = order.PlayerID
	}

	snapResp, err := utils.CreateSnapTransaction(
		order.OrderNo,
		order.TotalBayar,
		order.ProductName,
		pName,
		order.Whatsapp,
	)

	if err != nil {
		log.Printf("[CREATE-SNAP-ERROR] %v\n", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"snap_token":   snapResp.Token,
		"redirect_url": snapResp.RedirectURL,
		"client_key":   utils.GetMidtransClientKey(),
	})
}
