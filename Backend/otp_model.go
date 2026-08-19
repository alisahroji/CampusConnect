package main

import (
	"time"
)

type OTP struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email     string    `gorm:"not null;index"`
	Code      string    `gorm:"type:varchar(6);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
