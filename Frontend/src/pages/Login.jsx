import { Mail, Lock, EyeOff } from 'lucide-react';
import { Link } from 'react-router-dom';

export default function Login() {
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white p-8 rounded-3xl shadow-sm">
        
        {/* Logo / Title Area */}
        <div className="text-center mb-10">
          <div className="flex justify-center items-center gap-2 mb-4">
            {/* Placeholder Icon Logo */}
            <div className="w-10 h-10 bg-primary text-white rounded-lg flex items-center justify-center font-bold text-xl">
              C
            </div>
            <h1 className="text-2xl font-bold text-gray-800">CampusConnect</h1>
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">Let's Sign In.!</h2>
          <p className="text-sm text-gray-500">Login to Your Account to Continue your Courses</p>
        </div>

        {/* Form Area */}
        <form className="space-y-5">
          {/* Email Input */}
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
              <Mail className="h-5 w-5 text-gray-400" />
            </div>
            <input
              type="email"
              className="w-full pl-11 pr-4 py-3.5 bg-gray-50 border border-gray-200 rounded-full text-sm focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors"
              placeholder="Email"
              required
            />
          </div>

          {/* Password Input */}
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
              <Lock className="h-5 w-5 text-gray-400" />
            </div>
            <input
              type="password"
              className="w-full pl-11 pr-12 py-3.5 bg-gray-50 border border-gray-200 rounded-full text-sm focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors"
              placeholder="Password"
              required
            />
            <button type="button" className="absolute inset-y-0 right-0 pr-4 flex items-center text-gray-400 hover:text-gray-600">
              <EyeOff className="h-5 w-5" />
            </button>
          </div>

          {/* Options: Remember Me & Forgot Password */}
          <div className="flex items-center justify-between px-1">
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" className="w-4 h-4 text-primary rounded border-gray-300 focus:ring-primary" />
              <span className="text-sm text-gray-600">Remember Me</span>
            </label>
            <Link to="#" className="text-sm text-gray-500 hover:text-primary transition-colors">
              Forgot Password?
            </Link>
          </div>

          {/* Sign In Button */}
          <button
            type="submit"
            className="w-full bg-primary hover:bg-blue-700 text-white font-medium py-3.5 rounded-full flex items-center justify-center gap-2 transition-colors mt-4"
          >
            Sign In
            <div className="w-6 h-6 bg-white text-primary rounded-full flex items-center justify-center">
              <span className="text-lg leading-none mb-0.5">→</span>
            </div>
          </button>
        </form>

        {/* Register Link */}
        <div className="mt-8 text-center text-sm">
          <span className="text-gray-500">Don't have an Account? </span>
          <Link to="/register" className="text-primary font-semibold hover:underline">
            SIGN UP
          </Link>
        </div>
        
      </div>
    </div>
  );
}