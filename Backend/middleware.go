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

	// 6b. Minggu 6 Hari 4: tolak user yang diblokir admin (403, bukan 401 —
	// identitasnya valid, hanya akunnya dibekukan). User dimuat dari DB
	// per-request sehingga ban berlaku SEKETIKA meskipun access token lama
	// belum kedaluwarsa.
	if user.Banned {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Akun ini sedang diblokir oleh admin"})
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

// OptionalAuth mencoba membaca token JWT BILA header Authorization ada.
// Berbeda dengan RequireAuth, request tetap dilanjutkan (tanpa abort) saat
// token tidak ada / tidak valid — handler hanya melihat userID tidak di-set.
// Dipakai endpoint publik yang tetap butuh identitas viewer, mis. GET /likes
// agar bisa mengembalikan current_user.liked untuk user yang login.
func OptionalAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.Next() // anonymous guest
		return
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode enkripsi tidak valid")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		c.Next() // token rusak/expired → perlakukan sebagai guest
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.Next()
		return
	}

	if exp, ok := claims["exp"].(float64); !ok || float64(time.Now().Unix()) > exp {
		c.Next()
		return
	}

	var user User
	DB.First(&user, "id = ?", claims["sub"])
	if user.ID == "" {
		c.Next()
		return
	}

	// Minggu 6 Hari 4: user diblokir diperlakukan sebagai guest (tanpa
	// identitas), konsisten dengan RequireAuth yang menolaknya secara eksplisit.
	if user.Banned {
		c.Next()
		return
	}

	c.Set("currentUser", user)
	c.Set("userID", user.ID)
	c.Next()
}

// loadUserByID mengambil satu user dari DB berdasarkan ID.
// Helper package-level agar auth inline di main.go (login OTP, refresh,
// Google callback) bisa memakai cek ban yang sama dengan middleware.
func loadUserByID(id string) *User {
	var user User
	if err := DB.First(&user, "id = ?", id).Error; err != nil {
		return nil
	}
	return &user
}

// AdminGuard adalah lapisan otorisasi di ATAS RequireAuth (rute admin):
// RequireAuth sudah memverifikasi token & memuat user dari DB (sumber
// kebenaran) ke context. Role TIDAK dibaca dari token/client — melainkan
// dari DB via RequireAuth, sehingga role yang diubah admin langsung berlaku.
//  - anonymous / token rusak  -> 401 (oleh RequireAuth, AdminGuard tak tercapai)
//  - authenticated non-Admin  -> 403
//  - authenticated Admin      -> lanjut ke handler
func AdminGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, exists := c.Get("currentUser")
		if !exists {
			// Pertahanan: tidak boleh terjadi setelah RequireAuth.
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Belum terautentikasi"})
			c.Abort()
			return
		}

		user, ok := currentUser.(User)
		if !ok || user.Role != "Admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Endpoint ini khusus admin"})
			c.Abort()
			return
		}

		// Defense in depth (Minggu 6 Hari 4): walaupun RequireAuth sudah
		// menolak user banned sebelum guard ini, guard tetap mengecek agar
		// salah wiring rute di masa depan tidak membuka celah.
		if user.Banned {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Akun ini sedang diblokir oleh admin"})
			c.Abort()
			return
		}

		c.Next()
	}
}
