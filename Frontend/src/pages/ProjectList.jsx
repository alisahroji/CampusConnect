import { useEffect, useState } from 'react';
import api from '../services/api';
import ProjectCard from '../components/ProjectCard';

const ProjectList = () => {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedTag, setSelectedTag] = useState('Semua');

  useEffect(() => {
    const fetchProjects = async () => {
      try {
        setLoading(true);
        const res = await api.get('/projects');
        setProjects(res.data.data || []);
      } catch (err) {
        setError(err.response?.data?.error || 'Gagal memuat daftar project.');
      } finally {
        setLoading(false);
      }
    };
    fetchProjects();
  }, []);

  // Kumpulkan semua tag unik dari tech_stack seluruh project
  const allTags = [...new Set(
    projects.flatMap((project) =>
      (project.tech_stack || '').split(',').map((tag) => tag.trim()).filter(Boolean)
    )
  )];

  const filteredProjects = selectedTag === 'Semua'
    ? projects
    : projects.filter((project) =>
        (project.tech_stack || '').split(',').map((tag) => tag.trim()).includes(selectedTag)
      );

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body">
      <div className="max-w-6xl mx-auto px-6 py-12">
        <header className="mb-10">
          <h1 className="font-display text-4xl font-semibold text-[#1E293B]">Project Showcase</h1>
          <p className="text-[#64748B] mt-2">Jelajahi project kolaborasi dari komunitas kampus.</p>
        </header>

        {/* Filter pill berdasarkan tech stack */}
        <div className="flex flex-wrap gap-2 mb-8">
          {['Semua', ...allTags].map((tag) => (
            <button
              key={tag}
              type="button"
              onClick={() => setSelectedTag(tag)}
              className={`px-4 py-2 rounded-full text-sm font-semibold border transition-colors ${
                selectedTag === tag
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
        ) : filteredProjects.length === 0 ? (
          // Empty state
          <div className="text-center py-24">
            <p className="text-5xl mb-4">🗂️</p>
            <p className="font-display text-xl font-semibold text-[#1E293B]">Belum ada project</p>
            <p className="text-sm text-[#64748B] mt-1">Jadilah yang pertama membagikan project kamu!</p>
          </div>
        ) : (
          // Grid kartu project
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredProjects.map((project) => (
              <ProjectCard key={project.id} project={project} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default ProjectList;
