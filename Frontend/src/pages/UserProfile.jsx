import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import api from '../services/api';
import { getCurrentUserID } from '../utils/auth';

// Halaman profil publik user (Minggu 5 Hari 6 — Follow UI).
// State follow SELALU dibootstrap dari server (current_user.following pada
// GET /api/users/:id) sehingga tombol tetap sinkron setelah refresh.
const UserProfile = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const [user, setUser] = useState(null);
  const [following, setFollowing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [followBusy, setFollowBusy] = useState(false);
  const [error, setError] = useState('');

  const currentUID = getCurrentUserID();
  const isSelf = Boolean(currentUID) && currentUID === id;

  useEffect(() => {
    let cancelled = false;

    const loadProfile = async () => {
      try {
        setLoading(true);
        setError('');
        const res = await api.get(`/users/${id}`);
        if (cancelled) return;
        setUser(res.data.user);
        setFollowing(Boolean(res.data.current_user?.following));
      } catch (err) {
        if (cancelled) return;
        if (err.response?.status === 401) {
          navigate('/login');
          return;
        }
        setError(
          err.response?.status === 404
            ? 'User tidak ditemukan.'
            : err.response?.data?.error || 'Gagal memuat profil.'
        );
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    loadProfile();
    return () => {
      cancelled = true;
    };
  }, [id, navigate]);

  const handleToggleFollow = async () => {
    // Anonymous diminta login dulu (konsisten dengan gating like)
    if (!localStorage.getItem('access_token')) {
      navigate('/login');
      return;
    }
    try {
      setFollowBusy(true);
      setError('');
      const res = await api.post(`/users/${id}/follow`);
      // State baru diambil dari jawaban server, bukan dari asumsi UI
      setFollowing(Boolean(res.data.following));
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setError(err.response?.data?.error || 'Gagal memperbarui status follow.');
    } finally {
      setFollowBusy(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center font-body">
        <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error && !user) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex flex-col items-center justify-center font-body px-4">
        <p className="text-5xl mb-4">🫥</p>
        <p className="font-display text-xl font-semibold text-[#1E293B] mb-4">
          {error}
        </p>
        <Link
          to="/search"
          className="bg-[#D97757] hover:bg-[#C26244] text-white text-sm font-bold py-2.5 px-6 rounded-full transition-colors"
        >
          Cari User Lain
        </Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body selection:bg-[#D97757] selection:text-white">
      {/* Banner atas */}
      <div className="h-48 md:h-64 bg-[#112320] relative overflow-hidden">
        <div
          className="absolute inset-0 opacity-20"
          style={{
            backgroundImage:
              'linear-gradient(#F8F9FA 1px, transparent 1px), linear-gradient(90deg, #F8F9FA 1px, transparent 1px)',
            backgroundSize: '4rem 4rem',
          }}
        ></div>
      </div>

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-12 -mt-20 relative z-10 pb-16">
        <div className="flex flex-col sm:flex-row gap-6 sm:items-end">
          {/* Avatar */}
          <div className="relative shrink-0">
            <img
              src={
                user?.picture_url ||
                `https://ui-avatars.com/api/?name=${user?.name || 'A'}`
              }
              alt={user?.name}
              className="w-32 h-32 sm:w-40 sm:h-40 object-cover border-4 border-[#F8F9FA] shadow-lg bg-white"
              style={{ borderRadius: '4px' }}
            />
          </div>

          {/* Identitas */}
          <div className="flex-1 pb-1 min-w-0">
            <h1 className="font-display text-3xl md:text-4xl font-bold text-[#1E293B] mb-1 break-words">
              {user?.name || 'Tanpa Nama'}
            </h1>
            <p className="text-[#D97757] font-bold tracking-wide uppercase text-sm">
              {user?.role || 'Student'}
            </p>
            <p className="text-xs text-[#94A3B8] mt-1">
              Bergabung{' '}
              {user?.created_at
                ? new Date(user.created_at).toLocaleDateString('id-ID', {
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric',
                  })
                : '-'}
            </p>
          </div>

          {/* Tombol Follow / label self */}
          <div className="pb-1 shrink-0">
            {isSelf ? (
              <div className="flex flex-col gap-2 items-stretch">
                <span className="text-xs font-bold text-[#94A3B8] uppercase tracking-widest text-center">
                  Ini Profil Kamu
                </span>
                <Link
                  to="/edit-profile"
                  className="inline-block bg-white border border-[#E2E8F0] hover:border-[#112320] text-[#1E293B] font-bold py-2.5 px-6 text-sm tracking-wider uppercase transition-colors shadow-sm text-center"
                >
                  Edit Profil
                </Link>
              </div>
            ) : (
              <button
                type="button"
                onClick={handleToggleFollow}
                disabled={followBusy}
                className={`text-sm font-bold tracking-wider uppercase py-2.5 px-8 rounded-full transition-all duration-300 disabled:opacity-60 cursor-pointer ${
                  following
                    ? 'bg-[#112320] text-[#F8F9FA] hover:bg-[#0b1826]'
                    : 'bg-[#D97757] text-white hover:bg-[#C26244]'
                }`}
              >
                {followBusy
                  ? 'Memproses...'
                  : following
                    ? '✓ Following'
                    : '+ Follow'}
              </button>
            )}
          </div>
        </div>

        {error && user && (
          <div className="mt-6 p-3 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
            {error}
          </div>
        )}

        {/* Bio & detail */}
        <div className="mt-10">
          <h3 className="text-xs font-bold text-[#94A3B8] uppercase tracking-widest border-b border-[#E2E8F0] pb-4 mb-6">
            Tentang
          </h3>
          <p className="text-[#1E293B] leading-relaxed text-lg font-light">
            {user?.bio || 'Belum ada bio yang ditulis.'}
          </p>
        </div>
      </div>
    </div>
  );
};

export default UserProfile;
