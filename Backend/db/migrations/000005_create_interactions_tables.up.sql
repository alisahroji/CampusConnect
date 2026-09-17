-- Fitur Interaksi Minggu 4 - Hari 3 (Like & Comment)

-- Tabel untuk fitur Komentar pada project showcase
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabel untuk fitur Like (satu user hanya bisa like satu kali per project)
CREATE TABLE likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_likes_project_user UNIQUE (project_id, user_id)
);

-- Index untuk mempercepat query komentar & like milik suatu project
CREATE INDEX idx_comments_project_id ON comments(project_id);
CREATE INDEX idx_likes_project_id ON likes(project_id);
