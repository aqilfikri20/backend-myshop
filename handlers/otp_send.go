package handlers

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SendOTP(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type Request struct {
			Phone string `json:"no_hp"`
		}

		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
		}

		if req.Phone == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Phone number required"})
		}

		otp := fmt.Sprintf("%06d", rand.Intn(1000000))
		expiry := time.Now().Add(5 * time.Minute)

		// Simpan atau update OTP
		_, err := db.Exec(`INSERT INTO otp_codes (no_hp, otp, expires_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (no_hp) DO UPDATE SET otp=$2, expires_at=$3`,
			req.Phone, otp, expiry)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}

		// Kirim OTP ke WhatsApp via Fonnte
		token := os.Getenv("FONNTE_TOKEN")
		message := fmt.Sprintf("Kode OTP kamu: %s (berlaku 5 menit). Jangan bagikan ke siapa pun.", otp)

		formData := url.Values{}
		formData.Set("target", req.Phone)
		formData.Set("message", message)

		reqFonnte, _ := http.NewRequest("POST", "https://api.fonnte.com/send", strings.NewReader(formData.Encode()))
		reqFonnte.Header.Add("Authorization", token)
		reqFonnte.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(reqFonnte)
		if err != nil || resp.StatusCode != 200 {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to send WhatsApp message"})
		}
		defer resp.Body.Close()

		return c.JSON(fiber.Map{
			"message": "OTP sent to WhatsApp",
			"no_hp":   req.Phone,
		})
	}
}
