package appsetup

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"ml-topup-v2/database"
	"ml-topup-v2/handlers"
	"ml-topup-v2/middleware"
)

//go:embed public
var embedPublic embed.FS

//go:embed admin-views
var embedAdmin embed.FS

//go:embed orders.db
var embedDB []byte


func Setup() *fiber.App {
	// Load environment variables dari .env
	_ = godotenv.Load() // Ignore error on production/Vercel where .env file is missing

	// Get subdirectory FS for public
	publicFS, err := fs.Sub(embedPublic, "public")
	if err != nil {
		log.Fatalf("[FATAL] Gagal inisialisasi public embed FS: %v\n", err)
	}

	// Get subdirectory FS for admin-views
	adminFS, err := fs.Sub(embedAdmin, "admin-views")
	if err != nil {
		log.Fatalf("[FATAL] Gagal inisialisasi admin embed FS: %v\n", err)
	}

	// Detect if running on Vercel or AWS Lambda
	dbPath := "./orders.db"
	if os.Getenv("VERCEL") != "" || os.Getenv("NOW_REGION") != "" || os.Getenv("LAMBDA_TASK_ROOT") != "" {
		// Write the embedded DB to /tmp/orders.db so it's writable
		err := os.WriteFile("/tmp/orders.db", embedDB, 0644)
		if err != nil {
			log.Printf("[WARNING] Could not write to /tmp/orders.db: %v", err)
		}
		
		dbPath = "/tmp/orders.db"
		log.Println("[INFO] Serverless environment detected. Embedded DB written to: " + dbPath)
	}

	// Inisialisasi Database SQLite
	err = database.InitDB(dbPath)
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

	// 2. Admin Frontend (Hidden & Secret Path)
	adminSecretPath := os.Getenv("ADMIN_SECRET_PATH")
	if adminSecretPath == "" {
		adminSecretPath = "kelola-dabrong-99" // Default rahasia jika tidak di-set
	}
	
	log.Printf("[SERVER] Halaman admin tersembunyi dapat diakses di: /%s\n", adminSecretPath)
	
	// Serve static files admin pada path rahasia
	app.Use("/"+adminSecretPath, filesystem.New(filesystem.Config{
		Root:   http.FS(adminFS),
		Browse: false,
		Index:  "index.html",
	}))

	// Redirect path umum /admin ke 404 Not Found
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString("404 Page Not Found")
	})

	// Route status tanpa ekstensi .html
	app.Get("/status", func(c *fiber.Ctx) error {
		fileContent, err := fs.ReadFile(publicFS, "status.html")
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("404 Page Not Found")
		}
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.Send(fileContent)
	})

	// ============================================================
	//  API ROUTES
	// ============================================================

	api := app.Group("/api")

	// Debug endpoint to list file structure at runtime
	api.Get("/debug-files", func(c *fiber.Ctx) error {
		cwd, _ := os.Getwd()
		var files []string
		_ = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				files = append(files, path+"/")
			} else {
				files = append(files, path)
			}
			return nil
		})
		return c.JSON(fiber.Map{
			"cwd":   cwd,
			"files": files,
		})
	})

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

	// 1. Customer Frontend (Public) - registered last as wildcard fallback
	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.FS(publicFS),
		Browse: false,
		Index:  "index.html",
	}))

	return app
}
