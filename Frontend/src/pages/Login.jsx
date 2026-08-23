import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import axios from 'axios';

const Login = () => {
  // State Management untuk mengontrol form
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [step, setStep] = useState(1); // Step 1: Input Email, Step 2: Input OTP
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');
  const navigate = useNavigate();

  // Fungsi menembak API Request OTP
  const handleRequestOTP = async (e) => {
    e.preventDefault();
    setLoading(true);
    setErrorMsg('');
    
    try {
      await axios.post('http://localhost:8080/api/auth/request-otp', { email });
      setStep(2); // Pindah ke tampilan input OTP
    } catch (error) {
      setErrorMsg(error.response?.data?.error || 'Gagal mengirim email OTP.');
    } finally {
      setLoading(false);
    }
  };

  // Fungsi menembak API Verify OTP
  const handleVerifyOTP = async (e) => {
    e.preventDefault();
    setLoading(true);
    setErrorMsg('');

    try {
      const response = await axios.post('http://localhost:8080/api/auth/verify-otp', { 
        email: email, 
        code: otp 
      });
      
      // Simpan "Tiket VIP" ke localStorage browser
      localStorage.setItem('access_token', response.data.access_token);
      
      // Jika sukses, arahkan user ke halaman Dashboard/Profile
      navigate('/'); 
    } catch (error) {
      setErrorMsg(error.response?.data?.error || 'Kode OTP salah atau kedaluwarsa.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <style>{`
        @import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;700&family=Playfair+Display:ital,wght@0,600;0,700;1,600&display=swap');
        .font-display { font-family: 'Playfair Display', serif; }
        .font-body { font-family: 'DM Sans', sans-serif; }
      `}</style>

      <div className="min-h-screen flex flex-col md:flex-row font-body bg-[#F8F9FA] selection:bg-[#D97757] selection:text-white">
        
        {/* LEFT PANEL: Editorial Brand Area */}
        <div className="relative w-full md:w-[60%] bg-[#112320] text-[#F8F9FA] p-10 md:p-16 lg:p-24 flex flex-col justify-between overflow-hidden">
          <div className="absolute inset-0 opacity-20 pointer-events-none" 
               style={{ 
                 backgroundImage: 'linear-gradient(#F8F9FA 1px, transparent 1px), linear-gradient(90deg, #F8F9FA 1px, transparent 1px)', 
                 backgroundSize: '4rem 4rem' 
               }}>
            <div className="absolute top-[8rem] left-[8rem] w-2 h-2 bg-[#D97757] rounded-full"></div>
            <div className="absolute top-[16rem] left-[12rem] w-3 h-3 bg-[#F8F9FA] rounded-none rotate-45"></div>
            <div className="absolute bottom-[12rem] right-[8rem] w-2 h-2 border border-[#D97757] rounded-full"></div>
          </div>

          <div className="relative z-10">
            <h1 className="text-xl font-bold tracking-widest uppercase text-[#D97757]">
              CampusConnect
            </h1>
          </div>

          <div className="relative z-10 mt-16 md:mt-0 max-w-xl">
            <h2 className="font-display text-4xl md:text-5xl lg:text-6xl font-semibold leading-tight mb-6">
              Kembali ke ruang kerjamu.
            </h2>
            <p className="text-[#94A3B8] text-lg font-light leading-relaxed max-w-md">
              Akses kembali portofoliomu, lanjutkan diskusi komunitas, dan pantau event akademik terkini.
            </p>
          </div>
          
          <div className="relative z-10 hidden md:block">
            <p className="text-xs text-[#94A3B8] tracking-widest uppercase">© {new Date().getFullYear()} Universitas Nasional PASIM</p>
          </div>
        </div>

        {/* RIGHT PANEL: Clean Form Area */}
        <div className="w-full md:w-[40%] flex items-center justify-center p-8 md:p-12 lg:p-16 bg-[#F8F9FA]">
          <div className="w-full max-w-sm">
            <div className="mb-10 md:hidden">
              <h1 className="text-2xl font-display font-bold text-[#1E293B]">Masuk Akun</h1>
            </div>

            {/* Menampilkan pesan error jika ada */}
            {errorMsg && (
              <div className="mb-6 p-3 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium">
                {errorMsg}
              </div>
            )}

            {/* FORM DINAMIS (Bisa berubah dari Email ke OTP) */}
            <form onSubmit={step === 1 ? handleRequestOTP : handleVerifyOTP} className="space-y-6">
              
              {step === 1 ? (
                // TAMPILAN STEP 1: MINTA EMAIL
                <div className="space-y-2">
                  <label htmlFor="email" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider">
                    Email Institusi / Pribadi
                  </label>
                  <input 
                    type="email" 
                    id="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full bg-transparent border-b-2 border-[#E2E8F0] py-3 text-[#1E293B] focus:border-[#112320] focus:outline-none transition-colors rounded-none placeholder:text-[#94A3B8]"
                    placeholder="mahasiswa@campus.ac.id"
                    disabled={loading}
                    required
                  />
                </div>
              ) : (
                // TAMPILAN STEP 2: MASUKKAN OTP
                <div className="space-y-2 animate-pulse">
                  <label htmlFor="otp" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider">
                    Masukkan 6 Digit OTP
                  </label>
                  <p className="text-xs text-[#94A3B8] mb-2">Dikirim ke: <span className="font-bold text-[#D97757]">{email}</span></p>
                  <input 
                    type="text" 
                    id="otp"
                    maxLength="6"
                    value={otp}
                    onChange={(e) => setOtp(e.target.value)}
                    className="w-full bg-transparent border-b-2 border-[#D97757] py-3 text-2xl tracking-[1em] text-center text-[#1E293B] focus:border-[#112320] focus:outline-none transition-colors rounded-none placeholder:text-[#E2E8F0]"
                    placeholder="••••••"
                    disabled={loading}
                    required
                  />
                </div>
              )}

              <button 
                type="submit"
                disabled={loading}
                className="w-full bg-[#D97757] hover:bg-[#C26244] disabled:bg-[#E2E8F0] disabled:text-[#94A3B8] text-white py-4 text-sm font-bold tracking-wider uppercase transition-colors rounded-none"
              >
                {loading ? 'Memproses...' : (step === 1 ? 'Kirim Kode OTP' : 'Verifikasi & Masuk')}
              </button>
              
              {/* Tombol kembali ke email jika salah ketik (hanya muncul di Step 2) */}
              {step === 2 && !loading && (
                <button 
                  type="button" 
                  onClick={() => setStep(1)}
                  className="w-full text-xs font-bold text-[#94A3B8] uppercase tracking-wider hover:text-[#1E293B] mt-4"
                >
                  ← Ganti Email
                </button>
              )}
            </form>

            <div className="my-8 flex items-center gap-4">
              <div className="h-px bg-[#E2E8F0] flex-1"></div>
              <span className="text-xs font-bold text-[#94A3B8] uppercase tracking-wider">Atau</span>
              <div className="h-px bg-[#E2E8F0] flex-1"></div>
            </div>

            <button 
              type="button"
              className="w-full flex items-center justify-center gap-3 bg-white border border-[#E2E8F0] hover:border-[#112320] text-[#1E293B] font-bold py-3 px-4 rounded-full transition-all duration-300"
            >
              <svg className="w-5 h-5" viewBox="0 0 24 24">
                <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
              </svg>
              <span className="text-sm tracking-wide">SIGN IN WITH GOOGLE</span>
            </button>

            <div className="mt-12 text-sm text-[#64748B]">
              Belum punya portofolio?{' '}
              <Link to="/register" className="text-[#112320] font-bold border-b border-transparent hover:border-[#112320] transition-colors pb-0.5">
                Mulai dari sini.
              </Link>
            </div>
          </div>
        </div>

      </div>
    </>
  );
};

export default Login;