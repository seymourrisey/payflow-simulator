import { NavLink } from "react-router-dom";

const navItems = [
  { to: "/dashboard", icon: "⬡", label: "Dashboard" },
  { to: "/pay", icon: "⬢", label: "Bayar" },
  { to: "/topup", icon: "↓", label: "Top Up" },
  { to: "/history", icon: "≡", label: "Riwayat" },
];

export default function SidebarNav() {
  return (
    <nav className="flex-1 px-3 space-y-1">
      {navItems.map(({ to, icon, label }) => (
        <NavLink
          key={to}
          to={to}
          className={({ isActive }) =>
            `flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200
             ${
               isActive
                 ? "bg-jade/10 text-jade border border-jade/20"
                 : "text-slate hover:text-white hover:bg-ink-muted"
             }`
          }
        >
          <span className="text-base w-5 text-center">{icon}</span>
          <span className="font-body font-medium text-sm">{label}</span>
        </NavLink>
      ))}
    </nav>
  );
}
