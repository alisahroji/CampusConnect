package service

import (
	"campusconnect/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// MOCK REPOSITORY NOTIFICATION (Minggu 6 Hari 2)
// =============================================================================

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(notification *repository.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) ListByUser(recipientID string, limit int) ([]repository.Notification, error) {
	args := m.Called(recipientID, limit)
	notifications, _ := args.Get(0).([]repository.Notification)
	return notifications, args.Error(1)
}

func (m *MockNotificationRepository) CountUnread(recipientID string) (int64, error) {
	args := m.Called(recipientID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNotificationRepository) MarkRead(id, recipientID string) error {
	args := m.Called(id, recipientID)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAllRead(recipientID string) error {
	args := m.Called(recipientID)
	return args.Error(0)
}

func newNotificationTestService() (NotificationService, *MockNotificationRepository) {
	repo := new(MockNotificationRepository)
	return NewNotificationService(repo), repo
}

// Notify tidak boleh membuat notification ketika recipient == actor
// (jangan notifikasi diri sendiri — like/comment/follow konten sendiri).
func TestNotify_SkipSelfNotify(t *testing.T) {
	svc, repo := newNotificationTestService()

	svc.Notify("user-A", "user-A", repository.NotifTypeLikePost, "post-1")

	repo.AssertNotCalled(t, "Create", mock.Anything)
}

// Notify mengabaikan input kosong (recipient/actor wajib ada).
func TestNotify_SkipEmptyFields(t *testing.T) {
	svc, repo := newNotificationTestService()

	svc.Notify("", "user-B", repository.NotifTypeFollow, "")
	svc.Notify("user-A", "", repository.NotifTypeFollow, "")

	repo.AssertNotCalled(t, "Create", mock.Anything)
}

// Notify menyimpan notification dengan recipient/actor/type/entity benar.
func TestNotify_CreatesNotification(t *testing.T) {
	svc, repo := newNotificationTestService()

	repo.On("Create", mock.MatchedBy(func(n *repository.Notification) bool {
		return n.RecipientID == "user-A" &&
			n.ActorID != nil && *n.ActorID == "user-B" &&
			n.Type == repository.NotifTypeLikePost &&
			n.EntityID != nil && *n.EntityID == "post-1"
	})).Return(nil)

	svc.Notify("user-A", "user-B", repository.NotifTypeLikePost, "post-1")

	repo.AssertExpectations(t)
}

// Notify TANPA entity (mis. follow) tetap valid — EntityID dibiarkan nil.
func TestNotify_FollowWithoutEntity(t *testing.T) {
	svc, repo := newNotificationTestService()

	repo.On("Create", mock.MatchedBy(func(n *repository.Notification) bool {
		return n.Type == repository.NotifTypeFollow && n.EntityID == nil
	})).Return(nil)

	svc.Notify("user-A", "user-B", repository.NotifTypeFollow, "")

	repo.AssertExpectations(t)
}

// Kegagalan menyimpan notification tidak boleh menggagalkan aksi utama:
// Notify harus tetap "sukses" (tidak panic, tidak mengembalikan error ke
// pemanggil like/comment/follow). Best-effort.
func TestNotify_DBFailureIsBestEffort(t *testing.T) {
	svc, repo := newNotificationTestService()

	repo.On("Create", mock.Anything).Return(errors.New("db down"))

	assert.NotPanics(t, func() {
		svc.Notify("user-A", "user-B", repository.NotifTypeCommentPost, "post-1")
	})
	repo.AssertExpectations(t)
}

// List membatasi limit ke 50 bila caller mengirim nilai tidak wajar.
func TestList_ClampsLimit(t *testing.T) {
	svc, repo := newNotificationTestService()

	repo.On("ListByUser", "user-A", 50).Return([]repository.Notification{}, nil).Twice()

	_, err := svc.List("user-A", 0)    // tidak valid -> clamp ke 50
	assert.NoError(t, err)
	_, err = svc.List("user-A", 99999) // terlalu besar -> clamp ke 50
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

// UnreadCount meneruskan hasil repository (dasar badge polling).
func TestUnreadCount(t *testing.T) {
	svc, repo := newNotificationTestService()

	repo.On("CountUnread", "user-A").Return(int64(3), nil)

	count, err := svc.UnreadCount("user-A")

	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

// MarkRead meneruskan error sentinel repository (ErrNotFound untuk
// notification yang bukan milik recipient).
func TestMarkRead_NotOwnerPropagatesNotFound(t *testing.T) {
	svc, repo := newNotificationTestService()

	repo.On("MarkRead", "notif-1", "user-B").Return(repository.ErrNotFound)

	err := svc.MarkRead("notif-1", "user-B")

	assert.ErrorIs(t, err, repository.ErrNotFound)
}

// MarkAllRead sukses tanpa error.
func TestMarkAllRead(t *testing.T) {
	svc, repo := newNotificationTestService()

	repo.On("MarkAllRead", "user-A").Return(nil)

	assert.NoError(t, svc.MarkAllRead("user-A"))
	repo.AssertExpectations(t)
}
