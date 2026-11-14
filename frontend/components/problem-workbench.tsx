"use client";

import { useTheme } from "@/lib/theme";
import { configureEditor } from "@/lib/editor-theme";
import { AnswerWorkbench } from "@/components/answer-workbench";

import dynamic from "next/dynamic";
import { useEffect, useMemo, useState } from "react";
import { ProblemDetail, Submission, SubmissionCaseResult } from "@/lib/types";
import { loadOrCreateUserKey, loadStoredUsername } from "@/lib/username";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), {
  ssr: false,
  loading: () => <div className="editor-loading">Loading editor…</div>,
});

type Props = {
  problem: ProblemDetail;
};

type SubmissionMode = "run" | "submit";

function formatVerdict(verdict: string): string {
  return verdict.replaceAll("_", " ");
}

function normalizeError(input: unknown): string {
  if (typeof input === "string") {
    return input;
  }
  return "Request failed";
}

function normalizeSubmission(payload: Submission): Submission {
  return {
    ...payload,
    results: Array.isArray(payload.results) ? payload.results : [],
  };
}

async function fetchSubmission(submissionID: string): Promise<Submission> {
  const response = await fetch(`/api/backend/v1/submissions/${submissionID}`, {
    cache: "no-store",
  });

  const data = (await response.json()) as Submission | { error: string };
  if (!response.ok || "error" in data) {
    const message = "error" in data ? data.error : `Failed to load submission (${response.status})`;
    throw new Error(message);
  }

  return normalizeSubmission(data);
}

function caseStatusClass(status: string): string {
  if (status === "passed") {
    return "case-status-passed";
  }
  if (status === "failed") {
    return "case-status-failed";
  }
  if (status === "error") {
    return "case-status-error";
  }
  return "case-status-pending";
}

export function ProblemWorkbench({ problem }: Props) {
  if (["single_choice", "multiple_select", "numeric"].includes(problem.problemType)) return <AnswerWorkbench key={problem.slug} problem={problem} />;
  return <CodeWorkbench key={problem.slug} problem={problem} />;
}

