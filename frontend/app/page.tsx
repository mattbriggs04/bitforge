import Link from "next/link";
import { DailyPractice } from "@/components/daily-practice";
import { SiteHeader } from "@/components/site-header";

const fields = [
  ["embedded", "Embedded systems"],
  ["ai_ml", "AI & machine learning"],
  ["computer_vision", "Computer vision"],
  ["computer_architecture", "Computer architecture"],
  ["operating_systems", "Operating systems"],
];

export default function Home() {
  return (
    <div className="site-root">
      <SiteHeader />
      <main className="container home-page">
        <section className="home-intro">
          <div className="home-heading">
            <h1 className="typewriter" aria-label="Stay sharp.">
              <span className="type-line type-line-first" aria-hidden="true">Stay</span>
              <span className="type-line type-line-second" aria-hidden="true">sharp<span className="accent-period">.</span></span>
            </h1>
            <p className="hero-copy">Daily problems for computer enthusiasts.</p>
            <Link className="text-link" href="/problems">Browse problems <span aria-hidden>↗</span></Link>
          </div>
          <DailyPractice />
        </section>
        <section className="fields-section" id="fields" aria-labelledby="fields-heading">
          <div className="section-heading"><h2 id="fields-heading">Pick a field.</h2><span className="eyebrow">01—05</span></div>
          <div className="field-list">
            {fields.map(([slug, title], index) => (
              <Link href={`/problems?category=${slug}`} className="field-link" key={slug}>
                <span className="field-index">0{index + 1}</span><span>{title}</span><span className="field-arrow" aria-hidden>↗</span>
              </Link>
            ))}
          </div>
        </section>
      </main>
      <footer className="container site-footer"><span>BitForge</span><span>C / C++ / Python / Rust</span></footer>
    </div>
  );
}
