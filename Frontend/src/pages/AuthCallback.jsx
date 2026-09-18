import { useEffect, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

const AuthCallback = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const handled = useRef(false); // Hindari pemrosesan ganda saat StrictMode re-render

  useEffect(() => {
    // Pastikan token hanya diproses satu kali
    if (handled.current) return;
    handled.current = true;

    const accessToken = searchParams.get('access_token');
    const refreshToken = searchParams.get('refresh_token');

    if (accessToken) {
      // Simpan token ke localStorage (sama seperti alur OTP)
      localStorage.setItem('access_token', accessToken);
      if (refreshToken) {
        localStorage.setItem('refresh_token', refreshToken);
      }
      // Login sukses → Home utama aplikasi (Feed), bukan profil
      navigate('/feed', { replace: true });
    } else {
      // Tidak ada token di URL → kembali ke halaman login
      navigate('/login', { replace: true });
    }
  }, [navigate, searchParams]);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-[#F8F9FA] font-body">
      <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin mb-4"></div>
      <p className="text-sm font-medium text-[#64748B]">Menghubungkan akun Google Anda...</p>
    </div>
  );
};

export default AuthCallback;
