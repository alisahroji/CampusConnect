import { Link } from 'react-router-dom';
import { User } from 'lucide-react';

export default function Landing() {
  return (
    <div className="min-h-screen bg-white">
      {/* Navigation Bar */}
      <nav className="flex items-center justify-between px-8 py-4 bg-white shadow-sm sticky top-0 z-50">
        <div className="flex items-center gap-2">
          {/* Logo */}
          <div className="w-8 h-8 bg-primary text-white rounded-lg flex items-center justify-center font-bold">
            C
          </div>
          <span className="text-xl font-bold text-gray-800 tracking-tight">CampusConnect</span>
        </div>

        {/* Center Links (Desktop only) */}
        <div className="hidden md:flex items-center gap-8 text-sm font-medium text-gray-600">
          <Link to="/" className="text-gray-900 hover:text-primary transition-colors">Home</Link>
          <Link to="/projects" className="hover:text-primary transition-colors">Projects</Link>
          <Link to="/search" className="hover:text-primary transition-colors">Search</Link>
        </div>

        {/* Right Actions */}
        <div className="flex items-center gap-4">
          <Link 
            to="/login" 
            aria-label="Masuk"
            className="text-gray-500 hover:text-primary transition-colors p-2 rounded-full hover:bg-gray-100"
          >
            <User className="w-5 h-5" />
          </Link>
          <Link 
            to="/login" 
            className="bg-primary hover:bg-blue-700 text-white px-5 py-2.5 rounded-full text-sm font-medium transition-colors"
          >
            Get Started
          </Link>
        </div>
      </nav>

      {/* Hero Section */}
      <div className="relative w-full h-[600px] bg-gray-900">
        {/* Background Image with Overlay */}
        <img 
          src="https://images.unsplash.com/photo-1522202176988-66273c2fd55f?q=80&w=2071&auto=format&fit=crop" 
          alt="Students learning" 
          className="absolute inset-0 w-full h-full object-cover object-center"
        />
        {/* Gradient Overlay for Text Readability */}
        <div className="absolute inset-0 bg-gradient-to-r from-black/80 via-black/50 to-transparent"></div>

        {/* Hero Content */}
        <div className="relative z-10 h-full max-w-7xl mx-auto px-8 flex flex-col justify-center w-full md:w-1/2 text-white">
          <h1 className="text-4xl md:text-5xl font-bold leading-tight mb-6">
            Connecting You to the Best Educational Resources
          </h1>
          <p className="text-lg text-gray-200 mb-8 max-w-lg leading-relaxed">
            Discover, learn, and grow with our curated courses and community resources. Platform eksklusif untuk memaksimalkan potensi mahasiswa di Universitas Nasional PASIM Bandung.
          </p>
          
          <div className="flex items-center gap-4">
            <Link 
              to="/login" 
              className="bg-primary hover:bg-blue-600 text-white px-6 py-3 rounded-full font-medium transition-all shadow-lg hover:shadow-xl"
            >
              Get Started
            </Link>
            <Link 
              to="/projects" 
              className="border border-white hover:bg-white hover:text-gray-900 text-white px-6 py-3 rounded-full font-medium transition-all"
            >
              Lihat Project Mahasiswa
            </Link>
          </div>
        </div>
      </div>

      {/* Placeholder for Next Sections (Featured Courses, dll) */}
      <div className="max-w-7xl mx-auto px-8 py-20">
        <h2 className="text-2xl font-bold text-center text-gray-800 mb-10">Featured Courses</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
           <div className="h-64 bg-gray-100 rounded-2xl animate-pulse"></div>
           <div className="h-64 bg-gray-100 rounded-2xl animate-pulse"></div>
           <div className="h-64 bg-gray-100 rounded-2xl animate-pulse"></div>
        </div>
      </div>
    </div>
  );
}