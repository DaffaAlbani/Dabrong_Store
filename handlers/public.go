package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"ml-topup-v2/database"
	"ml-topup-v2/tokovoucher"
)

var defaultProducts = []database.Product{
	// Mobile Legends (AGML)
	{ProductID: "AGML022", ProductName: "22 Diamond", Category: "AGML", Price: 5900, Status: "active"},
	{ProductID: "AGML056", ProductName: "56 Diamond", Category: "AGML", Price: 13900, Status: "active"},
	{ProductID: "AGML086", ProductName: "86 Diamond", Category: "AGML", Price: 21900, Status: "active"},
	{ProductID: "AGML172", ProductName: "172 Diamond", Category: "AGML", Price: 43900, Status: "active"},
	{ProductID: "AGML257", ProductName: "257 Diamond", Category: "AGML", Price: 64900, Status: "active"},
	{ProductID: "AGML344", ProductName: "344 Diamond", Category: "AGML", Price: 84900, Status: "active"},
	{ProductID: "AGML514", ProductName: "514 Diamond", Category: "AGML", Price: 124900, Status: "active"},
	{ProductID: "AGML600", ProductName: "600 Diamond", Category: "AGML", Price: 144900, Status: "active"},
	{ProductID: "AGML878", ProductName: "878 Diamond", Category: "AGML", Price: 209900, Status: "active"},
	{ProductID: "AGML1195", ProductName: "1195 Diamond", Category: "AGML", Price: 279900, Status: "active"},
	{ProductID: "AGML2010", ProductName: "2010 Diamond", Category: "AGML", Price: 469900, Status: "active"},
	{ProductID: "AGML3688", ProductName: "3688 Diamond", Category: "AGML", Price: 849900, Status: "active"},
	{ProductID: "AGML5532", ProductName: "5532 Diamond", Category: "AGML", Price: 1249900, Status: "active"},

	// Free Fire (AGFF)
	{ProductID: "AGFF05", ProductName: "5 Diamond", Category: "AGFF", Price: 1000, Status: "active"},
	{ProductID: "AGFF12", ProductName: "12 Diamond", Category: "AGFF", Price: 2000, Status: "active"},
	{ProductID: "AGFF50", ProductName: "50 Diamond", Category: "AGFF", Price: 8000, Status: "active"},
	{ProductID: "AGFF70", ProductName: "70 Diamond", Category: "AGFF", Price: 10000, Status: "active"},
	{ProductID: "AGFF140", ProductName: "140 Diamond", Category: "AGFF", Price: 20000, Status: "active"},
	{ProductID: "AGFF355", ProductName: "355 Diamond", Category: "AGFF", Price: 50000, Status: "active"},
	{ProductID: "AGFF720", ProductName: "720 Diamond", Category: "AGFF", Price: 100000, Status: "active"},

	// PUBG Mobile (AGPB)
	{ProductID: "AGPB30", ProductName: "30 UC", Category: "AGPB", Price: 7000, Status: "active"},
	{ProductID: "AGPB60", ProductName: "60 UC", Category: "AGPB", Price: 14000, Status: "active"},
	{ProductID: "AGPB325", ProductName: "325 UC", Category: "AGPB", Price: 65000, Status: "active"},
	{ProductID: "AGPB660", ProductName: "660 UC", Category: "AGPB", Price: 130000, Status: "active"},

	// Genshin Impact (AGGI)
	{ProductID: "AGGI60", ProductName: "60 Genesis Crystal", Category: "AGGI", Price: 15000, Status: "active"},
	{ProductID: "AGGI300", ProductName: "300 Genesis Crystal", Category: "AGGI", Price: 75000, Status: "active"},
	{ProductID: "AGGI980", ProductName: "980 Genesis Crystal", Category: "AGGI", Price: 230000, Status: "active"},

	// Valorant (AGVL)
	{ProductID: "AGVL125", ProductName: "125 Points", Category: "AGVL", Price: 15000, Status: "active"},
	{ProductID: "AGVL375", ProductName: "375 Points", Category: "AGVL", Price: 45000, Status: "active"},
	{ProductID: "AGVL1120", ProductName: "1120 Points", Category: "AGVL", Price: 130000, Status: "active"},

	// Honkai: Star Rail (AGHSR)
	{ProductID: "AGHSR60", ProductName: "60 Oneiric Shard", Category: "AGHSR", Price: 16000, Status: "active"},
	{ProductID: "AGHSR300", ProductName: "300 Oneiric Shard", Category: "AGHSR", Price: 79000, Status: "active"},

	// Call of Duty Mobile (AGCODM)
	{ProductID: "AGCODM31", ProductName: "31 CP", Category: "AGCODM", Price: 5000, Status: "active"},
	{ProductID: "AGCODM63", ProductName: "63 CP", Category: "AGCODM", Price: 10000, Status: "active"},
	{ProductID: "AGCODM128", ProductName: "128 CP", Category: "AGCODM", Price: 20000, Status: "active"},

	// Higgs Domino (AGHD)
	{ProductID: "AGHD30M", ProductName: "30M Emas", Category: "AGHD", Price: 5000, Status: "active"},
	{ProductID: "AGHD60M", ProductName: "60M Emas", Category: "AGHD", Price: 10000, Status: "active"},
	{ProductID: "AGHD200M", ProductName: "200M Emas", Category: "AGHD", Price: 30000, Status: "active"},

	// Roblox (AGRBLX)
	{ProductID: "AGRBLX80", ProductName: "80 Robux", Category: "AGRBLX", Price: 15000, Status: "active"},
	{ProductID: "AGRBLX400", ProductName: "400 Robux", Category: "AGRBLX", Price: 75000, Status: "active"},

	// Steam Wallet (AGSTM)
	{ProductID: "AGSTM12", ProductName: "Steam Wallet Rp 12.000", Category: "AGSTM", Price: 14000, Status: "active"},
	{ProductID: "AGSTM45", ProductName: "Steam Wallet Rp 45.000", Category: "AGSTM", Price: 49000, Status: "active"},

	// Google Play Voucher (AGGPL)
	{ProductID: "AGGPL20", ProductName: "Google Play Rp 20.000", Category: "AGGPL", Price: 22000, Status: "active"},
	{ProductID: "AGGPL50", ProductName: "Google Play Rp 50.000", Category: "AGGPL", Price: 53000, Status: "active"},

	// Garena Shells (AGGS)
	{ProductID: "AGGS33", ProductName: "33 Garena Shell", Category: "AGGS", Price: 10000, Status: "active"},
	{ProductID: "AGGS66", ProductName: "66 Garena Shell", Category: "AGGS", Price: 20000, Status: "active"},
	{ProductID: "AGGS165", ProductName: "165 Garena Shell", Category: "AGGS", Price: 50000, Status: "active"},
}

