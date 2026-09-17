package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log" // Ini log bawaan Go (dibiarkan agar kode lamamu tetap aman)
	"net/http"
	"net/url"
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

	// Kita beri nama samaran "zlog" agar tidak bentrok dengan "log" bawaan Go
	zlog "github.com/rs/zerolog/log"
)

var googleOauthConfig *oauth2.Config

const (
	// URL dasar frontend SPA (Vite) — dipakai untuk redirect pasca-OAuth
	frontendURL = "http://localhost:5173"
	// Tujuan redirect jika proses OAuth gagal
	oauthErrorRedirect = frontendURL + "/login?error=oauth_failed"
)

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

	// Inisiasi Layer User
	userRepo := repository.NewUserRepository(DB)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Inisiasi Layer Project
	projectRepo := repository.NewProjectRepository(DB)
	projectService := service.NewProjectService(projectRepo)
	projectHandler := handler.NewProjectHandler(projectService)

	// Inisiasi Layer Interaksi (Like & Comment) - Minggu 4 Hari 3
	commentRepo := repository.NewCommentRepository(DB)
	commentService := service.NewCommentService(commentRepo, projectRepo)
	commentHandler := handler.NewCommentHandler(commentService)

	likeRepo := repository.NewLikeRepository(DB)
	likeService := service.NewLikeService(likeRepo, projectRepo)
	likeHandler := handler.NewLikeHandler(likeService)	// Inisiasi Layer Post (Feed) - Minggu 5 Hari 1 & 2
	postRepo := repository.NewPostRepository(DB)
	postService := service.NewPostService(postRepo)
	postHandler := handler.NewPostHandler(postService)

	// Inisiasi Layer Interaksi Post (Like & Comment) - Minggu 5 Hari 3
	postLikeRepo := repository.NewPostLikeRepository(DB)
	postCommentRepo := repository.NewPostCommentRepository(DB)
	postLikeService := service.NewPostLikeService(postLikeRepo, postRepo)
	postCommentService := service.NewPostCommentService(postCommentRepo, postRepo)
	postInteractionHandler := handler.NewPostInteractionHandler(postLikeService, postCommentService)

	// Inisiasi Layer Follow & Feed (Social Graph) - Minggu 5 Hari 3
	followRepo := repository.NewFollowRepository(DB)
	followService := service.NewFollowService(followRepo, userRepo)
	followHandler := handler.NewFollowHandler(followService)
	feedService := service.NewFeedService(followRepo, postRepo)
	feedHandler := handler.NewFeedHandler(feedService)

	// Inisiasi Layer Search - Minggu 5 Hari 4
	searchRepo := repository.NewSearchRepository(DB)
	searchService := service.NewSearchService(searchRepo)
	searchHandler := handler.NewSearchHandler(searchService)



	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Izin khusus untuk frontend Vite
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// --- ENDPOINT HEALTH CHECK & READINESS ---
	r.GET("/health", handler.HealthCheck)
	r.GET("/ready", handler.ReadyCheck)

	// --- RUTE AUTH & LOGIN ---
	r.GET("/api/auth/google/login", func(c *gin.Context) {
		oauthStateString := "random-state-string"
		url := googleOauthConfig.AuthCodeURL(oauthStateString)
		c.Redirect(http.StatusTemporaryRedirect, url)
	})

	r.GET("/api/auth/google/callback", func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.Redirect(http.StatusFound, oauthErrorRedirect)
			return
		}

		token, err := googleOauthConfig.Exchange(context.Background(), code)
		if err != nil {
			c.Redirect(http.StatusFound, oauthErrorRedirect)
			return
		}

		response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
		if err != nil {
			c.Redirect(http.StatusFound, oauthErrorRedirect)
			return
		}
		defer response.Body.Close()

		userData, err := io.ReadAll(response.Body)
		if err != nil {
			c.Redirect(http.StatusFound, oauthErrorRedirect)
			return
		}

		var googleUser GoogleUserResult
		if err := json.Unmarshal(userData, &googleUser); err != nil {
			c.Redirect(http.StatusFound, oauthErrorRedirect)
			return
		}

		var user User
		result := DB.Where("email = ?", googleUser.Email).First(&user)

		if result.Error != nil {
			user = User{
				Email:      googleUser.Email,
				Name:       googleUser.Name,
				PictureURL: googleUser.Picture,
			}
			if err := DB.Create(&user).Error; err != nil {
				log.Printf("Gagal menyimpan user baru: %v\n", err)
				c.Redirect(http.StatusFound, oauthErrorRedirect)
				return
			}
			log.Println("User baru berhasil didaftarkan otomatis!")
		} else {
			log.Println("User lama berhasil login kembali!")
		}

		accessToken, refreshToken, err := GenerateTokens(user.ID)
		if err != nil {
			log.Printf("Gagal membuat token: %v\n", err)
			c.Redirect(http.StatusFound, oauthErrorRedirect)
			return
		}

		// Redirect kembali ke frontend SPA sambil membawa token via query params.
		// Frontend (AuthCallback) yang akan menyimpan token ke localStorage.
		redirectURL := frontendURL + "/auth/callback?" +
			"access_token=" + url.QueryEscape(accessToken) +
			"&refresh_token=" + url.QueryEscape(refreshToken)
		c.Redirect(http.StatusFound, redirectURL)
	})

	r.POST("/api/auth/refresh", func(c *gin.Context) {
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token tidak boleh kosong"})
			return
		}

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

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal membaca data token"})
			return
		}

		userID := claims["sub"].(string)

		newAccessToken, newRefreshToken, err := GenerateTokens(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token baru"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Token berhasil diperbarui!",
			"access_token":  newAccessToken,
			"refresh_token": newRefreshToken,
		})
	})

	r.POST("/api/auth/logout", RequireAuth, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Berhasil logout! Sesi diakhiri secara aman.",
			"instruction": "Frontend wajib menghapus access_token dan refresh_token dari storage lokal.",
		})
	})

	r.POST("/api/auth/request-otp", func(c *gin.Context) {
		var input RequestOTPInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format email tidak valid"})
			return
		}

		otpCode := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
		expiresAt := time.Now().Add(5 * time.Minute)

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

		err := SendOTPEmail(input.Email, otpCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim email OTP", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Kode OTP berhasil dikirim ke " + input.Email,
		})
	})

	r.POST("/api/auth/verify-otp", func(c *gin.Context) {
		var input VerifyOTPInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email dan kode OTP (6 digit) wajib diisi"})
			return
		}

		var storedOTP OTP
		result := DB.Where("email = ? AND code = ?", input.Email, input.Code).First(&storedOTP)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Kode OTP salah atau tidak ditemukan"})
			return
		}

		if time.Now().After(storedOTP.ExpiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Kode OTP sudah kedaluwarsa, silakan minta ulang"})
			return
		}

		var user User
		dbRes := DB.Where("email = ?", input.Email).First(&user)
		if dbRes.Error != nil {
			user = User{
				Email: input.Email,
				Name:  "Mahasiswa Baru",
				Role:  "Student",
			}
			DB.Create(&user)
		}

		DB.Delete(&storedOTP)

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

	// --- RUTE PROFILE ---
	r.GET("/api/profile", RequireAuth, userHandler.GetProfile)
	r.PUT("/api/profile", RequireAuth, userHandler.UpdateProfile)
	r.POST("/api/profile/avatar", RequireAuth, userHandler.UploadAvatar)

	// --- RUTE PROJECT SHOWCASE ---
	// Endpoint publik (bisa diakses tanpa login)
	r.GET("/api/projects", projectHandler.GetAll)
	r.GET("/api/projects/:id", projectHandler.GetByID)

	// Endpoint terproteksi (wajib login menggunakan middleware RequireAuth)
	r.POST("/api/projects", RequireAuth, projectHandler.Create)
	r.PUT("/api/projects/:id", RequireAuth, projectHandler.Update)
	r.DELETE("/api/projects/:id", RequireAuth, projectHandler.Delete)

	// --- RUTE INTERAKSI: LIKE & COMMENT ---
	// Komentar: daftar publik, tambah & hapus terproteksi
	r.GET("/api/projects/:id/comments", commentHandler.GetByProject)
	r.POST("/api/projects/:id/comments", RequireAuth, commentHandler.Create)
	r.DELETE("/api/comments/:id", RequireAuth, commentHandler.Delete)

	// Like: toggle terproteksi, hitung publik
	r.POST("/api/projects/:id/like", RequireAuth, likeHandler.Toggle)
	r.GET("/api/projects/:id/likes", likeHandler.GetLikes)

	// --- RUTE FEED POSTS (Minggu 5) ---
	// Endpoint publik (daftar feed bisa diakses tanpa login)
	r.GET("/api/posts", postHandler.GetAll)
	r.GET("/api/posts/:id", postHandler.GetByID)

	// Endpoint terproteksi (wajib login menggunakan middleware RequireAuth)
	r.POST("/api/posts", RequireAuth, postHandler.Create)
	r.DELETE("/api/posts/:id", RequireAuth, postHandler.Delete)

	// --- RUTE INTERAKSI POST: LIKE & COMMENT (Minggu 5) ---
	// Komentar post: daftar publik, tambah & hapus terproteksi
	r.GET("/api/posts/:id/comments", postInteractionHandler.GetComments)
	r.POST("/api/posts/:id/comments", RequireAuth, postInteractionHandler.CreateComment)
	r.DELETE("/api/post-comments/:id", RequireAuth, postInteractionHandler.DeleteComment)

	// Like post: toggle terproteksi, hitungan publik
	r.POST("/api/posts/:id/like", RequireAuth, postInteractionHandler.ToggleLike)
	r.GET("/api/posts/:id/likes", postInteractionHandler.GetLikes)

	// --- RUTE SOCIAL GRAPH: FOLLOW & FEED (Minggu 5) ---
	r.POST("/api/users/:id/follow", RequireAuth, followHandler.Toggle)
	r.GET("/api/feed", RequireAuth, feedHandler.GetFeed)

	// --- RUTE SEARCH (Minggu 5 Hari 4) ---
	r.GET("/api/search", searchHandler.Search)


	// --- RUTE DEBUG & TEST ---
	r.GET("/api/test-email", func(c *gin.Context) {
		targetEmail := "alinasution2401@gmail.com"
		dummyOTP := "889922"

		err := SendOTPEmail(targetEmail, dummyOTP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim email", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Email OTP berhasil dikirim ke " + targetEmail})
	})

	r.GET("/api/debug/users", func(c *gin.Context) {
		var users []User
		DB.Find(&users)
		c.JSON(http.StatusOK, gin.H{
			"total_user": len(users),
			"data_users": users,
		})
	})

	// --- LOGGING CANGGIH DARI ZEROLOG ---
	zlog.Info().Msg("🚀 Server CampusConnect berjalan mantap di port 8080...")

	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
