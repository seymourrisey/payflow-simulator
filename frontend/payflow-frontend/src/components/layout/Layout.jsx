import Sidebar from "./Sidebar";

export default function Layout({ children }) {
  return (
    <div
      className="min-h-screen flex bg-ink"
      style={{
        backgroundImage: `
          linear-gradient(rgba(0,229,160,0.025) 1px, transparent 1px),
          linear-gradient(90deg, rgba(0,229,160,0.025) 1px, transparent 1px)
        `,
        backgroundSize: "40px 40px",
      }}
    >
      {/* Sidebar */}
      <Sidebar />

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        <div className="max-w-4xl mx-auto px-8 py-8">{children}</div>
      </main>
    </div>
  );
}
