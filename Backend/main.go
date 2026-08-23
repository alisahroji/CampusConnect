package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"campusconnect/handler" // Sesuaikan dengan nama modul
	"campusconnect/repository"
	"campusconnect/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

// Struct untuk menerima request refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RequestOTPInput struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
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

	userRepo := repository.NewUserRepository(DB)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Izin khusus untuk frontend Vite
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

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

	// --- RUTE UNTUK PERPANJANG TOKEN ---
	r.POST("/api/auth/refresh", func(c *gin.Context) {
		var req RefreshTokenRequest

		// 1. Tangkap refresh token dari body request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token tidak boleh kosong"})
			return
		}

		// 2. Parse dan validasi Refresh Token menggunakan JWT_REFRESH_SECRET
		token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode enkripsi tidak valid")
			}
			return []byte(os.Getenv("JWT_REFRESH_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token tidak valid atau sudah kedaluwarsa, silakan login ulang"})
			return
		}

		// 3. Ambil data dari dalam token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal membaca data token"})
			return
		}

		// 4. Ambil ID User (sub)
		userID := claims["sub"].(string)

		// 5. Cetak pasangan token yang baru!
		newAccessToken, newRefreshToken, err := GenerateTokens(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token baru"})
			return
		}

		// 6. Kirim tiket baru tersebut
		c.JSON(http.StatusOK, gin.H{
			"message":       "Token berhasil diperbarui!",
			"access_token":  newAccessToken,
			"refresh_token": newRefreshToken,
		})
	})

	// --- RUTE YANG DILINDUNGI SATPAM ---
	// Perhatikan kita menyisipkan RequireAuth sebelum fungsi utamanya
	// --- RUTE YANG DILINDUNGI SATPAM (Clean Architecture) ---
	r.GET("/api/profile", RequireAuth, userHandler.GetProfile)
	r.PUT("/api/profile", RequireAuth, userHandler.UpdateProfile)
	r.POST("/api/profile/avatar", RequireAuth, userHandler.UploadAvatar)

	// --- RUTE UNTUK LOGOUT ---
	// Kita gunakan RequireAuth agar hanya orang yang sedang login yang bisa memanggil rute ini
	r.POST("/api/auth/logout", RequireAuth, func(c *gin.Context) {
		// Di tahap ini (Fase 1), backend hanya memberikan respons sukses.
		// (Di Fase lanjutan, kita akan memasukkan token ini ke Redis Blacklist di sini).

		c.JSON(http.StatusOK, gin.H{
			"message":     "Berhasil logout! Sesi diakhiri secara aman.",
			"instruction": "Frontend wajib menghapus access_token dan refresh_token dari storage lokal.",
		})
	})

	// --- RUTE REQUEST OTP ---
	r.POST("/api/auth/request-otp", func(c *gin.Context) {
		var input RequestOTPInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format email tidak valid"})
			return
		}

		// 1. Generate kode OTP acak 6 digit sederhana
		otpCode := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)

		// 2. Tentukan waktu kedaluwarsa (5 menit dari sekarang)
		expiresAt := time.Now().Add(5 * time.Minute)

		// 3. Simpan atau perbarui OTP di database untuk email tersebut
		// Hapus OTP lama yang belum dipakai (jika ada) untuk email ini
		DB.Unscoped().Where("email = ?", input.Email).Delete(&OTP{})
		newOTP := OTP{
			Email:     input.Email,
			Code:      otpCode,
			ExpiresAt: expiresAt,
		}

		if err := DB.Create(&newOTP).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan kode OTP", "details": err.Error()})
			return
		}
		// 4. Kirim email OTP menggunakan fungsi Resend yang sudah kita buat
		err := SendOTPEmail(input.Email, otpCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim email OTP", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Kode OTP berhasil dikirim ke " + input.Email,
		})
	})

	// --- RUTE VERIFIKASI OTP ---
	r.POST("/api/auth/verify-otp", func(c *gin.Context) {
		var input VerifyOTPInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email dan kode OTP (6 digit) wajib diisi"})
			return
		}

		// 1. Cari data OTP di database berdasarkan email dan kode
		var storedOTP OTP
		result := DB.Where("email = ? AND code = ?", input.Email, input.Code).First(&storedOTP)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Kode OTP salah atau tidak ditemukan"})
			return
		}

		// 2. Cek apakah OTP sudah kedaluwarsa
		if time.Now().After(storedOTP.ExpiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Kode OTP sudah kedaluwarsa, silakan minta ulang"})
			return
		}

		// 3. Cek apakah user sudah terdaftar di database utama, jika belum buat baru
		var user User
		dbRes := DB.Where("email = ?", input.Email).First(&user)
		if dbRes.Error != nil {
			// Daftarkan sebagai user baru otomatis jika belum ada
			user = User{
				Email: input.Email,
				Name:  "Mahasiswa Baru", // Default name, nanti bisa diubah di halaman profile
				Role:  "Student",
			}
			DB.Create(&user)
		}

		// 4. Hapus OTP yang sudah sukses digunakan agar tidak bisa dipakai 2x
		DB.Delete(&storedOTP)

		// 5. Terbitkan Token JWT (Access Token & Refresh Token) resmi untuk user!
		accessToken, refreshToken, err := GenerateTokens(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menerbitkan token autentikasi"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Verifikasi OTP Sukses!",
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user":          user,
		})
	})

	// --- RUTE UJI COBA KIRIM EMAIL ---
	r.GET("/api/test-email", func(c *gin.Context) {
		// Ganti dengan email aktifmu untuk pengujian
		targetEmail := "alinasution2401@gmail.com"
		dummyOTP := "889922"

		err := SendOTPEmail(targetEmail, dummyOTP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal mengirim email",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Email OTP berhasil dikirim ke " + targetEmail,
		})
	})

	// --- RUTE DEBUG (HANYA UNTUK CEK DATABASE SEMENTARA) ---
	r.GET("/api/debug/users", func(c *gin.Context) {
		var users []User
		DB.Find(&users)
		c.JSON(http.StatusOK, gin.H{
			"total_user": len(users),
			"data_users": users,
		})
	})

	log.Println("Server is running on port 8080...")
	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
