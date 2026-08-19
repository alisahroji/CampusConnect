package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email       string `gorm:"unique;not null"`
	Name        string `gorm:"not null"`
	PictureURL  string
	Role        string `gorm:"default:'Student'"`
	Bio         string // Kolom baru
	Skills      string // Kolom baru
	GithubURL   string // Kolom baru
	LinkedinURL string // Kolom baru
}

func connectDB() {
	// DSN untuk GORM
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
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

	log.Println("Berhasil terhubung ke PostgreSQL!")

	// Jalankan Migration otomatis via golang-migrate dengan format URL yang benar
	runMigrations()
}

func runMigrations() {
	// Format URL DSN khusus untuk golang-migrate (postgres://user:pass@host:port/dbname?sslmode=disable)
	migrationDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	m, err := migrate.New(
		"file://db/migrations",
		migrationDSN,
	)
	if err != nil {
		log.Printf("Gagal menginisialisasi migrasi: %v\n", err)
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("Info migrasi: %v\n", err)
	} else {
		log.Println("Skema migrasi database berhasil diterapkan dengan bersih!")
	}
}
