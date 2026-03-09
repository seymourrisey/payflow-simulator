import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { authApi } from "../lib/api";
import { useAuthStore } from "../store/authStore";
import toast from "react-hot-toast";

export default function RegisterPage() {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();

  // cofirm password state
  const [form, setForm] = useState({
    full_name: "",
    email: "",
    password: "",
    confirm_password: "",
  });

  const [loading, setLoading] = useState(false);

  // error state
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();

    // Reset error
    setError("");

    // Cek panjang password
    if (form.password.length < 8) {
      toast.error("Password minimal 8 karakter");
      return;
    }

    // Cek apakah password dan confirm password sama
    if (form.password !== form.confirm_password) {
      setError("Password dan konfirmasi password tidak sama");
      toast.error("Password tidak cocok");
      return;
    }

    setLoading(true);

    try {
      // Hapus confirm_password sebelum kirim ke API
      const { confirm_password, ...registerData } = form;

      const res = await authApi.register(registerData);
      const { token, user } = res.data.data;

      setAuth(token, user);
      toast.success("Akun berhasil dibuat! Wallet otomatis terisi.");
      navigate("/dashboard");
    } catch (err) {
      toast.error(err.response?.data?.error || "Registrasi gagal");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center bg-ink px-4"
      style={{
        backgroundImage: `radial-gradient(ellipse 80% 60% at 50% -20%, rgba(0,229,160,0.08) 0%, transparent 60%),
                          linear-gradient(rgba(0,229,160,0.02) 1px, transparent 1px),
                          linear-gradient(90deg, rgba(0,229,160,0.02) 1px, transparent 1px)`,
        backgroundSize: "auto, 40px 40px, 40px 40px",
      }}
    >
      <div className="w-full max-w-sm animate-fade-up">
        {/* Logo */}
        <div className="text-center mb-10">
          <div className="w-14 h-14 rounded-2xl bg-jade/10 border border-jade/30 flex items-center justify-center mx-auto mb-4 glow-jade">
            <span className="font-display font-black text-jade text-2xl">
              P
            </span>
          </div>
          <h1 className="font-display font-bold text-2xl">
            Pay<span className="text-gradient">Flow</span>
          </h1>
          <p className="text-slate-subtle text-sm mt-1">
            Payment Gateway Simulator
          </p>
        </div>

        <div className="card-glow p-8">
          <h2 className="font-display font-semibold text-xl mb-6">Buat Akun</h2>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="text-sm text-slate mb-2 block">
                Nama Lengkap
              </label>
              <input
                type="text"
                className="input"
                placeholder="Kurt Cobain"
                value={form.full_name}
                onChange={(e) =>
                  setForm({ ...form, full_name: e.target.value })
                }
                required
              />
            </div>

            <div>
              <label className="text-sm text-slate mb-2 block">Email</label>
              <input
                type="email"
                className="input"
                placeholder="nama@email.com"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                required
              />
            </div>

            <div>
              <label className="text-sm text-slate mb-2 block">Password</label>
              <input
                type="password"
                className="input"
                placeholder="Minimal 8 karakter"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                required
                minLength={8}
              />
            </div>

            <div>
              <label className="text-sm text-slate mb-2 block">
                Konfirmasi Password
              </label>
              <input
                type="password"
                className="input"
                placeholder="Masukkan password lagi"
                value={form.confirm_password}
                onChange={(e) =>
                  setForm({ ...form, confirm_password: e.target.value })
                }
                required
                minLength={8}
              />
              {error && <p className="text-red-400 text-xs mt-1.5">{error}</p>}
            </div>

            <div className="mt-2 p-3 rounded-xl bg-jade/5 border border-jade/10">
              <p className="text-xs text-slate-subtle">
                ✦ Wallet IDR akan otomatis dibuat saat daftar
              </p>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full mt-2"
            >
              {loading ? "Memproses..." : "Buat Akun →"}
            </button>
          </form>
        </div>

        <p className="text-center text-slate-subtle text-sm mt-6">
          Sudah punya akun?{" "}
          <Link to="/login" className="text-jade hover:underline">
            Masuk
          </Link>
        </p>
      </div>
    </div>
  );
}
