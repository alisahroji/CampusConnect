package main

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateTokens membuat pasangan Access Token dan Refresh Token
func GenerateTokens(userID string) (string, string, error) {
	// 1. Buat Access Token (Umur Pendek: 15 Menit)
	accessTokenClaims := jwt.MapClaims{
		"sub": userID,                                     // Subject (Siapa pemilik token ini)
		"exp": time.Now().Add(time.Minute * 15).Unix(),    // Waktu kadaluarsa
		"iat": time.Now().Unix(),                          // Waktu diterbitkan
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}

	// 2. Buat Refresh Token (Umur Panjang: 7 Hari)
	refreshTokenClaims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat": time.Now().Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}