import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';

// Format waktu relatif sederhana (mis. "5 jam lalu")
const formatRelativeTime = (isoDate) => {
  if (!isoDate) return '';
  const diffMs = Date.now() - new Date(isoDate).getTime();
  const minutes = Math.floor(diffMs / 60000);
  if (minutes < 1) return 'Baru saja';
  if (minutes < 60) return `${minutes} menit lalu`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} jam lalu`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days} hari lalu`;
  return new Date(isoDate).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
};

const PostCard = ({ post }) => {
  const navigate = useNavigate();

  const [liked, setLiked] = useState(false);
  const [likeCount, setLikeCount] = useState(post.like_count ?? 0);
  const [likeBusy, setLikeBusy] = useState(false);

  const [showComments, setShowComments] = useState(false);
  const [comments, setComments] = useState([]);
  const [commentsLoaded, setCommentsLoaded] = useState(false);
  const [commentsLoading, setCommentsLoading] = useState(false);
  const [newComment, setNewComment] = useState('');
  const [commentBusy, setCommentBusy] = useState(false);
  const [cardError, setCardError] = useState('');

  const isLoggedIn = () => Boolean(localStorage.getItem('access_token'));

  const handleLike = async () => {
    if (!isLoggedIn()) {
      navigate('/login');
      return;
    }
    try {
      setLikeBusy(true);
      setCardError('');
      const res = await api.post(`/posts/${post.id}/like`);
      setLiked(res.data.liked);
      setLikeCount(res.data.like_count);
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setCardError(err.response?.data?.error || 'Gagal memperbarui like');
    } finally {
      setLikeBusy(false);
    }
  };

  const toggleComments = async () => {
    const next = !showComments;
    setShowComments(next);
    if (next && !commentsLoaded) {
      try {
        setCommentsLoading(true);
        const res = await api.get(`/posts/${post.id}/comments`);
        setComments(res.data.data || []);
        setCommentsLoaded(true);
      } catch (err) {
        setCardError(err.response?.data?.error || 'Gagal memuat komentar');
      } finally {
        setCommentsLoading(false);
      }
    }
  };

  const handleCommentSubmit = async (e) => {
    e.preventDefault();
    const content = newComment.trim();
    if (!content) return;

    if (!isLoggedIn()) {
      navigate('/login');
      return;
    }

    try {
      setCommentBusy(true);
      setCardError('');
      const res = await api.post(`/posts/${post.id}/comments`, { content });
      setComments((prev) => [...prev, res.data.data]);
      setNewComment('');
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setCardError(err.response?.data?.error || 'Gagal mengirim komentar');
    } finally {
      setCommentBusy(false);
    }
  };

  return (
    <article className="bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-5 md:p-6">
      {/* Header penulis */}
      <div className="flex items-center gap-3 mb-4">
        <img
          src={
            post.user?.picture_url ||
            `https://ui-avatars.com/api/?name=${post.user?.name || 'A'}`
          }
          alt={post.user?.name}
          className="w-10 h-10 rounded-full object-cover border border-[#E2E8F0]"
        />
        <div className="flex-1">
          <p className="text-sm font-bold text-[#1E293B]">
            {post.user?.name || 'Anonim'}
          </p>
          <p className="text-xs text-[#94A3B8]">
            {formatRelativeTime(post.created_at)}
          </p>
        </div>
      </div>

      {/* Isi post */}
      <p className="text-[#1E293B] leading-relaxed whitespace-pre-line mb-4">
        {post.content}
      </p>

      {/* Gambar post (opsional) */}
      {post.image_url && (
        <img
          src={post.image_url}
          alt="Lampiran post"
          className="w-full rounded-xl border border-[#E2E8F0] mb-4"
        />
      )}

      {cardError && (
        <p className="text-xs text-red-600 mb-3">{cardError}</p>
      )}

      {/* Aksi */}
      <div className="flex items-center gap-3 pt-3 border-t border-[#E2E8F0]">
        <button
          type="button"
          onClick={handleLike}
          disabled={likeBusy}
          className={`flex items-center gap-2 text-sm font-bold py-2 px-4 rounded-full border transition-all duration-300 disabled:opacity-60 cursor-pointer ${
            liked
              ? 'bg-[#D97757] border-[#D97757] text-white hover:bg-[#C26244]'
              : 'bg-white border-[#E2E8F0] text-[#1E293B] hover:border-[#D97757] hover:text-[#D97757]'
          }`}
        >
          <span className="leading-none">{liked ? '♥' : '♡'}</span>
          {likeCount}
        </button>

        <button
          type="button"
          onClick={toggleComments}
          className="flex items-center gap-2 text-sm font-bold py-2 px-4 rounded-full border border-[#E2E8F0] bg-white text-[#1E293B] hover:border-[#112320] transition-colors cursor-pointer"
        >
          💬 Komentar
          {commentsLoaded && comments.length > 0 && (
            <span className="text-xs text-[#94A3B8]">({comments.length})</span>
          )}
        </button>
      </div>

      {/* Section komentar (inline expand) */}
      {showComments && (
        <div className="mt-4 pt-4 border-t border-[#E2E8F0] space-y-4">
          {commentsLoading ? (
            <div className="flex justify-center py-4">
              <div className="w-6 h-6 border-2 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
            </div>
          ) : comments.length === 0 ? (
            <p className="text-sm text-[#94A3B8] text-center py-2">
              Belum ada komentar. Mulai diskusinya!
            </p>
          ) : (
            comments.map((comment) => (
              <div key={comment.id} className="flex items-start gap-3">
                <img
                  src={
                    comment.user?.picture_url ||
                    `https://ui-avatars.com/api/?name=${comment.user?.name || 'A'}`
                  }
                  alt={comment.user?.name}
                  className="w-8 h-8 rounded-full object-cover border border-[#E2E8F0]"
                />
                <div className="bg-[#F8F9FA] rounded-xl px-4 py-3 flex-1">
                  <p className="text-xs font-bold text-[#1E293B]">
                    {comment.user?.name || 'Anonim'}
                  </p>
                  <p className="text-sm text-[#1E293B] leading-relaxed">
                    {comment.content}
                  </p>
                </div>
              </div>
            ))
          )}

          {/* Form komentar */}
          <form onSubmit={handleCommentSubmit} className="flex items-center gap-2">
            <input
              type="text"
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              placeholder={
                isLoggedIn()
                  ? 'Tulis komentar...'
                  : 'Login untuk berkomentar...'
              }
              className="flex-1 bg-[#F8F9FA] border border-[#E2E8F0] px-4 py-2.5 text-sm text-[#1E293B] focus:border-[#112320] focus:outline-none transition-colors rounded-full placeholder:text-[#94A3B8]"
            />
            <button
              type="submit"
              disabled={commentBusy || !newComment.trim()}
              className="bg-[#D97757] hover:bg-[#C26244] disabled:bg-[#E2E8F0] disabled:text-[#94A3B8] text-white text-sm font-bold py-2.5 px-5 rounded-full transition-colors cursor-pointer"
            >
              Kirim
            </button>
          </form>
        </div>
      )}
    </article>
  );
};

export default PostCard;
