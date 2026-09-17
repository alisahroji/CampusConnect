import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';
import ProjectCard from '../components/ProjectCard';
import { isLoggedIn } from '../utils/auth';

const STATUS_TABS = [
  { value: '', label: 'Semua Status' },
  { value: 'published', label: 'Published' },
  { value: 'draft', label: 'Draft Saya' },
];

const ProjectList = () => {
  const [projects, setProjects] = useState([]);
  const [allTags, setAllTags] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedTag, setSelectedTag] = useState(''); // '' = Semua
  const [statusFilter, setStatusFilter] = useState(''); // '' = Semua Status
  const loggedIn = isLoggedIn();

  // 1. Kumpulkan daftar tag unik untuk pill filter (dari project published).
  //    Daftar tag ini hanya untuk MEMILIH filter — penyaringan aslinya terjadi di backend.
  useEffect(() => {
    const fetchTags = async () => {
      try {
        const res = await api.get('/projects', { params: { status: 'published' } });
        const tags = [
          ...new Set(
            (res.data.data || []).flatMap((project) =>
              (project.tech_stack || '')
                .split(',')
                .map((tag) => tag.trim())
                .filter(Boolean)
            )
          ),
        ];
        setAllTags(tags);
      } catch {
        // Pill filter tetap tampil berisi "Semua" bila gagal memuat tag
      }
    };
    fetchTags();
  }, []);

  // 2. List project — difilter di BACKEND lewat ?tech_stack= dan ?status=
  //    (bukan lagi mengambil semuanya lalu menyaring di browser).
  useEffect(() => {
    const fetchProjects = async () => {
      try {
        setLoading(true);
        setError('');
        const params = {};
        if (statusFilter) params.status = statusFilter;
        if (selectedTag) params.tech_stack = selectedTag;
        const res = await api.get('/projects', { params });
        setProjects(res.data.data || []);
      } catch (err) {
        setError(err.response?.data?.error || 'Gagal memuat daftar project.');
      } finally {
        setLoading(false);
      }
    };
    fetchProjects();
  }, [selectedTag, statusFilter]);

  const isFiltered = selectedTag !== '' || statusFilter !== '';

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 py-12">
        <header className="mb-10 flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="font-display text-4xl font-semibold text-[#1E293B]">Project Showcase</h1>
            <p className="text-[#64748B] mt-2">Jelajahi project kolaborasi dari komunitas kampus.</p>
          </div>
          {loggedIn && (
            <Link
              to="/projects/new"
              className="inline-block bg-[#D97757] hover:bg-[#C26244] text-white text-sm font-bold uppercase tracking-wider py-3 px-6 rounded-full transition-colors"
            >
              + Project Baru
            </Link>
          )}
        </header>

        {/* Tab status (draft hanya relevan untuk user yang login) */}
        {loggedIn && (
          <div className="flex flex-wrap gap-2 mb-4">
            {STATUS_TABS.map((tab) => (
              <button
                key={tab.label}
                type="button"
                onClick={() => setStatusFilter(tab.value)}
                className={`px-4 py-1.5 rounded-full text-xs font-bold uppercase tracking-wider border transition-colors ${
                  statusFilter === tab.value
                    ? 'bg-[#D97757] text-white border-[#D97757]'
                    : 'bg-white text-[#64748B] border-[#E2E8F0] hover:border-[#D97757]'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        )}

        {/* Filter pill berdasarkan tech stack — query dikirim ke backend */}
        <div className="flex flex-wrap gap-2 mb-8">
          {['Semua', ...allTags].map((tag) => (
            <button
              key={tag}
              type="button"
              onClick={() => setSelectedTag(tag === 'Semua' ? '' : tag)}
              className={`px-4 py-2 rounded-full text-sm font-semibold border transition-colors ${
                selectedTag === (tag === 'Semua' ? '' : tag)
                  ? 'bg-[#112320] text-[#F8F9FA] border-[#112320]'
                  : 'bg-white text-[#64748B] border-[#E2E8F0] hover:border-[#112320]'
              }`}
            >
              {tag}
            </button>
          ))}
        </div>

        {loading ? (
          // Loading spinner
          <div className="flex items-center justify-center py-24">
            <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
          </div>
        ) : error ? (
          // Error state
          <div className="p-6 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
            {error}
          </div>
        ) : projects.length === 0 ? (
          // Empty state (bisa karena memang kosong atau hasil filter kosong)
          <div className="text-center py-24">
            <p className="text-5xl mb-4">🗂️</p>
            <p className="font-display text-xl font-semibold text-[#1E293B]">
              {isFiltered ? 'Tidak ada project yang cocok' : 'Belum ada project'}
            </p>
            <p className="text-sm text-[#64748B] mt-1">
              {isFiltered
                ? 'Coba pilih kombinasi filter lain.'
                : 'Jadilah yang pertama membagikan project kamu!'}
            </p>
          </div>
        ) : (
          // Grid kartu project
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {projects.map((project) => (
              <ProjectCard key={project.id} project={project} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default ProjectList;
