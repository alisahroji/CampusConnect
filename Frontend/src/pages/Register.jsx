import React from 'react';
import { Link } from 'react-router-dom';

const Register = () => {
  return (
    <>
      <style>{`
        @import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;700&family=Playfair+Display:ital,wght@0,600;0,700;1,600&display=swap');
        .font-display { font-family: 'Playfair Display', serif; }
        .font-body { font-family: 'DM Sans', sans-serif; }
      `}</style>

      <div className="min-h-screen flex flex-col md:flex-row-reverse font-body bg-[#F8F9FA] selection:bg-[#D97757] selection:text-white">
        
        {/* BRAND PANEL (Reversed for Register) */}
        <div className="relative w-full md:w-[60%] bg-[#112320] text-[#F8F9FA] p-10 md:p-16 lg:p-24 flex flex-col justify-between overflow-hidden">
          
          <div className="absolute inset-0 opacity-20 pointer-events-none" 
               style={{ 
                 backgroundImage: 'linear-gradient(#F8F9FA 1px, transparent 1px), linear-gradient(90deg, #F8F9FA 1px, transparent 1px)', 
                 backgroundSize: '4rem 4rem',
                 backgroundPosition: '2rem 2rem'
               }}>
            <div className="absolute top-[12rem] right-[8rem] w-3 h-3 bg-[#D97757] rounded-none rotate-12"></div>
            <div className="absolute bottom-[16rem] left-[12rem] w-2 h-2 border-2 border-[#F8F9FA] rounded-full"></div>
          </div>

          <div className="relative z-10 flex justify-end">
            <h1 className="text-xl font-bold tracking-widest uppercase text-[#D97757]">
              CampusConnect
            </h1>
          </div>

          <div className="relative z-10 mt-16 md:mt-0 max-w-xl self-end text-right">
            <h2 className="font-display text-4xl md:text-5xl lg:text-6xl font-semibold leading-tight mb-6">
              Bangun presensimu.
            </h2>
            <p className="text-[#94A3B8] text-lg font-light leading-relaxed ml-auto max-w-md">
              Karya akademikmu layak mendapat panggung. Bergabunglah untuk menyatukan portofolio, event, dan koneksimu dalam satu identitas.
            </p>
          </div>
          
          <div className="relative z-10 hidden md:block text-right">
            <p className="text-xs text-[#94A3B8] tracking-widest uppercase">Est. 2026 • Portofolio Masa Depan</p>
          </div>
        </div>

        {/* FORM PANEL */}
        <div className="w-full md:w-[40%] flex items-center justify-center p-8 md:p-12 lg:p-16 bg-[#F8F9FA]">
          <div className="w-full max-w-sm">
            <div className="mb-10 md:hidden">
              <h1 className="text-2xl font-display font-bold text-[#1E293B]">Registrasi</h1>
            </div>

            <form className="space-y-6">
              <div className="space-y-2">
                <label htmlFor="email" className="block text-xs font-bold text-[#1E293B] uppercase tracking-wider">
                  Email Institusi / Pribadi
                </label>
                <input 
                  type="email" 
                  id="email"
                  className="w-full bg-transparent border-b-2 border-[#E2E8F0] py-3 text-[#1E293B] focus:border-[#112320] focus:outline-none transition-colors rounded-none placeholder:text-[#94A3B8]"
                  placeholder="mahasiswa@campus.ac.id"
                  required
                />
              </div>

              <button 
                type="button"
                className="w-full bg-[#112320] hover:bg-[#1E293B] text-white py-4 text-sm font-bold tracking-wider uppercase transition-colors rounded-none"
              >
                Mulai Registrasi
              </button>
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
              {/* Google SVG */}
              <svg className="w-5 h-5" viewBox="0 0 24 24">
                <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
              </svg>
              <span className="text-sm tracking-wide">SIGN UP WITH GOOGLE</span>
            </button>

            <div className="mt-12 text-sm text-[#64748B]">
              Sudah punya akun?{' '}
              <Link to="/login" className="text-[#D97757] font-bold border-b border-transparent hover:border-[#D97757] transition-colors pb-0.5">
                Masuk di sini.
              </Link>
            </div>
          </div>
        </div>

      </div>
    </>
  );
};

export default Register;