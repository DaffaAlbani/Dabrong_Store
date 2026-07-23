package handlers

import (
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"ml-topup-v2/apigames"
	"ml-topup-v2/database"
	"ml-topup-v2/middleware"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Whatsapp string `json:"whatsapp"`
}

// POST /api/member/register
func MemberRegister(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format request tidak valid"})
	}

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := req.Password
	whatsapp := strings.TrimSpace(req.Whatsapp)

	if username == "" || email == "" || password == "" || whatsapp == "" {
		return c.JSON(fiber.Map{"success": false, "message": "Semua kolom wajib diisi"})
	}

	// Validasi WhatsApp
	waRegexp := regexp.MustCompile(`^(08|628)\d{8,12}$`)
	if !waRegexp.MatchString(whatsapp) {
		return c.JSON(fiber.Map{"success": false, "message": "Nomor WhatsApp tidak valid"})
	}

	// Cek apakah username sudah dipakai
	if _, err := database.GetUserByUsername(username); err == nil {
		return c.JSON(fiber.Map{"success": false, "message": "Username sudah terdaftar"})
	}

	// Cek apakah email sudah dipakai
	if _, err := database.GetUserByEmail(email); err == nil {
		return c.JSON(fiber.Map{"success": false, "message": "Email sudah terdaftar"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal mengenkripsi password"})
	}

	user := database.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Whatsapp: whatsapp,
		Role:     "member",
		Saldo:    0,
	}

	err = database.CreateUser(user)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal melakukan pendaftaran: " + err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Pendaftaran berhasil! Silakan login."})
}

type LoginRequest struct {
	Identity string `json:"identity"` // username atau email
	Password string `json:"password"`
}

// POST /api/member/login
func MemberLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Format request tidak valid"})
	}

	identity := strings.TrimSpace(req.Identity)
	password := req.Password

	if identity == "" || password == "" {
		return c.JSON(fiber.Map{"success": false, "message": "Username/Email dan Password wajib diisi"})
	}

	adminSecretPath := os.Getenv("ADMIN_SECRET_PATH")
	if adminSecretPath == "" {
		adminSecretPath = "kelola-dabrong-99"
	}

	// 1. CEK KREDENSIAL ADMIN DARI ENV / DEFAULT
	envAdminUser := os.Getenv("ADMIN_USERNAME")
	envAdminPass := os.Getenv("ADMIN_PASSWORD")
	if envAdminUser == "" {
		envAdminUser = "admin_dabrong"
	}
	if envAdminPass == "" {
		envAdminPass = "Rahasia123!"
	}

	if identity == envAdminUser && password == envAdminPass {
		token, err := middleware.GenerateToken(0, envAdminUser, "admin")
		if err != nil {
			return c.JSON(fiber.Map{"success": false, "message": "Gagal membuat sesi login admin"})
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

		c.Cookie(&fiber.Cookie{
			Name:     "member_token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			Secure:   os.Getenv("NODE_ENV") == "production",
			SameSite: "Lax",
			Path:     "/",
		})

		return c.JSON(fiber.Map{
			"success":    true,
			"message":    "Login admin berhasil!",
			"role":       "admin",
			"admin_path": "/" + adminSecretPath,
		})
	}

	// 2. CEK KREDENSIAL KE DATABASE USERS
	var user database.User
	var err error

	if strings.Contains(identity, "@") {
		user, err = database.GetUserByEmail(strings.ToLower(identity))
	} else {
		user, err = database.GetUserByUsername(identity)
	}

	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Akun tidak ditemukan"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Password salah"})
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal membuat sesi login"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "member_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   os.Getenv("NODE_ENV") == "production",
		SameSite: "Lax",
		Path:     "/",
	})

	if user.Role == "admin" {
		c.Cookie(&fiber.Cookie{
			Name:     "admin_token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			Secure:   os.Getenv("NODE_ENV") == "production",
			SameSite: "Lax",
			Path:     "/",
		})

		return c.JSON(fiber.Map{
			"success":    true,
			"message":    "Login admin berhasil!",
			"role":       "admin",
			"admin_path": "/" + adminSecretPath,
			"user": fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"whatsapp": user.Whatsapp,
				"role":     user.Role,
				"saldo":    user.Saldo,
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login berhasil!",
		"role":    "member",
		"user": fiber.Map{
			"username": user.Username,
			"email":    user.Email,
			"whatsapp": user.Whatsapp,
			"role":     user.Role,
			"saldo":    user.Saldo,
		},
	})
}

