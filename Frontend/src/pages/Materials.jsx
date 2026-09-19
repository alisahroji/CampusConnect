import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ExternalLink, FileText, Pencil, Trash2, UploadCloud } from 'lucide-react';
import api from '../services/api';
import { isLoggedIn, getCurrentUserID } from '../utils/auth';

// ============================================================================
// Materials (Minggu 7 Day 3) — halaman daftar materi kuliah.
//
// Sumber kebenaran:
// - Data 100% dari backend Day 2 (GET/POST/PUT/DELETE /api/materials).
// - Role dibaca dari BACKEND (GET /api/profile, pola AdminDashboard W6 Day 5),
//   bukan dari input client. Form upload hanya dirender untuk Lecturer.
// - Ownership edit/delete dibaca dari uploader.id pada response API vs JWT sub.
//
// PENTING: frontend restriction BUKAN security boundary — backend
// LecturerGuard (upload) + ownership check service (update/delete) tetap
// authority. Student/Admin yang memaksa via API akan ditolak 403.
//
// Kebijakan file di bawah = MIRROR kebijakan backend Day 2 (early validation
// UX saja; backend tetap final validator):
//   Allowed: pdf doc docx ppt pptx xls xlsx txt zip jpg png — Max 20 MB.
// ============================================================================

const ALLOWED_EXTENSIONS = [
  'pdf', 'doc', 'docx', 'ppt', 'pptx', 'xls', 'xlsx', 'txt', 'zip', 'jpg', 'png',
];
const MAX_FILE_SIZE = 20 * 1024 * 1024; // 20 MB

const formatBytes = (bytes) => {
  if (!bytes || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / 1024 ** i).toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
};

const formatDate = (iso) => {
  try {
    return new Date(iso).toLocaleDateString('id-ID', {
      day: 'numeric', month: 'short', year: 'numeric',
    });
  } catch {
    return '';
  }
};

const inputClass = (hasError) =>
  `w-full bg-white border-2 p-3 text-sm text-[#1E293B] rounded-xl placeholder:text-[#94A3B8] focus:outline-none transition-colors ${
    hasError ? 'border-red-400 focus:border-red-500' : 'border-[#E2E8F0] focus:border-[#112320]'
  }`;

