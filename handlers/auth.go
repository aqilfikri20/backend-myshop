package handlers

import (
	"backend/models"
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Simpan secret di .env di produksi
var jwtSecret = []byte("JWT_SECRET")

type RegisterInput struct {
	FullName string `json:"full_name"`
	Phone    string `json:"no_hp"`
	Password string `json:"password"`
}

type LoginInput struct {
	Phone    string `json:"no_hp"`
	Password string `json:"password"`
}

// ✅ REGISTER TANPA HASH
func RegisterUser(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input RegisterInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON input"})
		}

		input.FullName = strings.TrimSpace(input.FullName)
		input.Phone = strings.TrimSpace(input.Phone)

		if input.FullName == "" || input.Phone == "" || input.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All fields are required"})
		}

		var exists bool
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE no_hp=$1)`, input.Phone).Scan(&exists)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}
		if exists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Phone number already registered"})
		}

		// Simpan password langsung (⚠️ hanya untuk testing!)
		_, err = db.Exec(`INSERT INTO users (full_name, no_hp, password_user) VALUES ($1, $2, $3)`,
			input.FullName, input.Phone, input.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to register user"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
	}
}

// ✅ LOGIN TANPA HASH
func LoginUser(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input LoginInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON input"})
		}

		input.Phone = strings.TrimSpace(input.Phone)
		if input.Phone == "" || input.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Phone and password are required"})
		}

		var user models.User
		err := db.QueryRow(`
			SELECT user_id, full_name, no_hp, password_user, profile_image
			FROM users WHERE no_hp=$1
		`, input.Phone).Scan(&user.UserID, &user.FullName, &user.Phone, &user.PasswordUser, &user.ProfileImage)
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid phone or password"})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}

		// Bandingkan langsung plaintext password
		if user.PasswordUser != input.Password {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid phone or password"})
		}

		claims := jwt.MapClaims{
			"user_id":   user.UserID,
			"full_name": user.FullName,
			"phone":     user.Phone,
			"exp":       time.Now().Add(72 * time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		user.PasswordUser = ""
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Login successful",
			"token":   signedToken,
			"user":    user,
		})
	}
}
