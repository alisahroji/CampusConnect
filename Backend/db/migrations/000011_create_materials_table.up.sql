-- Minggu 7 (Hari 1): Desain Schema Material — Fase 3 Materials & Events
--
-- Roadmap field minimal: judul, kategori/mata kuliah, file, uploader.
-- Keputusan desain (lihat laporan Day 1):
--   * file reference: file_url (public URL Cloudinary, pola project_galleries
--     000007 / image_url) + original_filename + file_size supaya UI bisa
--     menampilkan nama & ukuran unduhan tanpa hit ke storage.
--   * no updated_at di migration: tabel ini append-only pada Week 7 (update
--     kembali mengunggah file baru). Model tetap punya UpdatedAt agar konsisten
--     dengan pola GORM existing; kolomnya akan ditambahkan saat kebutuhan edit
--     material benar-benar ada (dengan timestamp convention yang sama).
--   * owner FK: ON DELETE CASCADE mengikuti seluruh precedent entity
--     (projects.user_id, posts.user_id) — milik user dihapus bersama user.
--   * Tanpa unique constraint pada judul/kategori: dua materi berbeda boleh
--     memiliki judul sama (mis. "Modul 1" di dua mata kuliah berbeda).
--   * Tanpa CHECK role di sini: penegakan "siapa boleh upload" adalah urusan
--     service/middleware (Day 2), bukan schema.

CREATE TABLE materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(150) NOT NULL,
    category VARCHAR(100) NOT NULL,
    file_url VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0 CHECK (file_size >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Query utama Day 2+: daftar materi per kategori/mata kuliah (halaman Materials
-- dikelompokkan per mata kuliah).
CREATE INDEX idx_materials_category ON materials(category);
-- Query ownership: materi milik satu uploader (edit/delete + "materi saya").
CREATE INDEX idx_materials_uploader_id ON materials(uploader_id);
-- Feed materi terbaru dulu (list utama tanpa filter).
CREATE INDEX idx_materials_created_at ON materials(created_at DESC);