// GET /api/check-user
func CheckUser(c *fiber.Ctx) error {
	game := strings.TrimSpace(c.Query("game"))
	userID := strings.TrimSpace(c.Query("user_id"))
	serverID := strings.TrimSpace(c.Query("server_id"))

	if game == "" {
		game = "mobilelegends" // Default fallback
	}

	if userID == "" {
		return c.JSON(fiber.Map{"success": false, "message": "User ID wajib diisi"})
	}

	g := strings.ToLower(game)
	isML := strings.Contains(g, "legend") || g == "ml" || g == "mlbb"
	
	if isML && serverID == "" {
		return c.JSON(fiber.Map{"success": false, "message": "Server ID wajib diisi untuk Mobile Legends"})
	}

	// Format validation
	if g == "valorant" || g == "vl" {
		if !strings.Contains(userID, "#") {
			return c.JSON(fiber.Map{"success": false, "message": "Format Riot ID salah. Harus menggunakan format Username#TAG (contoh: User#123)"})
		}
	} else {
		numericCheck := regexp.MustCompile(`^\d+$`)
		if !numericCheck.MatchString(userID) {
			return c.JSON(fiber.Map{"success": false, "message": "ID Player hanya boleh berupa angka"})
		}
	}

	log.Printf("[CHECK-USER] game=%s user_id=%s server_id=%s\n", game, userID, serverID)

	res, err := tokovoucher.CheckUsername(game, userID, serverID)
	if err != nil {
		log.Printf("[CHECK-USER] Error calling Apigames: %v\n", err)
		return c.JSON(fiber.Map{
			"success":   true,
			"nickname":  fmt.Sprintf("ID %s", userID),
			"user_id":   userID,
			"server_id": serverID,
			"warning":   "Verifikasi nickname tidak tersedia saat ini. Pastikan ID benar.",
			"bypass":    true,
		})
	}

	// Parsing nickname dari Map response Apigames
	var nickname string
	if apiMap, ok := res.(map[string]any); ok {
		var statusOk bool
		if stFloat, ok := apiMap["status"].(float64); ok && stFloat == 1 {
			statusOk = true
		} else if stStr, ok := apiMap["status"].(string); ok && (stStr == "1" || strings.ToLower(stStr) == "success") {
			statusOk = true
		}
		
		var rcOk bool
		if rcFloat, ok := apiMap["rc"].(float64); ok && rcFloat == 200 {
			rcOk = true
		} else if rcInt, ok := apiMap["rc"].(int); ok && rcInt == 200 {
			rcOk = true
		} else if rcStr, ok := apiMap["rc"].(string); ok && rcStr == "200" {
			rcOk = true
		}

		if statusOk || rcOk {
			nickname = extractNickname(apiMap)
			log.Printf("[CHECK-USER] statusOk=%v rcOk=%v nickname='%s' apiMap=%+v\n", statusOk, rcOk, nickname, apiMap)
		} else {
			errorMsg, _ := apiMap["error_msg"].(string)
			if errorMsg == "" {
				errorMsg, _ = apiMap["message"].(string)
			}
			var rcVal float64
			if r, ok := apiMap["rc"].(float64); ok {
				rcVal = r
			} else if r, ok := apiMap["rc"].(int); ok {
				rcVal = float64(r)
			}
			errMsgLower := strings.ToLower(errorMsg)
			isConfigError := strings.Contains(errMsgLower, "signature") || strings.Contains(errMsgLower, "unauthorized") || rcVal == 401
			
			if isConfigError {
				log.Println("[CHECK-USER] ⚠️ Signature error — bypass mode aktif")
				return c.JSON(fiber.Map{
					"success":   true,
					"nickname":  fmt.Sprintf("ID %s", userID),
					"user_id":   userID,
					"server_id": serverID,
					"warning":   "Verifikasi nickname tidak tersedia saat ini. Pastikan ID benar.",
					"bypass":    true,
				})
			}
		}
	}

	if nickname != "" {
		return c.JSON(fiber.Map{
			"success":   true,
			"nickname":  nickname,
			"user_id":   userID,
			"server_id": serverID,
		})
	}

	return c.JSON(fiber.Map{
		"success": false,
		"message": "ID Player tidak ditemukan. Pastikan User ID dan Server ID benar.",
	})
}

