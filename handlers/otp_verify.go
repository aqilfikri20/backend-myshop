package handlers

import (
	"database/sql"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func VerifyOTP(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type Request struct {
			Phone string `json:"no_hp"`
			OTP   string `json:"otp"`
		}
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
		}

		var storedOTP string
		var expiresAt time.Time

		err := db.QueryRow(`SELECT otp, expires_at FROM otp_codes WHERE no_hp=$1`, req.Phone).
			Scan(&storedOTP, &expiresAt)
		if err == sql.ErrNoRows {
			return c.Status(400).JSON(fiber.Map{"error": "OTP not found"})
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}

		if time.Now().After(expiresAt) {
			return c.Status(400).JSON(fiber.Map{"error": "OTP expired"})
		}
		if storedOTP != req.OTP {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid OTP"})
		}

		var userID int
		err = db.QueryRow(`SELECT user_id FROM users WHERE no_hp=$1`, req.Phone).Scan(&userID)
		if err == sql.ErrNoRows {
			err = db.QueryRow(`INSERT INTO users (full_name, no_hp) VALUES ($1, $2) RETURNING user_id`,
				"User "+req.Phone[len(req.Phone)-4:], req.Phone).Scan(&userID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
			}
		}

		// Buat JWT
		claims := jwt.MapClaims{
			"user_id": userID,
			"no_hp":   req.Phone,
			"exp":     time.Now().Add(72 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

		return c.JSON(fiber.Map{
			"message": "OTP verified successfully",
			"token":   signed,
			"user": fiber.Map{
				"user_id": userID,
				"no_hp":   req.Phone,
			},
		})
	}
}
