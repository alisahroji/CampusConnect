-- Minggu 6 (Hari 1): Bookmark untuk Project & Post
-- Mengikuti keputusan arsitektur migration 000006: interaksi user disimpan
-- pada tabel dedicated (bukan polymorphic) agar foreign key tetap jelas dan
-- unique constraint dapat ditegakkan langsung oleh database.

CREATE TABLE project_bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Satu user tidak bisa mem-bookmark project yang sama dua kali
    CONSTRAINT uq_project_bookmarks_pair UNIQUE (project_id, user_id)
);

CREATE TABLE post_bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_post_bookmarks_pair UNIQUE (post_id, user_id)
);

-- Query utama: "semua bookmark milik seorang user" (list halaman Bookmark).
-- Sisi project_id/post_id sudah tercakup oleh prefix unique constraint
-- (project_id, user_id) / (post_id, user_id), sehingga index tambahan
-- hanya diperlukan untuk kolom user_id.
CREATE INDEX idx_project_bookmarks_user_id ON project_bookmarks(user_id);
CREATE INDEX idx_post_bookmarks_user_id ON post_bookmarks(user_id);
