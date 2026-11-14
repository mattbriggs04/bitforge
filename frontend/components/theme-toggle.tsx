"use client";

import { useTheme } from "@/lib/theme";

export function ThemeToggle() {
  const theme = useTheme();
  return (
    <button className="theme-toggle" type="button" aria-label={theme === "dark" ? "Switch to light theme" : "Switch to dark theme"} onClick={() => {
      const next = theme === "dark" ? "light" : "dark";
      document.documentElement.dataset.theme = next;
      try { localStorage.setItem("bf-theme", next); } catch { /* Theme still works without storage. */ }
    }}>
      <svg className="theme-sun" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" aria-hidden="true"><circle cx="12" cy="12" r="3.5" /><path d="M12 2v3m0 14v3M2 12h3m14 0h3M5 5l2 2m10 10 2 2M5 19l2-2M17 7l2-2" /></svg>
      <svg className="theme-moon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" aria-hidden="true"><path d="M20 15.1A8.5 8.5 0 0 1 8.9 4 8.5 8.5 0 1 0 20 15.1Z" /></svg>
    </button>
  );
}
