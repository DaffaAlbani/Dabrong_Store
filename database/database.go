package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
"encoding/json"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

type Order struct {
	ID              int64          `json:"id"`
	OrderNo         string         `json:"order_no"`
	PlayerID        string         `json:"player_id"`
	ServerID        string         `json:"server_id"`
	PlayerName      string         `json:"player_name"`
	ProductID       string         `json:"product_id"`
	ProductName     string         `json:"product_name"`
	Diamond         int            `json:"diamond"`
	Price           int            `json:"price"`
	UniqueCode      int            `json:"unique_code"`
	TotalBayar      int            `json:"total_bayar"`
	BankName        string         `json:"bank_name"`
	BankAccount     string         `json:"bank_account"`
	BankHolder      string         `json:"bank_holder"`
	Whatsapp        string         `json:"whatsapp"`
	PaymentMethod   string         `json:"payment_method"`
	Status          string         `json:"status"`
	ApigamesRefID   sql.NullString `json:"apigames_ref_id"`
	ApigamesStatus  sql.NullString `json:"apigames_status"`
	ApigamesMessage sql.NullString `json:"apigames_message"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
	ConfirmedAt     sql.NullString `json:"confirmed_at"`
	CompletedAt     sql.NullString `json:"completed_at"`
	UserID          sql.NullInt64  `json:"user_id"`
}

type Product struct {
	ID            int64  `json:"id"`
	ProductID     string `json:"product_id"`
	ProductName   string `json:"product_name"`
	Category      string `json:"category"`
	Price         int    `json:"price"`
	OriginalPrice int    `json:"original_price"`
	Status        string `json:"status"`
	CachedAt      string `json:"cached_at"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	Whatsapp  string `json:"whatsapp"`
	Role      string `json:"role"`
	Saldo     int    `json:"saldo"`
	CreatedAt string `json:"created_at"`
}

type OrderStats struct {
	Total   int `json:"total"`
	Pending int `json:"pending"`
	Proses  int `json:"proses"`
	Sukses  int `json:"sukses"`
	Gagal   int `json:"gagal"`
	Revenue int `json:"revenue"`
}

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("gagal membuka database: %v", err)
	}

	// Optimize SQLite performance & concurrency
	_, _ = DB.Exec("PRAGMA journal_mode=WAL;")
	_, _ = DB.Exec("PRAGMA busy_timeout=5000;")
	_, _ = DB.Exec("PRAGMA synchronous=NORMAL;")

	// Buat tabel orders jika belum ada
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			order_no     TEXT UNIQUE NOT NULL,
			player_id    TEXT NOT NULL,
			server_id    TEXT NOT NULL,
			player_name  TEXT,
			product_id   TEXT NOT NULL,
			product_name TEXT NOT NULL,
			diamond      INTEGER NOT NULL,
			price        INTEGER NOT NULL,
			unique_code  INTEGER NOT NULL DEFAULT 0,
			total_bayar  INTEGER NOT NULL,
			bank_name    TEXT,
			bank_account TEXT,
			bank_holder  TEXT,
			whatsapp     TEXT,
			payment_method TEXT DEFAULT 'bank_transfer',
			status       TEXT NOT NULL DEFAULT 'PENDING',
			apigames_ref_id  TEXT,
			apigames_status  TEXT,
			apigames_message TEXT,
			created_at   TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
			updated_at   TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
			confirmed_at TEXT,
			completed_at TEXT,
			user_id      INTEGER
		);
	`)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel orders: %v", err)
	}

	// Buat tabel products_cache jika belum ada
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS products_cache (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id     TEXT UNIQUE NOT NULL,
			product_name   TEXT NOT NULL,
			category       TEXT,
			price          INTEGER NOT NULL,
			original_price INTEGER DEFAULT 0,
			status         TEXT,
			cached_at      TEXT NOT NULL DEFAULT (datetime('now', 'localtime'))
		);
	`)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel products_cache: %v", err)
	}

	// Buat tabel users jika belum ada
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			username     TEXT UNIQUE NOT NULL,
			email        TEXT UNIQUE NOT NULL,
			password     TEXT NOT NULL,
			whatsapp     TEXT NOT NULL,
			role         TEXT DEFAULT 'member',
			saldo        INTEGER DEFAULT 0,
			created_at   TEXT NOT NULL DEFAULT (datetime('now', 'localtime'))
		);
	`)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel users: %v", err)
	}

	// Jalankan migrasi kolom jika orders lama belum punya kolom user_id
	_, _ = DB.Exec("ALTER TABLE orders ADD COLUMN user_id INTEGER")
	_, _ = DB.Exec("ALTER TABLE products_cache ADD COLUMN original_price INTEGER DEFAULT 0")

	// Always sync products_cache from embedded products.json if available
	if len(EmbeddedProducts) > 0 {
		var embeddedList []Product
		if err := json.Unmarshal(EmbeddedProducts, &embeddedList); err == nil && len(embeddedList) > 0 {
			_, _ = DB.Exec("DELETE FROM products_cache")
			_ = CacheProducts(embeddedList)
			log.Printf("[DATABASE] Synced %d curated products from embedded products.json\n", len(embeddedList))
		}
	}

	log.Println("[DATABASE] SQLite berhasil terhubung & diinisialisasi dengan mode WAL.")
	return nil
}

// ============ USER QUERIES ============

func CreateUser(user User) error {
	_, err := DB.Exec(`
		INSERT INTO users (username, email, password, whatsapp, role, saldo)
		VALUES (?, ?, ?, ?, ?, ?)
	`, user.Username, user.Email, user.Password, user.Whatsapp, user.Role, user.Saldo)
	return err
}

func GetUserByUsername(username string) (User, error) {
	row := DB.QueryRow("SELECT id, username, email, password, whatsapp, role, saldo, created_at FROM users WHERE username = ?", username)
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Whatsapp, &u.Role, &u.Saldo, &u.CreatedAt)
	return u, err
}

func GetUserByEmail(email string) (User, error) {
	row := DB.QueryRow("SELECT id, username, email, password, whatsapp, role, saldo, created_at FROM users WHERE email = ?", email)
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Whatsapp, &u.Role, &u.Saldo, &u.CreatedAt)
	return u, err
}

func UpdateUserSaldo(userID int64, newSaldo int) error {
	_, err := DB.Exec("UPDATE users SET saldo = ? WHERE id = ?", newSaldo, userID)
	return err
}

// ============ ORDER QUERIES ============

func CreateOrder(o Order) error {
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(`
		INSERT INTO orders (
			order_no, player_id, server_id, player_name,
			product_id, product_name, diamond, price,
			unique_code, total_bayar, bank_name, bank_account,
			bank_holder, whatsapp, payment_method, status,
			created_at, updated_at, user_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, o.OrderNo, o.PlayerID, o.ServerID, o.PlayerName,
		o.ProductID, o.ProductName, o.Diamond, o.Price,
		o.UniqueCode, o.TotalBayar, o.BankName, o.BankAccount,
		o.BankHolder, o.Whatsapp, o.PaymentMethod, o.Status,
		nowStr, nowStr, o.UserID)
	return err
}

