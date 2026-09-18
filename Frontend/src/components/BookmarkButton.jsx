import { useEffect, useState } from 'react';
import { Bookmark, BookmarkCheck } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import { isLoggedIn } from '../utils/auth';

// Tombol bookmark reusable untuk Project & Post (Minggu 6 Day 3).
//  - State awal TIDAK hardcoded false: bila user login dan initialBookmarked
//    tidak diberikan, status di-bootstrap dari endpoint status backend
//    (GET /{projects|posts}/:id/bookmark, OptionalAuth) sehingga ikon tetap
//    benar setelah refresh.
//  - Toggle selalu memakai hasil server (POST .../bookmark -> {bookmarked})
//    sebagai sumber kebenaran — pelajaran dari BUG-1/BUG-5A like.
//  - Anonymous: tidak ada request status, klik diarahkan ke /login
//    (konsisten dengan pola like/komentar existing).
const BookmarkButton = ({ type, id, initialBookmarked = null, onError }) => {
  const navigate = useNavigate();
  const [bookmarked, setBookmarked] = useState(Boolean(initialBookmarked));
  // Anonymous langsung "ready" (klik -> /login, konsisten pola like/komentar);
  // user login tanpa initialBookmarked menunggu bootstrap status dari server.
  const [bootstrapped, setBootstrapped] = useState(
    initialBookmarked !== null || !isLoggedIn()
  );
  const [busy, setBusy] = useState(false);

  // Bootstrap status dari server (hanya untuk user login & state belum diketahui)
  useEffect(() => {
    let cancelled = false;

    const bootstrap = async () => {
      if (!isLoggedIn() || initialBookmarked !== null) return;
      try {
        const res = await api.get(`/${type}s/${id}/bookmark`);
        if (!cancelled) {
          setBookmarked(Boolean(res.data.bookmarked));
          setBootstrapped(true);
        }
      } catch {
        // Gagal bootstrap: tombol tetap tampil sebagai belum di-bookmark;
        // toggle berikutnya tetap akan dikoreksi oleh response server.
        if (!cancelled) setBootstrapped(true);
      }
    };

    bootstrap();
    return () => {
      cancelled = true;
    };
  }, [type, id, initialBookmarked]);

  const handleToggle = async () => {
    if (!isLoggedIn()) {
      navigate('/login');
      return;
    }
    try {
      setBusy(true);
      const res = await api.post(`/${type}s/${id}/bookmark`);
      setBookmarked(Boolean(res.data.bookmarked));
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      onError?.(err.response?.data?.error || 'Gagal memperbarui bookmark');
    } finally {
      setBusy(false);
    }
  };

  return (
    <button
      type="button"
      onClick={handleToggle}
      disabled={busy || !bootstrapped}
      title={bookmarked ? 'Hapus bookmark' : 'Simpan ke bookmark'}
      aria-label={bookmarked ? 'Hapus bookmark' : 'Simpan ke bookmark'}
      aria-pressed={bookmarked}
      className={`flex items-center gap-2 font-bold text-sm py-2.5 px-5 rounded-full border transition-all duration-300 disabled:opacity-60 cursor-pointer ${
        bookmarked
          ? 'bg-[#112320] border-[#112320] text-[#F8F9FA] hover:bg-[#1E293B]'
          : 'bg-white border-[#E2E8F0] text-[#1E293B] hover:border-[#112320] hover:text-[#112320]'
      }`}
    >
      {bookmarked ? (
        <BookmarkCheck className="w-4 h-4" />
      ) : (
        <Bookmark className="w-4 h-4" />
      )}
      {busy ? '…' : bookmarked ? 'Tersimpan' : 'Simpan'}
    </button>
  );
};

export default BookmarkButton;
