package handlers

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"ml-topup-v2/apigames"
	"ml-topup-v2/database"
	"ml-topup-v2/middleware"
)

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// POST /api/admin/login
func AdminLogin(c *fiber.Ctx) error {
	var req AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format request tidak valid"})
	}

	adminUser := os.Getenv("ADMIN_USERNAME")
	adminPass := os.Getenv("ADMIN_PASSWORD")

	if adminUser == "" || adminPass == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Konfigurasi kredensial admin di server belum lengkap (.env)",
		})
	}

	if req.Username == adminUser && req.Password == adminPass {
		// Generate Admin JWT Token
		token, err := middleware.GenerateToken(0, req.Username, "admin")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membuat sesi login"})
		}

		c.Cookie(&fiber.Cookie{
			Name:     "admin_token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			Secure:   os.Getenv("NODE_ENV") == "production",
			SameSite: "Lax",
			Path:     "/",
		})

		return c.JSON(fiber.Map{"success": true, "message": "Login admin berhasil"})
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Username atau password salah"})
}

// POST /api/admin/logout
func AdminLogout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "admin_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Path:     "/",
	})
	return c.JSON(fiber.Map{"success": true, "message": "Logout berhasil"})
}

// GET /api/admin/me
func AdminMe(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	return c.JSON(fiber.Map{"success": true, "username": username})
}

// GET /api/admin/stats
func AdminStats(c *fiber.Ctx) error {
	stats, err := database.GetOrderStats()
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal mengambil statistik order"})
	}

	// Cek saldo Apigames
	balanceRes, err := apigames.CheckBalance()
	var balanceData any
	if err == nil {
		if balMap, ok := balanceRes.(map[string]any); ok {
			if balData, exists := balMap["data"]; exists {
				balanceData = balData
			} else {
				balanceData = balMap
			}
		}
	} else {
		log.Printf("[ADMIN-STATS-WARN] Gagal mengambil saldo Apigames: %v\n", err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"stats":   stats,
		"balance": balanceData,
	})
}

// GET /api/admin/orders
func AdminOrders(c *fiber.Ctx) error {
	status := c.Query("status")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var orders []database.Order
	var err error

	if status != "" {
		orders, err = database.GetOrdersByStatus(strings.ToUpper(status))
	} else {
		orders, err = database.GetAllOrders(limit, offset)
	}

	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal mengambil order: " + err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "orders": orders})
}

// GET /api/admin/order/:order_no
func AdminOrderDetail(c *fiber.Ctx) error {
	orderNo := strings.ToUpper(c.Params("order_no"))
	order, err := database.GetOrderByOrderNo(orderNo)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Order tidak ditemukan"})
	}

	return c.JSON(fiber.Map{"success": true, "order": order})
}

type ConfirmRequest struct {
	OrderNo string `json:"order_no"`
}

