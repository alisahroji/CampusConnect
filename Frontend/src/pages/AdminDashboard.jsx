import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ShieldAlert } from 'lucide-react';
import api from '../services/api';
import { isLoggedIn } from '../utils/auth';
import UserManagement from '../components/admin/UserManagement';

// Halaman Admin Dashboard (Minggu 6 Day 5) — skeleton + User Management.
//
// Access control UI: role dibaca dari BACKEND (GET /api/profile, dipakai
// RequireAuth), bukan dari data yang diketik client. Frontend restriction
// BUKAN security boundary — backend AdminGuard (Day 4) tetap authority;
// Student/Lecturer yang memaksa lewat API akan tetap ditolak 403.
const AdminDashboard = () => {
  const navigate = useNavigate();
  const [checking, setChecking] = useState(true);
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    let cancelled = false;

    const checkRole = async () => {
      if (!isLoggedIn()) {
        navigate('/login');
        return;
      }
      try {
        const res = await api.get('/profile');
        if (cancelled) return;
        const role = res.data?.user?.role;
        if (role === 'Admin') {
          setIsAdmin(true);
        }
        setChecking(false);
      } catch (err) {
        if (cancelled) return;
        if (err.response?.status === 401) {
          navigate('/login');
          return;
        }
        // Error lain (mis. server down): perlakukan sebagai bukan admin —
        // backend tetap menolak API admin bila sebenarnya admin.
        setChecking(false);
      }
    };

    checkRole();
    return () => {
      cancelled = true;
    };
  }, [navigate]);

  if (checking) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center font-body">
        <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (!isAdmin) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex flex-col items-center justify-center font-body px-6 text-center">
        <ShieldAlert className="w-14 h-14 text-[#D97757] mb-4" />
        <h1 className="font-display text-2xl font-semibold text-[#1E293B]">Akses Ditolak</h1>
        <p className="text-[#64748B] mt-2 max-w-sm">
          Halaman ini khusus admin. Akun kamu tidak memiliki izin untuk membukanya.
        </p>
        <Link
          to="/feed"
          className="mt-6 inline-block bg-[#112320] text-[#F8F9FA] font-bold text-sm uppercase tracking-wider py-3 px-6 rounded-full hover:bg-[#1E293B] transition-colors"
        >
          ← Kembali ke Feed
        </Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 py-12">
        {/* Skeleton dashboard: judul + konteks admin + entry User Management.
            Tanpa chart/analytics/metrics dummy — scope Week 6 Day 5. */}
        <header className="mb-10">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-[#112320] text-white rounded-xl flex items-center justify-center font-bold">
              A
            </div>
            <div>
              <h1 className="font-display text-3xl md:text-4xl font-semibold text-[#1E293B]">
                Admin Dashboard
              </h1>
              <p className="text-[#64748B] mt-1">
                Kelola pengguna CampusConnect — role dan status akses.
              </p>
            </div>
          </div>
        </header>

        {/* Entry utama skeleton: User Management (satu-satunya modul Week 6) */}
        <section>
          <UserManagement />
        </section>
      </div>
    </div>
  );
};

export default AdminDashboard;
