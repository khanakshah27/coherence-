import React, { useEffect, useState } from 'react';
import { Routes, Route, Link, useLocation } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import DriftsPage from './pages/Drifts';
import RemediationsPage from './pages/Remediations';
import ReportsPage from './pages/Reports';
import { scansApi } from './api/client';

interface Scan {
  id: string;
  cloud_provider: string;
  status: string;
  drift_count: number;
  created_at: string;
}

function NavLink({ to, children }: { to: string; children: React.ReactNode }) {
  const { pathname } = useLocation();
  const active = pathname === to || (to !== '/' && pathname.startsWith(to));
  return (
    <Link
      to={to}
      className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
        active ? 'bg-blue-700 text-white' : 'text-blue-100 hover:bg-blue-700 hover:text-white'
      }`}
    >
      {children}
    </Link>
  );
}

export default function App() {
  const [scans, setScans] = useState<Scan[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchScans = async () => {
    try {
      setLoading(true);
      const data = await scansApi.list();
      setScans(Array.isArray(data) ? data : []);
    } catch {
      setScans([]);
    } finally {
      setLoading(false);
    }
  };

  const triggerScan = async (cloudProvider: string, regions: string[]) => {
    try {
      await scansApi.create({
        cloud_provider: cloudProvider,
        regions,
        resource_types: ['ec2', 's3', 'rds'],
      });
      fetchScans();
    } catch (err) {
      console.error('Failed to trigger scan:', err);
    }
  };

  useEffect(() => { fetchScans(); }, []);

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-blue-800 shadow-lg">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex items-center gap-8">
              <Link to="/" className="flex items-center gap-2">
                <span className="text-2xl">🔍</span>
                <span className="text-xl font-bold text-white tracking-tight">Coherence</span>
              </Link>
              <div className="hidden md:flex gap-1">
                <NavLink to="/">Dashboard</NavLink>
                <NavLink to="/drifts">Drifts</NavLink>
                <NavLink to="/remediations">Remediations</NavLink>
                <NavLink to="/reports">Reports</NavLink>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-blue-200 text-xs">v1.0.0</span>
              <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" title="Connected" />
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <Routes>
          <Route path="/" element={
            <Dashboard
              scans={scans}
              loading={loading}
              onRefresh={fetchScans}
              onTriggerScan={triggerScan}
            />
          } />
          <Route path="/drifts" element={<DriftsPage />} />
          <Route path="/remediations" element={<RemediationsPage />} />
          <Route path="/reports" element={<ReportsPage />} />
        </Routes>
      </main>

      <footer className="border-t border-gray-200 bg-white mt-16">
        <div className="max-w-7xl mx-auto px-4 py-4 text-center text-xs text-gray-400">
          Coherence — Infrastructure State Drift Detection &amp; Auto-Remediation · MIT License
        </div>
      </footer>
    </div>
  );
}
