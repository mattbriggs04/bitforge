"use client";

import { useCallback, useEffect, useRef, useState, type PointerEvent as ReactPointerEvent } from "react";
import { Markdown } from "@/components/markdown";
import { ProblemDetail } from "@/lib/types";
import { ProblemWorkbench } from "@/components/problem-workbench";

type Props = {
  problem: ProblemDetail;
};

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

function stripDuplicateStatementTitle(markdown: string, title: string): string {
  const lines = markdown.split(/\r?\n/);
  let firstContentIndex = 0;

  while (firstContentIndex < lines.length && lines[firstContentIndex].trim() === "") {
    firstContentIndex += 1;
  }

  const firstLine = lines[firstContentIndex]?.trim();
  if (!firstLine?.startsWith("# ")) {
    return markdown;
  }

  const heading = firstLine.replace(/^#\s+/, "").trim();
  if (heading.toLowerCase() !== title.trim().toLowerCase()) {
    return markdown;
  }

  const remaining = lines.slice(firstContentIndex + 1);
  while (remaining[0]?.trim() === "") {
    remaining.shift();
  }
  return remaining.join("\n");
}

function optionalString(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function getSolutionMarkdown(problem: ProblemDetail): string {
  const direct = optionalString(problem.metadata.solutionMarkdown);
  if (direct) {
    return direct;
  }

  const solution = problem.metadata.solution;
  if (typeof solution === "string") {
    return solution.trim();
  }

  const code = optionalString(problem.metadata.solutionCode);
  if (code) {
    const language = optionalString(problem.metadata.solutionLanguage) || "c";
    return ["```" + language, code, "```"].join("\n");
  }

  if (solution && typeof solution === "object") {
    const data = solution as Record<string, unknown>;
    const markdown = optionalString(data.markdown);
    if (markdown) {
      return markdown;
    }

    const explanation = optionalString(data.explanation);
    const nestedCode = optionalString(data.code);
    const language = optionalString(data.language) || "c";
    if (explanation && nestedCode) {
      return [explanation, "", "```" + language, nestedCode, "```"].join("\n");
    }
    if (nestedCode) {
      return ["```" + language, nestedCode, "```"].join("\n");
    }
  }

  return "";
}

export function ProblemWorkspace({ problem }: Props) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [leftPanePercent, setLeftPanePercent] = useState(40);
  const [dragging, setDragging] = useState(false);
  const statementMarkdown = stripDuplicateStatementTitle(problem.statementMarkdown, problem.title);
  const solutionMarkdown = getSolutionMarkdown(problem);

  const updateSplit = useCallback((clientX: number): void => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    const bounds = container.getBoundingClientRect();
    if (bounds.width <= 0) {
      return;
    }

    const ratio = ((clientX - bounds.left) / bounds.width) * 100;
    setLeftPanePercent(clamp(ratio, 20, 70));
  }, []);

  useEffect(() => {
    if (!dragging) {
      return;
    }

    const onMove = (event: PointerEvent): void => {
      updateSplit(event.clientX);
    };
    const stopDragging = (): void => {
      setDragging(false);
    };

    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", stopDragging);
    window.addEventListener("pointercancel", stopDragging);

    const previousCursor = document.body.style.cursor;
    const previousSelection = document.body.style.userSelect;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    return () => {
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerup", stopDragging);
      window.removeEventListener("pointercancel", stopDragging);
      document.body.style.cursor = previousCursor;
      document.body.style.userSelect = previousSelection;
    };
  }, [dragging, updateSplit]);

  const handleDragStart = (event: ReactPointerEvent<HTMLButtonElement>): void => {
    event.preventDefault();
    updateSplit(event.clientX);
    setDragging(true);
  };

  return (
    <section className="problem-workspace-container">
      <div ref={containerRef} className="problem-workspace-resizable">
        <article className="problem-panel problem-panel-left" style={{ width: `${leftPanePercent}%` }}>
          <div className="problem-pane-section markdown-body">
            <h2>Problem</h2>
            <Markdown>{statementMarkdown}</Markdown>
          </div>

          <div className="problem-pane-section markdown-body">
            <h2>Constraints</h2>
            <Markdown>{problem.constraintsMarkdown}</Markdown>
          </div>

          {problem.samples.length > 0 ? (
            <div className="problem-pane-section">
              <h2>Examples</h2>
              <div className="sample-grid">{problem.samples.map((sample) => (
                <article key={sample.name} className="sample-card">
                  <h3>{sample.name}</h3>
                  <p><strong>Input</strong> {sample.input}</p>
                  <p><strong>Expected</strong> {sample.expected}</p>
                  {sample.explanation ? <p className="sample-explanation"><Markdown inline>{sample.explanation}</Markdown></p> : null}
                </article>
              ))}</div>
            </div>
          ) : null}

          {solutionMarkdown ? (
            <div className="problem-pane-section">
              <details className="solution-panel">
                <summary>Solution</summary>
                <div className="solution-body markdown-body">
                  <Markdown>{solutionMarkdown}</Markdown>
                </div>
              </details>
            </div>
          ) : null}

          {problem.assets.length > 0 ? (
            <div className="problem-pane-section">
              <h2>Reference</h2>
              <div className="asset-grid">
                {problem.assets.map((asset) => (
                  <article key={asset.name} className="asset-card">
                    <h3>{asset.name}</h3>
                    <p className="asset-meta">
                      {asset.assetType} · {asset.mimeType}
                    </p>
                    <pre>{asset.content}</pre>
                  </article>
                ))}
              </div>
            </div>
          ) : null}
        </article>

        <button
          className={`problem-pane-divider ${dragging ? "is-dragging" : ""}`}
          type="button"
          aria-label="Resize problem and editor panels"
          onPointerDown={handleDragStart}
          onKeyDown={(event) => { if (event.key === "ArrowLeft" || event.key === "ArrowRight") { event.preventDefault(); setLeftPanePercent((n) => clamp(n + (event.key === "ArrowLeft" ? -5 : 5), 20, 70)); } }}
        >
          <span aria-hidden />
        </button>

        <div className="problem-pane-right">
          <ProblemWorkbench problem={problem} />
        </div>
      </div>
    </section>
  );
}
