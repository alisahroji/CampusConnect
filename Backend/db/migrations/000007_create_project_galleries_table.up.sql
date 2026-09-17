-- Fitur Minggu 4: Project Gallery (satu project memiliki banyak gambar)

CREATE TABLE project_galleries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    image_url VARCHAR(255) NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk mempercepat pengambilan galeri milik suatu project
CREATE INDEX idx_project_galleries_project_id ON project_galleries(project_id);
