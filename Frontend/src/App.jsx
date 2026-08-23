import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Landing from './pages/Landing';
import Login from './pages/Login';
import Register from './pages/Register';
import DashboardLayout from './components/layout/DashboardLayout';
import Dashboard from './pages/Dashboard';
import Profile from './pages/Profile';
import EditProfile from './pages/EditProfile';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        
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