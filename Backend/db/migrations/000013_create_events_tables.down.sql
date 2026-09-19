-- Rollback Minggu 7 Day 4: hapus schema Event.
-- Urutan: event_rsvps dulu (mereferensi events), lalu events.
DROP TABLE IF EXISTS event_rsvps;
DROP TABLE IF EXISTS events;
