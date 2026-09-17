import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import PostCard from '../components/PostCard';

const PAGE_SIZE = 5;

const Feed = () => {
  const navigate = useNavigate();

  // Pengunjung anonim diarahkan ke tab publik agar tidak langsung dilempar ke login
  const [tab, setTab] = useState(
    () => (localStorage.getItem('access_token') ? 'feed' : 'explore')
  );
  const [posts, setPosts] = useState([]);
  const [nextCursor, setNextCursor] = useState('');
  const [initialLoading, setInitialLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');

  // Composer
  const [newPost, setNewPost] = useState('');
  const [posting, setPosting] = useState(false);
  const [composerError, setComposerError] = useState('');

  const isLoggedIn = () => Boolean(localStorage.getItem('access_token'));

  // ---- Data fetching (cursor pagination) ----
  const fetchPosts = useCallback(
    async (cursor, tabToLoad) => {
      const endpoint = tabToLoad === 'feed' ? '/feed' : '/posts';
      const params = { limit: PAGE_SIZE };
      if (cursor) params.cursor = cursor;

      const res = await api.get(endpoint, { params });
      return {
        items: res.data.data || [],
        nextCursor: res.data.next_cursor || '',
      };
    },
    []
  );

  // Muat halaman pertama setiap kali tab berubah
  useEffect(() => {
    let cancelled = false;

    const loadFirstPage = async () => {
      try {
        setInitialLoading(true);
        setError('');
        setPosts([]);
        setNextCursor('');

        const { items, nextCursor: nc } = await fetchPosts('', tab);
        if (cancelled) return;
        setPosts(items);
        setNextCursor(nc);
      } catch (err) {
        if (cancelled) return;
        if (err.response?.status === 401) {
          navigate('/login');
          return;
        }
        setError(err.response?.data?.error || 'Gagal memuat feed.');
      } finally {
        if (!cancelled) setInitialLoading(false);
      }
    };

    loadFirstPage();
    return () => {
      cancelled = true;
    };
  }, [tab, fetchPosts, navigate]);

  // Muat halaman berikutnya (dipanggil IntersectionObserver)
  const loadMore = useCallback(async () => {
    if (!nextCursor || loadingMore || initialLoading) return;

    try {
      setLoadingMore(true);
      const { items, nextCursor: nc } = await fetchPosts(nextCursor, tab);
      setPosts((prev) => [...prev, ...items]);
      setNextCursor(nc);
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setError(err.response?.data?.error || 'Gagal memuat post berikutnya.');
    } finally {
      setLoadingMore(false);
    }
  }, [nextCursor, loadingMore, initialLoading, tab, fetchPosts, navigate]);

  // ---- Infinite scroll via IntersectionObserver ----
  const sentinelRef = useRef(null);

  useEffect(() => {
    const node = sentinelRef.current;
    if (!node) return undefined;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          loadMore();
        }
      },
      { rootMargin: '200px' } // mulai memuat sedikit sebelum mencapai dasar
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, [loadMore]);

  // ---- Composer: buat post baru ----
  const handlePostSubmit = async (e) => {
    e.preventDefault();
    const content = newPost.trim();
    if (!content) return;

    if (!isLoggedIn()) {
      navigate('/login');
      return;
    }

    try {
      setPosting(true);
      setComposerError('');
      const res = await api.post('/posts', { content });
      // Optimistic prepend: tampilkan post baru di atas tanpa refetch
      setPosts((prev) => [res.data.data, ...prev]);
      setNewPost('');
      // Post baru muncul di tab mana pun karena dibuat oleh user sendiri;
      // kalau user sedang di tab "feed", feed server akan memuatnya saat reload.
      if (tab === 'explore') {
        setTab('feed');
      }
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setComposerError(err.response?.data?.error || 'Gagal membuat post.');
    } finally {
      setPosting(false);
    }
  };

  const switchTab = (target) => {
    if (target === 'feed' && !isLoggedIn()) {
      navigate('/login');
      return;
    }
    setTab(target);
  };

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body selection:bg-[#D97757] selection:text-white">
      <div className="max-w-2xl mx-auto px-4 sm:px-6 py-10">
        <header className="mb-8">
          <h1 className="font-display text-3xl md:text-4xl font-semibold text-[#1E293B]">
            Feed
          </h1>
          <p className="text-[#64748B] mt-1">
            Cerita dan kabar terbaru dari komunitas kampus.
          </p>
        </header>

        {/* Tab: Feed Saya vs Eksplorasi */}
        <div className="flex gap-2 mb-6 bg-white border border-[#E2E8F0] rounded-full p-1.5 shadow-sm">
          <button
            type="button"
            onClick={() => switchTab('feed')}
            className={`flex-1 text-sm font-bold py-2.5 rounded-full transition-all duration-300 cursor-pointer ${
              tab === 'feed'
                ? 'bg-[#112320] text-[#F8F9FA]'
                : 'text-[#64748B] hover:text-[#112320]'
            }`}
          >
            Feed Saya
          </button>
          <button
            type="button"
            onClick={() => switchTab('explore')}
            className={`flex-1 text-sm font-bold py-2.5 rounded-full transition-all duration-300 cursor-pointer ${
              tab === 'explore'
                ? 'bg-[#112320] text-[#F8F9FA]'
                : 'text-[#64748B] hover:text-[#112320]'
            }`}
          >
            Eksplorasi
          </button>
        </div>

        {/* Composer */}
        <form
          onSubmit={handlePostSubmit}
          className="bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-5 mb-6"
        >
          <textarea
            value={newPost}
            onChange={(e) => setNewPost(e.target.value)}
            rows="3"
            maxLength={2000}
            placeholder={
              isLoggedIn()
                ? 'Apa yang sedang kamu kerjakan?'
                : 'Login untuk berbagi cerita...'
            }
            className="w-full bg-transparent text-[#1E293B] text-sm leading-relaxed focus:outline-none resize-none placeholder:text-[#94A3B8]"
          />
          {composerError && (
            <p className="text-xs text-red-600 mb-2">{composerError}</p>
          )}
          <div className="flex items-center justify-between pt-3 border-t border-[#E2E8F0]">
            <span className="text-xs text-[#94A3B8]">
              {newPost.length}/2000
            </span>
            <button
              type="submit"
              disabled={posting || !newPost.trim()}
              className="bg-[#D97757] hover:bg-[#C26244] disabled:bg-[#E2E8F0] disabled:text-[#94A3B8] text-white text-sm font-bold tracking-wider uppercase py-2.5 px-6 rounded-full transition-colors cursor-pointer"
            >
              {posting ? 'Mengirim...' : 'Posting'}
            </button>
          </div>
        </form>

        {/* Error global */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
            {error}
          </div>
        )}

        {/* Daftar post */}
        {initialLoading ? (
          <div className="flex items-center justify-center py-24">
            <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
          </div>
        ) : posts.length === 0 ? (
          <div className="text-center py-20 bg-white rounded-2xl border border-[#E2E8F0]">
            <p className="text-5xl mb-4">🌱</p>
            <p className="font-display text-xl font-semibold text-[#1E293B]">
              {tab === 'feed'
                ? 'Feed kamu masih kosong'
                : 'Belum ada post sama sekali'}
            </p>
            <p className="text-sm text-[#64748B] mt-2 max-w-sm mx-auto">
              {tab === 'feed'
                ? 'Follow orang lain lewat halaman komunitas, atau tulis post pertamamu di atas!'
                : 'Jadilah yang pertama berbagi cerita!'}
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {posts.map((post) => (
              <PostCard key={post.id} post={post} />
            ))}
          </div>
        )}

        {/* Sentinel infinite scroll: terpasang selama masih ada next_cursor.
            loadMore() punya guard (loadingMore/nextCursor) sehingga aman
            dari pemicu ganda saat observasi terjadi berulang. */}
        {!initialLoading && posts.length > 0 && nextCursor && (
          <div ref={sentinelRef} className="h-1" aria-hidden="true" />
        )}
        {!initialLoading && posts.length > 0 && (
          <div className="py-8 flex items-center justify-center">
            {loadingMore ? (
              <div className="w-7 h-7 border-2 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
            ) : nextCursor ? (
              <p className="text-xs text-[#94A3B8]">Gulir untuk memuat lagi…</p>
            ) : (
              <p className="text-xs text-[#94A3B8]">
                Semua post sudah ditampilkan.
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default Feed;
