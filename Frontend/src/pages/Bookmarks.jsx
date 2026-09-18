import { useEffect, useState } from 'react';
import { BookmarkCheck, Link as LinkIcon } from 'lucide-react';
import { Link } from 'react-router-dom';
import api from '../services/api';

// Halaman Bookmarks (Minggu 6 Day 3): daftar project & post yang disimpan
// user. Data 100% dari backend (endpoint Day 1) — list post sudah ter-hydrate
// author/like_count/liked_by_me, list project berisi data penuh project.
const Bookmarks = () => {
  const [projects, setProjects] = useState([]);
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;

    const fetchBookmarks = async () => {
      try {
        setLoading(true);
        setError('');
        const [projRes, postRes] = await Promise.all([
          api.get('/bookmarks/projects'),
          api.get('/bookmarks/posts'),
        ]);
        if (cancelled) return;
        setProjects(projRes.data.data || []);
        setPosts(postRes.data.data || []);
      } catch (err) {
        if (!cancelled) {
          setError(err.response?.data?.error || 'Gagal memuat bookmark.');
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    fetchBookmarks();
    return () => {
      cancelled = true;
    };
  }, []);

  // Kartu ringkas untuk post yang di-bookmark
  const PostRow = ({ post }) => (
    <Link
      to="/feed"
      className="block bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-5 hover:shadow-lg hover:-translate-y-0.5 transition-all"
    >
      <div className="flex items-center gap-3 mb-2">
        <img
          src={
            post.user?.picture_url ||
            `https://ui-avatars.com/api/?name=${post.user?.name || 'A'}`
          }
          alt={post.user?.name}
          className="w-8 h-8 rounded-full object-cover border border-[#E2E8F0]"
        />
        <span className="text-sm font-bold text-[#1E293B]">{post.user?.name || 'Anonim'}</span>
        <span className="ml-auto flex items-center gap-1 text-xs text-[#94A3B8]">
          ♥ {post.like_count ?? 0}
        </span>
      </div>
      <p className="text-sm text-[#1E293B] leading-relaxed line-clamp-2 whitespace-pre-line">
        {post.content}
      </p>
    </Link>
  );

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 py-12">
        <header className="mb-10">
          <h1 className="font-display text-4xl font-semibold text-[#1E293B]">Bookmarks</h1>
          <p className="text-[#64748B] mt-2">Project dan post yang kamu simpan.</p>
        </header>

        {loading ? (
          <div className="flex items-center justify-center py-24">
            <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
          </div>
        ) : error ? (
          <div className="p-6 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
            {error}
          </div>
        ) : (
          <div className="space-y-12">
            {/* Section project */}
            <section>
              <h2 className="font-display text-2xl font-semibold text-[#1E293B] mb-5">
                Project ({projects.length})
              </h2>
              {projects.length === 0 ? (
                <div className="text-center py-10 bg-white rounded-2xl border border-[#E2E8F0]">
                  <p className="text-4xl mb-3">🗂️</p>
                  <p className="font-display font-semibold text-[#1E293B]">
                    Belum ada project tersimpan
                  </p>
                  <p className="text-sm text-[#64748B] mt-1">
                    Buka{' '}
                    <Link to="/projects" className="text-[#D97757] font-bold hover:underline">
                      Project Showcase
                    </Link>{' '}
                    dan simpan yang menarik.
                  </p>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  {projects.map((project) => (
                    <div
                      key={project.id}
                      className="flex items-start gap-4 bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-5"
                    >
                      <div className="w-12 h-12 shrink-0 rounded-xl bg-gradient-to-br from-[#112320] to-[#1E293B] flex items-center justify-center text-[#D97757] font-display font-semibold text-xl">
                        {project.title?.charAt(0)?.toUpperCase() ?? 'P'}
                      </div>
                      <div className="flex-1 min-w-0">
                        <Link
                          to={`/projects/${project.id}`}
                          className="font-display font-semibold text-[#1E293B] hover:text-[#D97757] transition-colors line-clamp-1"
                        >
                          {project.title}
                        </Link>
                        <p className="text-sm text-[#64748B] line-clamp-2 mt-1">
                          {project.description}
                        </p>
                        <span className="inline-flex items-center gap-1 mt-2 text-[10px] font-bold uppercase tracking-wider text-[#94A3B8]">
                          <BookmarkCheck className="w-3 h-3" /> tersimpan
                        </span>
                      </div>
                      <LinkIcon className="w-4 h-4 text-[#E2E8F0] shrink-0 mt-1" />
                    </div>
                  ))}
                </div>
              )}
            </section>

            {/* Section post */}
            <section>
              <h2 className="font-display text-2xl font-semibold text-[#1E293B] mb-5">
                Post ({posts.length})
              </h2>
              {posts.length === 0 ? (
                <div className="text-center py-10 bg-white rounded-2xl border border-[#E2E8F0]">
                  <p className="text-4xl mb-3">🔖</p>
                  <p className="font-display font-semibold text-[#1E293B]">
                    Belum ada post tersimpan
                  </p>
                  <p className="text-sm text-[#64748B] mt-1">
                    Bookmark post dari{' '}
                    <Link to="/feed" className="text-[#D97757] font-bold hover:underline">
                      Feed
                    </Link>{' '}
                    untuk dibaca nanti.
                  </p>
                </div>
              ) : (
                <div className="space-y-4">
                  {posts.map((post) => (
                    <PostRow key={post.id} post={post} />
                  ))}
                </div>
              )}
            </section>
          </div>
        )}
      </div>
    </div>
  );
};

export default Bookmarks;
