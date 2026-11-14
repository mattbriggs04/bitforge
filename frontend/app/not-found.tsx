import Link from "next/link";
import { SiteHeader } from "@/components/site-header";

export default function NotFound() {
  return (
    <div className="site-root">
      <SiteHeader />
      <main className="container" style={{ paddingTop: "5rem", paddingBottom: "5rem" }}>
        <h1>Not found.</h1>
        <p>This page doesn’t exist.</p>
        <Link href="/problems" className="btn btn-primary">
          All problems
        </Link>
      </main>
    </div>
  );
}
