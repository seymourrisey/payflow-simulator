import { useState, useEffect, useRef } from "react";
import { QRCodeSVG } from "qrcode.react";
import { paymentApi, webhookApi } from "../lib/api";
import { useWalletStore } from "../store/walletStore";
import { formatIDR, getTimeLeft } from "../lib/utils";
import toast from "react-hot-toast";

const PAYMENT_FEE_RATE = 0.007;

// Icon mapping berdasarkan nama merchant
const getMerchantIcon = (name = "") => {
  if (name.toLowerCase().includes("tokopedia")) return "🛒";
  if (name.toLowerCase().includes("gojek")) return "🛵";
  if (name.toLowerCase().includes("pln")) return "⚡";
  return "🏪";
};

export default function PayPage() {
  const { wallet, fetchWallet } = useWalletStore();
  const [step, setStep] = useState("form");

  // Merchants dari API
  const [merchants, setMerchants] = useState([]);
  const [merchantsLoading, setMerchantsLoading] = useState(true);

  // Form
  const [merchantId, setMerchantId] = useState(""); // string ID
  const [amount, setAmount] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(false);

  // QR state
  const [qrData, setQrData] = useState(null);
  const [referenceId, setReferenceId] = useState("");
  const [expiredAt, setExpiredAt] = useState(null);
  const [timeLeft, setTimeLeft] = useState("");

  // Result
  const [payResult, setPayResult] = useState(null);

  // Countdown timer
  const timerRef = useRef(null);

  // ── Fetch merchants dari API saat mount ──────────────
  useEffect(() => {
    const loadMerchants = async () => {
      try {
        const res = await webhookApi.getMerchants();
        const data = res.data.data || [];
        setMerchants(data);
        // Set default merchant pertama
        if (data.length > 0) setMerchantId(data[0].id);
      } catch (err) {
        toast.error("Gagal memuat daftar merchant");
      } finally {
        setMerchantsLoading(false);
      }
    };
    loadMerchants();
    fetchWallet();
  }, []);

  // Countdown timer
  useEffect(() => {
    if (step === "qr_ready" && expiredAt) {
      timerRef.current = setInterval(() => {
        const left = getTimeLeft(expiredAt);
        if (!left) {
          clearInterval(timerRef.current);
          setStep("form");
          toast.error("QR Code sudah expired. Silakan generate ulang.");
        } else {
          setTimeLeft(left);
        }
      }, 1000);
    }
    return () => clearInterval(timerRef.current);
  }, [step, expiredAt]);

  const fee = amount ? parseFloat(amount) * PAYMENT_FEE_RATE : 0;
  const total = amount ? parseFloat(amount) + fee : 0;

  const selectedMerchant = merchants.find((m) => m.id === merchantId);

  // ── Step 1: Generate QR ──────────────────────────────
  const handleGenerateQR = async (e) => {
    e.preventDefault();
    if (!amount || parseFloat(amount) <= 0) {
      toast.error("Masukkan nominal yang valid");
      return;
    }
    if (!merchantId) {
      toast.error("Pilih merchant terlebih dahulu");
      return;
    }
    if (!wallet || wallet.balance < total) {
      toast.error("Saldo tidak mencukupi");
      return;
    }

    setLoading(true);
    try {
      const res = await paymentApi.generateQR({
        merchant_id: merchantId, // ✅ string: "MRC-TOKOPEDIA0001"
        amount: parseFloat(amount), // ✅ number: 25000
        description,
      });
      const { qr_data, reference_id, expired_at } = res.data.data;
      setQrData(qr_data);
      setReferenceId(reference_id);
      setExpiredAt(expired_at);
      setTimeLeft(getTimeLeft(expired_at));
      setStep("qr_ready");
      toast.success("QR Code berhasil dibuat!");
    } catch (err) {
      toast.error(err.response?.data?.error || "Gagal generate QR");
    } finally {
      setLoading(false);
    }
  };

  // ── Step 2: Simulasi Scan & Bayar ────────────────────
  const handleSimulatePay = async () => {
    setStep("paying");
    setLoading(true);
    try {
      const res = await paymentApi.pay(
        {
          merchant_id: merchantId, // ✅ string
          amount: parseFloat(amount), // ✅ number
          description,
        },
        referenceId,
      );
      setPayResult(res.data.data);
      setStep("success");
      fetchWallet();
      toast.success("Pembayaran berhasil!");
    } catch (err) {
      const msg = err.response?.data?.error || "Pembayaran gagal";
      toast.error(msg);
      setStep("failed");
      setPayResult({ error: msg });
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    setStep("form");
    setQrData(null);
    setAmount("");
    setDescription("");
    setPayResult(null);
    clearInterval(timerRef.current);
  };

  return (
    <div className="max-w-md mx-auto space-y-6">
      <div>
        <h1 className="font-display font-bold text-3xl">Bayar</h1>
        <p className="text-slate-subtle text-sm mt-1">
          Simulasi pembayaran via QR Code
        </p>
      </div>

      {/* ── STEP: FORM ─────────────────────────────────── */}
      {step === "form" && (
        <form
          onSubmit={handleGenerateQR}
          className="card-glow p-6 space-y-5 animate-fade-up"
        >
          {/* Saldo */}
          <div className="flex items-center justify-between p-3 rounded-xl bg-jade/5 border border-jade/10">
            <span className="text-sm text-slate-subtle">Saldo tersedia</span>
            <span className="balance-text text-jade font-semibold">
              {formatIDR(wallet?.balance || 0)}
            </span>
          </div>

          {/* Pilih merchant dari API */}
          <div>
            <label className="text-sm text-slate mb-2 block">
              Merchant Tujuan
            </label>
            {merchantsLoading ? (
              <div className="grid grid-cols-3 gap-2">
                {[1, 2, 3].map((i) => (
                  <div
                    key={i}
                    className="h-16 rounded-xl bg-ink-muted animate-pulse"
                  />
                ))}
              </div>
            ) : (
              <div className="grid grid-cols-3 gap-2">
                {merchants.map((m) => (
                  <button
                    key={m.id}
                    type="button"
                    onClick={() => setMerchantId(m.id)}
                    className={`p-3 rounded-xl border text-center transition-all duration-200
                      ${
                        merchantId === m.id
                          ? "border-jade/60 bg-jade/10 text-white"
                          : "border-ink-muted bg-ink-soft text-slate hover:border-jade/30"
                      }`}
                  >
                    <div className="text-xl mb-1">
                      {getMerchantIcon(m.merchant_name)}
                    </div>
                    <div className="text-xs leading-tight truncate">
                      {m.merchant_name.split(" ")[0]}
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Nominal */}
          <div>
            <label className="text-sm text-slate mb-2 block">
              Nominal Pembayaran
            </label>
            <div className="relative">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-subtle text-sm font-mono">
                Rp
              </span>
              <input
                type="number"
                className="input pl-10 font-mono"
                placeholder="50.000"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                min={1000}
                required
              />
            </div>
          </div>

          {/* Keterangan */}
          <div>
            <label className="text-sm text-slate mb-2 block">
              Keterangan (opsional)
            </label>
            <input
              type="text"
              className="input"
              placeholder="cth: Belanja bulanan"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          {/* Fee preview */}
          {amount > 0 && (
            <div className="space-y-2 p-3 rounded-xl bg-ink-muted/40 border border-ink-muted text-sm">
              <div className="flex justify-between text-slate-subtle">
                <span>Nominal</span>
                <span className="font-mono">
                  {formatIDR(parseFloat(amount))}
                </span>
              </div>
              <div className="flex justify-between text-slate-subtle">
                <span>Biaya admin (0.7%)</span>
                <span className="font-mono">{formatIDR(fee)}</span>
              </div>
              <div className="gradient-line" />
              <div className="flex justify-between font-semibold text-white">
                <span>Total</span>
                <span className="balance-text">{formatIDR(total)}</span>
              </div>
            </div>
          )}

          <button
            type="submit"
            disabled={loading || merchantsLoading}
            className="btn-primary w-full"
          >
            {loading ? "Membuat QR..." : "🔲 Generate QR Code"}
          </button>
        </form>
      )}

      {/* ── STEP: QR READY ─────────────────────────────── */}
      {step === "qr_ready" && (
        <div className="card-glow p-6 animate-fade-up text-center space-y-5">
          <div>
            <p className="text-sm text-slate-subtle">Bayar ke</p>
            <p className="font-display font-semibold text-lg mt-0.5">
              {getMerchantIcon(selectedMerchant?.merchant_name)}{" "}
              {selectedMerchant?.merchant_name}
            </p>
          </div>

          <div className="flex justify-center">
            <div className="p-4 rounded-2xl bg-white animate-pulse-jade">
              <QRCodeSVG
                value={qrData}
                size={180}
                bgColor="#FFFFFF"
                fgColor="#0A0A0F"
                level="M"
              />
            </div>
          </div>

          <div className="flex items-center justify-center gap-2">
            <span className="w-2 h-2 rounded-full bg-jade animate-ping" />
            <span className="font-mono text-jade font-semibold text-lg">
              {timeLeft}
            </span>
            <span className="text-slate-subtle text-sm">tersisa</span>
          </div>

          <div className="p-4 rounded-xl bg-ink-muted/40 border border-ink-muted space-y-1.5 text-sm">
            <div className="flex justify-between text-slate-subtle">
              <span>Nominal</span>
              <span className="font-mono">{formatIDR(parseFloat(amount))}</span>
            </div>
            <div className="flex justify-between text-slate-subtle">
              <span>Biaya admin</span>
              <span className="font-mono">{formatIDR(fee)}</span>
            </div>
            <div className="gradient-line" />
            <div className="flex justify-between text-white font-semibold">
              <span>Total Debit</span>
              <span className="balance-text">{formatIDR(total)}</span>
            </div>
          </div>

          <p className="text-xs text-slate-subtle font-mono break-all">
            {referenceId}
          </p>

          <button
            onClick={handleSimulatePay}
            disabled={loading}
            className="btn-pay w-full"
          >
            {loading ? "Memproses..." : "✅ Simulasi Scan & Bayar"}
          </button>
          <button
            onClick={handleReset}
            className="btn-secondary w-full text-sm"
          >
            ← Batal
          </button>
        </div>
      )}

      {/* ── STEP: PAYING ───────────────────────────────── */}
      {step === "paying" && (
        <div className="card-glow p-10 animate-fade-up text-center space-y-4">
          <div className="w-16 h-16 rounded-full border-2 border-jade border-t-transparent animate-spin mx-auto" />
          <p className="font-display font-semibold text-lg">
            Memproses Pembayaran...
          </p>
          <p className="text-slate-subtle text-sm">Jangan tutup halaman ini</p>
        </div>
      )}

      {/* ── STEP: SUCCESS ──────────────────────────────── */}
      {step === "success" && (
        <div className="card-glow p-8 animate-fade-up text-center space-y-5">
          <div className="w-16 h-16 rounded-full bg-jade/10 border-2 border-jade flex items-center justify-center mx-auto glow-jade">
            <span className="text-jade text-2xl">✓</span>
          </div>
          <div>
            <p className="font-display font-bold text-2xl text-jade">
              Berhasil!
            </p>
            <p className="text-slate-subtle text-sm mt-1">
              Pembayaran telah diproses
            </p>
          </div>
          <div className="p-4 rounded-xl bg-ink-muted/40 border border-ink-muted space-y-2 text-sm text-left">
            <div className="flex justify-between">
              <span className="text-slate-subtle">Reference ID</span>
              <span className="font-mono text-xs text-jade">
                {payResult?.reference_id}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-subtle">Nominal</span>
              <span className="font-mono">{formatIDR(payResult?.amount)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-subtle">Biaya admin</span>
              <span className="font-mono">{formatIDR(payResult?.fee)}</span>
            </div>
            <div className="gradient-line" />
            <div className="flex justify-between font-semibold text-white">
              <span>Status</span>
              <span className="badge-success">{payResult?.status}</span>
            </div>
          </div>
          <button onClick={handleReset} className="btn-primary w-full">
            ← Bayar Lagi
          </button>
        </div>
      )}

      {/* ── STEP: FAILED ───────────────────────────────── */}
      {step === "failed" && (
        <div className="card-glow p-8 animate-fade-up text-center space-y-4">
          <div className="w-16 h-16 rounded-full bg-danger/10 border-2 border-danger flex items-center justify-center mx-auto">
            <span className="text-danger text-2xl">✕</span>
          </div>
          <div>
            <p className="font-display font-bold text-2xl text-danger">Gagal</p>
            <p className="text-slate-subtle text-sm mt-1">{payResult?.error}</p>
          </div>
          <button onClick={handleReset} className="btn-secondary w-full">
            ← Coba Lagi
          </button>
        </div>
      )}
    </div>
  );
}
