import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { authApi } from "../lib/api";
import { useAuthStore } from "../store/authStore";
import toast from "react-hot-toast";

export default function LoginPage() {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();
  const [form, setForm] = useState({ email: "", password: "" });
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await authApi.login(form);
      const { token, user } = res.data.data;
      setAuth(token, user);
      toast.success(`Selamat datang, ${user.full_name}!`);
      navigate("/dashboard");
    } catch (err) {
      toast.error(err.response?.data?.error || "Login gagal");
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

        {/* Card */}
        <div className="card-glow p-8">
          <h2 className="font-display font-semibold text-xl mb-6">Masuk</h2>

          <form onSubmit={handleSubmit} className="space-y-4">
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
                placeholder="••••••••"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                required
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full mt-2"
            >
              {loading ? "Memproses..." : "Masuk →"}
            </button>
          </form>
        </div>

        <p className="text-center text-slate-subtle text-sm mt-6">
          Belum punya akun?{" "}
          <Link to="/register" className="text-jade hover:underline">
            Daftar sekarang
          </Link>
        </p>
      </div>
    </div>
  );
}
