import { useState, useEffect } from "react";
import { transactionApi } from "../lib/api";
import { formatIDR, formatDate, txTypeLabel, txTypeIcon } from "../lib/utils";

export default function HistoryPage() {
  const [transactions, setTransactions] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const LIMIT = 10;

  useEffect(() => {
    fetchHistory();
  }, [page]);

  const fetchHistory = async () => {
    setLoading(true);
    try {
      const res = await transactionApi.getHistory(page, LIMIT);
      setTransactions(res.data.data?.data || []);
      setTotal(res.data.data?.total || 0);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const totalPages = Math.ceil(total / LIMIT);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-display font-bold text-3xl">Riwayat</h1>
          <p className="text-slate-subtle text-sm mt-1">
            {total} transaksi total
          </p>
        </div>
      </div>

      {/* ── Transaction list ── */}
      {loading ? (
        <div className="space-y-2">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="card px-4 py-4 flex gap-4 animate-pulse">
              <div className="w-9 h-9 rounded-xl bg-ink-muted flex-shrink-0" />
              <div className="flex-1 space-y-2">
                <div className="h-4 bg-ink-muted rounded w-1/3" />
                <div className="h-3 bg-ink-muted rounded w-1/2" />
              </div>
              <div className="space-y-2">
                <div className="h-4 bg-ink-muted rounded w-20" />
                <div className="h-3 bg-ink-muted rounded w-14" />
              </div>
            </div>
          ))}
        </div>
      ) : transactions.length === 0 ? (
        <div className="card p-12 text-center">
          <p className="text-4xl mb-4">⬡</p>
          <p className="font-display font-semibold">Belum ada transaksi</p>
          <p className="text-slate-subtle text-sm mt-1">
            Lakukan top up atau pembayaran pertamamu
          </p>
        </div>
      ) : (
        <div className="space-y-2 animate-fade-up">
          {transactions.map((tx, idx) => (
            <div
              key={tx.reference_id}
              className="card px-4 py-4 flex items-center gap-4 hover:border-jade/20 transition-colors duration-200"
              style={{ animationDelay: `${idx * 0.05}s` }}
            >
              {/* Type icon */}
              <div
                className={`w-10 h-10 rounded-xl flex items-center justify-center text-sm flex-shrink-0
                ${
                  tx.type === "TOPUP"
                    ? "bg-jade/10 text-jade border border-jade/20"
                    : "bg-amber-pay/10 text-amber-pay border border-amber-pay/20"
                }`}
              >
                {txTypeIcon[tx.type]}
              </div>

              {/* Info */}
              <div className="flex-1 min-w-0">
                <p className="font-medium text-sm">{txTypeLabel[tx.type]}</p>
                <p className="text-slate-subtle text-xs font-mono truncate mt-0.5">
                  {tx.reference_id}
                </p>
                {tx.description && (
                  <p className="text-slate-subtle text-xs mt-0.5 truncate">
                    {tx.description}
                  </p>
                )}
              </div>

              {/* Date */}
              <div className="text-right flex-shrink-0">
                <p className="text-xs text-slate-subtle">
                  {formatDate(tx.created_at)}
                </p>
              </div>

              {/* Amount + status */}
              <div className="text-right flex-shrink-0 min-w-[100px]">
                <p
                  className={`balance-text text-sm font-semibold
                  ${tx.type === "TOPUP" ? "text-jade" : "text-white"}`}
                >
                  {tx.type === "TOPUP" ? "+" : "-"}
                  {formatIDR(tx.amount)}
                </p>
                {tx.fee > 0 && (
                  <p className="text-xs text-slate-subtle font-mono">
                    fee: {formatIDR(tx.fee)}
                  </p>
                )}
                <div className="mt-1">
                  <span
                    className={
                      tx.status === "SUCCESS"
                        ? "badge-success"
                        : tx.status === "PENDING"
                          ? "badge-pending"
                          : "badge-failed"
                    }
                  >
                    {tx.status}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

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
    </div>
  );
}
