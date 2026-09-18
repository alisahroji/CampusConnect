-- Minggu 6 Hari 4: kolom banned untuk Admin User Management.
-- Satu pola saja: banned BOOLEAN NOT NULL DEFAULT FALSE.
-- Existing users otomatis FALSE lewat DEFAULT.
ALTER TABLE users
    ADD COLUMN banned BOOLEAN NOT NULL DEFAULT FALSE;