func extractNickname(m map[string]any) string {
	keys := []string{"username", "nickname", "name"}
	for _, k := range keys {
		if val, ok := m[k].(string); ok && val != "" {
			return val
		}
	}
	// Coba cek properti data
	if dataObj, ok := m["data"].(map[string]any); ok {
		for _, k := range keys {
			if val, ok := dataObj[k].(string); ok && val != "" {
				return val
			}
		}
	}
	// Coba cek properti result
	if resObj, ok := m["result"].(map[string]any); ok {
		for _, k := range keys {
			if val, ok := resObj[k].(string); ok && val != "" {
				return val
			}
		}
	}
	return ""
}

// GET /api/products
func GetProducts(c *fiber.Ctx) error {
	category := strings.TrimSpace(c.Query("category"))
	var cached []database.Product
	var err error

	if category != "" {
		cached, err = database.GetProductsByCategory(category)
	} else {
		cached, err = database.GetCachedProducts()
	}

	if err != nil || len(cached) == 0 {
		log.Printf("[PRODUCTS] Gagal membaca orders.db atau kosong. Error: %v, len: %d. Inisialisasi data default...\n", err, len(cached))
		_ = database.CacheProducts(defaultProducts)
		if category != "" {
			cached, _ = database.GetProductsByCategory(category)
		} else {
			cached, _ = database.GetCachedProducts()
		}
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"products": cached,
	})
}

