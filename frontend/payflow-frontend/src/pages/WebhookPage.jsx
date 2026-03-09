import { useState, useEffect, useCallback } from "react";
import { webhookApi } from "../lib/api";
import { formatDate } from "../lib/utils";

// ── Status badge ─────────────────────────────────────────
function DeliveryBadge({ delivered, retryCount }) {
  if (delivered) {
    return <span className="badge-success">✓ Delivered</span>;
  }
  return (
    <span className="badge-failed">
      ✕ Failed {retryCount > 0 ? `(${retryCount} retry)` : ""}
    </span>
  );
}

// ── Stat card ─────────────────────────────────────────────
function StatCard({ label, value, sub, accent }) {
  return (
    <div className="card px-5 py-4">
      <p className="text-slate-subtle text-xs uppercase tracking-wider mb-2">
        {label}
      </p>
      <p
        className={`font-display font-bold text-2xl ${accent || "text-white"}`}
      >
        {value}
      </p>
      {sub && <p className="text-slate-subtle text-xs mt-1">{sub}</p>}
    </div>
  );
}

// ── Payload modal ─────────────────────────────────────────
function PayloadModal({ log, onClose }) {
  if (!log) return null;
  let parsed = null;
  try {
    parsed = JSON.parse(log.payload || "{}");
  } catch {
    parsed = log.payload;
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center px-4"
      onClick={onClose}
    >
      <div className="absolute inset-0 bg-ink/80 backdrop-blur-sm" />
      <div
        className="relative card-glow w-full max-w-lg p-6 animate-fade-up max-h-[80vh] overflow-auto"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-start justify-between mb-5">
          <div>
            <p className="font-display font-semibold text-lg">{log.event}</p>
            <p className="text-slate-subtle text-xs font-mono mt-0.5">
              {log.merchant_name}
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-slate-subtle hover:text-white transition-colors text-xl leading-none"
          >
            ×
          </button>
        </div>

        {/* Meta */}
        <div className="grid grid-cols-2 gap-3 mb-4 text-sm">
          {[
            ["Log ID", `#${log.id}`],
            ["Transaction", `#${log.transaction_id}`],
            [
              "Status",
              log.response_status ? `HTTP ${log.response_status}` : "—",
            ],
            ["Retries", log.retry_count],
            ["Dikirim", formatDate(log.sent_at)],
            ["Delivered", log.is_delivered ? "Ya" : "Tidak"],
          ].map(([k, v]) => (
            <div key={k} className="bg-ink-muted/40 rounded-lg px-3 py-2">
              <p className="text-slate-subtle text-xs">{k}</p>
              <p className="font-mono text-xs text-white mt-0.5">{v}</p>
            </div>
          ))}
        </div>

        {/* Payload */}
        <div>
          <p className="text-xs text-slate-subtle uppercase tracking-wider mb-2">
            Webhook Payload
          </p>
          <pre
            className="bg-ink rounded-xl p-4 text-xs font-mono text-jade overflow-auto max-h-48
                          border border-ink-muted leading-relaxed"
          >
            {JSON.stringify(parsed, null, 2)}
          </pre>
        </div>

        {/* Response body */}
        {log.response_body && (
          <div className="mt-4">
            <p className="text-xs text-slate-subtle uppercase tracking-wider mb-2">
              Merchant Response
            </p>
            <pre
              className="bg-ink rounded-xl p-4 text-xs font-mono text-slate overflow-auto max-h-32
                            border border-ink-muted"
            >
              {log.response_body}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}

// ── Main page ─────────────────────────────────────────────
export default function WebhookPage() {
  const [logs, setLogs] = useState([]);
  const [stats, setStats] = useState(null);
  const [merchants, setMerchants] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [selectedLog, setSelectedLog] = useState(null);
  const [filterMerchant, setFilterMerchant] = useState("all");
  const [filterStatus, setFilterStatus] = useState("all");
  const LIMIT = 20;

  const fetchAll = useCallback(async () => {
    try {
      const [logsRes, statsRes, merchantsRes] = await Promise.all([
        webhookApi.getLogs(page, LIMIT),
        webhookApi.getStats(),
        webhookApi.getMerchants(),
      ]);
      setLogs(logsRes.data.data?.data || []);
      setTotal(logsRes.data.data?.total || 0);
      setStats(statsRes.data.data);
      setMerchants(merchantsRes.data.data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  // Auto-refresh setiap 5 detik
  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(fetchAll, 5000);
    return () => clearInterval(interval);
  }, [autoRefresh, fetchAll]);

  // Filter logs di client-side
  const filteredLogs = logs.filter((l) => {
    if (filterMerchant !== "all" && l.merchant_id !== parseInt(filterMerchant))
      return false;
    if (filterStatus === "delivered" && !l.is_delivered) return false;
    if (filterStatus === "failed" && l.is_delivered) return false;
    return true;
  });

  const totalPages = Math.ceil(total / LIMIT);

  return (
    <div className="space-y-6">
      {/* ── Header ── */}
      <div className="flex items-start justify-between animate-fade-up">
        <div>
          <h1 className="font-display font-bold text-3xl">Webhook Panel</h1>
          <p className="text-slate-subtle text-sm mt-1">
            Monitor pengiriman notifikasi ke merchant
          </p>
        </div>
        <div className="flex items-center gap-3">
          {/* Auto-refresh toggle */}
          <button
            onClick={() => setAutoRefresh((v) => !v)}
            className={`flex items-center gap-2 px-3 py-2 rounded-xl border text-xs font-medium transition-all
              ${
                autoRefresh
                  ? "border-jade/40 bg-jade/10 text-jade"
                  : "border-ink-muted bg-ink-soft text-slate-subtle"
              }`}
          >
            <span
              className={`w-1.5 h-1.5 rounded-full ${autoRefresh ? "bg-jade animate-ping" : "bg-slate-subtle"}`}
            />
            {autoRefresh ? "Live" : "Paused"}
          </button>
          <button
            onClick={fetchAll}
            className="btn-secondary text-xs px-4 py-2"
          >
            ↻ Refresh
          </button>
        </div>
      </div>

      {/* ── Stats ── */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 animate-fade-up delay-100">
          <StatCard label="Total Dikirim" value={stats.total} />
          <StatCard
            label="Delivered"
            value={stats.delivered}
            accent="text-jade"
          />
          <StatCard label="Failed" value={stats.failed} accent="text-danger" />
          <StatCard
            label="Success Rate"
            value={`${stats.success_rate?.toFixed(1)}%`}
            accent={stats.success_rate >= 90 ? "text-jade" : "text-amber-pay"}
            sub={`${stats.last_hour} dalam 1 jam terakhir`}
          />
        </div>
      )}

      {/* ── Filters ── */}
      <div className="flex items-center gap-3 animate-fade-up delay-200">
        {/* Merchant filter */}
        <select
          value={filterMerchant}
          onChange={(e) => setFilterMerchant(e.target.value)}
          className="input text-sm w-auto pr-8 cursor-pointer"
          style={{ width: "auto" }}
        >
          <option value="all">Semua Merchant</option>
          {merchants.map((m) => (
            <option key={m.id} value={m.id}>
              {m.merchant_name}
            </option>
          ))}
        </select>

        {/* Status filter */}
        <div className="flex rounded-xl border border-ink-muted overflow-hidden">
          {[
            { val: "all", label: "Semua" },
            { val: "delivered", label: "Delivered" },
            { val: "failed", label: "Failed" },
          ].map(({ val, label }) => (
            <button
              key={val}
              onClick={() => setFilterStatus(val)}
              className={`px-4 py-2 text-xs font-medium transition-colors
                ${
                  filterStatus === val
                    ? "bg-jade/15 text-jade"
                    : "text-slate-subtle hover:text-white"
                }`}
            >
              {label}
            </button>
          ))}
        </div>

        <span className="text-slate-subtle text-xs ml-auto">
          {filteredLogs.length} ditampilkan
        </span>
      </div>

      {/* ── Log table ── */}
      <div className="animate-fade-up delay-300">
        {loading ? (
          <div className="space-y-2">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="card px-4 py-4 flex gap-4 animate-pulse">
                <div className="w-8 h-8 rounded-lg bg-ink-muted flex-shrink-0" />
                <div className="flex-1 space-y-2">
                  <div className="h-4 bg-ink-muted rounded w-1/4" />
                  <div className="h-3 bg-ink-muted rounded w-1/3" />
                </div>
              </div>
            ))}
          </div>
        ) : filteredLogs.length === 0 ? (
          <div className="card p-12 text-center">
            <p className="text-3xl mb-3">⬡</p>
            <p className="font-display font-semibold">Belum ada webhook</p>
            <p className="text-slate-subtle text-sm mt-1">
              Lakukan pembayaran untuk melihat webhook dikirim ke merchant
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {filteredLogs.map((log, idx) => (
              <button
                key={log.id}
                onClick={() => setSelectedLog(log)}
                className="card w-full px-4 py-3.5 flex items-center gap-4 text-left
                           hover:border-jade/25 transition-all duration-200 group"
                style={{ animationDelay: `${idx * 0.03}s` }}
              >
                {/* Delivery status dot */}
                <div
                  className={`w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 text-sm
                  ${
                    log.is_delivered
                      ? "bg-jade/10 text-jade border border-jade/20"
                      : "bg-danger/10 text-danger border border-danger/20"
                  }`}
                >
                  {log.is_delivered ? "✓" : "✕"}
                </div>

                {/* Event + merchant */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="font-mono text-xs text-jade font-medium">
                      {log.event}
                    </p>
                    <span className="text-ink-muted">·</span>
                    <p className="text-sm text-white">{log.merchant_name}</p>
                  </div>
                  <p className="text-slate-subtle text-xs mt-0.5 font-mono">
                    TX #{log.transaction_id} · Log #{log.id}
                  </p>
                </div>

                {/* Retry */}
                {log.retry_count > 0 && (
                  <div className="badge-pending text-xs">
                    {log.retry_count} retry
                  </div>
                )}

                {/* HTTP status */}
                <div className="text-center flex-shrink-0 min-w-[52px]">
                  {log.response_status ? (
                    <span
                      className={`font-mono text-sm font-semibold
                      ${
                        log.response_status < 300
                          ? "text-jade"
                          : log.response_status < 500
                            ? "text-amber-pay"
                            : "text-danger"
                      }`}
                    >
                      {log.response_status}
                    </span>
                  ) : (
                    <span className="text-slate-subtle text-xs">—</span>
                  )}
                  <p className="text-slate-subtle text-xs">HTTP</p>
                </div>

                {/* Time */}
                <div className="text-right flex-shrink-0">
                  <p className="text-xs text-slate-subtle">
                    {formatDate(log.sent_at)}
                  </p>
                  <DeliveryBadge
                    delivered={log.is_delivered}
                    retryCount={log.retry_count}
                  />
                </div>

                {/* Arrow */}
                <span className="text-slate-subtle group-hover:text-jade transition-colors text-sm">
                  →
                </span>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* ── Pagination ── */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
            className="btn-secondary px-4 py-2 text-sm"
          >
            ←
          </button>
          <span className="text-slate-subtle text-sm font-mono">
            {page} / {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
            className="btn-secondary px-4 py-2 text-sm"
          >
            →
          </button>
        </div>
      )}

      {/* ── How it works info ── */}
      <div className="card p-5 border-jade/10 animate-fade-up">
        <p className="font-display font-semibold text-sm mb-3 text-jade">
          ⬡ Cara Kerja Webhook
        </p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-xs text-slate-subtle">
          {[
            [
              "1. Pembayaran Sukses",
              "Setelah ACID transaction commit, PayFlow langsung dispatch webhook ke URL merchant secara async (non-blocking).",
            ],
            [
              "2. Retry Otomatis",
              "Jika merchant endpoint gagal merespons, sistem retry dengan exponential backoff: 1s → 2s → 4s (maks 3x).",
            ],
            [
              "3. HMAC Signature",
              "Setiap request ditandai header X-Payflow-Signature (HMAC-SHA256). Merchant bisa verifikasi keaslian webhook.",
            ],
          ].map(([title, desc]) => (
            <div key={title}>
              <p className="text-white font-medium mb-1">{title}</p>
              <p>{desc}</p>
            </div>
          ))}
        </div>
      </div>

      {/* ── Payload modal ── */}
      {selectedLog && (
        <PayloadModal log={selectedLog} onClose={() => setSelectedLog(null)} />
      )}
    </div>
  );
}
