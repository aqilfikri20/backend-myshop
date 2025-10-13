package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

type Product struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	ImageURL    *string `json:"image_url,omitempty"`
	CategoryID  *int    `json:"category_id,omitempty"`
	Category    *string `json:"category_name,omitempty"`
	StoreID     *int    `json:"store_id,omitempty"`
	Store       *string `json:"store_name,omitempty"`
}

func GetProducts(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query(`
			SELECT p.product_id, p.product_name, p.price, p.image_url,
			       p.category_id, c.category_name, p.store_id, s.store_name
			FROM products p
			LEFT JOIN categories c ON c.category_id = p.category_id
			LEFT JOIN stores s ON s.store_id = p.store_id
			ORDER BY p.product_id`)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var list []Product
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ProductID, &p.ProductName, &p.Price, &p.ImageURL,
				&p.CategoryID, &p.Category, &p.StoreID, &p.Store); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			list = append(list, p)
		}
		return c.JSON(list)
	}
}

func CreateProduct(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		p := new(Product)
		if err := c.BodyParser(p); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
		}
		_, err := db.Exec(`INSERT INTO products (product_name, price, image_url, category_id, store_id)
		                   VALUES ($1,$2,$3,$4,$5)`,
			p.ProductName, p.Price, p.ImageURL, p.CategoryID, p.StoreID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(fiber.Map{"message": "product created"})
	}
}

func GetProduct(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var p Product
		err := db.QueryRow(`
			SELECT p.product_id, p.product_name, p.price, p.image_url,
			       p.category_id, c.category_name, p.store_id, s.store_name
			FROM products p
			LEFT JOIN categories c ON c.category_id = p.category_id
			LEFT JOIN stores s ON s.store_id = p.store_id
			WHERE p.product_id = $1`, id).
			Scan(&p.ProductID, &p.ProductName, &p.Price, &p.ImageURL, &p.CategoryID, &p.Category, &p.StoreID, &p.Store)

		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "product not found"})
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(p)
	}
}

func UpdateProduct(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		p := new(Product)
		if err := c.BodyParser(p); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
		}
		_, err := db.Exec(`UPDATE products SET product_name=$1, price=$2, image_url=$3, category_id=$4, store_id=$5 WHERE product_id=$6`,
			p.ProductName, p.Price, p.ImageURL, p.CategoryID, p.StoreID, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"message": "product updated"})
	}
}

func DeleteProduct(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		_, err := db.Exec(`DELETE FROM products WHERE product_id=$1`, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(204)
	}
}