type OrderRequest struct {
	PlayerID      string `json:"player_id"`
	ServerID      string `json:"server_id"`
	PlayerName    string `json:"player_name"`
	ProductID     string `json:"product_id"`
	ProductName   string `json:"product_name"`
	Diamond       string `json:"diamond"`
	Price         string `json:"price"`
	Whatsapp      string `json:"whatsapp"`
	PaymentMethod string `json:"payment_method"`
}

// POST /api/order
func CreateOrder(c *fiber.Ctx) error {
	var req OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format request tidak valid"})
	}

	playerID := strings.TrimSpace(req.PlayerID)
	serverID := strings.TrimSpace(req.ServerID)
	productID := strings.TrimSpace(req.ProductID)
	whatsapp := strings.TrimSpace(req.Whatsapp)

	if playerID == "" || productID == "" || req.Price == "" || req.Diamond == "" {
		return c.JSON(fiber.Map{"success": false, "message": "Data tidak lengkap"})
	}

	waRegexp := regexp.MustCompile(`^(08|628)\d{8,12}$`)
	if whatsapp == "" || !waRegexp.MatchString(whatsapp) {
		return c.JSON(fiber.Map{"success": false, "message": "Nomor WhatsApp tidak valid (contoh: 08123456789)"})
	}

	priceVal, err := strconv.Atoi(req.Price)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format harga tidak valid"})
	}
	diamondVal, err := strconv.Atoi(req.Diamond)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format jumlah diamond tidak valid"})
	}

	// Determine game name and order prefix based on productID
	gameName := "Game"
	prefix := "DBR"
	prodIDUpper := strings.ToUpper(productID)
	switch {
	case strings.HasPrefix(prodIDUpper, "AGML"):
		gameName = "Mobile Legends"
		prefix = "DML"
	case strings.HasPrefix(prodIDUpper, "AGFF"):
		gameName = "Free Fire"
		prefix = "DFF"
	case strings.HasPrefix(prodIDUpper, "AGHOK"):
		gameName = "Honor of Kings"
		prefix = "DHK"
	case strings.HasPrefix(prodIDUpper, "AGPUBG"):
		gameName = "PUBG Mobile"
		prefix = "DPB"
	case strings.HasPrefix(prodIDUpper, "AGVALO"):
		gameName = "Valorant"
		prefix = "DVL"
	case strings.HasPrefix(prodIDUpper, "AGGI"):
		gameName = "Genshin Impact"
		prefix = "DGI"
	case strings.HasPrefix(prodIDUpper, "AGHSR"):
		gameName = "Honkai Star Rail"
		prefix = "DSR"
	case strings.HasPrefix(prodIDUpper, "AGCODM"):
		gameName = "CODM"
		prefix = "DCM"
	}

	// Generate order number with random suffix to prevent UNIQUE constraint collisions
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderNo := fmt.Sprintf("%s%d%03d", prefix, time.Now().Unix()%100000000, r.Intn(1000))
	uniqueCode := r.Intn(900) + 100 // 100 - 999
	totalBayar := priceVal + uniqueCode

	bankName := os.Getenv("BANK_NAME")
	if bankName == "" {
		bankName = "BCA"
	}
	bankAccount := os.Getenv("BANK_ACCOUNT")
	if bankAccount == "" {
		bankAccount = "0000000000"
	}
	bankHolder := os.Getenv("BANK_HOLDER")
	if bankHolder == "" {
		bankHolder = "Admin"
	}

	// Hubungkan dengan user terdaftar jika dia login
	var userID sql.NullInt64
	if c.Locals("user_id") != nil {
		userID.Int64 = c.Locals("user_id").(int64)
		userID.Valid = true
	}

	pName := req.PlayerName
	if pName == "" {
		pName = playerID
	}

	prodName := req.ProductName
	if prodName == "" {
		prodName = fmt.Sprintf("%d Diamond", diamondVal)
	}

	payMethod := req.PaymentMethod
	if payMethod != "qris" && payMethod != "bank_transfer" {
		payMethod = "bank_transfer"
	}

	order := database.Order{
		OrderNo:       orderNo,
		PlayerID:      playerID,
		ServerID:      serverID,
		PlayerName:    pName,
		ProductID:     productID,
		ProductName:   prodName,
		Diamond:       diamondVal,
		Price:         priceVal,
		UniqueCode:    uniqueCode,
		TotalBayar:    totalBayar,
		BankName:      bankName,
		BankAccount:   bankAccount,
		BankHolder:    bankHolder,
		Whatsapp:      whatsapp,
		PaymentMethod: payMethod,
		Status:        "PENDING",
		UserID:        userID,
	}

	err = database.CreateOrder(order)
	if err != nil {
		log.Printf("[ORDER-ERROR] Gagal menyimpan order: %v\n", err)
		return c.JSON(fiber.Map{"success": false, "message": "Gagal menyimpan transaksi: " + err.Error()})
	}

	log.Printf("[ORDER-CREATED] %s | %s (%s/%s) | %d💎 | Rp%d\n", orderNo, pName, playerID, serverID, diamondVal, totalBayar)

	// Buat pesan WhatsApp otomatis
	waAdminNumber := os.Getenv("WHATSAPP_NUMBER")
	if waAdminNumber == "" {
		waAdminNumber = "6281234567890"
	}

	waMsg := fmt.Sprintf(
		"Halo Admin, saya sudah transfer untuk order top-up %s:\n\n"+
			"📋 No. Order: %s\n"+
			"👤 Player: %s (ID: %s | Server: %s)\n"+
			"💎 Paket: %s\n"+
			"💰 Total Transfer: Rp %s\n"+
			"🏦 Ke: %s %s a/n %s\n\n"+
			"Mohon segera dikonfirmasi. Terima kasih!",
		gameName, orderNo, pName, playerID, serverID, prodName,
		formatRupiah(totalBayar), bankName, bankAccount, bankHolder,
	)

	waURL := fmt.Sprintf("https://wa.me/%s?text=%s", waAdminNumber, url.QueryEscape(waMsg))

	return c.JSON(fiber.Map{
		"success": true,
		"order": fiber.Map{
			"order_no":       orderNo,
			"player_name":    pName,
			"player_id":      playerID,
			"server_id":      serverID,
			"product_name":   prodName,
			"diamond":        diamondVal,
			"price":          priceVal,
			"unique_code":    uniqueCode,
			"total_bayar":    totalBayar,
			"bank_name":      bankName,
			"bank_account":   bankAccount,
			"bank_holder":    bankHolder,
			"payment_method": payMethod,
			"status":         "PENDING",
		},
		"whatsapp_url": waURL,
	})
}

