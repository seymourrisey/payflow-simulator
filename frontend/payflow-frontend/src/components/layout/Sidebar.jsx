import SidebarLogo from "./SidebarLogo";
import SidebarWallet from "./SidebarWallet";
import SidebarNav from "./SidebarNav";
import SidebarUser from "./SidebarUser";

export default function Sidebar() {
  return (
    <aside className="w-64 flex-shrink-0 flex flex-col border-r border-ink-muted/60 bg-ink/80 backdrop-blur-xl">
      {/* Logo */}
      <SidebarLogo />

      {/* Divider */}
      <div className="gradient-line mx-6 mb-6" />

      {/* Mini wallet preview */}
      <SidebarWallet />

      {/* Navigation links */}
      <SidebarNav />

      {/* User info + logout */}
      <SidebarUser />
    </aside>
  );
}
