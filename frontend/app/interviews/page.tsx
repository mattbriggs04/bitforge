import Link from "next/link";
import { SiteHeader } from "@/components/site-header";
import { getInterviewsServer } from "@/lib/server-api";
import { problemLabel } from "@/lib/problem-labels";
import type { InterviewSummary } from "@/lib/types";

export default async function InterviewsPage() {
  let interviews: InterviewSummary[] = [];
  let unavailable = false;
  try { interviews = await getInterviewsServer(); } catch { unavailable = true; }
  return (
    <div className="site-root">
      <SiteHeader />
      <main className="container catalog-page">
        <section className="catalog-header"><h1>Interviews<span className="accent-period">.</span></h1></section>
        {unavailable ? <p className="catalog-warning">Backend unavailable. Try again shortly.</p> : null}
        <div className="interview-list">
          {!unavailable && interviews.length === 0 ? <p className="empty-state">No interviews yet.</p> : null}
          {interviews.map((interview) => (
            <Link className="interview-card" key={interview.slug} href={`/interviews/${interview.slug}`}>
              <div><span className="tag tag-category">{problemLabel(interview.category)}</span><h2>{interview.title}</h2><p>{interview.description}</p><div className="interview-card-meta"><span>{interview.questionCount} questions</span><span>~{interview.targetMinutes} min</span></div></div>
              <span className="row-arrow" aria-hidden>↗</span>
            </Link>
          ))}
        </div>
      </main>
    </div>
  );
}
