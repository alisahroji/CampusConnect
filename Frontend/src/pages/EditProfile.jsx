import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import api from '../services/api';

const EditProfile = () => {
  const [formData, setFormData] = useState({ name: '', bio: '' });
  const [avatar, setAvatar] = useState(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    // Tarik data saat ini untuk mengisi form default
    const fetchProfile = async () => {
      try {
        const response = await api.get('/profile');
        const data = response.data?.data || response.data?.user || response.data;
        setFormData({ 
          name: data?.name || '',
          bio: data?.bio || ''
        });
      } catch (err) {
        console.error(err);
        setError('Gagal memuat data profil. Sesi mungkin telah berakhir.');
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, []);

  const handleInputChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleFileChange = (e) => {
    // Mengambil file gambar pertama yang dipilih
    if (e.target.files && e.target.files[0]) {
      setAvatar(e.target.files[0]);
    }
  };

const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');

    try {
      // 1. KIRIM DATA TEKS (sebagai JSON murni sesuai struct Golang)
      // Perhatikan huruf kecil "name" dan "bio" sesuai tag json di Golang
      await api.put('/profile', {
        name: formData.name,
        bio: formData.bio,
        skills: "", // Bisa dikosongkan jika belum ada di UI
        github_url: "",
        linkedin_url: ""
      });

      // 2. KIRIM FOTO (Jika ada file yang dipilih)
      if (avatar) {
        const avatarData = new FormData();
        avatarData.append('avatar', avatar);

        // CATATAN: Pastikan endpoint POST '/profile/avatar' ini 
        // sudah sesuai dengan rute UploadAvatar di router Golang-mu.
        await api.post('/profile/avatar', avatarData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        });
      }

      // 3. Kembali ke halaman profil setelah sukses
      navigate('/profile');
    } catch (err) {
      console.error("Error Update:", err);
      setError(err.response?.data?.error || 'Terjadi kesalahan saat memperbarui profil.');
    } finally {
      setSaving(false);
    }
  };
  if (loading) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center font-body">
        <div className="animate-pulse text-[#D97757] font-bold tracking-widest uppercase">Membuka Dokumen...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body selection:bg-[#D97757] selection:text-white pb-20">
      {/* Mini Header */}
      <div className="bg-[#112320] py-12 relative overflow-hidden">
        <div className="absolute inset-0 opacity-20" 
             style={{ backgroundImage: 'linear-gradient(#F8F9FA 1px, transparent 1px), linear-gradient(90deg, #F8F9FA 1px, transparent 1px)', backgroundSize: '4rem 4rem' }}>
        </div>
        <div className="max-w-3xl mx-auto px-6 relative z-10">
          <Link to="/profile" className="text-[#94A3B8] text-sm font-bold tracking-wider uppercase hover:text-white transition-colors">
            ← Kembali ke Profil
          </Link>
          <h1 className="font-display text-4xl font-bold text-[#F8F9FA] mt-6">Edit Identitas</h1>
        </div>
      </div>

      <div className="max-w-3xl mx-auto px-6 mt-12">
        {error && (
          <div className="mb-8 p-4 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-10">
          
          {/* Section: Informasi Dasar */}
          <div className="bg-white p-8 border border-[#E2E8F0] shadow-sm">
            <h2 className="text-xs font-bold text-[#94A3B8] uppercase tracking-widest border-b border-[#E2E8F0] pb-4 mb-8">Informasi Dasar</h2>
            
            <div className="space-y-8">
              <div>
                <label htmlFor="name" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
                  Nama Lengkap
                </label>
                <input 
                  type="text" id="name"

                  name="name"

                  value={formData.name}
                  onChange={handleInputChange}
                  className="w-full bg-transparent border-b-2 border-[#E2E8F0] py-3 text-[#1E293B] focus:border-[#112320] focus:outline-none transition-colors rounded-none placeholder:text-[#94A3B8] text-lg"
                  placeholder="Masukkan nama lengkapmu..."
                  required
                />
              </div>

              <div>
                <label htmlFor="bio" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-2">
                  Biografi Singkat
                </label>
                <textarea

                  id="bio"

                  name="bio"

                  value={formData.bio}
                  onChange={handleInputChange}
                  rows="4"
                  className="w-full bg-transparent border-2 border-[#E2E8F0] p-4 text-[#1E293B] focus:border-[#112320] focus:outline-none transition-colors rounded-none placeholder:text-[#94A3B8] resize-none"
                  placeholder="Ceritakan minat dan keahlian utamamu..."
                />
              </div>
            </div>
          </div>

          {/* Section: Avatar */}
          <div className="bg-white p-8 border border-[#E2E8F0] shadow-sm">
            <h2 className="text-xs font-bold text-[#94A3B8] uppercase tracking-widest border-b border-[#E2E8F0] pb-4 mb-8">Foto Profil</h2>
            
            <div>
              <label className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider mb-4">
                Unggah Avatar Baru
              </label>
              <input 
                type="file" 
                accept="image/*"
                onChange={handleFileChange}
                className="block w-full text-sm text-[#94A3B8]
                  file:mr-4 file:py-3 file:px-6
                  file:rounded-none file:border-0
                  file:text-xs file:font-bold file:uppercase file:tracking-wider
                  file:bg-[#112320] file:text-white
                  hover:file:bg-[#1E293B] transition-all cursor-pointer border border-[#E2E8F0] p-2"
              />
              <p className="mt-3 text-xs text-[#94A3B8]">Maksimal ukuran file: 2MB (JPG, PNG).</p>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex justify-end gap-4 pt-4">
            <Link to="/profile" className="px-8 py-4 text-sm font-bold text-[#1E293B] uppercase tracking-wider hover:bg-[#E2E8F0] transition-colors">
              Batal
            </Link>
            <button 
              type="submit"
              disabled={saving}
              className="bg-[#D97757] hover:bg-[#C26244] disabled:bg-[#E2E8F0] disabled:text-[#94A3B8] text-white px-10 py-4 text-sm font-bold tracking-wider uppercase transition-colors rounded-none"
            >
              {saving ? 'Menyimpan...' : 'Simpan Perubahan'}
            </button>
          </div>

        </form>
      </div>
    </div>
  );
};

export default EditProfile;