const Materials = () => {
  const navigate = useNavigate();

  // ---- List state ----
  const [materials, setMaterials] = useState([]);
  const [categories, setCategories] = useState([]); // derived dari list tanpa filter
  const [categoryFilter, setCategoryFilter] = useState(''); // '' = Semua
  const [refreshKey, setRefreshKey] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // ---- Role (dari backend) ----
  const [role, setRole] = useState('');
  const isLecturer = role === 'Lecturer';

  // ---- Upload form state (Lecturer only) ----
  const [form, setForm] = useState({ title: '', category: '' });
  const [file, setFile] = useState(null);
  const [formErrors, setFormErrors] = useState({});
  const [formError, setFormError] = useState('');
  const [uploadMessage, setUploadMessage] = useState('');
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const fileInputRef = useRef(null);

  // ---- Edit/Delete per-card state ----
  const [editingID, setEditingID] = useState(null);
  const [editForm, setEditForm] = useState({ title: '', category: '' });
  const [actionError, setActionError] = useState('');
  const [busyID, setBusyID] = useState(null); // anti double-click per aksi

  const currentUserId = getCurrentUserID();

  // Role dari backend (sumber kebenaran) — pola yang sama dengan AdminDashboard.
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
        setRole(res.data?.user?.role || '');
      } catch (err) {
        if (cancelled) return;
        if (err.response?.status === 401) {
          navigate('/login');
        }
        // Error lain: role tetap kosong → form upload disembunyikan.
        // Backend LecturerGuard tetap menolak API upload bila sebenarnya Lecturer.
      }
    };
    checkRole();
    return () => {
      cancelled = true;
    };
  }, [navigate]);

  // List material — filter kategori terjadi di BACKEND (?category=), pola
  // ProjectList. Daftar kategori untuk pill diambil dari list tanpa filter.
  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      try {
        setLoading(true);
        setError('');
        const params = { limit: 50 }; // bounded, tanpa pagination palsu
        if (categoryFilter) params.category = categoryFilter;
        const res = await api.get('/materials', { params });
        if (cancelled) return;
        const list = res.data.data || [];
        setMaterials(list);
        if (!categoryFilter) {
          setCategories([...new Set(list.map((m) => m.category).filter(Boolean))].sort());
        }
      } catch (err) {
        if (cancelled) return;
        if (err.response?.status === 401) {
          navigate('/login');
          return;
        }
        setError(err.response?.data?.error || 'Gagal memuat daftar materi.');
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    load();
    return () => {
      cancelled = true;
    };
  }, [categoryFilter, refreshKey, navigate]);

  // ---- Upload ----
  const handleFileChange = (e) => {
    const selected = e.target.files?.[0] || null;
    e.target.value = ''; // reset agar file sama bisa dipilih ulang
    setFile(selected);
    setFormErrors((prev) => ({ ...prev, file: undefined }));
  };

  const validateUpload = () => {
    const errs = {};
    if (!form.title.trim()) errs.title = 'Judul wajib diisi';
    if (!form.category.trim()) errs.category = 'Kategori/mata kuliah wajib diisi';
    if (!file) {
      errs.file = 'File materi wajib dipilih';
    } else {
      const ext = file.name.split('.').pop()?.toLowerCase() || '';
      if (!ALLOWED_EXTENSIONS.includes(ext)) {
        errs.file = `Tipe file .${ext} tidak didukung. Format diizinkan: ${ALLOWED_EXTENSIONS.join(', ')}`;
      } else if (file.size > MAX_FILE_SIZE) {
        errs.file = `Ukuran file maksimal ${formatBytes(MAX_FILE_SIZE)}.`;
      } else if (file.size === 0) {
        errs.file = 'File tidak boleh kosong.';
      }
    }
    setFormErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleUpload = async (e) => {
    e.preventDefault();
    setFormError('');
    setUploadMessage('');
    if (!validateUpload()) return;

    try {
      setUploading(true);
      setUploadProgress(0);
      const fd = new FormData();
      fd.append('title', form.title.trim());
      fd.append('category', form.category.trim());
      fd.append('file', file);
      const res = await api.post('/materials', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
        onUploadProgress: (progressEvent) => {
          if (progressEvent.total) {
            setUploadProgress(Math.round((progressEvent.loaded / progressEvent.total) * 100));
          }
        },
      });
      // Sinkronkan UI dengan data nyata dari response backend (bukan state lokal)
      const created = res.data.data;
      setMaterials((prev) => [created, ...prev]);
      setCategories((prev) =>
        prev.includes(created.category) ? prev : [...prev, created.category].sort()
      );
      setForm({ title: '', category: '' });
      setFile(null);
      if (fileInputRef.current) fileInputRef.current.value = '';
      setUploadMessage('Materi berhasil diunggah.');
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      // Gagal: list tidak diubah, form tidak false-positive
      setUploadMessage('');
      setFormError(err.response?.data?.error || 'Gagal mengunggah materi.');
    } finally {
      setUploading(false);
      setUploadProgress(0);
    }
  };

  // ---- Edit (metadata: judul/kategori) ----
  const startEdit = (material) => {
    setEditingID(material.id);
    setEditForm({ title: material.title, category: material.category });
    setActionError('');
    setUploadMessage('');
  };

  const cancelEdit = () => {
    setEditingID(null);
    setActionError('');
  };

  const handleUpdate = async (material) => {
    if (!editForm.title.trim() || !editForm.category.trim()) {
      setActionError('Judul dan kategori wajib diisi.');
      return;
    }
    try {
      setBusyID(material.id);
      setActionError('');
      const res = await api.put(`/materials/${material.id}`, {
        title: editForm.title.trim(),
        category: editForm.category.trim(),
      });
      // Update UI dari response backend
      const updated = res.data.data;
      setMaterials((prev) => prev.map((m) => (m.id === updated.id ? updated : m)));
      setCategories((prev) =>
        prev.includes(updated.category) ? prev : [...prev, updated.category].sort()
      );
      setEditingID(null);
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setActionError(err.response?.data?.error || 'Gagal memperbarui materi.');
    } finally {
      setBusyID(null);
    }
  };

  // ---- Delete ----
  const handleDelete = async (material) => {
    if (!window.confirm(`Hapus materi "${material.title}" secara permanen?`)) return;
    try {
      setBusyID(material.id);
      setActionError('');
      await api.delete(`/materials/${material.id}`);
      setMaterials((prev) => prev.filter((m) => m.id !== material.id));
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      // Gagal: row tetap ada, tampilkan error (bukan false-positive)
      setActionError(err.response?.data?.error || 'Gagal menghapus materi.');
    } finally {
      setBusyID(null);
    }
  };

  const isFiltered = categoryFilter !== '';

  // Kartu material dirender lewat fungsi (bukan komponen dalam komponen)
  // agar input form edit tidak kehilangan fokus saat re-render.
  const renderMaterialCard = (material) => {
    const isOwner = material.uploader?.id && material.uploader.id === currentUserId;
    const isEditing = editingID === material.id;
    const isBusy = busyID === material.id;

    return (
      <div
        key={material.id}
        className="bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-5 flex flex-col gap-3"
      >
        {isEditing ? (
          /* ---------- Mode edit (metadata saja; ganti file ditunda backend) ---------- */
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-1.5">
                Judul
              </label>
              <input
                type="text"
                value={editForm.title}
                maxLength={150}
                onChange={(e) => setEditForm((f) => ({ ...f, title: e.target.value }))}
                className={inputClass(false)}
              />
            </div>
            <div>
              <label className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-1.5">
                Kategori / Mata Kuliah
              </label>
              <input
                type="text"
                value={editForm.category}
                maxLength={100}
                onChange={(e) => setEditForm((f) => ({ ...f, category: e.target.value }))}
                className={inputClass(false)}
              />
            </div>
            <p className="text-xs text-[#94A3B8]">
              Penggantian file belum didukung — hanya judul dan kategori.
            </p>
            <div className="flex items-center gap-2 pt-1">
              <button
                type="button"
                onClick={() => handleUpdate(material)}
                disabled={isBusy}
                className="bg-[#D97757] hover:bg-[#C26244] disabled:opacity-60 text-white text-xs font-bold uppercase tracking-wider px-4 py-2 rounded-full transition-colors cursor-pointer"
              >
                {isBusy ? 'Menyimpan...' : 'Simpan'}
              </button>
              <button
                type="button"
                onClick={cancelEdit}
                disabled={isBusy}
                className="text-xs font-bold text-[#64748B] hover:text-[#1E293B] px-3 py-2 transition-colors cursor-pointer"
              >
                Batal
              </button>
            </div>
          </div>
        ) : (
          /* ---------- Mode tampil ---------- */
          <>
            <div className="flex items-start gap-3">
              <div className="w-10 h-10 shrink-0 rounded-xl bg-[#FEF3EB] text-[#C26244] flex items-center justify-center">
                <FileText className="w-5 h-5" />
              </div>
              <div className="flex-1 min-w-0">
                <h3
                  className="font-display font-semibold text-[#1E293B] leading-snug break-words"
                  title={material.title}
                >
                  {material.title}
                </h3>
                <span className="inline-block mt-1 px-2.5 py-0.5 rounded-full bg-[#FEF3EB] text-[#C26244] text-[10px] font-bold uppercase tracking-wider">
                  {material.category}
                </span>
              </div>
            </div>

            <div className="flex items-center gap-2 text-sm text-[#64748B] min-w-0">
              <img
                src={
                  material.uploader?.picture_url ||
                  `https://ui-avatars.com/api/?name=${encodeURIComponent(material.uploader?.name || 'A')}`
                }
                alt={material.uploader?.name || 'Uploader'}
                className="w-6 h-6 rounded-full object-cover border border-[#E2E8F0] shrink-0"
              />
              <span className="font-semibold truncate">
                {material.uploader?.name || 'Anonim'}
              </span>
            </div>

            <div className="text-xs text-[#94A3B8] space-y-0.5 min-w-0">
              <p className="truncate" title={material.original_filename}>
                📄 {material.original_filename} · {formatBytes(material.file_size)}
              </p>
              <p>Diunggah {formatDate(material.created_at)}</p>
            </div>

            <div className="flex flex-wrap items-center gap-2 pt-1 border-t border-[#F1F5F9]">
              <a
                href={material.file_url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 text-xs font-bold text-[#D97757] hover:text-[#C26244] hover:underline transition-colors"
              >
                Buka File <ExternalLink className="w-3.5 h-3.5" />
              </a>
              {isOwner && (
                <div className="flex items-center gap-2 ml-auto">
                  <button
                    type="button"
                    onClick={() => startEdit(material)}
                    disabled={isBusy}
                    className="inline-flex items-center gap-1 text-xs font-bold text-[#64748B] hover:text-[#112320] transition-colors cursor-pointer disabled:opacity-60"
                  >
                    <Pencil className="w-3.5 h-3.5" /> Edit
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(material)}
                    disabled={isBusy}
                    className="inline-flex items-center gap-1 text-xs font-bold text-red-600 hover:text-red-700 transition-colors cursor-pointer disabled:opacity-60"
                  >
                    <Trash2 className="w-3.5 h-3.5" /> {isBusy ? 'Menghapus...' : 'Hapus'}
                  </button>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body selection:bg-[#D97757] selection:text-white">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 py-12">
        <header className="mb-10">
          <h1 className="font-display text-4xl font-semibold text-[#1E293B]">Materials</h1>
          <p className="text-[#64748B] mt-2">
            Materi kuliah dan bahan belajar dari dosen.
          </p>
        </header>

        {/* ---------- Form upload — HANYA Lecturer (backend LecturerGuard authority) ---------- */}
        {isLecturer && (
          <form
            onSubmit={handleUpload}
            className="bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-6 mb-10 space-y-5"
          >
            <div className="flex items-center gap-2">
              <UploadCloud className="w-5 h-5 text-[#D97757]" />
              <h2 className="font-display text-lg font-semibold text-[#1E293B]">
                Unggah Materi Baru
              </h2>
            </div>

            {formError && (
              <div className="p-4 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
                {formError}
              </div>
            )}
            {uploadMessage && (
              <div className="p-4 bg-green-50 border-l-4 border-green-500 text-green-700 text-sm font-medium rounded">
                {uploadMessage}
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
              <div>
                <label htmlFor="material-title" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
                  Judul Materi *
                </label>
                <input
                  id="material-title"
                  type="text"
                  value={form.title}
                  maxLength={150}
                  onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
                  placeholder="cth: Slide Pertemuan 1 — Pengenalan Algoritma"
                  className={inputClass(formErrors.title)}
                />
                {formErrors.title && <p className="text-xs text-red-500 mt-1">{formErrors.title}</p>}
              </div>
              <div>
                <label htmlFor="material-category" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
                  Kategori / Mata Kuliah *
                </label>
                <input
                  id="material-category"
                  type="text"
                  value={form.category}
                  maxLength={100}
                  onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
                  placeholder="cth: Algoritma dan Pemrograman"
                  className={inputClass(formErrors.category)}
                />
                {formErrors.category && (
                  <p className="text-xs text-red-500 mt-1">{formErrors.category}</p>
                )}
              </div>
            </div>

            <div>
              <label className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
                File Materi *
              </label>
              <label
                htmlFor="material-file"
                className={`inline-flex items-center gap-2 border-2 border-dashed rounded-xl px-5 py-3 text-sm font-semibold transition-colors cursor-pointer max-w-full ${
                  formErrors.file
                    ? 'border-red-300 text-red-500'
                    : 'border-[#E2E8F0] text-[#64748B] hover:border-[#D97757] hover:text-[#D97757]'
                } ${uploading ? 'opacity-60 pointer-events-none' : ''}`}
              >
                <FileText className="w-4 h-4 shrink-0" />
                <span className="truncate max-w-[280px] sm:max-w-md" title={file?.name}>
                  {file ? `${file.name} (${formatBytes(file.size)})` : 'Pilih file...'}
                </span>
              </label>
              <input
                id="material-file"
                ref={fileInputRef}
                type="file"
                onChange={handleFileChange}
                className="hidden"
              />
              {formErrors.file && <p className="text-xs text-red-500 mt-1">{formErrors.file}</p>}
              <p className="text-xs text-[#94A3B8] mt-1.5">
                Format: {ALLOWED_EXTENSIONS.join(', ')} · Maksimal {formatBytes(MAX_FILE_SIZE)}.
              </p>
            </div>

            <div className="pt-2">
              <button
                type="submit"
                disabled={uploading}
                className="bg-[#D97757] hover:bg-[#C26244] disabled:bg-[#E2E8F0] disabled:text-[#94A3B8] text-white px-6 py-2.5 text-sm font-bold tracking-wider uppercase transition-colors rounded-full cursor-pointer"
              >
                {uploading
                  ? `Mengunggah...${uploadProgress > 0 ? ` ${uploadProgress}%` : ''}`
                  : 'Unggah Materi'}
              </button>
            </div>
          </form>
        )}

        {/* ---------- Filter kategori (query dikirim ke backend) ---------- */}
        {categories.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-8">
            {['Semua', ...categories].map((cat) => (
              <button
                key={cat}
                type="button"
                onClick={() => setCategoryFilter(cat === 'Semua' ? '' : cat)}
                className={`px-4 py-2 rounded-full text-sm font-semibold border transition-colors ${
                  categoryFilter === (cat === 'Semua' ? '' : cat)
                    ? 'bg-[#112320] text-[#F8F9FA] border-[#112320]'
                    : 'bg-white text-[#64748B] border-[#E2E8F0] hover:border-[#112320]'
                }`}
              >
                {cat}
              </button>
            ))}
          </div>
        )}

        {/* ---------- Error aksi (edit/delete gagal) ---------- */}
        {actionError && (
          <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
            {actionError}
          </div>
        )}

        {/* ---------- List ---------- */}
        {loading ? (
          <div className="flex items-center justify-center py-24">
            <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
          </div>
        ) : error ? (
          <div className="p-6 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded flex flex-wrap items-center justify-between gap-3">
            <span>{error}</span>
            <button
              type="button"
              onClick={() => setRefreshKey((k) => k + 1)}
              className="bg-[#112320] text-[#F8F9FA] text-xs font-bold uppercase tracking-wider px-4 py-2 rounded-full hover:bg-[#1E293B] transition-colors cursor-pointer"
            >
              ↻ Muat ulang
            </button>
          </div>
        ) : materials.length === 0 ? (
          <div className="text-center py-24">
            <p className="text-5xl mb-4">📚</p>
            <p className="font-display text-xl font-semibold text-[#1E293B]">
              {isFiltered ? 'Tidak ada materi yang cocok' : 'Belum ada materi'}
            </p>
            <p className="text-sm text-[#64748B] mt-1">
              {isFiltered
                ? 'Coba pilih kategori lain.'
                : isLecturer
                  ? 'Unggah materi pertama melalui form di atas.'
                  : 'Materi dari dosen akan tampil di sini.'}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
            {materials.map(renderMaterialCard)}
          </div>
        )}
      </div>
    </div>
  );
};

export default Materials;
