"use client";

import { useEffect, useState } from "react";
import ReactMarkdown from "react-markdown";
import { problemLabel } from "@/lib/problem-labels";
import { ProblemDetail, Submission } from "@/lib/types";
import { loadOrCreateUserKey, loadStoredUsername } from "@/lib/username";

type Option = { id: string; text: string };

export function AnswerWorkbench({ problem }: { problem: ProblemDetail }) {
  const format = String(problem.metadata.format ?? problem.problemType);
  const options = (problem.metadata.options ?? []) as Option[];
  const [selected, setSelected] = useState<string[]>([]);
  const [numeric, setNumeric] = useState("");
  const [submission, setSubmission] = useState<Submission | null>(null);
  const [submissionId, setSubmissionId] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [explanation, setExplanation] = useState("");
  const [retry, setRetry] = useState(0);
  const draftKey = `bitforge:answer:${problem.slug}:${problem.metadata.bundleHash ?? "v1"}`;

  useEffect(() => {
    try {
      const draft = JSON.parse(localStorage.getItem(draftKey) ?? "null");
      if (draft) { setSelected(Array.isArray(draft.selected) ? draft.selected : []); setNumeric(typeof draft.numeric === "string" ? draft.numeric : ""); }
    } catch { /* Practice remains usable when browser storage is unavailable. */ }
  }, [draftKey]);

  function save(nextSelected: string[], nextNumeric: string) {
    setSelected(nextSelected); setNumeric(nextNumeric);
    try { localStorage.setItem(draftKey, JSON.stringify({ selected: nextSelected, numeric: nextNumeric })); } catch { /* Storage can be disabled. */ }
  }
  useEffect(() => {
    if (!submissionId) return;
    let active = true;
    let timer: ReturnType<typeof setTimeout>;
    let failures = 0;
    async function poll() {
      try {
        const response = await fetch(`/api/backend/v1/submissions/${submissionId}`, { cache: "no-store" });
        if (!response.ok) throw new Error("Could not retrieve grading results.");
        const result: Submission = await response.json();
        if (!active) return;
        setSubmission(result); failures = 0; setError("");
        if (result.status === "queued" || result.status === "running") { timer = setTimeout(poll, 1000); }
        else setBusy(false);
      } catch (e) {
        if (!active) return;
        if (++failures < 4) timer = setTimeout(poll, failures * 1500);
        else { setError(e instanceof Error ? e.message : "Grading connection interrupted."); setBusy(false); }
      }
    }
    void poll();
    return () => { active = false; clearTimeout(timer); };
  }, [submissionId, retry]);

  async function submit() {
    setBusy(true); setError(""); setExplanation(""); setSubmission(null); setSubmissionId("");
    try {
      const sourceCode = format === "numeric" ? numeric.trim() : JSON.stringify(format === "single_choice" ? selected[0] : selected);
      const response = await fetch("/api/backend/v1/submissions", {
        method: "POST", headers: { "content-type": "application/json", "x-user-key": loadOrCreateUserKey(), "x-user-handle": loadStoredUsername() || "demo" },
        body: JSON.stringify({ problemSlug: problem.slug, language: "answer", mode: "submit", sourceCode }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Submission failed.");
      setSubmissionId(data.submissionId);
    } catch (e) { setError(e instanceof Error ? e.message : "Submission failed."); setBusy(false); }
  }
  async function reveal() {
    try {
      const response = await fetch(`/api/backend/v1/problems/${problem.slug}/explanation?submissionId=${submissionId}`, { headers: { "x-user-key": loadOrCreateUserKey(), "x-user-handle": loadStoredUsername() || "demo" } });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Could not load explanation.");
      setExplanation(data.explanation);
    } catch (e) { setError(e instanceof Error ? e.message : "Could not load explanation."); }
  }
  return (
    <section className="answer-workbench">
      <p className="eyebrow">{problemLabel(format)}</p>
      <h2>Your answer</h2>

      {format === "numeric" ? (
        <label className="numeric-answer">Answer<input type="text" inputMode="decimal" value={numeric} onChange={(e) => save(selected, e.target.value)} placeholder="0.00" aria-label="Numeric answer" /></label>
      ) : (
        <fieldset className="answer-options" disabled={busy}>
          <legend>{format === "multiple_select" ? "Select all that apply" : "Choose one answer"}</legend>
          {options.map((option) => (
            <label key={option.id} className={`answer-option ${selected.includes(option.id) ? "is-selected" : ""}`}>
              <input type={format === "multiple_select" ? "checkbox" : "radio"} name="answer" checked={selected.includes(option.id)} onChange={() => save(format === "single_choice" ? [option.id] : selected.includes(option.id) ? selected.filter((id) => id !== option.id) : [...selected, option.id], numeric)} />
              <span><strong>{option.id}</strong> {option.text}</span>
            </label>
          ))}
        </fieldset>
      )}
      <button className="btn btn-primary" disabled={busy || (format === "numeric" ? !numeric.trim() : selected.length === 0)} onClick={() => void submit()}>{busy ? "Grading…" : "Submit"}</button>
      {error ? <p role="alert">{error} {submissionId ? <button className="btn btn-muted" onClick={() => { setError(""); setBusy(true); setRetry((n) => n + 1); }}>Retry results</button> : null}</p> : null}
      <div aria-live="polite" className="answer-feedback">
        {submission ? <p>{submission.status === "queued" || submission.status === "running" ? "Grading…" : submission.verdict === "accepted" ? "Correct." : submission.verdict === "wrong_answer" ? "Incorrect. Try again or view the explanation." : submission.errorMessage || "Grading could not finish."}</p> : null}
        {submission?.status === "completed" ? <button className="btn btn-muted" onClick={() => void reveal()}>Explanation</button> : null}
      </div>
      {explanation ? <article className="markdown-body answer-explanation"><ReactMarkdown>{explanation}</ReactMarkdown></article> : null}
    </section>
  );
}
