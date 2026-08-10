import { Mail, Lock, EyeOff, User } from 'lucide-react';
import { Link } from 'react-router-dom';

export default function Register() {
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white p-8 rounded-3xl shadow-sm">
        
        {/* Logo / Title Area */}
        <div className="text-center mb-8">
          <div className="flex justify-center items-center gap-2 mb-4">
            <div className="w-10 h-10 bg-primary text-white rounded-lg flex items-center justify-center font-bold text-xl">
              C
            </div>
            <h1 className="text-2xl font-bold text-gray-800">CampusConnect</h1>
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">Getting Started.!</h2>
          <p className="text-sm text-gray-500">Create an Account to Continue your Courses</p>
        </div>

        {/* Form Area */}
        <form className="space-y-4">
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

          {/* Terms & Conditions */}
          <div className="flex items-start px-1 mt-2">
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" className="w-4 h-4 text-primary rounded border-gray-300 focus:ring-primary mt-0.5" required />
              <span className="text-sm text-gray-600">
                Agree to <Link to="#" className="text-gray-900 font-medium hover:underline">Terms & Conditions</Link>
              </span>
            </label>
          </div>

          {/* Sign Up Button */}
          <button
            type="submit"
            className="w-full bg-primary hover:bg-blue-700 text-white font-medium py-3.5 rounded-full flex items-center justify-center gap-2 transition-colors mt-6"
          >
            Sign Up
            <div className="w-6 h-6 bg-white text-primary rounded-full flex items-center justify-center">
              <span className="text-lg leading-none mb-0.5">→</span>
            </div>
          </button>
        </form>

        {/* Login Link */}
        <div className="mt-8 text-center text-sm">
          <span className="text-gray-500">Already have an Account? </span>
          <Link to="/login" className="text-primary font-semibold hover:underline">
            SIGN IN
          </Link>
        </div>
        
      </div>
    </div>
  );
}