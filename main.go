// menjalankan server web menggunakan framework Fiber
package main

import (
	"backend/handlers"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	InitDB()
	defer DB.Close()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000", // Next.js frontend
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	// --- ROUTE PRODUK ---
	app.Get("/api/products", handlers.GetProducts(DB))
	app.Get("/api/products/:id", handlers.GetProduct(DB))
	app.Post("/api/products", handlers.CreateProduct(DB))
	app.Put("/api/products/:id", handlers.UpdateProduct(DB))
	app.Delete("/api/products/:id", handlers.DeleteProduct(DB))

	// --- ROUTE USER ---
	app.Get("/api/users", handlers.GetUsers(DB))
	app.Post("/api/users", handlers.CreateUser(DB))

	// --- LOGIN & REGISTER MANUAL ---
	app.Post("/api/register", handlers.RegisterUser(DB))
	app.Post("/api/login", handlers.LoginUser(DB))

	// --- LOGIN GOOGLE (OAuth2) ---
	app.Get("/api/auth/google/login", handlers.GoogleLogin)
	app.Get("/api/auth/google/callback", handlers.GoogleCallback(DB))

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
