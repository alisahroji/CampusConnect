package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Variabel global untuk menyimpan koneksi database
var DB *gorm.DB

// Struct User yang merepresentasikan tabel users hasil migrasi kita
type User struct {
	ID         string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email      string `gorm:"unique;not null"`
	Name       string `gorm:"not null"`
	PictureURL string
	Role       string `gorm:"default:'Student'"`
}

func connectDB() {
	// Merakit string koneksi (DSN) dari variabel .env
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	
	log.Println("Berhasil terhubung ke PostgreSQL menggunakan GORM!")
}