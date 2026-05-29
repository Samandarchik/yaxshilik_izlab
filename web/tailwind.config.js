/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        // Brend ranglari — to'q sariq asosiy urg'u
        accent: {
          DEFAULT: "#FF6B35",
          soft: "#FFF1EA",
          ring: "#FFD9C7",
          dark: "#E55A26",
        },
        ink: {
          DEFAULT: "#0F172A", // asosiy matn (slate-900)
          2: "#475569", // ikkilamchi (slate-600)
          3: "#94A3B8", // uchlamchi (slate-400)
        },
        line: "#E9EDF3",
        page: "#F7F8FB",
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
        serif: ["'Instrument Serif'", "Georgia", "serif"],
      },
      boxShadow: {
        soft: "0 1px 2px rgba(15,23,42,0.04), 0 8px 24px rgba(15,23,42,0.06)",
        card: "0 1px 3px rgba(15,23,42,0.06), 0 12px 32px rgba(15,23,42,0.08)",
        glow: "0 10px 30px rgba(255,107,53,0.30)",
      },
      borderRadius: {
        "2xl": "1.25rem",
        "3xl": "1.75rem",
      },
      keyframes: {
        "fade-up": {
          "0%": { opacity: "0", transform: "translateY(12px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        pop: {
          "0%": { transform: "scale(1)" },
          "50%": { transform: "scale(1.25)" },
          "100%": { transform: "scale(1)" },
        },
        pulse2: {
          "0%,100%": { opacity: "1" },
          "50%": { opacity: "0.4" },
        },
      },
      animation: {
        "fade-up": "fade-up 0.5s ease both",
        pop: "pop 0.4s ease",
        pulse2: "pulse2 1.6s ease-in-out infinite",
      },
    },
  },
  plugins: [],
};
