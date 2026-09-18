package service

import (
	"campusconnect/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// TEST ADMIN SERVICE (Minggu 6 Hari 4)
// MockUserRepository (dengan method List/UpdateBanned/UpdateRole/CountByRole)
// sudah tersedia di user_service_test.go — package yang sama.
// =============================================================================

const (
	testActorID  = "11111111-1111-1111-1111-111111111111" // admin pelaku aksi
	testTargetID = "22222222-2222-2222-2222-222222222222" // user target
)

func newTestAdminService(repo *MockUserRepository) AdminService {
	return NewAdminService(repo)
}

// -----------------------------------------------------------------------------
// LIST USERS
// -----------------------------------------------------------------------------

func TestListUsers_ReturnsUsersFromRepo(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	expected := []repository.User{
		{ID: testTargetID, Name: "Ali", Email: "ali@kampus.ac.id", Role: "Student"},
		{ID: testActorID, Name: "Bahlil", Email: "bahlil@kampus.ac.id", Role: "Admin"},
	}
	repo.On("List", 100).Return(expected, nil)

	users, err := svc.ListUsers(100)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "Ali", users[0].Name)
	repo.AssertExpectations(t)
}

func TestListUsers_PassesLimitThrough(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("List", 5).Return([]repository.User{}, nil)

	users, err := svc.ListUsers(5)

	assert.NoError(t, err)
	assert.Empty(t, users)
	repo.AssertExpectations(t)
}

// -----------------------------------------------------------------------------
// BAN / UNBAN
// -----------------------------------------------------------------------------

func TestSetBanned_SelfBanRejected(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	err := svc.SetBanned(testActorID, testActorID, true)

	assert.ErrorIs(t, err, ErrSelfBan)
	repo.AssertNotCalled(t, "UpdateBanned")
}

func TestSetBanned_TargetNotFound(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("FindByID", testTargetID).Return(nil, repository.ErrNotFound)

	err := svc.SetBanned(testActorID, testTargetID, true)

	assert.ErrorIs(t, err, repository.ErrNotFound)
	repo.AssertNotCalled(t, "UpdateBanned")
	repo.AssertExpectations(t)
}

func TestSetBanned_Success(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID, Role: "Student"}, nil)
	repo.On("UpdateBanned", testTargetID, true).Return(nil)

	err := svc.SetBanned(testActorID, testTargetID, true)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSetBanned_UnbanSuccess(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID, Role: "Student"}, nil)
	repo.On("UpdateBanned", testTargetID, false).Return(nil)

	err := svc.SetBanned(testActorID, testTargetID, false)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSetBanned_RepoErrorPropagates(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID}, nil)
	repo.On("UpdateBanned", testTargetID, true).Return(errors.New("db down"))

	err := svc.SetBanned(testActorID, testTargetID, true)

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrSelfBan)
	repo.AssertExpectations(t)
}

// -----------------------------------------------------------------------------
// CHANGE ROLE
// -----------------------------------------------------------------------------

func TestSetRole_InvalidRolesRejected(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	cases := []string{
		"",
		"   ",
		"student", // lowercase ditolak (whitelist case-sensitive)
		"ADMIN",   // uppercase ditolak
		"Superuser",
		"<script>",
	}

	for _, role := range cases {
		_, err := svc.SetRole(testActorID, testTargetID, role)
		assert.ErrorIs(t, err, ErrInvalidRole, "role %q harusnya ditolak", role)
	}
	repo.AssertNotCalled(t, "UpdateRole")
	repo.AssertNotCalled(t, "FindByID")
}

func TestSetRole_TrimsWhitespace(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID, Role: "Student"}, nil)
	repo.On("UpdateRole", testTargetID, "Lecturer").Return(nil)

	updated, err := svc.SetRole(testActorID, testTargetID, " Lecturer ")

	assert.NoError(t, err)
	assert.Equal(t, "Lecturer", updated.Role)
	repo.AssertExpectations(t)
}

