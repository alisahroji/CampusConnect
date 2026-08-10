import { Search, ChevronRight } from 'lucide-react';

export default function Dashboard() {
  const categories = ['All', 'Graphic Design', '3D Design', 'Web Dev'];

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      
      {/* Mobile Greeting & Search (Visible on small screens) */}
      <div className="md:hidden space-y-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">HI, ALEX</h1>
          <p className="text-sm text-gray-500">What Would you like to learn Today? Search Below.</p>
        </div>
        <div className="relative">
          <input
            type="text"
            placeholder="Search for..."
            className="w-full pl-4 pr-12 py-3 bg-white border border-gray-200 rounded-2xl text-sm focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary shadow-sm"
          />
          <button className="absolute right-2 top-2 bottom-2 bg-primary text-white p-2 rounded-xl">
            <Search className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Blue Banner Promo */}
      <div className="bg-primary rounded-3xl p-6 md:p-8 text-white relative overflow-hidden shadow-lg shadow-blue-200">
        <div className="relative z-10 w-2/3 md:w-1/2">
          <p className="text-sm font-semibold mb-1 opacity-90">25% OFF*</p>
          <h2 className="text-2xl md:text-3xl font-bold mb-2">Today's Special</h2>
          <p className="text-sm opacity-80 mb-4 line-clamp-2 text-blue-100">Get a Discount for Every Course Order only Valid for Today.!</p>
        </div>
        {/* Decorative Circles */}
        <div className="absolute -right-10 -top-10 w-40 h-40 bg-white opacity-10 rounded-full"></div>
        <div className="absolute right-10 -bottom-10 w-24 h-24 bg-white opacity-10 rounded-full"></div>
      </div>

      {/* Categories Horizontal Scroll */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="font-bold text-gray-900">Categories</h3>
          <button className="text-sm text-primary font-medium flex items-center hover:underline">
            SEE ALL <ChevronRight className="w-4 h-4 ml-1" />
          </button>
        </div>
        <div className="flex gap-3 overflow-x-auto pb-2 scrollbar-hide">
          {categories.map((cat, idx) => (
            <button
              key={cat}
              className={`px-5 py-2 rounded-full text-sm font-medium whitespace-nowrap transition-colors ${
                idx === 1 ? 'bg-primary text-white' : 'bg-white text-gray-600 border border-gray-200 hover:border-primary'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Popular Courses Grid */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="font-bold text-gray-900">Popular Courses</h3>
          <button className="text-sm text-primary font-medium flex items-center hover:underline">
            SEE ALL <ChevronRight className="w-4 h-4 ml-1" />
          </button>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {/* Dummy Card 1 */}
          <div className="bg-white rounded-3xl p-3 shadow-sm border border-gray-100">
            <div className="w-full h-40 bg-gray-900 rounded-2xl mb-4"></div>
            <div className="px-2 pb-2">
              <p className="text-xs text-orange-500 font-medium mb-1">Graphic Design</p>
              <h4 className="font-bold text-gray-900 mb-2 line-clamp-1">Graphic Design Advanced</h4>
              <div className="flex items-center justify-between">
                <div>
                  <span className="text-primary font-bold">850/-</span>
                  <span className="text-gray-400 text-xs line-through ml-2">1250</span>
                </div>
                <div className="flex items-center text-xs font-medium text-gray-500">
                  <span className="text-yellow-500 mr-1">⭐ 4.2</span> | 7830 Std
                </div>
              </div>
            </div>
          </div>
          {/* Dummy Card 2 */}
          <div className="bg-white rounded-3xl p-3 shadow-sm border border-gray-100">
            <div className="w-full h-40 bg-gray-900 rounded-2xl mb-4"></div>
            <div className="px-2 pb-2">
              <p className="text-xs text-orange-500 font-medium mb-1">Web Development</p>
              <h4 className="font-bold text-gray-900 mb-2 line-clamp-1">React Masterclass</h4>
              <div className="flex items-center justify-between">
                <div>
                  <span className="text-primary font-bold">999/-</span>
                </div>
                <div className="flex items-center text-xs font-medium text-gray-500">
                  <span className="text-yellow-500 mr-1">⭐ 4.8</span> | 5210 Std
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

    </div>
  );
}