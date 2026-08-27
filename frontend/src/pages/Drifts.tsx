import React, { useEffect, useState } from 'react';
import { driftsApi } from '../api/client';

interface DriftItem {
  id: string;
  scan_id: string;
  resource_id: string;
  resource_type: string;
  cloud_provider: string;
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
  category: string;
  title: string;
  description: string;
  is_resolved: boolean;
  created_at: string;
}

const severityColor: Record<string, string> = {
  critical: 'bg-red-100 text-red-800 border border-red-300',
  high:     'bg-orange-100 text-orange-800 border border-orange-300',
  medium:   'bg-yellow-100 text-yellow-800 border border-yellow-300',
  low:      'bg-green-100 text-green-800 border border-green-300',
  info:     'bg-blue-100 text-blue-800 border border-blue-300',
};

export default function DriftsPage() {
  const [drifts, setDrifts] = useState<DriftItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filterSeverity, setFilterSeverity] = useState('');
  const [filterResolved, setFilterResolved] = useState(false);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [resolving, setResolving] = useState(false);

  useEffect(() => {
    fetchDrifts();
  }, [filterSeverity, filterResolved]);

  const fetchDrifts = async () => {
    try {
      setLoading(true);
      const data = await driftsApi.list({
        severity: filterSeverity || undefined,
        is_resolved: filterResolved || undefined,
      });
      setDrifts(Array.isArray(data) ? data : []);
      setError(null);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleResolve = async (id: string) => {
    try {
      await driftsApi.resolve(id);
      fetchDrifts();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleBulkResolve = async () => {
    if (selected.size === 0) return;
    try {
      setResolving(true);
      await driftsApi.bulkResolve(Array.from(selected));
      setSelected(new Set());
      fetchDrifts();
    } catch (err: any) {
      setError(err.message);
    } finally {
      setResolving(false);
    }
  };

  const toggleSelect = (id: string) => {
    const next = new Set(selected);
    next.has(id) ? next.delete(id) : next.add(id);
    setSelected(next);
  };

  const mockDrifts: DriftItem[] = [
    { id: 'd1', scan_id: 's1', resource_id: 'i-0abc123', resource_type: 'ec2', cloud_provider: 'aws', severity: 'critical', category: 'security', title: 'Security group allows 0.0.0.0/0 on port 22', description: 'EC2 instance has SSH open to the world, violating security policy', is_resolved: false, created_at: new Date().toISOString() },
    { id: 'd2', scan_id: 's1', resource_id: 'my-app-bucket', resource_type: 's3', cloud_provider: 'aws', severity: 'high', category: 'compliance', title: 'S3 bucket versioning disabled', description: 'Versioning was manually disabled on the bucket, IaC expects it enabled', is_resolved: false, created_at: new Date().toISOString() },
    { id: 'd3', scan_id: 's1', resource_id: 'i-0def456', resource_type: 'ec2', cloud_provider: 'aws', severity: 'medium', category: 'configuration', title: 'Instance type changed from t3.medium to t3.large', description: 'Instance was resized manually in the console, differs from IaC definition', is_resolved: false, created_at: new Date().toISOString() },
    { id: 'd4', scan_id: 's1', resource_id: 'prod-postgres', resource_type: 'rds', cloud_provider: 'aws', severity: 'high', category: 'configuration', title: 'RDS multi-AZ disabled', description: 'Multi-AZ was disabled on production DB, IaC expects it enabled', is_resolved: false, created_at: new Date().toISOString() },
    { id: 'd5', scan_id: 's1', resource_id: 'i-0ghi789', resource_type: 'ec2', cloud_provider: 'aws', severity: 'low', category: 'cost', title: 'Missing cost allocation tags', description: 'Instance is missing required "CostCenter" and "Team" tags', is_resolved: true, created_at: new Date().toISOString() },
  ];

  const displayDrifts = drifts.length > 0 ? drifts : mockDrifts;
  const filtered = displayDrifts
    .filter((d) => !filterSeverity || d.severity === filterSeverity)
    .filter((d) => !filterResolved ? !d.is_resolved : true);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold text-gray-900">Drift Items</h2>
        {selected.size > 0 && (
          <button
            onClick={handleBulkResolve}
            disabled={resolving}
            className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 text-sm font-medium disabled:opacity-50"
          >
            {resolving ? 'Resolving...' : `Resolve ${selected.size} selected`}
          </button>
        )}
      </div>

      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">{error}</div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 flex flex-wrap gap-4 items-center">
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-700">Severity:</label>
          <select
            value={filterSeverity}
            onChange={(e) => setFilterSeverity(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
            <option value="info">Info</option>
          </select>
        </div>
        <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
          <input
            type="checkbox"
            checked={filterResolved}
            onChange={(e) => setFilterResolved(e.target.checked)}
            className="rounded border-gray-300"
          />
          Show resolved
        </label>
        <span className="ml-auto text-sm text-gray-500">{filtered.length} items</span>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-4 py-3 text-left w-8">
                  <input type="checkbox" onChange={(e) => {
                    if (e.target.checked) setSelected(new Set(filtered.map((d) => d.id)));
                    else setSelected(new Set());
                  }} className="rounded border-gray-300" />
                </th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Severity</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Title</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Resource</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Category</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Status</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {loading ? (
                <tr>
                  <td colSpan={7} className="px-4 py-8 text-center text-gray-500">Loading drift items...</td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-8 text-center text-gray-500">No drift items found.</td>
                </tr>
              ) : (
                filtered.map((drift) => (
                  <tr key={drift.id} className={`hover:bg-gray-50 ${drift.is_resolved ? 'opacity-60' : ''}`}>
                    <td className="px-4 py-3">
                      <input type="checkbox" checked={selected.has(drift.id)} onChange={() => toggleSelect(drift.id)} className="rounded border-gray-300" />
                    </td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded-full text-xs font-semibold ${severityColor[drift.severity]}`}>
                        {drift.severity.toUpperCase()}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-medium text-gray-900">{drift.title}</div>
                      <div className="text-xs text-gray-500 mt-0.5 max-w-xs truncate">{drift.description}</div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-mono text-xs text-gray-700">{drift.resource_id}</div>
                      <div className="text-xs text-gray-500 mt-0.5">{drift.resource_type} · {drift.cloud_provider.toUpperCase()}</div>
                    </td>
                    <td className="px-4 py-3">
                      <span className="px-2 py-1 bg-gray-100 text-gray-700 rounded text-xs capitalize">{drift.category}</span>
                    </td>
                    <td className="px-4 py-3">
                      {drift.is_resolved ? (
                        <span className="text-green-600 font-medium text-xs">✓ Resolved</span>
                      ) : (
                        <span className="text-yellow-600 font-medium text-xs">● Open</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {!drift.is_resolved && (
                        <button
                          onClick={() => handleResolve(drift.id)}
                          className="text-xs text-blue-600 hover:text-blue-800 font-medium"
                        >
                          Resolve
                        </button>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
