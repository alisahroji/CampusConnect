package repository

import "errors"

// Sentinel errors standar agar handler bisa memetakan error ke status HTTP
// menggunakan errors.Is(...) tanpa perlu membandingkan string error secara manual.
var (
	// ErrNotFound dipakai ketika record tidak ditemukan di database.
	ErrNotFound = errors.New("data tidak ditemukan")

	// ErrForbidden dipakai ketika user tidak berhak melakukan aksi (bukan pemilik resource).
	ErrForbidden = errors.New("akses ditolak")

	// ErrInvalidCursor dipakai ketika cursor pagination tidak dapat dibaca.
	ErrInvalidCursor = errors.New("cursor tidak valid")
)
