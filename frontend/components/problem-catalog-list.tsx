"use client";

import Link from "next/link";
import { problemLabel } from "@/lib/problem-labels";
import { useEffect, useState } from "react";
import { ProblemSummary } from "@/lib/types";
import { loadOrCreateUserKey, loadStoredUsername } from "@/lib/username";

type Props = {
  initialProblems: ProblemSummary[];
  q: string;
  difficulty: string;
  category: string;
  tag: string;
};

function buildQuery(params: Omit<Props, "initialProblems">): string {
  const query = new URLSearchParams();
  if (params.q) query.set("q", params.q);
  if (params.difficulty) query.set("difficulty", params.difficulty);
  if (params.category) query.set("category", params.category);
  if (params.tag) query.set("tag", params.tag);
  return query.toString();
}

export function ProblemCatalogList({ initialProblems, q, difficulty, category, tag }: Props) {
  const [problems, setProblems] = useState(initialProblems);

  useEffect(() => {
    setProblems(initialProblems);
  }, [initialProblems]);

  useEffect(() => {
    const userKey = loadOrCreateUserKey();
    const userHandle = loadStoredUsername() || "demo";
    const query = buildQuery({ q, difficulty, category, tag });
    const url = query ? `/api/backend/v1/problems?${query}` : "/api/backend/v1/problems";
    let active = true;

    const refreshCompletion = async (): Promise<void> => {
      try {
        const response = await fetch(url, {
          cache: "no-store",
          headers: {
            "x-user-handle": userHandle,
            "x-user-key": userKey,
          },
        });
        const payload = (await response.json()) as { items?: ProblemSummary[] };
        if (active && response.ok && Array.isArray(payload.items)) {
          setProblems(payload.items);
        }
      } catch {
        // Keep the server-rendered list if completion refresh fails.
      }
    };

    void refreshCompletion();
    return () => { active = false; };
  }, [initialProblems, q, difficulty, category, tag]);

  return (
    <section className="problem-list" aria-label="Problems">
      <div className="problem-list-heading"><span>Problem</span><span>{problems.length} total</span></div>
      {problems.length === 0 ? <div className="empty-state">No matches.</div> : problems.map((problem, index) => (
        <Link key={problem.id} href={`/problems/${problem.slug}`} className={`problem-row ${problem.completed ? "is-complete" : ""}`}>
          <span className="problem-index" aria-label={problem.completed ? "Completed" : undefined}>{problem.completed ? "✓" : String(index + 1).padStart(2, "0")}</span>
          <div className="problem-row-body"><h2>{problem.title}</h2><p>{problem.shortDescription}</p></div>
          <div className="problem-row-meta"><span className="problem-row-category tag tag-category">{problemLabel(problem.category)}</span>
          <span className={`problem-difficulty tag tag-difficulty tag-${problem.difficulty}`}><span className={`difficulty-meter level-${problem.difficulty}`} aria-hidden><i /><i /><i /></span>{problem.difficulty}</span></div>
          <span className="row-arrow" aria-hidden>↗</span>
        </Link>
      ))}
    </section>
  );
}
