import axios, { AxiosInstance } from 'axios';

const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

const api: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 30000,
});

// Request interceptor – attach JWT if present
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('coherence_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// Response interceptor – normalise errors
api.interceptors.response.use(
  (res) => res,
  (err) => {
    const msg = err.response?.data?.error || err.message || 'Unknown error';
    return Promise.reject(new Error(msg));
  }
);

/* ── Scans ─────────────────────────────────────────────── */
export const scansApi = {
  list: () => api.get('/scans').then((r) => r.data ?? []),
  get: (id: string) => api.get(`/scans/${id}`).then((r) => r.data),
  create: (payload: { cloud_provider: string; regions: string[]; resource_types: string[] }) =>
    api.post('/scans', payload).then((r) => r.data),
  retry: (id: string) => api.post(`/scans/${id}/retry`).then((r) => r.data),
  remove: (id: string) => api.delete(`/scans/${id}`).then((r) => r.data),
};

/* ── Drifts ─────────────────────────────────────────────── */
export const driftsApi = {
  list: (params?: { scan_id?: string; severity?: string; is_resolved?: boolean }) =>
    api.get('/drifts', { params }).then((r) => r.data ?? []),
  get: (id: string) => api.get(`/drifts/${id}`).then((r) => r.data),
  resolve: (id: string) => api.post(`/drifts/${id}/resolve`).then((r) => r.data),
  bulkResolve: (ids: string[]) => api.post('/drifts/bulk-resolve', { drift_ids: ids }).then((r) => r.data),
};

/* ── Remediations ───────────────────────────────────────── */
export const remediationsApi = {
  list: () => api.get('/remediations').then((r) => r.data ?? []),
  get: (id: string) => api.get(`/remediations/${id}`).then((r) => r.data),
  create: (payload: { drift_id: string; action_type: string; dry_run: boolean }) =>
    api.post('/remediations', payload).then((r) => r.data),
  approve: (id: string) => api.post(`/remediations/${id}/approve`).then((r) => r.data),
  reject: (id: string) => api.post(`/remediations/${id}/reject`).then((r) => r.data),
  execute: (id: string) => api.post(`/remediations/${id}/execute`).then((r) => r.data),
  rollback: (id: string) => api.post(`/remediations/${id}/rollback`).then((r) => r.data),
};

/* ── Reports ────────────────────────────────────────────── */
export const reportsApi = {
  list: () => api.get('/reports').then((r) => r.data ?? []),
  get: (id: string) => api.get(`/reports/${id}`).then((r) => r.data),
  generate: (scan_id: string) => api.post('/reports/generate', { scan_id }).then((r) => r.data),
  exportUrl: (id: string, format = 'json') => `${BASE_URL}/reports/${id}/export?format=${format}`,
};

/* ── Health ─────────────────────────────────────────────── */
export const healthApi = {
  check: () => api.get('/health').then((r) => r.data),
};

export default api;