// POST /api/admin/confirm
func AdminConfirmOrder(c *fiber.Ctx) error {
	var req ConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format request tidak valid"})
	}

	orderNo := strings.ToUpper(req.OrderNo)
	if orderNo == "" {
		return c.JSON(fiber.Map{"success": false, "message": "order_no wajib diisi"})
	}

	order, err := database.GetOrderByOrderNo(orderNo)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Order tidak ditemukan"})
	}

	if order.Status != "PENDING" {
		return c.JSON(fiber.Map{
			"success": false,
			"message": "Order sudah berstatus " + order.Status + ", tidak bisa dikonfirmasi ulang.",
		})
	}

	// Update status menjadi PROSES
	_ = database.UpdateOrderStatus(order.OrderNo, "PROSES", "PROSES", "Sedang diproses oleh admin")
	log.Printf("[ADMIN-CONFIRM] %s → PROSES\n", order.OrderNo)

	// Kirim transaksi ke Apigames
	res, err := apigames.SendTransaction(order.OrderNo, order.ProductID, order.PlayerID, order.ServerID)
	if err != nil {
		log.Printf("[APIGAMES-ERROR] Gagal mengirim transaksi ke Apigames: %v\n", err)
		_ = database.UpdateOrderStatus(order.OrderNo, "GAGAL", "FAILED_API", "Error API: "+err.Error())
		
		// Kembalikan saldo jika transaksi dibayar dengan saldo
		refundUserIfSaldoPaid(order)

		return c.JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengirim ke Apigames: " + err.Error(),
		})
	}

	log.Printf("[APIGAMES-RESPONSE] %s: %v\n", order.OrderNo, res)

	if apiMap, ok := res.(map[string]any); ok {
		statusVal, _ := apiMap["status"].(float64)
		rcVal, _ := apiMap["rc"].(float64)
		errorMsg, _ := apiMap["error_msg"].(string)

		if statusVal == 1 || rcVal == 200 {
			var apiStatus string
			if dataObj, ok := apiMap["data"].(map[string]any); ok {
				if st, exists := dataObj["status"].(string); exists {
					apiStatus = st
				} else if st, exists := dataObj["trx_status"].(string); exists {
					apiStatus = st
				}
			}
			if apiStatus == "" {
				apiStatus = "SUKSES"
			}

			messageText, _ := apiMap["message"].(string)
			if messageText == "" {
				messageText = "Transaksi diproses oleh provider"
			}

			_ = database.UpdateOrderStatus(order.OrderNo, "SUKSES", apiStatus, messageText)
			
			// Ambil data terupdate
			updatedOrder, _ := database.GetOrderByOrderNo(order.OrderNo)

			return c.JSON(fiber.Map{
				"success":  true,
				"message":  "Diamond berhasil dikirim ke player!",
				"order":    updatedOrder,
				"apigames": apiMap,
			})
		} else {
			// Gagal di Apigames
			_ = database.UpdateOrderStatus(order.OrderNo, "GAGAL", "FAILED_PROVIDER", errorMsg)
			
			// Refund saldo jika transaksi dibayar dengan saldo
			refundUserIfSaldoPaid(order)

			updatedOrder, _ := database.GetOrderByOrderNo(order.OrderNo)

			return c.JSON(fiber.Map{
				"success": false,
				"message": "Gagal dari Apigames: " + errorMsg,
				"order":   updatedOrder,
			})
		}
	}

	return c.JSON(fiber.Map{"success": false, "message": "Format respons Apigames tidak dikenal"})
}

func refundUserIfSaldoPaid(order database.Order) {
	if order.UserID.Valid && order.ApigamesStatus.String == "SUCCESS_SALDO" {
		userID := order.UserID.Int64
		// Dapatkan saldo user saat ini
		row := database.DB.QueryRow("SELECT saldo FROM users WHERE id = ?", userID)
		var currentSaldo int
		if err := row.Scan(&currentSaldo); err == nil {
			newSaldo := currentSaldo + order.TotalBayar
			_ = database.UpdateUserSaldo(userID, newSaldo)
			log.Printf("[REFUND-SALDO] Pengembalian saldo Rp%d sukses untuk User ID %d (Order %s)\n", order.TotalBayar, userID, order.OrderNo)
		}
	}
}

