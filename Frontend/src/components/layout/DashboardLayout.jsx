import { Outlet, Link, useLocation } from 'react-router-dom';
import { Home, BookOpen, MessageCircle, CreditCard, User, Bell, Search } from 'lucide-react';

export default function DashboardLayout() {
  const location = useLocation();

  // Item yang belum punya halaman (scope fase berikutnya) dinonaktifkan secara
  // jujur: tetap terlihat, tapi non-klik dengan label "Segera" — bukan dead-link.
  const navItems = [
    { name: 'Home', path: '/dashboard', icon: Home },
    { name: 'My Courses', path: null, icon: BookOpen },
    { name: 'Inbox', path: null, icon: MessageCircle },
    { name: 'Transaction', path: null, icon: CreditCard },
    { name: 'Profile', path: '/profile', icon: User },
  ];

  return (
    <div className="min-h-screen bg-background flex flex-col md:flex-row">
      
      {/* Sidebar (Desktop) */}
      <aside className="hidden md:flex flex-col w-64 bg-white shadow-sm h-screen sticky top-0 z-40">
        <div className="p-6 flex items-center gap-3 border-b border-gray-100">
          <div className="w-8 h-8 bg-primary text-white rounded-lg flex items-center justify-center font-bold">C</div>
          <span className="text-xl font-bold text-gray-800">CampusConnect</span>
        </div>
        <nav className="flex-1 px-4 py-6 space-y-2">
          {navItems.map((item) => {
            const isActive = item.path !== null && location.pathname === item.path;
            const Icon = item.icon;
            const baseClass = `flex items-center gap-3 px-4 py-3 rounded-2xl transition-colors ${
              isActive ? 'bg-primary text-white shadow-md' : 'text-gray-500 hover:bg-gray-50 hover:text-primary'
            }`;

            if (item.path === null) {
              return (
                <span
                  key={item.name}
                  aria-disabled="true"
                  title="Segera hadir"
                  className={`${baseClass} opacity-50 cursor-not-allowed select-none`}
                >
                  <Icon className="w-5 h-5" />
                  <span className="font-medium text-sm">{item.name}</span>
                  <span className="ml-auto text-[10px] uppercase tracking-wide text-gray-400">Segera</span>
                </span>
              );
            }

            return (
              <Link key={item.name} to={item.path} className={baseClass}>
                <Icon className="w-5 h-5" />
                <span className="font-medium text-sm">{item.name}</span>
              </Link>
            );
          })}
        </nav>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col min-h-screen pb-20 md:pb-0">
        
        {/* Top Navbar (Mobile & Desktop) */}
        <header className="bg-white/80 backdrop-blur-md sticky top-0 z-30 px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-xl font-bold text-gray-800 hidden md:block">Hi, Alex</h1>
            {/* Mobile Header (replaces sidebar logo on small screens) */}
            <div className="md:hidden flex items-center gap-2">
              <div className="w-8 h-8 bg-primary text-white rounded-lg flex items-center justify-center font-bold text-sm">C</div>
              <span className="font-bold text-gray-800">CampusConnect</span>
            </div>
          </div>
          
          <div className="flex items-center gap-4">
            {/* Search Bar (Hidden on mobile, moved to content) */}
            <div className="hidden md:flex items-center bg-gray-50 rounded-full px-4 py-2 border border-gray-100">
              <Search className="w-4 h-4 text-gray-400 mr-2" />
              <input type="text" placeholder="Search..." className="bg-transparent border-none outline-none text-sm w-48" />
            </div>
            {/* Notification Bell */}
            <button className="w-10 h-10 bg-white border border-gray-200 rounded-full flex items-center justify-center text-gray-500 hover:text-primary hover:border-primary transition-colors">
              <Bell className="w-5 h-5" />
            </button>
          </div>
        </header>

        {/* Dynamic Content (Dashboard Pages will render here) */}
        <div className="flex-1 p-6 overflow-y-auto">
          <Outlet />
        </div>
      </main>

      {/* Bottom Navigation (Mobile Only) */}
      <nav className="md:hidden fixed bottom-0 w-full bg-white border-t border-gray-100 flex items-center justify-around py-3 z-50">
        {navItems.map((item) => {
          const isActive = item.path !== null && location.pathname === item.path;
          const Icon = item.icon;
          const itemClass = `flex flex-col items-center gap-1 ${
            isActive ? 'text-primary' : 'text-gray-400'
          } ${item.path === null ? 'opacity-40 cursor-not-allowed' : ''}`;
          const content = (
            <>
              <Icon className="w-6 h-6" />
              <span className="text-[10px] font-medium">{item.name}</span>
            </>
          );

          if (item.path === null) {
            return (
              <span key={item.name} aria-disabled="true" className={itemClass}>
                {content}
              </span>
            );
          }

          return (
            <Link key={item.name} to={item.path} className={itemClass}>
              {content}
            </Link>
          );
        })}
      </nav>

    </div>
  );
}