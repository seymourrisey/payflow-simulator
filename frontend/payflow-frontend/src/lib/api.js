import axios from "axios";
const apiUrl = import.meta.env.API_BASE_URL;

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: { "Content-Type": "application/json" },
  timeout: 10000,
});

// Helper decode JWT expiry
const isTokenExpired = (token) => {
  if (!token) return true;
  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    return payload.exp * 1000 < Date.now();
  } catch {
    return true;
  }
};

// Request interceptor: inject JWT + cek expiry sebelum request
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("payflow_token");

  // Kalau token expired → hapus dan redirect ke login
  if (token && isTokenExpired(token)) {
    localStorage.removeItem("payflow_token");
    localStorage.removeItem("payflow_user");
    window.location.href = "/login";
    return Promise.reject(new Error("Token expired"));
  }

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: handle 401 globally
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("payflow_token");
      localStorage.removeItem("payflow_user");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  },
);

// auth endpoints
export const authApi = {
  register: (data) => api.post("/auth/register", data),
  login: (data) => api.post("/auth/login", data),
  logout: () => api.post("/auth/logout"), // ← tambah ini
};

// wallet endpoints
export const walletApi = {
  getWallet: () => api.get("/wallet"),
  topUp: (data) => api.post("/wallet/topup", data),
};

// payment endpoints
export const paymentApi = {
  generateQR: (data) => api.post("/payment/qr", data),

  // payment endpoints
  pay: (data, idempotencyKey) =>
    api.post("/payment/pay", data, {
      headers: { "X-Idempotency-Key": idempotencyKey },
    }),
};

// transaction endpoints
export const transactionApi = {
  getHistory: (page = 1, limit = 10) =>
    api.get(`/transactions?page=${page}&limit=${limit}`),
};

export const webhookApi = {
  getLogs: (page = 1, limit = 20) =>
    api.get(`/webhooks?page=${page}&limit=${limit}`),
  getStats: () => api.get("/webhooks/stats"),
  getMerchants: () => api.get("/webhooks/merchants"),
};

export default api;
