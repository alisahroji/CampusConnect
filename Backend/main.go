package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config

// Struct untuk menampung balasan profil dari Google
type GoogleUserResult struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func init() {
	// Memuat variabel dari file .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Konfigurasi OAuth Google
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func main() {
	r := gin.Default()

	connectDB()

	// Endpoint dasar
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Backend OK!"})
	})

	// Rute untuk menginisiasi Login Google
	r.GET("/api/auth/google/login", func(c *gin.Context) {
		// State ini fungsinya sebagai token anti-CSRF (untuk sekarang kita isi string statis dulu)
		oauthStateString := "random-state-string"

		// Membuat URL menuju halaman persetujuan Google
		url := googleOauthConfig.AuthCodeURL(oauthStateString)

		// Redirect user ke URL Google tersebut
		c.Redirect(http.StatusTemporaryRedirect, url)
	})

	// Rute Callback setelah user memilih akun Google
	r.GET("/api/auth/google/callback", func(c *gin.Context) {
		// 1. Ambil authorization code dari URL
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found in URL"})
			return
		}

		// 2. Tukar kode tersebut dengan Access Token dari Google
		token, err := googleOauthConfig.Exchange(context.Background(), code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
			return
		}

		// 3. Gunakan Access Token untuk meminta data profil user
		response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
			return
		}
		defer response.Body.Close()

		// 4. Baca balasan dari Google
		userData, err := io.ReadAll(response.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
			return
		}

		// 5. Parse JSON ke dalam struct Go
		var googleUser GoogleUserResult
		if err := json.Unmarshal(userData, &googleUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
			return
		}

		// 6. Integrasi dengan Database (Mencari atau Membuat User Baru)
		var user User

		// Kita cari user berdasarkan email dari Google
		result := DB.Where("email = ?", googleUser.Email).First(&user)

		if result.Error != nil {
			// Jika error (kemungkinan besar karena data tidak ditemukan), kita daftarkan sebagai user baru!
			user = User{
				Email:      googleUser.Email,
				Name:       googleUser.Name,
				PictureURL: googleUser.Picture,
				// Kolom ID akan otomatis digenerate oleh PostgreSQL
				// Kolom Role otomatis menjadi 'Student'
			}

			// Simpan ke database
			if err := DB.Create(&user).Error; err != nil {
				log.Printf("Gagal menyimpan user baru: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data user ke database"})
				return
			}
			log.Println("User baru berhasil didaftarkan otomatis!")
		} else {
			log.Println("User lama berhasil login kembali!")
		}

		// 7. Terbitkan JWT (Access Token & Refresh Token)
		accessToken, refreshToken, err := GenerateTokens(user.ID)
		if err != nil {
			log.Printf("Gagal membuat token: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menerbitkan token autentikasi"})
			return
		}

		// 8. Kirim token dan data profil ke Frontend
		c.JSON(http.StatusOK, gin.H{
			"message":       "Autentikasi & Integrasi Database Sukses!",
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user":          user,
		})
	})

	log.Println("Server is running on port 8080...")
	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}