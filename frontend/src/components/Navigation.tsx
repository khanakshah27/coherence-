import React from 'react';

export default function Navigation() {
  return (
    <nav className="bg-white shadow-sm border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <h1 className="text-2xl font-bold text-gray-900">
                🔍 Coherence
              </h1>
            </div>
            <div className="hidden md:block ml-10">
              <div className="flex space-x-8">
                <a href="#" className="text-gray-700 hover:text-gray-900 px-3 py-2 text-sm font-medium">
                  Dashboard
                </a>
                <a href="#" className="text-gray-700 hover:text-gray-900 px-3 py-2 text-sm font-medium">
                  Scans
                </a>
                <a href="#" className="text-gray-700 hover:text-gray-900 px-3 py-2 text-sm font-medium">
                  Drifts
                </a>
                <a href="#" className="text-gray-700 hover:text-gray-900 px-3 py-2 text-sm font-medium">
                  Reports
                </a>
              </div>
            </div>
          </div>
          <div className="flex items-center">
            <button className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">
              + New Scan
            </button>
          </div>
        </div>
      </div>
    </nav>
  );
}
