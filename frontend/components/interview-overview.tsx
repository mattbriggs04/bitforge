"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { problemLabel } from "@/lib/problem-labels";
import { loadOrCreateUserKey, loadStoredUsername } from "@/lib/username";
import type { InterviewDetail } from "@/lib/types";

export function InterviewOverview({ initial }: { initial: InterviewDetail }) {
  const [interview, setInterview] = useState(initial);
  const [progressAvailable, setProgressAvailable] = useState(false);
  useEffect(() => {
    const controller = new AbortController();
    async function loadProgress() {
      try {
        const response = await fetch(`/api/backend/v1/interviews/${encodeURIComponent(initial.slug)}`, { signal: controller.signal, cache: "no-store", headers: { "x-user-key": loadOrCreateUserKey(), "x-user-handle": loadStoredUsername() || "demo" } });
        if (!response.ok) return;
        const value: InterviewDetail = await response.json();
        if (!controller.signal.aborted) { setInterview(value); setProgressAvailable(true); }
      } catch { /* Keep the server-rendered question list if progress is unavailable. */ }
    }
    void loadProgress(); return () => controller.abort();
  }, [initial.slug]);
  const completed = interview.problems.filter((problem) => problem.completed).length;
  const next = interview.problems.find((problem) => !problem.completed) ?? interview.problems[0];
  const questionURL = (slug: string) => `/problems/${slug}?interview=${encodeURIComponent(interview.slug)}`;
  return (
    <>
      <div className="interview-start">
        <div><span className="tag tag-category">{problemLabel(interview.category)}</span><span>{interview.questionCount} questions</span><span>~{interview.targetMinutes} min</span>{progressAvailable ? <span>{completed}/{interview.questionCount} solved</span> : null}</div>
        {next ? <Link className="btn btn-primary" href={questionURL(next.slug)}>{completed === interview.questionCount ? "Review" : completed ? "Continue" : "Start interview"}<span aria-hidden>↗</span></Link> : null}
      </div>
      <ol className="interview-questions">
        {interview.problems.map((problem, index) => (
          <li key={problem.id}>
            <Link className="interview-question" href={questionURL(problem.slug)}>
              <span className="problem-index">{progressAvailable && problem.completed ? "✓" : String(index + 1).padStart(2, "0")}</span>
              <div><h2>{problem.title}</h2><p>{problem.shortDescription}</p></div>
              <div className="interview-question-tags"><span className="tag tag-format">{problemLabel(problem.problemType)}</span><span className={`tag tag-difficulty tag-${problem.difficulty}`}>{problem.difficulty}</span></div>
              <span className="row-arrow" aria-hidden>↗</span>
            </Link>
          </li>
        ))}
      </ol>
    </>
  );
}
