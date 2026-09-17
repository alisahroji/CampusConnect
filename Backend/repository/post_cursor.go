package repository

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// EncodePostCursor membungkus (created_at, id) post terakhir pada sebuah halaman
// menjadi cursor opaque ber-base64. Cursor ini dipakai client untuk meminta
// halaman berikutnya (keyset pagination, lebih scalable daripada offset).
func EncodePostCursor(createdAt time.Time, id string) string {
	raw := fmt.Sprintf("%s|%s", createdAt.UTC().Format(time.RFC3339Nano), id)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

// DecodePostCursor membuka cursor hasil EncodePostCursor menjadi
// timestamp dan ID untuk query halaman berikutnya.
func DecodePostCursor(cursor string) (time.Time, string, error) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", ErrInvalidCursor
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", ErrInvalidCursor
	}

	cursorTime, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", ErrInvalidCursor
	}

	return cursorTime, parts[1], nil
}
