import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import api from '../services/api';

const ProjectDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const [project, setProject] = useState(null);
  const [comments, setComments] = useState([]);
  const [likeCount, setLikeCount] = useState(0);
  const [liked, setLiked] = useState(false);
  const [newComment, setNewComment] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [likeBusy, setLikeBusy] = useState(false);

  const isLoggedIn = () => Boolean(localStorage.getItem('access_token'));

  const techStack = (project?.tech_stack || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean);

  useEffect(() => {
    const fetchDetail = async () => {
      try {
        setLoading(true);
        const [projectRes, commentsRes, likesRes] = await Promise.all([
          api.get(`/projects/${id}`),
          api.get(`/projects/${id}/comments`),
          api.get(`/projects/${id}/likes`),
        ]);
        setProject(projectRes.data.data || null);
        setComments(commentsRes.data.data || []);
        setLikeCount(likesRes.data.like_count ?? 0);
      } catch (err) {
        setError(
          err.response?.status === 404
            ? 'Project tidak ditemukan.'
            : err.response?.data?.error || 'Gagal memuat detail project.'
        );
      } finally {
        setLoading(false);
      }
    };
    fetchDetail();
  }, [id]);

  const handleLike = async () => {
    if (!isLoggedIn()) {
      navigate('/login');
      return;
    }
    try {
      setLikeBusy(true);
      const res = await api.post(`/projects/${id}/like`);
      // Backend mengembalikan status like terbaru: { liked, like_count }
      setLiked(res.data.liked);
      setLikeCount(res.data.like_count);
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setError(err.response?.data?.error || 'Gagal memperbarui like.');
    } finally {
      setLikeBusy(false);
    }
  };

  const handleCommentSubmit = async (e) => {
    e.preventDefault();
    if (!newComment.trim()) return;

    if (!isLoggedIn()) {
      navigate('/login');
      return;
    }

    try {
      setSubmitting(true);
      const res = await api.post(`/projects/${id}/comments`, {
        content: newComment.trim(),
      });
      // Tampilkan komentar baru secara instan tanpa refetch
      setComments((prev) => [...prev, res.data.data]);
      setNewComment('');
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setError(err.response?.data?.error || 'Gagal mengirim komentar.');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center font-body">
        <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error && !project) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex flex-col items-center justify-center font-body px-6 text-center">
        <p className="text-5xl mb-4">🔍</p>
        <p className="font-display text-xl font-semibold text-[#1E293B]">{error}</p>
        <Link
          to="/projects"
          className="mt-6 inline-block bg-[#112320] text-[#F8F9FA] font-bold text-sm uppercase tracking-wider py-3 px-6 rounded-full hover:bg-[#1E293B] transition-colors"
        >
          ← Kembali ke Daftar Project
        </Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body selection:bg-[#D97757] selection:text-white">
      {/* Banner gambar project */}
      <div className="h-64 md:h-80 bg-gradient-to-br from-[#112320] to-[#1E293B] flex items-center justify-center overflow-hidden">
        {project?.image_url ? (
          <img
            src={project.image_url}
            alt={project?.title}
            className="w-full h-full object-cover"
          />
        ) : (
          <span className="font-display text-7xl font-semibold text-[#D97757]">
            {project?.title?.charAt(0)?.toUpperCase() ?? 'P'}
          </span>
        )}
      </div>

      <div className="max-w-4xl mx-auto px-6 -mt-12 relative z-10 pb-20">
        <Link
          to="/projects"
          className="inline-block mb-6 text-[#64748B] text-sm font-bold tracking-wider uppercase hover:text-[#112320] transition-colors"
        >
          ← Kembali ke Daftar
        </Link>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
            {error}
          </div>
        )}

        {/* Kartu utama detail project */}
        <article className="bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-6 md:p-10">
          <div className="flex flex-wrap items-start justify-between gap-4 mb-2">
            <h1 className="font-display text-3xl md:text-4xl font-semibold text-[#1E293B] leading-tight">
              {project?.title}
            </h1>
            <span className="shrink-0 text-[10px] font-bold uppercase tracking-wider text-[#94A3B8] border border-[#E2E8F0] rounded-full px-3 py-1">
              {project?.status || 'published'}
            </span>
          </div>

          {/* Penulis project */}
          <div className="flex items-center gap-3 mb-8">
            <img
              src={
                project?.user?.picture_url ||
                `https://ui-avatars.com/api/?name=${project?.user?.name || 'A'}`
              }
              alt={project?.user?.name}
              className="w-10 h-10 rounded-full object-cover border border-[#E2E8F0]"
            />
            <div>
              <p className="text-sm font-bold text-[#1E293B]">
                {project?.user?.name || 'Anonim'}
              </p>
              <p className="text-xs text-[#94A3B8]">
                {project?.created_at
                  ? new Date(project.created_at).toLocaleDateString('id-ID', {
                      day: 'numeric',
                      month: 'long',
                      year: 'numeric',
                    })
                  : ''}
              </p>
            </div>
          </div>

          <p className="text-[#1E293B] leading-relaxed whitespace-pre-line mb-8">
            {project?.description}
          </p>

          {/* Pills tech stack */}
          {techStack.length > 0 && (
            <div className="flex flex-wrap gap-2 mb-8">
              {techStack.map((tag) => (
                <span
                  key={tag}
                  className="text-xs font-semibold text-[#112320] bg-[#D97757]/10 border border-[#D97757]/20 rounded-full px-3 py-1"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}

          {/* Aksi: Like + link demo/repo */}
          <div className="flex flex-wrap items-center gap-4 pt-6 border-t border-[#E2E8F0]">
            <button
              type="button"
              onClick={handleLike}
              disabled={likeBusy}
              className={`flex items-center gap-2 font-bold text-sm py-2.5 px-5 rounded-full border transition-all duration-300 disabled:opacity-60 cursor-pointer ${
                liked
                  ? 'bg-[#D97757] border-[#D97757] text-white hover:bg-[#C26244]'
                  : 'bg-white border-[#E2E8F0] text-[#1E293B] hover:border-[#D97757] hover:text-[#D97757]'
              }`}
            >
              <span className="text-base leading-none">{liked ? '♥' : '♡'}</span>
              {likeCount} {likeCount === 1 ? 'Like' : 'Likes'}
            </button>

            {project?.demo_url && (
              <a
                href={project.demo_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm font-bold text-[#D97757] hover:text-[#C26244] transition-colors"
              >
                Lihat Demo ↗
              </a>
            )}
            {project?.repo_url && (
              <a
                href={project.repo_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm font-bold text-[#64748B] hover:text-[#1E293B] transition-colors"
              >
                Kode Sumber ↗
              </a>
            )}
          </div>
        </article>

        {/* Section komentar */}
        <section className="mt-10">
          <h2 className="font-display text-2xl font-semibold text-[#1E293B] mb-6">
            Diskusi ({comments.length})
          </h2>

          {/* Form komentar */}
          <form
            onSubmit={handleCommentSubmit}
            className="bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-6 mb-8"
          >
            <textarea
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              rows="3"
              placeholder={
                isLoggedIn()
                  ? 'Tulis komentar atau pertanyaanmu...'
                  : 'Login terlebih dahulu untuk menulis komentar...'
              }
              className="w-full bg-transparent border-2 border-[#E2E8F0] p-4 text-sm text-[#1E293B] focus:border-[#112320] focus:outline-none transition-colors rounded-xl placeholder:text-[#94A3B8] resize-none"
            />
            <div className="flex justify-end mt-4">
              <button
                type="submit"
                disabled={submitting || !newComment.trim()}
                className="bg-[#D97757] hover:bg-[#C26244] disabled:bg-[#E2E8F0] disabled:text-[#94A3B8] text-white px-6 py-3 text-sm font-bold tracking-wider uppercase transition-colors rounded-full cursor-pointer"
              >
                {submitting ? 'Mengirim...' : 'Kirim Komentar'}
              </button>
            </div>
          </form>

          {/* Daftar komentar */}
          {comments.length === 0 ? (
            <div className="text-center py-12 bg-white rounded-2xl border border-[#E2E8F0]">
              <p className="text-4xl mb-3">💬</p>
              <p className="font-display font-semibold text-[#1E293B]">
                Belum ada komentar
              </p>
              <p className="text-sm text-[#64748B] mt-1">
                Jadilah yang pertama memulai diskusi!
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {comments.map((comment) => (
                <div
                  key={comment.id}
                  className="bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-5"
                >
                  <div className="flex items-center gap-3 mb-3">
                    <img
                      src={
                        comment.user?.picture_url ||
                        `https://ui-avatars.com/api/?name=${comment.user?.name || 'A'}`
                      }
                      alt={comment.user?.name}
                      className="w-9 h-9 rounded-full object-cover border border-[#E2E8F0]"
                    />
                    <div>
                      <p className="text-sm font-bold text-[#1E293B]">
                        {comment.user?.name || 'Anonim'}
                      </p>
                      <p className="text-xs text-[#94A3B8]">
                        {comment.created_at
                          ? new Date(comment.created_at).toLocaleDateString('id-ID', {
                              day: 'numeric',
                              month: 'short',
                              year: 'numeric',
                            })
                          : ''}
                      </p>
                    </div>
                  </div>
                  <p className="text-sm text-[#1E293B] leading-relaxed whitespace-pre-line">
                    {comment.content}
                  </p>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  );
};

export default ProjectDetail;
