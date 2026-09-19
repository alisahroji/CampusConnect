import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Landing from './pages/Landing';
import Login from './pages/Login';
import Register from './pages/Register';
import AuthCallback from './pages/AuthCallback';
import ProjectList from './pages/ProjectList';
import ProjectDetail from './pages/ProjectDetail';
import ProjectForm from './pages/ProjectForm';
import Feed from './pages/Feed';
import UserProfile from './pages/UserProfile';
import Search from './pages/Search';
import DashboardLayout from './components/layout/DashboardLayout';
import AppNavbar from './components/layout/AppNavbar';
import Dashboard from './pages/Dashboard';
import Profile from './pages/Profile';
import EditProfile from './pages/EditProfile';
import Bookmarks from './pages/Bookmarks';
import Notifications from './pages/Notifications';
import AdminDashboard from './pages/AdminDashboard';

function App() {
  return (
    <BrowserRouter>
      {/* Navbar global application routes (hide sendiri di Landing/Login/Register/Callback/Dashboard) */}
      <AppNavbar />
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/auth/callback" element={<AuthCallback />} />
        <Route path="/projects" element={<ProjectList />} />
        <Route path="/projects/new" element={<ProjectForm />} />
        <Route path="/projects/:id" element={<ProjectDetail />} />
        <Route path="/projects/:id/edit" element={<ProjectForm />} />
        <Route path="/feed" element={<Feed />} />
        {/* Minggu 5 Hari 6: Follow UI + Search UI */}
        <Route path="/users/:id" element={<UserProfile />} />
        <Route path="/search" element={<Search />} />

        {/* Minggu 6 Day 3: Bookmark UI + Notification */}
        <Route path="/bookmarks" element={<Bookmarks />} />
        <Route path="/notifications" element={<Notifications />} />

        {/* Minggu 6 Day 5: Admin Dashboard (guard role via /api/profile;
            authority tetap AdminGuard backend) */}
        <Route path="/admin" element={<AdminDashboard />} />
        
        {/* Rute Profile ditambahkan di sini, sejajar dengan rute utama */}
        <Route path="/profile" element={<Profile />} />
        <Route path="/edit-profile" element={<EditProfile />} />        
        {/* Rute Dashboard dengan Layout */}
        <Route path="/dashboard" element={<DashboardLayout />}>
          <Route index element={<Dashboard />} />
          {/* Halaman lain seperti /dashboard/courses, /dashboard/inbox akan ditambahkan di sini pada minggu-minggu berikutnya */}
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;