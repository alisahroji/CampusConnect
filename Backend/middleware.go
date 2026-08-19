package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth adalah "satpam" yang mengecek validitas Access Token
func RequireAuth(c *gin.Context) {
	// 1. Ambil header Authorization dari request Frontend
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Token tidak ditemukan"})
		c.Abort() // Hentikan proses, jangan lanjut ke fungsi utama
		return
	}

	// 2. Pisahkan kata "Bearer " dari token aslinya
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	// 3. Parse dan validasi keaslian token menggunakan rahasia kita (.env)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Pastikan algoritma enkripsinya benar
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode enkripsi tidak valid")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Token tidak valid atau rusak"})
		c.Abort()
		return
	}

	// 4. Ambil isi data dari dalam token (Claims)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Gagal membaca data token"})
		c.Abort()
		return
	}

	// 5. Cek apakah token sudah kedaluwarsa (expired)
	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Token sudah kedaluwarsa"})
		c.Abort()
		return
	}

	// 6. Cari user di database berdasarkan ID (sub) yang ada di token
	var user User
	DB.First(&user, "id = ?", claims["sub"])

	if user.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: User tidak ditemukan di sistem"})
		c.Abort()
		return
	}

	// 7. Simpan data profil user ke dalam memori context Gin
	// Ini agar endpoint selanjutnya tahu siapa yang sedang melakukan request
	c.Set("currentUser", user)
	c.Set("userID", user.ID)

	// 8. Silakan masuk! Lanjut ke endpoint yang dituju
	c.Next()
}
