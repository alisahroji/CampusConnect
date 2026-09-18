package service

import (
	"campusconnect/repository"
	"errors"
	"strings"
)

// Whitelist role (Minggu 6 Hari 4) — nilai yang sama dengan convention DB
// (migration 000001: DEFAULT 'Student').
const (
	RoleStudent  = "Student"
	RoleLecturer = "Lecturer"
	RoleAdmin    = "Admin"
)

// ValidRoles: whitelist eksplisit. Role di luar daftar ini (termasuk empty
// string, lowercase variant, atau string arbitrer) ditolak di service.
var ValidRoles = map[string]bool{
	RoleStudent:  true,
	RoleLecturer: true,
	RoleAdmin:    true,
}

// Sentinel errors Day 4 — di-mapping handler ke status code:
//   - ErrSelfBan / ErrInvalidRole / ErrLastAdmin -> 400
//   - repository.ErrNotFound                     -> 404
//   - selain itu                                 -> 500
var (
	ErrInvalidRole = errors.New("role tidak valid (gunakan: Student, Lecturer, atau Admin)")
	ErrSelfBan     = errors.New("admin tidak bisa memblokir dirinya sendiri")
	ErrLastAdmin   = errors.New("operasi ditolak: minimal satu Admin harus tetap tersisa")
)

// AdminService: operasi manajemen user untuk admin (Minggu 6 Hari 4).
// Authorization dilakukan di middleware (RequireAuth + AdminGuard); service
// fokus pada business rule (self-ban, whitelist role, last-admin guard).
type AdminService interface {
	ListUsers(limit int) ([]repository.User, error)
	SetBanned(actorID, targetID string, banned bool) error
	SetRole(actorID, targetID string, role string) (*repository.User, error)
}

type adminService struct {
	userRepo repository.UserRepository
}

func NewAdminService(userRepo repository.UserRepository) AdminService {
	return &adminService{userRepo: userRepo}
}

// ListUsers mengembalikan daftar user untuk management (data dari DB,
// tanpa dummy). Penyaringan field sensitif dilakukan DTO di handler.
func (s *adminService) ListUsers(limit int) ([]repository.User, error) {
	return s.userRepo.List(limit)
}

// SetBanned memblokir / melepas blokir user.
// Self-ban selalu ditolak: admin tidak boleh mengunci dirinya sendiri.
// (Banned admin lain oleh admin aktif tidak bisa menghabiskan admin —
// actor sendiri selalu tersisa aktif, sehingga guard last-admin di sini
// tidak diperlukan untuk ban.)
func (s *adminService) SetBanned(actorID, targetID string, banned bool) error {
	if actorID == targetID {
		return ErrSelfBan
	}
	// Pastikan target ada (404, bukan silent update 0 row).
	if _, err := s.userRepo.FindByID(targetID); err != nil {
		return err
	}
	return s.userRepo.UpdateBanned(targetID, banned)
}

// SetRole mengubah role user dengan whitelist ketat.
// Guard last-admin: demote Admin (termasuk oleh dirinya sendiri) ditolak
// bila jumlah Admin tepat 1 — sistem tidak boleh kehilangan admin terakhir.
func (s *adminService) SetRole(actorID, targetID string, role string) (*repository.User, error) {
	role = strings.TrimSpace(role)
	if !ValidRoles[role] {
		return nil, ErrInvalidRole
	}

	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return nil, err
	}

	if target.Role == RoleAdmin && role != RoleAdmin {
		totalAdmin, err := s.userRepo.CountByRole(RoleAdmin)
		if err != nil {
			return nil, err
		}
		if totalAdmin <= 1 {
			return nil, ErrLastAdmin
		}
	}

	if err := s.userRepo.UpdateRole(targetID, role); err != nil {
		return nil, err
	}
	target.Role = role
	return target, nil
}
