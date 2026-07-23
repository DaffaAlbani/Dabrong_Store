package middleware

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	rateLimitMap = make(map[string][]time.Time)
	rateLimitMu  sync.Mutex
)

// Clean old rate limit entries every 5 minutes
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			rateLimitMu.Lock()
			now := time.Now()
			for ip, times := range rateLimitMap {
				var validTimes []time.Time
				for _, t := range times {
					if now.Sub(t) < time.Minute {
						validTimes = append(validTimes, t)
					}
				}
				if len(validTimes) == 0 {
					delete(rateLimitMap, ip)
				} else {
					rateLimitMap[ip] = validTimes
				}
			}
			rateLimitMu.Unlock()
		}
	}()
}

// RateLimit: Maksimal 10 request per menit untuk IP tertentu
func RateLimit(c *fiber.Ctx) error {
	ip := c.IP()
	now := time.Now()

	rateLimitMu.Lock()
	times := rateLimitMap[ip]
	
	// Saring request yang terjadi dalam 1 menit terakhir
	var activeTimes []time.Time
	for _, t := range times {
		if now.Sub(t) < time.Minute {
			activeTimes = append(activeTimes, t)
		}
	}

	if len(activeTimes) >= 10 {
		rateLimitMu.Unlock()
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"success": false,
			"message": "Terlalu banyak permintaan. Silakan coba beberapa saat lagi (Anti-Spam).",
		})
	}

	activeTimes = append(activeTimes, now)
	rateLimitMap[ip] = activeTimes
	rateLimitMu.Unlock()

	return c.Next()
}

// SecurityHeaders middleware
func SecurityHeaders(c *fiber.Ctx) error {
	c.Set("X-Frame-Options", "DENY")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Set("Content-Security-Policy", "default-src 'self' https: 'unsafe-inline' 'unsafe-eval'; img-src 'self' data: https:; font-src 'self' data: https:;")
	return c.Next()
}

// Secret key untuk JWT signing
func getJWTSecret() []byte {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "dabrong-default-secret-key-change-this"
	}
	return []byte(secret)
}

// Generate JWT token untuk Admin atau Member
func GenerateToken(userID int64, username, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// AuthAdmin middleware
func AuthAdmin(c *fiber.Ctx) error {
	tokenStr := c.Cookies("admin_token")
	if tokenStr == "" {
		tokenStr = c.Cookies("member_token")
	}
	if tokenStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized. Silakan login kembali.",
		})
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return getJWTSecret(), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Token tidak valid atau kedaluwarsa.",
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["role"] != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Forbidden. Akses ditolak.",
		})
	}

	c.Locals("username", claims["username"].(string))
	c.Locals("role", claims["role"].(string))
	return c.Next()
}

// AuthMember middleware
func AuthMember(c *fiber.Ctx) error {
	tokenStr := c.Cookies("member_token")
	if tokenStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Silakan login terlebih dahulu.",
		})
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return getJWTSecret(), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Sesi Anda telah berakhir. Silakan login kembali.",
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Sesi tidak valid.",
		})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User ID tidak valid dalam token.",
		})
	}

	c.Locals("user_id", int64(userIDFloat))
	c.Locals("username", claims["username"].(string))
	c.Locals("role", claims["role"].(string))
	return c.Next()
}
