import Link from "next/link";

export function SiteHeader() {
  return (
    <header className="site-header">
      <div className="container site-header-inner">
        <Link href="/" className="brand-mark">
          <span className="brand-dot" aria-hidden>
            BF
          </span>
          <span className="brand-text">
            <strong>BitForge</strong>
            <small>systems interview lab</small>
          </span>
        </Link>
        <nav className="main-nav">
          <Link href="/problems">Problem Catalog</Link>
          <Link href="/competitions">Competitions</Link>
          <a href="#tracks">Tracks</a>
        </nav>
      </div>
    </header>
  );
}
