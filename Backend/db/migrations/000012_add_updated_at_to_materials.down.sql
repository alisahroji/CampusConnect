-- Rollback migration 000012: hapus kolom updated_at dari materials.
ALTER TABLE materials
    DROP COLUMN IF EXISTS updated_at;
