import { useNavigate } from "react-router-dom";
import { useAuthStore } from "../../store/authStore";

export default function SidebarUser() {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const initial = user?.full_name?.[0]?.toUpperCase() || "U";

  return (
    <div className="p-4 border-t border-ink-muted/60">
      {/* Avatar + info */}
      <div className="flex items-center gap-3 px-2 py-2">
        <div className="w-8 h-8 rounded-full bg-jade/20 border border-jade/30 flex items-center justify-center flex-shrink-0">
          <span className="text-jade text-sm font-bold">{initial}</span>
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-white truncate">
            {user?.full_name || "User"}
          </p>
          <p className="text-xs text-slate-subtle truncate">{user?.email}</p>
        </div>
      </div>

      {/* Logout button */}
      <button
        onClick={handleLogout}
        className="w-full mt-2 text-sm text-slate-subtle hover:text-danger
                   transition-colors py-2 px-4 rounded-lg hover:bg-danger/10 text-left"
      >
        → Keluar
      </button>
    </div>
  );
}
