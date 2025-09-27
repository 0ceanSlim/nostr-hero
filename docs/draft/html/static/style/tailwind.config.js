/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.{html,js}"],
  theme: {
    extend: {
      colors: {
        bgPrimary: "var(--color-bgPrimary)",
        bgSecondary: "var(--color-bgSecondary)",
        bgTertiary: "var(--color-bgTertiary)",
        bgInverted: "var(--color-bgInverted)",
        textPrimary: "var(--color-textPrimary)",
        textSecondary: "var(--color-textSecondary)",
        textMuted: "var(--color-textMuted)",
        textInverted: "var(--color-textInverted)",
        textHighlighted: "var(--color-textHighlighted)",
      },
      keyframes: {
        spin: {
          "0%": { transform: "rotate(0deg)" },
          "100%": { transform: "rotate(360deg)" },
        },
        fadeInUp: {
          "0%": { opacity: 0, transform: "translateY(100%)" },
          "100%": { opacity: 1, transform: "translateY(0)" },
        },
      },
      animation: {
        spin: "spin 1s linear infinite",
        fadeInUp: "fadeInUp 0.5s ease-out",
      },
      borderWidth: {
        5: "5px",
      },
    },
  },
  plugins: [require("@tailwindcss/aspect-ratio")],
};
