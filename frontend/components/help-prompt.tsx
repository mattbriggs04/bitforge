"use client";

import { useState } from "react";
import { copyPrompt, tutorPrompt, type HelpMode } from "@/lib/tutor-prompt";
import type { ProblemDetail, Submission } from "@/lib/types";

export function HelpPrompt({ problem, work, language, submission, expanded = false }: { problem: ProblemDetail; work: string; language: string; submission?: Submission | null; expanded?: boolean }) {
  const [mode, setMode] = useState<HelpMode>("scaffold");
  const [stuck, setStuck] = useState("");
  const [status, setStatus] = useState("");
  const [preview, setPreview] = useState(false);
  const prompt = tutorPrompt(problem, mode, stuck, work, language, submission);
  async function copy() {
    try { await copyPrompt(prompt); setStatus("Copied. Paste into your agent."); }
    catch { setPreview(true); setStatus("Select and copy the prompt below."); }
  }
  return (
    <details className="help-prompt" open={expanded || undefined}>
      <summary>Ask an agent</summary>
      <div className="help-prompt-body">
        <p className="form-note">Copies this problem and your current work into a tutor prompt.</p>
        <label>Help level<select value={mode} onChange={(event) => { setMode(event.target.value as HelpMode); setStatus(""); }}>
          <option value="nudge">Nudge</option><option value="approach">Approach</option><option value="scaffold">Scaffold</option><option value="walkthrough">Full walkthrough</option>
        </select></label>
        <label>Where are you stuck?<textarea rows={3} value={stuck} onChange={(event) => { setStuck(event.target.value); setStatus(""); }} placeholder="What have you tried? What is unclear?" /></label>
        <div className="prompt-actions"><button type="button" className="btn btn-primary" onClick={() => void copy()}>Copy help prompt</button><button type="button" className="btn btn-muted" aria-expanded={preview} onClick={() => setPreview(!preview)}>{preview ? "Hide prompt" : "Preview"}</button></div>
        <p className="form-note" role="status">{status}</p>
        {preview ? <textarea className="prompt-preview" rows={14} readOnly value={prompt} aria-label="Help prompt" /> : null}
      </div>
    </details>
  );
}
