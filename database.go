package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func InitDB() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	ssl := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, ssl,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Gagal koneksi database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Tidak bisa ping database: %v", err)
	}

	DB = db
	log.Println("Koneksi database berhasil")
}
