// Format angka ke format Rupiah
export const formatIDR = (amount) => {
  if (amount === null || amount === undefined) return "Rp 0";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
};

// Format tanggal ke bahasa Indonesia
export const formatDate = (dateStr) => {
  const date = new Date(dateStr);
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
};

// Hitung sisa waktu countdown dari expiry timestamp
export const getTimeLeft = (expiredAt) => {
  const diff = new Date(expiredAt) - new Date();
  if (diff <= 0) return null;
  const minutes = Math.floor(diff / 60000);
  const seconds = Math.floor((diff % 60000) / 1000);
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
};

// Transaction type label in Bahasa
export const txTypeLabel = {
  PAYMENT: "Pembayaran",
  TOPUP: "Top Up",
  TRANSFER: "Transfer",
};

export const txTypeIcon = {
  PAYMENT: "↑",
  TOPUP: "↓",
  TRANSFER: "⇄",
};
