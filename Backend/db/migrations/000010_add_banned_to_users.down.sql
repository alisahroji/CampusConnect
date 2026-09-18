-- Rollback migration 000010: hapus kolom banned.
ALTER TABLE users
    DROP COLUMN IF EXISTS banned;
