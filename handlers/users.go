// users.go
package handlers

import (
	"backend/models"
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	UserID       int     `json:"user_id"`
	FullName     string  `json:"full_name"`
	Phone        string  `json:"no_hp"`
	PasswordUser string  `json:"-"`
	ProfileImage *string `json:"profile_image,omitempty"`
}

func GetUsers(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token tidak ditemukan "})
		}

		// Format token biasanya: "Bearer <token>"
		var jwtToken string
		_, err := fmt.Sscanf(tokenString, "Bearer %s", &jwtToken)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid token format"})
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		userID := int(claims["user_id"].(float64))
		var user models.User

		err = db.QueryRow(`SELECT user_id, full_name, no_hp, profile_image FROM users WHERE user_id=$1`, userID).
			Scan(&user.UserID, &user.FullName, &user.Phone, &user.ProfileImage)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found"})
		}

		return c.JSON(fiber.Map{"user": user})
	}
}

func CreateUser(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		u := new(User)
		if err := c.BodyParser(u); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
		}
		_, err := db.Exec(`INSERT INTO users (full_name, no_hp, password_user, profile_image) VALUES ($1,$2,$3,$4)`,
			u.FullName, u.Phone, u.PasswordUser, u.ProfileImage)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(fiber.Map{"message": "User berhasil dibuat"})
	}
}
