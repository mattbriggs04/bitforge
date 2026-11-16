import Link from "next/link";
import type { InterviewDetail } from "@/lib/types";

export function InterviewNavigation({ interview, problemSlug }: { interview: InterviewDetail; problemSlug: string }) {
  const index = interview.problems.findIndex((problem) => problem.slug === problemSlug);
  if (index < 0) return null;
  const previous = interview.problems[index - 1];
  const next = interview.problems[index + 1];
  const questionURL = (slug: string) => `/problems/${slug}?interview=${encodeURIComponent(interview.slug)}`;
  return (
    <nav className="interview-navigation" aria-label="Interview questions">
      <div><Link className="interview-context" href={`/interviews/${interview.slug}`}>{interview.title}</Link><span className="interview-position">Question {index + 1} of {interview.problems.length}</span></div>
      <div className="interview-steps">{previous ? <Link className="btn btn-muted" href={questionURL(previous.slug)}>← Previous</Link> : null}{next ? <Link className="btn btn-muted" href={questionURL(next.slug)}>Next →</Link> : <Link className="btn btn-muted" href={`/interviews/${interview.slug}`}>Back to overview →</Link>}</div>
    </nav>
  );
}