// POST /api/member/logout
func MemberLogout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "member_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Path:     "/",
	})
	return c.JSON(fiber.Map{"success": true, "message": "Logout berhasil"})
}

// GET /api/member/profile
func MemberProfile(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	
	user, err := database.GetUserByUsername(username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "User tidak ditemukan"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"user": fiber.Map{
			"username": user.Username,
			"email":    user.Email,
			"whatsapp": user.Whatsapp,
			"role":     user.Role,
			"saldo":    user.Saldo,
		},
	})
}

// GET /api/member/orders
func MemberOrders(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	orders, err := database.GetOrdersByUserID(userID)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal mengambil riwayat transaksi"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"orders":  orders,
	})
}

// POST /api/member/pay-with-saldo
func PayWithSaldo(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	orderNo := strings.ToUpper(c.FormValue("order_no"))

	if orderNo == "" {
		return c.JSON(fiber.Map{"success": false, "message": "Nomor Order wajib diisi"})
	}

	// Ambil data order
	order, err := database.GetOrderByOrderNo(orderNo)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Order tidak ditemukan"})
	}

	if order.Status != "PENDING" {
		return c.JSON(fiber.Map{"success": false, "message": "Order sudah diproses atau dibatalkan"})
	}

	// Ambil data user
	row := database.DB.QueryRow("SELECT id, username, email, password, whatsapp, role, saldo FROM users WHERE id = ?", userID)
	var user database.User
	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Whatsapp, &user.Role, &user.Saldo)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Akun user tidak ditemukan"})
	}

	if user.Saldo < order.TotalBayar {
		return c.JSON(fiber.Map{"success": false, "message": "Saldo Anda tidak mencukupi untuk melakukan transaksi"})
	}

	// Potong saldo
	newSaldo := user.Saldo - order.TotalBayar
	err = database.UpdateUserSaldo(userID, newSaldo)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "Gagal mendebet saldo: " + err.Error()})
	}

	// Update status order menjadi PROSES
	_ = database.UpdateOrderStatus(orderNo, "PROSES", "SUCCESS_SALDO", "Dibayar menggunakan Saldo Akun")

	// Kirim transaksi secara otomatis ke Apigames
	log.Printf("[SALDO-PAY] Auto-checkout Apigames untuk Order %s...\n", orderNo)
	
	go func(o database.Order) {
		res, err := apigames.SendTransaction(o.OrderNo, o.ProductID, o.PlayerID, o.ServerID)
		if err != nil {
			log.Printf("[SALDO-PAY-ERROR] Gagal auto-checkout Apigames: %v\n", err)
			_ = database.UpdateOrderStatus(o.OrderNo, "PROSES", "FAILED_API", "Error API: "+err.Error())
			return
		}

		if apiMap, ok := res.(map[string]any); ok {
			statusVal, _ := apiMap["status"].(float64)
			rcVal, _ := apiMap["rc"].(float64)
			errorMsg, _ := apiMap["error_msg"].(string)
			
			if statusVal == 1 || rcVal == 200 {
				log.Printf("[SALDO-PAY-SUCCESS] Sukses auto-checkout Apigames untuk Order %s\n", o.OrderNo)
				_ = database.UpdateOrderStatus(o.OrderNo, "PROSES", "PROSES", "Sedang diproses oleh provider Apigames")
			} else {
				log.Printf("[SALDO-PAY-FAILED] Apigames gagal: %s\n", errorMsg)
				_ = database.UpdateOrderStatus(o.OrderNo, "PROSES", "FAILED_PROVIDER", errorMsg)
			}
		}
	}(order)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Pembayaran menggunakan saldo sukses! Pesanan sedang diproses.",
		"saldo":   newSaldo,
	})
}
