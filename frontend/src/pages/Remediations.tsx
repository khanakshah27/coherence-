import React, { useEffect, useState } from 'react';
import { remediationsApi } from '../api/client';

interface Remediation {
  id: string;
  drift_id: string;
  action_type: string;
  status: string;
  approval_status: string;
  dry_run: boolean;
  approved_by?: string;
  created_at: string;
}

const mockRemediations: Remediation[] = [
  { id: 'r1', drift_id: 'd2', action_type: 'apply_iac', status: 'pending', approval_status: 'pending', dry_run: true, created_at: new Date().toISOString() },
  { id: 'r2', drift_id: 'd3', action_type: 'safe_remediate', status: 'success', approval_status: 'approved', dry_run: false, approved_by: 'arn:aws:iam::123:user/alice', created_at: new Date(Date.now() - 3600000).toISOString() },
  { id: 'r3', drift_id: 'd4', action_type: 'apply_iac', status: 'failed', approval_status: 'approved', dry_run: false, approved_by: 'arn:aws:iam::123:user/bob', created_at: new Date(Date.now() - 7200000).toISOString() },
];

const statusColor: Record<string, string> = {
  pending:   'bg-yellow-100 text-yellow-800',
  executing: 'bg-blue-100 text-blue-800',
  success:   'bg-green-100 text-green-800',
  failed:    'bg-red-100 text-red-800',
  rolled_back: 'bg-gray-100 text-gray-800',
};

export default function RemediationsPage() {
  const [items, setItems] = useState<Remediation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchRemediations();
  }, []);

  const fetchRemediations = async () => {
    try {
      setLoading(true);
      const data = await remediationsApi.list();
      setItems(Array.isArray(data) && data.length > 0 ? data : mockRemediations);
      setError(null);
    } catch {
      setItems(mockRemediations);
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = async (id: string) => {
    try {
      await remediationsApi.approve(id);
      fetchRemediations();
    } catch (err: any) { setError(err.message); }
  };

  const handleExecute = async (id: string) => {
    try {
      await remediationsApi.execute(id);
      fetchRemediations();
    } catch (err: any) { setError(err.message); }
  };

  const handleRollback = async (id: string) => {
    try {
      await remediationsApi.rollback(id);
      fetchRemediations();
    } catch (err: any) { setError(err.message); }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold text-gray-900">Remediations</h2>
        <button
          onClick={fetchRemediations}
          className="bg-gray-600 text-white px-4 py-2 rounded-lg hover:bg-gray-700 text-sm font-medium"
        >
          ↻ Refresh
        </button>
      </div>

      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">{error}</div>
      )}

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: 'Pending Approval', count: items.filter(i => i.approval_status === 'pending').length, color: 'text-yellow-600' },
          { label: 'Executing', count: items.filter(i => i.status === 'executing').length, color: 'text-blue-600' },
          { label: 'Succeeded', count: items.filter(i => i.status === 'success').length, color: 'text-green-600' },
          { label: 'Failed', count: items.filter(i => i.status === 'failed').length, color: 'text-red-600' },
        ].map((s) => (
          <div key={s.label} className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{s.label}</div>
            <div className={`text-2xl font-bold mt-1 ${s.color}`}>{s.count}</div>
          </div>
        ))}
      </div>

      {/* Remediation list */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">ID</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Drift</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Action</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Status</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Approval</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Dry Run</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Created</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {loading ? (
                <tr><td colSpan={8} className="px-4 py-8 text-center text-gray-500">Loading...</td></tr>
              ) : items.length === 0 ? (
                <tr><td colSpan={8} className="px-4 py-8 text-center text-gray-500">No remediations yet.</td></tr>
              ) : (
                items.map((rem) => (
                  <tr key={rem.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-xs text-gray-700">{rem.id}</td>
                    <td className="px-4 py-3 font-mono text-xs text-gray-700">{rem.drift_id}</td>
                    <td className="px-4 py-3">
                      <span className="px-2 py-1 bg-purple-100 text-purple-800 rounded text-xs font-medium capitalize">
                        {rem.action_type.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded-full text-xs font-semibold capitalize ${statusColor[rem.status] ?? 'bg-gray-100 text-gray-800'}`}>
                        {rem.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-xs font-medium capitalize ${rem.approval_status === 'approved' ? 'text-green-600' : rem.approval_status === 'rejected' ? 'text-red-600' : 'text-yellow-600'}`}>
                        {rem.approval_status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-600">{rem.dry_run ? '✓ Yes' : '✗ No'}</td>
                    <td className="px-4 py-3 text-xs text-gray-500">
                      {new Date(rem.created_at).toLocaleString()}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        {rem.approval_status === 'pending' && (
                          <button onClick={() => handleApprove(rem.id)} className="text-xs text-green-600 hover:text-green-800 font-medium">Approve</button>
                        )}
                        {rem.approval_status === 'approved' && rem.status === 'pending' && (
                          <button onClick={() => handleExecute(rem.id)} className="text-xs text-blue-600 hover:text-blue-800 font-medium">Execute</button>
                        )}
                        {rem.status === 'success' && (
                          <button onClick={() => handleRollback(rem.id)} className="text-xs text-red-600 hover:text-red-800 font-medium">Rollback</button>
                        )}
                      </div>
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