func GetOrderById(id int64) (Order, error) {
	row := DB.QueryRow("SELECT * FROM orders WHERE id = ?", id)
	return scanOrder(row)
}

func GetOrderByOrderNo(orderNo string) (Order, error) {
	row := DB.QueryRow("SELECT * FROM orders WHERE order_no = ?", orderNo)
	return scanOrder(row)
}

func GetAllOrders(limit, offset int) ([]Order, error) {
	rows, err := DB.Query("SELECT * FROM orders ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func GetOrdersByStatus(status string) ([]Order, error) {
	rows, err := DB.Query("SELECT * FROM orders WHERE status = ? ORDER BY id DESC", status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func GetOrdersByUserID(userID int64) ([]Order, error) {
	rows, err := DB.Query("SELECT * FROM orders WHERE user_id = ? ORDER BY id DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func UpdateOrderStatus(orderNo, status, apigamesStatus, apigamesMessage string) error {
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	var err error
	if status == "SUCCESS" {
		_, err = DB.Exec(`
			UPDATE orders
			SET status = ?, apigames_status = ?, apigames_message = ?, updated_at = ?, completed_at = ?
			WHERE order_no = ?
		`, status, apigamesStatus, apigamesMessage, nowStr, nowStr, orderNo)
	} else if status == "PROSES" {
		_, err = DB.Exec(`
			UPDATE orders
			SET status = ?, apigames_status = ?, apigames_message = ?, updated_at = ?, confirmed_at = ?
			WHERE order_no = ?
		`, status, apigamesStatus, apigamesMessage, nowStr, nowStr, orderNo)
	} else {
		_, err = DB.Exec(`
			UPDATE orders
			SET status = ?, apigames_status = ?, apigames_message = ?, updated_at = ?
			WHERE order_no = ?
		`, status, apigamesStatus, apigamesMessage, nowStr, orderNo)
	}
	return err
}

func GetOrderStats() (OrderStats, error) {
	var s OrderStats
	
	rowTotal := DB.QueryRow("SELECT COUNT(*) FROM orders")
	_ = rowTotal.Scan(&s.Total)
	
	rowPending := DB.QueryRow("SELECT COUNT(*) FROM orders WHERE status = 'PENDING'")
	_ = rowPending.Scan(&s.Pending)

	rowProses := DB.QueryRow("SELECT COUNT(*) FROM orders WHERE status = 'PROSES'")
	_ = rowProses.Scan(&s.Proses)

	rowSukses := DB.QueryRow("SELECT COUNT(*) FROM orders WHERE status = 'SUCCESS'")
	_ = rowSukses.Scan(&s.Sukses)

	rowGagal := DB.QueryRow("SELECT COUNT(*) FROM orders WHERE status = 'FAILED'")
	_ = rowGagal.Scan(&s.Gagal)

	rowRevenue := DB.QueryRow("SELECT COALESCE(SUM(total_bayar), 0) FROM orders WHERE status = 'SUCCESS'")
	_ = rowRevenue.Scan(&s.Revenue)

	return s, nil
}

// ============ PRODUCTS CACHE & CRUD ============

func CacheProducts(products []Product) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO products_cache (product_id, product_name, category, price, original_price, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range products {
		_, err = stmt.Exec(p.ProductID, p.ProductName, p.Category, p.Price, p.OriginalPrice, p.Status)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func GetCachedProducts() ([]Product, error) {
	rows, err := DB.Query("SELECT id, product_id, product_name, category, price, original_price, status, cached_at FROM products_cache ORDER BY price ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err = rows.Scan(&p.ID, &p.ProductID, &p.ProductName, &p.Category, &p.Price, &p.OriginalPrice, &p.Status, &p.CachedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func GetProductsByCategory(category string) ([]Product, error) {
	rows, err := DB.Query("SELECT id, product_id, product_name, category, price, original_price, status, cached_at FROM products_cache WHERE category = ? ORDER BY price ASC", category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err = rows.Scan(&p.ID, &p.ProductID, &p.ProductName, &p.Category, &p.Price, &p.OriginalPrice, &p.Status, &p.CachedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func ClearProductsCache() error {
	_, err := DB.Exec("DELETE FROM products_cache")
	return err
}

func AddProduct(p Product) (Product, error) {
	_, err := DB.Exec(`
		INSERT INTO products_cache (product_id, product_name, category, price, original_price, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, p.ProductID, p.ProductName, p.Category, p.Price, p.OriginalPrice, p.Status)
	if err != nil {
		return p, err
	}

	row := DB.QueryRow("SELECT id, product_id, product_name, category, price, original_price, status, cached_at FROM products_cache WHERE product_id = ?", p.ProductID)
	var created Product
	err = row.Scan(&created.ID, &created.ProductID, &created.ProductName, &created.Category, &created.Price, &created.OriginalPrice, &created.Status, &created.CachedAt)
	return created, err
}

func UpdateProduct(productID string, p Product) (Product, error) {
	_, err := DB.Exec(`
		UPDATE products_cache
		SET product_name = ?, category = ?, price = ?, original_price = ?, status = ?
		WHERE product_id = ?
	`, p.ProductName, p.Category, p.Price, p.OriginalPrice, p.Status, productID)
	if err != nil {
		return p, err
	}

	row := DB.QueryRow("SELECT id, product_id, product_name, category, price, original_price, status, cached_at FROM products_cache WHERE product_id = ?", productID)
	var updated Product
	err = row.Scan(&updated.ID, &updated.ProductID, &updated.ProductName, &updated.Category, &updated.Price, &updated.OriginalPrice, &updated.Status, &updated.CachedAt)
	return updated, err
}

func DeleteProduct(productID string) error {
	_, err := DB.Exec("DELETE FROM products_cache WHERE product_id = ?", productID)
	return err
}

// ============ HELPER SCANNERS ============

type RowOrRows interface {
	Scan(dest ...any) error
}

func scanOrder(row RowOrRows) (Order, error) {
	var o Order
	var refID, statusMsg, refMsg sql.NullString
	err := row.Scan(
		&o.ID, &o.OrderNo, &o.PlayerID, &o.ServerID, &o.PlayerName,
		&o.ProductID, &o.ProductName, &o.Diamond, &o.Price,
		&o.UniqueCode, &o.TotalBayar, &o.BankName, &o.BankAccount,
		&o.BankHolder, &o.Whatsapp, &o.PaymentMethod, &o.Status,
		&refID, &statusMsg, &refMsg,
		&o.CreatedAt, &o.UpdatedAt, &o.ConfirmedAt, &o.CompletedAt, &o.UserID,
	)
	if err == nil {
		o.ApigamesRefID = refID
		o.ApigamesStatus = statusMsg
		o.ApigamesMessage = refMsg
	}
	return o, err
}
