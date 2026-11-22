"use client";

import Link from "next/link";
import { useState } from "react";
import { Markdown } from "@/components/markdown";
import type { ProblemDetail } from "@/lib/types";

export function ProblemHints({ problem }: { problem: ProblemDetail }) {
  const [approach, setApproach] = useState(false);
  if (!problem.hints?.nudge) return null;
  return <details className="problem-hints">
    <summary>Hints</summary>
    <div className="hint-step markdown-body"><h3>Nudge</h3><Markdown>{problem.hints.nudge}</Markdown></div>
    {approach ? <div className="hint-step markdown-body"><h3>Approach</h3><Markdown>{problem.hints.approach}</Markdown></div> : <button type="button" className="btn btn-muted" onClick={() => setApproach(true)}>Reveal approach</button>}
    <Link className="hint-help-link" href={`/prompts?mode=help&problem=${encodeURIComponent(problem.slug)}`}>Scaffold or walkthrough with an agent ↗</Link>
  </details>;
}
