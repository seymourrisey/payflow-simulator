export default function SidebarLogo() {
  return (
    <div className="px-6 pt-8 pb-6">
      <div className="flex items-center gap-2.5">
        <div className="w-8 h-8 rounded-lg bg-jade flex items-center justify-center">
          <span className="text-ink font-display font-bold text-sm">P</span>
        </div>
        <span className="font-display font-bold text-lg tracking-tight">
          Pay<span className="text-gradient">Flow</span>
        </span>
      </div>
      <p className="text-slate-subtle text-xs mt-1.5 ml-10">Simulator v1.0</p>
    </div>
  );
}
