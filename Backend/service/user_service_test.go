package service

import (
	"campusconnect/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. MOCK REPOSITORY
// Membuat tiruan dari UserRepository agar kita tidak perlu terkoneksi ke database asli
type MockUserRepository struct {
	mock.Mock
}

// Meniru fungsi FindByEmail (Tambahan wajib agar interface terpenuhi)
func (m *MockUserRepository) FindByEmail(email string) (*repository.User, error) {
	args := m.Called(email)
	if args.Get(0) != nil {
		return args.Get(0).(*repository.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// Meniru fungsi FindByID
func (m *MockUserRepository) FindByID(userID string) (*repository.User, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).(*repository.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// Meniru fungsi Create (Tambahan wajib agar interface terpenuhi)
func (m *MockUserRepository) Create(user *repository.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// Meniru fungsi Update
func (m *MockUserRepository) Update(user *repository.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// ==============================================================================

// 2. SKENARIO TEST: Mengambil Profil Berhasil (Success)
func TestGetUserProfile_Success(t *testing.T) {
	// A. Persiapan Mock & Service
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// B. Data Tiruan
	mockUser := &repository.User{
		ID:    "user-123",
		Name:  "Lian Sang Developer",
		Email: "alinasution2401@gmail.com",
	}

	// C. Aturan main: Jika FindByID dipanggil dengan "user-123", kembalikan mockUser tanpa error
	mockRepo.On("FindByID", "user-123").Return(mockUser, nil)

	// D. Eksekusi fungsi asli
	result, err := userService.GetUserProfile("user-123")

	// E. Validasi (Harus sukses, data tidak boleh kosong, nama harus cocok)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Lian Sang Developer", result.Name)

	// Pastikan mock benar-benar dipanggil
	mockRepo.AssertExpectations(t)
}

// 3. SKENARIO TEST: Mengambil Profil Gagal (Not Found)
func TestGetUserProfile_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Aturan main: Jika dipanggil dengan ID "user-999", kembalikan error
	mockRepo.On("FindByID", "user-999").Return(nil, errors.New("user tidak ditemukan"))

	result, err := userService.GetUserProfile("user-999")

	// Validasi (Harus error, hasil harus kosong)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user tidak ditemukan", err.Error())

	mockRepo.AssertExpectations(t)
}

// 4. SKENARIO TEST: Update Profil Berhasil
func TestUpdateProfile_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	userToUpdate := &repository.User{
		ID:   "user-123",
		Name: "Lian Update",
		Bio:  "Mahasiswi D3 yang sedang membangun mahakarya",
	}

	// Aturan main: Jika fungsi Update dipanggil membawa userToUpdate, kembalikan tanpa error
	mockRepo.On("Update", userToUpdate).Return(nil)

	err := userService.UpdateProfile(userToUpdate)

	// Validasi
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
