"use client";

import { useSyncExternalStore } from "react";

function subscribe(callback: () => void) {
  const observer = new MutationObserver(callback);
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
  return () => observer.disconnect();
}

export function useTheme() {
  return useSyncExternalStore(subscribe, () => document.documentElement.dataset.theme === "light" ? "light" : "dark", () => "dark");
}
