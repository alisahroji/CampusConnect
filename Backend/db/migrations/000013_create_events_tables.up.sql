-- Minggu 7 (Hari 4): Desain Schema Event — Fase 3 Materials & Events
--
-- Roadmap field minimal: judul, tanggal, lokasi/link, RSVP.
-- Keputusan desain (lihat laporan Day 4):
--   * creator_id: nama mengikuti konsep "pembuat" entity (precedent user_id/
--     uploader_id pada posts/materials). Roadmap belum menetapkan siapa yang
--     boleh membuat event — policy CRUD ditetapkan di Day 5; schema hanya
--     menyimpan referensi pembuat agar ownership dapat ditegakkan nanti.
--   * starts_at TIMESTAMPTZ (bukan DATE): event akademik punya jam mulai;
--     TIMESTAMPTZ = konvensi seluruh tabel existing. Timezone disimpan UTC
--     di DB dan dirender di sisi aplikasi (pola created_at existing).
--     ends_at tidak dibuat — roadmap hanya menyebut "tanggal".
--   * location & event_url nullable berdua (VARCHAR 255, pola posts.image_url):
--     roadmap "lokasi/link" belum menentukan wajib-salah-satu, jadi CHECK
--     "minimal satu terisi" TIDAK dibuat (keputusan Day 5 bila policy final).
--   * description TEXT nullable (opsional, pola posts.content TEXT).
--   * RSVP = tabel dedicated event_rsvps mengikuti keputusan arsitektur
--     migration 000006/000008: interaksi user pada tabel dedicated agar FK
--     jelas dan unique constraint ditegakkan langsung oleh database.
--   * Tanpa kolom rsvp_count/cache di events: jumlah RSVP dihitung dari
--     tabel event_rsvps saat dibutuhkan (Day 5) — satu sumber kebenaran.

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(150) NOT NULL,
    description TEXT,
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    location VARCHAR(255),
    event_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Query "event yang saya buat" (Day 5: kelola event + ownership edit/delete).
CREATE INDEX idx_events_creator_id ON events(creator_id);
-- Feed/list event urut waktu ke depan (list utama halaman Events).
CREATE INDEX idx_events_starts_at ON events(starts_at);

CREATE TABLE event_rsvps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Satu user hanya boleh RSVP satu kali per event (duplicate ditolak DB).
    CONSTRAINT uq_event_rsvps_pair UNIQUE (event_id, user_id)
);

-- Sisi event_id sudah tercakup oleh prefix unique constraint
-- (event_id, user_id), sehingga index tambahan hanya diperlukan untuk
-- user_id ("semua event yang saya RSVP" pada halaman Events).
CREATE INDEX idx_event_rsvps_user_id ON event_rsvps(user_id);
