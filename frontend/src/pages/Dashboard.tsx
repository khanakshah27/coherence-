import React, { useState } from 'react';
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';

interface Scan {
  id: string;
  cloud_provider: string;
  status: string;
  drift_count: number;
  created_at: string;
}

interface DashboardProps {
  scans: Scan[];
  loading: boolean;
  onRefresh: () => void;
  onTriggerScan: (provider: string, regions: string[]) => void;
}

export default function Dashboard({ scans, loading, onRefresh, onTriggerScan }: DashboardProps) {
  const [selectedProvider, setSelectedProvider] = useState('aws');
  const [selectedRegions, setSelectedRegions] = useState(['us-east-1']);

  const severityData = [
    { name: 'Critical', value: 12, color: '#dc2626' },
    { name: 'High', value: 24, color: '#ea580c' },
    { name: 'Medium', value: 42, color: '#f59e0b' },
    { name: 'Low', value: 89, color: '#10b981' },
  ];

  const driftTrendData = [
    { date: '7d ago', drifts: 45 },
    { date: '6d ago', drifts: 52 },
    { date: '5d ago', drifts: 48 },
    { date: '4d ago', drifts: 61 },
    { date: '3d ago', drifts: 55 },
    { date: '2d ago', drifts: 67 },
    { date: 'Today', drifts: 89 },
  ];

  const resourceTypes = [
    { type: 'EC2', count: 234, drifted: 12 },
    { type: 'S3', count: 89, drifted: 5 },
    { type: 'RDS', count: 45, drifted: 8 },
    { type: 'Lambda', count: 156, drifted: 3 },
  ];

  const COLORS = ['#dc2626', '#ea580c', '#f59e0b', '#10b981'];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h2 className="text-3xl font-bold text-gray-900">Dashboard</h2>
        <div className="flex space-x-3">
          <button 
            onClick={onRefresh}
            className="bg-gray-600 text-white px-4 py-2 rounded-lg hover:bg-gray-700 text-sm font-medium"
          >
            ↻ Refresh
          </button>
          <button 
            onClick={() => onTriggerScan(selectedProvider, selectedRegions)}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium"
          >
            + Start Scan
          </button>
        </div>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="text-gray-500 text-sm font-medium">Total Resources</div>
          <div className="text-3xl font-bold text-gray-900 mt-2">524</div>
          <div className="text-xs text-gray-500 mt-2">Across 3 clouds</div>
        </div>
        <div className="bg-white rounded-lg shadow p-6 border-l-4 border-red-600">
          <div className="text-gray-500 text-sm font-medium">Critical Drift</div>
          <div className="text-3xl font-bold text-red-600 mt-2">12</div>
          <div className="text-xs text-gray-500 mt-2">Requires attention</div>
        </div>
        <div className="bg-white rounded-lg shadow p-6 border-l-4 border-yellow-600">
          <div className="text-gray-500 text-sm font-medium">Total Drift</div>
          <div className="text-3xl font-bold text-yellow-600 mt-2">167</div>
          <div className="text-xs text-gray-500 mt-2">Across all severities</div>
        </div>
        <div className="bg-white rounded-lg shadow p-6 border-l-4 border-blue-600">
          <div className="text-gray-500 text-sm font-medium">Compliance Status</div>
          <div className="text-3xl font-bold text-blue-600 mt-2">72%</div>
          <div className="text-xs text-gray-500 mt-2">Of rules passed</div>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Drift by Severity */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Drift Distribution</h3>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={severityData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, value }) => `${name}: ${value}`}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
              >
                {severityData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>

        {/* Drift Trend */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Drift Trend (Last 7 Days)</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={driftTrendData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Line type="monotone" dataKey="drifts" stroke="#3b82f6" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Recent Scans */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">Recent Scans</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Scan ID</th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Cloud Provider</th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Status</th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Drift Found</th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Created</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                    Loading scans...
                  </td>
                </tr>
              ) : scans.length > 0 ? (
                scans.map((scan) => (
                  <tr key={scan.id} className="border-b border-gray-200 hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm font-mono text-gray-900">{scan.id.substring(0, 8)}</td>
                    <td className="px-6 py-4 text-sm text-gray-900 capitalize">{scan.cloud_provider}</td>
                    <td className="px-6 py-4 text-sm">
                      <span className={`px-3 py-1 rounded-full text-xs font-semibold ${
                        scan.status === 'completed' ? 'bg-green-100 text-green-800' :
                        scan.status === 'running' ? 'bg-blue-100 text-blue-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {scan.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm font-semibold text-gray-900">{scan.drift_count}</td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {new Date(scan.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                    No scans found. Start a new scan to begin.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Resource Type Breakdown */}
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Resource Breakdown by Type</h3>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={resourceTypes}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="type" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="count" fill="#3b82f6" name="Total" />
            <Bar dataKey="drifted" fill="#ef4444" name="Drifted" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
