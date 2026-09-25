import type { Config } from "tailwindcss";

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
        background: "#000000",
        surface: "#121212",
        surfaceSubtle: "#181818",
        card: "#181818",
        border: "#272727",
        borderHighlight: "#383838",
        primary: {
          DEFAULT: "#FFFFFF",
          hover: "#E4E4E7",
          foreground: "#09090B",
        },
        accent: {
          DEFAULT: "#10B981", // Verification emerald
          hover: "#059669",
          foreground: "#FFFFFF",
        },
        muted: {
          DEFAULT: "#A1A1AA",
          foreground: "#71717A",
        },
      },
      fontFamily: {
        sans: ["var(--font-sans)", "system-ui", "sans-serif"],
        mono: ["var(--font-mono)", "ui-monospace", "SFMono-Regular", "monospace"],
      },
    },
  },
  plugins: [],
};

export default config;
