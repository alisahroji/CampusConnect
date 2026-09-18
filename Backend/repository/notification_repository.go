package repository

import (
	"time"

	"gorm.io/gorm"
)

// Notification menyimpan satu kejadian yang perlu diketahui recipient
// (like/comment/follow). Storage-only untuk Minggu 6 — realtime menyusul
// di Minggu 10. Schema tabel dibuat oleh migration 000009.
type Notification struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecipientID string     `gorm:"type:uuid;not null" json:"recipient_id"`
	ActorID     *string    `gorm:"type:uuid" json:"actor_id"` // boleh NULL (akun actor dihapus)
	Type        string     `gorm:"type:varchar(50);not null" json:"type"`
	EntityID    *string    `gorm:"type:uuid" json:"entity_id"` // project_id / post_id (NULL untuk follow)
	ReadAt      *time.Time `json:"read_at"`                    // NULL = belum dibaca
	CreatedAt   time.Time  `json:"created_at"`

	// Relasi ke users untuk menampilkan "siapa yang melakukan aksi".
	// Actor bisa NULL, sehingga dipreload manual (bukan foreignKey tag)
	// agar notification tanpa actor tetap ter-render.
	Actor *User `gorm:"-" json:"actor,omitempty"`
}

// Tipe notifikasi yang dikenal sistem (whitelist untuk query filter).
const (
	NotifTypeLikeProject    = "like_project"
	NotifTypeCommentProject = "comment_project"
	NotifTypeLikePost       = "like_post"
	NotifTypeCommentPost    = "comment_post"
	NotifTypeFollow         = "follow"
)

// NotificationRepository menyimpan & membaca notifikasi milik recipient.
type NotificationRepository interface {
	// Create menyimpan satu notification (best-effort oleh pemanggil:
	// kegagalan notifikasi tidak boleh menggagalkan aksi utamanya).
	Create(notification *Notification) error
	// ListByUser mengembalikan notifikasi milik satu user, terbaru dulu,
	// lengkap dengan data actor. Query dibatasi limit untuk keamanan.
	ListByUser(recipientID string, limit int) ([]Notification, error)
	// CountUnread menghitung notifikasi belum dibaca (badge).
	CountUnread(recipientID string) (int64, error)
	// MarkRead menandai satu notification milik recipient sebagai terbaca.
	// Mengembalikan ErrNotFound bila row tidak ada / bukan milik recipient.
	MarkRead(id, recipientID string) error
	// MarkAllRead menandai seluruh notifikasi milik recipient sebagai terbaca.
	MarkAllRead(recipientID string) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *Notification) error {
	return r.db.Create(notification).Error
}

// listNotifications menjalankan query list + preload actor dalam 2 query.
func (r *notificationRepository) listNotifications(recipientID string, limit int) ([]Notification, error) {
	var notifications []Notification
	err := r.db.
		Where("recipient_id = ?", recipientID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	// Preload actor (2 query: kumpulkan actor_id, ambil users terkait).
	actorIDs := make([]string, 0, len(notifications))
	actorIndex := make(map[string][]int, len(notifications))
	for i, n := range notifications {
		if n.ActorID != nil {
			if _, seen := actorIndex[*n.ActorID]; !seen {
				actorIDs = append(actorIDs, *n.ActorID)
			}
			actorIndex[*n.ActorID] = append(actorIndex[*n.ActorID], i)
		}
	}
	if len(actorIDs) > 0 {
		var actors []User
		if err := r.db.Where("id IN ?", actorIDs).Find(&actors).Error; err != nil {
			return nil, err
		}
		for _, actor := range actors {
			for _, idx := range actorIndex[actor.ID] {
				a := actor
				// Email tidak diekspos lewat notification (privasi, konsisten
				// dengan profil publik yang menyembunyikan email).
				a.Email = ""
				notifications[idx].Actor = &a
			}
		}
	}

	return notifications, nil
}

func (r *notificationRepository) ListByUser(recipientID string, limit int) ([]Notification, error) {
	return r.listNotifications(recipientID, limit)
}

func (r *notificationRepository) CountUnread(recipientID string) (int64, error) {
	var count int64
	err := r.db.Model(&Notification{}).
		Where("recipient_id = ? AND read_at IS NULL", recipientID).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) MarkRead(id, recipientID string) error {
	result := r.db.Model(&Notification{}).
		Where("id = ? AND recipient_id = ? AND read_at IS NULL", id, recipientID).
		Update("read_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Bisa berarti: tidak ada, bukan milik recipient, atau sudah dibaca.
		// Untuk idempotensi "sudah dibaca" kita cek keberadaan row miliknya.
		var exists int64
		r.db.Model(&Notification{}).
			Where("id = ? AND recipient_id = ?", id, recipientID).
			Count(&exists)
		if exists == 0 {
			return ErrNotFound
		}
	}
	return nil
}

func (r *notificationRepository) MarkAllRead(recipientID string) error {
	return r.db.Model(&Notification{}).
		Where("recipient_id = ? AND read_at IS NULL", recipientID).
		Update("read_at", time.Now()).Error
}
