import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import api from '../services/api';

// Halaman Search (Minggu 5 Hari 6) — terhubung penuh ke GET /api/search?q=
// (public). Tidak ada data dummy: semua hasil datang dari backend.
const Search = () => {
  const navigate = useNavigate();

  const [query, setQuery] = useState('');
  const [submitted, setSubmitted] = useState('');
  const [users, setUsers] = useState([]);
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [hasSearched, setHasSearched] = useState(false);

  const inputRef = useRef(null);

  const runSearch = async (q) => {
    const keyword = q.trim();
    if (!keyword) {
      setError('Kata kunci pencarian tidak boleh kosong.');
      return;
    }
    try {
      setLoading(true);
      setError('');
      const res = await api.get('/search', { params: { q: keyword } });
      setUsers(res.data.users || []);
      setProjects(res.data.projects || []);
      setSubmitted(keyword);
      setHasSearched(true);
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      // 400 (query kosong dsb.) ditampilkan jelas, bukan blank page
      setError(err.response?.data?.error || 'Gagal melakukan pencarian.');
      setUsers([]);
      setProjects([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    runSearch(query);
  };

  // Fokuskan input saat halaman dibuka
  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const noResult =
    hasSearched && !loading && users.length === 0 && projects.length === 0;

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body selection:bg-[#D97757] selection:text-white">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 py-10">
        <header className="mb-8">
          <h1 className="font-display text-3xl md:text-4xl font-semibold text-[#1E293B]">
            Pencarian
          </h1>
          <p className="text-[#64748B] mt-1">
            Temukan mahasiswa berbakat dan project keren di kampus.
          </p>
        </header>

        {/* Form pencarian */}
        <form onSubmit={handleSubmit} className="flex gap-2 mb-8">
          <div className="flex-1 relative">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-[#94A3B8]">
              🔍
            </span>
            <input
              ref={inputRef}
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Cari nama mahasiswa atau judul project..."
              className="w-full bg-white border border-[#E2E8F0] rounded-full pl-11 pr-5 py-3 text-sm text-[#1E293B] focus:border-[#D97757] focus:outline-none transition-colors shadow-sm placeholder:text-[#94A3B8]"
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="bg-[#D97757] hover:bg-[#C26244] disabled:bg-[#E2E8F0] disabled:text-[#94A3B8] text-white text-sm font-bold tracking-wider uppercase px-6 rounded-full transition-colors cursor-pointer shrink-0"
          >
            {loading ? '...' : 'Cari'}
          </button>
        </form>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
            {error}
          </div>
        )}

        {loading && (
          <div className="flex items-center justify-center py-20">
            <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
          </div>
        )}

        {!loading && noResult && (
          <div className="text-center py-16 bg-white rounded-2xl border border-[#E2E8F0]">
            <p className="text-5xl mb-4">🔭</p>
            <p className="font-display text-xl font-semibold text-[#1E293B]">
              Tidak ada hasil untuk &ldquo;{submitted}&rdquo;
            </p>
            <p className="text-sm text-[#64748B] mt-2">
              Coba kata kunci lain, misalnya nama atau tech stack.
            </p>
          </div>
        )}

        {/* Hasil: Users */}
        {!loading && users.length > 0 && (
          <section className="mb-8">
            <h2 className="text-xs font-bold text-[#94A3B8] uppercase tracking-widest border-b border-[#E2E8F0] pb-3 mb-4">
              Mahasiswa ({users.length})
            </h2>
            <div className="space-y-3">
              {users.map((user) => (
                <Link
                  key={user.id}
                  to={`/users/${user.id}`}
                  className="flex items-center gap-4 bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-4 hover:border-[#D97757] transition-colors"
                >
                  <img
                    src={
                      user.picture_url ||
                      `https://ui-avatars.com/api/?name=${user.name || 'A'}`
                    }
                    alt={user.name}
                    className="w-12 h-12 rounded-full object-cover border border-[#E2E8F0]"
                  />
                  <div className="flex-1 min-w-0">
                    <p className="font-bold text-[#1E293B] truncate">
                      {user.name}
                    </p>
                    <p className="text-xs text-[#94A3B8] uppercase tracking-wide">
                      {user.role || 'Student'}
                    </p>
                  </div>
                  <span className="text-[#D97757] font-bold">→</span>
                </Link>
              ))}
            </div>
          </section>
        )}

        {/* Hasil: Projects */}
        {!loading && projects.length > 0 && (
          <section>
            <h2 className="text-xs font-bold text-[#94A3B8] uppercase tracking-widest border-b border-[#E2E8F0] pb-3 mb-4">
              Project ({projects.length})
            </h2>
            <div className="space-y-3">
              {projects.map((project) => (
                <Link
                  key={project.id}
                  to={`/projects/${project.id}`}
                  className="block bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-4 hover:border-[#D97757] transition-colors"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="font-bold text-[#1E293B] truncate">
                        {project.title}
                      </p>
                      <p className="text-xs text-[#94A3B8] mt-0.5">
                        oleh {project.user?.name || 'Anonim'}
                      </p>
                      <p className="text-sm text-[#64748B] mt-2 line-clamp-2">
                        {project.description}
                      </p>
                    </div>
                    <span className="text-[#D97757] font-bold shrink-0">→</span>
                  </div>
                  {project.tech_stack && (
                    <div className="flex flex-wrap gap-1.5 mt-3">
                      {project.tech_stack
                        .split(',')
                        .map((t) => t.trim())
                        .filter(Boolean)
                        .slice(0, 5)
                        .map((t) => (
                          <span
                            key={t}
                            className="text-[11px] font-semibold bg-[#F8F9FA] border border-[#E2E8F0] text-[#112320] px-2.5 py-1 rounded-full"
                          >
                            {t}
                          </span>
                        ))}
                    </div>
                  )}
                </Link>
              ))}
            </div>
          </section>
        )}

        {/* Empty state sebelum mencari */}
        {!hasSearched && !loading && !error && (
          <div className="text-center py-16">
            <p className="text-5xl mb-4">🧭</p>
            <p className="text-[#64748B]">
              Mulai dengan mengetik nama mahasiswa atau judul project di atas.
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default Search;
