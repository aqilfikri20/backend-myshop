package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	},
	Endpoint: google.Endpoint,
}

// 🔹 STEP 1: Redirect user ke Google login page
func GoogleLogin(c *fiber.Ctx) error {
	conf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	url := conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

// 🔹 STEP 2: Handle callback dari Google
func GoogleCallback(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Query("code")
		if code == "" {
			return c.Status(400).SendString("No code provided")
		}

		conf := &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
		token, err := conf.Exchange(context.Background(), code)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to exchange token"})
		}

		client := googleOauthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal mendapatkan user info"})
		}
		defer resp.Body.Close()

		var userInfo struct {
			ID      string `json:"id"`
			Email   string `json:"email"`
			Name    string `json:"name"`
			Picture string `json:"picture"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decode user info"})
		}

		// Buat JWT setelah login
		claims := jwt.MapClaims{
			"email": userInfo.Email,
			"name":  userInfo.Name,
			"exp":   time.Now().Add(time.Hour * 72).Unix(),
		}
		appToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, _ := appToken.SignedString([]byte(os.Getenv("JWT_SECRET")))

		// ✅ Redirect ke frontend dengan token di query string
		frontendURL := os.Getenv("FRONTEND_URL") // contoh: http://localhost:3000
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}
		redirectURL := fmt.Sprintf("%s/?token=%s", frontendURL, url.QueryEscape(signed))
		return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
	}
}
