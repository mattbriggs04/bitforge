"use client";

import { useRef, useState } from "react";
import Link from "next/link";
import { HelpPrompt } from "@/components/help-prompt";
import { copyPrompt } from "@/lib/tutor-prompt";
import { problemLabel } from "@/lib/problem-labels";
import type { CompiledPrompt, GenerationConfig, ProblemDetail, ProblemSummary, PromptCatalog, QuestionMix } from "@/lib/types";

function download(name: string, value: string) {
  const url = URL.createObjectURL(new Blob([value], { type: "text/plain;charset=utf-8" }));
  const link = document.createElement("a"); link.href = url; link.download = name; link.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
export function PromptBuilder({ catalog, problems, initialProblem, initialMode }: { catalog: PromptCatalog; problems: ProblemSummary[]; initialProblem: ProblemDetail | null; initialMode: string }) {
  const role = catalog.roles.find((item) => item.id === "ai_ml") ?? catalog.roles[0];
  const [mode, setMode] = useState(initialMode);
  const [config, setConfig] = useState<GenerationConfig>({ schemaVersion: 1, kind: "interview", role: role.id, preset: "daily", category: role.category, format: "coding", language: role.language, difficulty: "medium", level: "Full-time interview", topic: "", targetMinutes: 25, seed: 0, avoid: [], objectives: [], mix: [] });
  const [compiled, setCompiled] = useState<CompiledPrompt | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("");
  const revision = useRef(0);
  function update(patch: Partial<GenerationConfig>) { revision.current += 1; setConfig((value) => ({ ...value, ...patch })); setCompiled(null); setError(""); setStatus(""); }
  const mix = config.preset === "custom" ? config.mix : catalog.defaults[config.role]?.[config.preset] ?? [];
  function changeMix(index: number, patch: Partial<QuestionMix>) { update({ mix: mix.map((item, i) => i === index ? { ...item, ...patch } : item), preset: "custom" }); }
  async function compile(event: React.FormEvent) {
    event.preventDefault(); const currentRevision = ++revision.current; setBusy(true); setError(""); setCompiled(null); setStatus("");
    try {
      if (!Number.isSafeInteger(config.seed) || config.seed < 0) throw new Error("Seed must be a nonnegative safe integer.");
      const response = await fetch("/api/backend/v1/prompts/compile", { method: "POST", headers: { "content-type": "application/json" }, body: JSON.stringify(config) });
      const data = await response.json(); if (!response.ok) throw new Error(data.error || "Could not build prompt.");
      if (currentRevision === revision.current) setCompiled(data);
    } catch (e) { setError(e instanceof Error ? e.message : "Could not build prompt."); }
    finally { setBusy(false); }
  }
  return <>
    <div className="prompt-modes" role="group" aria-label="Prompt purpose">
      <button className="btn btn-muted" aria-pressed={mode === "generate"} onClick={() => setMode("generate")}>Create practice</button>
      <button className="btn btn-muted" aria-pressed={mode === "help"} onClick={() => setMode("help")}>Get help</button>
    </div>
    {mode === "help" ? <TutorBuilder problems={problems} initialProblem={initialProblem} /> : <div className="prompt-layout">
      <form className="prompt-form" onSubmit={(event) => void compile(event)}>
        <div className="prompt-field-grid">
          <label>Create<select value={config.kind} onChange={(event) => update({ kind: event.target.value })}><option value="interview">Interview / practice set</option><option value="question">Single question</option></select></label>
          <label>Role<select value={config.role} onChange={(event) => { const next = catalog.roles.find((item) => item.id === event.target.value); update({ role: event.target.value, category: next?.category ?? "random", language: next?.language ?? "random" }); }}>
            {catalog.roles.map((item) => <option key={item.id} value={item.id}>{item.label}</option>)}<option value="random">Random</option>
          </select></label>
          {config.kind === "interview" ? <label>Preset<select value={config.preset} onChange={(event) => { const preset = catalog.presets.find((item) => item.id === event.target.value); update({ preset: event.target.value, targetMinutes: preset?.targetMinutes ?? 45, mix: event.target.value === "custom" ? mix.length ? mix : [{ format: "coding", focus: "applied", count: 1 }] : [] }); }}>
            {catalog.presets.map((item) => <option key={item.id} value={item.id}>{item.label}</option>)}
          </select></label> : <label>Format<select value={config.format} onChange={(event) => update({ format: event.target.value })}>{catalog.formats.map((format) => <option key={format} value={format}>{problemLabel(format)}</option>)}<option value="random">Random</option></select></label>}
          <label>Time budget (min)<input type="number" required min={1} max={180} value={config.targetMinutes} onChange={(event) => update({ targetMinutes: event.target.valueAsNumber })} /></label>
          <label>Level<select value={config.level} onChange={(event) => update({ level: event.target.value })}><option>Full-time interview</option><option>Graduate course</option><option>Undergraduate course</option></select></label>
          <label>Difficulty<select value={config.difficulty} onChange={(event) => update({ difficulty: event.target.value })}>{["easy", "medium", "hard", "random"].map((item) => <option key={item} value={item}>{item[0].toUpperCase() + item.slice(1)}</option>)}</select></label>
          <label>Language<select value={config.language} onChange={(event) => update({ language: event.target.value })}>{[["c", "C11"], ["cpp", "C++20"], ["python", "Python / NumPy"], ["rust", "Rust"], ["random", "Random"]].map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></label>
          <label>Topic<input value={config.topic} maxLength={2000} onChange={(event) => update({ topic: event.target.value })} placeholder="Random, or something specific" /></label>
        </div>
        {config.kind === "interview" ? <fieldset className="mix-editor"><legend>Question mix</legend>
          {config.role === "random" && config.preset !== "custom" ? <p className="form-note">The selected role determines the mix when you build the prompt.</p> : mix.map((item, index) => <div className="mix-row" key={index}>
            <select aria-label={`Question group ${index + 1} format`} value={item.format} disabled={config.preset !== "custom"} onChange={(event) => changeMix(index, { format: event.target.value })}>{catalog.formats.map((format) => <option key={format} value={format}>{problemLabel(format)}</option>)}</select>
            <select aria-label={`Question group ${index + 1} focus`} value={item.focus} disabled={config.preset !== "custom"} onChange={(event) => changeMix(index, { focus: event.target.value })}>{catalog.focuses.map((focus) => <option key={focus} value={focus}>{focus[0].toUpperCase() + focus.slice(1)}</option>)}</select>
            <input type="number" min={1} max={20} required aria-label={`Question group ${index + 1} count`} value={item.count} disabled={config.preset !== "custom"} onChange={(event) => changeMix(index, { count: event.target.valueAsNumber })} />
            {config.preset === "custom" ? <button type="button" className="btn btn-muted" aria-label={`Remove question group ${index + 1}`} onClick={() => update({ mix: mix.filter((_, i) => i !== index) })}>×</button> : null}
          </div>)}
          <button type="button" className="btn btn-muted" onClick={() => update({ preset: "custom", mix: config.preset === "custom" ? [...mix, { format: "single_choice", focus: "reasoning", count: 1 }] : mix.length ? mix : [{ format: "coding", focus: "applied", count: 1 }] })}>{config.preset === "custom" ? "+ Add group" : "Customize mix"}</button>
        </fieldset> : null}
        <details className="prompt-options"><summary>Objectives & exclusions</summary><div className="prompt-advanced">
          <label>Objectives<textarea rows={3} value={config.objectives.join("\n")} onChange={(event) => update({ objectives: event.target.value.split("\n") })} placeholder="One per line, or let the agent propose them" /></label>
          <label>Avoid<textarea rows={3} value={config.avoid.join("\n")} onChange={(event) => update({ avoid: event.target.value.split("\n") })} placeholder="Topics or problems, one per line" /></label>
          <label>Seed<input type="number" min={0} max={Number.MAX_SAFE_INTEGER} step={1} value={config.seed} onChange={(event) => update({ seed: event.target.valueAsNumber })} /></label><p className="form-note">0 picks a fresh seed. Save the resolved config to repeat a selection.</p>
        </div></details>
        <button className="btn btn-primary" disabled={busy} type="submit">{busy ? "Building…" : "Build prompt"}<span aria-hidden>↗</span></button>
        {error ? <p role="alert" className="catalog-warning">{error}</p> : null}
      </form>
      <section className="prompt-output" aria-label="Generated prompt">
        {compiled ? <>
          <div className="prompt-output-meta"><span className="tag tag-category">{problemLabel(compiled.config.role)}</span><span>{compiled.config.targetMinutes} min</span><span>Seed {compiled.config.seed}</span></div>
          <div className="prompt-actions"><button className="btn btn-primary" onClick={async () => { try { await copyPrompt(compiled.prompt); setStatus("Copied. Paste into an agent working in this repository."); } catch { setStatus("Select and copy the prompt below."); } }}>Copy prompt</button><button className="btn btn-muted" onClick={() => download("bitforge-request.json", JSON.stringify(compiled.config, null, 2) + "\n")}>Save config</button><button className="btn btn-muted" onClick={() => download("bitforge-prompt.txt", compiled.prompt)}>Save prompt</button></div>
          <p className="form-note" role="status">{status}</p>
          <label>Agent prompt<textarea className="prompt-preview" rows={24} readOnly value={compiled.prompt} /></label>
        </> : <div className="prompt-empty"><span className="prompt-cursor" aria-hidden>_</span><h2>Agent brief</h2><p>{catalog.roles.find((item) => item.id === config.role)?.emphasis ?? "Let the seed choose a role and its defaults."}</p><p>Your agent authors the questions, generates tests, validates, and imports. Two hints come with each question.</p><Link href="/interviews">Browse interviews ↗</Link></div>}
      </section>
    </div>}
  </>;
}

function TutorBuilder({ problems, initialProblem }: { problems: ProblemSummary[]; initialProblem: ProblemDetail | null }) {
  const [problem, setProblem] = useState(initialProblem);
  const [slug, setSlug] = useState(initialProblem?.slug ?? "");
  const [language, setLanguage] = useState(initialProblem?.languageTemplates[0]?.language ?? "answer");
  const [work, setWork] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function load(nextSlug: string) {
    setSlug(nextSlug); setProblem(null); setWork(""); setError("");
    if (!nextSlug) return;
    setBusy(true);
    try { const response = await fetch(`/api/backend/v1/problems/${encodeURIComponent(nextSlug)}`); const data = await response.json(); if (!response.ok) throw new Error(data.error || "Could not load problem."); setProblem(data); setLanguage(data.languageTemplates[0]?.language ?? "answer"); }
    catch (e) { setError(e instanceof Error ? e.message : "Could not load problem."); }
    finally { setBusy(false); }
  }
  function loadDraft() {
    if (!problem) return;
    try {
      const key = language === "answer" ? `bitforge:answer:${problem.slug}:${problem.metadata.bundleHash ?? "v1"}` : `bitforge:code:${problem.slug}:${problem.metadata.bundleHash ?? "legacy"}:${language}`;
      const value = localStorage.getItem(key); if (!value) { setError("No saved draft for this question and language."); return; }
      setWork(value); setError("");
    } catch { setError("Browser storage is unavailable. Paste your work below."); }
  }
  return <div className="tutor-builder">
    <label>Problem<select value={slug} disabled={busy} onChange={(event) => void load(event.target.value)}><option value="">Choose a problem</option>{problems.map((item) => <option key={item.slug} value={item.slug}>{item.title}</option>)}</select></label>
    {busy ? <p role="status">Loading…</p> : null}{error ? <p role="alert">{error}</p> : null}
    {problem ? <>
      <div className="tutor-work-head"><label>Language<select value={language} onChange={(event) => { setLanguage(event.target.value); setWork(""); }}>{problem.languageTemplates.length ? problem.languageTemplates.map((item) => <option key={item.language} value={item.language}>{item.language}</option>) : <option value="answer">Answer / reasoning</option>}</select></label><button className="btn btn-muted" onClick={loadDraft}>Load saved draft</button><Link className="btn btn-muted" href={`/problems/${problem.slug}`}>Open problem ↗</Link></div>
      <label>Your work<textarea className="prompt-preview" rows={10} value={work} onChange={(event) => setWork(event.target.value)} placeholder="Paste code, your reasoning, or an error message." /></label>
      <HelpPrompt key={`${problem.slug}:${language}`} problem={problem} work={work} language={language} expanded />
    </> : null}
  </div>;
}
