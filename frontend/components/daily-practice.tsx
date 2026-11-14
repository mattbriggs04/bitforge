"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { problemLabel } from "@/lib/problem-labels";
import { ProblemSummary } from "@/lib/types";
import { loadOrCreateUserKey, loadStoredUsername } from "@/lib/username";

export function DailyPractice() {
  const [problem, setProblem] = useState<ProblemSummary | null>(null);
  const [message, setMessage] = useState("Loading…");
  useEffect(() => {
    const controller = new AbortController();
    async function load() {
      try {
        const key = loadOrCreateUserKey();
        const response = await fetch("/api/backend/v1/problems", { signal: controller.signal, headers: { "x-user-key": key, "x-user-handle": loadStoredUsername() || "demo" } });
        if (!response.ok) throw new Error();
        const { items }: { items: ProblemSummary[] } = await response.json();
        const now = new Date();
        const day = `${now.getFullYear()}-${now.getMonth()+1}-${now.getDate()}`;
        // Store a daily choice so completion and refresh do not change it mid-day.
        const storageKey = `bitforge:daily:${day}:${key}`;
        let stored = "";
        try { stored = localStorage.getItem(storageKey) ?? ""; } catch { /* optional */ }
        let choice = items.find((item) => item.slug === stored);
        if (!choice) {
          const pending = items.filter((item) => !item.completed);
          const pool = pending.length ? pending : items;
          const categories = [...new Set(pool.map((item) => item.category))].sort();
          let hash = 0; for (const c of day + key) hash = (Math.imul(hash, 31) + c.charCodeAt(0)) >>> 0;
          const category = categories[hash % categories.length];
          const track = pool.filter((item) => item.category === category);
          choice = track[hash % track.length];
          if (choice) { try { localStorage.setItem(storageKey, choice.slug); } catch { /* optional */ } }
        }
        if (!controller.signal.aborted) { setProblem(choice ?? null); setMessage(choice ? "" : "No questions yet."); }
      } catch { if (!controller.signal.aborted) setMessage("Connect the backend to load today’s problem."); }
    }
    void load(); return () => controller.abort();
  }, []);
  return (
    <section className="daily-practice" aria-labelledby="daily-title">
      <div className="daily-top"><span className="eyebrow"><span className="status-light" />Today’s problem</span></div>
      <div className="daily-content">
        <p className="daily-category">{problem ? <span className="tag tag-category">{problemLabel(problem.category)}</span> : "Daily practice"}</p>
        <h2 id="daily-title">{problem?.title ?? (message === "Loading…" ? "One moment." : "Your next problem.")}</h2>
        <p className="daily-description">{problem?.shortDescription ?? message}</p>
        {problem ? <div className="daily-meta"><span className="tag tag-format">{problemLabel(problem.problemType)}</span><span className={`tag tag-difficulty tag-${problem.difficulty}`}>{problem.difficulty}</span>{problem.completed ? <span className="tag tag-category">Complete ✓</span> : null}</div> : null}
      </div>
      <Link className="btn btn-primary daily-action" href={problem ? `/problems/${problem.slug}` : "/problems"}>{problem ? problem.completed ? "Review" : "Solve" : "Browse"}<span aria-hidden>↗</span></Link>
    </section>
  );
}