function CodeWorkbench({ problem }: Props) {
  const theme = useTheme();
  const [language, setLanguage] = useState(problem.languageTemplates[0]?.language ?? "c");
  const defaultTemplate = useMemo(
    () => problem.languageTemplates.find((item) => item.language === language) ?? problem.languageTemplates[0],
    [problem.languageTemplates, language],
  );

  const [code, setCode] = useState(defaultTemplate?.starterCode ?? "");
  const [submissionID, setSubmissionID] = useState<string>("");
  const [submission, setSubmission] = useState<Submission | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [pollRetry, setPollRetry] = useState(0);
  const [pollPaused, setPollPaused] = useState(false);
  const [userHandle, setUserHandle] = useState("demo");
  const [userKey, setUserKey] = useState("");

  const draftKey = `bitforge:code:${problem.slug}:${problem.metadata.bundleHash ?? "legacy"}:${language}`;
  useEffect(() => {
    try { setCode(localStorage.getItem(draftKey) ?? defaultTemplate?.starterCode ?? ""); }
    catch { setCode(defaultTemplate?.starterCode ?? ""); }
  }, [defaultTemplate?.starterCode, draftKey]);
  function updateCode(value: string) {
    setCode(value);
    try { localStorage.setItem(draftKey, value); } catch { /* Storage may be disabled. */ }
  }

  useEffect(() => {
    setUserKey(loadOrCreateUserKey());
    const stored = loadStoredUsername();
    if (stored) {
      setUserHandle(stored);
    }
  }, []);

  useEffect(() => {
    if (!submissionID) {
      return;
    }

    let active = true;
    let failures = 0;
    setPollPaused(false);
    let timer: ReturnType<typeof setTimeout> | null = null;

    const poll = async (): Promise<void> => {
      try {
        const latest = await fetchSubmission(submissionID);
        if (!active) {
          return;
        }
        setSubmission(latest);
        failures = 0; setError("");
        if (latest.status === "queued" || latest.status === "running") {
          timer = setTimeout(poll, 900);
        }
      } catch (pollErr) {
        if (!active) {
          return;
        }
        if (++failures < 4) { timer = setTimeout(poll, failures * 1500); return; }
        setPollPaused(true);
        setError(pollErr instanceof Error ? pollErr.message : "Failed to poll submission");
      }
    };

    void poll();

    return () => {
      active = false;
      if (timer) {
        clearTimeout(timer);
      }
    };
  }, [submissionID, pollRetry]);

  const startSubmission = async (mode: SubmissionMode): Promise<void> => {
    if (!code.trim()) {
      setError("Source code is empty");
      return;
    }
    const activeUserKey = userKey || loadOrCreateUserKey();
    if (!userKey && activeUserKey) {
      setUserKey(activeUserKey);
    }

    setIsSubmitting(true);
    setError("");
    setSubmission(null);
    setSubmissionID("");

    try {
      const response = await fetch("/api/backend/v1/submissions", {
        method: "POST",
        headers: {
          "content-type": "application/json",
          "x-user-handle": userHandle || "demo",
          "x-user-key": activeUserKey,
        },
        body: JSON.stringify({
          problemSlug: problem.slug,
          language,
          mode,
          sourceCode: code,
        }),
      });

      const data = (await response.json()) as { submissionId?: string; error?: string };
      if (!response.ok || !data.submissionId) {
        throw new Error(data.error ?? `Submission failed (${response.status})`);
      }

      setSubmissionID(data.submissionId);
    } catch (submitErr) {
      setError(normalizeError(submitErr instanceof Error ? submitErr.message : submitErr));
    } finally {
      setIsSubmitting(false);
    }
  };

  const submissionResults: SubmissionCaseResult[] = useMemo(
    () => (submission && Array.isArray(submission.results) ? submission.results : []),
    [submission],
  );

  const inFlight = !pollPaused && Boolean(submissionID) && (!submission || submission.status === "queued" || submission.status === "running");

  return (
    <section className="solve-section">
      <div className="solve-toolbar">
        <label>Language <select value={language} onChange={(e) => setLanguage(e.target.value)} disabled={isSubmitting || inFlight}>
          {problem.languageTemplates.map((template) => <option key={template.language} value={template.language}>{({ c: "C11", cpp: "C++20", python: "Python 3 + NumPy", rust: "Rust 2021" } as Record<string,string>)[template.language] ?? template.language}</option>)}
        </select></label>
        <p className="debug-hint">
          {problem.problemType === "coding" ? "stdin → stdout" : "Function implementation"}
        </p>
      </div>
      <div className="editor-wrap">
        <MonacoEditor
          height="560px"
          language={language === "cpp" ? "cpp" : language}
          theme={`bitforge-${theme}`}
          beforeMount={configureEditor}
          value={code}
          onChange={(value) => updateCode(value ?? "")}
          options={{
            fontSize: 13,
            fontFamily: "SFMono-Regular, Consolas, Liberation Mono, monospace",
            fontLigatures: false,
            overviewRulerLanes: 0,
            hideCursorInOverviewRuler: true,
            renderLineHighlight: "none",
            scrollbar: { verticalScrollbarSize: 5, horizontalScrollbarSize: 5 },
            minimap: { enabled: false },
            lineNumbersMinChars: 3,
            padding: { top: 16 },
            scrollBeyondLastLine: false,
            automaticLayout: true,
          }}
        />
      </div>

      <div className="solve-actions solve-actions-footer">
        <span className="draft-status">Draft saved locally</span>
        <button className="btn btn-muted" type="button" onClick={() => { if (window.confirm("Reset this draft to the starter code?")) updateCode(defaultTemplate?.starterCode ?? ""); }}>
          Reset
        </button>
        <button
          className="btn btn-muted"
          type="button"
          disabled={isSubmitting || inFlight}
          onClick={() => void startSubmission("run")}
        >
          Run
        </button>
        <button
          className="btn btn-primary"
          type="button"
          disabled={isSubmitting || inFlight}
          onClick={() => void startSubmission("submit")}
        >
          Submit
        </button>
      </div>

      <section className="judge-terminal" aria-live="polite">
        <div className="judge-terminal-head">
          <h3>Results</h3>
          {submissionID ? <code>{submissionID.slice(0, 8)}</code> : <code>idle</code>}
        </div>

        <div className="judge-terminal-body">
          {error ? <p className="terminal-line terminal-line-error">[error] {error} {pollPaused ? <button className="btn btn-muted" onClick={() => setPollRetry((n) => n + 1)}>Retry results</button> : null}</p> : null}

          {!submission && !error ? (
            <p className="terminal-line terminal-line-dim">Run examples or submit.</p>
          ) : null}

          {inFlight ? <p className="terminal-line terminal-line-dim">Judging…</p> : null}

          {submission ? (
            <>
              <div className="result-overview terminal-overview">
                <div>
                  <span className="result-label">Status</span>
                  <strong>{submission.status}</strong>
                </div>
                <div>
                  <span className="result-label">Verdict</span>
                  <strong>{formatVerdict(submission.verdict)}</strong>
                </div>
                <div>
                  <span className="result-label">Tests</span>
                  <strong>
                    {submission.passedTests}/{submission.totalTests}
                  </strong>
                </div>
                <div>
                  <span className="result-label">Score</span>
                  <strong>{submission.score}</strong>
                </div>
              </div>

              {submission.errorMessage ? (
                <p className="terminal-line terminal-line-error">[judge] {submission.errorMessage}</p>
              ) : null}

              {submissionResults.length > 0 ? (
                <ul className="terminal-case-list">
                  {submissionResults.map((item) => (
                    <li key={`${item.sortOrder}-${item.caseName}`} className={caseStatusClass(item.status)}>
                      <span className="terminal-case-status">{item.status.toUpperCase()}</span>
                      <span className="terminal-case-name">{item.caseName}</span>
                      <span className="terminal-case-message">{item.message}</span>
                    </li>
                  ))}
                </ul>
              ) : null}

              {submission.compileOutput ? (
                <details className="terminal-details">
                  <summary>Compiler</summary>
                  <pre>{submission.compileOutput}</pre>
                </details>
              ) : null}

              {submission.runtimeOutput ? (
                <details className="terminal-details">
                  <summary>Output</summary>
                  <pre>{submission.runtimeOutput}</pre>
                </details>
              ) : null}
            </>
          ) : null}
        </div>
      </section>
    </section>
  );
}
