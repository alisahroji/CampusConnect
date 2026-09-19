-- Minggu 7 Day 2: kolom updated_at untuk fitur edit material.
-- Tabel materials (000011) dibuat append-only; kini CRUD update metadata
-- (judul/kategori) dibutuhkan, jadi timestamp edit mengikuti pola entity
-- existing (projects/posts: updated_at = created_at saat insert, GORM
-- men-set otomatis saat UPDATE).
ALTER TABLE materials
    ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP;
