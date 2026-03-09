export type ProblemSummary = {
  id: string;
  slug: string;
  title: string;
  difficulty: "easy" | "medium" | "hard" | string;
  category: string;
  problemType: string;
  shortDescription: string;
  tags: string[];
};

export type ProblemSample = {
  name: string;
  input: string;
  expected: string;
  explanation: string;
  sortOrder: number;
};

export type LanguageTemplate = {
  language: string;
  starterCode: string;
  notes: string;
};

export type ProblemAsset = {
  assetType: string;
  name: string;
  mimeType: string;
  content: string;
  metadata: Record<string, unknown>;
};

export type ProblemDetail = ProblemSummary & {
  statementMarkdown: string;
  constraintsMarkdown: string;
  samples: ProblemSample[];
  languageTemplates: LanguageTemplate[];
  assets: ProblemAsset[];
  metadata: Record<string, unknown>;
};

export type SubmissionCaseResult = {
  caseName: string;
  isHidden: boolean;
  status: "passed" | "failed" | "error" | "skipped" | string;
  message: string;
  executionMs: number;
  sortOrder: number;
};

export type Submission = {
  id: string;
  problemId: string;
  problemSlug: string;
  userId: string;
  language: string;
  mode: "run" | "submit" | string;
  status: "queued" | "running" | "completed" | "failed" | string;
  verdict:
    | "pending"
    | "accepted"
    | "wrong_answer"
    | "compile_error"
    | "runtime_error"
    | "system_error"
    | string;
  score: number;
  totalTests: number;
  passedTests: number;
  compileOutput?: string;
  runtimeOutput?: string;
  errorMessage?: string;
  queuedAt: string;
  startedAt?: string;
  completedAt?: string;
  results?: SubmissionCaseResult[];
};

export type APIError = {
  error: string;
};
