import { useEffect } from "react";
import { Link } from "react-router-dom";
import { useWalletStore } from "../store/walletStore";
import { useAuthStore } from "../store/authStore";
import { transactionApi } from "../lib/api";
import { formatIDR, formatDate, txTypeLabel, txTypeIcon } from "../lib/utils";
import { useState } from "react";

export default function DashboardPage() {
  const { user } = useAuthStore();
  const { wallet, fetchWallet, loading } = useWalletStore();
  const [recentTx, setRecentTx] = useState([]);

  useEffect(() => {
    fetchWallet();
    transactionApi
      .getHistory(1, 5)
      .then((res) => {
        setRecentTx(res.data.data?.data || []);
      })
      .catch(() => {});
  }, []);

  const hour = new Date().getHours();
  const greeting =
    hour < 12 ? "Selamat pagi" : hour < 17 ? "Selamat siang" : "Selamat malam";

  return (
    <div className="space-y-6">
      {/* ── Header ── */}
      <div className="animate-fade-up">
        <p className="text-slate-subtle text-sm">{greeting},</p>
        <h1 className="font-display font-bold text-3xl mt-0.5">
          {user?.full_name?.split(" ")[0] || "User"} 👋
        </h1>
      </div>

      {/* ── Balance Card ── */}
      <div
        className="animate-fade-up delay-100 relative overflow-hidden rounded-2xl p-6"
        style={{
          background:
            "linear-gradient(135deg, #0D1F17 0%, #0A1A12 50%, #0A0A0F 100%)",
          border: "1px solid rgba(0, 229, 160, 0.2)",
          boxShadow:
            "0 0 60px rgba(0, 229, 160, 0.08), inset 0 1px 0 rgba(0, 229, 160, 0.1)",
        }}
      >
        {/* Decorative glow */}
        <div
          className="absolute top-0 right-0 w-40 h-40 rounded-full blur-3xl"
          style={{
            background: "rgba(0, 229, 160, 0.07)",
            transform: "translate(20%, -30%)",
          }}
        />

        <div className="relative">
          <p className="text-slate-subtle text-xs uppercase tracking-widest mb-2">
            Total Saldo
          </p>
          {loading ? (
            <div className="h-10 w-48 bg-ink-muted rounded-lg animate-pulse" />
          ) : (
            <p className="balance-text text-4xl text-white">
              {formatIDR(wallet?.balance || 0)}
            </p>
          )}
          <p className="text-jade/60 text-sm mt-1.5 font-mono">
            {wallet?.currency || "IDR"} · Wallet #{wallet?.id || "—"}
          </p>
        </div>

        {/* Quick actions */}
        <div className="flex gap-3 mt-6">
          <Link
            to="/topup"
            className="btn-primary text-sm px-5 py-2.5 flex-1 text-center"
          >
            ↓ Top Up
          </Link>
          <Link
            to="/pay"
            className="btn-pay text-sm px-5 py-2.5 flex-1 text-center"
          >
            ⬢ Bayar
          </Link>
        </div>
      </div>

      {/* ── Quick Stats ── */}
      <div className="grid grid-cols-3 gap-4 animate-fade-up delay-200">
        {[
          {
            label: "Total Transaksi",
            value: recentTx.length > 0 ? `${recentTx.length}+` : "—",
            icon: "⬡",
          },
          { label: "Status", value: "AKTIF", icon: "◉" },
          { label: "Mata Uang", value: "IDR", icon: "◈" },
        ].map(({ label, value, icon }) => (
          <div key={label} className="card px-4 py-4">
            <p className="text-slate-subtle text-xs mb-2">{label}</p>
            <div className="flex items-center gap-2">
              <span className="text-jade text-sm">{icon}</span>
              <span className="font-display font-semibold text-sm">
                {value}
              </span>
            </div>
          </div>
        ))}
      </div>

      {/* ── Recent Transactions ── */}
      <div className="animate-fade-up delay-300">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-display font-semibold text-lg">
            Transaksi Terbaru
          </h2>
          <Link to="/history" className="text-jade text-sm hover:underline">
            Lihat semua →
          </Link>
        </div>

        {recentTx.length === 0 ? (
          <div className="card p-8 text-center">
            <p className="text-slate-subtle text-sm">Belum ada transaksi</p>
            <p className="text-slate-subtle/60 text-xs mt-1">
              Coba top up dulu atau lakukan pembayaran
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {recentTx.map((tx) => (
              <div
                key={tx.reference_id}
                className="card px-4 py-3.5 flex items-center gap-4"
              >
                {/* Icon */}
                <div
                  className={`w-9 h-9 rounded-xl flex items-center justify-center text-sm
                  ${tx.type === "TOPUP" ? "bg-jade/10 text-jade" : "bg-amber-pay/10 text-amber-pay"}`}
                >
                  {txTypeIcon[tx.type]}
                </div>
                {/* Info */}
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-sm">{txTypeLabel[tx.type]}</p>
                  <p className="text-slate-subtle text-xs font-mono truncate">
                    {tx.reference_id}
                  </p>
                </div>
                {/* Amount + status */}
                <div className="text-right">
                  <p
                    className={`balance-text text-sm font-semibold
                    ${tx.type === "TOPUP" ? "text-jade" : "text-white"}`}
                  >
                    {tx.type === "TOPUP" ? "+" : "-"}
                    {formatIDR(tx.amount)}
                  </p>
                  <span
                    className={`text-xs ${
                      tx.status === "SUCCESS"
                        ? "badge-success"
                        : tx.status === "PENDING"
                          ? "badge-pending"
                          : "badge-failed"
                    }`}
                  >
                    {tx.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
