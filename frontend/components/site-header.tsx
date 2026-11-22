"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ThemeToggle } from "@/components/theme-toggle";

export function SiteHeader() {
  const pathname = usePathname();
  return (
    <header className="site-header">
      <div className="container site-header-inner">
        <Link href="/" className="brand-mark" aria-label="BitForge home">
          <svg className="brand-symbol" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M4 5h16M4 12h11M4 19h6" stroke="currentColor" strokeWidth="3" /></svg>
          <span>bitforge<span className="brand-period">.</span></span>
        </Link>
        <div className="header-actions">
          <nav className="main-nav" aria-label="Main navigation">
            <Link href="/" aria-current={pathname === "/" ? "page" : undefined}>Today</Link>
            <Link href="/interviews" aria-current={pathname.startsWith("/interviews") ? "page" : undefined}>Interviews</Link>
            <Link href="/problems" aria-current={pathname.startsWith("/problems") ? "page" : undefined}>Problems</Link>
          <Link href="/prompts" aria-current={pathname.startsWith("/prompts") ? "page" : undefined}>Prompts</Link>
          </nav>
          <ThemeToggle />
        </div>
      </div>
    </header>
  );
}
