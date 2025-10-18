package handlers

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Simpan secret di .env agar lebih aman
var jwtSecret = []byte("super_secret")

type RegisterInput struct {
	FullName string `json:"full_name"`
	Gmail    string `json:"gmail"`
	Password string `json:"password"`
}

type LoginInput struct {
	Gmail    string `json:"gmail"`
	Password string `json:"password"`
}

// ✅ REGISTER HANDLER
func RegisterUser(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input RegisterInput

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON input"})
		}

		// Validasi ringan
		input.FullName = strings.TrimSpace(input.FullName)
		input.Gmail = strings.TrimSpace(input.Gmail)
		if input.FullName == "" || input.Gmail == "" || input.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All fields are required"})
		}

		// Cek apakah user sudah ada
		var exists bool
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE gmail=$1)`, input.Gmail).Scan(&exists)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}
		if exists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email already registered"})
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
		}

		// Insert user
		_, err = db.Exec(`INSERT INTO users (full_name, gmail, password_user) VALUES ($1, $2, $3)`,
			input.FullName, input.Gmail, string(hashedPassword))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to register user"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
	}
}

// ✅ LOGIN HANDLER
func LoginUser(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input LoginInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON input"})
		}

		input.Gmail = strings.TrimSpace(input.Gmail)
		if input.Gmail == "" || input.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email and password are required"})
		}

		var (
			id             int
			name           string
			hashedPassword string
		)

		err := db.QueryRow(`SELECT user_id, full_name, password_user FROM users WHERE gmail=$1`, input.Gmail).
			Scan(&id, &name, &hashedPassword)
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}

		// Bandingkan password
		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
		}

		// Generate JWT token
		claims := jwt.MapClaims{
			"user_id": id,
			"name":    name,
			"email":   input.Gmail,
			"exp":     time.Now().Add(time.Hour * 72).Unix(), // berlaku 3 hari
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Login successful",
			"token":   signedToken,
			"user": fiber.Map{
				"user_id":   id,
				"full_name": name,
				"gmail":     input.Gmail,
			},
		})
	}
}
