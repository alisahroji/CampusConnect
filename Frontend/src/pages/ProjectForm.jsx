import { useEffect, useState } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import api from '../services/api';
import { isLoggedIn } from '../utils/auth';

const ProjectForm = () => {
  const { id } = useParams(); // Ada ID = mode edit
  const isEdit = Boolean(id);
  const navigate = useNavigate();

  const [form, setForm] = useState({
    title: '',
    description: '',
    tech_stack: '',
    repo_url: '',
    demo_url: '',
    image_url: '',
    status: 'published',
  });
  const [gallery, setGallery] = useState([]); // array of URL string
  const [statuses, setStatuses] = useState(['draft', 'published']);

  const [loading, setLoading] = useState(isEdit);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [fieldErrors, setFieldErrors] = useState({});

  const setField = (key) => (e) => setForm((f) => ({ ...f, [key]: e.target.value }));

  useEffect(() => {
    // Form project wajib login
    if (!isLoggedIn()) {
      navigate('/login');
      return;
    }

    // Ambil daftar status valid dari backend (fallback draft/published)
    api
      .get('/projects/statuses')
      .then((res) => {
        if (Array.isArray(res.data.statuses) && res.data.statuses.length > 0) {
          setStatuses(res.data.statuses);
        }
      })
      .catch(() => {});

    // Mode edit: muat data project milik sendiri
    if (isEdit) {
      api
        .get(`/projects/${id}`)
        .then((res) => {
          const p = res.data.data;
          setForm({
            title: p.title || '',
            description: p.description || '',
            tech_stack: p.tech_stack || '',
            repo_url: p.repo_url || '',
            demo_url: p.demo_url || '',
            image_url: p.image_url || '',
            status: p.status || 'published',
          });
          setGallery((p.gallery || []).map((g) => g.image_url));
          setLoading(false);
        })
        .catch((err) => {
          setError(
            err.response?.status === 404
              ? 'Project tidak ditemukan atau bukan milik Anda.'
              : err.response?.data?.error || 'Gagal memuat project.'
          );
          setLoading(false);
        });
    }
  }, [id, isEdit, navigate]);

  const validate = () => {
    const errs = {};
    if (!form.title.trim()) errs.title = 'Judul wajib diisi';
    if (!form.description.trim()) errs.description = 'Deskripsi wajib diisi';
    if (form.repo_url && !/^https?:\/\//i.test(form.repo_url.trim())) {
      errs.repo_url = 'URL repo harus diawali http:// atau https://';
    }
    if (form.demo_url && !/^https?:\/\//i.test(form.demo_url.trim())) {
      errs.demo_url = 'URL demo harus diawali http:// atau https://';
    }
    setFieldErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleUploadGallery = async (e) => {
    const files = Array.from(e.target.files || []);
    e.target.value = ''; // reset agar file sama bisa dipilih lagi
    if (files.length === 0) return;

    setUploading(true);
    setError('');
    try {
      const uploadedUrls = [];
      for (const file of files) {
        const fd = new FormData();
        fd.append('image', file);
        const res = await api.post('/images/project', fd, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
        uploadedUrls.push(res.data.image_url);
      }
      setGallery((prev) => [...prev, ...uploadedUrls]);
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setError(err.response?.data?.error || 'Gagal mengunggah gambar.');
    } finally {
      setUploading(false);
    }
  };

  const removeGalleryItem = (index) => {
    setGallery((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validate()) return;

    try {
      setSaving(true);
      setError('');
      setSuccess('');
      const payload = {
        ...form,
        gallery, // urutan array = display_order di backend
      };
      const res = isEdit
        ? await api.put(`/projects/${id}`, payload)
        : await api.post('/projects', payload);
      setSuccess(res.data.message || 'Berhasil disimpan');
      // Sinkronkan UI dengan data terbaru dari backend, lalu ke halaman detail
      const savedId = res.data.data?.id || id;
      navigate(`/projects/${savedId}`);
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      if (err.response?.status === 403) {
        setError(err.response.data.error || 'Anda tidak memiliki izin.');
        return;
      }
      setError(err.response?.data?.error || 'Gagal menyimpan project.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm('Hapus project ini secara permanen?')) return;
    try {
      setDeleting(true);
      setError('');
      await api.delete(`/projects/${id}`);
      navigate('/projects');
    } catch (err) {
      if (err.response?.status === 401) {
        navigate('/login');
        return;
      }
      setError(err.response?.data?.error || 'Gagal menghapus project.');
      setDeleting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center font-body">
        <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  const inputClass = (hasError) =>
    `w-full bg-white border-2 p-3.5 text-sm text-[#1E293B] rounded-xl placeholder:text-[#94A3B8] focus:outline-none transition-colors ${
      hasError
        ? 'border-red-400 focus:border-red-500'
        : 'border-[#E2E8F0] focus:border-[#112320]'
    }`;

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body selection:bg-[#D97757] selection:text-white">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        <Link
          to="/projects"
          className="inline-block mb-6 text-[#64748B] text-sm font-bold tracking-wider uppercase hover:text-[#112320] transition-colors"
        >
          ← Kembali ke Daftar
        </Link>

        <h1 className="font-display text-3xl md:text-4xl font-semibold text-[#1E293B] mb-2">
          {isEdit ? 'Edit Project' : 'Project Baru'}
        </h1>
        <p className="text-[#64748B] mb-8">
          {isEdit
            ? 'Perbarui informasi project showcase kamu.'
            : 'Bagikan karya terbaikmu ke komunitas kampus.'}
        </p>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
            {error}
          </div>
        )}
        {success && (
          <div className="mb-6 p-4 bg-green-50 border-l-4 border-green-500 text-green-700 text-sm font-medium rounded">
            {success}
          </div>
        )}

        <form onSubmit={handleSubmit} className="bg-white rounded-2xl border border-[#E2E8F0] shadow-sm p-6 md:p-8 space-y-6">
          {/* Judul */}
          <div>
            <label htmlFor="title" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
              Judul Project *
            </label>
            <input
              id="title"
              type="text"
              value={form.title}
              onChange={setField('title')}
              maxLength={150}
              placeholder="cth: Smart Kampus App"
              className={inputClass(fieldErrors.title)}
            />
            {fieldErrors.title && <p className="text-xs text-red-500 mt-1">{fieldErrors.title}</p>}
          </div>

          {/* Deskripsi */}
          <div>
            <label htmlFor="description" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
              Deskripsi *
            </label>
            <textarea
              id="description"
              rows={5}
              value={form.description}
              onChange={setField('description')}
              placeholder="Ceritakan project kamu: masalah yang diselesaikan, fitur utama, dsb."
              className={`${inputClass(fieldErrors.description)} resize-none`}
            />
            {fieldErrors.description && (
              <p className="text-xs text-red-500 mt-1">{fieldErrors.description}</p>
            )}
          </div>

          {/* Tech stack */}
          <div>
            <label htmlFor="tech_stack" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
              Tech Stack
            </label>
            <input
              id="tech_stack"
              type="text"
              value={form.tech_stack}
              onChange={setField('tech_stack')}
              placeholder="cth: React,Go,PostgreSQL (dipisah koma)"
              className={inputClass(false)}
            />
          </div>

          {/* Repo & Demo */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label htmlFor="repo_url" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
                URL Repository
              </label>
              <input
                id="repo_url"
                type="text"
                value={form.repo_url}
                onChange={setField('repo_url')}
                placeholder="https://github.com/..."
                className={inputClass(fieldErrors.repo_url)}
              />
              {fieldErrors.repo_url && <p className="text-xs text-red-500 mt-1">{fieldErrors.repo_url}</p>}
            </div>
            <div>
              <label htmlFor="demo_url" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
                URL Demo
              </label>
              <input
                id="demo_url"
                type="text"
                value={form.demo_url}
                onChange={setField('demo_url')}
                placeholder="https://demo-project.com"
                className={inputClass(fieldErrors.demo_url)}
              />
              {fieldErrors.demo_url && <p className="text-xs text-red-500 mt-1">{fieldErrors.demo_url}</p>}
            </div>
          </div>

          {/* Gambar cover (URL tunggal, kompatibel dengan card) */}
          <div>
            <label htmlFor="image_url" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
              Gambar Cover (URL)
            </label>
            <input
              id="image_url"
              type="text"
              value={form.image_url}
              onChange={setField('image_url')}
              placeholder="https://res.cloudinary.com/... (opsional)"
              className={inputClass(false)}
            />
            <p className="text-xs text-[#94A3B8] mt-1">
              Digunakan sebagai tampilan utama pada list. Galeri lengkap ada di bawah.
            </p>
          </div>

          {/* Status draft/published */}
          <div>
            <label htmlFor="status" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
              Status
            </label>
            <select id="status" value={form.status} onChange={setField('status')} className={inputClass(false)}>
              {statuses.map((s) => (
                <option key={s} value={s}>
                  {s === 'draft' ? 'Draft (hanya terlihat oleh kamu)' : 'Published (publik)'}
                </option>
              ))}
            </select>
          </div>

          {/* Galeri multi-image */}
          <div>
            <label className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
              Galeri Gambar ({gallery.length})
            </label>
            {gallery.length > 0 && (
              <div className="grid grid-cols-3 sm:grid-cols-4 gap-3 mb-3">
                {gallery.map((url, index) => (
                  <div key={`${url}-${index}`} className="relative group">
                    <img
                      src={url}
                      alt={`Galeri ${index + 1}`}
                      className="w-full h-24 object-cover rounded-xl border border-[#E2E8F0]"
                    />
                    <button
                      type="button"
                      onClick={() => removeGalleryItem(index)}
                      className="absolute top-1 right-1 w-6 h-6 rounded-full bg-[#1E293B]/80 text-white text-xs font-bold opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
                      aria-label="Hapus gambar"
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            )}
            <label
              htmlFor="gallery-upload"
              className={`inline-flex items-center gap-2 border-2 border-dashed border-[#E2E8F0] hover:border-[#D97757] rounded-xl px-5 py-3 text-sm font-semibold text-[#64748B] hover:text-[#D97757] transition-colors cursor-pointer ${
                uploading ? 'opacity-60 pointer-events-none' : ''
              }`}
            >
              {uploading ? (
                <>
                  <span className="w-4 h-4 border-2 border-[#D97757] border-t-transparent rounded-full animate-spin"></span>
                  Mengunggah...
                </>
              ) : (
                <>📷 Tambah Gambar (bisa pilih beberapa)</>
              )}
            </label>
            <input
              id="gallery-upload"
              type="file"
              accept="image/jpeg,image/png,image/gif,image/webp"
              multiple
              onChange={handleUploadGallery}
              className="hidden"
            />
          </div>

          {/* Aksi */}
          <div className="flex flex-wrap items-center justify-between gap-4 pt-4 border-t border-[#E2E8F0]">
            {isEdit && (
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleting || saving}
                className="text-sm font-bold text-red-600 hover:text-red-700 border-2 border-red-200 hover:border-red-400 px-5 py-2.5 rounded-full transition-colors disabled:opacity-60 cursor-pointer"
              >
                {deleting ? 'Menghapus...' : 'Hapus Project'}
              </button>
            )}
            <div className="flex items-center gap-3 ml-auto">
              <Link
                to={isEdit ? `/projects/${id}` : '/projects'}
                className="text-sm font-bold text-[#64748B] hover:text-[#1E293B] px-4 py-2.5 transition-colors"
              >
                Batal
              </Link>
              <button
                type="submit"
                disabled={saving || uploading}
                className="bg-[#D97757] hover:bg-[#C26244] disabled:bg-[#E2E8F0] disabled:text-[#94A3B8] text-white px-6 py-2.5 text-sm font-bold tracking-wider uppercase transition-colors rounded-full cursor-pointer"
              >
                {saving ? 'Menyimpan...' : isEdit ? 'Simpan Perubahan' : 'Publikasikan'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};

export default ProjectForm;