// GET /api/order/:order_no
func GetOrder(c *fiber.Ctx) error {
	orderNo := strings.ToUpper(c.Params("order_no"))
	order, err := database.GetOrderByOrderNo(orderNo)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Order tidak ditemukan"})
	}

	// Sembunyikan informasi privat dari customer luar
	return c.JSON(fiber.Map{
		"success": true,
		"order": fiber.Map{
			"order_no":       order.OrderNo,
			"player_name":    order.PlayerName,
			"player_id":      order.PlayerID,
			"product_name":   order.ProductName,
			"diamond":        order.Diamond,
			"total_bayar":    order.TotalBayar,
			"bank_name":      order.BankName,
			"bank_account":   order.BankAccount,
			"bank_holder":    order.BankHolder,
			"payment_method": order.PaymentMethod,
			"status":         order.Status,
			"created_at":     order.CreatedAt,
			"confirmed_at":   order.ConfirmedAt.String,
			"completed_at":   order.CompletedAt.String,
		},
	})
}

func formatRupiah(amount int) string {
	str := strconv.Itoa(amount)
	var result []string
	for len(str) > 3 {
		result = append([]string{str[len(str)-3:]}, result...)
		str = str[:len(str)-3]
	}
	result = append([]string{str}, result...)
	return strings.Join(result, ".")
}
