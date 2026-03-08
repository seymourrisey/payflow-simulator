/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,jsx}"],
  theme: {
    extend: {
      fontFamily: {
        // Display font — tegas, financial feel
        display: ["Syne", "sans-serif"],
        // Body font — clean & readable
        body: ["DM Sans", "sans-serif"],
        // Monospace untuk angka saldo
        mono: ["JetBrains Mono", "monospace"],
      },
      colors: {
        ink: {
          DEFAULT: "#0A0A0F",
          soft: "#1A1A24",
          muted: "#2E2E3E",
        },
        slate: {
          subtle: "#8888A0",
          DEFAULT: "#C0C0D0",
        },
        jade: {
          DEFAULT: "#00E5A0",
          dim: "#00B87E",
          glow: "rgba(0, 229, 160, 0.15)",
        },
        amber: {
          pay: "#F5A623",
          glow: "rgba(245, 166, 35, 0.15)",
        },
        danger: "#FF4D6D",
      },
      backgroundImage: {
        "grid-pattern": `linear-gradient(rgba(0,229,160,0.03) 1px, transparent 1px),
                         linear-gradient(90deg, rgba(0,229,160,0.03) 1px, transparent 1px)`,
      },
      backgroundSize: {
        grid: "40px 40px",
      },
      animation: {
        "fade-up": "fadeUp 0.5s ease forwards",
        "pulse-jade": "pulseJade 2s ease-in-out infinite",
        "spin-slow": "spin 3s linear infinite",
        countdown: "countdown linear forwards",
      },
      keyframes: {
        fadeUp: {
          from: { opacity: 0, transform: "translateY(16px)" },
          to: { opacity: 1, transform: "translateY(0)" },
        },
        pulseJade: {
          "0%, 100%": { boxShadow: "0 0 0 0 rgba(0, 229, 160, 0.4)" },
          "50%": { boxShadow: "0 0 0 12px rgba(0, 229, 160, 0)" },
        },
      },
    },
  },
  plugins: [],
};
