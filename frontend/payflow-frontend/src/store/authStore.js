import { create } from "zustand";

// Helper: cek apakah JWT token sudah expired
const isTokenExpired = (token) => {
  if (!token) return true;
  try {
    // Decode payload JWT (bagian tengah, base64)
    const payload = JSON.parse(atob(token.split(".")[1]));
    // exp adalah Unix timestamp dalam detik
    return payload.exp * 1000 < Date.now();
  } catch {
    return true; // Kalau decode gagal, anggap expired
  }
};

// Cek token saat app load — kalau expired, langsung hapus
const storedToken = localStorage.getItem("payflow_token");
if (storedToken && isTokenExpired(storedToken)) {
  localStorage.removeItem("payflow_token");
  localStorage.removeItem("payflow_user");
}

export const useAuthStore = create((set) => ({
  token: !isTokenExpired(localStorage.getItem("payflow_token"))
    ? localStorage.getItem("payflow_token")
    : null,
  user: !isTokenExpired(localStorage.getItem("payflow_token"))
    ? JSON.parse(localStorage.getItem("payflow_user") || "null")
    : null,

  setAuth: (token, user) => {
    localStorage.setItem("payflow_token", token);
    localStorage.setItem("payflow_user", JSON.stringify(user));
    set({ token, user });
  },

  logout: () => {
    localStorage.removeItem("payflow_token");
    localStorage.removeItem("payflow_user");
    set({ token: null, user: null });
  },

  // Cek apakah user benar-benar authenticated (token ada DAN tidak expired)
  isAuthenticated: () => {
    const token = localStorage.getItem("payflow_token");
    return !!token && !isTokenExpired(token);
  },
}));
