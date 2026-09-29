import type { Config } from "tailwindcss";

/**
 * Verinode design system.
 *
 * Identity: an institutional commodities desk, not a crypto dashboard. Warm ink
 * base (not pure black), a signal-gold primary that reads as "physical commodity /
 * forward market", and a verification-green used ONLY for pass/attested states.
 * Serif display for gravitas, clean sans for body, mono strictly for data/hashes.
 */
const config: Config = {
  darkMode: ["class"],
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        // Warm ink base ramp (very slightly blue-warm, never pure #000).
        ink: {
          950: "#0A0B0E", // page base
          900: "#101217", // raised surface
          850: "#151821", // card
          800: "#1B1F29", // card hover / inputs
          700: "#252A36", // hairline-strong
          600: "#333A48",
        },
        line: "#20242E", // default hairline
        lineSoft: "#191C24",
        // Signal gold — the commodity/forward accent.
        signal: {
          DEFAULT: "#E5A94E",
          bright: "#F2C078",
          deep: "#B4801F",
          wash: "#2A2113",
        },
        // Verification green — pass/attested ONLY.
        verify: {
          DEFAULT: "#4ADE9B",
          deep: "#0F9D6B",
          wash: "#0F1E19",
        },
        // Alert / failure.
        alert: {
          DEFAULT: "#F26D6D",
          wash: "#241416",
        },
        parchment: "#EDE7DA", // warm near-white text
        muted: {
          DEFAULT: "#9AA0AD",
          soft: "#6B7180",
          foreground: "#6B7180",
        },

        // ---------------------------------------------------------------
        // Legacy token aliases — mapped onto the new palette so existing
        // portal pages keep rendering with correct contrast. Prefer the
        // ink-*/signal/verify/parchment tokens above in new code.
        // ---------------------------------------------------------------
        background: "#0A0B0E",
        foreground: "#EDE7DA",
        surface: "#151821",
        surfaceSubtle: "#1B1F29",
        card: "#151821",
        border: "#20242E",
        borderHighlight: "#252A36",
        primary: {
          DEFAULT: "#E5A94E",
          hover: "#F2C078",
          foreground: "#0A0B0E",
        },
        accent: {
          DEFAULT: "#4ADE9B",
          hover: "#0F9D6B",
          foreground: "#0A0B0E",
        },
      },
      fontFamily: {
        serif: ["var(--font-serif)", "Georgia", "serif"],
        sans: ["var(--font-sans)", "system-ui", "sans-serif"],
        mono: ["var(--font-mono)", "ui-monospace", "SFMono-Regular", "monospace"],
      },
      fontSize: {
        // Tighter display scale with balanced line-heights.
        display: ["clamp(2.5rem, 5vw, 4.25rem)", { lineHeight: "1.02", letterSpacing: "-0.02em" }],
        headline: ["clamp(1.75rem, 3vw, 2.75rem)", { lineHeight: "1.08", letterSpacing: "-0.015em" }],
        title: ["1.375rem", { lineHeight: "1.2", letterSpacing: "-0.01em" }],
      },
      borderRadius: {
        card: "14px",
        pill: "999px",
      },
      boxShadow: {
        raise: "0 1px 0 0 rgba(255,255,255,0.03) inset, 0 12px 40px -12px rgba(0,0,0,0.6)",
        glow: "0 0 0 1px rgba(229,169,78,0.25), 0 8px 30px -8px rgba(229,169,78,0.25)",
      },
      backgroundImage: {
        "grid-faint":
          "linear-gradient(to right, rgba(255,255,255,0.025) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.025) 1px, transparent 1px)",
      },
      keyframes: {
        rise: {
          "0%": { opacity: "0", transform: "translateY(10px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        pulseDot: {
          "0%, 100%": { opacity: "1" },
          "50%": { opacity: "0.35" },
        },
      },
      animation: {
        rise: "rise 0.5s cubic-bezier(0.22, 1, 0.36, 1) both",
        pulseDot: "pulseDot 2s ease-in-out infinite",
      },
    },
  },
  plugins: [],
};

export default config;