func TestSetRole_TargetNotFound(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("FindByID", testTargetID).Return(nil, repository.ErrNotFound)

	_, err := svc.SetRole(testActorID, testTargetID, "Lecturer")

	assert.ErrorIs(t, err, repository.ErrNotFound)
	repo.AssertNotCalled(t, "UpdateRole")
	repo.AssertExpectations(t)
}

func TestSetRole_StudentToLecturer_PersistsAndReturns(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID, Role: "Student"}, nil)
	repo.On("UpdateRole", testTargetID, "Lecturer").Return(nil)

	updated, err := svc.SetRole(testActorID, testTargetID, "Lecturer")

	assert.NoError(t, err)
	assert.Equal(t, "Lecturer", updated.Role)
	repo.AssertExpectations(t)
}

func TestSetRole_PromoteToAdmin_NoLastAdminGuardNeeded(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	// Target Student -> Admin: guard CountByRole tidak boleh dipanggil.
	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID, Role: "Student"}, nil)
	repo.On("UpdateRole", testTargetID, "Admin").Return(nil)

	updated, err := svc.SetRole(testActorID, testTargetID, "Admin")

	assert.NoError(t, err)
	assert.Equal(t, "Admin", updated.Role)
	repo.AssertExpectations(t)
}

func TestSetRole_DemoteLastAdmin_Rejected(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	// Target Admin, role saat ini Admin, jumlah admin di DB = 1 -> ditolak.
	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID, Role: "Admin"}, nil)
	repo.On("CountByRole", "Admin").Return(int64(1), nil)

	_, err := svc.SetRole(testActorID, testTargetID, "Student")

	assert.ErrorIs(t, err, ErrLastAdmin)
	repo.AssertNotCalled(t, "UpdateRole")
	repo.AssertExpectations(t)
}

func TestSetRole_DemoteSelfLastAdmin_Rejected(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	// Admin turunkan role dirinya sendiri padahal dia satu-satunya admin.
	repo.On("FindByID", testActorID).Return(&repository.User{ID: testActorID, Role: "Admin"}, nil)
	repo.On("CountByRole", "Admin").Return(int64(1), nil)

	_, err := svc.SetRole(testActorID, testActorID, "Student")

	assert.ErrorIs(t, err, ErrLastAdmin)
	repo.AssertNotCalled(t, "UpdateRole")
	repo.AssertExpectations(t)
}

func TestSetRole_DemoteAdmin_WithOtherAdminsAllowed(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	// Ada 2 admin -> demote satu admin sah.
	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID, Role: "Admin"}, nil)
	repo.On("CountByRole", "Admin").Return(int64(2), nil)
	repo.On("UpdateRole", testTargetID, "Student").Return(nil)

	updated, err := svc.SetRole(testActorID, testTargetID, "Student")

	assert.NoError(t, err)
	assert.Equal(t, "Student", updated.Role)
	repo.AssertExpectations(t)
}

func TestSetRole_RepoUpdateErrorPropagates(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID, Role: "Student"}, nil)
	repo.On("UpdateRole", testTargetID, "Lecturer").Return(errors.New("db down"))

	_, err := svc.SetRole(testActorID, testTargetID, "Lecturer")

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestSetRole_CountByRoleErrorPropagates(t *testing.T) {
	repo := new(MockUserRepository)
	svc := newTestAdminService(repo)

	repo.On("FindByID", testTargetID).Return(&repository.User{ID: testTargetID, Role: "Admin"}, nil)
	repo.On("CountByRole", "Admin").Return(int64(0), errors.New("db down"))

	_, err := svc.SetRole(testActorID, testTargetID, "Student")

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

// Guard statis: AdminService tetap memenuhi interface (mencegah drift).
var _ AdminService = (*adminService)(nil)
var _ mock.TestingT = (*testing.T)(nil)
