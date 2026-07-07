package appsetup

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"ml-topup-v2/database"
	"ml-topup-v2/handlers"
	"ml-topup-v2/middleware"
)

func Setup() *fiber.App {
	// Load environment variables dari .env
	_ = godotenv.Load() // Ignore error on production/Vercel where .env file is missing

	// Detect if running on Vercel or AWS Lambda
	dbPath := "./orders.db"
	if os.Getenv("VERCEL") != "" || os.Getenv("NOW_REGION") != "" || os.Getenv("LAMBDA_TASK_ROOT") != "" {
		dbPath = "file::memory:?cache=shared"
		log.Println("[INFO] Serverless environment detected. Database path set to in-memory: " + dbPath)
	}

	// Inisialisasi Database SQLite
	err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("[FATAL] Gagal inisialisasi database: %v\n", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "Dabrong Top-Up ML Golang Backend",
	})

	// Logger middleware untuk melihat request log di terminal
	app.Use(logger.New())

	// Security Headers
	app.Use(middleware.SecurityHeaders)

	// ============================================================
	//  STATIC ASSETS SERVING (Pemisahan Customer & Admin)
	// ============================================================

	// 1. Customer Frontend (Public)
	app.Static("/", "./public")
	
	// Route status tanpa ekstensi .html
	app.Get("/status", func(c *fiber.Ctx) error {
		return c.SendFile("./public/status.html")
	})

	// 2. Admin Frontend (Hidden & Secret Path)
	adminSecretPath := os.Getenv("ADMIN_SECRET_PATH")
	if adminSecretPath == "" {
		adminSecretPath = "kelola-dabrong-99" // Default rahasia jika tidak di-set
	}
	
	log.Printf("[SERVER] Halaman admin tersembunyi dapat diakses di: /%s\n", adminSecretPath)
	
	// Serve static files admin pada path rahasia
	app.Static("/"+adminSecretPath, "./admin-views")

	// Redirect path umum /admin ke 404 Not Found
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString("404 Page Not Found")
	})

	// ============================================================
	//  API ROUTES
	// ============================================================

	api := app.Group("/api")

	// 1. Public / Customer Endpoints (Diproteksi Rate Limiter)
	api.Get("/check-user", middleware.RateLimit, handlers.CheckUser)
	api.Get("/products", handlers.GetProducts)
	api.Post("/order", middleware.RateLimit, handlers.CreateOrder)
	api.Get("/order/:order_no", handlers.GetOrder) // Dukung GET /api/order/DMLxxx

	// 2. Customer Member Endpoints (Auth & Profile)
	api.Post("/member/register", handlers.MemberRegister)
	api.Post("/member/login", handlers.MemberLogin)
	api.Post("/member/logout", handlers.MemberLogout)
	api.Get("/member/profile", middleware.AuthMember, handlers.MemberProfile)
	api.Get("/member/orders", middleware.AuthMember, handlers.MemberOrders)
	api.Post("/member/pay-with-saldo", middleware.AuthMember, handlers.PayWithSaldo)

	// 3. Admin Endpoints
	api.Post("/admin/login", handlers.AdminLogin)
	api.Post("/admin/logout", handlers.AdminLogout)
	
	// Admin area yang dilindungi AuthAdmin middleware
	adminAPI := api.Group("/admin", middleware.AuthAdmin)
	adminAPI.Get("/me", handlers.AdminMe)
	adminAPI.Get("/stats", handlers.AdminStats)
	adminAPI.Get("/orders", handlers.AdminOrders)
	adminAPI.Get("/order/:order_no", handlers.AdminOrderDetail)
	adminAPI.Post("/confirm", handlers.AdminConfirmOrder)
	adminAPI.Post("/check-trx", handlers.AdminCheckTrx)
	adminAPI.Post("/reject", handlers.AdminRejectOrder)
	
	// CRUD Produk manual via Admin
	adminAPI.Post("/products", handlers.AdminAddProduct)
	adminAPI.Put("/products/:product_id", handlers.AdminUpdateProduct)
	adminAPI.Delete("/products/:product_id", handlers.AdminDeleteProduct)

	return app
}
