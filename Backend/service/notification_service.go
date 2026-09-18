package service

import (
	"campusconnect/repository"
	"log"
)

// Notifier adalah interface minimal yang dipakai layer interaksi
// (like/comment/follow) untuk membuat notification secara best-effort.
// Dipisah dari NotificationService agar mock test interaksi existing
// tidak perlu diubah (pola yang sama dengan SetFollowService).
type Notifier interface {
	// Notify membuat satu notification. Implementasi WAJIB tidak pernah
	// panic dan mengembalikan error hanya untuk logging pemanggil:
	// kegagalan notifikasi tidak boleh menggagalkan aksi utama user.
	Notify(recipientID, actorID, notifType, entityID string)
}

type NotificationService interface {
	Notifier
	// List mengembalikan notification milik recipient (terbaru dulu).
	List(recipientID string, limit int) ([]repository.Notification, error)
	// UnreadCount menghitung notifikasi belum dibaca (badge polling).
	UnreadCount(recipientID string) (int64, error)
	// MarkRead menandai satu notification milik recipient terbaca.
	MarkRead(id, recipientID string) error
	// MarkAllRead menandai semua notification milik recipient terbaca.
	MarkAllRead(recipientID string) error
}

type notificationService struct {
	notifRepo repository.NotificationRepository
}

func NewNotificationService(notifRepo repository.NotificationRepository) NotificationService {
	return &notificationService{notifRepo: notifRepo}
}

// Notify membuat notification dengan aturan:
//   - recipient == actor -> tidak ada notifikasi (jangan notifikasi diri sendiri)
//   - field kosong aman: actor/entity boleh kosong (kolom DB nullable)
//   - kegagalan DB hanya di-log; aksi utama user tetap sukses
func (s *notificationService) Notify(recipientID, actorID, notifType, entityID string) {
	if recipientID == "" || actorID == "" {
		return
	}
	if recipientID == actorID {
		return // tidak ada notifikasi untuk aksi terhadap diri sendiri
	}

	notification := &repository.Notification{
		RecipientID: recipientID,
		ActorID:     &actorID,
		Type:        notifType,
	}
	if entityID != "" {
		notification.EntityID = &entityID
	}

	if err := s.notifRepo.Create(notification); err != nil {
		// Best-effort: notification gagal tidak boleh merusak aksi utama.
		log.Printf("[notification] gagal menyimpan notification type=%s: %v", notifType, err)
	}
}

func (s *notificationService) List(recipientID string, limit int) ([]repository.Notification, error) {
	// Batas wajar untuk list dasar Minggu 6 (belum pagination cursor).
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.notifRepo.ListByUser(recipientID, limit)
}

func (s *notificationService) UnreadCount(recipientID string) (int64, error) {
	return s.notifRepo.CountUnread(recipientID)
}

func (s *notificationService) MarkRead(id, recipientID string) error {
	return s.notifRepo.MarkRead(id, recipientID)
}

func (s *notificationService) MarkAllRead(recipientID string) error {
	return s.notifRepo.MarkAllRead(recipientID)
}
