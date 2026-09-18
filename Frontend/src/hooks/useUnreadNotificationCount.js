import { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import api from '../services/api';
import { isLoggedIn } from '../utils/auth';

// Interval polling badge notifikasi (Minggu 6 = polling, realtime menyusul
// Minggu 10). 30 detik cukup responsif tanpa membebani backend.
const POLL_INTERVAL_MS = 30000;

// Hook kecil untuk badge notifikasi di navbar:
//  - polling GET /notifications/unread-count HANYA saat user login;
//  - berhenti (interval dibersihkan) saat logout / unmount;
//  - efek bergantung pada pathname karena login/logout selalu navigasi,
//    sehingga sesi baru otomatis memulai polling dengan user context baru;
//  - kegagalan request tidak crash navbar (fallback 0, tanpa angka palsu).
export default function useUnreadNotificationCount() {
  const location = useLocation();
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    // Anonymous / baru logout: tanpa request. Count tetap 0 (nilai awal),
    // dan navbar seluruhnya hilang di halaman auth sehingga tak terlihat.
    if (!isLoggedIn()) {
      return undefined;
    }

    let cancelled = false;

    const fetchCount = async () => {
      // Guard di dalam interval juga: token bisa hilang di antara tick
      // (logout dari tab lain) — jangan kirim request tanpa sesi.
      if (!isLoggedIn()) {
        setUnreadCount(0);
        return;
      }
      try {
        const res = await api.get('/notifications/unread-count');
        if (!cancelled) setUnreadCount(res.data.unread_count ?? 0);
      } catch {
        // Fallback aman: badge hilang, navbar tetap hidup.
        if (!cancelled) setUnreadCount(0);
      }
    };

    fetchCount();
    const timer = setInterval(fetchCount, POLL_INTERVAL_MS);

    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [location.pathname]);

  return unreadCount;
}
