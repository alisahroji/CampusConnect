-- Minggu 6 (Hari 2): Notification dasar (storage-only, belum realtime)
-- Realtime (WebSocket/Redis Pub-Sub) baru masuk scope Minggu 10.

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- Actor boleh NULL (mis. akun actor dihapus): FK pakai SET NULL
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL, -- like_project | comment_project | like_post | comment_post | follow
    -- ID entitas terkait (project_id / post_id) untuk deep-link di frontend.
    -- Untuk type 'follow' kosong karena aktor sudah tersedia di actor_id.
    entity_id UUID,
    -- NULL berarti belum dibaca (unread)
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Query utama: daftar notifikasi milik recipient (terbaru dulu)
CREATE INDEX idx_notifications_recipient_id ON notifications(recipient_id, created_at DESC);
-- Partial index khusus badge unread count (hanya row yang belum dibaca)
CREATE INDEX idx_notifications_recipient_unread ON notifications(recipient_id) WHERE read_at IS NULL;
