// mengelola request terkait user
package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
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
		rows, err := db.Query(`SELECT user_id, full_name, no_hp, profile_image FROM users ORDER BY user_id`)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.UserID, &u.FullName, &u.Phone, &u.ProfileImage); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			users = append(users, u)
		}
		return c.JSON(users)
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