// POST /api/admin/check-trx
func AdminCheckTrx(c *fiber.Ctx) error {
	var req ConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format request tidak valid"})
	}

	orderNo := strings.ToUpper(req.OrderNo)
	if orderNo == "" {
		return c.JSON(fiber.Map{"success": false, "message": "order_no wajib diisi"})
	}

	order, err := database.GetOrderByOrderNo(orderNo)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Order tidak ditemukan"})
	}

	res, err := apigames.CheckTransactionStatus(order.OrderNo)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal cek status ke Apigames: " + err.Error()})
	}

	if apiMap, ok := res.(map[string]any); ok {
		var remoteStatus string
		var messageText string
		
		if dataObj, ok := apiMap["data"].(map[string]any); ok {
			if st, exists := dataObj["status"].(string); exists {
				remoteStatus = st
			} else if st, exists := dataObj["trx_status"].(string); exists {
				remoteStatus = st
			}
			messageText, _ = dataObj["message"].(string)
		}

		statusVal, _ := apiMap["status"].(float64)

		if statusVal == 1 && (strings.ToLower(remoteStatus) == "sukses" || strings.ToLower(remoteStatus) == "success") {
			_ = database.UpdateOrderStatus(order.OrderNo, "SUKSES", remoteStatus, messageText)
		} else if statusVal == 1 && (strings.ToLower(remoteStatus) == "gagal" || strings.ToLower(remoteStatus) == "failed") {
			_ = database.UpdateOrderStatus(order.OrderNo, "GAGAL", remoteStatus, messageText)
			refundUserIfSaldoPaid(order)
		}

		return c.JSON(fiber.Map{"success": true, "data": apiMap})
	}

	return c.JSON(fiber.Map{"success": false, "message": "Format respons tidak valid"})
}

type RejectRequest struct {
	OrderNo string `json:"order_no"`
	Reason  string `json:"reason"`
}

// POST /api/admin/reject
func AdminRejectOrder(c *fiber.Ctx) error {
	var req RejectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format request tidak valid"})
	}

	orderNo := strings.ToUpper(req.OrderNo)
	if orderNo == "" {
		return c.JSON(fiber.Map{"success": false, "message": "order_no wajib diisi"})
	}

	order, err := database.GetOrderByOrderNo(orderNo)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Order tidak ditemukan"})
	}

	if order.Status == "SUKSES" || order.Status == "FAILED" {
		return c.JSON(fiber.Map{"success": false, "message": "Order sudah final, tidak bisa dibatalkan"})
	}

	reason := req.Reason
	if reason == "" {
		reason = "Dibatalkan oleh admin"
	}

	_ = database.UpdateOrderStatus(order.OrderNo, "GAGAL", "REJECTED_ADMIN", reason)
	refundUserIfSaldoPaid(order)

	updatedOrder, _ := database.GetOrderByOrderNo(order.OrderNo)
	return c.JSON(fiber.Map{"success": true, "message": "Order berhasil dibatalkan", "order": updatedOrder})
}

// ============ PRODUCTS CRUD ENDPOINTS ============

// POST /api/admin/products
func AdminAddProduct(c *fiber.Ctx) error {
	var p database.Product
	if err := c.BodyParser(&p); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format produk tidak valid"})
	}

	if p.ProductID == "" || p.ProductName == "" || p.Price <= 0 {
		return c.JSON(fiber.Map{"success": false, "message": "Data produk tidak lengkap"})
	}

	p.ProductID = strings.ToUpper(strings.TrimSpace(p.ProductID))
	p.ProductName = strings.TrimSpace(p.ProductName)
	if p.Category == "" {
		p.Category = "AGML"
	}
	if p.Status == "" {
		p.Status = "active"
	}

	created, err := database.AddProduct(p)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal menambah produk: " + err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Produk berhasil ditambahkan", "product": created})
}

// PUT /api/admin/products/:product_id
func AdminUpdateProduct(c *fiber.Ctx) error {
	productID := strings.ToUpper(c.Params("product_id"))
	
	var p database.Product
	if err := c.BodyParser(&p); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format produk tidak valid"})
	}

	p.ProductName = strings.TrimSpace(p.ProductName)
	if p.Category == "" {
		p.Category = "AGML"
	}
	if p.Status == "" {
		p.Status = "active"
	}

	updated, err := database.UpdateProduct(productID, p)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal update produk: " + err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Produk berhasil diupdate", "product": updated})
}

// DELETE /api/admin/products/:product_id
func AdminDeleteProduct(c *fiber.Ctx) error {
	productID := strings.ToUpper(c.Params("product_id"))

	err := database.DeleteProduct(productID)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal menghapus produk: " + err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Produk berhasil dihapus"})
}
