package service

import (
	"campusconnect/repository"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// TEST MATERIAL SERVICE (Minggu 7 Day 2)
// Mock repository mengikuti pola user_service_test.go.
// =============================================================================

type MockMaterialRepository struct {
	mock.Mock
}

func (m *MockMaterialRepository) Create(material *repository.Material) error {
	args := m.Called(material)
	return args.Error(0)
}

func (m *MockMaterialRepository) FindByID(id string) (*repository.Material, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*repository.Material), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMaterialRepository) List(category string, uploaderID string, limit int) ([]repository.Material, error) {
	args := m.Called(category, uploaderID, limit)
	return args.Get(0).([]repository.Material), args.Error(1)
}

func (m *MockMaterialRepository) Update(material *repository.Material) error {
	args := m.Called(material)
	return args.Error(0)
}

func (m *MockMaterialRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockFileUploader struct {
	mock.Mock
}

func (m *MockFileUploader) Upload(filename string, size int64, content io.Reader) (string, error) {
	args := m.Called(filename, size)
	return args.String(0), args.Error(1)
}

const (
	lecturerID = "11111111-1111-1111-1111-111111111111"
	studentID  = "22222222-2222-2222-2222-222222222222"
)

func newMaterialService() (MaterialService, *MockMaterialRepository, *MockFileUploader) {
	repo := new(MockMaterialRepository)
	up := new(MockFileUploader)
	svc := NewMaterialService(repo, up)
	return svc, repo, up
}

// ---------------------------------------------------------------------------
// AUTHORIZATION: upload khusus Lecturer
// ---------------------------------------------------------------------------

func TestCreateMaterial_StudentForbidden(t *testing.T) {
	svc, repo, up := newMaterialService()
	// Tidak ada ekspektasi repo/uploader: harus ditolak sebelum keduanya dipakai.
	_, err := svc.CreateMaterial(studentID, "Student", "Judul", "Kategori", "materi.pdf", 100, strings.NewReader("isi"))
	assert.ErrorIs(t, err, repository.ErrForbidden)
	repo.AssertNotCalled(t, "Create")
	up.AssertNotCalled(t, "Upload")
}

func TestCreateMaterial_AdminForbidden(t *testing.T) {
	svc, _, _ := newMaterialService()
	// Policy Day 2: Admin TIDAK otomatis boleh upload (roadmap hanya menyebut Lecturer).
	_, err := svc.CreateMaterial(lecturerID, "Admin", "Judul", "Kategori", "materi.pdf", 100, strings.NewReader("isi"))
	assert.ErrorIs(t, err, repository.ErrForbidden)
}

func TestCreateMaterial_LecturerAllowed(t *testing.T) {
	svc, repo, up := newMaterialService()
	up.On("Upload", "materi.pdf", int64(100)).Return("https://res.cloudinary.com/raw/materi.pdf", nil)
	repo.On("Create", mock.AnythingOfType("*repository.Material")).Return(nil)

	material, err := svc.CreateMaterial(lecturerID, "Lecturer", "  Basis Data  ", " IF4021 ", "materi.pdf", 100, strings.NewReader("isi"))
	assert.NoError(t, err)
	assert.Equal(t, "Basis Data", material.Title) // trim
	assert.Equal(t, "IF4021", material.Category)  // trim
	assert.Equal(t, lecturerID, material.UploaderID)
	assert.Equal(t, "https://res.cloudinary.com/raw/materi.pdf", material.FileURL)
}

// ---------------------------------------------------------------------------
// VALIDASI INPUT & FILE
// ---------------------------------------------------------------------------

func TestCreateMaterial_EmptyTitle(t *testing.T) {
	svc, repo, _ := newMaterialService()
	_, err := svc.CreateMaterial(lecturerID, "Lecturer", "   ", "Kategori", "materi.pdf", 100, strings.NewReader("isi"))
	assert.Error(t, err)
	repo.AssertNotCalled(t, "Create")
}

func TestCreateMaterial_EmptyCategory(t *testing.T) {
	svc, _, _ := newMaterialService()
	_, err := svc.CreateMaterial(lecturerID, "Lecturer", "Judul", "", "materi.pdf", 100, strings.NewReader("isi"))
	assert.Error(t, err)
}

func TestCreateMaterial_TitleTooLong(t *testing.T) {
	svc, _, _ := newMaterialService()
	long := strings.Repeat("a", 151)
	_, err := svc.CreateMaterial(lecturerID, "Lecturer", long, "Kategori", "materi.pdf", 100, strings.NewReader("isi"))
	assert.Error(t, err)
}

func TestCreateMaterial_MissingFile(t *testing.T) {
	svc, _, _ := newMaterialService()
	_, err := svc.CreateMaterial(lecturerID, "Lecturer", "Judul", "Kategori", "", 100, strings.NewReader("isi"))
	assert.Error(t, err)
}

func TestCreateMaterial_InvalidExtension(t *testing.T) {
	svc, _, _ := newMaterialService()
	_, err := svc.CreateMaterial(lecturerID, "Lecturer", "Judul", "Kategori", "virus.exe", 100, strings.NewReader("isi"))
	assert.Error(t, err) // executable ditolak
}

func TestCreateMaterial_EmptyFile(t *testing.T) {
	svc, _, _ := newMaterialService()
	_, err := svc.CreateMaterial(lecturerID, "Lecturer", "Judul", "Kategori", "materi.pdf", 0, strings.NewReader(""))
	assert.Error(t, err)
}

func TestCreateMaterial_OversizeFile(t *testing.T) {
	svc, _, _ := newMaterialService()
	_, err := svc.CreateMaterial(lecturerID, "Lecturer", "Judul", "Kategori", "materi.pdf", MaxMaterialSize+1, strings.NewReader("isi"))
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// STORAGE ORKESTRASI
// ---------------------------------------------------------------------------

func TestCreateMaterial_UploadFailure(t *testing.T) {
	svc, repo, up := newMaterialService()
	up.On("Upload", "materi.pdf", int64(100)).Return("", errors.New("cloud down"))
	_, err := svc.CreateMaterial(lecturerID, "Lecturer", "Judul", "Kategori", "materi.pdf", 100, strings.NewReader("isi"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "penyimpanan cloud") // pesan aman, tanpa detail provider
	repo.AssertNotCalled(t, "Create")                    // DB tidak tertulis saat upload gagal
}

func TestCreateMaterial_DBFailureAfterUpload(t *testing.T) {
	svc, repo, up := newMaterialService()
	up.On("Upload", mock.Anything, mock.Anything).Return("https://res.cloudinary.com/raw/materi.pdf", nil)
	repo.On("Create", mock.AnythingOfType("*repository.Material")).Return(errors.New("db down"))
	_, err := svc.CreateMaterial(lecturerID, "Lecturer", "Judul", "Kategori", "materi.pdf", 100, strings.NewReader("isi"))
	assert.Error(t, err)
	// Orphan file di storage adalah tradeoff yang didokumentasikan (bukan bug tersembunyi).
}

// ---------------------------------------------------------------------------
// READ
// ---------------------------------------------------------------------------

func TestGetMaterial_NotFound(t *testing.T) {
	svc, repo, _ := newMaterialService()
	repo.On("FindByID", "tidak-ada").Return(nil, repository.ErrNotFound)
	_, err := svc.GetMaterial("tidak-ada")
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestListMaterials_TrimsCategory(t *testing.T) {
	svc, repo, _ := newMaterialService()
	// Clamp default/max limit terjadi di repository.List; service meneruskan
	// limit apa adanya. Yang diuji di service: trim kategori.
	repo.On("List", "IF4021", "", mock.Anything).Return([]repository.Material{}, nil)
	list, err := svc.ListMaterials("  IF4021  ", "", 0)
	assert.NoError(t, err)
	assert.Empty(t, list)
}

// ---------------------------------------------------------------------------
// OWNERSHIP UPDATE
// ---------------------------------------------------------------------------

func TestUpdateMaterial_OwnerSuccess(t *testing.T) {
	svc, repo, _ := newMaterialService()
	existing := &repository.Material{ID: "m1", UploaderID: lecturerID, Title: "Lama", Category: "Lama"}
	repo.On("FindByID", "m1").Return(existing, nil)
	repo.On("Update", mock.AnythingOfType("*repository.Material")).Return(nil)

	updated, err := svc.UpdateMaterial("m1", lecturerID, "Baru", "Baru")
	assert.NoError(t, err)
	assert.Equal(t, "Baru", updated.Title)
}

func TestUpdateMaterial_NotOwnerForbidden(t *testing.T) {
	svc, repo, _ := newMaterialService()
	existing := &repository.Material{ID: "m1", UploaderID: lecturerID, Title: "Lama", Category: "Lama"}
	repo.On("FindByID", "m1").Return(existing, nil)

	_, err := svc.UpdateMaterial("m1", studentID, "Baru", "Baru")
	assert.ErrorIs(t, err, repository.ErrForbidden)
	repo.AssertNotCalled(t, "Update") // tidak ada penulisan
}

func TestUpdateMaterial_InvalidInput(t *testing.T) {
	svc, repo, _ := newMaterialService()
	_, err := svc.UpdateMaterial("m1", lecturerID, "", "Kategori")
	assert.Error(t, err)
	repo.AssertNotCalled(t, "FindByID")
}

func TestUpdateMaterial_NotFound(t *testing.T) {
	svc, repo, _ := newMaterialService()
	repo.On("FindByID", "m404").Return(nil, repository.ErrNotFound)
	_, err := svc.UpdateMaterial("m404", lecturerID, "Judul", "Kategori")
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

// ---------------------------------------------------------------------------
// OWNERSHIP DELETE
// ---------------------------------------------------------------------------

func TestDeleteMaterial_OwnerSuccess(t *testing.T) {
	svc, repo, _ := newMaterialService()
	existing := &repository.Material{ID: "m1", UploaderID: lecturerID}
	repo.On("FindByID", "m1").Return(existing, nil)
	repo.On("Delete", "m1").Return(nil)

	err := svc.DeleteMaterial("m1", lecturerID)
	assert.NoError(t, err)
}

func TestDeleteMaterial_NotOwnerForbidden(t *testing.T) {
	svc, repo, _ := newMaterialService()
	existing := &repository.Material{ID: "m1", UploaderID: lecturerID}
	repo.On("FindByID", "m1").Return(existing, nil)

	err := svc.DeleteMaterial("m1", studentID)
	assert.ErrorIs(t, err, repository.ErrForbidden)
	repo.AssertNotCalled(t, "Delete")
}

func TestDeleteMaterial_NotFound(t *testing.T) {
	svc, repo, _ := newMaterialService()
	repo.On("FindByID", "m404").Return(nil, repository.ErrNotFound)
	err := svc.DeleteMaterial("m404", lecturerID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
