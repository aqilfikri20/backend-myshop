// menentukan struktur data untuk user, store, category, dan product
package main

import "time"

type User struct {
	UserID       int       `json:"user_id"`
	FullName     string    `json:"full_name"`
	Gmail        string    `json:"gmail"`
	PasswordUser string    `json:"-"`
	ProfileImage *string   `json:"profile_image,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Store struct {
	StoreID      int     `json:"store_id"`
	StoreName    string  `json:"store_name"`
	StoreAddress string  `json:"store_address"`
	StorePhone   string  `json:"store_phone"`
	OwnerID      int     `json:"owner_id"`
	OwnerName    *string `json:"owner_name,omitempty"`
}

type Category struct {
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
}

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
