package repository

import (
	"time"
)

// Event adalah satu kegiatan akademik/kampus (Minggu 7 Fase 3). Schema tabel
// dibuat oleh migration 000013. Pada Day 4 ini file HANYA berisi model untuk
// kebutuhan schema & compile — repository CRUD, service, handler, dan route
// menyusul di Day 5 sesuai roadmap.
//
// File reference: location/event_url nullable keduanya (roadmap "lokasi/link"
// belum menentukan wajib-salah-satu); starts_at TIMESTAMPTZ mengikuti konvensi
// seluruh tabel existing (UTC di DB, render di aplikasi).
type Event struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatorID   string    `gorm:"type:uuid;not null" json:"creator_id"`
	Title       string    `gorm:"type:varchar(150);not null" json:"title"`
	Description *string   `gorm:"type:text" json:"description"`
	StartsAt    time.Time `gorm:"not null" json:"starts_at"`
	Location    *string   `gorm:"type:varchar(255)" json:"location"`
	EventURL    *string   `gorm:"type:varchar(255)" json:"event_url"`
	CreatedAt   time.Time `json:"created_at"`

	// Relasi ke tabel User (satu event dibuat satu user) — pola Material.Uploader.
	// Dipreload manual saat query (Day 5); DTO yang menentukan field user tampil.
	Creator *User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

// EventRSVP adalah keikutsertaan satu user pada satu event. Duplikat
// dicegah DB lewat uq_event_rsvps_pair (migration 000013) — pola
// project_bookmarks/post_bookmarks (000008).
type EventRSVP struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EventID   string    `gorm:"type:uuid;not null" json:"event_id"`
	UserID    string    `gorm:"type:uuid;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`

	// Relasi ke tabel Event (satu RSVP milik satu event).
	Event *Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
	// Relasi ke tabel User (satu RSVP milik satu user).
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
