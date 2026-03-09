import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { walletApi } from "../lib/api";
import { useWalletStore } from "../store/walletStore";
import { formatIDR } from "../lib/utils";
import toast from "react-hot-toast";

const CHANNELS = [
  {
    id: "BANK_TRANSFER",
    label: "Transfer Bank",
    desc: "BCA, Mandiri, BRI, BNI",
    icon: "🏦",
  },
  {
    id: "VIRTUAL_ACCOUNT",
    label: "Virtual Account",
    desc: "Nomor VA otomatis",
    icon: "📱",
  },
];

const QUICK_AMOUNTS = [25000, 50000, 100000, 200000, 500000];

export default function TopUpPage() {
  const navigate = useNavigate();
  const { wallet, fetchWallet } = useWalletStore();
  const [channel, setChannel] = useState("BANK_TRANSFER");
  const [amount, setAmount] = useState("");
  const [loading, setLoading] = useState(false);

  const handleTopUp = async (e) => {
    e.preventDefault();
    if (!amount || parseFloat(amount) <= 0) {
      toast.error("Masukkan nominal yang valid");
      return;
    }
    setLoading(true);
    try {
      const res = await walletApi.topUp({
        amount: parseFloat(amount),
        payment_channel: channel,
      });
      toast.success(`Top Up ${formatIDR(parseFloat(amount))} berhasil!`);
      fetchWallet();
      navigate("/dashboard");
    } catch (err) {
      toast.error(err.response?.data?.error || "Top up gagal");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-md mx-auto space-y-6">
      <div>
        <h1 className="font-display font-bold text-3xl">Top Up</h1>
        <p className="text-slate-subtle text-sm mt-1">Isi saldo wallet kamu</p>
      </div>

      {/* Saldo sekarang */}
      <div className="card px-5 py-4 flex items-center justify-between">
        <div>
          <p className="text-xs text-slate-subtle">Saldo saat ini</p>
          <p className="balance-text text-jade text-xl font-semibold mt-0.5">
            {formatIDR(wallet?.balance || 0)}
          </p>
        </div>
        <div className="w-10 h-10 rounded-xl bg-jade/10 flex items-center justify-center">
          <span className="text-jade text-lg">◈</span>
        </div>
      </div>

      <form
        onSubmit={handleTopUp}
        className="card-glow p-6 space-y-5 animate-fade-up"
      >
        {/* Pilih channel */}
        <div>
          <label className="text-sm text-slate mb-3 block">
            Metode Pembayaran
          </label>
          <div className="space-y-2">
            {CHANNELS.map((ch) => (
              <button
                key={ch.id}
                type="button"
                onClick={() => setChannel(ch.id)}
                className={`w-full flex items-center gap-4 p-4 rounded-xl border transition-all duration-200 text-left
                  ${
                    channel === ch.id
                      ? "border-jade/60 bg-jade/5"
                      : "border-ink-muted bg-ink-soft hover:border-jade/30"
                  }`}
              >
                <span className="text-2xl">{ch.icon}</span>
                <div>
                  <p
                    className={`font-medium text-sm ${channel === ch.id ? "text-jade" : "text-white"}`}
                  >
                    {ch.label}
                  </p>
                  <p className="text-slate-subtle text-xs">{ch.desc}</p>
                </div>
                <div
                  className={`ml-auto w-4 h-4 rounded-full border-2 flex items-center justify-center flex-shrink-0
                  ${channel === ch.id ? "border-jade" : "border-ink-muted"}`}
                >
                  {channel === ch.id && (
                    <div className="w-2 h-2 rounded-full bg-jade" />
                  )}
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Quick amounts */}
        <div>
          <label className="text-sm text-slate mb-2 block">Nominal Cepat</label>
          <div className="flex flex-wrap gap-2">
            {QUICK_AMOUNTS.map((v) => (
              <button
                key={v}
                type="button"
                onClick={() => setAmount(String(v))}
                className={`px-3 py-1.5 rounded-lg text-xs font-mono border transition-all duration-200
                  ${
                    amount === String(v)
                      ? "bg-jade/10 border-jade/50 text-jade"
                      : "bg-ink-muted border-ink-muted text-slate hover:border-jade/30"
                  }`}
              >
                {formatIDR(v)}
              </button>
            ))}
          </div>
        </div>

        {/* Input nominal */}
        <div>
          <label className="text-sm text-slate mb-2 block">
            Atau masukkan nominal
          </label>
          <div className="relative">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-subtle text-sm font-mono">
              Rp
            </span>
            <input
              type="number"
              className="input pl-10 font-mono"
              placeholder="0"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              min={10000}
            />
          </div>
          <p className="text-xs text-slate-subtle mt-1.5">
            Minimum top up Rp 10.000
          </p>
        </div>

        {/* Simulation notice */}
        <div className="p-3 rounded-xl bg-amber-pay/5 border border-amber-pay/15">
          <p className="text-xs text-amber-pay/80">
            ⚠ Ini adalah simulasi — saldo akan langsung bertambah tanpa proses
            pembayaran nyata
          </p>
        </div>

        <button
          type="submit"
          disabled={loading || !amount}
          className="btn-primary w-full"
        >
          {loading
            ? "Memproses..."
            : `↓ Top Up ${amount ? formatIDR(parseFloat(amount)) : ""}`}
        </button>
      </form>
    </div>
  );
}
