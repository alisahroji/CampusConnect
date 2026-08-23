import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';

const Profile = () => {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        console.log("1. Memulai pemanggilan API...");
        const token = localStorage.getItem('access_token');
        console.log("2. Status Token:", token ? "Token Tersimpan Aman" : "TOKEN KOSONG!");

        const response = await api.get('/profile');
        console.log("3. Respon utuh dari Golang:", response);

        // Pengambilan data super aman dengan Opsional Chaining (?.)
        const actualProfileData = response.data?.data || response.data?.user || response.data;
        console.log("4. Data yang siap dirender:", actualProfileData);

        setProfile(actualProfileData);
      } catch (err) {
        console.error("!!! ERROR FATAL TERDETEKSI !!!", err);
        const errorMessage = err.response?.data?.error || err.message || 'Error tidak diketahui.';
        setError(`Gagal: ${errorMessage}`);
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, []);
  
  if (loading) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center font-body">
        <div className="animate-pulse text-[#D97757] font-bold tracking-widest uppercase">Menyiapkan Ruang Kerja...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center font-body text-red-600">
        {error}
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#F8F9FA] font-body selection:bg-[#D97757] selection:text-white">
      {/* Banner Atas - Deep Forest */}
      <div className="h-48 md:h-64 bg-[#112320] relative overflow-hidden">
        <div className="absolute inset-0 opacity-20" 
             style={{ backgroundImage: 'linear-gradient(#F8F9FA 1px, transparent 1px), linear-gradient(90deg, #F8F9FA 1px, transparent 1px)', backgroundSize: '4rem 4rem' }}>
        </div>
      </div>

      {/* Konten Utama Profil */}
      <div className="max-w-4xl mx-auto px-6 sm:px-8 lg:px-12 -mt-20 relative z-10">
        <div className="flex flex-col md:flex-row gap-8 md:items-end">
          
          {/* Avatar Cloudinary */}
          <div className="relative">
            <img 
              src={profile?.PictureURL || 'https://ui-avatars.com/api/?name=' + profile?.Name} 
              alt="Profile Avatar" 
              className="w-40 h-40 object-cover border-4 border-[#F8F9FA] shadow-lg bg-white"
              style={{ borderRadius: '4px' }} // Sharp tapi sedikit halus
            />
          </div>

          {/* Info Identitas */}
          <div className="flex-1 pb-2">
            <h1 className="font-display text-4xl font-bold text-[#1E293B] mb-2">{profile?.Name || 'Nama Belum Diatur'}</h1>
            <p className="text-[#D97757] font-bold tracking-wide uppercase text-sm">{profile?.Email}</p>
          </div>

          {/* Tombol Edit */}
          <div className="pb-2">
            <Link to="/edit-profile" className="inline-block bg-white border border-[#E2E8F0] hover:border-[#112320] text-[#1E293B] font-bold py-3 px-6 text-sm tracking-wider uppercase transition-colors shadow-sm">
              Edit Profil
            </Link>
          </div>
        </div>

        {/* Bio & Detail */}
        <div className="mt-12 grid grid-cols-1 md:grid-cols-3 gap-12">
          <div className="md:col-span-2">
            <h3 className="text-xs font-bold text-[#94A3B8] uppercase tracking-widest border-b border-[#E2E8F0] pb-4 mb-6">Tentang Saya</h3>
            <p className="text-[#1E293B] leading-relaxed text-lg font-light">
              {profile?.Bio || 'Belum ada bio yang ditulis. Ceritakan sedikit tentang dirimu!'}
            </p>
          </div>
          
          <div>
            <h3 className="text-xs font-bold text-[#94A3B8] uppercase tracking-widest border-b border-[#E2E8F0] pb-4 mb-6">Informasi Akademik</h3>
            <div className="space-y-4 text-sm text-[#1E293B]">
              <p><span className="font-bold">Role:</span> {profile?.Role}</p>
              <p><span className="font-bold">Bergabung:</span> {profile?.CreatedAt ? new Date(profile?.CreatedAt).toLocaleDateString('id-ID') : 'Data tanggal tidak tersedia'}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Profile;