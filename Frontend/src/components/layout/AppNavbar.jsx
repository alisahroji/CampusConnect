import { useEffect } from 'react';
import { Link, NavLink, useLocation, useNavigate } from 'react-router-dom';
import { Folder, Home, LogOut, Search, User } from 'lucide-react';
import { logout } from '../../utils/auth';

// Navbar utama hanya tampil pada application routes. Landing, halaman auth,
// dan Dashboard (sudah memakai DashboardLayout sendiri) tidak memakai navbar ini.
const HIDDEN_PREFIXES = ['/login', '/register', '/auth/callback', '/dashboard'];

const isNavbarHidden = (pathname) =>
  pathname === '/' || HIDDEN_PREFIXES.some((prefix) => pathname.startsWith(prefix));

const NAV_ITEMS = [
  { to: '/feed', label: 'Home', icon: Home },
  { to: '/projects', label: 'Projects', icon: Folder },
  { to: '/search', label: 'Search', icon: Search },
];

export default function AppNavbar() {
  const location = useLocation();
  const navigate = useNavigate();
  const visible = !isNavbarHidden(location.pathname);

  // Beri ruang untuk bottom nav mobile agar konten paling bawah tetap terjangkau
  // (padding dibersihkan otomatis saat navbar tidak tampil / komponen unmount).
  useEffect(() => {
    document.body.style.paddingBottom = visible ? '4.5rem' : '';
    return () => {
      document.body.style.paddingBottom = '';
    };
  }, [visible]);

  if (!visible) return null;

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const isProfileActive = location.pathname === '/profile';

  const desktopLinkClass = ({ isActive }) =>
    `px-4 py-2 text-sm font-bold uppercase tracking-wide rounded-full transition-colors ${
      isActive
        ? 'bg-[#112320] text-[#F8F9FA]'
        : 'text-[#64748B] hover:text-[#112320] hover:bg-[#F8F9FA]'
    }`;

  return (
    <>
      {/* Desktop: top bar */}
      <header className="hidden md:block sticky top-0 z-40 bg-white/90 backdrop-blur border-b border-[#E2E8F0]">
        <div className="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between gap-6">
          <Link to="/feed" className="flex items-center gap-2 shrink-0">
            <div className="w-8 h-8 bg-[#D97757] text-white rounded-lg flex items-center justify-center font-bold">
              C
            </div>
            <span className="text-lg font-bold tracking-tight text-[#1E293B] font-display">
              CampusConnect
            </span>
          </Link>

          <nav className="flex items-center gap-1">
            {NAV_ITEMS.map(({ to, label }) => (
              <NavLink key={to} to={to} className={desktopLinkClass}>
                {label}
              </NavLink>
            ))}
          </nav>

          <div className="flex items-center gap-3">
            <Link
              to="/profile"
              title="Profil"
              aria-label="Profil"
              className={`w-9 h-9 rounded-full flex items-center justify-center transition-shadow ${
                isProfileActive
                  ? 'bg-[#D97757] text-white ring-2 ring-[#C26244]'
                  : 'bg-[#112320] text-[#F8F9FA] hover:ring-2 hover:ring-[#D97757]'
              }`}
            >
              <User className="w-5 h-5" />
            </Link>
            <button
              type="button"
              onClick={handleLogout}
              className="flex items-center gap-2 text-sm font-bold uppercase tracking-wide text-[#64748B] hover:text-red-600 border border-[#E2E8F0] hover:border-red-200 hover:bg-red-50 rounded-full px-4 py-2 transition-colors cursor-pointer"
            >
              <LogOut className="w-4 h-4" />
              Keluar
            </button>
          </div>
        </div>
      </header>

      {/* Mobile: bottom navigation (5 slot, flex-1 agar tidak overflow di layar sempit) */}
      <nav className="md:hidden fixed bottom-0 inset-x-0 z-40 bg-white border-t border-[#E2E8F0] flex items-stretch justify-around py-2">
        {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              `flex-1 flex flex-col items-center gap-0.5 px-1 py-1 min-w-0 ${
                isActive ? 'text-[#D97757]' : 'text-[#94A3B8]'
              }`
            }
          >
            <Icon className="w-6 h-6" />
            <span className="text-[10px] font-bold">{label}</span>
          </NavLink>
        ))}
        <Link
          to="/profile"
          className={`flex-1 flex flex-col items-center gap-0.5 px-1 py-1 min-w-0 ${
            isProfileActive ? 'text-[#D97757]' : 'text-[#94A3B8]'
          }`}
        >
          <User className="w-6 h-6" />
          <span className="text-[10px] font-bold">Profil</span>
        </Link>
        <button
          type="button"
          onClick={handleLogout}
          className="flex-1 flex flex-col items-center gap-0.5 px-1 py-1 min-w-0 text-[#94A3B8] cursor-pointer"
        >
          <LogOut className="w-6 h-6" />
          <span className="text-[10px] font-bold">Keluar</span>
        </button>
      </nav>
    </>
  );
}
