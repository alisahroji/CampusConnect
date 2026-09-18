import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import { isLoggedIn } from '../utils/auth';

// Waktu relatif sederhana (sama semangatnya dengan PostCard)
const formatRelativeTime = (isoDate) => {
  if (!isoDate) return '';
  const diffMs = Date.now() - new Date(isoDate).getTime();
  const minutes = Math.floor(diffMs / 60000);
  if (minutes < 1) return 'Baru saja';
  if (minutes < 60) return `${minutes} menit lalu`;
  if (minutes < 1440) return `${Math.floor(minutes / 60)} jam lalu`;
  return new Date(isoDate).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
};

// Label per tipe notifikasi (event nyata dari Day 2)
const NOTIF_META = {
  like_project: { icon: '❤️', text: 'menyukai project kamu' },
  comment_project: { icon: '💬', text: 'mengomentari project kamu' },
  like_post: { icon: '❤️', text: 'menyukai post kamu' },
  comment_post: { icon: '💬', text: 'mengomentari post kamu' },
  follow: { icon: '👤', text: 'mulai mengikuti kamu' },
};

// Halaman Notifikasi minimal (Minggu 6 Day 3): list notification milik user
// (endpoint Day 2) + aksi mark-read & read-all. Scope sengaja kecil — cukup
// agar badge navbar dapat digunakan secara normal.
const Notifications = () => {
  const navigate = useNavigate();
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [busyID, setBusyID] = useState('');

  useEffect(() => {
    if (!isLoggedIn()) {
      navigate('/login');
      return;
    }

    let cancelled = false;

    const fetchNotifications = async () => {
      try {
        setLoading(true);
        setError('');
        const res = await api.get('/notifications', { params: { limit: 50 } });
        if (cancelled) return;
        setNotifications(res.data.data || []);
        setUnreadCount(res.data.unread_count ?? 0);
      } catch (err) {
        if (!cancelled) {
          if (err.response?.status === 401) {
            navigate('/login');
            return;
          }
          setError(err.response?.data?.error || 'Gagal memuat notifikasi.');
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    fetchNotifications();
    return () => {
      cancelled = true;
    };
  }, [navigate]);

  const refreshUnread = async () => {
    try {
      const res = await api.get('/notifications/unread-count');
      setUnreadCount(res.data.unread_count ?? 0);
    } catch {
      // Badge di navbar sendiri yang polling; di sini cukup diam.
    }
  };

  const handleMarkRead = async (id) => {
    setBusyID(id);
    try {
      await api.post(`/notifications/${id}/read`);
      // Update lokal agar tak perlu refetch penuh
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, read_at: new Date().toISOString() } : n))
      );
      await refreshUnread();
    } catch (err) {
      setError(err.response?.data?.error || 'Gagal menandai notifikasi.');
    } finally {
      setBusyID('');
    }
  };

  const handleReadAll = async () => {
    try {
      await api.post('/notifications/read-all');
      setNotifications((prev) =>
        prev.map((n) => ({ ...n, read_at: n.read_at || new Date().toISOString() }))
      );
      setUnreadCount(0);
    } catch (err) {
      setError(err.response?.data?.error || 'Gagal menandai semua notifikasi.');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center font-body">
        <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body">
      <div className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <header className="mb-8 flex items-start justify-between gap-4">
          <div>
            <h1 className="font-display text-4xl font-semibold text-[#1E293B]">Notifikasi</h1>
            <p className="text-[#64748B] mt-2">
              {unreadCount > 0
                ? `${unreadCount} notifikasi belum dibaca.`
                : 'Semua notifikasi sudah dibaca.'}
            </p>
          </div>
          {unreadCount > 0 && (
            <button
              type="button"
              onClick={handleReadAll}
              className="shrink-0 bg-white border border-[#E2E8F0] hover:border-[#D97757] hover:text-[#D97757] text-[#1E293B] text-xs font-bold uppercase tracking-wider py-2.5 px-4 rounded-full transition-colors cursor-pointer"
            >
              Tandai semua dibaca
            </button>
          )}
        </header>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
            {error}
          </div>
        )}

        {notifications.length === 0 ? (
          <div className="text-center py-20 bg-white rounded-2xl border border-[#E2E8F0]">
            <p className="text-5xl mb-4">🔔</p>
            <p className="font-display text-xl font-semibold text-[#1E293B]">
              Belum ada notifikasi
            </p>
            <p className="text-sm text-[#64748B] mt-2 max-w-sm mx-auto">
              Notifikasi muncul saat ada yang menyukai, mengomentari, atau mengikuti kamu.
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {notifications.map((notif) => {
              const meta = NOTIF_META[notif.type] || { icon: '🔔', text: notif.type };
              const unread = !notif.read_at;

              return (
                <div
                  key={notif.id}
                  className={`flex items-start gap-3 rounded-2xl border p-4 transition-colors ${
                    unread
                      ? 'bg-[#D97757]/5 border-[#D97757]/30'
                      : 'bg-white border-[#E2E8F0]'
                  }`}
                >
                  <span className="text-xl leading-none mt-0.5">{meta.icon}</span>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-[#1E293B] leading-relaxed">
                      <span className="font-bold">
                        {notif.actor?.name || 'Seseorang'}
                      </span>{' '}
                      {meta.text}
                    </p>
                    <p className="text-xs text-[#94A3B8] mt-0.5">
                      {formatRelativeTime(notif.created_at)}
                    </p>
                  </div>
                  {unread && (
                    <button
                      type="button"
                      onClick={() => handleMarkRead(notif.id)}
                      disabled={busyID === notif.id}
                      className="shrink-0 text-xs font-bold text-[#64748B] hover:text-[#D97757] px-3 py-1.5 rounded-full hover:bg-[#F8F9FA] transition-colors disabled:opacity-60 cursor-pointer"
                    >
                      {busyID === notif.id ? '…' : 'Tandai dibaca'}
                    </button>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

export default Notifications;
