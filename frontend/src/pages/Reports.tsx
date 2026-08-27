import React, { useEffect, useState } from 'react';
import { reportsApi, scansApi } from '../api/client';
import { RadarChart, PolarGrid, PolarAngleAxis, Radar, ResponsiveContainer, Tooltip } from 'recharts';

const complianceData = [
  { subject: 'Encryption', A: 90, fullMark: 100 },
  { subject: 'Access Ctrl', A: 68, fullMark: 100 },
  { subject: 'Tagging', A: 55, fullMark: 100 },
  { subject: 'Backup', A: 82, fullMark: 100 },
  { subject: 'Networking', A: 72, fullMark: 100 },
  { subject: 'Logging', A: 95, fullMark: 100 },
];

export default function ReportsPage() {
  const [reports, setReports] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);

  useEffect(() => {
    fetchReports();
  }, []);

  const fetchReports = async () => {
    try {
      setLoading(true);
      const data = await reportsApi.list();
      setReports(Array.isArray(data) ? data : []);
    } catch {
      setReports([]);
    } finally {
      setLoading(false);
    }
  };

  const generateReport = async () => {
    try {
      setGenerating(true);
      const scans = await scansApi.list();
      if (scans && scans.length > 0) {
        await reportsApi.generate(scans[0].id);
        fetchReports();
      }
    } catch {
      // fallback silently
    } finally {
      setGenerating(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold text-gray-900">Reports</h2>
        <button
          onClick={generateReport}
          disabled={generating}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium disabled:opacity-50"
        >
          {generating ? 'Generating...' : '+ Generate Report'}
        </button>
      </div>

      {/* Compliance radar */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Compliance Score by Domain</h3>
          <ResponsiveContainer width="100%" height={300}>
            <RadarChart data={complianceData}>
              <PolarGrid />
              <PolarAngleAxis dataKey="subject" />
              <Radar name="Score" dataKey="A" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.3} />
              <Tooltip />
            </RadarChart>
          </ResponsiveContainer>
        </div>

        {/* Overall compliance summary */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Overall Compliance Summary</h3>
          <div className="space-y-4 mt-2">
            {[
              { rule: 'S3 Bucket Encryption Required', passed: true },
              { rule: 'RDS Multi-AZ in Production', passed: false },
              { rule: 'EC2 IMDSv2 Enforced', passed: true },
              { rule: 'CloudTrail Logging Enabled', passed: true },
              { rule: 'VPC Flow Logs Enabled', passed: false },
              { rule: 'IAM Password Policy Enforced', passed: true },
            ].map((item, i) => (
              <div key={i} className="flex items-center justify-between py-2 border-b border-gray-100">
                <span className="text-sm text-gray-700">{item.rule}</span>
                <span className={`text-sm font-semibold ${item.passed ? 'text-green-600' : 'text-red-600'}`}>
                  {item.passed ? '✓ Pass' : '✗ Fail'}
                </span>
              </div>
            ))}
          </div>
          <div className="mt-4 pt-4 border-t border-gray-200">
            <div className="flex justify-between text-sm font-semibold">
              <span className="text-gray-700">Overall Score</span>
              <span className="text-blue-600">4/6 (67%)</span>
            </div>
          </div>
        </div>
      </div>

      {/* Reports list */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">Generated Reports</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Report ID</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Cloud</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Total Resources</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Drifted</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Drift %</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Cost Impact</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Created</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-900">Export</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {loading ? (
                <tr><td colSpan={8} className="px-4 py-8 text-center text-gray-500">Loading...</td></tr>
              ) : reports.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-gray-500">
                    No reports yet. Generate one from a completed scan.
                  </td>
                </tr>
              ) : (
                reports.map((r) => (
                  <tr key={r.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-xs text-gray-700">{r.id}</td>
                    <td className="px-4 py-3 text-xs capitalize">{r.cloud_provider}</td>
                    <td className="px-4 py-3 text-xs">{r.total_resources}</td>
                    <td className="px-4 py-3 text-xs text-red-600 font-semibold">{r.drifted_resources}</td>
                    <td className="px-4 py-3 text-xs">{r.drift_percentage?.toFixed(1)}%</td>
                    <td className="px-4 py-3 text-xs text-orange-600 font-semibold">${r.cost_impact?.toFixed(2)}</td>
                    <td className="px-4 py-3 text-xs text-gray-500">{new Date(r.created_at).toLocaleDateString()}</td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        <a href={reportsApi.exportUrl(r.id, 'json')} target="_blank" rel="noreferrer" className="text-xs text-blue-600 hover:underline">JSON</a>
                        <a href={reportsApi.exportUrl(r.id, 'csv')} target="_blank" rel="noreferrer" className="text-xs text-blue-600 hover:underline">CSV</a>
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
