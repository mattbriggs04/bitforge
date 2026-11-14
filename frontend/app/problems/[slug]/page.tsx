import Link from "next/link";
import { problemLabel } from "@/lib/problem-labels";
import { notFound } from "next/navigation";
import { ProblemWorkspace } from "@/components/problem-workspace";
import { SiteHeader } from "@/components/site-header";
import { getProblemServer } from "@/lib/server-api";

type PageProps = {
  params: Promise<{ slug: string }>;
};

function difficultyClass(difficulty: string): string {
  switch (difficulty.toLowerCase()) {
    case "easy":
      return "badge badge-easy";
    case "hard":
      return "badge badge-hard";
    default:
      return "badge badge-medium";
  }
}

export default async function ProblemDetailPage({ params }: PageProps) {
  const { slug } = await params;
  const problem = await getProblemServer(slug);

  if (!problem) {
    notFound();
  }

  return (
    <div className="site-root">
      <SiteHeader />
      <main className="problem-page problem-page-full">
        <div className="problem-page-shell">
          <div className="problem-page-head">
            <div>
              <h1>{problem.title}</h1>
            </div>
            <Link href="/problems" className="btn btn-muted">
              ← All problems
            </Link>
          </div>

          <div className="problem-meta-row">
            <span className={difficultyClass(problem.difficulty)}>{problem.difficulty}</span>
            <span className="badge badge-category">{problemLabel(problem.category)}</span>
            <span className="badge badge-type">{problemLabel(problem.problemType)}</span>
            {problem.completed ? <span className="badge badge-completed">Completed</span> : null}

          </div>
        </div>

        <ProblemWorkspace problem={problem} />
      </main>
    </div>
  );
}
